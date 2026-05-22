package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"openriot/notify"
	"openriot/paths"
)

// stateFile returns the path to the volume state file.
func stateFile() string {
	if paths.HomeDir() == "" {
		return ""
	}
	return paths.Join(".config", "openriot", "volume.state")
}

// saveVolumeState reads current sndioctl values and writes them to disk.
func saveVolumeState() {
	path := stateFile()
	if path == "" {
		return
	}
	level, _ := exec.Command("sndioctl", "-n", "output.level").Output()
	mute, _ := exec.Command("sndioctl", "-n", "output.mute").Output()
	micLevel, _ := exec.Command("sndioctl", "-n", "input.level").Output()
	micMute, _ := exec.Command("sndioctl", "-n", "input.mute").Output()

	content := fmt.Sprintf("output.level=%s\noutput.mute=%s\ninput.level=%s\ninput.mute=%s\n",
		strings.TrimSpace(string(level)),
		strings.TrimSpace(string(mute)),
		strings.TrimSpace(string(micLevel)),
		strings.TrimSpace(string(micMute)),
	)
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(content), 0644)
}

// restoreVolumeState reads saved values and applies them via sndioctl.
func restoreVolumeState() {
	path := stateFile()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		ctrl := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if ctrl == "" || val == "" {
			continue
		}
		exec.Command("sndioctl", fmt.Sprintf("%s=%s", ctrl, val)).Run()
	}
}

func sndioctlValue(ctrl string) string {
	out, _ := exec.Command("sndioctl", "-n", ctrl).Output()
	return strings.TrimSpace(string(out))
}

func vol() string    { return sndioctlValue("output.level") }
func micVol() string { return sndioctlValue("input.level") }
func isMuted() bool {
	v, err := strconv.ParseFloat(sndioctlValue("output.mute"), 64)
	return err == nil && v == 1.0
}
func micMuted() bool {
	v, err := strconv.ParseFloat(sndioctlValue("input.mute"), 64)
	return err == nil && v == 1.0
}

// Run executes volume subcommands using OpenBSD's sndioctl.
// Supported: toggle, inc, dec, get, restore, mic-toggle, mic-inc, mic-dec, mic-get
func Run(args []string) int {
	usage := func() int {
		fmt.Fprintln(os.Stderr, "Usage: openriot --volume [toggle|inc|dec|get|restore|mic-toggle|mic-inc|mic-dec|mic-get]")
		return 1
	}

	sendNotify := func(icon, msg string) {
		notify.SendNotify(icon, "Settings", msg, "normal", 1500, 0)
	}

	sndioctl := func(cmd string) error {
		parts := strings.Fields(cmd)
		c := exec.Command("sndioctl", parts...)
		out, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sndioctl %s: %w: %s", cmd, err, string(out))
		}
		return nil
	}

	// toPercent converts sndioctl float (0-1) to percentage string (0-100)
	toPercent := func(raw string) string {
		f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return raw
		}
		return strconv.Itoa(int(f * 100))
	}

	if len(args) < 1 {
		return usage()
	}

	switch args[0] {
	case "toggle":
		// Toggle using sh -c like original working code
		cmd := exec.Command("sh", "-c", "sndioctl output.mute=!")
		out, err := cmd.CombinedOutput()
		if err != nil {
			f, _ := os.OpenFile("/tmp/openriot-mute.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				fmt.Fprintf(f, "TOGGLE ERROR: %v: %s\n", err, string(out))
				f.Close()
			}
			sendNotify("speaker", fmt.Sprintf("Mute error: %v", err))
			return 1
		}
		saveVolumeState()
		// Check mute state after toggle
		cur, _ := exec.Command("sndioctl", "-n", "output.mute").Output()
		curStr := strings.TrimSpace(string(cur))
		if curStr == "1" || curStr == "1.0" {
			sendNotify("speaker-muted", "Speaker: Muted")
		} else {
			sendNotify("speaker", fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		}
		return 0
	case "inc":
		sndioctl("output.level=+0.05")
		saveVolumeState()
		sendNotify("speaker", fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		return 0
	case "dec":
		sndioctl("output.level=-0.05")
		saveVolumeState()
		sendNotify("speaker", fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		return 0
	case "get":
		fmt.Println(vol())
		return 0
	case "restore":
		restoreVolumeState()
		return 0
	case "mic-toggle":
		sndioctl("input.mute=!")
		saveVolumeState()
		if micMuted() {
			sendNotify("mic-muted", "Mic: Muted")
		} else {
			sendNotify("mic", fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		}
		return 0
	case "mic-inc":
		sndioctl("input.level=+0.05")
		saveVolumeState()
		sendNotify("mic", fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		return 0
	case "mic-dec":
		sndioctl("input.level=-0.05")
		saveVolumeState()
		sendNotify("mic", fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		return 0
	case "mic-get":
		fmt.Println(micVol())
		return 0
	default:
		return usage()
	}
}

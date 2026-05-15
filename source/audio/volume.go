package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"openriot/notify"
)

// stateFile returns the path to the volume state file.
func stateFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "openriot", "volume.state")
}

// saveVolumeState reads current sndioctl values and writes them to disk.
func saveVolumeState() {
	path := stateFile()
	if path == "" {
		return
	}
	level, _ := exec.Command("sh", "-c", "sndioctl -n output.level 2>/dev/null").Output()
	mute, _ := exec.Command("sh", "-c", "sndioctl -n output.mute 2>/dev/null").Output()
	micLevel, _ := exec.Command("sh", "-c", "sndioctl -n input.level 2>/dev/null").Output()
	micMute, _ := exec.Command("sh", "-c", "sndioctl -n input.mute 2>/dev/null").Output()

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
		return c.Run()
	}

	vol := func() string {
		out, _ := exec.Command("sh", "-c", "sndioctl output.level 2>/dev/null | cut -d= -f2").Output()
		return strings.TrimSpace(string(out))
	}

	micVol := func() string {
		out, _ := exec.Command("sh", "-c", "sndioctl input.level 2>/dev/null | cut -d= -f2").Output()
		return strings.TrimSpace(string(out))
	}

	isMuted := func() bool {
		out, _ := exec.Command("sh", "-c", "sndioctl output.mute 2>/dev/null | cut -d= -f2").Output()
		return strings.TrimSpace(string(out)) == "1"
	}

	// toPercent converts sndioctl float (0-1) to percentage string (0-100)
	toPercent := func(raw string) string {
		f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return raw
		}
		return strconv.Itoa(int(f * 100))
	}

	micMuted := func() bool {
		out, _ := exec.Command("sh", "-c", "sndioctl input.mute 2>/dev/null | cut -d= -f2").Output()
		return strings.TrimSpace(string(out)) == "1"
	}

	if len(args) < 1 {
		return usage()
	}

	switch args[0] {
	case "toggle":
		sndioctl("output.mute=!")
		saveVolumeState()
		if isMuted() {
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

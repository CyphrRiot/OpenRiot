package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Run executes volume subcommands using OpenBSD's sndioctl.
// Supported: toggle, inc, dec, get, mic-toggle, mic-inc, mic-dec, mic-get
func Run(args []string) int {
	usage := func() int {
		fmt.Fprintln(os.Stderr, "Usage: openriot --volume [toggle|inc|dec|get|mic-toggle|mic-inc|mic-dec|mic-get]")
		return 1
	}

		home := os.Getenv("HOME")
	iconPath := filepath.Join(home, ".local/share/openriot/config/icons")
	speakerIcon := filepath.Join(iconPath, "speaker.png")
	speakerMutedIcon := filepath.Join(iconPath, "speaker-muted.png")
	micIcon := filepath.Join(iconPath, "mic.png")
	micMutedIcon := filepath.Join(iconPath, "mic-muted.png")
	notify := func(icon, msg string) {
		exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "1500", "Settings", msg).Start()
	}
	_ = speakerIcon // used in switch below

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
		if isMuted() {
			notify(speakerMutedIcon, "Speaker: Muted")
		} else {
			notify(speakerIcon, fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		}
		return 0
	case "inc":
		sndioctl("output.level=+0.05")
		notify(speakerIcon, fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		return 0
	case "dec":
		sndioctl("output.level=-0.05")
		notify(speakerIcon, fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		return 0
	case "get":
		fmt.Println(vol())
		return 0
	case "mic-toggle":
		sndioctl("input.mute=!")
		if micMuted() {
			notify(micMutedIcon, "Mic: Muted")
		} else {
			notify(micIcon, fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		}
		return 0
	case "mic-inc":
		sndioctl("input.level=+0.05")
		notify(micIcon, fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		return 0
	case "mic-dec":
		sndioctl("input.level=-0.05")
		notify(micIcon, fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		return 0
	case "mic-get":
		fmt.Println(micVol())
		return 0
	default:
		return usage()
	}
}

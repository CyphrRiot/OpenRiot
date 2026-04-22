package audio

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"openriot/notify"
)

// Run executes volume subcommands using OpenBSD's sndioctl.
// Supported: toggle, inc, dec, get, mic-toggle, mic-inc, mic-dec, mic-get
func Run(args []string) int {
	usage := func() int {
		fmt.Fprintln(os.Stderr, "Usage: openriot --volume [toggle|inc|dec|get|mic-toggle|mic-inc|mic-dec|mic-get]")
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
		if isMuted() {
			sendNotify("speaker-muted", "Speaker: Muted")
		} else {
			sendNotify("speaker", fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		}
		return 0
	case "inc":
		sndioctl("output.level=+0.05")
		sendNotify("speaker", fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		return 0
	case "dec":
		sndioctl("output.level=-0.05")
		sendNotify("speaker", fmt.Sprintf("Speaker: %s%%", toPercent(vol())))
		return 0
	case "get":
		fmt.Println(vol())
		return 0
	case "mic-toggle":
		sndioctl("input.mute=!")
		if micMuted() {
			sendNotify("mic-muted", "Mic: Muted")
		} else {
			sendNotify("mic", fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		}
		return 0
	case "mic-inc":
		sndioctl("input.level=+0.05")
		sendNotify("mic", fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		return 0
	case "mic-dec":
		sndioctl("input.level=-0.05")
		sendNotify("mic", fmt.Sprintf("Mic: %s%%", toPercent(micVol())))
		return 0
	case "mic-get":
		fmt.Println(micVol())
		return 0
	default:
		return usage()
	}
}

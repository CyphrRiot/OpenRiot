package audio

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes volume subcommands using OpenBSD's sndioctl.
// Supported: toggle, inc, dec, get, mic-toggle, mic-inc, mic-dec, mic-get
func Run(args []string) int {
	usage := func() int {
		fmt.Fprintln(os.Stderr, "Usage: openriot --volume [toggle|inc|dec|get|mic-toggle|mic-inc|mic-dec|mic-get]")
		return 1
	}

	notify := func(msg string) {
		// Try makoctl (OpenBSD notification daemon)
		if _, err := exec.LookPath("makoctl"); err == nil {
			_ = exec.Command("makoctl", "dismiss", "--all").Run()
			_ = exec.Command("makoctl", "append", "--app-name=Volume", msg).Start()
		}
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
			notify("Speaker: Muted")
		} else {
			notify(fmt.Sprintf("Speaker: %s%%", vol()))
		}
		return 0
	case "inc":
		sndioctl("output.level=+0.05")
		notify(fmt.Sprintf("Volume Up: %s%%", vol()))
		return 0
	case "dec":
		sndioctl("output.level=-0.05")
		notify(fmt.Sprintf("Volume Down: %s%%", vol()))
		return 0
	case "get":
		fmt.Println(vol())
		return 0
	case "mic-toggle":
		sndioctl("input.mute=!")
		if micMuted() {
			notify("Microphone: Muted")
		} else {
			notify(fmt.Sprintf("Microphone: %s%%", micVol()))
		}
		return 0
	case "mic-inc":
		sndioctl("input.level=+0.05")
		notify(fmt.Sprintf("Mic Up: %s%%", micVol()))
		return 0
	case "mic-dec":
		sndioctl("input.level=-0.05")
		notify(fmt.Sprintf("Mic Down: %s%%", micVol()))
		return 0
	case "mic-get":
		fmt.Println(micVol())
		return 0
	default:
		return usage()
	}
}

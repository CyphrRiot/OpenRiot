package display

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"openriot/notify"
)

// Run executes brightness subcommands.
// Tries wsconsctl (OpenBSD console) first, falls back to xbacklight (X11).
// Always shows notification (like volume) - even on failure.
func Run(args []string) int {
	usage := func() int {
		fmt.Fprintln(os.Stderr, "Usage: openriot --brightness [up|down|set <0-100>|get]")
		return 1
	}

	showNotify := func(msg string) {
		exec.Command("openriot", "--notify-dismiss").Run()
		notify.SendNotify("nightlight-on", "Settings", msg, "normal", 3000, 0)
	}

	// Try wsconsctl (OpenBSD console), then xbacklight (X11)
	runCmd := func(action string) bool {
		// Try wsconsctl first
		if exec.Command("wsconsctl", fmt.Sprintf("display.brightness=%s", action)).Run() == nil {
			return true
		}
		// Fallback to xbacklight
		var xbArgs []string
		switch action {
		case "+10":
			xbArgs = []string{"-inc", "10"}
		case "-10":
			xbArgs = []string{"-dec", "10"}
		default:
			// For set command, action is the value
			xbArgs = []string{"-set", action}
		}
		if exec.Command("xbacklight", xbArgs...).Run() == nil {
			return true
		}
		return false
	}

	// Get brightness - wsconsctl (0-255) or xbacklight (0-100)
	getBrightness := func() (int, bool) {
		// Try wsconsctl first
		out, _ := exec.Command("sh", "-c", "wsconsctl display.brightness 2>/dev/null | cut -d= -f2").Output()
		val, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err == nil && val > 0 {
			return val * 100 / 255, true
		}
		// Fallback to xbacklight
		out, _ = exec.Command("xbacklight", "-get").Output()
		val, err = strconv.Atoi(strings.TrimSpace(string(out)))
		if err == nil && val > 0 {
			return val, true
		}
		return 0, false
	}

	if len(args) < 1 {
		return usage()
	}

	switch args[0] {
	case "up":
		if !runCmd("+10") {
			showNotify("No backlight")
			return 1
		}
		b, ok := getBrightness()
		if ok {
			showNotify(fmt.Sprintf("%d%%", b))
		} else {
			showNotify("Brightness up")
		}
		return 0
	case "down":
		if !runCmd("-10") {
			showNotify("No backlight")
			return 1
		}
		b, ok := getBrightness()
		if ok {
			showNotify(fmt.Sprintf("%d%%", b))
		} else {
			showNotify("Brightness down")
		}
		return 0
	case "set":
		if len(args) < 2 {
			return usage()
		}
		val, err := strconv.Atoi(args[1])
		if err != nil || val < 0 || val > 100 {
			fmt.Fprintln(os.Stderr, "Error: brightness must be 0-100")
			return 1
		}
		// Try wsconsctl first (0-255), then xbacklight (0-100)
		wsval := val * 255 / 100
		if exec.Command("wsconsctl", fmt.Sprintf("display.brightness=%d", wsval)).Run() != nil {
			if exec.Command("xbacklight", "-set", args[1]).Run() != nil {
				showNotify("No backlight")
				return 1
			}
		}
		showNotify(fmt.Sprintf("%d%%", val))
		return 0
	case "get":
		b, ok := getBrightness()
		if ok {
			fmt.Println(b)
		}
		return 0
	default:
		return usage()
	}
}
package display

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Run executes brightness subcommands using OpenBSD's wsconsctl.
// NOTE: wsconsctl controls console brightness. Wayland compositors on OpenBSD
// do not have a standard brightness API. This may not work on all laptops.
// Supported: up, down, set <0-100>, get
func Run(args []string) int {
	usage := func() int {
		fmt.Fprintln(os.Stderr, "Usage: openriot --brightness [up|down|set <0-100>|get]")
		return 1
	}

	notify := func(msg string) {
		// Dismiss any existing notifications, show new one via waybar (auto-expires in 3s)
		exec.Command("openriot", "--notify-dismiss").Run()
		exec.Command("openriot", "--notify", "Brightness", msg, "--expires-in", "3").Start()
	}

	wsconsctl := func(cmd string) error {
		parts := strings.Fields(cmd)
		c := exec.Command("wsconsctl", parts...)
		return c.Run()
	}

	getBrightness := func() int {
		out, _ := exec.Command("sh", "-c", "wsconsctl display.brightness 2>/dev/null | cut -d= -f2").Output()
		val, _ := strconv.Atoi(strings.TrimSpace(string(out)))
		return val
	}

	if len(args) < 1 {
		return usage()
	}

	switch args[0] {
	case "up":
		if err := wsconsctl("display.brightness=+10"); err != nil {
			fmt.Fprintln(os.Stderr, "Error: wsconsctl failed (may require root)")
			return 1
		}
		notify(fmt.Sprintf("Brightness: %d", getBrightness()))
		return 0
	case "down":
		if err := wsconsctl("display.brightness=-10"); err != nil {
			fmt.Fprintln(os.Stderr, "Error: wsconsctl failed (may require root)")
			return 1
		}
		notify(fmt.Sprintf("Brightness: %d", getBrightness()))
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
		// wsconsctl takes 0-255 range
		wsval := val * 255 / 100
		if err := wsconsctl(fmt.Sprintf("display.brightness=%d", wsval)); err != nil {
			fmt.Fprintln(os.Stderr, "Error: wsconsctl failed")
			return 1
		}
		notify(fmt.Sprintf("Brightness: %d%%", val))
		return 0
	case "get":
		b := getBrightness()
		fmt.Println(b)
		return 0
	default:
		return usage()
	}
}

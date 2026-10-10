package display

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"openriot/notify"
)

// GetBrightness returns the current backlight percentage (0-100).
// Tries wsconsctl (0-255) then xbacklight (0-100). Returns 0 if unavailable.
func GetBrightness() int {
	out, _ := exec.Command("sh", "-c", "wsconsctl display.brightness 2>/dev/null | cut -d= -f2").Output()
	val, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err == nil && val > 0 {
		return val * 100 / 255
	}
	out, _ = exec.Command("xbacklight", "-get").Output()
	f, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err == nil {
		return int(f + 0.5)
	}
	return 0
}

// SetBrightness sets the backlight to the given percentage (0-100).
// Tries wsconsctl first, falls back to xbacklight.
func SetBrightness(pct int) error {
	wsval := pct * 255 / 100
	if exec.Command("wsconsctl", fmt.Sprintf("display.brightness=%d", wsval)).Run() == nil {
		return nil
	}
	return exec.Command("xbacklight", "-set", strconv.Itoa(pct)).Run()
}

// AdjustBrightness raises or lowers the backlight by pct percent.
func AdjustBrightness(delta int) error {
	if exec.Command("wsconsctl", fmt.Sprintf("display.brightness=%+d", delta)).Run() == nil {
		return nil
	}
	verb := "-inc"
	if delta < 0 {
		verb = "-dec"
		delta = -delta
	}
	return exec.Command("xbacklight", verb, strconv.Itoa(delta)).Run()
}

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

	if len(args) < 1 {
		return usage()
	}

	switch args[0] {
	case "up":
		if err := AdjustBrightness(10); err != nil {
			showNotify("No backlight")
			return 1
		}
		showNotify(fmt.Sprintf("%d%%", GetBrightness()))
		return 0
	case "down":
		if err := AdjustBrightness(-10); err != nil {
			showNotify("No backlight")
			return 1
		}
		showNotify(fmt.Sprintf("%d%%", GetBrightness()))
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
		if err := SetBrightness(val); err != nil {
			showNotify("No backlight")
			return 1
		}
		showNotify(fmt.Sprintf("%d%%", val))
		return 0
	case "get":
		fmt.Println(GetBrightness())
		return 0
	default:
		return usage()
	}
}
package display

import (
	"fmt"
	"os/exec"
	"strings"

	"openriot/notify"
	"openriot/polybar"
)

const (
	iconBoth = "󰍺"
	iconHDMI = "󰍹"
)

var lastLidActionState *bool

// RunHDMI outputs the HDMI icon based on current display mode.
// 󰍺 = Laptop + HDMI both active,  󰍹 = HDMI only (laptop disabled).
// Also auto-sets lid suspend: external display → no suspend, no display → allow suspend.
func RunHDMI() {
	if HasExternalDisplay() {
		setLidAction(false) // suspend disabled when docked
		if isLaptopMonitorEnabled() {
			fmt.Println(polybar.Icon(iconBoth))
		} else {
			fmt.Println(polybar.Icon(iconHDMI))
		}
	} else {
		setLidAction(true) // suspend enabled when undocked
		// No icon when no external display
		// Auto-restore laptop display if HDMI was unplugged while in HDMI-only mode
		laptop := getLaptopDisplay()
		if laptop != "" && !isLaptopMonitorEnabled() {
			exec.Command("xrandr", "--output", laptop, "--auto").Run()
		}
	}
}

func setLidAction(enable bool) {
	if lastLidActionState != nil && *lastLidActionState == enable {
		return
	}
	lastLidActionState = &enable
	val := "machdep.lidaction=0"
	if enable {
		val = "machdep.lidaction=1"
	}
	_ = exec.Command("doas", "sysctl", val).Run()
}

// ToggleHDMI toggles between Laptop+HDMI and HDMI-only modes.
// When switching to HDMI-only: laptop off, lid suspend disabled.
// When switching back: laptop on, lid suspend enabled.
func ToggleHDMI() {
	laptop := getLaptopDisplay()
	if laptop == "" {
		notify.SendNotify("hdmi", "Display", "No laptop display found", "critical", 5000, 0)
		return
	}

	if isLaptopMonitorEnabled() {
		// Switch to HDMI-only mode
		exec.Command("xrandr", "--output", laptop, "--off").Run()
		setLidAction(false)
		notify.SendNotify("hdmi", "Display Mode", "HDMI Only 󰍹\nLaptop display disabled", "normal", 8000, 0)
	} else {
		// Switch to both mode
		exec.Command("xrandr", "--output", laptop, "--auto").Run()
		setLidAction(true)
		notify.SendNotify("hdmi", "Display Mode", "Laptop + HDMI 󰍺\nLaptop re-enabled", "normal", 8000, 0)
	}
	// Restart polybar after display reconfiguration
	exec.Command("pkill", "-9", "polybar").Run()
}

// HasExternalDisplay returns true if a non-internal monitor is connected.
func HasExternalDisplay() bool {
	out, err := exec.Command("xrandr").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, " connected ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		if !strings.HasPrefix(name, "eDP") && !strings.HasPrefix(name, "LVDS") {
			return true
		}
	}
	return false
}

// getLaptopDisplay returns the xrandr output name for the built-in display.
// Uses plain xrandr (not --listmonitors) so disabled outputs are still found.
func getLaptopDisplay() string {
	out, err := exec.Command("xrandr").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		state := fields[1]
		if state != "connected" && state != "disconnected" {
			continue
		}
		if strings.HasPrefix(name, "eDP") || strings.HasPrefix(name, "LVDS") {
			return name
		}
	}
	return ""
}

// isLaptopMonitorEnabled returns true if the laptop display is currently active.
func isLaptopMonitorEnabled() bool {
	display := getLaptopDisplay()
	if display == "" {
		return false
	}
	out, err := exec.Command("xrandr", "--listactivemonitors").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), display)
}

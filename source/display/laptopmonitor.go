package display

import (
	"fmt"
	"os/exec"
	"strings"

	"openriot/notify"
	"openriot/polybar"
)

const (
	iconEnabled  = "󰌢"
	iconDisabled = "󰛧"
)

// RunLaptopMonitor outputs the laptop monitor icon for polybar.
// Returns empty string (hidden) if not a laptop or no external monitor.
func RunLaptopMonitor() {
	if !isLaptop() || !hasExternalMonitor() {
		return
	}
	if isLaptopMonitorEnabled() {
		fmt.Println(polybar.Icon(iconEnabled))
	} else {
		fmt.Println(polybar.Icon(iconDisabled))
	}
}

// ToggleLaptopMonitor toggles the laptop display and sends notification.
func ToggleLaptopMonitor() {
	display := getLaptopDisplay()
	if display == "" {
		notify.SendNotify("laptop-monitor", "Laptop Monitor", "No laptop display found", "critical", 5000, 0)
		return
	}

	if isLaptopMonitorEnabled() {
		exec.Command("xrandr", "--output", display, "--off").Run()
		notify.SendNotify("laptop-monitor", "Laptop Monitor", "Disabling Laptop Monitor", "normal", 3000, 0)
	} else {
		exec.Command("xrandr", "--output", display, "--auto").Run()
		notify.SendNotify("laptop-monitor", "Laptop Monitor", "Enabling Laptop Monitor", "normal", 3000, 0)
	}
}

// isLaptop returns true if the system has a battery (laptop indicator).
func isLaptop() bool {
	out, err := exec.Command("sh", "-c", "sysctl hw.sensors acpibat0 2>/dev/null | grep -q present && echo yes || echo no").Output()
	return err == nil && strings.TrimSpace(string(out)) == "yes"
}

// hasExternalMonitor returns true if 2+ monitors are active.
func hasExternalMonitor() bool {
	out, err := exec.Command("xrandr", "--listactivemonitors").Output()
	if err != nil {
		return false
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	// First line is "0: +*DP-1 1920/508x1080/286+0+0  DP-1", count remaining lines
	return len(lines) >= 3
}

// getLaptopDisplay returns the xrandr output name for the built-in display.
func getLaptopDisplay() string {
	out, err := exec.Command("xrandr", "--listmonitors").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		// Last field is the output name
		name := fields[len(fields)-1]
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

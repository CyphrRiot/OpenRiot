package display

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"openriot/notify"
	"openriot/polybar"
)

const (
	iconBoth   = "󰍺"
	iconHDMI   = "󰍹"
	iconLaptop = ""
)

// i3Output mirrors the JSON shape returned by  i3-msg -t get_outputs .
type i3Output struct {
	Name             string `json:"name"`
	Active           bool   `json:"active"`
	Primary          bool   `json:"primary"`
	CurrentWorkspace string `json:"current_workspace"`
}

// i3Outputs is a thin wrapper around parsed i3 get_outputs output.
type i3Outputs []i3Output

func parseI3Outputs() i3Outputs {
	out, err := exec.Command("i3-msg", "-t", "get_outputs").Output()
	if err != nil {
		return nil
	}
	var outs []i3Output
	if err := json.Unmarshal(out, &outs); err != nil {
		return nil
	}
	return outs
}

func (o i3Outputs) laptopName() string {
	for _, out := range o {
		if strings.HasPrefix(out.Name, "eDP") || strings.HasPrefix(out.Name, "LVDS") {
			return out.Name
		}
	}
	return ""
}

func (o i3Outputs) externalName() string {
	for _, out := range o {
		if out.Name == "xroot-0" {
			continue
		}
		if !strings.HasPrefix(out.Name, "eDP") && !strings.HasPrefix(out.Name, "LVDS") {
			return out.Name
		}
	}
	return ""
}

func (o i3Outputs) isActive(name string) bool {
	if name == "" {
		return false
	}
	for _, out := range o {
		if out.Name == name {
			return out.Active
		}
	}
	return false
}

// RunHDMI prints the polybar icon for the current display configuration.
// It uses i3's get_outputs IPC to avoid the xrandr fork+exec and DRM
// round-trip that stalls X11 event delivery every invocation.
func RunHDMI() {
	outputs := parseI3Outputs()
	laptop := outputs.laptopName()
	ext := outputs.externalName()
	hasExt := ext != ""

	if hasExt {
		setLidAction(false) // suspend disabled when docked
		laptopActive := outputs.isActive(laptop)
		extActive := outputs.isActive(ext)

		if laptopActive && !extActive {
			fmt.Println(polybar.Icon(iconLaptop))
		} else if laptopActive && extActive {
			fmt.Println(polybar.Icon(iconBoth))
		} else {
			fmt.Println(polybar.Icon(iconHDMI))
		}
	} else {
		setLidAction(true) // suspend enabled when undocked
		fmt.Println(polybar.Icon(iconLaptop))
		// Auto-restore laptop display if it was disabled while HDMI-only
		if laptop != "" && !outputs.isActive(laptop) {
			exec.Command("xrandr", "--output", laptop, "--auto").Run()
		}
	}
}

func getSysctl(name string) string {
	out, _ := exec.Command("sysctl", "-n", name).Output()
	return strings.TrimSpace(string(out))
}

func setSysctlIfChanged(name, want string) {
	if getSysctl(name) == want {
		return
	}
	_ = exec.Command("doas", "sysctl", name+"="+want).Run()
}

func setLidAction(enable bool) {
	lidWant := "0"
	powerWant := "0"
	perfWant := "auto"
	if enable {
		lidWant = "1"
		powerWant = "1"
	} else if isOnAC() {
		perfWant = "high"
	}
	setSysctlIfChanged("machdep.lidaction", lidWant)
	setSysctlIfChanged("hw.allowpowerdown", powerWant)
	setSysctlIfChanged("hw.perfpolicy", perfWant)
}

func isOnAC() bool {
	out, _ := exec.Command("apm", "-a").Output()
	return strings.TrimSpace(string(out)) == "1"
}

// ToggleHDMI cycles through three display modes: Both → Laptop Only → HDMI Only → Both.
// It still uses xrandr for actual display reconfiguration (that only happens on click).
func ToggleHDMI() {
	outputs := parseI3Outputs()
	laptop := outputs.laptopName()
	if laptop == "" {
		notify.SendNotify("hdmi", "Display", "No laptop display found", "critical", 5000, 0)
		return
	}

	ext := outputs.externalName()
	if ext == "" {
		notify.SendNotify("hdmi", "Display", "No External Monitor", "normal", 5000, 0)
		return
	}

	laptopActive := outputs.isActive(laptop)
	extActive := outputs.isActive(ext)

	if laptopActive && extActive {
		// Both → Laptop only: disable external
		exec.Command("xrandr", "--output", ext, "--off").Run()
		setLidAction(true)
		notify.SendNotify("hdmi", "Display Mode", "Laptop Only \nExternal display disabled", "normal", 8000, 0)
	} else if laptopActive && !extActive {
		// Laptop only → HDMI only: disable laptop, enable external
		exec.Command("xrandr", "--output", laptop, "--off").Run()
		if !extActive {
			exec.Command("xrandr", "--output", ext, "--auto").Run()
		}
		setLidAction(false)
		notify.SendNotify("hdmi", "Display Mode", "HDMI Only 󰍹\nLaptop display disabled", "normal", 8000, 0)
	} else {
		// HDMI only → Both: enable laptop
		exec.Command("xrandr", "--output", laptop, "--auto").Run()
		setLidAction(true)
		notify.SendNotify("hdmi", "Display Mode", "Laptop + HDMI 󰍺\nLaptop re-enabled", "normal", 8000, 0)
	}
	// Restart polybar after display reconfiguration so i3 re-syncs output state
	exec.Command("pkill", "-9", "polybar").Run()
}

// HasExternalDisplay returns true if a non-internal monitor is connected.
func HasExternalDisplay() bool {
	return parseI3Outputs().externalName() != ""
}

// getLaptopDisplay returns the name of the built-in display (eDP-1 / LVDS-1).
func getLaptopDisplay() string {
	return parseI3Outputs().laptopName()
}

// isLaptopMonitorEnabled returns true if the laptop display is currently active.
func isLaptopMonitorEnabled() bool {
	display := getLaptopDisplay()
	if display == "" {
		return false
	}
	return isDisplayActive(display)
}

// getExternalDisplay returns the name of the first connected external display.
func getExternalDisplay() string {
	return parseI3Outputs().externalName()
}

// isDisplayActive returns true if the given display output is currently active.
// Uses i3 get_outputs instead of xrandr for status queries.
func isDisplayActive(display string) bool {
	if display == "" {
		return false
	}
	return parseI3Outputs().isActive(display)
}

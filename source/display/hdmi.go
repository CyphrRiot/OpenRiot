package display

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"openriot/notify"
	"openriot/paths"
	"openriot/polybar"
)

const (
	iconBoth   = "󰍺"
	iconHDMI   = "󰍹"
	iconLaptop = ""
)

var cachedMode = make(map[string]string)
var cachedRate = make(map[string]string)

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

// xrandrLaptopName finds the built-in display name (eDP-1 / LVDS-1) via
// xrandr --query. Used as a fallback when i3's get_outputs doesn't report
// the laptop panel — i3 drops disabled outputs from its tracking after
// `xrandr --output eDP-1 --off`, which breaks the polybar toggle in
// clamshell mode.
func xrandrLaptopName() string {
	out, err := exec.Command("xrandr", "--query").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		if (strings.HasPrefix(name, "eDP") || strings.HasPrefix(name, "LVDS")) && fields[1] == "connected" {
			return name
		}
	}
	return ""
}

func (o i3Outputs) laptopName() string {
	for _, out := range o {
		if strings.HasPrefix(out.Name, "eDP") || strings.HasPrefix(out.Name, "LVDS") {
			return out.Name
		}
	}
	// i3 drops disabled outputs from its tracking; fall back to xrandr
	// so the polybar toggle works in clamshell mode.
	return xrandrLaptopName()
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
		setLidAction(true) // enable suspend when undocked
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

// xrandrOutput holds the parsed state of one xrandr output.
type xrandrOutput struct {
	Name   string
	Active bool   // has geometry like 1920x1080+0+0
	Mode   string // e.g. "1920x1080" (empty if not active)
	Rate   string // e.g. "60.00" (empty if not active)
}

// parseXrandrOutputs parses `xrandr --query` into a map of name → state.
// Authoritative for active state and current mode/rate — i3 drops disabled
// outputs from its tracking after `xrandr --output <name> --off`.
func parseXrandrOutputs() map[string]xrandrOutput {
	out, err := exec.Command("xrandr", "--query").Output()
	if err != nil {
		return nil
	}
	result := make(map[string]xrandrOutput)
	var current string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if current == "" {
				continue
			}
			// Mode line: "   1920x1080     60.00 +* 60.00" (active)
			// or off:    "   1920x1080     60.00 +  60.00" (no *)
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			res := parts[0]
			for _, p := range parts[1:] {
				if strings.Contains(p, "*") {
					s := result[current]
				s.Mode = res
				r := strings.TrimRight(strings.TrimLeft(p, "*"), "+")
				s.Rate = r
				cachedMode[current] = res
				cachedRate[current] = r
				saveDisplayMode(current, res, r)
				result[current] = s
					break
				}
			}
		} else {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[1] == "connected" {
				name := fields[0]
				active := false
				for _, f := range fields[2:] {
					if strings.Contains(f, "x") && strings.Contains(f, "+") {
						active = true
						break
					}
				}
				result[name] = xrandrOutput{Name: name, Active: active}
				current = name
			} else {
				current = ""
			}
		}
	}
	return result
}

// turnOn activates an xrandr output with its previous mode+rate, or
// falls back to --auto if no mode was captured. Mode+rate are persisted
// to a display-mode cache file so they survive across process invocations.
func turnOn(name, mode, rate string) {
	if name == "" {
		return
	}
	if mode == "" {
		if m, ok := cachedMode[name]; ok && m != "" {
			mode = m
			rate = cachedRate[name]
		}
	}
	if mode == "" {
		if fm, fr := loadDisplayMode(name); fm != "" {
			mode = fm
			rate = fr
		}
	}
	if mode == "" {
		exec.Command("xrandr", "--output", name, "--auto").Run()
		return
	}
	args := []string{"--output", name, "--mode", mode}
	if rate != "" {
		args = append(args, "--rate", rate)
	}
	exec.Command("xrandr", args...).Run()
}

// displayModeCachePath returns the path to the per-display mode cache file.
func displayModeCachePath() string {
	return paths.Join(".cache", "openriot", "display-modes")
}

// saveDisplayMode persists a display's current mode+rate to disk so it
// survives process invocations. Offline displays retain their last-known
// values.
func saveDisplayMode(name, mode, rate string) {
	if name == "" || mode == "" {
		return
	}
	cachePath := displayModeCachePath()
	os.MkdirAll(paths.Join(".cache", "openriot"), 0700)

	entries := make(map[string][2]string)
	if data, err := os.ReadFile(cachePath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				entries[fields[0]] = [2]string{fields[1], fields[2]}
			}
		}
	}
	entries[name] = [2]string{mode, rate}

	var lines []string
	for n, mr := range entries {
		lines = append(lines, fmt.Sprintf("%s %s %s", n, mr[0], mr[1]))
	}
	os.WriteFile(cachePath, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

// loadDisplayMode reads a saved mode+rate for a display from disk.
func loadDisplayMode(name string) (mode, rate string) {
	if name == "" {
		return "", ""
	}
	data, err := os.ReadFile(displayModeCachePath())
	if err != nil {
		return "", ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == name {
			return fields[1], fields[2]
		}
	}
	return "", ""
}

// ToggleHDMI cycles through three display modes:
// HDMI Only → Laptop Only → Both → HDMI Only (repeats).
// Uses xrandr as the authoritative source for active state and current
// mode/rate, since i3 drops disabled outputs from its tracking.
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

	states := parseXrandrOutputs()
	laptopOut := states[laptop]
	extOut := states[ext]
	laptopActive := laptopOut.Active
	extActive := extOut.Active

	if laptopActive && extActive {
		// Both → Laptop only: disable external
		exec.Command("xrandr", "--output", ext, "--off").Run()
		setLidAction(true)
		notify.SendNotify("hdmi", "Display Mode", "Laptop Only \nExternal display disabled", "normal", 8000, 0)
	} else if laptopActive && !extActive {
		// Laptop only → HDMI only: disable laptop, enable external
		exec.Command("xrandr", "--output", laptop, "--off").Run()
		turnOn(ext, extOut.Mode, extOut.Rate)
		setLidAction(false)
		notify.SendNotify("hdmi", "Display Mode", "HDMI Only 󰍹\nLaptop display disabled", "normal", 8000, 0)
	} else {
		// HDMI only → Both: enable laptop
		turnOn(laptop, laptopOut.Mode, laptopOut.Rate)
		setLidAction(true)
		notify.SendNotify("hdmi", "Display Mode", "Laptop + HDMI 󰍺\nLaptop re-enabled", "normal", 8000, 0)
	}
	// Restart polybar after display reconfiguration so i3 re-syncs output state
	exec.Command("pkill", "-9", "polybar").Run()
}

// RestoreDisplays re-detects and enables all connected displays after resume
// from suspend. It ensures the laptop display is active and updates lid-action
// sysctls based on whether an external monitor is connected.
func RestoreDisplays() {
	exec.Command("xrandr", "--auto").Run()

	outputs := parseI3Outputs()
	laptop := outputs.laptopName()
	ext := outputs.externalName()

	if ext != "" && laptop != "" && !outputs.isActive(laptop) {
		exec.Command("xrandr", "--output", laptop, "--auto").Run()
	}

	if ext != "" {
		setLidAction(false)
	} else {
		setLidAction(true)
	}

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

package resolution

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"openriot/backgrounds"
	"openriot/lock"
	"openriot/notify"
)

// Display represents a connected monitor output.
type Display struct {
	Name      string
	Primary   bool
	Current   string // e.g. "1920x1080@60.00"
	Modes     []Mode
}

// Mode represents a single resolution with available refresh rates.
type Mode struct {
	Resolution string   // e.g. "1920x1080"
	Rates      []Rate
}

// Rate represents a single refresh rate with metadata flags.
type Rate struct {
	Value     float64
	Current   bool // * flag
	Preferred bool // + flag
}

// String returns a human-readable rate string.
func (r Rate) String() string {
	s := fmt.Sprintf("%.2f", r.Value)
	if r.Current {
		s += " *"
	}
	if r.Preferred {
		s += " +"
	}
	return s
}

// GetDis parses xrandr output and returns all connected displays.
func GetDis() ([]Display, error) {
	out, err := exec.Command("xrandr").Output()
	if err != nil {
		return nil, fmt.Errorf("xrandr failed: %w", err)
	}
	return parseXrandr(string(out)), nil
}

// parseXrandr parses raw xrandr text into a slice of Display structs.
func parseXrandr(text string) []Display {
	var displays []Display
	var currentDisplay *Display

	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t") {
			// Display line
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			if fields[1] != "connected" {
				continue
			}
			d := Display{Name: fields[0]}
			for _, f := range fields[2:] {
				if f == "primary" {
					d.Primary = true
					break
				}
			}
			displays = append(displays, d)
			currentDisplay = &displays[len(displays)-1]
		} else if currentDisplay != nil {
			// Mode line for the current display
			line = strings.TrimSpace(line)
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			res := parts[0]
			var rates []Rate
			for _, p := range parts[1:] {
				if p == "+" {
					// Standalone + applies to the previous rate
					if len(rates) > 0 {
						rates[len(rates)-1].Preferred = true
					}
					continue
				}
				current := strings.Contains(p, "*")
				preferred := strings.Contains(p, "+")
				rateStr := strings.TrimRight(p, "*+")
				val, err := strconv.ParseFloat(rateStr, 64)
				if err != nil {
					continue
				}
				rates = append(rates, Rate{Value: val, Current: current, Preferred: preferred})
				if current {
					currentDisplay.Current = fmt.Sprintf("%s@%.2f", res, val)
				}
			}
			if len(rates) > 0 {
				currentDisplay.Modes = append(currentDisplay.Modes, Mode{Resolution: res, Rates: rates})
			}
		}
	}
	return displays
}

const stateFile = ".config/openriot/resolution.state"

// statePath returns the full path to the resolution state file.
func statePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, stateFile)
}

// SaveResolution writes the selected display, resolution, and rate to disk.
func SaveResolution(displayName, resolution string, rate float64) error {
	file := statePath()
	if file == "" {
		return fmt.Errorf("cannot determine state path")
	}
	data := fmt.Sprintf("%s\n%s\n%.2f\n", displayName, resolution, rate)
	return os.WriteFile(file, []byte(data), 0600)
}

// LoadResolution reads the persisted resolution state.
func LoadResolution() (displayName, resolution string, rate float64, err error) {
	file := statePath()
	if file == "" {
		return "", "", 0, fmt.Errorf("cannot determine state path")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", "", 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 3 {
		return "", "", 0, fmt.Errorf("corrupt state file")
	}
	displayName = lines[0]
	resolution = lines[1]
	if _, err := fmt.Sscanf(lines[2], "%f", &rate); err != nil {
		return "", "", 0, fmt.Errorf("invalid rate in state: %w", err)
	}
	return displayName, resolution, rate, nil
}

// RestoreResolution applies the saved resolution, or returns an error if none is saved.
func RestoreResolution() error {
	displayName, resolution, rate, err := LoadResolution()
	if err != nil {
		return err
	}
	return Apply(displayName, resolution, rate)
}

// Apply executes xrandr, persists the choice, and triggers desktop refresh.
func Apply(displayName, resolution string, rate float64) error {
	cmd := exec.Command("xrandr",
		"--output", displayName,
		"--mode", resolution,
		"--rate", fmt.Sprintf("%.2f", rate),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xrandr --mode failed: %w", err)
	}

	// Persist for next reboot
	_ = SaveResolution(displayName, resolution, rate)

	// Restart polybar so it scales to the new resolution
	exec.Command("pkill", "-9", "polybar").Run()

	// Reload wallpaper at new resolution
	backgrounds.Load()

	// Rebuild lock screen cache asynchronously (can take 30-60s)
	go func() {
		notify.SendNotify("lock", "Screen Lock", "Rebuilding lock screen cache... (~1 min)", "normal", 5000, 0)
		if err := lock.BuildCache(); err != nil {
			notify.SendNotify("lock", "Screen Lock", fmt.Sprintf("Cache rebuild failed: %v", err), "critical", 5000, 0)
		} else {
			notify.SendNotify("lock", "Screen Lock", "Lock screen cache rebuilt", "normal", 3000, 0)
		}
	}()

	return nil
}

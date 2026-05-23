package settings

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"openriot/macspoof"
	"openriot/nightlight"
	"openriot/notify"
	"openriot/paths"
	"openriot/polybar"
	"openriot/theme"
	"openriot/wireguard"
)

// RunMenu launches a rofi settings menu with toggles for VPN, Transmission,
// Night Light, and Proton Sync. Each entry shows its current ON/OFF state.
func RunMenu() {
	entries := buildEntries()
	if len(entries) == 0 {
		return
	}

	themePath := findTheme()

	var rofiInput bytes.Buffer
	for _, e := range entries {
		stateStr := "(Turn on)"
		if e.on {
			stateStr = "(Turn off)"
		}
		if e.label != "" {
			stateStr = e.label
		}
		fmt.Fprintf(&rofiInput, "%s %s %s\n", e.icon, e.name, stateStr)
	}

	cmd := exec.Command("rofi", "-dmenu", "-i", "-p", "Settings", "-format", "i", "-theme", themePath, "-theme-str", fmt.Sprintf("window { width: 450px; border: 2px; border-color: %s; }", theme.GetAccent()))
	cmd.Stdin = &rofiInput
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return // user cancelled
	}

	selected := strings.TrimSpace(out.String())
	if selected == "" {
		return
	}

	var idx int
	if _, err := fmt.Sscanf(selected, "%d", &idx); err != nil {
		return
	}
	if idx < 0 || idx >= len(entries) {
		return
	}

	entries[idx].toggle()
}

type entry struct {
	icon   string
	name   string
	label  string
	on     bool
	toggle func()
}

func buildEntries() []entry {
	var entries []entry

	// WireGuard
	wgOn := wireguard.IsRunning()
	entries = append(entries, entry{
		icon: "󰱓",
		name: "WireGuard VPN",
		on:   wgOn,
		toggle: func() {
			_ = wireguard.Toggle()
		},
	})

	// Night Light
	nlOn := nightlight.IsOn()
	entries = append(entries, entry{
		icon: "󰌵",
		name: "Night Light",
		on:   nlOn,
		toggle: func() {
			_ = nightlight.Toggle()
		},
	})

	// Monitor Resolution
	entries = append(entries, entry{
		icon:  "󰹑",
		name:  "Monitor Resolution",
		label: "(Change)",
		toggle: func() {
			exec.Command("alacritty", "--class", "openriot_resolution", "-e", "openriot", "--resolution-tui").Start()
		},
	})

	// Proton Sync
	entries = append(entries, entry{
		icon:  "󱥾",
		name:  "Proton Sync",
		label: "(Sync)",
		on:    polybar.IsProtonDriveConfigured(),
		toggle: func() {
			_ = polybar.TriggerSync()
		},
	})

	// Stealth Mode
	entries = append(entries, entry{
		icon: "󰝴",
		name: "Stealth Mode",
		on:   macspoof.IsStealthEnabled(),
		toggle: func() {
			notify.SendNotify("stealth", "Stealth", "Restarting Networking Services", "normal", 5000, 0)
			if err := macspoof.StealthToggle(); err != nil {
				notify.SendNotify("stealth", "Stealth", "Failed: "+err.Error(), "critical", 5000, 0)
				return
			}
			enabled := macspoof.IsStealthEnabled()
			if enabled {
				notify.SendNotify("stealth", "Stealth", "Enabled [Stealth]", "normal", 3000, 0)
			} else {
				notify.SendNotify("stealth", "Stealth", "Disabled", "normal", 3000, 0)
			}
		},
	})

	// Select WiFi
	entries = append(entries, entry{
		icon:  "󱚹",
		name:  "Select WiFi",
		label: "(Launch)",
		on:    false,
		toggle: func() {
			exec.Command("alacritty", "--class", "openriot_wifi", "-e", "openriot", "--nmtui").Start()
		},
	})

	// Benchmark
	entries = append(entries, entry{
		icon:  "󰓅",
		name:  "Benchmark",
		label: "(Launch)",
		on:    false,
		toggle: func() {
			exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "openriot", "--benchmark").Start()
		},
	})

	return entries
}

func findTheme() string {
	var candidates []string
	candidates = append(candidates, paths.OpenRiotDir("config", "rofi", "simple-tokyonight.rasi"))
	candidates = append(candidates, "/usr/local/share/openriot/config/rofi/simple-tokyonight.rasi")
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "simple-tokyonight"
}

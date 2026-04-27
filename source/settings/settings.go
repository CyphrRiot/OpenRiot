package settings

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/macspoof"
	"openriot/nightlight"
	"openriot/polybar"
	"openriot/wireguard"
)

var homeDir, _ = os.UserHomeDir()

// RunMenu launches a rofi settings menu with toggles for VPN, Transmission,
// Night Light, and Proton Sync. Each entry shows its current ON/OFF state.
func RunMenu() {
	entries := buildEntries()
	if len(entries) == 0 {
		return
	}

	theme := findTheme()

	var rofiInput bytes.Buffer
	for _, e := range entries {
		stateStr := "(Turn on)"
		if e.on {
			stateStr = "(Turn off)"
		}
		fmt.Fprintf(&rofiInput, "%s %s %s\n", e.icon, e.name, stateStr)
	}

	cmd := exec.Command("rofi", "-dmenu", "-i", "-p", "Settings", "-format", "i", "-theme", theme, "-theme-str", "window { width: 450px; }")
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
	on     bool
	toggle func()
}

func buildEntries() []entry {
	var entries []entry

	// WireGuard
	wgOn := wireguard.Status() != ""
	entries = append(entries, entry{
		icon: "󰱓",
		name: "WireGuard VPN",
		on:   wgOn,
		toggle: func() {
			exec.Command("sh", "-c", "$HOME/.local/share/openriot/install/openriot --wireguard").Start()
		},
	})

	// Night Light
	nlOn := nightlight.Get() != ""
	entries = append(entries, entry{
		icon: "󰌵",
		name: "Night Light",
		on:   nlOn,
		toggle: func() {
			exec.Command("sh", "-c", "$HOME/.local/share/openriot/install/openriot --night-light").Start()
		},
	})

	// Proton Sync
	entries = append(entries, entry{
		icon: "󱥾",
		name: "Proton Sync",
		on:   polybar.IsProtonDriveConfigured(),
		toggle: func() {
			exec.Command("sh", "-c", "$HOME/.local/share/openriot/install/openriot --proton-drive-sync").Start()
		},
	})

	// Stealth Mode
	entries = append(entries, entry{
		icon: "󰝴",
		name: "Stealth Mode",
		on:   macspoof.IsStealthEnabled(),
		toggle: func() {
			exec.Command("sh", "-c", "$HOME/.local/share/openriot/install/openriot --stealth").Start()
		},
	})

	return entries
}

func findTheme() string {
	candidates := []string{
		filepath.Join(homeDir, ".local/share/openriot/config/rofi/simple-tokyonight.rasi"),
		"/home/grendel/.local/share/openriot/config/rofi/simple-tokyonight.rasi",
		"/usr/local/share/openriot/config/rofi/simple-tokyonight.rasi",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "simple-tokyonight"
}

package wireguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/network"
	"openriot/notify"
	"openriot/polybar"
)

const (
	ConfigPath = "/etc/wireguard/wg0.conf"
	stateFile  = ".config/openriot/wireguard.enabled"
)

func isConfigured() bool {
	cmd := exec.Command("doas", "test", "-f", ConfigPath)
	err := cmd.Run()
	return err == nil
}

func IsRunning() bool {
	cmd := exec.Command("ifconfig", "wg0")
	out, _ := cmd.Output()
	return strings.Contains(string(out), "UP") && strings.Contains(string(out), "RUNNING")
}

func isRunning() bool {
	return IsRunning()
}

// GetTunnelIP returns the IPv4 address assigned to wg0, or empty string if down.
func GetTunnelIP() string {
	cmd := exec.Command("ifconfig", "wg0")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") && !strings.HasPrefix(line, "inet6 ") {
			// Format: "inet 10.75.64.36 netmask 0xffffff00"
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}
	return ""
}

// GetTunnelIP returns the IPv4 address assigned to wg0, or empty string if down.

func isAutostartEnabled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, stateFile))
	return err == nil
}

func setAutostart(enabled bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path := filepath.Join(home, stateFile)
	if enabled {
		os.WriteFile(path, []byte("1"), 0600)
	} else {
		os.Remove(path)
	}
}

const rcLocalMarker = "# OpenRiot: wireguard autostart"
const rcLocalCmd = "[ -f /etc/wireguard/wg0.conf ] && wg-quick up /etc/wireguard/wg0.conf 2>/dev/null"

func setBootPersistence(enabled bool) {
	data, err := os.ReadFile("/etc/rc.local")
	exists := err == nil
	lines := strings.Split(string(data), "\n")

	var out []string
	for _, line := range lines {
		if strings.Contains(line, rcLocalMarker) {
			continue
		}
		out = append(out, line)
	}

	if enabled {
		out = append(out, rcLocalMarker, rcLocalCmd)
	}

	if !exists && !enabled {
		return
	}

	newContent := strings.Join(out, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	cmd := exec.Command("doas", "tee", "/etc/rc.local")
	cmd.Stdin = strings.NewReader(newContent)
	_ = cmd.Run()
}

func Status() string {
	if !isConfigured() {
		return ""
	}
	if isRunning() {
		return polybar.Icon("󰱓")
	}
	return polybar.Icon("󰅛")
}

func Start() error {
	notify.SendNotify("wireguard", "VPN", "Starting WireGuard...", "normal", 3000, 0)
	cmd := exec.Command("doas", "wg-quick", "up", ConfigPath)
	if err := cmd.Run(); err != nil {
		return err
	}
	// wg-quick up kills WiFi; restore it
	go network.ReconnectWifi()
	return nil
}

func Stop() error {
	notify.SendNotify("wireguard", "VPN", "Stopping WireGuard...", "normal", 3000, 0)
	cmd := exec.Command("doas", "wg-quick", "down", ConfigPath)
	return cmd.Run()
}

func Toggle() error {
	if !isConfigured() {
		notify.SendNotify("wireguard", "VPN", "Not configured\nGo to OpenRiot.org\nRead directions.", "critical", 5000, 0)
		return nil
	}
	if isRunning() {
		setAutostart(false)
		setBootPersistence(false)
		return Stop()
	}
	setAutostart(true)
	setBootPersistence(true)
	return Start()
}

func AutoStart() error {
	if !isConfigured() || !isAutostartEnabled() {
		return nil
	}
	return Start()
}

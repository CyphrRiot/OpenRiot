package wireguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ConfigPath = "/etc/wireguard/wg0.conf"
)

func isConfigured() bool {
	cmd := exec.Command("doas", "test", "-f", ConfigPath)
	err := cmd.Run()
	return err == nil
}

func isRunning() bool {
	cmd := exec.Command("ifconfig", "wg0")
	out, _ := cmd.Output()
	return strings.Contains(string(out), "UP") && strings.Contains(string(out), "RUNNING")
}

func Status() string {
	if !isConfigured() {
		return "󰛳"
	}
	if isRunning() {
		return "󰱓"
	}
	return "󰅛"
}

func getHome() string {
	home, _ := os.UserHomeDir()
	return home
}

func Start() error {
	home := getHome()
	iconPath := filepath.Join(home, ".local/share/openriot/config/icons")
	vpnIcon := filepath.Join(iconPath, "vpn.png")
	exec.Command("/usr/local/bin/notify-send", "-i", vpnIcon, "-u", "normal", "VPN", "Starting WireGuard...").Run()
	cmd := exec.Command("doas", "wg-quick", "up", ConfigPath)
	return cmd.Run()
}

func Stop() error {
	home := getHome()
	iconPath := filepath.Join(home, ".local/share/openriot/config/icons")
	vpnIcon := filepath.Join(iconPath, "vpn.png")
	exec.Command("/usr/local/bin/notify-send", "-i", vpnIcon, "-u", "normal", "VPN", "Stopping WireGuard...").Run()
	cmd := exec.Command("doas", "wg-quick", "down", ConfigPath)
	return cmd.Run()
}

func Toggle() error {
	if !isConfigured() {
		home := getHome()
		iconPath := filepath.Join(home, ".local/share/openriot/config/icons")
		vpnErrIcon := filepath.Join(iconPath, "vpn-error.png")
		exec.Command("/usr/local/bin/notify-send", "-i", vpnErrIcon, "-u", "critical", "VPN", "Not configured\nGo to OpenRiot.org\nRead directions.").Run()
		return nil
	}
	if isRunning() {
		return Stop()
	}
	return Start()
}

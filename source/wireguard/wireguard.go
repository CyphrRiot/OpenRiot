package wireguard

import (
	"os/exec"
	"strings"

	"openriot/notify"
	"openriot/polybar"
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
	if !isConfigured() || !isRunning() {
		return ""
	}
	return polybar.Icon("󰱓")
}

func Start() error {
	notify.SendNotify("wireguard", "VPN", "Starting WireGuard...", "normal", 3000, 0)
	cmd := exec.Command("doas", "wg-quick", "up", ConfigPath)
	return cmd.Run()
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
		return Stop()
	}
	return Start()
}

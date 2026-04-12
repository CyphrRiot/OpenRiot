package wireguard

import (
	"os"
	"os/exec"
	"strings"
)

const (
	IconNotConfigured = "󰛳"
	IconDisconnected  = "󰅛"
	IconConnected     = "󰱓"
	ConfigPath        = "/etc/wireguard/wg0.conf"
)

func isConfigured() bool {
	_, err := os.Stat(ConfigPath)
	return err == nil
}

func isRunning() bool {
	cmd := exec.Command("ifconfig", "wg0")
	out, _ := cmd.Output()
	return strings.Contains(string(out), "UP") && strings.Contains(string(out), "RUNNING")
}

func Status() string {
	if !isConfigured() {
		return IconNotConfigured
	}
	if isRunning() {
		return IconConnected
	}
	return IconDisconnected
}

func Start() error {
	exec.Command("notify-send", "-u", "normal", "Starting WireGuard...").Run()
	cmd := exec.Command("doas", "wg-quick", "up", ConfigPath)
	return cmd.Run()
}

func Stop() error {
	exec.Command("notify-send", "-u", "normal", "Stopping WireGuard...").Run()
	cmd := exec.Command("doas", "wg-quick", "down", ConfigPath)
	return cmd.Run()
}

func Toggle() error {
	if !isConfigured() {
		exec.Command("notify-send", "-u", "critical", "WireGuard is not configured.\nGo to OpenRiot.org\nRead directions.").Run()
		return nil
	}
	if isRunning() {
		return Stop()
	}
	return Start()
}

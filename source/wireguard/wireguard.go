package wireguard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"openriot/notify"
	"openriot/paths"
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
	_, err := os.Stat(paths.Join(stateFile))
	return err == nil
}

func setAutostart(enabled bool) {
	path := paths.Join(stateFile)
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
		if strings.Contains(line, rcLocalMarker) || strings.Contains(line, rcLocalCmd) {
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

func getDNSFromConfig() string {
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "DNS") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

type mullvadStatus struct {
	Hostname string `json:"mullvad_exit_ip_hostname"`
	City     string `json:"city"`
	ExitIP   bool   `json:"mullvad_exit_ip"`
}

func fetchMullvadStatus() (mullvadStatus, error) {
	var status mullvadStatus
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://am.i.mullvad.net/json")
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return status, err
	}
	return status, nil
}

func IsConnected() bool {
	status, err := fetchMullvadStatus()
	return err == nil && status.ExitIP
}

func GetServerName() string {
	status, err := fetchMullvadStatus()
	if err != nil || status.Hostname == "" {
		return ""
	}
	if status.City != "" {
		return status.Hostname + "\n" + status.City
	}
	return status.Hostname
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
	if dns := getDNSFromConfig(); dns != "" {
		exec.Command("doas", "sh", "-c",
			fmt.Sprintf("echo 'nameserver %s' >> /etc/resolv.conf", dns)).Run()
		exec.Command("doas", "rcctl", "restart", "resolvd").Run()
	}
	time.Sleep(2 * time.Second)
	if !IsConnected() {
		notify.SendNotify("wireguard", "WireGuard VPN",
			"Failed to connect. Mullvad account may be expired or out of credits.", "critical", 0, 0)
		return nil
	}
	if server := GetServerName(); server != "" {
		notify.SendNotify("wireguard", "WireGuard VPN",
			fmt.Sprintf("Connected to %s", server), "normal", 5000, 0)
	}
	return nil
}

func Stop() error {
	notify.SendNotify("wireguard", "VPN", "Stopping WireGuard...", "normal", 3000, 0)
	server := GetServerName()
	cmd := exec.Command("doas", "wg-quick", "down", ConfigPath)
	if err := cmd.Run(); err != nil {
		return err
	}
	exec.Command("doas", "rcctl", "restart", "resolvd").Run()
	if server != "" {
		notify.SendNotify("wireguard", "WireGuard VPN",
			fmt.Sprintf("Disconnected from %s", server), "normal", 5000, 0)
	}
	return nil
}

func Restart() error {
	if isRunning() {
		_ = Stop()
		time.Sleep(500 * time.Millisecond)
	}
	return Start()
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

package network

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const connectivityFile = "network-online"

func Get() string {
	return GetWifi()
}

func GetWifi() string {
	// Check connectivity on every call
	CheckConnectivity()

	iface, connected := getWifiInterface()
	if !connected {
		// Use WasOnline for hysteresis - don't flash "down" for brief hiccups
		if WasOnline() {
			signal := getSignal(iface)
			if signal > 0 {
				return getWifiIcon(signal)
			}
		}
		return ""
	}

	// Connected but no internet - use WasOnline for hysteresis
	if !IsOnline() {
		if WasOnline() {
			// Still show signal, internet will come back
			signal := getSignal(iface)
			if signal > 0 {
				return getWifiIcon(signal)
			}
		}
		return "󱛅"
	}

	signal := getSignal(iface)
	if signal > 0 {
		return getWifiIcon(signal)
	}
	return ""
}

func GetEth() string {
	iface, hasCarrier := getEthInterface()
	if iface == "" {
		return ""
	}

	if hasCarrier {
		return "󰈀"
	}
	return "󰌙"
}

func GetWifiDetails() string {
	iface, connected := getWifiInterface()
	if !connected {
		return fmt.Sprintf("Not Connected\nInterface: %s\n\n1. Update /etc/hostname.%s\n2. Run    doas /etc/netstart %s\n3. Pray to the ancient gods of BSD", iface, iface, iface)
	}

	cmd := exec.Command("/sbin/ifconfig", iface)
	output, _ := cmd.Output()
	details := string(output)

	ap := extractAP(details)
	ip := extractIP(details)
	mac := extractMAC(details)
	randomized := isMACRandomized(iface)

	macInfo := fmt.Sprintf("MAC: %s", mac)
	if randomized {
		macInfo += " [Stealth]"
	}

	return fmt.Sprintf("AP: %s\nIP: %s\nInterface: %s\n%s", ap, ip, iface, macInfo)
}

func GetEthDetails() string {
	iface, hasCarrier := getEthInterface()
	if iface == "" {
		return "No Ethernet interface"
	}

	if !hasCarrier {
		return fmt.Sprintf("No Carrier\nInterface: %s\n\nCheck cable connection", iface)
	}

	cmd := exec.Command("/sbin/ifconfig", iface)
	output, _ := cmd.Output()
	ip := extractIP(string(output))
	mac := extractMAC(string(output))
	randomized := isMACRandomized(iface)

	macInfo := fmt.Sprintf("MAC: %s", mac)
	if randomized {
		macInfo += " [Stealth]"
	}

	return fmt.Sprintf("IP: %s\nInterface: %s\n%s", ip, iface, macInfo)
}

// extractMAC extracts the MAC address from ifconfig output
func extractMAC(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "lladdr") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "lladdr" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}
	return "N/A"
}

// isMACRandomized checks if lladdr random is configured in hostname file
func isMACRandomized(iface string) bool {
	content, err := os.ReadFile("/etc/hostname." + iface)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), "lladdr random")
}

// IsConnected returns true if wifi is connected
func IsConnected() bool {
	_, connected := getWifiInterface()
	return connected
}

func getWifiInterface() (string, bool) {
	cmd := exec.Command("/sbin/ifconfig")
	output, _ := cmd.Output()

	var current string
	isWifi := false
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if matched, _ := regexp.MatchString(`^[a-z]+[0-9]+:`, line); matched {
			current = strings.TrimSuffix(strings.SplitN(line, ":", 2)[0], ":")
			isWifi = false
		}
		if strings.Contains(line, "ieee80211") && current != "ieee80211" {
			isWifi = true
		}
		// Connected if: join (WPA) or inet (DHCP/static IP assigned)
		if isWifi && (strings.Contains(line, "join") || strings.Contains(line, "inet ")) {
			return current, true
		}
	}
	return "", false
}

func getEthInterface() (string, bool) {
	cmd := exec.Command("/sbin/ifconfig")
	output, _ := cmd.Output()

	var current string
	var isEth bool
	hasCarrier := false
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if matched, _ := regexp.MatchString(`^[a-z]+[0-9]+:`, line); matched {
			current = strings.TrimSuffix(strings.SplitN(line, ":", 2)[0], ":")
			isEth = false
			hasCarrier = false
		}
		if strings.Contains(line, "media: Ethernet") {
			isEth = true
		}
		if isEth && strings.Contains(line, "status: active") {
			return current, true
		}
		// Only check fallback on status: lines AFTER "status: active" check
		// This prevents returning false on media: line before we see status:
		if isEth && current != "" && current != "lo0" && current != "ieee80211" && strings.Contains(line, "status:") {
			// Connected if: status active OR inet (IP assigned)
			hasCarrier = hasCarrier || strings.Contains(line, "inet ")
			return current, hasCarrier
		}
	}
	return "", false
}

func getSignal(iface string) int {
	cmd := exec.Command("/sbin/ifconfig", iface)
	output, _ := cmd.Output()

	re := strings.Fields(string(output))
	for _, s := range re {
		if strings.HasPrefix(s, "-") && strings.HasSuffix(s, "dBm") {
			var signal int
			fmt.Sscanf(s, "%ddBm", &signal)
			return -signal
		}
	}
	return 0
}

func getWifiIcon(signal int) string {
	percent := max(0, min((signal+100)*100/70, 100))

	if percent >= 70 {
		return "󰤨"
	} else if percent >= 50 {
		return "󰤥"
	} else if percent >= 30 {
		return "󰤢"
	} else if percent >= 20 {
		return "󰤟"
	}
	return "󰤯"
}

func extractAP(output string) string {
	if !strings.Contains(output, "join") {
		return "N/A"
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "join") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "join" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}
	return "Unknown"
}

func extractIP(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet ") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "inet" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}
	return "N/A"
}

// connectivityPath returns the path to the connectivity state file
func connectivityPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache/openriot", connectivityFile)
}

// IsOnline returns true if we have confirmed internet connectivity
func IsOnline() bool {
	path := connectivityPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	// File contains timestamp of last successful ping
	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	// Consider online if ping was within last 90 seconds
	return time.Since(timestamp) < 90*time.Second
}

// WasOnline returns true if we were online recently (5 min grace period)
// This provides hysteresis - don't show "down" immediately after connectivity loss
func WasOnline() bool {
	path := connectivityPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	// Grace period: still show "online" for 5 minutes after last ping
	return time.Since(timestamp) < 5*time.Minute
}

// CheckConnectivity pings 8.8.8.8 and updates the connectivity state
func CheckConnectivity() {
	_, connected := getWifiInterface()
	if !connected {
		// Not connected at all - clear connectivity
		os.Remove(connectivityPath())
		return
	}

	// Ping with 1 packet, timeout wrapper for 3 second timeout
	cmd := exec.Command("sh", "-c", "timeout 3 ping -c 1 8.8.8.8 >/dev/null 2>&1")
	err := cmd.Run()

	if err == nil {
		// Success - record timestamp
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, ".cache/openriot")
		os.MkdirAll(dir, 0755)
		os.WriteFile(connectivityPath(), []byte(time.Now().Format(time.RFC3339)), 0600)
	} else {
		// Failed - clear connectivity file
		os.Remove(connectivityPath())
	}
}

// ReconnectWifi restarts the wifi interface to reconnect
func ReconnectWifi() error {
	iface, connected := getWifiInterface()
	if !connected || iface == "" {
		return fmt.Errorf("no wifi interface found")
	}
	// Run netstart for the interface
	cmd := exec.Command("doas", "sh", "/etc/netstart", iface)
	return cmd.Run()
}
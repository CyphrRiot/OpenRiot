package network

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"openriot/logger"
	"openriot/paths"
	"openriot/polybar"
)

const connectivityFile = "network-online"

var ifaceLineRE = regexp.MustCompile(`^[a-z]+[0-9]+:`)

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
		return polybar.Icon("󱛅")
	}

	signal := getSignal(iface)
	return getWifiIcon(signal)
}

func GetEth() string {
	iface, hasCarrier := getEthInterface()
	if iface == "" {
		return ""
	}
	if hasCarrier {
		return polybar.Icon("󰌘")
	}
	return polybar.Icon("󰈀")
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

	lines := strings.Split(string(output), "\n")
	var current string
	isWifi := false
	isUp := false
	hasJoin := false
	noNetwork := false

	for _, line := range lines {
		// New interface block starts
		if ifaceLineRE.MatchString(line) {
			// Check previous interface first
			if current != "" && isWifi && isUp && hasJoin && !noNetwork {
				return current, true
			}
			current = strings.TrimSuffix(strings.SplitN(line, ":", 2)[0], ":")
			isWifi = false
			isUp = strings.Contains(line, "<UP")
			hasJoin = false
			noNetwork = false
			continue
		}
		if current == "" {
			continue
		}
		if strings.Contains(line, "ieee80211") && current != "ieee80211" {
			isWifi = true
		}
		if isWifi {
			if strings.Contains(line, "join") || strings.Contains(line, "nwid") {
				hasJoin = true
			}
			if strings.Contains(line, "status: no network") {
				noNetwork = true
			}
		}
	}

	// Check the last interface in the output
	if current != "" && isWifi && isUp && hasJoin && !noNetwork {
		return current, true
	}
	return "", false
}

func getEthInterface() (string, bool) {
	cmd := exec.Command("/sbin/ifconfig")
	output, _ := cmd.Output()

	var current string
	var isEth bool
	var lastEth string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if ifaceLineRE.MatchString(line) {
			current = strings.TrimSuffix(strings.SplitN(line, ":", 2)[0], ":")
			isEth = false
		}
		if strings.Contains(line, "media: Ethernet") {
			isEth = true
			lastEth = current
		}
		if isEth && strings.Contains(line, "status: active") {
			return current, true
		}
	}
	if lastEth != "" {
		return lastEth, false
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
		// Percentage format: 56%
		if strings.HasSuffix(s, "%") {
			var pct int
			if n, _ := fmt.Sscanf(s, "%d%%", &pct); n == 1 {
				return pct
			}
		}
	}
	return 0
}

func getWifiIcon(signal int) string {
	var percent int
	if signal >= 0 && signal <= 100 {
		percent = signal
	} else {
		percent = max(0, min((signal+100)*100/70, 100))
	}

	var icon string
	if percent >= 70 {
		icon = "󰤨"
	} else if percent >= 50 {
		icon = "󰤥"
	} else if percent >= 30 {
		icon = "󰤢"
	} else if percent >= 20 {
		icon = "󰤟"
	} else {
		icon = "󰤯"
	}
	return polybar.Icon(icon)
}

func extractAP(output string) string {
	if !strings.Contains(output, "join") && !strings.Contains(output, "nwid") {
		return "N/A"
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		for _, kw := range []string{"join", "nwid"} {
			if !strings.Contains(line, kw) {
				continue
			}
			parts := strings.Fields(line)
			for i, p := range parts {
				if p != kw || i+1 >= len(parts) {
					continue
				}
				word := parts[i+1]
				if strings.HasPrefix(word, "\"") {
					var b strings.Builder
					b.WriteString(word)
					for j := i + 2; j < len(parts); j++ {
						b.WriteByte(' ')
						b.WriteString(parts[j])
						if strings.HasSuffix(parts[j], "\"") {
							break
						}
					}
					return strings.Trim(b.String(), "\"")
				}
				return word
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
func connectivityPath() (string, error) {
	home := paths.HomeDir()
	if home == "" {
		return "", fmt.Errorf("cannot get home dir")
	}
	return filepath.Join(home, ".cache", "openriot", connectivityFile), nil
}

// IsOnline returns true if we have confirmed internet connectivity
func IsOnline() bool {
	path, err := connectivityPath()
	if err != nil {
		return false
	}
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
	path, err := connectivityPath()
	if err != nil {
		return false
	}
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
		if path, err := connectivityPath(); err == nil {
			os.Remove(path)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "8.8.8.8")
	err := cmd.Run()

	path, err2 := connectivityPath()
	if err2 != nil {
		return
	}
	if err == nil {
		// Success - record timestamp
		if mkdirErr := os.MkdirAll(filepath.Dir(path), 0700); mkdirErr != nil {
			logger.Warn(fmt.Sprintf("connectivity: mkdir: %v", mkdirErr))
		}
		if writeErr := os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0600); writeErr != nil {
			logger.Warn(fmt.Sprintf("connectivity: write: %v", writeErr))
		}
	} else {
		// Failed - clear connectivity file
		if removeErr := os.Remove(path); removeErr != nil {
			logger.Warn(fmt.Sprintf("connectivity: remove: %v", removeErr))
		}
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
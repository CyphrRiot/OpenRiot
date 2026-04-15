package network

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func Get() string {
	return GetWifi()
}

func GetWifi() string {
	iface, connected := getWifiInterface()
	if !connected {
		return "󰤯"
	}

	signal := getSignal(iface)
	if signal > 0 {
		return getWifiIcon(signal)
	}
	return "󰤯"
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

	return fmt.Sprintf("AP: %s\nIP: %s\nInterface: %s", ap, ip, iface)
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

	return fmt.Sprintf("IP: %s\nInterface: %s", ip, iface)
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
			hasCarrier = true
		}
		if isEth && current != "" && current != "lo0" && current != "ieee80211" {
			// Connected if: status active OR inet (IP assigned)
			if strings.Contains(line, "media:") || strings.Contains(line, "status:") || strings.Contains(line, "inet ") {
				hasCarrier = hasCarrier || strings.Contains(line, "inet ")
				return current, hasCarrier
			}
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
	percent := (signal + 100) * 100 / 70
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

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
package network

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func Get() string {
	iface := getWifiInterface()
	if iface == "" {
		return "󰤯"
	}

	signal := getSignal(iface)
	if signal == 0 {
		return "󰤯"
	}

	return getWifiIcon(signal)
}

func GetDetails() string {
	iface := getWifiInterface()
	if iface == "" {
		return "No interface found"
	}

	cmd := exec.Command("/sbin/ifconfig", iface)
	output, _ := cmd.Output()
	outputStr := string(output)

	if !strings.Contains(outputStr, "status: active") {
		return "Not connected"
	}

	// Get AP name (join the network)
	ap := ""
	if strings.Contains(outputStr, "join") {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "join") {
				parts := strings.Fields(line)
				for i, p := range parts {
					if p == "join" && i+1 < len(parts) {
						ap = parts[i+1]
						break
					}
				}
				break
			}
		}
	}
	if ap == "" {
		ap = "Unknown"
	}

	// Get IP address
	ip := ""
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet ") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "inet" && i+1 < len(parts) {
					ip = parts[i+1]
					break
				}
			}
			break
		}
	}
	if ip == "" {
		ip = "N/A"
	}

	return fmt.Sprintf("AP: %s\nIP: %s\nNC: %s", ap, ip, iface)
}

func getWifiInterface() string {
	cmd := exec.Command("/sbin/ifconfig")
	output, _ := cmd.Output()

	var current string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if matched, _ := regexp.MatchString(`^[a-z]+[0-9]+:`, line); matched {
			current = strings.TrimSuffix(strings.SplitN(line, ":", 2)[0], ":")
		}
		if strings.Contains(line, "ieee80211") && current != "" {
			return current
		}
	}
	return ""
}

func getSignal(iface string) int {
	cmd := exec.Command("/sbin/ifconfig", iface)
	output, _ := cmd.Output()

	re := regexp.MustCompile(`-[0-9]+dBm`)
	match := re.FindString(string(output))
	if match == "" {
		return 0
	}

	signalStr := strings.Trim(match, "-dBm")
	signal, err := strconv.Atoi(signalStr)
	if err != nil {
		return 0
	}
	return signal
}

func getWifiIcon(signal int) string {
	// Convert dBm to percentage (-100dBm = 0%, -30dBm = 100%)
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

package polybar

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// RunMetrics outputs CPU icon for polybar (memory is separate module)
func RunMetrics() error {
	cpu := getCPU()
	cpuPct := getCPUPercent()
	fmt.Printf(" %s\nCPU: %s%%\n", cpu, cpuPct)
	return nil
}

func getCPU() string {
	cpu := getCPUPercentInt()
	// 0-25%: 󰡳, 25-50%: 󰡵, 50-90%: 󰊚, 90%+: 󰡴
	switch {
	case cpu >= 90:
		return "󰡴"
	case cpu >= 50:
		return "󰊚"
	case cpu >= 25:
		return "󰡵"
	default:
		return "󰡳"
	}
}

func getCPUPercent() string {
	return strconv.Itoa(getCPUPercentInt())
}

func getCPUPercentInt() int {
	out, err := exec.Command("/usr/bin/vmstat", "1", "1").Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0
	}
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) > 0 {
		idle, err := strconv.ParseFloat(fields[len(fields)-1], 64)
		if err != nil {
			return 0
		}
		return int(100 - idle)
	}
	return 0
}

// GetRAM returns memory icon (for polybar)
func GetRAM() string {
	ram := getRAMPercentInt()
	// 0-25%: 󱊔, 25-50%: 󱊗, 50-90%: 󱊖, 90%+: 󱊕
	switch {
	case ram >= 90:
		return "󱊕"
	case ram >= 50:
		return "󱊖"
	case ram >= 25:
		return "󱊗"
	default:
		return "󱊔"
	}
}

func getRAM() string {
	ram := getRAMPercentInt()
	// 0-25%: 󱊔, 25-50%: 󱊗, 50-90%: 󱊖, 90%+: 󱊕
	switch {
	case ram >= 90:
		return "󱊕"
	case ram >= 50:
		return "󱊖"
	case ram >= 25:
		return "󱊗"
	default:
		return "󱊔"
	}
}

func getRAMPercent() string {
	return strconv.Itoa(getRAMPercentInt())
}

// GetCPUPercent returns CPU usage as "XX%" string (for notifications)
func GetCPUPercent() string {
	return getCPUPercent() + "%"
}

// GetMemPercent returns memory usage as "XX%" string (for notifications)
func GetMemPercent() string {
	return getRAMPercent() + "%"
}

// GetMemDetails returns detailed memory info for notifications
// Format: "X.XX GiB of Y.YY GiB\nTotal Used: Z%"
func GetMemDetails() string {
	totalOut, _ := exec.Command("/sbin/sysctl", "-n", "hw.usermem").Output()
	total, _ := strconv.ParseInt(strings.TrimSpace(string(totalOut)), 10, 64)

	topOut, _ := exec.Command("/usr/bin/top", "-n", "1").Output()
	lines := strings.Split(strings.TrimSpace(string(topOut)), "\n")
	var usedMB int64
	for _, line := range lines {
		if strings.HasPrefix(line, "Memory:") {
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "Real:" && i+1 < len(fields) {
					ram := fields[i+1]
					if idx := strings.Index(ram, "/"); idx > 0 {
						ram = ram[:idx]
					}
					ram = strings.TrimSuffix(ram, "M")
					usedMB, _ = strconv.ParseInt(ram, 10, 64)
					break
				}
			}
			break
		}
	}

	usedGB := float64(usedMB) / 1024
	totalGB := float64(total) / (1024 * 1024 * 1024)
	usedPct := int(float64(usedMB) * 1024 * 1024 / float64(total) * 100)

	return fmt.Sprintf("%.2f GiB of %.2f GiB\nTotal Used: %d%%", usedGB, totalGB, usedPct)
}

func getRAMPercentInt() int {
	totalOut, err := exec.Command("/sbin/sysctl", "-n", "hw.usermem").Output()
	if err != nil {
		return 0
	}
	total, err := strconv.ParseInt(strings.TrimSpace(string(totalOut)), 10, 64)
	if err != nil || total == 0 {
		return 0
	}

	topOut, err := exec.Command("/usr/bin/top", "-n", "1").Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(topOut)), "\n")
	var usedMB int64
	for _, line := range lines {
		if strings.HasPrefix(line, "Memory:") {
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "Real:" && i+1 < len(fields) {
					ram := fields[i+1]
					if idx := strings.Index(ram, "/"); idx > 0 {
						ram = ram[:idx]
					}
					ram = strings.TrimSuffix(ram, "M")
					usedMB, _ = strconv.ParseInt(ram, 10, 64)
					break
				}
			}
			break
		}
	}
	if usedMB == 0 {
		return 0
	}
	return int(float64(usedMB) * 1024 * 1024 / float64(total) * 100)
}

// RunVolume outputs volume icon for polybar
func RunVolume() error {
	vol := getVolume()
	mute := isMuted()

	var icon string
	if mute || vol == 0 {
		icon = ""
	} else if vol >= 67 {
		icon = "󰕾"
	} else if vol >= 34 {
		icon = "󰖀"
	} else {
		icon = "󰕿"
	}

	// Output: icon + tooltip with percentage
	fmt.Printf("%s\nVolume: %d%%\n", icon, vol)
	return nil
}

func getVolume() int {
	out, err := exec.Command("sh", "-c", "sndioctl output.level 2>/dev/null | cut -d= -f2").Output()
	if err != nil {
		return 0
	}
	vol, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return int(vol * 100)
}

func isMuted() bool {
	out, err := exec.Command("sh", "-c", "sndioctl output.mute 2>/dev/null | cut -d= -f2").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "1"
}

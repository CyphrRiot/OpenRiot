package polybar

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// RunMetrics outputs CPU and RAM usage for polybar
// Mimics system-metrics.sh: uses vmstat and sysctl
func RunMetrics() error {
	// CPU usage — OpenBSD: use vmstat
	cpu := getCPU()

	// RAM usage — OpenBSD
	ram := getRAM()

	fmt.Printf(" %s%% %s%%\n", cpu, ram)
	return nil
}

func getCPU() string {
	out, err := exec.Command("vmstat", "1", "1").Output()
	if err != nil {
		return "0"
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return "0"
	}
	fields := strings.Fields(lines[len(lines)-1])
	// vmstat output: idle CPU% is the last column
	if len(fields) > 0 {
		idle, err := strconv.ParseFloat(fields[len(fields)-1], 64)
		if err != nil {
			return "0"
		}
		cpu := int(100 - idle)
		return strconv.Itoa(cpu)
	}
	return "0"
}

func getRAM() string {
	// Get total memory
	totalOut, err := exec.Command("sysctl", "-n", "hw.physmem").Output()
	if err != nil {
		return "0"
	}
	total, err := strconv.ParseInt(strings.TrimSpace(string(totalOut)), 10, 64)
	if err != nil || total == 0 {
		return "0"
	}

	// Get free pages and page size
	pageSizeOut, err := exec.Command("sysctl", "-n", "hw.pagesize").Output()
	if err != nil {
		return "0"
	}
	pageSize, err := strconv.ParseInt(strings.TrimSpace(string(pageSizeOut)), 10, 64)
	if err != nil {
		return "0"
	}

	vmstatOut, err := exec.Command("vmstat", "2", "1").Output()
	if err != nil {
		return "0"
	}
	lines := strings.Split(strings.TrimSpace(string(vmstatOut)), "\n")
	if len(lines) < 3 {
		return "0"
	}
	fields := strings.Fields(lines[2])
	// free pages is typically the 5th column (index 4)
	if len(fields) > 4 {
		freePages, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil {
			return "0"
		}
		freeBytes := freePages * pageSize
		usedBytes := total - freeBytes
		ramPct := int(usedBytes * 100 / total)
		return strconv.Itoa(ramPct)
	}
	return "0"
}

// RunVolume outputs volume with icon for polybar
// Mimics volume.sh: gets volume from sndioctl, outputs icon + percentage
func RunVolume() error {
	vol := getVolume()
	mute := isMuted()

	var icon string
	if mute {
		icon = "🔇"
	} else if vol >= 70 {
		icon = "🔊"
	} else if vol >= 30 {
		icon = "🔉"
	} else {
		icon = "🔈"
	}

	fmt.Printf("%s %d%%\n", icon, vol)
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

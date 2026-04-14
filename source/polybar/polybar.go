package polybar

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RunMetrics outputs CPU icon for polybar (memory is separate module)
func RunMetrics() error {
	cpu := getCPU()
	cpuPct := getCPUPercent()
	fmt.Printf("%s\nCPU: %s%%\n", cpu, cpuPct)
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

// RunProtonDrive outputs proton-drive icon for polybar
func RunProtonDrive() error {
	icon := GetProtonDriveIcon()
	fmt.Printf("%s\nProton Drive: %s\n", icon, getProtonDriveTooltip())
	return nil
}

// GetProtonDriveIcon returns sync state icon
func GetProtonDriveIcon() string {
	if !isProtonDriveConfigured() {
		return ""
	}
	state := checkProtonDriveSync()
	if state == "synced" {
		return "󱥾" // ✓ checkmark
	}
	return "󰴋" // sync arrows
}

// IsProtonDriveConfigured returns true if Proton Drive sync is set up
func IsProtonDriveConfigured() bool {
	return isProtonDriveConfigured()
}

func isProtonDriveConfigured() bool {
	// Check for ~/ProtonSync folder
	home := os.Getenv("HOME")
	syncFolder := home + "/ProtonSync"
	if _, err := os.Stat(syncFolder); err != nil {
		return false
	}

	// Check for rclone config
	rcloneConf := home + "/.config/rclone/rclone.conf"
	if _, err := os.Stat(rcloneConf); err != nil {
		return false
	}

	return true
}

// checkProtonDriveSync returns "synced" or "needs-sync"
// If cache files missing but configured, auto-init first
func checkProtonDriveSync() string {
	home := os.Getenv("HOME")
	bisyncDir := home + "/.cache/rclone/bisync"

	path1 := bisyncDir + "/home_grendel_ProtonSync..proton_ProtonSync.path1.lst"
	path2 := bisyncDir + "/home_grendel_ProtonSync..proton_ProtonSync.path2.lst"

	// If cache files missing but configured, auto-init
	if !cacheFilesExist(path1, path2) && isProtonDriveConfigured() {
		InitProtonDriveCache()
	}

	// Check if either file is missing
	if _, err := os.Stat(path1); os.IsNotExist(err) {
		return "needs-sync"
	}
	if _, err := os.Stat(path2); os.IsNotExist(err) {
		return "needs-sync"
	}

	// Compare files (skip first line - timestamp differs each run)
	content1, _ := os.ReadFile(path1)
	content2, _ := os.ReadFile(path2)

	lines1 := strings.Split(string(content1), "\n")
	lines2 := strings.Split(string(content2), "\n")
	if len(lines1) > 1 && len(lines2) > 1 && strings.Join(lines1[1:], "\n") == strings.Join(lines2[1:], "\n") {
		return "synced"
	}
	return "needs-sync"
}

// CheckProtonDriveSyncState exports checkProtonDriveSync for main.go
func CheckProtonDriveSyncState() string {
	return checkProtonDriveSync()
}

// cacheFilesExist checks if both bisync cache files exist
func cacheFilesExist(path1, path2 string) bool {
	_, err1 := os.Stat(path1)
	_, err2 := os.Stat(path2)
	return err1 == nil && err2 == nil
}

// InitProtonDriveCache runs bisync --dry-run to populate cache files
func InitProtonDriveCache() error {
	home := os.Getenv("HOME")
	cmd := exec.Command("rclone", "bisync",
		home+"/ProtonSync",
		"proton:ProtonSync",
		"--dry-run",
		"--work-dir", home+"/.cache/rclone/bisync")
	return cmd.Run()
}

func getProtonDriveTooltip() string {
	if isProtonDriveConfigured() {
		return "Ready to sync"
	}
	return "Not configured"
}

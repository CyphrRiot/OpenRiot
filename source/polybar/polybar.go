package polybar

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"openriot/screen"
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
	// Use 2 samples - first one is often stale, last is current
	out, err := exec.Command("/usr/bin/vmstat", "1", "2").Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0
	}
	// Get the LAST line (most recent sample)
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
	// 0-25%: 󰢿, 25-50%: 󰢼, 50-75%: 󰢽, 75-100%: 󰢾
	switch {
	case ram >= 75:
		return "󰢾"
	case ram >= 50:
		return "󰢽"
	case ram >= 25:
		return "󰢼"
	default:
		return "󰢿"
	}
}

func getRAMPercent() string {
	return strconv.Itoa(getRAMPercentInt())
}

// GetCPUPercent returns CPU usage as "XX%" string (for notifications)
func GetCPUPercent() string {
	return getCPUPercent() + "%"
}

// GetCPUDetails returns detailed CPU info for notifications
// Format: "CPU in Use: XX%\nProcessors: N\nModel: ..."
func GetCPUDetails() string {
	cpuPct := getCPUPercent()

	// Get number of CPUs
	ncpuOut, _ := exec.Command("/sbin/sysctl", "-n", "hw.ncpu").Output()
	ncpu := strings.TrimSpace(string(ncpuOut))

	// Get CPU model
	modelOut, _ := exec.Command("/sbin/sysctl", "-n", "hw.model").Output()
	model := strings.TrimSpace(string(modelOut))

	return fmt.Sprintf("CPU in Use: %s%%\nProcessors: %s\nModel: %s", cpuPct, ncpu, model)
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
	// Check for ~/Documents/ProtonSync folder
	home := os.Getenv("HOME")
	syncFolder := home + "/Documents/ProtonSync"
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

// getCurrentUsername returns the current system username
func getCurrentUsername() string {
	user, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return user.Username
}

// checkProtonDriveSync returns "synced" or "needs-sync"
func checkProtonDriveSync() string {
	home := os.Getenv("HOME")
	bisyncDir := home + "/.cache/rclone/bisync"

	// Find cache files using glob (works regardless of naming)
	matches1, _ := filepath.Glob(bisyncDir + "/*path1.lst")
	matches2, _ := filepath.Glob(bisyncDir + "/*path2.lst")

	if len(matches1) == 0 || len(matches2) == 0 {
		return "needs-sync"
	}

	path1 := matches1[0]
	path2 := matches2[0]

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

	// Check if any local file was modified after last bisync run
	content1, _ := os.ReadFile(path1)
	lines1 := strings.Split(string(content1), "\n")
	localFiles := getLocalFileList(home + "/Documents/ProtonSync")
	for _, name := range localFiles {
		localPath := home + "/Documents/ProtonSync/" + name
		localMtime := getFileMtime(localPath)
		cachedMtime := getCachedMtime(name, lines1[1:])
		if cachedMtime.IsZero() || localMtime.After(cachedMtime) {
			return "needs-sync"
		}
	}

	// Also check bidirectional: cache→local (files deleted locally should trigger sync)
	cachedFiles := extractFilenames(lines1[1:])
	if !filesInCache(cachedFiles, localFiles) {
		return "needs-sync"
	}

	return "synced"
}

// getDirMtime returns the modification time of a directory (newest file inside)
func getDirMtime(dir string) time.Time {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return time.Time{}
	}
	var newest time.Time
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
	}
	return newest
}

// getFileMtime returns the modification time of a single file
func getFileMtime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// getCachedMtime extracts the mtime for a file from bisync cache lines
func getCachedMtime(filename string, lines []string) time.Time {
	prefix := `"` + filename + `"`
	for _, line := range lines {
		if strings.Contains(line, prefix) {
			// Format: "-    18169 - - 2026-04-14T01:41:11.186703920+0000 \"filename\""
			// Find timestamp after " - - " delimiter, before the opening quote
			idx := strings.Index(line, " - - ")
			quoteIdx := strings.Index(line, `"`)
			if idx > 0 && quoteIdx > 0 {
				ts := strings.TrimSpace(line[idx+5 : quoteIdx])
				t, err := time.Parse("2006-01-02T15:04:05.999999999-0700", ts)
				if err == nil {
					return t
				}
			}
		}
	}
	return time.Time{}
}

// getLocalFileList returns list of files in directory
func getLocalFileList(dir string) []string {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	return files
}

// extractFilenames extracts filenames from cache file lines
func extractFilenames(lines []string) []string {
	var names []string
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Format: "-    18070 - - timestamp \"filename\""
		start := strings.Index(line, "\"")
		end := strings.LastIndex(line, "\"")
		if start >= 0 && end > start {
			name := line[start+1 : end]
			names = append(names, name)
		}
	}
	return names
}

// filesInCache returns true if all local files are in cache
func filesInCache(local, cached []string) bool {
	cacheMap := make(map[string]bool)
	for _, c := range cached {
		cacheMap[c] = true
	}
	for _, f := range local {
		if !cacheMap[f] {
			return false
		}
	}
	return true
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
		home+"/Documents/ProtonSync",
		"proton:ProtonSync",
		"--dry-run",
		"--work-dir", home+"/.cache/rclone/bisync")
	return cmd.Run()
}

func getProtonDriveTooltip() string {
	if !isProtonDriveConfigured() {
		return "Not configured"
	}
	state := checkProtonDriveSync()
	if state == "synced" {
		return "Synced"
	}
	return "Needs sync"
}

// GetProtonDriveTooltipText returns a formatted sync status string for notifications
func GetProtonDriveTooltipText() string {
	if !isProtonDriveConfigured() {
		return "Not configured"
	}
	state := checkProtonDriveSync()
	if state == "synced" {
		return getSyncTime()
	}
	return "Needs sync"
}

// getSyncTime returns the last sync time formatted for display
func getSyncTime() string {
	home := os.Getenv("HOME")
	bisyncDir := home + "/.cache/rclone/bisync"

	// Find path1.lst to get last sync time from mtime
	matches, _ := filepath.Glob(bisyncDir + "/*path1.lst")
	if len(matches) == 0 {
		return "Recently"
	}

	info, err := os.Stat(matches[0])
	if err != nil {
		return "Recently"
	}

	// Format: "April 20, 09:00 AM"
	return info.ModTime().Format("January 2, 03:04 PM")
}

// Setup generates scaled polybar config (doesn't launch polybar - i3 handles that)
func Setup() int {
	home := os.Getenv("HOME")

	// Get screen resolution
	width := screen.GetWidth()

	// Determine scale factors based on resolution
	height, font0, font1, modMargin := getScaleFactors(width)

	// Read template config
	templatePath := filepath.Join(home, ".local/share/openriot/config/polybar/config.ini")
	configPath := filepath.Join(home, ".config/polybar/config.ini")

	template, err := os.ReadFile(templatePath)
	if err != nil {
		return 1
	}

	// Apply scaling transformations (do larger sizes first to avoid collision)
	content := string(template)
	replacements := []struct{ old, new string }{
		{"height = 26", "height = " + height},
		{"module-margin = 1", "module-margin = " + modMargin},
		{"Hurmit Nerd Font:size=20", "Hurmit Nerd Font:" + font1},
		{"Hurmit Nerd Font:size=11", "Hurmit Nerd Font:" + font0},
	}

	for _, r := range replacements {
		content = strings.ReplaceAll(content, r.old, r.new)
	}

	// Write scaled config
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return 1
	}
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return 1
	}
	fmt.Println("[DONE] Polybar scaled. Run `Super+Shift+R` to apply changes.")
	return 0
}

func getScaleFactors(width int) (height, font0, font1, modMargin string) {
	switch {
	case width >= 2560: // 1440p or 4K
		height = "32"
		font0 = "size=13"
		font1 = "size=17"
		modMargin = "2"
	case width >= 1920: // 1080p
		height = "28"
		font0 = "size=11"
		font1 = "size=15"
		modMargin = "2"
	default: // Below 1080p
		height = "28"
		font0 = "size=11"
		font1 = "size=15"
		modMargin = "2"
	}
	return
}

func startPolybar() int {
	cmd := exec.Command("polybar", "main")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return 1
	}
	return 0
}

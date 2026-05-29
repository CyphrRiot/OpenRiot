package polybar

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"encoding/binary"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/sys/unix"
	"openriot/installer"
	"openriot/notify"
	"openriot/paths"
	"openriot/screen"
	"openriot/windowicon"
	"openriot/windowtitle"
)

// Icon wraps a Nerd Font glyph with polybar click-area offset to fix right-edge clickability on right-side modules.
func Icon(icon string) string {
	if icon == "" {
		return ""
	}
	return icon + "%{O2}"
}

// IsProcessRunning checks if a process with the exact name is active.
func IsProcessRunning(name string) bool {
	out, _ := exec.Command("pgrep", "-x", name).Output()
	return len(strings.TrimSpace(string(out))) > 0
}

// RunNzbget outputs the nzbget polybar icon.
// Running → 󱑤; installed but not running → 󱑥; otherwise empty.
func RunNzbget() error {
	if IsProcessRunning("nzbget") {
		fmt.Print(Icon("󱑤"))
		return nil
	}
	if _, err := os.Stat("/usr/local/bin/nzbget"); err == nil {
		fmt.Print(Icon("󱑥"))
	}
	return nil
}

// RunMetrics outputs CPU icon for polybar (memory is separate module)
func RunMetrics() error {
	cpu := getCPU()
	fmt.Print(Icon(cpu))
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

var cpuCache struct {
	Value     int
	Time      time.Time
	ValidFor  time.Duration
}

func init() {
	cpuCache.ValidFor = 5 * time.Second
}

func getCPUPercentInt() int {
	if time.Since(cpuCache.Time) < cpuCache.ValidFor {
		return cpuCache.Value
	}

	raw1, err := unix.SysctlRaw("kern.cp_time")
	if err != nil || len(raw1) < 48 {
		return 0
	}
	c1 := parseCPTime(raw1)

	time.Sleep(500 * time.Millisecond)

	raw2, err := unix.SysctlRaw("kern.cp_time")
	if err != nil || len(raw2) < 48 {
		return 0
	}
	c2 := parseCPTime(raw2)

	var totalDiff, idleDiff uint64
	for i := 0; i < 6; i++ {
		diff := c2[i] - c1[i]
		totalDiff += diff
		if i == 5 {
			idleDiff = diff
		}
	}
	if totalDiff == 0 {
		return 0
	}
	cpuCache.Value = int((1.0 - float64(idleDiff)/float64(totalDiff)) * 100)
	cpuCache.Time = time.Now()
	return cpuCache.Value
}

func parseCPTime(raw []byte) []uint64 {
	counters := make([]uint64, 6)
	for i := 0; i < 6 && (i+1)*8 <= len(raw); i++ {
		counters[i] = binary.LittleEndian.Uint64(raw[i*8 : i*8+8])
	}
	return counters
}

// GetRAM returns memory icon (for polybar)
func GetRAM() string {
	ram := getRAMPercentInt()
	// 0-25%: 󰢿, 25-50%: 󰢼, 50-75%: 󰢽, 75-100%: 󰢾
	switch {
	case ram >= 75:
		return Icon("󰢾")
	case ram >= 50:
		return Icon("󰢽")
	case ram >= 25:
		return Icon("󰢼")
	default:
		return Icon("󰢿")
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

	ncpu, _ := unix.SysctlUint32("hw.ncpu")
	model, _ := unix.Sysctl("hw.model")

	return fmt.Sprintf("CPU in Use: %s%%\nProcessors: %d\nModel: %s", cpuPct, ncpu, model)
}

// GetMemPercent returns memory usage as "XX%" string (for notifications)
func GetMemPercent() string {
	return getRAMPercent() + "%"
}

// GetMemDetails returns detailed memory info for notifications
// Format: "X.XX GiB of Y.YY GiB\nTotal Used: Z%"
func GetMemDetails() string {
	usedMB, totalMB := getMemStats()
	if totalMB == 0 {
		return "Memory: unavailable"
	}
	usedGB := float64(usedMB) / 1024
	totalGB := float64(totalMB) / 1024
	usedPct := int(float64(usedMB) / float64(totalMB) * 100)

	return fmt.Sprintf("%.2f GiB of %.2f GiB\nTotal Used: %d%%", usedGB, totalGB, usedPct)
}

func getRAMPercentInt() int {
	usedMB, totalMB := getMemStats()
	if totalMB == 0 {
		return 0
	}
	return int(float64(usedMB) / float64(totalMB) * 100)
}

func getMemStats() (usedMB, totalMB int64) {
	uvm, err := unix.SysctlUvmexp("vm.uvmexp")
	if err != nil {
		return 0, 0
	}
	p := int64(uvm.Pagesize)
	totalPages := int64(uvm.Npages) - int64(uvm.Free)
	usedPages := int64(uvm.Active)
	return usedPages * p / (1024 * 1024), totalPages * p / (1024 * 1024)
}

// RunVolume outputs volume icon for polybar
func RunVolume() error {
	vol := getVolume()
	mute := isMuted()

	var icon string
	if mute || vol == 0 {
		icon = ""
	} else if vol > 75 {
		icon = ""
	} else if vol >= 45 {
		icon = "󰕾"
	} else if vol >= 10 {
		icon = "󱄠"
	} else {
		icon = ""
	}

	// Output: icon only
	fmt.Print(Icon(icon))
	return nil
}

func getVolume() int {
	out, err := exec.Command("sndioctl", "-n", "output.level").Output()
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
	out, err := exec.Command("sndioctl", "-n", "output.mute").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "1"
}

// RunProtonDrive outputs proton-drive icon for polybar
func RunProtonDrive() error {
	if !isProtonDriveConfigured() {
		return nil
	}
	fmt.Print(Icon(GetProtonDriveIcon()))
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
	if _, err := os.Stat(paths.Join("Documents", "ProtonSync")); err != nil {
		return false
	}
	if _, err := os.Stat(paths.Join(".config", "rclone", "rclone.conf")); err != nil {
		return false
	}
	return true
}

// checkProtonDriveSync returns "synced" or "needs-sync"
func checkProtonDriveSync() string {
	bisyncDir := paths.Join(".cache", "rclone", "bisync")

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
	syncDir := paths.Join("Documents", "ProtonSync")
	localFiles := getLocalFileList(syncDir)
	for _, name := range localFiles {
		localPath := filepath.Join(syncDir, name)
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

// TriggerSync runs the full Proton Drive sync with notifications
func TriggerSync() error {
	if !IsProtonDriveConfigured() {
		notify.SendNotify("proton-drive", "Proton Drive", "Not configured\nSee OpenRiot.org for setup info", "critical", 5000, 0)
		return nil
	}
	state := CheckProtonDriveSyncState()
	if state == "synced" {
		notify.SendNotify("proton-drive", "Proton Drive", "Synchronized: "+GetProtonDriveTooltipText(), "normal", 5000, 0)
		return nil
	}
	notify.SendNotify("proton-drive", "Proton Drive", "Syncing...", "normal", 2000, 0)
	cmd := `printf "Proton Drive Sync\nFrom: ~/Documents/ProtonSync -> Proton Drive Cloud\n\nWould you like to do a bi-directional Sync or one-way\n  and replace items in the Cloud with local items?\n\n[Y]es for bi-directional sync (or ENTER),\n[O]ne-way for One-Way sync or\n[Q]uit or [N]o ?\n\nChoose your adventure [Y/o/q/n] -> "; read -r ans; case "$ans" in o|O) echo "One-way sync selected..."; rclone copy ~/Documents/ProtonSync proton:ProtonSync --progress; printf "\nDone. Press Enter to close..."; read -r ans ;; [yY]|"") echo "Bi-directional sync selected..."; rclone bisync ~/Documents/ProtonSync proton:ProtonSync --resync --progress; printf "\nDone. Press Enter to close..."; read -r ans ;; *) echo "Canceled."; sleep 1 ;; esac`
	exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
	return nil
}

// cacheFilesExist checks if both bisync cache files exist
func cacheFilesExist(path1, path2 string) bool {
	_, err1 := os.Stat(path1)
	_, err2 := os.Stat(path2)
	return err1 == nil && err2 == nil
}

// InitProtonDriveCache runs bisync --dry-run to populate cache files
func InitProtonDriveCache() error {
	cmd := exec.Command("rclone", "bisync",
		paths.Join("Documents", "ProtonSync"),
		"proton:ProtonSync",
		"--dry-run",
		"--work-dir", paths.Join(".cache", "rclone", "bisync"))
	return cmd.Run()
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
	bisyncDir := paths.Join(".cache", "rclone", "bisync")

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

// Setup generates scaled polybar config (doesn't launch polybar - i3
// handles that)
func Setup() int {
	// Get screen resolution
	width := screen.GetWidth()

	// Determine scale factors based on resolution
	height, font0, font1, font2, modMargin := getScaleFactors(
		width)

	// Render color template
	templatePath := paths.OpenRiotDir("config", "polybar",
		"config.ini.tmpl")
	configPath := paths.Join(".config", "polybar", "config.ini")

	content, _, err := installer.RenderTemplateString(templatePath)
	if err != nil {
		return 1
	}

	// Apply scaling transformations (do larger sizes first to
	// avoid collision)
	replacements := []struct{ old, new string }{
		{"height = 26", "height = " + height},
		{"module-margin = 1",
			"module-margin = " + modMargin},
		{"Hurmit Nerd Font:size=20",
			"Hurmit Nerd Font:" + font1},
		{"Hurmit Nerd Font:size=11",
			"Hurmit Nerd Font:" + font0},
		{"Hurmit Nerd Font:size=13",
			"Hurmit Nerd Font:" + font2},
	}

	for _, r := range replacements {
		content = strings.ReplaceAll(content, r.old, r.new)
	}

	// On small screens reduce title length and tighten bar padding
	if width < 1360 {
		content = strings.ReplaceAll(content,
			"padding-right = 3", "padding-right = 0")
		windowtitle.SetMaxLen(24)
	}

	// Write scaled config
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return 1
	}
	if err := os.WriteFile(configPath, []byte(content),
		0600); err != nil {
		return 1
	}
	fmt.Println("[DONE] Polybar scaled. Run `Super+Shift+R` to " +
		"apply changes.")
	return 0
}

func getScaleFactors(width int) (height, font0, font1, font2, modMargin string) {
	switch {
	case width >= 2560: // 1440p or 4K
		height = "32"
		font0 = "size=13"
		font1 = "size=17"
		font2 = "size=16"
		modMargin = "1"
	case width >= 1920: // 1080p
		height = "28"
		font0 = "size=11"
		font1 = "size=15"
		font2 = "size=13"
		modMargin = "1"
	case width >= 1360: // WXGA+ / 900p
		height = "26"
		font0 = "size=10"
		font1 = "size=13"
		font2 = "size=11"
		modMargin = "1"
	default: // Below 1360 — e.g. 1280x720
		height = "22"
		font0 = "size=8.5"
		font1 = "size=12"
		font2 = "size=9"
		modMargin = "0"
	}
	return
}


// RunAll outputs a single polybar-formatted line combining workspace icons
// and focused window title, avoiding separate process spawns for both.
func RunAll() error {
	tree := getFullTree()
	workspaces := getAllWorkspaces()

	// Build icon map from cached TOML mappings + already-parsed tree.
	// This avoids a second expensive i3-msg -t get_tree call.
	mappings := windowicon.GetAllMappings()
	windowIcons := buildWindowIconsFromTree(tree, mappings)

	wsStr := buildWorkspacesPolybar(tree, workspaces, windowIcons)
	titleStr := findFocusedTitleInTree(tree)

	fmt.Print(wsStr + "  %{F#565F89}" + titleStr + "%{F-}\n")
	return nil
}

// buildWindowIconsFromTree extracts all unique window class→icon mappings
// from an already-parsed i3 tree, avoiding a redundant i3-msg call.
func buildWindowIconsFromTree(tree *i3Tree, mappings map[string]string) map[string]string {
	seen := make(map[string]bool)
	collectAllClasses(tree.Nodes, seen)
	collectAllClasses(tree.FloatingNodes, seen)

	result := make(map[string]string)
	for class := range seen {
		icon := ""
		if v, ok := mappings[strings.ToLower(class)]; ok {
			icon = v
		}
		result[class] = icon
	}
	return result
}

func collectAllClasses(nodes []i3Node, seen map[string]bool) {
	for _, n := range nodes {
		if n.Window != 0 && n.WindowProps.Class != "" {
			cls := n.WindowProps.Class
			if windowicon.IsPrivateFirefox(cls, n.Name) {
				seen["firefox-private"] = true
			} else {
				seen[cls] = true
			}
		}
		collectAllClasses(n.Nodes, seen)
		collectAllClasses(n.FloatingNodes, seen)
	}
}

// buildWorkspacesPolybar replicates workspaceicons.GetAll() output for use in
// a single module, consuming an already-parsed tree.
func buildWorkspacesPolybar(tree *i3Tree, workspaces []i3Workspace, windowIcons map[string]string) string {
	var results []string
	for wsNum := 1; wsNum <= 3; wsNum++ {
		classes := findWindowsInWorkspaceFromTree(tree.Nodes, wsNum)
		state := getStateFromWorkspaces(workspaces, wsNum)

		var icons []string
		for _, cls := range classes {
			if icon, ok := windowIcons[cls]; ok {
				icons = append(icons, icon)
			}
		}
		if len(icons) > 4 {
			icons = icons[:4]
		}

		allSlots := make([]string, 0, 4)
		for _, icon := range icons {
			allSlots = append(allSlots, icon)
		}
		for len(allSlots) < 4 {
			allSlots = append(allSlots, "\uec03")
		}

		indicator := getIndicator(state, len(icons) > 0)

		if state == "unfocused" && len(icons) > 0 {
			for i := 0; i < len(icons); i++ {
				allSlots[i] = fmt.Sprintf("%%{T0}%%{F#565F89}%s%%{F-}%%{T-}", icons[i])
			}
			indicator = fmt.Sprintf("%%{F#565F89}%s%%{F-}", indicator)
		}
		if state == "unfocused" && len(icons) == 0 {
			indicator = fmt.Sprintf("%%{F#1A1D26}%s%%{F-}", indicator)
		}

		content := fmt.Sprintf("%s %s", indicator, strings.Join(allSlots, " "))
		if state == "focused" {
			content = fmt.Sprintf("%%{F#9ECE6A}%s%%{F-} %%{F#8BB85A}%s%%{F-}", indicator, strings.Join(allSlots, " "))
		}
		if state == "urgent" {
			content = fmt.Sprintf("%%{F#F7768E}%s%%{F-}", content)
		}
		results = append(results,
			fmt.Sprintf("%%{A:$HOME/.local/share/openriot/install/openriot --workspace-switch %d:}%s%%{A}", wsNum, content))
	}
	return strings.Join(results, "  ")
}

// getIndicator mirrors workspaceicons.getIndicator.
func getIndicator(state string, hasApps bool) string {
	switch state {
	case "focused":
		return ""
	case "urgent":
		return ""
	case "unfocused":
		if hasApps {
			return ""
		}
		return ""
	default:
		return ""
	}
}

// i3Tree and i3Node mirror workspaceicons types.
type i3Tree struct {
	Nodes         []i3Node `json:"nodes"`
	FloatingNodes []i3Node `json:"floating_nodes"`
}

type i3Node struct {
	Type          string    `json:"type"`
	Num           int       `json:"num"`
	Focused       bool      `json:"focused"`
	Urgent        bool      `json:"urgent"`
	Window        int       `json:"window"`
	Name          string    `json:"name"`
	WindowProps   windowProps `json:"window_properties"`
	Nodes         []i3Node  `json:"nodes"`
	FloatingNodes []i3Node  `json:"floating_nodes"`
}

type windowProps struct {
	Class string `json:"class"`
}

type i3Workspace struct {
	Num     int  `json:"num"`
	Focused bool `json:"focused"`
	Urgent  bool `json:"urgent"`
}

func getFullTree() *i3Tree {
	cmd := exec.Command("i3-msg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	var tree i3Tree
	if err := json.Unmarshal(output, &tree); err != nil {
		return nil
	}
	return &tree
}

func getAllWorkspaces() []i3Workspace {
	cmd := exec.Command("i3-msg", "-t", "get_workspaces")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	var workspaces []i3Workspace
	if err := json.Unmarshal(output, &workspaces); err != nil {
		return nil
	}
	return workspaces
}

func findWindowsInWorkspaceFromTree(nodes []i3Node, wsNum int) []string {
	var classes []string
	findWindowsInWorkspace(nodes, wsNum, &classes)
	return classes
}

func findWindowsInWorkspace(nodes []i3Node, wsNum int, classes *[]string) {
	for _, n := range nodes {
		if n.Type == "workspace" && n.Num == wsNum {
			collectWindows(n, classes)
			return
		}
		findWindowsInWorkspace(n.Nodes, wsNum, classes)
		findWindowsInWorkspace(n.FloatingNodes, wsNum, classes)
	}
}

func collectWindows(node i3Node, classes *[]string) {
	if node.Window != 0 && node.WindowProps.Class != "" {
		cls := node.WindowProps.Class
		if windowicon.IsPrivateFirefox(cls, node.Name) {
			*classes = append(*classes, "firefox-private")
		} else {
			*classes = append(*classes, cls)
		}
	}
	for _, n := range node.Nodes {
		collectWindows(n, classes)
	}
	for _, n := range node.FloatingNodes {
		collectWindows(n, classes)
	}
}

func getStateFromWorkspaces(workspaces []i3Workspace, wsNum int) string {
	if workspaces == nil {
		return "unfocused"
	}
	for _, ws := range workspaces {
		if ws.Num == wsNum {
			if ws.Urgent {
				return "urgent"
			}
			if ws.Focused {
				return "focused"
			}
			return "unfocused"
		}
	}
	return "unfocused"
}

func findFocusedTitleInTree(tree *i3Tree) string {
	if tree == nil {
		return ""
	}
	title := findFocusedTitleInNodes(tree.Nodes)
	if title == "" {
		title = findFocusedTitleInNodes(tree.FloatingNodes)
	}
	return formatWindowTitle(title)
}

func findFocusedTitleInNodes(nodes []i3Node) string {
	for _, n := range nodes {
		if n.Focused && n.Type == "con" && n.Window != 0 && n.Name != "" {
			return n.Name
		}
		if title := findFocusedTitleInNodes(n.Nodes); title != "" {
			return title
		}
		if title := findFocusedTitleInNodes(n.FloatingNodes); title != "" {
			return title
		}
	}
	return ""
}

func formatWindowTitle(title string) string {
	if title == "" {
		return " ~"
	}
	title = stripEmojis(title)
	runes := []rune(title)
	if len(runes) > 36 {
		trunc := 33
		if trunc < 1 {
			trunc = 1
		}
		return fmt.Sprintf("%-*s", trunc, string(runes[:trunc])) + "..."
	}
	return fmt.Sprintf("%-*s", 36, string(runes))
}

func stripEmojis(s string) string {
	runes := []rune(s)
	var cleaned []rune
	for _, r := range runes {
		if r >= 0x1F300 && r <= 0x1F9FF {
			continue
		}
		if r >= 0x2600 && r <= 0x26FF {
			continue
		}
		if r >= 32 && r <= 126 {
			cleaned = append(cleaned, r)
		} else if r > 127 && r < 0x11000 {
			if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r) || unicode.IsSpace(r) {
				cleaned = append(cleaned, r)
			}
		}
	}
	return string(cleaned)
}


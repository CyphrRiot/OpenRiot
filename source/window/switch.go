package window

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"openriot/paths"
	"openriot/theme"
	"openriot/windowicon"
)

type windowRect struct {
	X      int64 `json:"x"`
	Y      int64 `json:"y"`
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

type i3Node struct {
	ID            int64       `json:"id"`
	Type          string      `json:"type"`
	Num           int         `json:"num"`
	Focused       bool        `json:"focused"`
	Window        int         `json:"window"`
	Name          string      `json:"name"`
	Rect          windowRect  `json:"rect"`
	WindowProps   windowProps `json:"window_properties"`
	Nodes         []i3Node    `json:"nodes"`
	FloatingNodes []i3Node    `json:"floating_nodes"`
}

type windowProps struct {
	Class    string `json:"class"`
	Instance string `json:"instance"`
	Title    string `json:"title"`
}

type windowEntry struct {
	ID        int64
	Icon      string
	Base      string
	Label     string
	Class     string
	Focused   bool
	Workspace int
}

var terminals = map[string]bool{
	"alacritty":      true,
	"urxvt":          true,
	"rxvt":           true,
	"kitty":          true,
	"xterm":          true,
	"gnome-terminal": true,
	"konsole":        true,
	"terminator":     true,
	"tilix":          true,
	"st":             true,
	"foot":           true,
	"wezterm":        true,
	"ghostty":        true,
	"lxterminal":     true,
	"qterminal":      true,
	"xfce4-terminal": true,
}

var skipClasses = map[string]bool{
	"polybar": true,
	"dunst":   true,
	"i3bar":   true,
	"i3lock":  true,
}

// RunSwitch launches the window switcher.
func RunSwitch() error {
	tree, err := getTree()
	if err != nil {
		return fmt.Errorf("i3 tree: %w", err)
	}

	classOverrides, cmdOverrides := loadAppNames()
	entries := collectWindows(tree, classOverrides, cmdOverrides)
	if len(entries) == 0 {
		return fmt.Errorf("no windows found")
	}

	// Compute max base width for fixed-width alignment
	maxBaseWidth := 0
	for _, e := range entries {
		if w := len(e.Base); w > maxBaseWidth {
			maxBaseWidth = w
		}
	}

	var input bytes.Buffer
	for _, e := range entries {
		pad := strings.Repeat(" ", maxBaseWidth-len(e.Base))
		fmt.Fprintf(&input, "%s  %s%s  %s\n", e.Icon, e.Base, pad, e.Label)
	}

	lines := len(entries)
	if lines > 16 {
		lines = 16
	}
	themeStr := fmt.Sprintf(
		"window { width: 750px; border: 2px; border-color: %s; } listview { columns: 1; lines: %d; flow: vertical; scrollbar: false; padding: 8px 0px; } element { padding: 6px 8px; border-radius: 4px; } inputbar { padding: 8px 12px; } icon-search { size: 14px; }",
		theme.GetAccent(),
		lines,
	)

	theme := paths.Join(".config", "rofi", "simple-tokyonight.rasi")

	args := []string{"-dmenu", "-i", "-p", "Windows", "-format", "i", "-theme", theme, "-theme-str", themeStr}
	if _, err := os.Stat(theme); os.IsNotExist(err) {
		args = []string{"-dmenu", "-i", "-p", "Windows", "-format", "i", "-theme", "simple-tokyonight", "-theme-str", themeStr}
	}

	cmd := exec.Command("rofi", args...)
	cmd.Stdin = &input
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil
	}

	idxStr := strings.TrimSpace(out.String())
	if idxStr == "" {
		return nil
	}

	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return fmt.Errorf("invalid selection: %s", idxStr)
	}

	if idx < 0 || idx >= len(entries) {
		return fmt.Errorf("selection out of range: %d", idx)
	}

	selected := entries[idx]
	focusCmd := exec.Command("i3-msg", fmt.Sprintf("[con_id=%d] focus", selected.ID))
	focusCmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return focusCmd.Run()
}

func getTree() (*i3Node, error) {
	cmd := exec.Command("i3-msg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var tree i3Node
	if err := json.Unmarshal(output, &tree); err != nil {
		return nil, err
	}
	return &tree, nil
}

func collectWindows(root *i3Node, classOverrides, cmdOverrides map[string]string) []windowEntry {
	var entries []windowEntry
	var ws int
	walk(root, &entries, &ws, classOverrides, cmdOverrides)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Workspace != entries[j].Workspace {
			return entries[i].Workspace < entries[j].Workspace
		}
		if entries[i].Focused != entries[j].Focused {
			return entries[i].Focused
		}
		return entries[i].Label < entries[j].Label
	})
	return entries
}

func walk(node *i3Node, entries *[]windowEntry, currentWS *int, classOverrides, cmdOverrides map[string]string) {
	if node == nil {
		return
	}

	if node.Type == "workspace" && node.Num > 0 {
		*currentWS = node.Num
	}

	if node.Type == "con" && node.Window != 0 {
		class := strings.ToLower(node.WindowProps.Class)
		instance := strings.ToLower(node.WindowProps.Instance)

		if skipClasses[class] || skipClasses[instance] {
			return
		}

		icon := windowicon.Get(class)
		if windowicon.IsPrivateFirefox(class, node.Name) {
			icon = windowicon.Get("firefox-private")
		} else if icon == "\uf059" && instance != "" {
			icon = windowicon.Get(instance)
		}
		base, label := formatLabel(node.Name, node.WindowProps.Title, class, instance, classOverrides, cmdOverrides)

		*entries = append(*entries, windowEntry{
			ID:        node.ID,
			Icon:      icon,
			Base:      base,
			Label:     label,
			Class:     class,
			Focused:   node.Focused,
			Workspace: *currentWS,
		})
	}

	for i := range node.Nodes {
		walk(&node.Nodes[i], entries, currentWS, classOverrides, cmdOverrides)
	}
	for i := range node.FloatingNodes {
		walk(&node.FloatingNodes[i], entries, currentWS, classOverrides, cmdOverrides)
	}
}

func formatLabel(name, title, class, instance string, classOverrides, cmdOverrides map[string]string) (string, string) {
	// 1. Determine base name from apps.txt or capitalized class
	base := ""
	if override, ok := classOverrides[class]; ok {
		base = override
	} else if override, ok := classOverrides[instance]; ok {
		base = override
	} else if override, ok := cmdOverrides[class]; ok {
		base = override
	} else if override, ok := cmdOverrides[instance]; ok {
		base = override
	} else {
		display := class
		if display == "" {
			display = instance
		}
		if display == "" {
			display = name
		}
		if display == "" {
			base = "Window"
		} else {
			base = strings.ToUpper(display[:1]) + display[1:]
		}
	}

	// 2. Use window title if present and different from base name
	displayTitle := title
	if displayTitle == "" {
		displayTitle = name
	}
	if displayTitle != "" && strings.ToLower(displayTitle) != strings.ToLower(base) {
		return base, base + " - " + displayTitle
	}
	return base, base
}

func loadAppNames() (classOverrides, cmdOverrides map[string]string) {
	classOverrides = make(map[string]string)
	cmdOverrides = make(map[string]string)

	appsFile := paths.Join(".config", "rofi", "apps.txt")
	f, err := os.Open(appsFile)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		appName := strings.TrimSpace(parts[0])
		cmd := strings.TrimSpace(parts[1])

		// Extract --class value
		if idx := strings.Index(cmd, "--class"); idx != -1 {
			after := cmd[idx+len("--class"):]
			after = strings.TrimLeft(after, " \t=")
			fields := strings.Fields(after)
			if len(fields) > 0 {
				classOverrides[strings.ToLower(fields[0])] = appName
			}
		}

		// Base command override
		base := strings.Fields(cmd)[0]
		base = strings.ToLower(filepath.Base(base))
		if _, exists := cmdOverrides[base]; !exists {
			cmdOverrides[base] = appName
		}
	}

	return
}

package window

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"openriot/lock"
	"openriot/notify"
	"openriot/paths"
)

// WindowInfo captures the state of one window for save/restore.
type WindowInfo struct {
	Class     string     `json:"class"`
	Instance  string     `json:"instance"`
	Workspace int        `json:"workspace"`
	Rect      windowRect `json:"rect"`
	Floating  bool       `json:"floating"`
	Focused   bool       `json:"focused"`
}

// Snapshot is the full window layout saved to disk.
type Snapshot struct {
	Timestamp time.Time    `json:"timestamp"`
	Windows   []WindowInfo `json:"windows"`
}

var shutdownSkipClasses = map[string]bool{
	"polybar":   true,
	"dunst":     true,
	"i3bar":     true,
	"i3lock":    true,
	"picom":     true,
	"xautolock": true,
}

func snapshotFilePath() string {
	return paths.Join(".cache", "openriot", "window-snapshot.json")
}

// SaveLayout captures the current i3 window layout and writes it to disk.
func SaveLayout() error {
	tree, err := getTree()
	if err != nil {
		return fmt.Errorf("i3 tree: %w", err)
	}

	var windows []WindowInfo
	var wsNum int
	collectForSnapshot(tree, &windows, &wsNum, false)

	snap := Snapshot{
		Timestamp: time.Now(),
		Windows:   windows,
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	cacheDir := paths.Join(".cache", "openriot")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("mkdir cache: %w", err)
	}

	return os.WriteFile(snapshotFilePath(), data, 0600)
}

// RestoreLayout reads the saved snapshot and restores windows to their
// previous workspaces and positions. Deletes the snapshot file on success.
func RestoreLayout() error {
	if os.Getenv("DISPLAY") == "" {
		return nil
	}

	path := snapshotFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		os.Remove(path)
		return nil
	}

	if len(snap.Windows) == 0 {
		os.Remove(path)
		return nil
	}

	classCommands := loadRestoreCommands()
	launched := make(map[string]bool)

	for _, w := range snap.Windows {
		key := strings.ToLower(w.Class)
		if key == "" {
			key = strings.ToLower(w.Instance)
		}
		if key == "" || shutdownSkipClasses[key] {
			continue
		}
		if launched[key] {
			continue
		}

		cmdLine, ok := classCommands[key]
		if !ok {
			cmdLine = key
		}
		launched[key] = true

		args := strings.Fields(cmdLine)
		if len(args) == 0 {
			continue
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		cmd.Start()
	}

	time.Sleep(3 * time.Second)

	tree, err := getTree()
	if err != nil {
		os.Remove(path)
		return nil
	}

	var currentWindows []windowPlacement
	var cws int
	collectForPlacement(tree, &currentWindows, &cws)
	currentByClass := make(map[string][]int64)
	for _, cw := range currentWindows {
		cls := strings.ToLower(cw.Class)
		if cls == "" {
			cls = strings.ToLower(cw.Instance)
		}
		currentByClass[cls] = append(currentByClass[cls], cw.ID)
	}

	for _, w := range snap.Windows {
		cls := strings.ToLower(w.Class)
		if cls == "" {
			cls = strings.ToLower(w.Instance)
		}
		ids := currentByClass[cls]
		if len(ids) == 0 {
			continue
		}

		conID := ids[0]
		currentByClass[cls] = ids[1:]

		exec.Command("i3-msg",
			fmt.Sprintf("[con_id=%d] move container to workspace %d", conID, w.Workspace),
		).Run()

		if w.Floating {
			exec.Command("i3-msg",
				fmt.Sprintf("[con_id=%d] floating enable", conID),
			).Run()
			if w.Rect.Width > 0 && w.Rect.Height > 0 {
				exec.Command("i3-msg",
					fmt.Sprintf("[con_id=%d] resize set %d %d", conID, w.Rect.Width, w.Rect.Height),
				).Run()
				if w.Rect.X >= 0 && w.Rect.Y >= 0 {
					exec.Command("i3-msg",
						fmt.Sprintf("[con_id=%d] move position %d %d", conID, w.Rect.X, w.Rect.Y),
					).Run()
				}
			}
		}
	}

	exec.Command("i3-msg", "workspace 1").Run()
	os.Remove(path)
	return nil
}

// GracefulShutdown closes applications cleanly, saves the window layout,
// locks the screen, and shuts down or reboots the system.
func GracefulShutdown(action string) error {
	gracefulCloseWindows()

	notify.SendNotify("power", "Power", "Saving window layout...", "normal", 2000, 0)
	SaveLayout()

	notify.SendNotify("power", "Power", "Locking...", "normal", 2000, 0)
	lock.Lock()

	var notifyMsg string
	var shutdownArgs []string
	switch action {
	case "reboot":
		notifyMsg = "Rebooting..."
		shutdownArgs = []string{"shutdown", "-r", "now"}
	case "shutdown":
		notifyMsg = "Shutting down..."
		shutdownArgs = []string{"shutdown", "-p", "now"}
	default:
		return fmt.Errorf("unknown power action: %s", action)
	}

	notify.SendNotify("power", "Power", notifyMsg, "normal", 3000, 0)
	exec.Command("doas", shutdownArgs...).Run()
	return nil
}

// gracefulCloseWindows sends close events to all application windows and
// waits up to 10 seconds for them to close.
func gracefulCloseWindows() {
	notify.SendNotify("power", "Power", "Closing applications...", "normal", 3000, 0)

	tree, err := getTree()
	if err != nil {
		return
	}

	var conIDs []int64
	var wsNum int
	collectConIDs(tree, &conIDs, &wsNum)

	for _, id := range conIDs {
		exec.Command("i3-msg", fmt.Sprintf("[con_id=%d] kill", id)).Run()
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		tree, err := getTree()
		if err != nil {
			return
		}
		var remaining []int64
		collectConIDs(tree, &remaining, &wsNum)
		if len(remaining) == 0 {
			return
		}
	}
}

func collectConIDs(node *i3Node, ids *[]int64, currentWS *int) {
	if node == nil {
		return
	}
	if node.Type == "workspace" && node.Num > 0 {
		*currentWS = node.Num
	}
	if node.Type == "con" && node.Window != 0 {
		class := strings.ToLower(node.WindowProps.Class)
		instance := strings.ToLower(node.WindowProps.Instance)
		if !shutdownSkipClasses[class] && !shutdownSkipClasses[instance] {
			*ids = append(*ids, node.ID)
		}
	}
	for i := range node.Nodes {
		collectConIDs(&node.Nodes[i], ids, currentWS)
	}
	for i := range node.FloatingNodes {
		collectConIDs(&node.FloatingNodes[i], ids, currentWS)
	}
}

func collectForSnapshot(node *i3Node, windows *[]WindowInfo, currentWS *int, floating bool) {
	if node == nil {
		return
	}
	if node.Type == "workspace" && node.Num > 0 {
		*currentWS = node.Num
	}
	if node.Type == "con" && node.Window != 0 {
		class := strings.ToLower(node.WindowProps.Class)
		instance := strings.ToLower(node.WindowProps.Instance)
		if !shutdownSkipClasses[class] && !shutdownSkipClasses[instance] {
			*windows = append(*windows, WindowInfo{
				Class:     node.WindowProps.Class,
				Instance:  node.WindowProps.Instance,
				Workspace: *currentWS,
				Rect:      node.Rect,
				Floating:  floating,
				Focused:   node.Focused,
			})
		}
	}
	for i := range node.Nodes {
		collectForSnapshot(&node.Nodes[i], windows, currentWS, floating)
	}
	for i := range node.FloatingNodes {
		collectForSnapshot(&node.FloatingNodes[i], windows, currentWS, true)
	}
}

type windowPlacement struct {
	ID       int64
	Class    string
	Instance string
}

func collectForPlacement(node *i3Node, windows *[]windowPlacement, currentWS *int) {
	if node == nil {
		return
	}
	if node.Type == "workspace" && node.Num > 0 {
		*currentWS = node.Num
	}
	if node.Type == "con" && node.Window != 0 {
		class := strings.ToLower(node.WindowProps.Class)
		instance := strings.ToLower(node.WindowProps.Instance)
		if !shutdownSkipClasses[class] && !shutdownSkipClasses[instance] {
			*windows = append(*windows, windowPlacement{
				ID:       node.ID,
				Class:    node.WindowProps.Class,
				Instance: node.WindowProps.Instance,
			})
		}
	}
	for i := range node.Nodes {
		collectForPlacement(&node.Nodes[i], windows, currentWS)
	}
	for i := range node.FloatingNodes {
		collectForPlacement(&node.FloatingNodes[i], windows, currentWS)
	}
}

// loadRestoreCommands parses apps.txt to build a class→command mapping
// for restoring applications.
func loadRestoreCommands() map[string]string {
	cmds := make(map[string]string)

	appsFile := filepath.Join(paths.ConfigDir("rofi"), "apps.txt")
	f, err := os.Open(appsFile)
	if err != nil {
		return cmds
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}
		cmdLine := strings.TrimSpace(parts[1])
		if cmdLine == "" || strings.HasPrefix(cmdLine, "@") {
			continue
		}

		key := ""
		if idx := strings.Index(cmdLine, "--class"); idx != -1 {
			after := cmdLine[idx+len("--class"):]
			after = strings.TrimLeft(after, " \t=")
			fields := strings.Fields(after)
			if len(fields) > 0 {
				key = strings.ToLower(fields[0])
			}
		}
		if key == "" {
			key = strings.ToLower(strings.Fields(cmdLine)[0])
		}
		if key != "" && cmds[key] == "" {
			cmds[key] = cmdLine
		}
	}
	return cmds
}
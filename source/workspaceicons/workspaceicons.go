package workspaceicons

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"openriot/windowicon"
)

// i3 tree structures
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

// Get returns formatted workspace icons for polybar
func Get(wsNum int) string {
	// Get window icons (all at once)
	windowIcons := windowicon.GetAllWindowIcons()

	// Get windows for this workspace
	windowClasses := getWindowClasses(wsNum)

	// Get workspace state
	wsState := getWorkspaceState(wsNum)

	// Build icons string
	var icons []string
	for _, cls := range windowClasses {
		if icon, ok := windowIcons[cls]; ok {
			icons = append(icons, icon)
		}
	}

	// Determine indicator
	indicator := getIndicator(wsState, len(icons) > 0)

	// If unfocused with icons, dim them
	if wsState == "unfocused" && len(icons) > 0 {
		var dimmed []string
		for _, icon := range icons {
			dimmed = append(dimmed, fmt.Sprintf("%%{T0}%%{F#565f89}%s%%{F-}%%{T-}", icon))
		}
		icons = dimmed
	}

	// If unfocused, dim indicator
	if wsState == "unfocused" {
		indicator = fmt.Sprintf("%%{F#565f89}%s%%{F-}", indicator)
	}

	// Build output
	if len(icons) > 0 {
		return fmt.Sprintf("%s %s", indicator, strings.Join(icons, " "))
	}
	return indicator
}

func getWindowClasses(wsNum int) []string {
	cmd := exec.Command("i3-msg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var tree i3Tree
	if err := json.Unmarshal(output, &tree); err != nil {
		return nil
	}

	var classes []string
	findWindowsInWorkspace(tree.Nodes, wsNum, &classes)
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
		*classes = append(*classes, node.WindowProps.Class)
	}
	for _, n := range node.Nodes {
		collectWindows(n, classes)
	}
	for _, n := range node.FloatingNodes {
		collectWindows(n, classes)
	}
}

func getWorkspaceState(wsNum int) string {
	cmd := exec.Command("i3-msg", "-t", "get_workspaces")
	output, err := cmd.Output()
	if err != nil {
		return "unfocused"
	}

	var workspaces []i3Workspace
	if err := json.Unmarshal(output, &workspaces); err != nil {
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

// GetAll returns icons for workspaces 1-4 with a single i3-msg call each.
// Output uses polybar click markup for workspace switching:
//   %{A:openriot --workspace-switch N:}content%{A}
// Each workspace is fixed-width: indicator + 5 icon slots (padded with spaces)
func GetAll() string {
	// Single i3-msg call for all workspace windows
	windowIcons := windowicon.GetAllWindowIcons()
	tree := getFullTree()
	workspaces := getAllWorkspaces()

	var results []string
	for wsNum := 1; wsNum <= 3; wsNum++ {
		// Get windows for this workspace from cached tree
		classes := findWindowsInWorkspaceFromTree(tree.Nodes, wsNum)
		// Get workspace state from cached workspaces
		state := getStateFromWorkspaces(workspaces, wsNum)

		// Build icons for this workspace from window classes (max 5)
		var icons []string
		for _, cls := range classes {
			if icon, ok := windowIcons[cls]; ok {
				icons = append(icons, icon)
			}
		}
		if len(icons) > 5 {
			icons = icons[:5]
		}

		// Build fixed-width content: indicator + ALWAYS 5 icon slots
		allSlots := make([]string, 0, 5)
		for _, icon := range icons {
			allSlots = append(allSlots, icon)
		}
		// Pad to exactly 5 slots with Nerd Font blank icon
		for len(allSlots) < 5 {
			allSlots = append(allSlots, "\uec03")
		}

		// Determine indicator
		indicator := getIndicator(state, len(icons) > 0)

		// If unfocused with icons, dim them
		if state == "unfocused" && len(icons) > 0 {
			for i := 0; i < len(icons); i++ {
				allSlots[i] = fmt.Sprintf("%%{T0}%%{F#565f89}%s%%{F-}%%{T-}", icons[i])
			}
		}

		// If unfocused, dim indicator
		if state == "unfocused" {
			indicator = fmt.Sprintf("%%{F#565f89}%s%%{F-}", indicator)
		}

		content := fmt.Sprintf("%s %s", indicator, strings.Join(allSlots, " "))

		// Wrap with polybar click action
		results = append(results, fmt.Sprintf("%%{A:$HOME/.local/share/openriot/install/openriot --workspace-switch %d:}%s%%{A}", wsNum, content))
	}

	return strings.Join(results, "   ")
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

package windowtitle

import (
	"encoding/json"
	"os/exec"
)

const maxTitleLen = 50

func Get() string {
	cmd := exec.Command("i3-msg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	var tree i3Tree
	if err := json.Unmarshal(output, &tree); err != nil {
		return ""
	}

	// Find focused workspace first
	for _, node := range tree.Nodes {
		if node.Type == "workspace" && node.Focused {
			title := findFocusedTitle(node)
			if title != "" {
				return formatTitle(title)
			}
		}
	}

	// Fallback: find any focused window
	return formatTitle(findFocusedTitleInNodes(tree.Nodes))
}

func findFocusedTitle(node i3Node) string {
	if node.Focused && node.Type == "con" && node.Window != 0 && node.Name != "" {
		return node.Name
	}
	for _, n := range node.Nodes {
		if title := findFocusedTitle(n); title != "" {
			return title
		}
	}
	for _, n := range node.FloatingNodes {
		if title := findFocusedTitle(n); title != "" {
			return title
		}
	}
	return ""
}

func findFocusedTitleInNodes(nodes []i3Node) string {
	for _, n := range nodes {
		if title := findFocusedTitle(n); title != "" {
			return title
		}
	}
	return ""
}

func formatTitle(title string) string {
	if len(title) > maxTitleLen {
		return " " + title[:maxTitleLen-3] + "..."
	}
	return " " + title
}

type i3Tree struct {
	Nodes          []i3Node `json:"nodes"`
	FloatingNodes  []i3Node `json:"floating_nodes"`
}

type i3Node struct {
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	Focused       bool     `json:"focused"`
	Window        int      `json:"window"`
	Nodes         []i3Node `json:"nodes"`
	FloatingNodes []i3Node `json:"floating_nodes"`
}

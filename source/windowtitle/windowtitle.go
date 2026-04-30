package windowtitle

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"unicode"
)

var maxTitleLen = 36

func SetMaxLen(n int) {
	maxTitleLen = n
}

func GetMaxLen() int {
	return maxTitleLen
}

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
	title = stripEmojis(title)
	runes := []rune(title)
	if len(runes) > maxTitleLen {
		trunc := maxTitleLen - 3
		if trunc < 1 {
			trunc = 1
		}
		return fmt.Sprintf("%-*s", trunc, string(runes[:trunc])) + "..."
	}
	return fmt.Sprintf("%-*s", maxTitleLen, string(runes))
}

// stripEmojis removes emojis and problematic unicode from title
func stripEmojis(s string) string {
	runes := []rune(s)
	var cleaned []rune
	for _, r := range runes {
		// Skip emoji ranges (U+1F300 to U+1F9FF)
		if r >= 0x1F300 && r <= 0x1F9FF {
			continue
		}
		// Skip emoji modifiers and other problematic ranges
		if r >= 0x2600 && r <= 0x26FF { // Miscellaneous symbols
			continue
		}
		// Keep printable ASCII and common unicode (letters, numbers, basic punctuation)
		if r >= 32 && r <= 126 {
			cleaned = append(cleaned, r)
		} else if r > 127 && r < 0x11000 {
			// Keep other printable unicode (letters, etc)
			if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r) || unicode.IsSpace(r) {
				cleaned = append(cleaned, r)
			}
		}
	}
	return string(cleaned)
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

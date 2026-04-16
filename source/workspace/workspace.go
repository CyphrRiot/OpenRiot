package workspace

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"openriot/paths"
)

func GetCurrent() int {
	cmd := exec.Command("i3-msg", "-t", "get_workspaces")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	var workspaces []struct {
		Num     int  `json:"num"`
		Focused bool `json:"focused"`
	}

	if err := json.Unmarshal(output, &workspaces); err != nil {
		return 0
	}

	for _, ws := range workspaces {
		if ws.Focused {
			return ws.Num
		}
	}
	return 0
}

func Switch(target int) {
	current := GetCurrent()
	if current == target {
		return // Already on this workspace
	}

	// Switch via i3-msg
	cmd := exec.Command("i3-msg", fmt.Sprintf("workspace %d", target))
	cmd.Run()

	// Notify
	iconName := fmt.Sprintf("workspace%d.png", target)
	exec.Command("/usr/local/bin/notify-send", "-i", paths.GetIconPath(iconName), "-t", "1500", "Workspace", fmt.Sprintf("Switched to workspace %d", target)).Start()
}

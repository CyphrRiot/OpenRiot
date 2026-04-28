package workspace

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"openriot/notify"
)

const maxWorkspaces = 3

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

func isValid(target int) bool {
	return target >= 1 && target <= maxWorkspaces
}

func notifyUnavailable(target int) {
	notify.SendNotify("dialog-error", "Workspace", fmt.Sprintf("Workspace %d not available", target), "normal", 2000, 0)
}

func Switch(target int) {
	if !isValid(target) {
		notifyUnavailable(target)
		return
	}

	current := GetCurrent()
	iconName := fmt.Sprintf("workspace%d.png", target)
	if current == target {
		notify.SendNotify(iconName, "Workspace", fmt.Sprintf("On workspace %d", target), "normal", 1500, 0)
		return // Already on this workspace
	}

	// Switch via i3-msg
	cmd := exec.Command("i3-msg", fmt.Sprintf("workspace %d", target))
	cmd.Run()

	// Notify
	notify.SendNotify(iconName, "Workspace", fmt.Sprintf("Switched to workspace %d", target), "normal", 1500, 0)
}

func Move(target int) {
	if !isValid(target) {
		notifyUnavailable(target)
		return
	}

	// Move container and switch
	exec.Command("i3-msg", fmt.Sprintf("move container to workspace %d", target)).Run()
	exec.Command("i3-msg", fmt.Sprintf("workspace %d", target)).Run()

	iconName := fmt.Sprintf("workspace%d.png", target)
	notify.SendNotify(iconName, "Workspace", fmt.Sprintf("Moved to workspace %d", target), "normal", 1500, 0)
}

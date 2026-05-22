package gurk

import (
	"fmt"
	"os"
	"strings"

	"openriot/paths"
)

const configPath = "gurk/gurk.toml"
const sectionName = "[keybindings.message_selected]"

var keybindings = `
[keybindings.message_selected]
alt-e = "edit_message"
alt-y = "copy_message selected"
ctrl-t = "react :thumbsup:"
ctrl-f = "react 🔥"
ctrl-h = "react :purple_heart:"`

// Run is the main entry point for the --gurk-setup command
func Run() error {
	configFile := paths.Join(".config", configPath)

	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("[SKIP] ~/.config/gurk/gurk.toml not found")
		return nil
	}

	// Read file content
	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", configFile, err)
	}

	// Check if section already exists
	if strings.Contains(string(content), sectionName) {
		fmt.Println("[SKIP] gurk keybindings already exist")
		return nil
	}

	// Append keybindings
	newContent := string(content) + keybindings + "\n"
	if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configFile, err)
	}

	fmt.Println("[DONE] Added gurk keybindings")
	return nil
}

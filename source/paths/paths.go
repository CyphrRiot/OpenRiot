package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// GetIconPath returns absolute path to an icon file
// Falls back to generic info icon if the requested icon doesn't exist
// Automatically appends .png if no extension provided
func GetIconPath(filename string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	// Append .png if no extension provided
	if !strings.Contains(filename, ".") {
		filename = filename + ".png"
	}
	iconDir := filepath.Join(home, ".local", "share", "openriot", "config", "icons")
	path := filepath.Join(iconDir, filename)
	if _, err := os.Stat(path); err != nil {
		return filepath.Join(iconDir, "info.png")
	}
	return path
}
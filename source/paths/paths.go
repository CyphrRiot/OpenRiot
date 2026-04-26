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
	home := os.Getenv("HOME")
	// Append .png if no extension provided
	if !strings.Contains(filename, ".") {
		filename = filename + ".png"
	}
	path := filepath.Join(home, ".local/share/openriot/config/icons", filename)
	if _, err := os.Stat(path); err != nil {
		// Icon doesn't exist, return generic fallback
		return filepath.Join(home, ".local/share/openriot/config/icons", "info.png")
	}
	return path
}
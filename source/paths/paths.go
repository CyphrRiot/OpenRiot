package paths

import (
	"os"
	"path/filepath"
)

// GetIconPath returns absolute path to an icon file
// Falls back to generic info icon if the requested icon doesn't exist
func GetIconPath(filename string) string {
	home := os.Getenv("HOME")
	path := filepath.Join(home, ".local/share/openriot/config/icons", filename)
	if _, err := os.Stat(path); err != nil {
		// Icon doesn't exist, return generic fallback
		return filepath.Join(home, ".local/share/openriot/config/icons", "info.png")
	}
	return path
}

// GetInstallDir returns the installation directory
func GetInstallDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "openriot")
}
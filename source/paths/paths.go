package paths

import (
	"os"
	"path/filepath"
)

// GetIconPath returns absolute path to an icon file
func GetIconPath(filename string) string {
	return filepath.Join(os.Getenv("HOME"), ".local/share/openriot/config/icons", filename)
}

// GetInstallDir returns the installation directory
func GetInstallDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "openriot")
}
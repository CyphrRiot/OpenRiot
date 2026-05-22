package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// HomeDir returns the user's home directory.
func HomeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return ""
}

// Join joins path elements under the user's home directory.
func Join(elem ...string) string {
	return filepath.Join(append([]string{HomeDir()}, elem...)...)
}

// OpenRiotDir returns a path under ~/.local/share/openriot.
func OpenRiotDir(elem ...string) string {
	return filepath.Join(
		append([]string{HomeDir(), ".local", "share", "openriot"}, elem...)...,
	)
}

// ConfigDir returns a path under ~/.local/share/openriot/config.
func ConfigDir(elem ...string) string {
	return filepath.Join(
		append([]string{HomeDir(), ".local", "share", "openriot", "config"}, elem...)...,
	)
}

// IconDir returns the icons directory path.
func IconDir() string {
	return ConfigDir("icons")
}

// GetIconPath returns the absolute path to an icon file.
// Falls back to the generic info icon if the requested icon does not exist.
// Automatically appends .png if no extension is provided.
func GetIconPath(filename string) string {
	if filepath.Ext(filename) == "" {
		filename += ".png"
	}
	path := filepath.Join(IconDir(), filename)
	if _, err := os.Stat(path); err != nil {
		return filepath.Join(IconDir(), "info.png")
	}
	return path
}

// WithExt returns filename with .png appended only if it has no extension.
func WithExt(filename string) string {
	if filepath.Ext(filename) == "" {
		return filename + ".png"
	}
	return filename
}

// ExpandTilde replaces a leading "~/" with the user's home directory.
func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(HomeDir(), path[2:])
	}
	return path
}

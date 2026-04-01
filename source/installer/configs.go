package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"openriot/logger"
)

// CopyConfigs copies configuration files from the repo to user's home directory
func CopyConfigs(repoDir string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configSourceDir := filepath.Join(repoDir, "config")

	logger.LogMessage("INFO", fmt.Sprintf("Copying configs from: %s", configSourceDir))

	// List of configs to copy (source relative to config/, dest relative to ~/.config/)
	configs := []struct {
		source string
		dest   string
	}{
		// Shell
		{"fish/config.fish", "fish/config.fish"},

		// Neovim
		{"nvim/init.lua", "nvim/init.lua"},
		{"nvim/lazyvim.json", "nvim/lazyvim.json"},

		// Terminal
		{"foot/cypherriot.ini", "foot/cypherriot.ini"},

		// Desktop
		{"sway/config", "sway/config"},
		{"sway/keybindings.conf", "sway/keybindings.conf"},
		{"waybar/config", "waybar/config"},
	}

	configDir := filepath.Join(homeDir, ".config")

	// Create ~/.config if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Copy each config
	for _, cfg := range configs {
		srcPath := filepath.Join(configSourceDir, cfg.source)
		destPath := filepath.Join(configDir, cfg.dest)

		// Skip if source doesn't exist
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			logger.LogMessage("INFO", fmt.Sprintf("Skipping %s (not found)", cfg.source))
			continue
		}

		// Create destination directory
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Failed to create directory %s: %v", destDir, err))
			continue
		}

		// Copy file
		if err := copyFile(srcPath, destPath); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Failed to copy %s: %v", cfg.source, err))
			continue
		}

		logger.LogMessage("INFO", fmt.Sprintf("Copied %s", cfg.dest))
	}

	// Copy backgrounds to ~/.local/share/openriot/backgrounds
	if err := copyBackgrounds(repoDir, homeDir); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Background copy failed: %v", err))
	}

	return nil
}

// copyBackgrounds copies background images to the backgrounds directory
func copyBackgrounds(repoDir, homeDir string) error {
	bgSourceDir := filepath.Join(repoDir, "backgrounds")
	bgDestDir := filepath.Join(homeDir, ".local", "share", "openriot", "backgrounds")

	// Create destination directory
	if err := os.MkdirAll(bgDestDir, 0755); err != nil {
		return fmt.Errorf("creating backgrounds directory: %w", err)
	}

	// Check if source exists
	if _, err := os.Stat(bgSourceDir); os.IsNotExist(err) {
		logger.LogMessage("INFO", "No backgrounds directory found")
		return nil
	}

	// Copy all jpg files
	entries, err := os.ReadDir(bgSourceDir)
	if err != nil {
		return fmt.Errorf("reading backgrounds directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".jpg") {
			continue
		}

		srcPath := filepath.Join(bgSourceDir, name)
		destPath := filepath.Join(bgDestDir, name)

		if err := copyFile(srcPath, destPath); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Failed to copy background %s: %v", name, err))
			continue
		}

		logger.LogMessage("INFO", fmt.Sprintf("Copied background %s", name))
	}

	return nil
}

// copyFile copies a single file
func copyFile(source, dest string) error {
	sourceData, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	if err := os.WriteFile(dest, sourceData, 0644); err != nil {
		return fmt.Errorf("writing dest file: %w", err)
	}

	return nil
}

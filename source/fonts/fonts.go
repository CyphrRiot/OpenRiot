package fonts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const fontMarker = "FiraCodeNerdFont-Regular.ttf"
const fontSourceDir = "assets/fonts"
const fontDestDir = ".local/share/fonts"

// Run installs Nerd Fonts
func Run() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home: %w", err)
	}

	// Check if fonts already installed (check for marker file)
	destPath := filepath.Join(homeDir, fontDestDir, fontMarker)
	if _, err := os.Stat(destPath); err == nil {
		fmt.Println("[SKIP] Fonts already installed")
		return nil
	}

	// Source fonts
	sourcePath := filepath.Join(homeDir, ".local/share/openriot", fontSourceDir)
	sourceFonts, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("font source not found: %w", err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Join(homeDir, fontDestDir), 0755); err != nil {
		return fmt.Errorf("failed to create font directory: %w", err)
	}

	// Copy all fonts
	for _, f := range sourceFonts {
		if f.IsDir() {
			continue
		}
		src := filepath.Join(sourcePath, f.Name())
		dst := filepath.Join(homeDir, fontDestDir, f.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		os.WriteFile(dst, data, 0644)
	}

	// Refresh font cache
	cmd := exec.Command("fc-cache", "-f", filepath.Join(homeDir, fontDestDir))
	cmd.Run() // Ignore errors - fonts still installed

	fmt.Println("[DONE] Nerd Fonts installed")
	return nil
}

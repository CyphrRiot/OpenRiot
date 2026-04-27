package fonts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const fontSourceDir = "assets/fonts"
const fontDestDir = ".local/share/fonts"

// Run installs Nerd Fonts and refreshes the font cache
func Run() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home: %w", err)
	}

	sourcePath := filepath.Join(homeDir, ".local/share/openriot", fontSourceDir)
	sourceFonts, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("font source not found: %w", err)
	}

	destPath := filepath.Join(homeDir, fontDestDir)
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create font directory: %w", err)
	}

	copied := 0
	for _, f := range sourceFonts {
		if f.IsDir() {
			continue
		}
		src := filepath.Join(sourcePath, f.Name())
		dst := filepath.Join(destPath, f.Name())

		if _, err := os.Stat(dst); err == nil {
			continue
		}

		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			continue
		}
		copied++
	}

	// Always refresh font cache - fixes stale cache after reboot
	cmd := exec.Command("fc-cache", "-f", destPath)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] fc-cache failed: %v\n", err)
	}

	if copied > 0 {
		fmt.Printf("[DONE] Installed %d font(s), cache refreshed\n", copied)
	} else {
		fmt.Println("[DONE] All fonts present, cache refreshed")
	}
	return nil
}

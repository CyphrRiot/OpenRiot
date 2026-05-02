package assets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const iconDest = "icons"
const cursorDest = "/usr/local/share/icons"

// Run installs assets like cursors and icon themes
func Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: openriot --install-asset <name>\nSupported: bibata, kora")
	}

	name := args[0]
	switch name {
	case "bibata":
		return installBibata()
	case "kora":
		return installKora()
	default:
		return fmt.Errorf("unknown asset: %s\nSupported: bibata, kora", name)
	}
}

func installBibata() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home: %w", err)
	}

	installedPath := filepath.Join(cursorDest, "Bibata-Modern-Ice")
	sourcePath := filepath.Join(homeDir, ".local/share/openriot/assets/cursors/Bibata-Modern-Ice")

	// Check if already installed
	if _, err := os.Stat(installedPath); err == nil {
		fmt.Println("[SKIP] Bibata cursor already installed")
		return nil
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		fmt.Println("[SKIP] Bibata source not found (run setup.sh first)")
		return nil
	}

	// Copy to system icons directory
	cmd := exec.Command("doas", "cp", "-r", sourcePath, cursorDest)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy Bibata: %w", err)
	}

	fmt.Println("[DONE] Bibata cursor installed")
	return nil
}

func installKora() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home: %w", err)
	}

	destPath := filepath.Join(homeDir, ".local/share/icons/kora")

	// Check if already installed
	if _, err := os.Stat(destPath); err == nil {
		fmt.Println("[SKIP] Kora icon theme already installed")
		return nil
	}

	// Extract bundled kora.tgz from assets to ~/.local/share/icons/
	sourceTgz := filepath.Join(homeDir, ".local/share/openriot/assets/themes/kora.tgz")
	iconsDir := filepath.Join(homeDir, ".local/share/icons")

	if _, err := os.Stat(sourceTgz); os.IsNotExist(err) {
		fmt.Println("[SKIP] Kora theme archive not found (run setup.sh first)")
		return nil
	}

	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return fmt.Errorf("failed to create icons directory: %w", err)
	}

	cmd := exec.Command("tar", "xzf", sourceTgz, "-C", iconsDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract Kora theme: %w", err)
	}

	fmt.Println("[DONE] Kora icon theme installed")
	return nil
}

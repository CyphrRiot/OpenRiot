package assets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		fmt.Println("[SKIP] Bibata source not found (run setup.sh first)")
		return nil
	}

	// Check if already installed — but also verify it is not the broken quoted version
	needsInstall := true
	if _, err := os.Stat(installedPath); err == nil {
		indexPath := filepath.Join(installedPath, "index.theme")
		if data, err := os.ReadFile(indexPath); err == nil {
			if !strings.Contains(string(data), `Inherits="`) {
				needsInstall = false
			}
		}
	}

	if needsInstall {
		// Remove broken or stale system install first
		if _, err := os.Stat(installedPath); err == nil {
			cmd := exec.Command("doas", "rm", "-rf", installedPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to remove stale Bibata: %w", err)
			}
		}

		// Copy to system icons directory
		cmd := exec.Command("doas", "cp", "-r", sourcePath, cursorDest)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy Bibata: %w", err)
		}
	}

	// Remove stale local copies that shadow the system install
	localStale := filepath.Join(homeDir, ".local/share/icons/Bibata-Modern-Ice")
	if _, err := os.Stat(localStale); err == nil {
		if err := os.RemoveAll(localStale); err != nil {
			return fmt.Errorf("failed to remove stale local Bibata: %w", err)
		}
	}

	dotIconsStale := filepath.Join(homeDir, ".icons/Bibata-Modern-Ice")
	if _, err := os.Stat(dotIconsStale); err == nil {
		if err := os.RemoveAll(dotIconsStale); err != nil {
			return fmt.Errorf("failed to remove stale .icons Bibata: %w", err)
		}
	}

	if needsInstall {
		fmt.Println("[DONE] Bibata cursor installed")
	} else {
		fmt.Println("[SKIP] Bibata cursor already installed")
	}
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

package assets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/paths"
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
	installedPath := filepath.Join(cursorDest, "Bibata-Modern-Ice")
	sourcePath := paths.OpenRiotDir("assets", "cursors", "Bibata-Modern-Ice")

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
	localStale := paths.Join(".local", "share", "icons", "Bibata-Modern-Ice")
	if _, err := os.Stat(localStale); err == nil {
		if err := os.RemoveAll(localStale); err != nil {
			return fmt.Errorf("failed to remove stale local Bibata: %w", err)
		}
	}

	dotIconsStale := paths.Join(".icons", "Bibata-Modern-Ice")
	if _, err := os.Stat(dotIconsStale); err == nil {
		if err := os.RemoveAll(dotIconsStale); err != nil {
			return fmt.Errorf("failed to remove stale .icons Bibata: %w", err)
		}
	}

	// ALWAYS create ~/.icons/default fallback (Firefox/Electron need this)
	defaultDir := paths.Join(".icons", "default")
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		return fmt.Errorf("failed to create ~/.icons/default: %w", err)
	}

	fallbackTheme := filepath.Join(defaultDir, "index.theme")
	content := "[Icon Theme]\nName=Default\nComment=Default cursor theme\nInherits=Bibata-Modern-Ice\n"
	if err := os.WriteFile(fallbackTheme, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write ~/.icons/default/index.theme: %w", err)
	}

	if needsInstall {
		fmt.Println("[DONE] Bibata cursor installed")
	} else {
		fmt.Println("[SKIP] Bibata cursor already installed")
	}
	fmt.Println("[DONE] Cursor fallback set: ~/.icons/default -> Bibata-Modern-Ice")
	return nil
}

func installKora() error {
	destPath := paths.Join(".local", "share", "icons", "kora")
	panelPath := filepath.Join(destPath, "panel", "22")
	actionsPath := filepath.Join(destPath, "actions", "16")

	// Check if already installed by looking for key fixed-size directories.
	// This is more robust than checking index.theme, which can be reverted
	// by GTK cache rebuilds or other tools without removing the actual files.
	needsInstall := false
	if _, err := os.Stat(panelPath); os.IsNotExist(err) {
		needsInstall = true
	}
	if _, err := os.Stat(actionsPath); os.IsNotExist(err) {
		needsInstall = true
	}

	if needsInstall {
		// Remove broken or stale install first
		if _, err := os.Stat(destPath); err == nil {
			if err := os.RemoveAll(destPath); err != nil {
				return fmt.Errorf("failed to remove stale Kora: %w", err)
			}
		}
	}

	// Extract bundled kora.tgz from assets to ~/.local/share/icons/
	sourceTgz := paths.OpenRiotDir("assets", "themes", "kora.tgz")
	iconsDir := paths.Join(".local", "share", "icons")

	if _, err := os.Stat(sourceTgz); os.IsNotExist(err) {
		fmt.Println("[SKIP] Kora theme archive not found (run setup.sh first)")
		return nil
	}

	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return fmt.Errorf("failed to create icons directory: %w", err)
	}

	if needsInstall {
		fmt.Println("[INFO] Installing Kora icon theme... (be patient)")
		cmd := exec.Command("tar", "xzf", sourceTgz, "-C", iconsDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract Kora theme: %w", err)
		}
		fmt.Println("[DONE] Kora icon theme installed")
	} else {
		fmt.Println("[SKIP] Kora icon theme already installed")
	}
	return nil
}

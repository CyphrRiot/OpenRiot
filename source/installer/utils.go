package installer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ShareLog uploads a file to tmpfiles.org for easy sharing
func ShareLog(filename string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get home dir: %w", err)
	}
	logPath := filepath.Join(homeDir, ".cache", "openriot", filename)

	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("reading log file: %w", err)
	}

	// Upload to tmpfiles.org
	cmd := exec.Command("curl", "-s", "-F", "file=@-", "https://tmpfiles.org/api/v1/upload")
	cmd.Stdin = bytes.NewReader(data)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	url := strings.TrimSpace(string(output))
	fmt.Println(url)
	return nil
}

// findFontPath locates the FiraCode Nerd Font in multiple locations
func findFontPath() string {
	// 1. Bundled font relative to binary (production: install/openriot -> ../assets/fonts/)
	if ex, err := os.Executable(); err == nil {
		bundled := filepath.Join(filepath.Dir(ex), "..", "assets", "fonts", "FiraCodeNerdFont-Regular.ttf")
		if _, err := os.Stat(bundled); err == nil {
			return bundled
		}
	}

	// 2. Same-directory assets (development fallback)
	if ex, err := os.Executable(); err == nil {
		sameDir := filepath.Join(filepath.Dir(ex), "assets", "fonts", "FiraCodeNerdFont-Regular.ttf")
		if _, err := os.Stat(sameDir); err == nil {
			return sameDir
		}
	}

	// 3. User's local font installation
	if home, err := os.UserHomeDir(); err == nil {
		local := filepath.Join(home, ".local", "share", "fonts", "FiraCode", "FiraCodeNerdFont-Regular.ttf")
		if _, err := os.Stat(local); err == nil {
			return local
		}
	}

	return ""
}

// MakeIcon generates a PNG icon from a Nerd Font symbol
func MakeIcon(name, symbol string) error {
	font := findFontPath()
	if font == "" {
		return fmt.Errorf("FiraCode Nerd Font not found (checked bundled assets and ~/.local/share/fonts/FiraCode/)")
	}

	iconDir := ""
	if ex, err := os.Executable(); err == nil {
		repoDir := filepath.Join(filepath.Dir(ex), "..", "config", "icons")
		if _, err := os.Stat(repoDir); err == nil {
			iconDir = repoDir
		}
	}
	if iconDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get home dir: %w", err)
		}
		iconDir = filepath.Join(home, ".local", "share", "openriot", "config", "icons")
	}

	// Ensure icon directory exists
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		return fmt.Errorf("creating icon dir: %w", err)
	}

	output := filepath.Join(iconDir, name+".png")
	cmd := exec.Command("convert",
		"-background", "none",
		"-fill", "white",
		"-font", font,
		"-pointsize", "32",
		"-gravity", "center",
		"label:"+symbol,
		"-extent", "48x48",
		output)
	return cmd.Run()
}
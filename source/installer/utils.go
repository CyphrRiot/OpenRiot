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
	homeDir, _ := os.UserHomeDir()
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

// MakeIcon generates a PNG icon from a Nerd Font symbol
func MakeIcon(name, symbol string) error {
	home := os.Getenv("HOME")
	font := filepath.Join(home, ".local/share/fonts/FiraCode/FiraCodeNerdFont-Regular.ttf")
	iconDir := filepath.Join(home, ".local/share/openriot/config/icons")

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
		"label:"+symbol,
		"-resize", "48x48",
		output)
	return cmd.Run()
}
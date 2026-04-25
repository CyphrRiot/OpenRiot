package roficalc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const repoURL = "https://github.com/svenstaro/rofi-calc.git"
const tmpDir = "/tmp/rofi-calc"
const installPath = "/usr/local/lib/rofi/libcalc.so"

// Run builds and installs rofi-calc from source
func Run() error {
	// Check if already installed
	if _, err := os.Stat(installPath); err == nil {
		fmt.Println("[SKIP] rofi-calc already installed")
		return nil
	}

	// Clone repo
	if _, err := os.Stat(tmpDir); err == nil {
		os.RemoveAll(tmpDir)
	}
	cmd := exec.Command("git", "clone", repoURL, tmpDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Meson setup
	buildDir := filepath.Join(tmpDir, "build")
	cmd = exec.Command("meson", "setup", buildDir, "--prefix=/usr/local")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("meson setup failed: %w", err)
	}

	// Compile
	cmd = exec.Command("meson", "compile", "-C", buildDir)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("meson compile failed: %w", err)
	}

	// Install
	cmd = exec.Command("doas", "meson", "install", "-C", buildDir)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("meson install failed: %w", err)
	}

	// Cleanup
	os.RemoveAll(tmpDir)

	fmt.Println("[DONE] rofi-calc installed")
	return nil
}
package lock

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func Lock() error {
	home, _ := os.UserHomeDir()

	// Find locked.jpg (installed or repo)
	lockJpg := filepath.Join(home, ".local/share/openriot/assets/locked.jpg")
	if _, err := os.Stat(lockJpg); os.IsNotExist(err) {
		lockJpg = filepath.Join(home, "Code/OpenRiot/assets/locked.jpg")
	}

	lockPng := "/tmp/openriot-lock.png"

	if _, err := os.Stat(lockJpg); err == nil {
		// Get screen resolution
		res := getResolution()
		if res == "" {
			res = "1920x1080"
		}

		// Convert and resize
		cmd := exec.Command("convert", lockJpg, "-resize", res+"^", "-gravity", "center", "-extent", res, lockPng)
		cmd.Run()

		if _, err := os.Stat(lockPng); err == nil {
			cmd = exec.Command("i3lock", "-i", lockPng)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			cmd.Stdin = nil
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.Run() // Run (not Start) so we wait for screen unlock before cleanup
			os.Remove(lockPng)
			return nil
		}
	}

	// Fallback: solid color
	cmd := exec.Command("i3lock", "-c", "08090c")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func getResolution() string {
	cmd := exec.Command("xdpyinfo")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "dimensions:") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "dimensions:" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}
	return ""
}

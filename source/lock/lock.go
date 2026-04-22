package lock

import (
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"syscall"

	"openriot/notify"
)

func Lock() error {
	home, _ := os.UserHomeDir()

	// Check if i3lock is already running
	if checkLockRunning() {
		return nil // Already locked, skip
	}

	// Find random lock image
	lockDir := filepath.Join(home, ".local/share/openriot/Locked")

	// Look for locked_*.jpg in Locked directory
	matches, _ := filepath.Glob(filepath.Join(lockDir, "locked_*.jpg"))
	if len(matches) == 0 {
		// Fallback to old locked.jpg
		lockJpg := filepath.Join(home, ".local/share/openriot/assets/locked.jpg")
		if _, err := os.Stat(lockJpg); os.IsNotExist(err) {
			lockJpg = filepath.Join(home, "Code/OpenRiot/assets/locked.jpg")
		}
		if _, err := os.Stat(lockJpg); err == nil {
			matches = append(matches, lockJpg)
		}
	}

	if len(matches) > 0 {
		// Randomly select one
		lockJpg := matches[rand.Intn(len(matches))]

		// Notify user
		notify.SendNotify("lock", "Screen Lock", "Screen is locking...", "normal", 4000, 0)
		time.Sleep(500 * time.Millisecond)

		// Get screen resolution
		res := getResolution()
		if res == "" {
			res = "1920x1080"
		}

		// Convert to /tmp with exact resolution (centers and fills)
		lockPng := "/tmp/openriot-lock.png"
		cmd := exec.Command("convert", lockJpg, "-resize", res+"^", "-gravity", "center", "-extent", res, lockPng)
		cmd.Run()

		if _, err := os.Stat(lockPng); err == nil {
			cmd = exec.Command("i3lock", "-i", lockPng)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			cmd.Stdin = nil
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.Run() // Wait for unlock before cleanup
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

// checkLockRunning returns true if i3lock is already running
func checkLockRunning() bool {
	cmd := exec.Command("pgrep", "-x", "i3lock")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

package lock

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"openriot/macspoof"
	"openriot/notify"
	"openriot/screen"
)

func Lock() error {
	home, _ := os.UserHomeDir()

	// Check if i3lock is already running
	if checkLockRunning() {
		return nil // Already locked, skip
	}

	// Find random lock image
	lockDir := filepath.Join(home, ".local/share/openriot/Locked")

	// Check if stealth mode is enabled
	var matches []string
	if macspoof.IsStealthEnabled() {
		// In stealth: allow both default/ and stealth/
		matches, _ = filepath.Glob(filepath.Join(lockDir, "default", "*.jpg"))
		stealthMatches, _ := filepath.Glob(filepath.Join(lockDir, "stealth", "*.jpg"))
		matches = append(matches, stealthMatches...)
	} else {
		// Look for default/*.jpg in Locked directory
		matches, _ = filepath.Glob(filepath.Join(lockDir, "default", "*.jpg"))
	}

	// Holiday images are shown in both modes (if they exist)
	holidayMatches, _ := filepath.Glob(filepath.Join(lockDir, "holiday", "*.jpg"))
	matches = append(matches, holidayMatches...)

	// Retro images are shown in both modes (if they exist)
	retroMatches, _ := filepath.Glob(filepath.Join(lockDir, "retro", "*.jpg"))
	matches = append(matches, retroMatches...)

	// Product images are shown in both modes (if they exist)
	productMatches, _ := filepath.Glob(filepath.Join(lockDir, "products", "*.jpg"))
	matches = append(matches, productMatches...)

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

	if len(matches) == 0 {
		// No lock images found
		notify.SendNotify("lock", "Screen Lock", "No lock images found", "normal", 4000, 0)
	}

	if len(matches) > 0 {
		// Randomly select one
		lockJpg := matches[rand.Intn(len(matches))]

		// Notify user
		notify.SendNotify("lock", "Screen Lock", "Screen is locking...", "normal", 4000, 0)
		time.Sleep(500 * time.Millisecond)

		// Get screen resolution
		w, h := screen.GetResolution()
		res := fmt.Sprintf("%dx%d", w, h)

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
			err := cmd.Start()
			if err != nil {
				notify.SendNotify("lock", "Screen Lock", "Lock failed: i3lock error", "critical", 5000, 0)
				return err
			}
			cmd.Wait()
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
	err := cmd.Start()
	if err != nil {
		notify.SendNotify("lock", "Screen Lock", "Lock failed: i3lock error", "critical", 5000, 0)
	}
	return err
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

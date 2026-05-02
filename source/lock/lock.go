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

	// Get screen resolution
	w, h := screen.GetResolution()
	res := fmt.Sprintf("%dx%d", w, h)

	// Find random lock image
	lockDir := filepath.Join(home, ".local/share/openriot/Locked")
	cacheBase := filepath.Join(lockDir, ".cache", res)

	// Try cached PNGs first (fast path)
	matches, _ := filepath.Glob(filepath.Join(cacheBase, "*.png"))
	if macspoof.IsStealthEnabled() {
		// Stealth: use all images
	} else {
		// Non-stealth: filter out stealth images (*s.png)
		var normal []string
		for _, m := range matches {
			if !strings.HasSuffix(m, "s.png") {
				normal = append(normal, m)
			}
		}
		matches = normal
	}

	// Cache missing — build it now
	if len(matches) == 0 {
		notify.SendNotify("lock", "Screen Lock", "Building Lock screens... Expected time: ~2 mins", "normal", 120000, 0)
		if err := BuildCache(); err != nil {
			notify.SendNotify("lock", "Screen Lock", "Lock failed: could not build cache", "critical", 5000, 0)
			return err
		}

		// Retry after build
		matches, _ = filepath.Glob(filepath.Join(cacheBase, "*.png"))
		if !macspoof.IsStealthEnabled() {
			var normal []string
			for _, m := range matches {
				if !strings.HasSuffix(m, "s.png") {
					normal = append(normal, m)
				}
			}
			matches = normal
		}
	}

	if len(matches) == 0 {
		notify.SendNotify("lock", "Screen Lock", "No lock images found", "normal", 4000, 0)
		return nil
	}

	lockFile := matches[rand.Intn(len(matches))]

	// Give user time to see the notification before screen locks
	notify.SendNotify("lock", "Screen Lock", "Screen is locking...", "normal", 4000, 0)
	time.Sleep(1 * time.Second)

	cmd := exec.Command("i3lock", "-i", lockFile)
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
	return nil
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

// SmartLock prevents locking when a known video player is running.
func SmartLock() error {
	players := []string{"mpv", "vlc", "mplayer"}
	for _, p := range players {
		cmd := exec.Command("pgrep", "-x", p)
		if output, _ := cmd.Output(); len(strings.TrimSpace(string(output))) > 0 {
			return nil
		}
	}
	return Lock()
}

package lock

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"openriot/macspoof"
	"openriot/notify"
	"openriot/paths"
	"openriot/screen"
)

// filterStealth removes stealth images from matches when stealth is off.
func filterStealth(matches []string) []string {
	if macspoof.IsStealthEnabled() {
		return matches
	}
	var normal []string
	for _, m := range matches {
		if !strings.HasSuffix(m, "s.png") {
			normal = append(normal, m)
		}
	}
	return normal
}

// isProcessRunning returns true if a process with the exact name is active.
func isProcessRunning(name string) bool {
	out, _ := exec.Command("pgrep", "-x", name).Output()
	return len(strings.TrimSpace(string(out))) > 0
}

// cryptoRandIndex returns a cryptographically secure random int [0, max).
func cryptoRandIndex(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

func Lock() error {
	// Check if i3lock is already running
	if checkLockRunning() {
		return nil // Already locked, skip
	}

	// Get screen resolution
	w, h := screen.GetResolution()
	res := fmt.Sprintf("%dx%d", w, h)

	// Find random lock image
	lockDir := paths.OpenRiotDir("Locked")
	cacheBase := filepath.Join(lockDir, ".cache", res)

	// Try cached PNGs first (fast path)
	matches, _ := filepath.Glob(filepath.Join(cacheBase, "*.png"))
	matches = filterStealth(matches)

	// Cache missing — build it now
	if len(matches) == 0 {
		notify.SendNotify("lock", "Screen Lock", "Building Lock screens... Expected time: ~2 mins", "normal", 120000, 0)
		if err := BuildCache(); err != nil {
			notify.SendNotify("lock", "Screen Lock", "Lock failed: could not build cache", "critical", 5000, 0)
			return err
		}

		// Retry after build
		matches, _ = filepath.Glob(filepath.Join(cacheBase, "*.png"))
		matches = filterStealth(matches)
	}

	if len(matches) == 0 {
		notify.SendNotify("lock", "Screen Lock", "No lock images found", "normal", 4000, 0)
		return nil
	}

	lockFile := matches[cryptoRandIndex(len(matches))]

	// Give user time to see the notification before screen locks
	notify.SendNotify("lock", "Screen Lock", "Screen is locking...", "normal", 4000, 0)

	// Disable keyboard auto-repeat so i3lock doesn't see phantom keys
	// from held keys during the X11 grab transition window.
	exec.Command("xset", "r", "off").Run()
	time.Sleep(1500 * time.Millisecond)

	cmd := exec.Command("i3lock", "-n", "-i", lockFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Start()
	if err != nil {
		exec.Command("xset", "r", "on").Run()
		notify.SendNotify("lock", "Screen Lock", "Lock failed: i3lock error", "critical", 5000, 0)
		return err
	}
	cmd.Wait()
	// Re-enable auto-repeat after the screen is unlocked.
	exec.Command("xset", "r", "on").Run()
	return nil
}

// checkLockRunning returns true if i3lock is already running
func checkLockRunning() bool {
	return isProcessRunning("i3lock")
}

// SmartLock prevents locking when a known video player is running.
func SmartLock() error {
	players := []string{"mpv", "vlc", "mplayer"}
	for _, p := range players {
		if isProcessRunning(p) {
			return nil
		}
	}
	return Lock()
}

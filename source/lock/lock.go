package lock

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

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

// cryptoShuffle performs a Fisher-Yates shuffle using crypto/rand.
func cryptoShuffle(slice []string) {
	for i := len(slice) - 1; i > 0; i-- {
		j := cryptoRandIndex(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// shufflePick reads the shuffle state file, advances to the next lock screen,
// and returns the full path. If state is missing or stale, it generates a new
// shuffled order. Guarantees no repeats until all images have been shown once.
func shufflePick(cacheBase string, matches []string) string {
	stateFile := filepath.Join(cacheBase, "shuffle.state")
	oldLast := ""

	if f, err := os.Open(stateFile); err == nil {
		var names []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			names = append(names, line)
		}
		f.Close()

		if len(names) == len(matches)+1 {
			matchSet := make(map[string]bool)
			for _, m := range matches {
				matchSet[filepath.Base(m)] = true
			}
			valid := true
			for _, fn := range names[1:] {
				if !matchSet[fn] {
					valid = false
					break
				}
			}
			if valid {
				idx := 0
				if n, err := fmt.Sscanf(names[0], "%d", &idx); err == nil && n == 1 {
					if idx >= len(names)-1 {
						oldLast = names[len(names)-1]
					} else {
						filename := names[idx+1]
						for _, m := range matches {
							if filepath.Base(m) == filename {
								names[0] = fmt.Sprintf("%d", idx+1)
								if err := os.WriteFile(stateFile, []byte(strings.Join(names, "\n")+"\n"), 0644); err == nil {
									return m
								}
								break
							}
						}
					}
				}
			}
		}
	}

	// Regenerate: collect filenames, shuffle, write state
	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = filepath.Base(m)
	}
	sort.Strings(names)
	cryptoShuffle(names)

	if oldLast != "" && names[0] == oldLast && len(names) > 1 {
		names[0], names[1] = names[1], names[0]
	}

	lines := append([]string{"1"}, names...)
	_ = os.WriteFile(stateFile, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	for _, m := range matches {
		if filepath.Base(m) == names[0] {
			return m
		}
	}
	return matches[0]
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

	lockFile := shufflePick(cacheBase, matches)

	// Give user time to see the notification before screen locks
	notify.SendNotify("lock", "Screen Lock", "Screen is locking...", "normal", 4000, 0)

	exec.Command("xset", "r", "off").Run()

	cmd := exec.Command("i3lock", "-n", "-i", lockFile)
	err := cmd.Start()
	if err != nil {
		exec.Command("xset", "r", "on").Run()
		notify.SendNotify("lock", "Screen Lock", "Lock failed: i3lock error", "critical", 5000, 0)
		return err
	}
	cmd.Wait()
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

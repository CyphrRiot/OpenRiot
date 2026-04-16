package backgrounds

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)


// Load restores the last saved wallpaper or falls back to default.
func Load() int {
	home := os.Getenv("HOME")
	stateFile := filepath.Join(home, ".config", "openriot", ".current-background")
	bgsDir := filepath.Join(home, ".local", "share", "openriot", "backgrounds")
	defaultBg := filepath.Join(bgsDir, "riot_00.jpg")

	wallpaper := defaultBg
	if b, err := os.ReadFile(stateFile); err == nil {
		candidate := strings.TrimSpace(string(b))
		if _, err := os.Stat(candidate); err == nil {
			wallpaper = candidate
		}
	}

	cmd := exec.Command("feh", "--bg-fill", wallpaper)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()

	return 0
}

// Next cycles to the next wallpaper and restarts feh.
func Next() int {
	home := os.Getenv("HOME")
	bgsDir := filepath.Join(home, ".local", "share", "openriot", "backgrounds")
	stateFile := filepath.Join(home, ".config", "openriot", ".current-background")

	entries, err := os.ReadDir(bgsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Backgrounds directory not found: %s\n", bgsDir)
		return 1
	}


	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)
		if strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") ||
			strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".webp") {
			files = append(files, filepath.Join(bgsDir, name))
		}
	}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No background images found in %s\n", bgsDir)
		return 1
	}
	sort.Strings(files)

	current := ""
	if b, err := os.ReadFile(stateFile); err == nil {
		current = strings.TrimSpace(string(b))
	}

	idx := 0
	for i, f := range files {
		if f == current {
			idx = i
			break
		}
	}
	next := files[(idx+1)%len(files)]

	_ = os.MkdirAll(filepath.Dir(stateFile), 0o755)
	_ = os.WriteFile(stateFile, []byte(next+"\n"), 0o600)

	_ = exec.Command("pkill", "-x", "feh").Run()
	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("feh", "--bg-fill", next)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()

	time.Sleep(1 * time.Second)
	if exec.Command("pgrep", "-x", "feh").Run() == nil {
		fmt.Printf("Switched to: %s\n", filepath.Base(next))
		return 0
	}
	fmt.Println("Warning: feh may not have started")
	return 0
}

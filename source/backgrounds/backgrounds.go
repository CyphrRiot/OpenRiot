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

	"openriot/paths"
)

func getStateFile() string {
	newFile := paths.Join(".config", "openriot", "current-background")
	oldFile := paths.Join(".config", "openriot", ".current-background")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		if data, err := os.ReadFile(oldFile); err == nil {
			_ = os.WriteFile(newFile, data, 0o600)
			_ = os.Remove(oldFile)
		}
	}
	return newFile
}

// activeDisplayCount returns the number of currently active displays.
func activeDisplayCount() int {
	out, err := exec.Command("xrandr", "--listactivemonitors").Output()
	if err != nil {
		return 1
	}
	count := 0
	for _, line := range strings.Split(string(out), "\n") {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' {
			count++
		}
	}
	return count
}

// fehArgs builds the feh command so each active display gets its own scaled copy.
func fehArgs(wallpaper string) []string {
	args := []string{"--bg-fill"}
	n := activeDisplayCount()
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		args = append(args, wallpaper)
	}
	return args
}

// Load restores the last saved wallpaper or falls back to default.
func Load() int {
	stateFile := getStateFile()
	bgsDir := paths.OpenRiotDir("backgrounds")
	defaultBg := filepath.Join(bgsDir, "01.png")

	wallpaper := defaultBg
	if b, err := os.ReadFile(stateFile); err == nil {
		candidate := strings.TrimSpace(string(b))
		if _, err := os.Stat(candidate); err == nil {
			wallpaper = candidate
		}
	}

	cmd := exec.Command("feh", fehArgs(wallpaper)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()

	return 0
}

// cycle moves delta steps through the wallpaper list and applies the result.
func cycle(delta int) int {
	bgsDir := paths.OpenRiotDir("backgrounds")
	stateFile := getStateFile()

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
	target := files[(idx+delta+len(files))%len(files)]

	_ = os.MkdirAll(filepath.Dir(stateFile), 0o755)
	_ = os.WriteFile(stateFile, []byte(target+"\n"), 0o600)

	_ = exec.Command("pkill", "-x", "feh").Run()
	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("feh", fehArgs(target)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()

	time.Sleep(1 * time.Second)
	if exec.Command("pgrep", "-x", "feh").Run() == nil {
		fmt.Printf("Switched to: %s\n", filepath.Base(target))
		return 0
	}
	fmt.Println("Warning: feh may not have started")
	return 0
}

// Next cycles to the next wallpaper and restarts feh.
func Next() int {
	return cycle(1)
}

// Prev cycles to the previous wallpaper and restarts feh.
func Prev() int {
	return cycle(-1)
}

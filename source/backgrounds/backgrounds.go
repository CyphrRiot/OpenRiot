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


// Next cycles to the next wallpaper and restarts swaybg.
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
	_ = os.WriteFile(stateFile, []byte(next+"\n"), 0o644)

	_ = exec.Command("pkill", "-x", "swaybg").Run()
	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("swaybg", "-i", next, "-m", "fill")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()

	time.Sleep(1 * time.Second)
	if exec.Command("pgrep", "-x", "swaybg").Run() == nil {
		fmt.Printf("Switched to: %s\n", filepath.Base(next))
		return 0
	}
	fmt.Println("Warning: swaybg may not have started")
	return 0
}

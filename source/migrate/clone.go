package migrate

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	cloneOutputMu  sync.Mutex
	cloneOutput    []string
	cloneRunning   bool
	cloneCompleted bool
	cloneError     error
	cloneCancel    bool
	cloneCmd       *exec.Cmd
)

func resetCloneState() {
	cloneOutputMu.Lock()
	defer cloneOutputMu.Unlock()
	cloneCmd = nil
	cloneRunning = false
	cloneCompleted = false
	cloneError = nil
	cloneCancel = false
	cloneOutput = nil
}

func appendCloneOutput(line string) {
	cloneOutputMu.Lock()
	defer cloneOutputMu.Unlock()
	cloneOutput = append(cloneOutput, line)
}

func getCloneOutput() []string {
	cloneOutputMu.Lock()
	defer cloneOutputMu.Unlock()
	out := make([]string, len(cloneOutput))
	copy(out, cloneOutput)
	return out
}

// StartClone kicks off the clone operation (rsync + installboot) in a
// background goroutine and returns an initial progress-check command.
func StartClone() tea.Cmd {
	resetCloneState()

	cloneOutputMu.Lock()
	cloneRunning = true
	cloneOutputMu.Unlock()

	go runClone()

	return CheckCloneProgress()
}

func runClone() {
	defer func() {
		cloneOutputMu.Lock()
		cloneCompleted = true
		cloneOutputMu.Unlock()
	}()

	// ── 1. Check mount ──────────────────────────────────────────────────

	appendCloneOutput("Checking /mnt/backup mount...")
	out, err := exec.Command("mount").Output()
	if err != nil {
		cloneError = fmt.Errorf("mount: %w", err)
		return
	}
	if !strings.Contains(string(out), " on /mnt/backup ") {
		cloneError = fmt.Errorf("/mnt/backup is not mounted. Mount it first:\ndoas mount /dev/sdXi /mnt/backup")
		return
	}

	// Extract disk device for installboot
	var disk string
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, " on /mnt/backup ") {
			continue
		}
		dev := strings.Fields(line)[0]              // e.g. /dev/sd3a
		dev = strings.TrimPrefix(dev, "/dev/")      // sd3a
		disk = strings.TrimRight(dev, "abcdefghijklmnop") // sd3
		break
	}
	appendCloneOutput(fmt.Sprintf("Target disk: %s", disk))

	// ── 2. Space check ──────────────────────────────────────────────────

	appendCloneOutput("Checking space...")
	dfOut, err := exec.Command("df", "-P").Output()
	if err != nil {
		cloneError = fmt.Errorf("df: %w", err)
		return
	}

	var totalUsed, targetAvail int64
	for _, line := range strings.Split(string(dfOut), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		mp := fields[5]
		// Skip excluded mountpoints
		switch {
		case mp == "/mnt" || strings.HasPrefix(mp, "/mnt/"):
		case mp == "/dev":
		case mp == "/tmp":
		case mp == "/var/tmp":
		default:
			// Parse "used" column (index 2 in 512-blocks)
			var used int64
			fmt.Sscanf(fields[2], "%d", &used)
			totalUsed += used
		}
		// Target available
		if mp == "/mnt/backup" {
			fmt.Sscanf(fields[3], "%d", &targetAvail)
		}
	}

	sourceGB := totalUsed * 512 / 1024 / 1024 / 1024
	targetGB := targetAvail * 512 / 1024 / 1024 / 1024
	appendCloneOutput(fmt.Sprintf("Source used: ~%d GB, Target free: ~%d GB", sourceGB, targetGB))

	if targetAvail < totalUsed {
		cloneError = fmt.Errorf("Not enough space: %d GB free < %d GB needed", targetGB, sourceGB)
		return
	}

	// ── 3. Rsync ────────────────────────────────────────────────────────

	appendCloneOutput("Starting rsync...")
	rsync := exec.Command("doas", "rsync",
		"-aH", "--delete", "--numeric-ids", "--info=progress2",
		"--exclude=/dev",
		"--exclude=/tmp",
		"--exclude=/var/tmp",
		"--exclude=/mnt",
		"--exclude=/media",
		"--exclude=/lost+found",
		"/", "/mnt/backup/",
	)

	cloneOutputMu.Lock()
	cloneCmd = rsync
	cloneOutputMu.Unlock()

	stdout, _ := rsync.StdoutPipe()
	stderr, _ := rsync.StderrPipe()
	rsync.Start()

	// Read rsync progress from stderr
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				appendCloneOutput(strings.TrimSpace(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()
	// Discard stdout
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := stdout.Read(buf)
			if err != nil {
				break
			}
		}
	}()

	if err := rsync.Wait(); err != nil {
		cloneError = fmt.Errorf("rsync failed: %w", err)
		return
	}
	appendCloneOutput("Rsync complete.")

	// ── 3b. Create skeleton directories excluded by rsync ─────────────────

	for _, dir := range []string{
		"/mnt/backup/dev",
		"/mnt/backup/tmp",
		"/mnt/backup/var/tmp",
		"/mnt/backup/mnt",
		"/mnt/backup/media",
		"/mnt/backup/lost+found",
	} {
		os.MkdirAll(dir, 0755)
	}
	appendCloneOutput("Created skeleton directories.")

	// ── 4. Installboot ──────────────────────────────────────────────────

	appendCloneOutput(fmt.Sprintf("Installing boot blocks on %s...", disk))
	ib := exec.Command("doas", "installboot", "-r", "/mnt/backup", disk)
	ib.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if out, err := ib.CombinedOutput(); err != nil {
		cloneError = fmt.Errorf("installboot failed: %s: %w", string(out), err)
		return
	}

	appendCloneOutput(fmt.Sprintf("Boot blocks installed on %s.", disk))
	appendCloneOutput("Clone complete. Drive is bootable.")
}

// CheckCloneProgress returns a command that polls the clone operation.
func CheckCloneProgress() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		cloneOutputMu.Lock()
		running := cloneRunning
		completed := cloneCompleted
		err := cloneError
		cancelled := cloneCancel
		cloneOutputMu.Unlock()

		if cancelled {
			return ProgressUpdate{
				Percentage: -1,
				Message:    "Cancelled",
				Done:       true,
				Error:      fmt.Errorf("cancelled"),
			}
		}

		if completed {
			if err != nil {
				return ProgressUpdate{
					Percentage: -1,
					Message:    "",
					Done:       true,
					Error:      err,
				}
			}
			return ProgressUpdate{
				Percentage: 1.0,
				Message:    "Full backup complete — drive is bootable",
				Done:       true,
			}
		}

		if !running {
			return ProgressUpdate{
				Percentage: -1,
				Message:    "Waiting for clone to start...",
				Done:       false,
			}
		}

		// Indeterminate progress while running
		lines := getCloneOutput()
		msg := ""
		if len(lines) > 0 {
			msg = lines[len(lines)-1]
		}
		return ProgressUpdate{
			Percentage: -1,
			Message:    msg,
			Done:       false,
		}
	})
}
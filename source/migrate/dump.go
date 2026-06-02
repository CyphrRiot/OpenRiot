package migrate

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type dumpState struct {
	level      int
	filesystem string
	dumpFile   string
	isRestore  bool
}

var (
	currentDumpState dumpState

	dumpOutputMu   sync.Mutex
	dumpOutput     []string
	dumpRunning    bool
	dumpCompleted  bool
	dumpError      error
	dumpCancel     bool
	dumpCmd        *exec.Cmd
	dumpFilesTotal int
	dumpFilesDone  int
)

func resetDumpState() {
	dumpOutputMu.Lock()
	defer dumpOutputMu.Unlock()
	dumpCmd = nil
	dumpRunning = false
	dumpCompleted = false
	dumpError = nil
	dumpCancel = false
	dumpOutput = nil
	dumpFilesTotal = 0
	dumpFilesDone = 0
}

func appendDumpOutput(line string) {
	dumpOutputMu.Lock()
	defer dumpOutputMu.Unlock()
	dumpOutput = append(dumpOutput, line)
}

func getDumpOutput() []string {
	dumpOutputMu.Lock()
	defer dumpOutputMu.Unlock()
	out := make([]string, len(dumpOutput))
	copy(out, dumpOutput)
	return out
}

// CheckDumpProgress polls the background dump operation and returns progress
func CheckDumpProgress() tea.Cmd {
	return func() tea.Msg {
		dumpOutputMu.Lock()
		running := dumpRunning
		completed := dumpCompleted
		err := dumpError
		cancelled := dumpCancel
		done := dumpFilesDone
		total := dumpFilesTotal
		dumpOutputMu.Unlock()

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
				Message:    "Dump complete",
				Done:       true,
			}
		}

		if !running {
			return ProgressUpdate{
				Percentage: -1,
				Message:    "Starting...",
				Done:       false,
			}
		}

		progress := -1.0
		if total > 0 {
			progress = float64(done) / float64(total)
		}

		msg := fmt.Sprintf("Dumping %s", currentDumpState.filesystem)
		if currentDumpState.dumpFile != "" {
			msg = fmt.Sprintf("Dumping %s → %s", currentDumpState.filesystem, currentDumpState.dumpFile)
		}

		return ProgressUpdate{
			Percentage: progress,
			Message:    msg,
			Done:       false,
		}
	}
}

// runDumpAsync runs dump(8) in a background goroutine
func runDumpAsync(mountPoint, operation string) {
	resetDumpState()

	dir := mountPoint + "/backup"
	if err := exec.Command("doas", "mkdir", "-p", dir).Run(); err != nil {
		dumpOutputMu.Lock()
		dumpError = err
		dumpCompleted = true
		dumpOutputMu.Unlock()
		return
	}

	level := 0
	if operation == "dump_incr" {
		level = 1
	}

	dateStamp := time.Now().Format("2006-01-02_150405")
	filesystems := []string{"/", "/home", "/var"}

	dumpOutputMu.Lock()
	dumpRunning = true
	dumpFilesTotal = len(filesystems)
	dumpFilesDone = 0
	dumpOutputMu.Unlock()

	for _, fs := range filesystems {
		dumpOutputMu.Lock()
		if dumpCancel {
			dumpOutputMu.Unlock()
			return
		}
		dumpOutputMu.Unlock()

		fsName := strings.ReplaceAll(fs, "/", "_")
		if fsName == "_" {
			fsName = "root"
		}

		dumpFile := fmt.Sprintf("%s/level%d-%s-%s.dump", dir, level, fsName, dateStamp)

		dumpOutputMu.Lock()
		currentDumpState = dumpState{level: level, filesystem: fs, dumpFile: dumpFile}
		dumpOutputMu.Unlock()

		cmd := exec.Command("doas", "dump", fmt.Sprintf("-%du", level), "-f", dumpFile, fs)
		stderr, err := cmd.StderrPipe()
		if err != nil {
			dumpOutputMu.Lock()
			dumpError = fmt.Errorf("pipe error for %s: %v", fs, err)
			dumpCompleted = true
			dumpOutputMu.Unlock()
			return
		}

		if err := cmd.Start(); err != nil {
			dumpOutputMu.Lock()
			dumpError = fmt.Errorf("start error for %s: %v", fs, err)
			dumpCompleted = true
			dumpOutputMu.Unlock()
			return
		}

		dumpOutputMu.Lock()
		dumpCmd = cmd
		dumpOutputMu.Unlock()

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			appendDumpOutput(line)
		}

		if err := cmd.Wait(); err != nil {
			dumpOutputMu.Lock()
			dumpError = fmt.Errorf("dump failed for %s: %v", fs, err)
			dumpCompleted = true
			dumpOutputMu.Unlock()
			return
		}

		dumpOutputMu.Lock()
		dumpFilesDone++
		dumpOutputMu.Unlock()
	}

	dumpOutputMu.Lock()
	dumpCompleted = true
	dumpOutputMu.Unlock()
}

// startDump kicks off the async dump and returns an initial command batch
func startDump(mountPoint, operation string) tea.Cmd {
	go runDumpAsync(mountPoint, operation)
	return CheckDumpProgress()
}

// runDumpRestoreAsync runs restore(8) in a background goroutine
func runDumpRestoreAsync(mountPoint string) {
	resetDumpState()

	dir := mountPoint + "/backup"
	cmd := exec.Command("doas", "find", dir, "-maxdepth", "1", "-name", "*.dump", "-type", "f")
	out, err := cmd.Output()
	if err != nil {
		dumpOutputMu.Lock()
		dumpError = fmt.Errorf("failed to list dump files: %v", err)
		dumpCompleted = true
		dumpOutputMu.Unlock()
		return
	}

	files := strings.Fields(string(out))
	if len(files) == 0 {
		dumpOutputMu.Lock()
		dumpError = fmt.Errorf("no dump files found in %s", dir)
		dumpCompleted = true
		dumpOutputMu.Unlock()
		return
	}

	dumpOutputMu.Lock()
	dumpRunning = true
	dumpFilesTotal = len(files)
	dumpFilesDone = 0
	dumpOutputMu.Unlock()

	for _, f := range files {
		dumpOutputMu.Lock()
		if dumpCancel {
			dumpOutputMu.Unlock()
			return
		}
		dumpOutputMu.Unlock()

		currentDumpState = dumpState{isRestore: true, dumpFile: f}

		base := strings.TrimSuffix(f, ".dump")
		parts := strings.Split(base, "-")
		targetFS := "/"
		if len(parts) >= 2 && parts[1] != "root" {
			targetFS = "/" + parts[1]
		}

		appendDumpOutput(fmt.Sprintf("Restoring %s → %s", f, targetFS))

		restoreCmd := exec.Command("doas", "restore", "-xf", f)
		restoreCmd.Dir = targetFS
		output, err := restoreCmd.CombinedOutput()
		if err != nil {
			dumpOutputMu.Lock()
			dumpError = fmt.Errorf("restore failed for %s: %v\n%s", f, err, string(output))
			dumpCompleted = true
			dumpOutputMu.Unlock()
			return
		}

		dumpOutputMu.Lock()
		dumpFilesDone++
		dumpOutputMu.Unlock()
	}

	dumpOutputMu.Lock()
	dumpCompleted = true
	dumpOutputMu.Unlock()
}

// startDumpRestore kicks off the async restore
func startDumpRestore(mountPoint string) tea.Cmd {
	go runDumpRestoreAsync(mountPoint)
	return CheckDumpProgress()
}
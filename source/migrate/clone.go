package migrate

import (
	"bytes"
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
	cloneOutputMu   sync.Mutex
	cloneOutput     []string
	cloneRunning    bool
	cloneCompleted  bool
	cloneError      error
	cloneCancel     bool
	cloneProgress   float64
	cloneTargetBase int64
	cloneSourceKB   int64
	cloneCmd        *exec.Cmd
)

func resetCloneState() {
	cloneOutputMu.Lock()
	defer cloneOutputMu.Unlock()
	cloneCmd = nil
	cloneRunning = false
	cloneCompleted = false
	cloneError = nil
	cloneCancel = false
	cloneProgress = 0
	cloneTargetBase = 0
	cloneSourceKB = 0
	cloneOutput = nil
}

// dfUsedKB returns the used space of a mountpoint in 1K blocks.
func dfUsedKB(mount string) int64 {
	out, err := exec.Command("df", "-kP", mount).Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return 0
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 3 {
		return 0
	}
	var used int64
	fmt.Sscanf(fields[2], "%d", &used)
	return used
}

// watchCloneCancel kills the running clone process tree when the user
// cancels. dump forks several children, so pkill by name is used.
func watchCloneCancel(procs ...string) {
	for {
		cloneOutputMu.Lock()
		cancelled := cloneCancel
		cloneOutputMu.Unlock()
		if cancelled {
			for _, name := range procs {
				exec.Command("doas", "pkill", "-9", name).Run()
			}
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// dumpProgress parses "DUMP: nn.nn% done" into a fraction, or -1.
func dumpProgress(line string) float64 {
	i := strings.Index(line, "DUMP:")
	if i < 0 {
		return -1
	}
	rest := line[i+5:]
	j := strings.IndexByte(rest, '%')
	if j < 0 {
		return -1
	}
	var pct float64
	if _, err := fmt.Sscanf(strings.TrimSpace(rest[:j]), "%f", &pct); err != nil {
		return -1
	}
	return pct / 100
}

func appendCloneOutput(line string) {
	cloneOutputMu.Lock()
	defer cloneOutputMu.Unlock()
	cloneOutput = append(cloneOutput, line)
	if len(cloneOutput) > 500 {
		cloneOutput = cloneOutput[len(cloneOutput)-500:]
	}
}

func getCloneOutput() []string {
	cloneOutputMu.Lock()
	defer cloneOutputMu.Unlock()
	out := make([]string, len(cloneOutput))
	copy(out, cloneOutput)
	return out
}

func getSwapDevices() []string {
	out, err := exec.Command("swapctl", "-l").Output()
	if err != nil {
		return nil
	}
	var devs []string
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 1 && strings.HasPrefix(fields[0], "/dev/") {
			devs = append(devs, fields[0])
		}
	}
	return devs
}

func extractDisk() (string, error) {
	out, err := exec.Command("mount").Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, " on /mnt/backup ") {
			dev := strings.TrimPrefix(strings.Fields(line)[0], "/dev/")
			return strings.TrimRight(dev, "abcdefghijklmnop"), nil
		}
	}
	return "", fmt.Errorf("not mounted")
}

// mountAttempt runs fsck and mounts /dev/sd3a if /mnt/backup isn't up.
// Ignores errors — callers re-check the mount and surface a clear message.
func mountAttempt() {
	out, err := exec.Command("mount").Output()
	if err == nil && strings.Contains(string(out), " on /mnt/backup ") {
		return // already mounted
	}
	exec.Command("doas", "fsck", "-y", "/dev/sd3a").Run()
	exec.Command("doas", "mkdir", "-p", "/mnt/backup").Run()
	exec.Command("doas", "mount", "/dev/sd3a", "/mnt/backup").Run()
}

// ── Rsync clone ──────────────────────────────────────────────────────────

func StartClone() tea.Cmd {
	resetCloneState()
	cloneOutputMu.Lock()
	cloneRunning = true
	cloneOutputMu.Unlock()
	go watchCloneCancel("rsync")
	go runClone()
	return CheckCloneProgress()
}

func runClone() {
	defer func() {
		cloneOutputMu.Lock()
		cloneCompleted = true
		cloneOutputMu.Unlock()
	}()

	appendCloneOutput("Checking /mnt/backup mount...")
	out, err := exec.Command("mount").Output()
	if err != nil {
		cloneError = fmt.Errorf("mount: %w", err)
		return
	}
	if !strings.Contains(string(out), " on /mnt/backup ") {
		cloneError = fmt.Errorf("/mnt/backup is not mounted")
		return
	}
	disk, err := extractDisk()
	if err != nil {
		cloneError = err
		return
	}
	appendCloneOutput(fmt.Sprintf("Target disk: %s", disk))

	appendCloneOutput("Checking space...")
	dfOut, err := exec.Command("df", "-kP").Output()
	if err != nil {
		cloneError = fmt.Errorf("df: %w", err)
		return
	}
	var totalUsed, targetUsed, targetAvail int64
	for _, line := range strings.Split(string(dfOut), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		mp := fields[5]
		if fields[0] == "Filesystem" {
			continue
		}
		var used, avail int64
		fmt.Sscanf(fields[2], "%d", &used)
		fmt.Sscanf(fields[3], "%d", &avail)
		if mp == "/mnt/backup" {
			targetUsed = used
			targetAvail = avail
		}
		if mp == "/mnt" || strings.HasPrefix(mp, "/mnt/") || mp == "/dev" || mp == "/tmp" || mp == "/var/tmp" {
			continue
		}
		totalUsed += used
	}
	actualNeeded := (totalUsed - targetUsed) / 1024 / 1024
	totalUsedGB, targetUsedGB, targetAvailGB := totalUsed/1024/1024, targetUsed/1024/1024, targetAvail/1024/1024
	if targetUsed > 0 {
		appendCloneOutput(fmt.Sprintf("Source: ~%d GB, Target: ~%d GB used, %d GB free", totalUsedGB, targetUsedGB, targetAvailGB))
		if targetAvail < actualNeeded*1024*1024 {
			cloneError = fmt.Errorf("Not enough space: %d GB free < %d GB needed", targetAvailGB, actualNeeded)
			return
		}
	} else {
		appendCloneOutput(fmt.Sprintf("Source: ~%d GB, Target free: %d GB", totalUsedGB, targetAvailGB))
		if targetAvail < totalUsed {
			cloneError = fmt.Errorf("Not enough space: %d GB free < %d GB needed", targetAvailGB, totalUsedGB)
			return
		}
	}

	appendCloneOutput("Disabling swap...")
	swapDevices := getSwapDevices()
	for _, d := range swapDevices {
		exec.Command("doas", "swapctl", "-d", d).Run()
	}
	if len(swapDevices) > 0 {
		defer func() {
			appendCloneOutput("Re-enabling swap...")
			for _, d := range swapDevices {
				exec.Command("doas", "swapctl", "-a", d).Run()
			}
		}()
	}

	appendCloneOutput("Starting rsync...")
	rsync := exec.Command("doas", "nice", "-n", "20", "rsync",
		"-aH", "--delete", "--delete-delay", "--prune-empty-dirs", "--numeric-ids", "--info=progress2", "--bwlimit=30M",
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
	for _, dir := range []string{"/mnt/backup/dev", "/mnt/backup/tmp", "/mnt/backup/var/tmp", "/mnt/backup/mnt", "/mnt/backup/media", "/mnt/backup/lost+found"} {
		os.MkdirAll(dir, 0755)
	}
	appendCloneOutput("Finalizing Clone...")
	ib := exec.Command("doas", "installboot", "-r", "/mnt/backup", disk)
	ib.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if out, err := ib.CombinedOutput(); err != nil {
		cloneError = fmt.Errorf("installboot failed: %s: %w", string(out), err)
		return
	}
	appendCloneOutput(fmt.Sprintf("Clone complete. /dev/%s is bootable.", disk))
}

// ── Dump | restore sequential clone (safe for softraid crypto) ────────

func StartDumpClone() tea.Cmd {
	resetCloneState()
	cloneOutputMu.Lock()
	cloneRunning = true
	cloneOutputMu.Unlock()
	go watchCloneCancel("dump", "restore")
	go runDumpClone()
	return CheckCloneProgress()
}

func runDumpClone() {
	defer func() {
		cloneOutputMu.Lock()
		cloneCompleted = true
		cloneOutputMu.Unlock()
	}()

	appendCloneOutput("Checking /mnt/backup mount...")
	out, err := exec.Command("mount").Output()
	if err != nil {
		cloneError = fmt.Errorf("mount: %w", err)
		return
	}
	if !strings.Contains(string(out), " on /mnt/backup ") {
		cloneError = fmt.Errorf("/mnt/backup is not mounted")
		return
	}
	disk, err := extractDisk()
	if err != nil {
		cloneError = err
		return
	}
	appendCloneOutput(fmt.Sprintf("Target disk: %s", disk))

	for _, part := range []struct{ raw, label, mount string }{
		{"/dev/rsd1a", "root", "/"},
		{"/dev/rsd1d", "home", "/home"},
	} {
		// Progress is per-partition so the bar climbs 0-100% for each one
		// instead of crawling through a combined total.
		cloneOutputMu.Lock()
		cloneTargetBase = dfUsedKB("/mnt/backup")
		cloneSourceKB = dfUsedKB(part.mount)
		cloneOutputMu.Unlock()

		appendCloneOutput(fmt.Sprintf("Cloning %s partition (%s)...", part.label, part.raw))
		cmd := exec.Command("doas", "sh", "-c",
			fmt.Sprintf("dump 0af - %s | (cd /mnt/backup && restore -rf -)", part.raw))
		stderr, _ := cmd.StderrPipe()
		cmd.Start()

		// dump and restore both write to stderr; dump emits its own
		// "DUMP:" lines, restore emits per-file chatter. Keep only the
		// DUMP lines so the header doesn't churn.
		buf := make([]byte, 4096)
		var acc []byte
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				acc = append(acc, buf[:n]...)
				for {
					idx := bytes.IndexAny(acc, "\r\n")
					if idx < 0 {
						break
					}
					token := strings.TrimSpace(string(acc[:idx]))
					acc = acc[idx+1:]
					if strings.Contains(token, "DUMP:") {
						appendCloneOutput(token)
					}
				}
			}
			if err != nil {
				break
			}
		}
		if err := cmd.Wait(); err != nil {
			cloneError = fmt.Errorf("dump %s failed: %w", part.label, err)
			return
		}
		appendCloneOutput(fmt.Sprintf("%s partition cloned.", part.label))
	}

	appendCloneOutput("Finalizing clone...")
	ib := exec.Command("doas", "installboot", "-r", "/mnt/backup", disk)
	ib.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if out, err := ib.CombinedOutput(); err != nil {
		cloneError = fmt.Errorf("installboot failed: %s: %w", string(out), err)
		return
	}
	appendCloneOutput(fmt.Sprintf("Clone complete. /dev/%s is bootable.", disk))
}

// ── Progress ────────────────────────────────────────────────────────────

func CheckCloneProgress() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		cloneOutputMu.Lock()
		running := cloneRunning
		completed := cloneCompleted
		err := cloneError
		cancelled := cloneCancel
		cloneOutputMu.Unlock()

		if cancelled {
			return ProgressUpdate{Percentage: -1, Message: "Cancelled", Done: true, Error: fmt.Errorf("cancelled")}
		}
		if completed {
			if err != nil {
				return ProgressUpdate{Percentage: -1, Message: "", Done: true, Error: err}
			}
			return ProgressUpdate{Percentage: 1.0, Message: "Full clone complete — drive is bootable", Done: true}
		}
		if !running {
			return ProgressUpdate{Percentage: -1, Message: "Waiting for clone to start...", Done: false}
		}
		lines := getCloneOutput()
		msg := ""
		if len(lines) > 0 {
			msg = lines[len(lines)-1]
		}
		// Progress = how much the target has grown vs the source total.
		cloneOutputMu.Lock()
		base := cloneTargetBase
		src := cloneSourceKB
		cloneOutputMu.Unlock()
		prog := -1.0
		if src > 0 {
			used := dfUsedKB("/mnt/backup")
			prog = float64(used-base) / float64(src)
			if prog < 0 {
				prog = 0
			}
			if prog > 1 {
				prog = 1
			}
		}
		return ProgressUpdate{Percentage: prog, Message: msg, Done: false}
	})
}
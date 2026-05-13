package screenrec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"openriot/notify"
	"openriot/polybar"
)

const (
	iconIdle      = ""
	iconRecording = ""
	outputDir     = "Videos/Recordings"
	pidFile       = "recording.pid"
	pathFile      = "recording.path"
	sndioStateFile = "sndio.orig.flags"
)

func getCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cache", "openriot")
}

func pidFilePath() string {
	return filepath.Join(getCacheDir(), pidFile)
}

func pathFilePath() string {
	return filepath.Join(getCacheDir(), pathFile)
}

func sndioStateFilePath() string {
	return filepath.Join(getCacheDir(), sndioStateFile)
}

func saveSndioFlags() error {
	cmd := exec.Command("rcctl", "get", "sndiod", "flags")
	out, _ := cmd.Output()
	orig := strings.TrimSpace(string(out))
	return os.WriteFile(sndioStateFilePath(), []byte(orig), 0600)
}

func restoreSndioFlags() error {
	data, err := os.ReadFile(sndioStateFilePath())
	if err != nil {
		return err
	}
	orig := strings.TrimSpace(string(data))
	if orig == "" {
		cmd := exec.Command("doas", "rcctl", "set", "sndiod", "flags")
		cmd.Run()
	} else {
		cmd := exec.Command("doas", "rcctl", "set", "sndiod", "flags", orig)
		cmd.Run()
	}
	cmd := exec.Command("doas", "rcctl", "restart", "sndiod")
	cmd.Run()
	os.Remove(sndioStateFilePath())
	return nil
}

func enableAudioMonitor() error {
	if err := saveSndioFlags(); err != nil {
		return err
	}
	cmd := exec.Command("doas", "rcctl", "set", "sndiod", "flags", "-s", "default", "-m", "play,rec", "-s", "mon", "-m", "mon")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("doas", "rcctl", "restart", "sndiod")
	if err := cmd.Run(); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}

func isRecording() (bool, int) {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return false, 0
	}
	var pid int
	fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &pid)
	if pid <= 0 {
		return false, 0
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		cleanupState()
		return false, 0
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		cleanupState()
		return false, 0
	}
	return true, pid
}

func cleanupState() {
	os.Remove(pidFilePath())
	os.Remove(pathFilePath())
}

func Status() string {
	if recording, _ := isRecording(); recording {
		return fmt.Sprintf("%%{F#f7768e}%s%%{F-}%%{O2}", iconRecording)
	}
	return polybar.Icon(iconIdle)
}

func Toggle() error {
	recording, pid := isRecording()
	if recording {
		return stopRecording(pid)
	}
	return startRecording()
}

func startRecording() error {
	// Notify immediately — user should see feedback before any setup work
	notify.SendNotify("screenrec", "Screen Recorder", "Recording is starting...", "normal", 3000, 0)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get home dir: %w", err)
	}

	res, err := getResolution()
	if err != nil || res == "" {
		res = "1920x1080"
	}

	outDir := filepath.Join(home, outputDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-1504")
	outPath := filepath.Join(outDir, timestamp+".mp4")

	audioAvailable := false
	if err := enableAudioMonitor(); err == nil {
		audioAvailable = true
	}

	var cmdStr string
	if audioAvailable {
		cmdStr = fmt.Sprintf(
			"ffmpeg -y -f sndio -thread_queue_size 4096 -i snd/0.mon -f x11grab -thread_queue_size 4096 -framerate 30 -s %s -i :0.0 -c:v libx264 -crf 23 -preset ultrafast -pix_fmt yuv420p -c:a aac -b:a 192k '%s'",
			res, outPath,
		)
	} else {
		cmdStr = fmt.Sprintf(
			"ffmpeg -y -f x11grab -thread_queue_size 4096 -framerate 30 -s %s -i :0.0 -c:v libx264 -crf 23 -preset ultrafast -pix_fmt yuv420p -an '%s'",
			res, outPath,
		)
	}

	cmd := exec.Command("sh", "-c", cmdStr+" &")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Wait a moment for ffmpeg to spawn, then capture its PID
	time.Sleep(200 * time.Millisecond)
	pgrep := exec.Command("pgrep", "-n", "-f", "ffmpeg.*x11grab")
	out, _ := pgrep.Output()
	var pid int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &pid)
	if pid > 0 {
		os.WriteFile(pidFilePath(), []byte(fmt.Sprintf("%d", pid)), 0600)
		os.WriteFile(pathFilePath(), []byte(outPath), 0600)
	}

	if !audioAvailable {
		notify.SendNotify("screenrec", "Screen Recorder", "No audio monitor available\nScreen only", "normal", 3000, 0)
	}

	return nil
}

func stopRecording(pid int) error {
	proc, err := os.FindProcess(pid)
	if err == nil {
		proc.Signal(syscall.SIGINT)
	}

	// Wait briefly for ffmpeg to finalize the file
	time.Sleep(500 * time.Millisecond)

	outPath := ""
	if data, err := os.ReadFile(pathFilePath()); err == nil {
		outPath = strings.TrimSpace(string(data))
	}

	cleanupState()
	restoreSndioFlags()

	msg := "Recording is stopping..."
	if outPath != "" {
		home, _ := os.UserHomeDir()
		displayPath := strings.Replace(outPath, home, "~", 1)
		msg = fmt.Sprintf("Recording is stopping...\nSaved to:\n %s", displayPath)
	}
	notify.SendNotify("screenrec", "Screen Recorder", msg, "normal", 5000, 0)

	return nil
}

func getResolution() (string, error) {
	cmd := exec.Command("xrandr", "--current")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "*") {
			fields := strings.Fields(line)
			for _, f := range fields {
				if strings.Contains(f, "x") {
					return f, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no active display found")
}

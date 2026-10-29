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
	"openriot/paths"
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
	return paths.Join(".cache", "openriot")
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

// IsRecording reports whether a screen recording is currently active.
func IsRecording() bool {
	recording, _ := isRecording()
	return recording
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

	res, err := getResolution()
	if err != nil || res == "" {
		res = "1920x1080"
	}

	outDir := paths.Join(outputDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-1504")
	outPath := filepath.Join(outDir, timestamp+".mp4")

	audioAvailable := false
	if err := enableAudioMonitor(); err == nil {
		audioAvailable = true
	}

	audio := audioAvailable
	cmd, err := launchFFmpeg(audio, res, outPath)
	if err != nil {
		if audio {
			restoreSndioFlags()
		}
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	if !waitRunning(cmd, 700*time.Millisecond) {
		if audio {
			restoreSndioFlags()
			audio = false
		}
		cmd, err = launchFFmpeg(false, res, outPath)
		if err != nil {
			return fmt.Errorf("failed to start ffmpeg: %w", err)
		}
		if !waitRunning(cmd, 700*time.Millisecond) {
			return fmt.Errorf("ffmpeg exited before recording started")
		}
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFilePath(),
		[]byte(fmt.Sprintf("%d", pid)), 0600); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}
	os.WriteFile(pathFilePath(), []byte(outPath), 0600)

	if !audio {
		notify.SendNotify("screenrec", "Screen Recorder",
			"No audio monitor available\nScreen only",
			"normal", 3000, 0)
	}

	return nil
}

func waitRunning(cmd *exec.Cmd, d time.Duration) bool {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-time.After(d):
		return true
	case <-done:
		return false
	}
}

func launchFFmpeg(audio bool, res, outPath string) (*exec.Cmd, error) {
	args := []string{"-y"}
	if audio {
		args = append(args, "-f", "sndio", "-thread_queue_size",
			"4096", "-i", "snd/0.mon")
	}
	args = append(args, "-f", "x11grab", "-thread_queue_size", "4096",
		"-framerate", "30", "-s", res, "-i", ":0.0",
		"-c:v", "libx264", "-crf", "23", "-preset", "ultrafast",
		"-pix_fmt", "yuv420p")
	if audio {
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	} else {
		args = append(args, "-an")
	}
	args = append(args, outPath)

	cmd := exec.Command("ffmpeg", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	logPath := filepath.Join(getCacheDir(), "screenrec.log")
	if err := os.MkdirAll(getCacheDir(), 0755); err == nil {
		if logFile, err := os.OpenFile(logPath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			cmd.Stdout = logFile
			cmd.Stderr = logFile
			defer logFile.Close()
		}
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
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
		home := paths.HomeDir()
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

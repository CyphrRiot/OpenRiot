package notify

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"openriot/paths"
)

const (
	cooldownFile = "notify-cooldown"
	cooldownMs   = 500 // milliseconds between notifications
)

// ShouldSend checks if enough time has passed since the last notification
func ShouldSend() bool {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".cache/openriot", cooldownFile)

	data, err := os.ReadFile(path)
	if err != nil {
		// No cooldown file = allowed
		return true
	}

	lastNs, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		// Corrupted file = reset and allow
		return true
	}

	elapsed := time.Since(time.Unix(0, lastNs))
	return elapsed > time.Duration(cooldownMs)*time.Millisecond
}

// RecordSend writes the current timestamp to the cooldown file
func RecordSend() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".cache/openriot")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, cooldownFile)
	return os.WriteFile(path, []byte(strconv.FormatInt(time.Now().UnixNano(), 10)), 0600)
}

// SendNotify sends a notification with icon, title, body, urgency, and timeout (ms)
// If in cooldown period, notification is skipped.
// Icon uses GetIconPath for fallback handling.
// replaceID: 0 = no replacement, otherwise replace notification with this ID.
func SendNotify(iconName, title, body, urgency string, timeoutMs, replaceID int) error {
	if !ShouldSend() {
		return nil // cooldown active, skip
	}
	if err := RecordSend(); err != nil {
		// Continue anyway - don't block notification on file error
	}

	icon := paths.GetIconPath(iconName)

	// Verify icon exists before sending notification
	if _, err := os.Stat(icon); err != nil {
		// Fallback to info icon if requested icon doesn't exist
		icon = filepath.Join(filepath.Dir(icon), "info.png")
	}

	args := []string{
		"/usr/local/bin/notify-send",
		"-i", icon,
		"-t", strconv.Itoa(timeoutMs),
	}
	if urgency == "critical" {
		args = append(args, "-u", "critical")
	}
	if replaceID > 0 {
		args = append(args, "-r", strconv.Itoa(replaceID))
	}
	args = append(args, title, body)

	return exec.Command(args[0], args[1:]...).Run()
}
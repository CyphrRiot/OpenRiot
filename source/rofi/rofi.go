package rofi

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

var homeDir, _ = os.UserHomeDir()

func Run() error {
	appsFile := findAppsFile()
	if appsFile == "" {
		return fmt.Errorf("apps.txt not found")
	}

	entries, err := parseAppsFile(appsFile)
	if err != nil {
		return fmt.Errorf("failed to parse apps file: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no apps found in apps.txt")
	}

	configDir := filepath.Dir(appsFile)
	theme := filepath.Join(configDir, "simple-tokyonight.rasi")

	// Build rofi input: "Icon Name" per line
	var rofiInput bytes.Buffer
	for _, entry := range entries {
		rofiInput.WriteString(fmt.Sprintf("%s  %s\n", entry.Icon, entry.Name))
	}

	// Run rofi
	cmd := exec.Command("rofi", "-dmenu", "-i", "-p", "Apps", "-format", "i", "-theme", theme)
	cmd.Stdin = &rofiInput
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil // User cancelled
	}

	selected := strings.TrimSpace(out.String())
	if selected == "" {
		return nil
	}

	var idx int
	if _, err := fmt.Sscanf(selected, "%d", &idx); err != nil {
		return fmt.Errorf("invalid selection: %s", selected)
	}

	if idx < 0 || idx >= len(entries) {
		return fmt.Errorf("selection out of range: %d", idx)
	}

	entry := entries[idx]

	// Send launching/stopping notification
	go func() {
		iconPath := filepath.Join(homeDir, ".local/share/openriot/config/icons/applications.png")
		name := entry.Name
		msg := "Launching"
		if strings.Contains(name, "󰭽") {
			msg = "Stopping"
			name = "Transmission"
		} else if strings.Contains(name, "󰅤") {
			name = "Transmission"
		}
		exec.Command("/usr/local/bin/notify-send", "-i", iconPath, "-t", "2000", "Applications", fmt.Sprintf("%s %s...", msg, name)).Run()
	}()

	// Execute the command
	return executeCommand(entry.Cmd)
}

func findAppsFile() string {
	paths := []string{
		filepath.Join(homeDir, ".config/rofi/apps.txt"),
		filepath.Join(homeDir, ".config/openriot/rofi/apps.txt"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

type appEntry struct {
	Name string
	Cmd  string
	Icon string
}

func parseAppsFile(path string) ([]appEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []appEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		entry := appEntry{
			Name: strings.TrimSpace(parts[0]),
			Cmd:  strings.TrimSpace(parts[1]),
			Icon: strings.TrimSpace(parts[2]),
		}
		entries = append(entries, entry)
	}

	// Handle dynamic entries
	for i, entry := range entries {
		if entry.Name == "Transmission" {
			if IsTransmissionRunning() {
				entries[i].Name = "Transmission 󰅤"
				entries[i].Cmd = "pkill -u $USER transmission-daemon"
			} else {
				entries[i].Name = "Transmission 󰭽"
				entries[i].Cmd = "sh -c \"mkdir -p ~/.local/share/transmission ~/.config/transmission && transmission-daemon -f --logfile ~/.local/share/transmission/daemon.log &\""
			}
		}
	}

	return entries, scanner.Err()
}

func IsTransmissionRunning() bool {
	cmd := exec.Command("pgrep", "-u", os.Getenv("USER"), "transmission-daemon")
	err := cmd.Run()
	return err == nil
}

func executeCommand(cmd string) error {
	var execCmd *exec.Cmd

	switch {
	case strings.HasSuffix(cmd, ".desktop"):
		desktopFile := filepath.Join(homeDir, ".local/share/applications", cmd)
		desktopCmd, err := getDesktopExec(desktopFile)
		if err != nil {
			return err
		}
		execCmd = exec.Command("sh", "-c", desktopCmd)

	case strings.HasPrefix(cmd, "https://"), strings.HasPrefix(cmd, "http://"):
		execCmd = exec.Command("sh", "-c", cmd+" &")

	default:
		execCmd = exec.Command("sh", "-c", cmd+" &")
	}

	execCmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	execCmd.Stdout = nil
	execCmd.Stderr = nil
	return execCmd.Start()
}

func getDesktopExec(desktopFile string) (string, error) {
	data, err := os.ReadFile(desktopFile)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Exec=") {
			return strings.TrimPrefix(line, "Exec="), nil
		}
	}
	return "", fmt.Errorf("no Exec= line found in %s", desktopFile)
}

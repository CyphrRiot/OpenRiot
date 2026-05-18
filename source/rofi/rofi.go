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

	"openriot/notify"
	"openriot/wireguard"
)

func Run() error {
	appsFile := findAppsFile()
	if appsFile == "" {
		return fmt.Errorf("apps.txt not found")
	}
	return runAppsFile(appsFile, "Apps")
}

func RunGames() error {
	gamesFile := findGamesFile()
	if gamesFile == "" {
		return fmt.Errorf("games.txt not found")
	}
	return runAppsFile(gamesFile, "Games")
}

func findGamesFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	paths := []string{
		filepath.Join(home, ".config", "rofi", "games.txt"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func runAppsFile(appsFile, prompt string) error {
	entries, err := parseAppsFile(appsFile)
	if err != nil {
		return fmt.Errorf("failed to parse apps file: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no apps found in %s", appsFile)
	}

	configDir := filepath.Dir(appsFile)
	theme := filepath.Join(configDir, "simple-tokyonight.rasi")

	// Build rofi input: "Icon Name" per line
	var rofiInput bytes.Buffer
	for _, entry := range entries {
		fmt.Fprintf(&rofiInput, "%s  %s\n", entry.Icon, entry.Name)
	}

	// Run rofi
	cmd := exec.Command("rofi", "-dmenu", "-i", "-p", prompt, "-format", "i", "-theme", theme)
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

	// Handle sub-menu
	if submenuName, ok := strings.CutPrefix(entry.Cmd, "@submenu:"); ok {
		submenuFile := filepath.Join(configDir, submenuName+".txt")
		return runAppsFile(submenuFile, entry.Name)
	}

	// Check if already running
	if entry.Cmd == "transmission-gtk" && IsTransmissionRunning() {
		go notify.SendNotify("applications", "Transmission", "Already Running", "normal", 2000, 0)
		return nil
	}

	// Block Transmission if WireGuard is not active
	if entry.Cmd == "transmission-gtk" && !wireguard.IsRunning() {
		go notify.SendNotify("applications", "Transmission", "Wireguard is NOT running.\nCannot start Transmission without Wireguard.\n(This is a protective measure)", "critical", 5000, 0)
		return nil
	}

	// Send launching notification
	go func() {
		notify.SendNotify("applications", "Applications", fmt.Sprintf("Launching %s...", entry.Name), "normal", 2000, 0)
	}()

	// Execute the command
	return executeCommand(entry.Cmd)
}

func findAppsFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	paths := []string{
		filepath.Join(home, ".config", "rofi", "apps.txt"),
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

	return entries, scanner.Err()
}

func IsTransmissionRunning() bool {
	cmd := exec.Command("pgrep", "-u", os.Getenv("USER"), "transmission-gtk")
	err := cmd.Run()
	return err == nil
}

func executeCommand(cmd string) error {
	var execCmd *exec.Cmd

	switch {
	case strings.HasSuffix(cmd, ".desktop"):
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get home dir: %w", err)
		}
		desktopFile := filepath.Join(home, ".local", "share", "applications", cmd)
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
		if after, ok := strings.CutPrefix(line, "Exec="); ok {
			return after, nil
		}
	}
	return "", fmt.Errorf("no Exec= line found in %s", desktopFile)
}

package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// getReleaseNotesPath resolves the path to the current version's release notes.
// Returns an error if the home dir, VERSION file, or notes file are unavailable.
func getReleaseNotesPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(homeDir, ".local", "share", "openriot", "VERSION"))
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(data))
	if version == "" || version == "unknown" {
		return "", fmt.Errorf("version unknown")
	}
	docsDir := filepath.Join(homeDir, ".local", "share", "openriot", "docs")
	notesPath := filepath.Join(docsDir, fmt.Sprintf("v%s-Release-Notes.md", version))
	if _, err := os.Stat(notesPath); err == nil {
		return notesPath, nil
	}
	parts := strings.Split(version, ".")
	if len(parts) >= 3 {
		legacyPath := filepath.Join(docsDir, fmt.Sprintf("v%s.%s-Release-Notes.md", parts[0], parts[1]))
		if _, err := os.Stat(legacyPath); err == nil {
			return legacyPath, nil
		}
	}
	return "", fmt.Errorf("release notes not found for v%s", version)
}

// ShowReleaseNotes renders the current version's release notes using lowdown
// with a custom pager that reads one screen at a time.
func ShowReleaseNotes() {
	notesPath, err := getReleaseNotesPath()
	if err != nil {
		return
	}
	version := strings.TrimSuffix(filepath.Base(notesPath), "-Release-Notes.md")
	version = strings.TrimPrefix(version, "v")

	// Render markdown to terminal-formatted string
	cmd := exec.Command("lowdown", "-Tterm", notesPath)
	out, _ := cmd.Output()
	content := strings.Split(string(out), "\n")

	// Build full output with header and footer so page accounting is uniform
	header := []string{
		"",
		strings.Repeat("=", 72),
		fmt.Sprintf("  OpenRiot v%s Release Notes", version),
		strings.Repeat("=", 72),
		"",
	}
	footer := []string{"", strings.Repeat("=", 72), ""}
	lines := append(header, content...)
	lines = append(lines, footer...)

	promptLines := 3 // blank line + prompt + newline after key
	pageSize := terminalHeight() - promptLines
	if pageSize < 5 {
		pageSize = 5
	}

	for i := 0; i < len(lines); {
		end := i + pageSize
		if end > len(lines) {
			end = len(lines)
		}
		for _, line := range lines[i:end] {
			fmt.Println(line)
		}
		i = end
		if i >= len(lines) {
			break
		}
		fmt.Println()
		key := readKey("-- More -- (press any key to continue, q to quit)")
		if key == 'q' || key == 'Q' {
			fmt.Println("  [skipped]")
			break
		}
	}
}

// AskShowReleaseNotes prompts the user whether to display release notes.
// Returns true if the user presses Y, y, or Enter. Returns false for N, n,
// Q, q, or any other key. If release notes are missing, returns false
// silently.
func AskShowReleaseNotes() bool {
	notesPath, err := getReleaseNotesPath()
	if err != nil {
		return false
	}
	version := strings.TrimSuffix(filepath.Base(notesPath), "-Release-Notes.md")
	version = strings.TrimPrefix(version, "v")

	fmt.Println()
	prompt := fmt.Sprintf("* Read v%s Release Notes? [Y/n] ", version)
	key := readKey(prompt)
	return key == 'y' || key == 'Y' || key == '\r' || key == '\n'
}

func terminalHeight() int {
	// Try /dev/tty first — stdout may be redirected (e.g. via script or pipe)
	tty, err := os.Open("/dev/tty")
	if err == nil {
		defer tty.Close()
		ws, err := unix.IoctlGetWinsize(int(tty.Fd()), unix.TIOCGWINSZ)
		if err == nil && ws.Row >= 10 {
			return int(ws.Row)
		}
	}
	// Fall back to stdout
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err == nil && ws.Row >= 10 {
		return int(ws.Row)
	}
	return 24
}

func readKey(prompt string) byte {
	fmt.Print(prompt)

	// Open /dev/tty explicitly — stdin may be a pipe (e.g. curl | sh)
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Scanln()
		return 0
	}
	defer tty.Close()

	fd := int(tty.Fd())
	oldState, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		fmt.Scanln()
		return 0
	}

	newState := *oldState
	newState.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	newState.Iflag &^= unix.ICRNL | unix.INLCR | unix.IGNCR
	newState.Cc[unix.VMIN] = 1
	newState.Cc[unix.VTIME] = 0

	err = unix.IoctlSetTermios(fd, unix.TIOCSETAF, &newState)
	if err != nil {
		fmt.Scanln()
		return 0
	}
	defer unix.IoctlSetTermios(fd, unix.TIOCSETAF, oldState)

	b := make([]byte, 1)
	n, _ := tty.Read(b)
	fmt.Println()
	if n == 0 {
		return 0
	}
	return b[0]
}

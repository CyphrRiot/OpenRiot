package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"openriot/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	logFile      *os.File
	logPath      string
	emojiSupport bool = true
	program      *tea.Program
	programReady bool = false
)

// SetProgram sets the TUI program for log integration
func SetProgram(p *tea.Program) {
	program = p
}

// SetProgramReady marks the TUI program as running (safe to Send)
func SetProgramReady(ready bool) {
	programReady = ready
}

// Log writes a log entry to file and console
func Log(status, category, operation, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] %s | %s | %s | %s\n", timestamp, status, category, operation, message)

	if logFile != nil {
		logFile.WriteString(entry)
	}

	emoji := ""
	switch status {
	case "SUCCESS", "Complete":
		emoji = "✅ "
	case "ERROR", "Error":
		emoji = "❌ "
	case "WARNING", "Warning":
		emoji = "⚠️ "
	case "INFO", "Progress":
		emoji = "📦 "
	}

	fmt.Fprintf(os.Stdout, "%s%s\n", emoji, message)
}

// LogMessage writes a simple message log
func LogMessage(level, message string) {
	// If we have a TUI program and it's ready, send log to it
	if program != nil && programReady {
		program.Send(tui.LogMsg(message))
		return
	}
	// Otherwise, fall back to stdout
	Log(level, "General", "Log", message)
}

// InitLogger initializes the log file
func InitLogger() error {
	logDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "openriot", "logs")
	os.MkdirAll(logDir, 0755)

	logPath = filepath.Join(logDir, fmt.Sprintf("openriot-%s.log", time.Now().Format("2006-01-02")))

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = f

	LogMessage("INFO", fmt.Sprintf("Logger initialized: %s", logPath))
	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		LogMessage("INFO", "Closing logger")
		logFile.Close()
	}
}

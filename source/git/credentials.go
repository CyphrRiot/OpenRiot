package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"openriot/logger"
	"openriot/tui"
)

var (
	GitUsername   string
	GitEmail      string
	GitConfirmUse bool
	GitInputDone  chan bool
	Program       *tea.Program
)

// SetProgram sets the TUI program reference for sending messages
func SetProgram(p *tea.Program) {
	Program = p
}

// SetGitInputChannel sets the channel for git input completion
func SetGitInputChannel(ch chan bool) {
	GitInputDone = ch
}

// SetGitCredentials sets the git credentials from TUI callbacks
func SetGitCredentials(username, email string, confirmed bool) {
	if username != "" {
		GitUsername = username
	}
	if email != "" {
		GitEmail = email
	}
	GitConfirmUse = confirmed
}

// SetGitConfirm sets only the confirmation status
func SetGitConfirm(confirmed bool) {
	GitConfirmUse = confirmed
}

// SetGitUsername sets only the username
func SetGitUsername(username string) {
	GitUsername = username
}

// SetGitEmail sets only the email
func SetGitEmail(email string) {
	GitEmail = email
}

// HandleGitConfiguration applies Git configuration with beautiful styling
func HandleGitConfiguration() error {
	logger.LogMessage("INFO", "Setting up Git configuration...")

	logger.Log("Progress", "Git", "Git Setup", "Checking credentials")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	// Check for existing git credentials in git config first
	existingName, _ := runGitConfigGet("user.name")
	existingEmail, _ := runGitConfigGet("user.email")

	var userName, userEmail string

	// If we have existing git credentials, ask to use them
	if strings.TrimSpace(existingName) != "" || strings.TrimSpace(existingEmail) != "" {
		logger.Log("Complete", "Git", "Git Found", "Found existing git credentials")
		logger.Log("Info", "Git", "Name", existingName)
		logger.Log("Info", "Git", "Email", existingEmail)
		if Program != nil {
			Program.Send(tui.InputRequestMsg{Mode: "git-confirm", Prompt: ""})
		}

		// Wait for confirmation with timeout
		select {
		case <-GitInputDone:
			// Continue normally
		case <-time.After(5 * time.Minute):
			logger.Log("Error", "Git", "Timeout", "Git configuration timed out")
			return fmt.Errorf("git configuration timed out after 5 minutes")
		}

		if GitConfirmUse {
			userName = existingName
			userEmail = existingEmail
			logger.Log("Success", "Git", "Git Setup", "Using existing credentials")
		} else {
			// User said no, prompt for new credentials
			logger.Log("Info", "Git", "Git Setup", "Setting up new credentials")
			if Program != nil {
				Program.Send(tui.InputRequestMsg{Mode: "git-username", Prompt: "Git Username: "})
			}

			// Wait for new credentials with timeout
			select {
			case <-GitInputDone:
				// Continue normally
			case <-time.After(5 * time.Minute):
				logger.Log("Error", "Git", "Timeout", "Git credential input timed out")
				return fmt.Errorf("git credential input timed out after 5 minutes")
			}

			userName = GitUsername
			userEmail = GitEmail
		}
	} else {
		// No existing credentials, prompt for new ones
		logger.Log("Info", "Git", "Git Setup", "No credentials found, setting up")
		if Program != nil {
			Program.Send(tui.InputRequestMsg{Mode: "git-username", Prompt: "Git Username: "})
		}

		// Wait for credentials to be entered with timeout
		select {
		case <-GitInputDone:
			// Continue normally
		case <-time.After(5 * time.Minute):
			logger.Log("Error", "Git", "Timeout", "Git credential input timed out")
			return fmt.Errorf("git credential input timed out after 5 minutes")
		}

		userName = GitUsername
		userEmail = GitEmail
	}

	// Save credentials to user.env
	if err := saveGitCredentials(homeDir, userName, userEmail); err != nil {
		logger.LogMessage("WARNING", fmt.Sprintf("Failed to save git credentials: %v", err))
	}

	// Apply Git user configuration
	if strings.TrimSpace(userName) != "" {
		if err := runGitConfig("user.name", userName); err != nil {
			return fmt.Errorf("setting git user.name: %w", err)
		}
		logger.Log("Success", "Git", "Git Identity", "User name set to: "+userName)
		logger.LogMessage("SUCCESS", fmt.Sprintf("Git user.name set to: %s", userName))
	}

	if strings.TrimSpace(userEmail) != "" {
		if err := runGitConfig("user.email", userEmail); err != nil {
			return fmt.Errorf("setting git user.email: %w", err)
		}
		logger.Log("Success", "Git", "Git Identity", "Email set to: "+userEmail)
		logger.LogMessage("SUCCESS", fmt.Sprintf("Git user.email set to: %s", userEmail))
	}

	logger.Log("Success", "Git", "Git Setup", "Git configuration complete")
	return nil
}

// runGitConfig runs git config command
func runGitConfig(key, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config failed: %w - output: %s", err, string(output))
	}
	return nil
}

// runGitConfigGet gets a git config value
func runGitConfigGet(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", "--get", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// saveGitCredentials saves git credentials to a file
func saveGitCredentials(homeDir, userName, userEmail string) error {
	dotfilesDir := filepath.Join(homeDir, ".local", "share", "openriot")
	if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
		return fmt.Errorf("creating dotfiles directory: %w", err)
	}

	envPath := filepath.Join(dotfilesDir, "user.env")
	envContent := fmt.Sprintf("GIT_USERNAME=%s\nGIT_EMAIL=%s\n", userName, userEmail)

	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("writing user.env: %w", err)
	}

	return nil
}

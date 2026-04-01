package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"openriot/config"
	"openriot/git"
	"openriot/installer"
	"openriot/logger"
	"openriot/tui"
)

var version = "0.1"

// OpenRouter input completion channel
var openRouterInputDone chan bool

// Git input completion channel
var gitInputDone = make(chan bool, 1)

var testMode bool

func main() {
	// Check for test mode flag (for testing on Linux without OpenBSD)
	for _, arg := range os.Args[1:] {
		if arg == "--test" || arg == "-t" {
			testMode = true
			logger.LogMessage("INFO", "Running in TEST MODE (Linux)")
		}
		if arg == "--version" {
			fmt.Println("openriot", version)
			os.Exit(0)
		}
	}

	// Initialize logger
	if err := logger.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.LogMessage("INFO", "OpenRiot Installer starting...")

	// Set up version getter for TUI
	tui.SetVersionGetter(func() string { return version })

	// Load configuration from packages.yaml
	configPath := config.FindConfigFile()
	if configPath == "" {
		logger.LogMessage("ERROR", "Could not find packages.yaml")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.LogMessage("ERROR", fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	packages := cfg.GetPackages()

	if testMode {
		logger.LogMessage("INFO", "Package install skipped (test mode)")
	} else if err := installer.InstallPackages(packages); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Package install skipped (not OpenBSD?): %v", err))
	} else {
		logger.LogMessage("SUCCESS", "Packages installed successfully!")
	}

	// Deploy configuration files (from packages.yaml)
	logger.LogMessage("INFO", "Deploying configuration files...")

	// Determine repo directory based on mode
	var repoDir string
	if testMode {
		// In test mode, use development directory
		repoDir = os.Getenv("HOME") + "/Code/OpenRiot"
	} else {
		// In production, binary is in install/ relative to repo
		execPath, err := os.Executable()
		if err != nil {
			logger.LogMessage("WARN", "Could not determine executable path")
			repoDir = "/opt/openriot"
		} else {
			// Assume install/openriot -> repo is one level up
			repoDir = filepath.Dir(filepath.Dir(execPath))
		}
	}

	if err := installer.CopyConfigs(repoDir, cfg); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Config deployment skipped: %v", err))
	} else {
		logger.LogMessage("SUCCESS", "Configuration files deployed!")
	}

	// Execute commands from packages.yaml
	logger.LogMessage("INFO", "Running configuration commands...")
	if err := installer.ExecCommands(cfg, testMode); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Some commands failed: %v", err))
	}

	// Set Fish as default shell (skip in test mode)
	if !testMode {
		logger.LogMessage("INFO", "Setting Fish as default shell...")
		if err := setupFishShell(); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Fish shell setup skipped: %v", err))
		} else {
			logger.LogMessage("SUCCESS", "Fish shell set as default!")
		}
	} else {
		logger.LogMessage("INFO", "Fish shell setup skipped (test mode)")
	}

	// Initialize OpenRouter input channel
	openRouterInputDone = make(chan bool, 1)

	// Initialize git input channel
	gitInputDone = make(chan bool, 1)

	// Set up git credential callbacks
	tui.SetGitCallbacks(
		func(confirmed bool) {
			git.SetGitConfirm(confirmed)
			gitInputDone <- true
		},
		func(username string) {
			git.SetGitUsername(username)
		},
		func(email string) {
			git.SetGitEmail(email)
			gitInputDone <- true
		},
	)

	// Set up OpenRouter callbacks
	tui.SetOpenRouterCallbacks(
		func(confirmed bool) {
			openRouterInputDone <- confirmed
		},
		func(apiKey string) {
			writeOpenRouterToFish(apiKey)
			openRouterInputDone <- true
		},
	)

	model := tui.NewInstallModel()
	program := tea.NewProgram(model)

	// Wire git package to TUI program (OpenBSD only)
	git.SetProgram(program)
	if gitInputDone != nil {
		git.SetGitInputChannel(gitInputDone)
	}

	// Run git configuration and OpenRouter prompt in background (OpenBSD only)
	if !testMode {
		go func() {
			if err := git.HandleGitConfiguration(); err != nil {
				logger.LogMessage("WARN", fmt.Sprintf("Git configuration skipped: %v", err))
			}
			// After git, send OpenRouter prompt
			program.Send(tui.OpenRouterConfirmMsg(true))
		}()
	} else {
		logger.LogMessage("INFO", "Git configuration skipped (test mode)")
		// Skip OpenRouter setup in test mode - just let TUI display
	}

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// writeOpenRouterToFish writes OpenRouter API key to fish config
func writeOpenRouterToFish(apiKey string) {
	if apiKey == "" {
		return
	}

	usr, err := user.Current()
	if err != nil {
		logger.LogMessage("ERROR", "Failed to get current user: "+err.Error())
		return
	}

	fishConfigPath := filepath.Join(usr.HomeDir, ".config", "fish", "config.fish")

	// Read existing config
	content, err := os.ReadFile(fishConfigPath)
	if err != nil {
		logger.LogMessage("ERROR", "Failed to read fish config: "+err.Error())
		return
	}

	// Check if OpenRouter already exists
	if strings.Contains(string(content), "OPENROUTER_API_KEY") {
		logger.LogMessage("INFO", "OpenRouter already configured in fish config")
		return
	}

	// Append OpenRouter config
	openRouterConfig := `

# OpenRouter LLM Configuration
# Get your free key from https://openrouter.ai/settings
set -gx OPENROUTER_API_KEY "` + apiKey + `"
set -gx OPENROUTER_BASE_URL "https://openrouter.ai/api/v1"
`

	newContent := string(content) + openRouterConfig

	err = os.WriteFile(fishConfigPath, []byte(newContent), 0644)
	if err != nil {
		logger.LogMessage("ERROR", "Failed to write fish config: "+err.Error())
		return
	}

	logger.LogMessage("SUCCESS", "OpenRouter API key saved to fish config")
}

// setupFishShell sets Fish as the default shell
func setupFishShell() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	fishPath := "/usr/local/bin/fish"

	// Check if fish is installed
	if _, err := os.Stat(fishPath); os.IsNotExist(err) {
		return fmt.Errorf("fish not found at %s", fishPath)
	}

	// Add fish to /etc/shells if not already present
	shellsContent, err := os.ReadFile("/etc/shells")
	if err != nil {
		// May fail if not running as root, that's okay
		logger.LogMessage("INFO", "Could not check /etc/shells (may need root)")
	} else {
		if !strings.Contains(string(shellsContent), fishPath) {
			f, err := os.OpenFile("/etc/shells", os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				logger.LogMessage("WARN", "Could not add fish to /etc/shells (may need root)")
			} else {
				defer f.Close()
				f.WriteString(fishPath + "\n")
			}
		}
	}

	// Change default shell using chsh
	cmd := exec.Command("chsh", "-s", fishPath, usr.Username)
	if output, err := cmd.CombinedOutput(); err != nil {
		// chsh may fail without root, but fish will still work
		logger.LogMessage("WARN", "Could not set fish as default shell (may need root): "+string(output))
	}

	return nil
}

// deployConfigs copies configuration files from the repo to ~/.config/
func deployConfigs() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	homeDir := usr.HomeDir
	configDir := filepath.Join(homeDir, ".config")

	// List of configs to deploy
	configs := map[string]string{
		"fish/config.fish":      "fish/config.fish",
		"nvim/init.lua":         "nvim/init.lua",
		"nvim/lazyvim.json":     "nvim/lazyvim.json",
		"foot/cypherriot.ini":   "foot/cypherriot.ini",
		"sway/config":           "sway/config",
		"sway/keybindings.conf": "sway/keybindings.conf",
		"waybar/config":         "waybar/config",
	}

	// Get the directory where the openriot binary is located
	// This assumes the configs are bundled with the binary or in a standard location
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	repoDir := filepath.Dir(execPath)

	// Also check for ../config relative to executable (development mode)
	devConfigDir := filepath.Join(filepath.Dir(execPath), "..", "config")
	if _, err := os.Stat(devConfigDir); err == nil {
		repoDir = filepath.Dir(execPath)
	}

	// Backgrounds go to ~/.local/share/openriot/backgrounds
	bgDir := filepath.Join(homeDir, ".local", "share", "openriot", "backgrounds")
	if err := os.MkdirAll(bgDir, 0755); err != nil {
		logger.LogMessage("WARN", "Failed to create backgrounds directory: "+err.Error())
	} else {
		bgSrc := filepath.Join(repoDir, "..", "backgrounds")
		if _, err := os.Stat(bgSrc); err == nil {
			// Copy all jpg files from backgrounds directory
			entries, err := os.ReadDir(bgSrc)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jpg") {
						srcPath := filepath.Join(bgSrc, entry.Name())
						destPath := filepath.Join(bgDir, entry.Name())
						if data, err := os.ReadFile(srcPath); err == nil {
							if err := os.WriteFile(destPath, data, 0644); err == nil {
								logger.LogMessage("INFO", "Deployed background "+entry.Name())
							}
						}
					}
				}
			}
		}
	}

	// Create ~/.config if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	for src, dest := range configs {
		srcPath := filepath.Join(repoDir, src)
		destPath := filepath.Join(configDir, dest)

		// Skip if source doesn't exist
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			logger.LogMessage("INFO", "Skipping "+src+" (not found)")
			continue
		}

		// Create destination directory
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			logger.LogMessage("WARN", "Failed to create directory "+destDir+": "+err.Error())
			continue
		}

		// Copy file
		data, err := os.ReadFile(srcPath)
		if err != nil {
			logger.LogMessage("WARN", "Failed to read "+srcPath+": "+err.Error())
			continue
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			logger.LogMessage("WARN", "Failed to write "+destPath+": "+err.Error())
			continue
		}

		logger.LogMessage("INFO", "Deployed "+dest)
	}

	return nil
}

// deployConfigsTest uses local config directory for testing on Linux
func deployConfigsTest() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	homeDir := usr.HomeDir
	configDir := filepath.Join(homeDir, ".config")

	// List of configs to deploy
	configs := map[string]string{
		"fish/config.fish":      "fish/config.fish",
		"nvim/init.lua":         "nvim/init.lua",
		"nvim/lazyvim.json":     "nvim/lazyvim.json",
		"foot/cypherriot.ini":   "foot/cypherriot.ini",
		"sway/config":           "sway/config",
		"sway/keybindings.conf": "sway/keybindings.conf",
		"waybar/config":         "waybar/config",
	}

	// Backgrounds go to ~/.local/share/openriot/backgrounds
	bgDir := filepath.Join(homeDir, ".local", "share", "openriot", "backgrounds")
	if err := os.MkdirAll(bgDir, 0755); err != nil {
		logger.LogMessage("WARN", "Failed to create backgrounds directory: "+err.Error())
	} else {
		bgSrc := "/home/grendel/Code/OpenRiot/backgrounds"
		if _, err := os.Stat(bgSrc); err == nil {
			// Copy all jpg files from images directory
			entries, err := os.ReadDir(bgSrc)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jpg") {
						srcPath := filepath.Join(bgSrc, entry.Name())
						destPath := filepath.Join(bgDir, entry.Name())
						if data, err := os.ReadFile(srcPath); err == nil {
							if err := os.WriteFile(destPath, data, 0644); err == nil {
								logger.LogMessage("INFO", "Deployed background "+entry.Name())
							}
						}
					}
				}
			}
		}
	}

	// Create ~/.config if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Use local config directory from repo
	repoDir := "/home/grendel/Code/OpenRiot/config"

	for src, dest := range configs {
		srcPath := filepath.Join(repoDir, src)
		destPath := filepath.Join(configDir, dest)

		// Skip if source doesn't exist
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			logger.LogMessage("INFO", "Skipping "+src+" (not found)")
			continue
		}

		// Create destination directory
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			logger.LogMessage("WARN", "Failed to create directory "+destDir+": "+err.Error())
			continue
		}

		// Copy file
		data, err := os.ReadFile(srcPath)
		if err != nil {
			logger.LogMessage("WARN", "Failed to read "+srcPath+": "+err.Error())
			continue
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			logger.LogMessage("WARN", "Failed to write "+destPath+": "+err.Error())
			continue
		}

		logger.LogMessage("INFO", "Deployed "+dest)
	}

	return nil
}

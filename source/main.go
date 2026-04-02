package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"openriot/audio"
	"openriot/backgrounds"
	"openriot/config"
	"openriot/detect"
	"openriot/display"
	"openriot/git"
	"openriot/installer"
	"openriot/logger"
	"openriot/mullvad"
	"openriot/tui"
	"openriot/windows"
)

// Injected at build time via Makefile ldflags:
//
//	-X main.version=$(OPENRIOT_VERSION)
//	-X main.openbsdVersion=$(OPENBSD_VERSION)
//
// Do NOT hardcode these here — change Makefile instead.
var version = "dev"
var openbsdVersion = "7.9"

// OpenRouter input completion channel
var openRouterInputDone chan bool

// Git input completion channel
var gitInputDone = make(chan bool, 1)

var testMode bool

func main() {
	// CLI-only commands (run and exit immediately)
	if len(os.Args) >= 2 && os.Args[1] == "--volume" {
		os.Exit(audio.Run(os.Args[2:]))
	}
	if len(os.Args) >= 2 && os.Args[1] == "--brightness" {
		os.Exit(display.Run(os.Args[2:]))
	}
	if len(os.Args) >= 2 && os.Args[1] == "--lock" {
		cmd := exec.Command("swaylock", "-f")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Start()
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "--suspend" {
		exec.Command("zzz").Run()
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "--power-menu" {
		menu := "Lock\nSuspend\nReboot\nShutdown\nLogout"
		cmd := exec.Command("fuzzel", "--dmenu", "--prompt=Power: ", "--width=20", "--lines=5")
		cmd.Stdin = strings.NewReader(menu)
		out, err := cmd.Output()
		if err != nil {
			return
		}
		choice := strings.TrimSpace(string(out))
		switch choice {
		case "Lock":
			exec.Command("swaylock", "-f").Run()
		case "Suspend":
			exec.Command("zzz").Run()
		case "Reboot":
			exec.Command("shutdown", "-r", "now").Run()
		case "Shutdown":
			exec.Command("shutdown", "-p", "now").Run()
		case "Logout":
			exec.Command("swaymsg", "exit").Run()
		}
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "--swaybg-next" {
		os.Exit(backgrounds.Next())
	}
	if len(os.Args) >= 2 && os.Args[1] == "--suspend-if-undocked" {
		detect.SuspendIfUndocked()
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "--fix-offscreen-windows" {
		os.Exit(windows.FixOffscreenWindows())
	}
	if len(os.Args) >= 2 && os.Args[1] == "--mullvad-setup" {
		os.Exit(mullvad.Setup())
	}
	if len(os.Args) >= 2 && os.Args[1] == "--switch-window" {
		exec.Command("swaymsg", "-t", "get_tree").Run()
		return
	}
	// Check for test mode flag (for testing on Linux without OpenBSD)
	for _, arg := range os.Args[1:] {
		if arg == "--test" || arg == "-t" {
			testMode = true
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

	// Set up version getter for TUI
	tui.SetVersionGetter(func() string { return version })

	// ---- TUI MUST be created first, before any install steps ----
	model := tui.NewInstallModel()
	program := tea.NewProgram(model)

	// Set up unified logger with TUI program (must be before SetProgramReady)
	logger.SetProgram(program)

	// Wire progress and step updates to TUI
	logger.SetProgressCallback(func(p float64) {
		program.Send(tui.ProgressMsg(p))
	})
	logger.SetStepCallback(func(step string) {
		program.Send(tui.StepMsg(step))
	})

	// Wire git package to TUI program (OpenBSD only)
	git.SetProgram(program)
	if gitInputDone != nil {
		git.SetGitInputChannel(gitInputDone)
	}

	// Initialize OpenRouter input channel
	openRouterInputDone = make(chan bool, 1)

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

	// Mark program as ready — NOW logger.LogMessage will route to TUI
	logger.LogMessage("INFO", "OpenRiot Installer starting...")
	logger.SetProgramReady(true)

	// ---- Run all install steps in goroutines so TUI stays responsive ----
	packages := cfg.GetPackages()

	// Determine repo directory based on mode (needed by multiple goroutines)
	var repoDir string
	if testMode {
		repoDir = os.Getenv("HOME") + "/Code/OpenRiot"
	} else {
		execPath, err := os.Executable()
		if err != nil {
			logger.LogMessage("WARN", "Could not determine executable path")
			repoDir = "/opt/openriot"
		} else {
			repoDir = filepath.Dir(filepath.Dir(execPath))
		}
	}

	// Run package installation in background
	go func() {
		program.Send(tui.StepMsg("Installing packages..."))
		program.Send(tui.ProgressMsg(0.0))
		if testMode {
			logger.LogMessage("INFO", "Package install skipped (test mode)")
			program.Send(tui.ProgressMsg(1.0))
		} else if err := installer.InstallPackages(packages); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Package install skipped (not OpenBSD?): %v", err))
			program.Send(tui.ProgressMsg(1.0))
		} else {
			logger.LogMessage("SUCCESS", "Packages installed successfully!")
			program.Send(tui.ProgressMsg(1.0))
		}
	}()

	// Run config deployment in background
	go func() {
		program.Send(tui.StepMsg("Deploying configuration..."))
		program.Send(tui.ProgressMsg(0.3))
		if err := installer.CopyConfigs(repoDir, cfg); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Config deployment skipped: %v", err))
		} else {
			logger.LogMessage("SUCCESS", "Configuration files deployed!")
		}
		program.Send(tui.ProgressMsg(0.6))
	}()

	// Run command execution in background
	go func() {
		program.Send(tui.StepMsg("Running commands..."))
		program.Send(tui.ProgressMsg(0.7))
		if err := installer.ExecCommands(cfg, testMode); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Some commands failed: %v", err))
		}
		program.Send(tui.ProgressMsg(0.9))
	}()

	// Run source builds in background
	go func() {
		program.Send(tui.StepMsg("Building from source..."))
		program.Send(tui.ProgressMsg(0.8))
		if err := installer.SourceBuilds(cfg, testMode); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Source builds: %v", err))
		}
		program.Send(tui.ProgressMsg(0.92))
	}()

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

	// ---- Run TUI main loop ----
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

	content, err := os.ReadFile(fishConfigPath)
	if err != nil {
		logger.LogMessage("ERROR", "Failed to read fish config: "+err.Error())
		return
	}

	if strings.Contains(string(content), "OPENROUTER_API_KEY") {
		logger.LogMessage("INFO", "OpenRouter already configured in fish config")
		return
	}

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

	if _, err := os.Stat(fishPath); os.IsNotExist(err) {
		return fmt.Errorf("fish not found at %s", fishPath)
	}

	shellsContent, err := os.ReadFile("/etc/shells")
	if err != nil {
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

	cmd := exec.Command("chsh", "-s", fishPath, usr.Username)
	if output, err := cmd.CombinedOutput(); err != nil {
		logger.LogMessage("WARN", "Could not set fish as default shell (may need root): "+string(output))
	}

	return nil
}

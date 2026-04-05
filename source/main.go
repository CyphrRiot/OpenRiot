package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"openriot/audio"
	"openriot/backgrounds"
	"openriot/config"
	"openriot/crypto"
	"openriot/detect"
	"openriot/display"
	"openriot/git"
	"openriot/installer"
	"openriot/logger"
	"openriot/notify"
	"openriot/tui"
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
	// --install is handled by the default TUI install flow below
	if len(os.Args) >= 2 && os.Args[1] == "--install" {
		// Explicit --install flag: validated, falls through to TUI install
	}
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
	// --notify "title" "body" [--urgency normal|critical|low] [--expires-in seconds]
	if len(os.Args) >= 2 && os.Args[1] == "--notify" {
		title, body, urgency := "", "", "normal"
		expiresIn := 0
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--urgency" && i+1 < len(os.Args) {
				urgency = os.Args[i+1]
			} else if os.Args[i] == "--expires-in" && i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &expiresIn)
			} else if title == "" {
				title = os.Args[i]
			} else if body == "" {
				body = os.Args[i]
			}
		}
		if title == "" {
			fmt.Fprintln(os.Stderr, "Usage: openriot --notify \"title\" \"body\" [--urgency normal] [--expires-in seconds]")
			os.Exit(1)
		}
		var expiresAt int64
		if expiresIn > 0 {
			expiresAt = time.Now().Unix() + int64(expiresIn)
		}
		if err := notify.Add(title, body, urgency, expiresAt); err != nil {
			fmt.Fprintf(os.Stderr, "notify error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// --notify-dismiss [id]
	if len(os.Args) >= 2 && os.Args[1] == "--notify-dismiss" {
		id := 0
		if len(os.Args) >= 3 {
			fmt.Sscanf(os.Args[2], "%d", &id)
		}
		if err := notify.Dismiss(id); err != nil {
			fmt.Fprintf(os.Stderr, "notify dismiss error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// --notify-clear
	if len(os.Args) >= 2 && os.Args[1] == "--notify-clear" {
		if err := notify.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "notify clear error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// --notify-waybar
	if len(os.Args) >= 2 && os.Args[1] == "--notify-waybar" {
		if err := notify.Waybar(); err != nil {
			fmt.Fprintf(os.Stderr, "notify waybar error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Check for version flag first (before any other processing)
	for _, arg := range os.Args[1:] {
		if arg == "--version" {
			fmt.Println("openriot", version)
			os.Exit(0)
		}
	}

	// Crypto price commands
	if len(os.Args) >= 2 && os.Args[1] == "--crypto" {
		mode := "BTC"
		if len(os.Args) >= 3 {
			mode = os.Args[2]
		}
		if err := crypto.RunCrypto(mode); err != nil {
			fmt.Fprintf(os.Stderr, "crypto error: %v\n", err)
		}
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "--crypto-refresh" {
		// Clear cache and fetch fresh prices
		os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "openriot-crypto.json"))
		os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "openriot-crypto-prev.json"))
		if err := crypto.RunCrypto("ROWML"); err != nil {
			fmt.Fprintf(os.Stderr, "crypto error: %v\n", err)
		}
		return
	}

	// Check for test mode flag (for testing on Linux without OpenBSD)
	for _, arg := range os.Args[1:] {
		if arg == "--test" || arg == "-t" {
			testMode = true
		}
	}

	// Initialize logger (set testMode first so delay applies during startup)
	logger.SetTestMode(testMode)
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

	// ---- Start TUI loop FIRST in a goroutine so it can receive messages ----
	// This MUST run before any background tasks that send to the program
	tuiDone := make(chan error, 1)
	go func() {
		_, err := program.Run()
		tuiDone <- err
	}()

	// Small delay to ensure TUI loop is running before we send messages
	select {
	case err := <-tuiDone:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		}
		return
	case <-time.After(50 * time.Millisecond):
		// TUI loop is running, proceed with sequential install flow
	}

	// Track if user requested early quit
	userQuit := false

	// Helper to check if TUI quit during install
	checkQuit := func() bool {
		select {
		case <-tuiDone:
			userQuit = true
			return true
		default:
			return false
		}
	}

	// ---- Sequential install flow (no goroutines) ----
	// Run all install steps sequentially so progress and logs display in order
	packages := cfg.GetPackages()

	// Determine repo directory based on mode
	var repoDir string
	if testMode {
		repoDir = os.Getenv("HOME") + "/Code/OpenRiot"
	} else {
		homeDir, _ := os.UserHomeDir()
		repoDir = filepath.Join(homeDir, ".local", "share", "openriot")
		// Fallback for running directly from a repo checkout on OpenBSD
		if _, err := os.Stat(filepath.Join(repoDir, "install", "packages.yaml")); os.IsNotExist(err) {
			if execPath, err := os.Executable(); err == nil {
				repoDir = filepath.Dir(filepath.Dir(execPath))
			}
		}
	}

	// Step 1: Package installation
	if checkQuit() {
		<-tuiDone
		return
	}
	program.Send(tui.StepMsg("Installing packages..."))
	program.Send(tui.ProgressMsg(0.1))
	if testMode {
		logger.LogMessage("INFO", "Package install skipped (test mode)")
	} else if err := installer.InstallPackages(packages); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Package install skipped (not OpenBSD?): %v", err))
	} else {
		logger.LogMessage("SUCCESS", "Packages installed successfully!")
	}

	// Step 2: Config deployment
	if checkQuit() {
		<-tuiDone
		return
	}
	program.Send(tui.StepMsg("Deploying configuration..."))
	program.Send(tui.ProgressMsg(0.3))
	if err := installer.CopyConfigs(repoDir, cfg, testMode); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Config deployment skipped: %v", err))
	} else {
		logger.LogMessage("SUCCESS", "Configuration files deployed!")
	}

	// Step 3: Command execution
	if checkQuit() {
		<-tuiDone
		return
	}
	program.Send(tui.StepMsg("Running commands..."))
	program.Send(tui.ProgressMsg(0.6))
	if err := installer.ExecCommands(cfg, testMode); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Some commands failed: %v", err))
	}

	// Step 4: Source builds
	if checkQuit() {
		<-tuiDone
		return
	}
	program.Send(tui.StepMsg("Building from source..."))
	program.Send(tui.ProgressMsg(0.8))
	if err := installer.SourceBuilds(cfg, testMode); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Source builds: %v", err))
	}

	// Step 5: Git and OpenRouter configuration SKIPPED — these require interactive
	// TUI prompts which block non-interactive first-boot installs. Users can
	// configure git and OpenRouter manually after first boot if desired.
	logger.LogMessage("INFO", "Git and OpenRouter setup skipped (non-interactive mode)")

	// Signal completion
	program.Send(tui.ProgressMsg(1.0))
	program.Send(tui.DoneMsg{})

	// Wait for TUI to finish
	if userQuit {
		// User pressed q - TUI already exited, no need to wait
		fmt.Fprintln(os.Stderr, "⏳ Waiting to exit...")
		return
	}
	if err := <-tuiDone; err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
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

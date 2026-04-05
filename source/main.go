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

	"openriot/audio"
	"openriot/backgrounds"
	"openriot/config"
	"openriot/crypto"
	"openriot/detect"
	"openriot/display"
	"openriot/installer"
	"openriot/notify"
)

// Injected at build time via Makefile ldflags:
//
//	-X main.version=$(OPENRIOT_VERSION)
//	-X main.openbsdVersion=$(OPENBSD_VERSION)
//
// Do NOT hardcode these here — change Makefile instead.
var version = "dev"
var openbsdVersion = "7.9"

var testMode bool

func main() {
	// Check for version flag first
	for _, arg := range os.Args[1:] {
		if arg == "--version" {
			fmt.Println("openriot", version)
			os.Exit(0)
		}
	}

	// Check for test mode flag (for testing on Linux without OpenBSD)
	for _, arg := range os.Args[1:] {
		if arg == "--test" || arg == "-t" {
			testMode = true
		}
	}

	// Handle --install (simple CLI, no TUI)
	if len(os.Args) >= 2 && os.Args[1] == "--install" {
		runInstall()
		return
	}

	// All other CLI commands
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

	// No command or unknown command
	fmt.Fprintf(os.Stderr, "openriot %s\n", version)
	fmt.Fprintf(os.Stderr, "Usage: openriot <command>\n")
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  --install          Install OpenRiot (configs, not packages)\n")
	fmt.Fprintf(os.Stderr, "  --lock            Lock the screen\n")
	fmt.Fprintf(os.Stderr, "  --suspend         Suspend the system\n")
	fmt.Fprintf(os.Stderr, "  --power-menu       Show power menu\n")
	fmt.Fprintf(os.Stderr, "  --volume <args>    Adjust volume\n")
	fmt.Fprintf(os.Stderr, "  --brightness <args> Adjust brightness\n")
	fmt.Fprintf(os.Stderr, "  --notify \"title\" \"body\" Send notification\n")
	fmt.Fprintf(os.Stderr, "  --crypto [BTC|ETH] Show crypto prices\n")
	fmt.Fprintf(os.Stderr, "  --version         Show version\n")
	os.Exit(1)
}

// runInstall handles the --install command (runs as USER, no TTY/PTY needed)
func runInstall() {
	fmt.Println("[INFO]  OpenRiot installer starting...")

	// Load configuration from packages.yaml
	configPath := config.FindConfigFile()
	if configPath == "" {
		fmt.Fprintf(os.Stderr, "[ERR!]  Could not find packages.yaml\n")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR!]  Failed to load config: %v\n", err)
		os.Exit(1)
	}

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

	// Step 1: Config deployment
	fmt.Println("[INFO]  Deploying configuration files...")
	if err := installer.CopyConfigs(repoDir, cfg, testMode); err != nil {
		fmt.Printf("[WARN]  Config deployment skipped: %v\n", err)
	} else {
		fmt.Println("[INFO]  Configuration files deployed!")
	}

	// Step 2: Command execution
	fmt.Println("[INFO]  Running commands...")
	if err := installer.ExecCommands(cfg, testMode); err != nil {
		fmt.Printf("[WARN]  Some commands failed: %v\n", err)
	}

	// Step 3: Source builds
	fmt.Println("[INFO]  Building from source...")
	if err := installer.SourceBuilds(cfg, testMode); err != nil {
		fmt.Printf("[WARN]  Source builds: %v\n", err)
	}

	// Step 4: Copy binary to install directory
	if !testMode {
		installBinary(repoDir)
	}

	fmt.Println("[INFO]  OpenRiot installation complete!")
}

// installBinary copies the openriot binary to ~/.local/share/openriot/install/
func installBinary(repoDir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("[WARN]  Could not get home directory: %v\n", err)
		return
	}

	installDir := filepath.Join(homeDir, ".local", "share", "openriot", "install")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		fmt.Printf("[WARN]  Could not create install directory: %v\n", err)
		return
	}

	// Copy the running binary to the install directory
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("[WARN]  Could not find running executable: %v\n", err)
		return
	}

	destPath := filepath.Join(installDir, "openriot")
	if execPath == destPath {
		fmt.Println("[INFO]  Binary already in place")
		return
	}

	sourceData, err := os.ReadFile(execPath)
	if err != nil {
		fmt.Printf("[WARN]  Could not read binary: %v\n", err)
		return
	}

	if err := os.WriteFile(destPath, sourceData, 0755); err != nil {
		fmt.Printf("[WARN]  Could not write binary: %v\n", err)
		return
	}

	fmt.Printf("[INFO]  Binary installed to %s\n", destPath)
}

// writeOpenRouterToFish writes OpenRouter API key to fish config
func writeOpenRouterToFish(apiKey string) {
	if apiKey == "" {
		return
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Printf("[ERR!]  Failed to get current user: %v\n", err)
		return
	}

	fishConfigPath := filepath.Join(usr.HomeDir, ".config", "fish", "config.fish")

	content, err := os.ReadFile(fishConfigPath)
	if err != nil {
		fmt.Printf("[ERR!]  Failed to read fish config: %v\n", err)
		return
	}

	if strings.Contains(string(content), "OPENROUTER_API_KEY") {
		fmt.Println("[INFO]  OpenRouter already configured in fish config")
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
		fmt.Printf("[ERR!]  Failed to write fish config: %v\n", err)
		return
	}

	fmt.Println("[INFO]  OpenRouter API key saved to fish config")
}

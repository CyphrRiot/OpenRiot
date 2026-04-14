package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"openriot/audio"
	"openriot/backgrounds"
	"openriot/config"
	"openriot/crypto"
	"openriot/detect"
	"openriot/display"
	"openriot/installer"
	"openriot/lock"
	"openriot/notify"
	"openriot/network"
	"openriot/nightlight"
	"openriot/polybar"
	"openriot/battery"
	"openriot/rofi"
	"openriot/weather"
	"openriot/workspace"
	"openriot/wireguard"
	"openriot/windowicon"
	"openriot/windowtitle"
	"openriot/update"
)

// Injected at build time via Makefile ldflags:
//
//	-X main.version=$(OPENRIOT_VERSION)
//	-X main.openbsdVersion=$(OPENBSD_VERSION)
//
// Do NOT hardcode these here - change Makefile instead.
var version = "dev"
var openbsdVersion = "7.9"

var testMode bool

// logDebugCall logs each binary invocation to /tmp/openriot_calls.log
func logDebugCall() {
	if os.Getenv("OPENRIOT_DEBUG") != "1" {
		return
	}
	f, err := os.OpenFile("/tmp/openriot_calls.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: cannot open log: %v\n", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format("15:04:05.000"), strings.Join(os.Args[1:], " "))
}

// getIconPath returns absolute path to an icon file
func getIconPath(filename string) string {
	return filepath.Join(os.Getenv("HOME"), ".local/share/openriot/config/icons", filename)
}

func main() {
	logDebugCall()

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

	// Handle --source-builds (runs only the source builds phase, used by setup.sh)
	if len(os.Args) >= 2 && os.Args[1] == "--source-builds" {
		runSourceBuilds()
		return
	}

	// --install-packages (installs packages from packages.yaml, used by setup.sh)
	if len(os.Args) >= 2 && os.Args[1] == "--install-packages" {
		runInstallPackages()
		return
	}

	// --packages flag - outputs package list from packages.yaml (used by setup.sh)
	if len(os.Args) >= 2 && os.Args[1] == "--packages" {
		configPath := config.FindConfigFile()
		if configPath == "" {
			fmt.Fprintf(os.Stderr, "[ERR!] Could not find packages.yaml\n")
			os.Exit(1)
		}
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERR!] Failed to load config: %v\n", err)
			os.Exit(1)
		}
		for _, pkg := range cfg.GetPackages() {
			fmt.Println(pkg)
		}
		os.Exit(0)
	}

	// --version-check - checks if remote version is newer than local (used by setup.sh)
	if len(os.Args) >= 2 && os.Args[1] == "--version-check" {
		localVer := getLocalVersion()
		remoteVer := getRemoteVersion()
		if localVer == "unknown" || remoteVer == "unknown" {
			os.Exit(1)
		}
		if compareVersions(localVer, remoteVer) < 0 {
			fmt.Printf("Update available: %s -> %s\n", localVer, remoteVer)
			os.Exit(0)
		}
		fmt.Printf("Current: %s\n", localVer)
		os.Exit(1)
	}

	// --wireguard-status - for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--wireguard-status" {
		fmt.Print(wireguard.Status())
		return
	}

	// --update-status - outputs update icon for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--update-status" {
		fmt.Print(update.Get())
		return
	}

	// --update - handle update click
	if len(os.Args) >= 2 && os.Args[1] == "--update" {
		update.Click()
		return
	}

	// --rofi - app launcher
	if len(os.Args) >= 2 && os.Args[1] == "--rofi" {
		if err := rofi.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "rofi error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// --weather - outputs weather icon + temp for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--weather" {
		fmt.Print(weather.Get())
		return
	}

	// --network-wifi - outputs wifi icon for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--network-wifi" {
		fmt.Print(network.GetWifi())
		return
	}

	// --network-eth - outputs eth icon for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--network-eth" {
		fmt.Print(network.GetEth())
		return
	}

	// --wifi-info - shows wifi connection status notification
	if len(os.Args) >= 2 && os.Args[1] == "--wifi-info" {
		details := network.GetWifiDetails()
		icon := "wifi.png"
		if !network.IsConnected() {
			icon = "wifi-off.png"
		}
		exec.Command("/usr/local/bin/notify-send", "-i", getIconPath(icon), "-t", "5000", "WiFi", details).Start()
		return
	}

	// --eth-info - shows ethernet connection status notification
	if len(os.Args) >= 2 && os.Args[1] == "--eth-info" {
		details := network.GetEthDetails()
		icon := "ethernet.png"
		exec.Command("/usr/local/bin/notify-send", "-i", getIconPath(icon), "-t", "5000", "Ethernet", details).Start()
		return
	}

	// --battery - outputs battery icon + percentage for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--battery" {
		fmt.Print(battery.Get())
		return
	}

	// --night-light-status - outputs night light icon for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--night-light-status" {
		fmt.Print(nightlight.Get())
		return
	}

	// --polybar-transmission - outputs transmission icon for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--polybar-transmission" {
		if rofi.IsTransmissionRunning() {
			fmt.Print("󰐻")
		} else {
			fmt.Print("󱧝")
		}
		return
	}

	// --polybar-proton-drive - outputs proton drive icon for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--polybar-proton-drive" {
		if err := polybar.RunProtonDrive(); err != nil {
			fmt.Fprintf(os.Stderr, "polybar proton-drive error: %v\n", err)
		}
		return
	}

	// --proton-drive-sync - sync Proton Drive (click action)
	if len(os.Args) >= 2 && os.Args[1] == "--proton-drive-sync" {
		icon := getIconPath("proton-drive.png")
		if polybar.IsProtonDriveConfigured() {
			// Check sync state
			state := polybar.CheckProtonDriveSyncState()
			if state == "synced" {
				exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "2000", "Proton Drive", "Already Synced ✓").Run()
			} else {
				exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "2000", "Proton Drive", "Syncing...").Run()
				cmd := `echo "Proton Drive is now syncing... be patient..."; rclone bisync ~/ProtonSync proton:ProtonSync --resync --progress; printf "\nDone. Press Enter to close..."; read -r ans`
				exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
			}
		} else {
			exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "5000", "-u", "critical", "Proton Drive", "Not configured").Run()
			exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "5000", "-u", "critical", "Setup Required", "See OpenRiot.org for setup info").Run()
		}
		return
	}

	// --proton-drive-init - populate bisync cache files
	if len(os.Args) >= 2 && os.Args[1] == "--proton-drive-init" {
		icon := getIconPath("proton-drive.png")
		if polybar.IsProtonDriveConfigured() {
			if err := polybar.InitProtonDriveCache(); err != nil {
				fmt.Fprintf(os.Stderr, "proton-drive init error: %v\n", err)
				exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "5000", "-u", "critical", "Proton Drive", "Failed to init cache").Run()
			} else {
				exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "2000", "Proton Drive", "Cache initialized").Run()
			}
		} else {
			exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "5000", "-u", "critical", "Proton Drive", "Not configured").Run()
		}
		return
	}

	// --transmission-stop - stop transmission daemon
	if len(os.Args) >= 2 && os.Args[1] == "--transmission-stop" {
		var icon string
		if rofi.IsTransmissionRunning() {
			icon = getIconPath("transmission-on.png")
			exec.Command("pkill", "-INT", "transmission-daemon").Run()
			exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "2000", "Transmission", "Stopping Transmission...").Run()
		} else {
			icon = getIconPath("transmission-off.png")
			exec.Command("sh", "-c", "mkdir -p ~/.local/share/transmission ~/.config/transmission && transmission-daemon -f --logfile ~/.local/share/transmission/daemon.log &").Run()
			exec.Command("/usr/local/bin/notify-send", "-i", icon, "-t", "2000", "Transmission", "Starting Transmission...").Run()
		}
		return
	}

	// --night-light - toggle night light
	if len(os.Args) >= 2 && os.Args[1] == "--night-light" {
		nightlight.Toggle()
		return
	}

	// --window-icon <class> - outputs icon for window class
	if len(os.Args) >= 2 && os.Args[1] == "--window-icon" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: openriot --window-icon <class>")
			os.Exit(1)
		}
		fmt.Print(windowicon.Get(os.Args[2]))
		return
	}

	// --window-title - outputs focused window title for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--window-title" {
		fmt.Print(windowtitle.Get())
		return
	}

	// --workspace-switch N - switch to workspace N (no-op if already there)
	if len(os.Args) >= 2 && os.Args[1] == "--workspace-switch" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: openriot --workspace-switch <N>")
			os.Exit(1)
		}
		target, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid workspace number")
			os.Exit(1)
		}
		workspace.Switch(target)
		return
	}

	// --wireguard - toggle VPN
	if len(os.Args) >= 2 && os.Args[1] == "--wireguard" {
		if err := wireguard.Toggle(); err != nil {
			fmt.Fprintf(os.Stderr, "WireGuard error: %v\n", err)
			os.Exit(1)
		}
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
		lock.Lock()
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "--suspend" {
		exec.Command("zzz").Run()
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "--power-menu" {
		cmd := exec.Command("rofi", "-dmenu", "-p", "Power: ")
		cmd.Stdin = strings.NewReader("Lock\nSuspend\nReboot\nShutdown\nLogout")
		out, err := cmd.Output()
		if err != nil {
			return
		}
		choice := strings.TrimSpace(string(out))
		switch choice {
		case "Lock":
			lock.Lock()
		case "Suspend":
			lock.Lock()
			exec.Command("zzz").Run()
		case "Reboot":
			exec.Command("shutdown", "-r", "now").Run()
		case "Shutdown":
			exec.Command("shutdown", "-p", "now").Run()
		case "Logout":
			exec.Command("i3-msg", "exit").Run()
		}
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "--wallpaper-next" {
		os.Exit(backgrounds.Next())
	}
	if len(os.Args) >= 2 && os.Args[1] == "--wallpaper-load" {
		os.Exit(backgrounds.Load())
	}
	if len(os.Args) >= 2 && os.Args[1] == "--suspend-if-undocked" {
		detect.SuspendIfUndocked()
		return
	}
	// --notify "title" "body" [--urgency normal|critical|low] [--expires-in seconds] [--icon path]
	if len(os.Args) >= 2 && os.Args[1] == "--notify" {
		title, body, urgency, iconPath := "", "", "normal", ""
		expiresIn := 0
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--urgency" && i+1 < len(os.Args) {
				urgency = os.Args[i+1]
			} else if os.Args[i] == "--expires-in" && i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &expiresIn)
			} else if os.Args[i] == "--icon" && i+1 < len(os.Args) {
				iconPath = os.Args[i+1]
			} else if title == "" {
				title = os.Args[i]
			} else if body == "" {
				body = os.Args[i]
			}
		}
		if title == "" {
			fmt.Fprintln(os.Stderr, "Usage: openriot --notify \"title\" \"body\" [--urgency normal] [--expires-in seconds] [--icon path]")
			os.Exit(1)
		}
		// Call notify-send to display notification
		args := []string{"/usr/local/bin/notify-send"}
		if iconPath != "" {
			args = append(args, "-i", iconPath)
		}
		if urgency != "normal" {
			args = append(args, "-u", urgency)
		}
		if expiresIn > 0 {
			args = append(args, "-t", fmt.Sprintf("%d", expiresIn*1000))
		}
		args = append(args, title)
		if body != "" {
			args = append(args, body)
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Run()
		// Also save to JSON for polybar module
		var expiresAt int64
		if expiresIn > 0 {
			expiresAt = time.Now().Unix() + int64(expiresIn)
		}
		notify.Add(title, body, urgency, expiresAt)
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
	// --notify-dunst (alias: --notify-status)
	if len(os.Args) >= 2 && (os.Args[1] == "--notify-dunst" || os.Args[1] == "--notify-status") {
		if err := notify.Status(); err != nil {
			fmt.Fprintf(os.Stderr, "notify dunst error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// --polybar-metrics outputs CPU and RAM for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--polybar-metrics" {
		if err := polybar.RunMetrics(); err != nil {
			fmt.Fprintf(os.Stderr, "polybar metrics error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// --polybar-volume outputs volume with icon for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--polybar-volume" {
		if err := polybar.RunVolume(); err != nil {
			fmt.Fprintf(os.Stderr, "polybar volume error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// --polybar-memory outputs memory icon for polybar
	if len(os.Args) >= 2 && os.Args[1] == "--polybar-memory" {
		ram := polybar.GetRAM()
		ramPct := polybar.GetMemPercent()
		fmt.Printf(" %s\nMemory: %s\n", ram, ramPct)
		os.Exit(0)
	}

	// --cpu-notify shows CPU usage notification
	if len(os.Args) >= 2 && os.Args[1] == "--cpu-notify" {
		cpuPct := polybar.GetCPUPercent()
		exec.Command("/usr/local/bin/notify-send", "-i", getIconPath("cpu.png"), "-t", "1500", "CPU", cpuPct).Start()
		os.Exit(0)
	}

	// --mem-notify shows memory usage notification
	if len(os.Args) >= 2 && os.Args[1] == "--mem-notify" {
		memDetails := polybar.GetMemDetails()
		exec.Command("/usr/local/bin/notify-send", "-i", getIconPath("memory.png"), "-t", "5000", "Memory", memDetails).Start()
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
	if len(os.Args) >= 2 && os.Args[1] == "--crypto-notify" {
		exec.Command("/usr/local/bin/notify-send", "-i", getIconPath("crypto.png"), "-t", "0", "-r", "1", "Crypto", "Loading...").Start()
		time.Sleep(100 * time.Millisecond)
		if err := crypto.RunCrypto("NOTIFY_SEND"); err != nil {
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
	// --share-log [filename] - upload log to ix.io for sharing
	if len(os.Args) >= 2 && os.Args[1] == "--share-log" {
		filename := "setup.log"
		if len(os.Args) >= 3 {
			filename = os.Args[2]
		}
		if err := shareLog(filename); err != nil {
			fmt.Fprintf(os.Stderr, "share-log error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// --make-icon <name> <symbol> - generate icon PNG
	if len(os.Args) >= 2 && os.Args[1] == "--make-icon" {
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: openriot --make-icon <name> <symbol>\n")
			os.Exit(1)
		}
		name := os.Args[2]
		symbol := os.Args[3]
		if err := makeIcon(name, symbol); err != nil {
			fmt.Fprintf(os.Stderr, "make-icon error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Icon created: %s.png\n", name)
		return
	}

	// No command or unknown command
	fmt.Fprintf(os.Stderr, "openriot %s\n", version)
	fmt.Fprintf(os.Stderr, "Usage: openriot <command>\n")
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  --install          Install OpenRiot (configs, not packages)\n")
	fmt.Fprintf(os.Stderr, "  --install-packages Install packages from packages.yaml\n")
	fmt.Fprintf(os.Stderr, "  --source-builds    Build software from source\n")
	fmt.Fprintf(os.Stderr, "  --packages         List packages from packages.yaml\n")
	fmt.Fprintf(os.Stderr, "  --rofi            Show app launcher\n")
	fmt.Fprintf(os.Stderr, "  --lock            Lock the screen\n")
	fmt.Fprintf(os.Stderr, "  --suspend         Suspend the system\n")
	fmt.Fprintf(os.Stderr, "  --power-menu       Show power menu\n")
	fmt.Fprintf(os.Stderr, "  --volume <args>    Adjust volume\n")
	fmt.Fprintf(os.Stderr, "  --brightness <args> Adjust brightness\n")
	fmt.Fprintf(os.Stderr, "  --notify \"title\" \"body\" Send notification\n")
	fmt.Fprintf(os.Stderr, "  --polybar-metrics    Show CPU/RAM for polybar\n")
	fmt.Fprintf(os.Stderr, "  --polybar-volume    Show volume for polybar\n")
	fmt.Fprintf(os.Stderr, "  --crypto [BTC|ETH] Show crypto prices\n")
	fmt.Fprintf(os.Stderr, "  --share-log [file] Upload log to ix.io for sharing\n")
	fmt.Fprintf(os.Stderr, "  --make-icon <name> <symbol> Generate icon PNG\n")
	fmt.Fprintf(os.Stderr, "  --version         Show version\n")
	os.Exit(1)
}

// shareLog uploads a file to ix.io pastebin for easy sharing
func shareLog(filename string) error {
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".cache", "openriot", filename)
	
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("reading log file: %w", err)
	}

	// Upload to catbox.moe
	cmd := exec.Command("curl", "-s", "-F", "reqtype=fileupload", "-F", "fileToUpload=@-", "https://catbox.moe/user/api.php")
	cmd.Stdin = bytes.NewReader(data)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	url := strings.TrimSpace(string(output))
	fmt.Println(url)
	return nil
}

// makeIcon generates a PNG icon from a Nerd Font symbol
func makeIcon(name, symbol string) error {
	home := os.Getenv("HOME")
	font := filepath.Join(home, ".local/share/fonts/FiraCode/FiraCodeNerdFont-Regular.ttf")
	iconDir := filepath.Join(home, ".local/share/openriot/config/icons")

	// Ensure icon directory exists
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		return fmt.Errorf("creating icon dir: %w", err)
	}

	output := filepath.Join(iconDir, name+".png")
	cmd := exec.Command("convert",
		"-background", "none",
		"-fill", "white",
		"-font", font,
		"-pointsize", "32",
		"label:"+symbol,
		"-resize", "48x48",
		output)
	return cmd.Run()
}

// runInstall handles the --install command (runs as USER, no TTY/PTY needed)
func runInstall() {
	fmt.Println("[INFO] OpenRiot installer starting...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR!] Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR!] Failed to load config from %s: %v\n", configPath, err)
		os.Exit(1)
	}

	// Step 1: Config deployment
	if err := installer.CopyConfigs(repoDir, cfg, testMode); err != nil {
		fmt.Printf("%s[WARN]%s  Config deployment skipped: %v\n", installer.Yellow, installer.Reset, err)
	}

	// Step 2: Command execution
	fmt.Printf("%s[INFO]%s Running post-install commands...\n", installer.Blue, installer.Reset)
	if err := installer.ExecCommands(cfg, testMode); err != nil {
		fmt.Printf("%s[WARN]%s Some commands failed: %v\n", installer.Yellow, installer.Reset, err)
	}

	// Step 3: Source builds (crush, wlsunset, bibata-cursor, etc.)
	fmt.Printf("%s[INFO]%s Running source builds...\n", installer.Blue, installer.Reset)
	if err := installer.SourceBuilds(cfg, testMode); err != nil {
		fmt.Printf("[WARN] Source builds: %v\n", err)
	}

	// Source builds handled above, setup.sh shows completion box
}

// runSourceBuilds runs only the source builds phase (used by setup.sh)
func runSourceBuilds() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR!] Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR!] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := installer.SourceBuilds(cfg, testMode); err != nil {
		fmt.Printf("[WARN] Source builds: %v\n", err)
	}
	fmt.Println("[INFO] Source builds complete!")
}

// runInstallPackages installs packages from packages.yaml (used by setup.sh)
func runInstallPackages() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR!] Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR!] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s[INFO]%s Installing packages from packages.yaml (safe one-by-one mode)...\n", installer.Blue, installer.Reset)

	packages := cfg.GetPackages()
	if len(packages) == 0 {
		fmt.Fprintf(os.Stderr, "%s[ERR!]%s No packages found in packages.yaml\n", installer.Red, installer.Reset)
		os.Exit(1)
	}

	failed, _ := installer.InstallPackages(packages)
	if failed > 0 {
		os.Exit(1)
	}
}



// getLocalVersion reads the local VERSION file
func getLocalVersion() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "unknown"
	}
	versionPath := filepath.Join(homeDir, ".local", "share", "openriot", "VERSION")
	data, err := os.ReadFile(versionPath)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// getRemoteVersion fetches VERSION from openriot.org
func getRemoteVersion() string {
	resp, err := http.Get("https://openriot.org/VERSION")
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// compareVersions compares two semantic versions (a vs b)
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	for i := 0; i < 3; i++ {
		var vA, vB int
		if i < len(partsA) {
			vA, _ = strconv.Atoi(partsA[i])
		}
		if i < len(partsB) {
			vB, _ = strconv.Atoi(partsB[i])
		}
		if vA < vB {
			return -1
		}
		if vA > vB {
			return 1
		}
	}
	return 0
}

// getInstallDir returns the installation directory
func getInstallDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "openriot")
}

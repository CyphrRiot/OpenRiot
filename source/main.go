package main

import (
	"fmt"
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
	"openriot/screenshot"
	"openriot/weather"
	"openriot/workspace"
	"openriot/workspaceicons"
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
	f, err := os.OpenFile("/tmp/openriot_calls.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: cannot open log: %v\n", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format("15:04:05.000"), strings.Join(os.Args[1:], " "))
}

func main() {
	logDebugCall()

	// Command map for CLI dispatcher
	commands := map[string]func(){
		"--version": func() {
			fmt.Println("openriot", version)
			os.Exit(0)
		},
		"--test-notify": func() {
			// Usage: --test-notify {icon} {title} {body} {urgency} {timeoutMs} {replaceID}
			if len(os.Args) < 7 {
				fmt.Fprintf(os.Stderr, "Usage: --test-notify {icon} {title} {body} {urgency} {timeoutMs} {replaceID}\n")
				fmt.Fprintf(os.Stderr, "Example: --test-notify chart \"Test\" \"Hello\" normal 3000 0\n")
				os.Exit(1)
			}
			icon := os.Args[2]
			title := os.Args[3]
			body := os.Args[4]
			urgency := os.Args[5]
			timeout, _ := strconv.Atoi(os.Args[6])
			replaceID, _ := strconv.Atoi(os.Args[7])
			notify.SendNotify(icon, title, body, urgency, timeout, replaceID)
			fmt.Printf("Sent: [%s] %s - %s\n", icon, title, body)
		},
		"--install": func() {
			runInstall()
		},
		"--source-builds": func() {
			installer.RunSourceBuilds(testMode)
		},
		"--install-packages": func() {
			installer.RunInstallPackages()
		},
		"--packages": func() {
			configPath := config.FindConfigFile()
			if configPath == "" {
				fmt.Fprintf(os.Stderr, "[FAIL] Could not find packages.yaml\n")
				os.Exit(1)
			}
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[FAIL] Failed to load config: %v\n", err)
				os.Exit(1)
			}
			for _, pkg := range cfg.GetPackages() {
				fmt.Println(pkg)
			}
			os.Exit(0)
		},
		"--check-packages": func() {
			installer.RunCheckPackages()
		},
		"--sync-packages": func() {
			installer.RunSyncPackages()
		},
		"--version-check": func() {
			localVer := update.GetLocalVersion()
			remoteVer := update.GetRemoteVersion()
			if localVer == "unknown" || remoteVer == "unknown" {
				os.Exit(1)
			}
			if update.CompareVersions(localVer, remoteVer) < 0 {
				fmt.Printf("Update available: %s -> %s\n", localVer, remoteVer)
				os.Exit(0)
			}
			fmt.Printf("Current: %s\n", localVer)
			os.Exit(1)
		},
	}

	// Handle --test/-t flag first (affects other commands)
	for _, arg := range os.Args[1:] {
		if arg == "--test" || arg == "-t" {
			testMode = true
		}
	}

	// Dispatch command from map
	if len(os.Args) >= 2 {
		if fn, ok := commands[os.Args[1]]; ok {
			fn()
			return
		}
	}

	// --wireguard-status - for polybar
	commands["--wireguard-status"] = func() {
		fmt.Print(wireguard.Status())
	}

	// --update-status - outputs update icon for polybar
	commands["--update-status"] = func() {
		fmt.Print(update.Get())
	}

	// --update - handle update click
	commands["--update"] = func() {
		update.Click()
	}

	// --rofi - app launcher
	commands["--rofi"] = func() {
		if err := rofi.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "rofi error: %v\n", err)
			os.Exit(1)
		}
	}

	// --weather - outputs weather icon + temp for polybar
	commands["--weather"] = func() {
		fmt.Print(weather.Get())
	}

	// Dispatch Section 2 commands
	if len(os.Args) >= 2 {
		if fn, ok := commands[os.Args[1]]; ok {
			fn()
			return
		}
	}

	// Section 3: Network, battery, and status commands
	commands["--network-wifi"] = func() {
		fmt.Print(network.GetWifi())
	}
	commands["--network-eth"] = func() {
		fmt.Print(network.GetEth())
	}
	commands["--wifi-info"] = func() {
		details := network.GetWifiDetails()
		icon := "wifi.png"
		if !network.IsConnected() {
			icon = "wifi-off.png"
		}
		notify.SendNotify(icon, "WiFi", details, "normal", 5000, 0)
	}
	commands["--wifi-reconnect"] = func() {
		if network.IsOnline() {
			notify.SendNotify("wifi", "WiFi", "Already connected", "normal", 2000, 0)
			return
		}
		if !network.IsConnected() {
			notify.SendNotify("wifi-off", "WiFi", "Not connected", "normal", 2000, 0)
			return
		}
		notify.SendNotify("wifi", "WiFi", "Reconnecting...", "normal", 3000, 0)
		if err := network.ReconnectWifi(); err != nil {
			notify.SendNotify("wifi-off", "WiFi", "Reconnect failed: "+err.Error(), "normal", 5000, 0)
		}
	}
	commands["--eth-info"] = func() {
		details := network.GetEthDetails()
		notify.SendNotify("ethernet", "Ethernet", details, "normal", 5000, 0)
	}
	commands["--battery"] = func() {
		fmt.Print(battery.Get())
	}
	commands["--battery-notify"] = func() {
		batteryDetails := battery.GetNotifyDetails()
		notify.SendNotify("battery", "Battery", batteryDetails, "normal", 5000, 0)
		os.Exit(0)
	}
	commands["--night-light-status"] = func() {
		fmt.Print(nightlight.Get())
	}
	commands["--polybar-transmission"] = func() {
		if rofi.IsTransmissionRunning() {
			fmt.Print("󰐻")
		} else {
			fmt.Print("󱧝")
		}
	}
	commands["--polybar-proton-drive"] = func() {
		if err := polybar.RunProtonDrive(); err != nil {
			fmt.Fprintf(os.Stderr, "polybar proton-drive error: %v\n", err)
		}
	}

	// Dispatch Section 3 commands
	if len(os.Args) >= 2 {
		if fn, ok := commands[os.Args[1]]; ok {
			fn()
			return
		}
	}

	commands["--proton-drive-sync"] = func() {
		if polybar.IsProtonDriveConfigured() {
			state := polybar.CheckProtonDriveSyncState()
			if state == "synced" {
				notify.SendNotify("proton-drive", "Proton Drive", "Synchronized: "+polybar.GetProtonDriveTooltipText(), "normal", 5000, 0)
				return
			}
			notify.SendNotify("proton-drive", "Proton Drive", "Syncing...", "normal", 2000, 0)
			cmd := `printf "Proton Drive Sync\nFrom: ~/ProtonDrive -> Proton Drive Cloud\n\nWould you like to do a bi-directional Sync or one-way\n  and replace items in the Cloud with local items?\n\n[Y]es for bi-directional sync (or ENTER),\n[O]ne-way for One-Way sync or\n[Q]uit or [N]o ?\n\nChoose your adventure [Y/o/q/n] -> "; read -r ans; case "$ans" in o|O) echo "One-way sync selected..."; rclone copy ~/ProtonSync proton:ProtonSync --progress; printf "\nDone. Press Enter to close..."; read -r ans ;; [yY]|"") echo "Bi-directional sync selected..."; rclone bisync ~/ProtonSync proton:ProtonSync --resync --progress; printf "\nDone. Press Enter to close..."; read -r ans ;; *) echo "Canceled."; sleep 1 ;; esac`
			exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
		} else {
			notify.SendNotify("proton-drive", "Proton Drive", "Not configured\nSee OpenRiot.org for setup info", "critical", 5000, 0)
		}
	}
	commands["--proton-drive-init"] = func() {
		if polybar.IsProtonDriveConfigured() {
			if err := polybar.InitProtonDriveCache(); err != nil {
				fmt.Fprintf(os.Stderr, "proton-drive init error: %v\n", err)
				notify.SendNotify("proton-drive", "Proton Drive", "Failed to init cache", "critical", 5000, 0)
			} else {
				notify.SendNotify("proton-drive", "Proton Drive", "Cache initialized", "normal", 2000, 0)
			}
		} else {
			notify.SendNotify("proton-drive", "Proton Drive", "Not configured", "critical", 5000, 0)
		}
	}
	commands["--transmission-toggle"] = func() {
		if rofi.IsTransmissionRunning() {
			exec.Command("pkill", "-INT", "transmission-daemon").Run()
			notify.SendNotify("transmission", "Transmission", "Stopping Transmission...", "normal", 2000, 0)
		} else {
			exec.Command("sh", "-c", "mkdir -p ~/.local/share/transmission ~/.config/transmission && transmission-daemon -f --logfile ~/.local/share/transmission/daemon.log &").Run()
			notify.SendNotify("transmission", "Transmission", "Starting Transmission...", "normal", 2000, 0)
		}
	}
	commands["--night-light"] = func() {
		nightlight.Toggle()
	}
	commands["--window-title"] = func() {
		fmt.Print(windowtitle.Get())
	}
	commands["--wireguard"] = func() {
		if err := wireguard.Toggle(); err != nil {
			fmt.Fprintf(os.Stderr, "WireGuard error: %v\n", err)
			os.Exit(1)
		}
	}

	// Dispatch Section 4 commands
	if len(os.Args) >= 2 {
		if fn, ok := commands[os.Args[1]]; ok {
			fn()
			return
		}
	}

	// --window-icon <class> - outputs icon for window class (requires arg)
	commands["--window-icon"] = func() {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: openriot --window-icon <class>")
			os.Exit(1)
		}
		fmt.Print(windowicon.Get(os.Args[2]))
	}

	// --all-window-icons - outputs all window class→icon mappings (one i3-msg call)
	commands["--all-window-icons"] = func() {
		windows := windowicon.GetAllWindowIcons()
		for class, icon := range windows {
			fmt.Printf("%s=%s\n", class, icon)
		}
	}
	// --workspace-switch N - switch to workspace N (requires arg)
	commands["--workspace-switch"] = func() {
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
	}

	// --workspace-icons N - output workspace icons for polybar (requires arg)
	commands["--workspace-icons"] = func() {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: openriot --workspace-icons <N>")
			os.Exit(1)
		}
		target, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid workspace number")
			os.Exit(1)
		}
		fmt.Print(workspaceicons.Get(target))
	}

	// Dispatch commands with args
	if len(os.Args) >= 2 {
		if fn, ok := commands[os.Args[1]]; ok {
			fn()
			return
		}
	}

	// Section 5: Volume, brightness, lock, power menu
	commands["--lock"] = func() {
		lock.Lock()
	}
	commands["--signal-launch"] = func() {
		home, _ := os.UserHomeDir()
		cmd := exec.Command("pgrep", "-f", "gurk")
		output, _ := cmd.Output()
		if len(strings.TrimSpace(string(output))) > 0 {
			// Already running
			notify.SendNotify("signal", "Signal", "Signal already launched", "normal", 2000, 0)
			return
		}
		// Launch Signal
		notify.SendNotify("signal", "Signal", "Starting Signal...", "normal", 2000, 0)
		exec.Command("alacritty", "--class", "gurk", "--title", "Signal", "-e", home+"/.local/share/openriot/config/bin/gurk").Start()
	}
	commands["--browser"] = func() {
		cmd := exec.Command("pgrep", "-f", "firefox")
		output, _ := cmd.Output()
		if len(strings.TrimSpace(string(output))) > 0 {
			notify.SendNotify("firefox", "Firefox", "Already running", "normal", 2000, 0)
			return
		}
		notify.SendNotify("firefox", "Firefox", "Starting Browser...", "normal", 2000, 0)
		exec.Command("firefox", os.Args[2:]...).Start()
	}
	commands["--proton"] = func() {
		notify.SendNotify("proton-mail", "Proton Mail", "Opening...", "normal", 2000, 0)
		exec.Command("firefox", "https://mail.proton.me/u/11/inbox").Start()
	}
	commands["--twitter"] = func() {
		notify.SendNotify("twitter", "X (Twitter)", "Opening...", "normal", 2000, 0)
		exec.Command("firefox", "https://x.com/").Start()
	}
	commands["--crush"] = func() {
		home, _ := os.UserHomeDir()
		cmd := exec.Command("pgrep", "-f", "crush")
		output, _ := cmd.Output()
		if len(strings.TrimSpace(string(output))) > 0 {
			notify.SendNotify("crush", "Crush AI", "Already running", "normal", 2000, 0)
			return
		}
		notify.SendNotify("crush", "Crush AI", "Starting Crush...", "normal", 2000, 0)
		exec.Command("alacritty", "--class", "crush", "--title", "Crush AI", "-e", home+"/.local/bin/crush").Start()
	}
	commands["--suspend"] = func() {
		exec.Command("zzz").Run()
	}
	commands["--power-menu"] = func() {
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
			notify.SendNotify("power", "Power", "Suspending...", "normal", 2000, 0)
			lock.Lock()
			exec.Command("doas", "zzz").Run()
		case "Reboot":
			notify.SendNotify("power", "Power", "Rebooting...", "normal", 3000, 0)
			exec.Command("doas", "shutdown", "-r", "now").Run()
		case "Shutdown":
			notify.SendNotify("power", "Power", "Shutting down...", "normal", 5000, 0)
			exec.Command("doas", "shutdown", "-p", "now").Run()
		case "Logout":
			notify.SendNotify("power", "Power", "Logging out...", "normal", 2000, 0)
			exec.Command("i3-msg", "exit").Run()
		}
	}
	commands["--wallpaper-next"] = func() {
		os.Exit(backgrounds.Next())
	}
	commands["--wallpaper-load"] = func() {
		os.Exit(backgrounds.Load())
	}
	commands["--polybar-setup"] = func() {
		os.Exit(polybar.Setup())
	}
	commands["--dunst-setup"] = func() {
		os.Exit(notify.Setup())
	}
	commands["--suspend-if-undocked"] = func() {
		detect.SuspendIfUndocked()
	}
	commands["--screenshot"] = func() {
		// --screenshot with no args = full screen
		// --screenshot select = area selection
		selectArea := len(os.Args) >= 3 && os.Args[2] == "select"
		if err := screenshot.Run(selectArea); err != nil {
			fmt.Fprintf(os.Stderr, "Screenshot failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Dispatch Section 5 commands
	if len(os.Args) >= 2 {
		if fn, ok := commands[os.Args[1]]; ok {
			fn()
			return
		}
	}

	// Commands with args (volume, brightness)
	if len(os.Args) >= 2 && os.Args[1] == "--volume" {
		os.Exit(audio.Run(os.Args[2:]))
	}
	if len(os.Args) >= 2 && os.Args[1] == "--brightness" {
		os.Exit(display.Run(os.Args[2:]))
	}

	// Section 6: Notify, polybar metrics, crypto commands
	commands["--notify-dismiss"] = func() {
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
	commands["--notify-clear"] = func() {
		if err := notify.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "notify clear error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	commands["--notify-dunst"] = func() {
		if err := notify.Status(); err != nil {
			fmt.Fprintf(os.Stderr, "notify dunst error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	commands["--notify-status"] = func() {
		if err := notify.Status(); err != nil {
			fmt.Fprintf(os.Stderr, "notify dunst error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	commands["--polybar-metrics"] = func() {
		if err := polybar.RunMetrics(); err != nil {
			fmt.Fprintf(os.Stderr, "polybar metrics error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	commands["--polybar-volume"] = func() {
		if err := polybar.RunVolume(); err != nil {
			fmt.Fprintf(os.Stderr, "polybar volume error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	commands["--polybar-memory"] = func() {
		ram := polybar.GetRAM()
		ramPct := polybar.GetMemPercent()
		fmt.Printf(" %s\nMemory: %s\n", ram, ramPct)
		os.Exit(0)
	}
	commands["--cpu-notify"] = func() {
		cpuDetails := polybar.GetCPUDetails()
		notify.SendNotify("cpu", "CPU", cpuDetails, "normal", 5000, 0)
		os.Exit(0)
	}
	commands["--mem-notify"] = func() {
		memDetails := polybar.GetMemDetails()
		notify.SendNotify("memory", "Memory", memDetails, "normal", 5000, 0)
		os.Exit(0)
	}
	commands["--crypto-notify"] = func() {
		notify.SendNotify("chart", "Crypto", "Loading...", "normal", 0, 1)
		time.Sleep(1 * time.Second)
		if err := crypto.RunCrypto("NOTIFY_SEND"); err != nil {
			fmt.Fprintf(os.Stderr, "crypto error: %v\n", err)
		}
	}
	commands["--crypto-refresh"] = func() {
		os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "openriot-crypto.json"))
		os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "openriot-crypto-prev.json"))
		if err := crypto.RunCrypto("ROWML"); err != nil {
			fmt.Fprintf(os.Stderr, "crypto error: %v\n", err)
		}
	}
	// Commands with args: --notify, --crypto, --share-log, --make-icon

	// Dispatch Section 6 commands
	if len(os.Args) >= 2 {
		if fn, ok := commands[os.Args[1]]; ok {
			fn()
			return
		}
	}

	// Commands with args (keep as if-statements)
	// --notify "title" "body" [--urgency...] [--expires-in...] [--icon...]
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
		var expiresAt int64
		if expiresIn > 0 {
			expiresAt = time.Now().Unix() + int64(expiresIn)
		}
		notify.Add(title, body, urgency, expiresAt)
		os.Exit(0)
	}
	// --crypto [BTC|ETH]
	if len(os.Args) >= 2 && os.Args[1] == "--crypto" {
		mode := "BTC"
		if len(os.Args) >= 3 {
			mode = os.Args[2]
		}
		if err := crypto.RunCrypto(mode); err != nil {
			fmt.Fprintf(os.Stderr, "crypto error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// --share-log [filename]
	if len(os.Args) >= 2 && os.Args[1] == "--share-log" {
		filename := "setup.log"
		if len(os.Args) >= 3 {
			filename = os.Args[2]
		}
		if err := installer.ShareLog(filename); err != nil {
			fmt.Fprintf(os.Stderr, "share-log error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// --make-icon <name> <symbol>
	if len(os.Args) >= 2 && os.Args[1] == "--make-icon" {
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: openriot --make-icon <name> <symbol>\n")
			os.Exit(1)
		}
		name := os.Args[2]
		symbol := os.Args[3]
		if err := installer.MakeIcon(name, symbol); err != nil {
			fmt.Fprintf(os.Stderr, "make-icon error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Icon created: %s.png\n", name)
		os.Exit(0)
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
	fmt.Fprintf(os.Stderr, "  --screenshot [select]  Take screenshot (use 'select' for area)\n")
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



// runInstall handles the --install command (runs as USER, no TTY/PTY needed)
func runInstall() {
	fmt.Println("[INFO] OpenRiot installer starting...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Failed to load config from %s: %v\n", configPath, err)
		os.Exit(1)
	}

	// Step 0: Package installation
	fmt.Printf("%s[INFO]%s Installing packages...\n", installer.Cyan, installer.Reset)
	installer.RunInstallPackages()

	// Step 1: Config deployment
	if err := installer.CopyConfigs(repoDir, cfg, testMode); err != nil {
		fmt.Printf("%s[WARN]%s  Config deployment skipped: %v\n", installer.Yellow, installer.Reset, err)
	}

	// Step 2: Command execution
	fmt.Printf("%s[INFO]%s Running post-install commands...\n", installer.Cyan, installer.Reset)
	if err := installer.ExecCommands(cfg, testMode); err != nil {
		fmt.Printf("%s[WARN]%s Some commands failed: %v\n", installer.Yellow, installer.Reset, err)
	}

	// Step 3: Source builds (crush, wlsunset, bibata-cursor, etc.)
	fmt.Printf("%s[INFO]%s Running source builds...\n", installer.Cyan, installer.Reset)
	if err := installer.SourceBuilds(cfg, testMode); err != nil {
		fmt.Printf("[WARN] Source builds: %v\n", err)
	}

	// Source builds handled above, setup.sh shows completion box
}

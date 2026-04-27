package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"openriot/assets"
	"openriot/audio"
	"openriot/backgrounds"
	"openriot/battery"
	"openriot/config"
	"openriot/crypto"
	"openriot/detect"
	"openriot/display"
	"openriot/fonts"
	"openriot/gurk"
	"openriot/imaging"
	"openriot/installer"
	"openriot/lock"
	"openriot/logger"
	"openriot/macspoof"
	"openriot/network"
	"openriot/nightlight"
	"openriot/notify"
	"openriot/polybar"
	"openriot/rofi"
	"openriot/roficalc"
	"openriot/screenshot"
	"openriot/update"
	"openriot/weather"
	"openriot/windowicon"
	"openriot/windowtitle"
	"openriot/wireguard"
	"openriot/workspace"
	"openriot/workspaceicons"
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

	// Handle --test/-t flag first (affects other commands)
	for _, arg := range os.Args[1:] {
		if arg == "--test" || arg == "-t" {
			testMode = true
		}
	}

	commands := initCommands()

	if len(os.Args) >= 2 {
		if fn, ok := commands[os.Args[1]]; ok {
			fn()
			return
		}
	}

	printUsage()
}

// hasAudioPlaying checks if any audio is playing via sndio (OpenBSD) or PulseAudio (Linux)
func hasAudioPlaying() bool {
	cmd := exec.Command("sndioctl", "-n")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "app/") && strings.Contains(line, ".level=") {
				return true
			}
		}
	}
	cmd = exec.Command("pactl", "list", "sink-inputs")
	output, err = cmd.Output()
	if err == nil && strings.Contains(string(output), "State: RUNNING") {
		return true
	}
	return false
}

func initVersionInstallCommands(cmds map[string]func()) {
	cmds["--version"] = func() {
			fmt.Println("openriot", version)
			os.Exit(0)
	}
	cmds["--install"] = func() {
			runInstall()
	}
}

func initPackageManagementCommands(cmds map[string]func()) {
	cmds["--source-builds"] = func() {
			installer.RunSourceBuilds(testMode)
	}
	cmds["--install-packages"] = func() {
			installer.RunInstallPackages()
	}
	cmds["--packages"] = func() {
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
	}
	cmds["--check-packages"] = func() {
			installer.RunCheckPackages()
	}
	cmds["--sync-packages"] = func() {
			installer.RunSyncPackages()
	}
	cmds["--check-dependencies"] = func() {
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
			refs, err := cfg.GetAllModulesOrdered()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[FAIL] Dependency error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Dependency order:")
			for i, ref := range refs {
				deps := ""
				if len(ref.Module.Depends) > 0 {
					deps = fmt.Sprintf("  (depends on: %s)", strings.Join(ref.Module.Depends, ", "))
				}
				fmt.Printf("  %d. %s.%s%s\n", i+1, ref.Category, ref.Name, deps)
			}
			os.Exit(0)
	}
	cmds["--mirrors"] = func() {
			installer.RunMirrors()
	}
}

func initToolsUpgradesCommands(cmds map[string]func()) {
	cmds["--random-mac"] = func() {
			if err := macspoof.Run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "random-mac error: %v\n", err)
				os.Exit(1)
			}
	}
	cmds["--crush-upgrade"] = func() {
			installer.NewCrushUpgrade().Run()
	}
	cmds["--gurk-setup"] = func() {
			if err := gurk.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "gurk-setup error: %v\n", err)
				os.Exit(1)
			}
	}
	cmds["--benchmark"] = func() {
			home, _ := os.UserHomeDir()
			cmd := home + "/.local/bin/benchmark; printf \"\\n\\nPress any key to continue...\"; read -r ans"
			exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
	}
	cmds["--install-asset"] = func() {
			if err := assets.Run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "install-asset error: %v\n", err)
				os.Exit(1)
			}
	}
	cmds["--install-fonts"] = func() {
			if err := fonts.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "install-fonts error: %v\n", err)
				os.Exit(1)
			}
	}
	cmds["--install-rofi-calc"] = func() {
			if err := roficalc.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "install-rofi-calc error: %v\n", err)
				os.Exit(1)
			}
	}
}

func initVersionCheckImagingCommands(cmds map[string]func()) {
	cmds["--version-check"] = func() {
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
	}
	cmds["--make-image"] = func() {
			imaging.RunMakeImage(os.Args[2:])
	}
}

func initSystemControlsCommands(cmds map[string]func()) {
	cmds["--volume"] = func() {
			os.Exit(audio.Run(os.Args[2:]))
	}
	cmds["--brightness"] = func() {
			os.Exit(display.Run(os.Args[2:]))
	}
	cmds["--notify"] = func() {
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
	cmds["--crypto"] = func() {
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
	cmds["--share-log"] = func() {
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
	cmds["--make-icon"] = func() {
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
}

func initPolybarStatusCommands(cmds map[string]func()) {
	cmds["--wireguard-status"] = func() {
			fmt.Print(wireguard.Status())
	}
	cmds["--stealth-status"] = func() {
			fmt.Print(macspoof.StealthStatus())
	}
	cmds["--stealth"] = func() {
			notify.SendNotify("stealth", "Stealth", " Restarting Networking Services", "normal", 5000, 0)
			if err := macspoof.StealthToggle(); err != nil {
				notify.SendNotify("stealth", "Stealth", "Failed: "+err.Error(), "critical", 5000, 0)
				os.Exit(1)
			}
			enabled := macspoof.IsStealthEnabled()
			if enabled {
				notify.SendNotify("stealth", "Stealth", "Enabled [Stealth]", "normal", 3000, 0)
			} else {
				notify.SendNotify("stealth", "Stealth", "Disabled", "normal", 3000, 0)
			}
	}
	cmds["--update-status"] = func() {
			fmt.Print(update.Get())
	}
	cmds["--update"] = func() {
			update.Click()
	}
	cmds["--rofi"] = func() {
			if err := rofi.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "rofi error: %v\n", err)
				os.Exit(1)
			}
	}
	cmds["--weather"] = func() {
			fmt.Print(weather.Get())
	}
}

func initNetworkBatteryCommands(cmds map[string]func()) {
	cmds["--network-wifi"] = func() {
			fmt.Print(network.GetWifi())
	}
	cmds["--network-eth"] = func() {
			fmt.Print(network.GetEth())
	}
	cmds["--wifi-info"] = func() {
			details := network.GetWifiDetails()
			icon := "wifi.png"
			if !network.IsConnected() {
				icon = "wifi-off.png"
			}
			notify.SendNotify(icon, "WiFi", details, "normal", 5000, 0)
	}
	cmds["--wifi-reconnect"] = func() {
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
	cmds["--eth-info"] = func() {
			details := network.GetEthDetails()
			notify.SendNotify("ethernet", "Ethernet", details, "normal", 5000, 0)
	}
	cmds["--battery"] = func() {
			fmt.Print(battery.Get())
	}
	cmds["--battery-notify"] = func() {
			batteryDetails := battery.GetNotifyDetails()
			notify.SendNotify("battery", "Battery", batteryDetails, "normal", 5000, 0)
			os.Exit(0)
	}
	cmds["--night-light-status"] = func() {
			fmt.Print(nightlight.Get())
	}
	cmds["--laptop-monitor"] = func() {
			display.RunLaptopMonitor()
	}
	cmds["--laptop-monitor-toggle"] = func() {
			display.ToggleLaptopMonitor()
	}
	cmds["--polybar-transmission"] = func() {
			if rofi.IsTransmissionRunning() {
				fmt.Print("󰐻")
			} else {
				fmt.Print("󱧝")
			}
	}
	cmds["--polybar-proton-drive"] = func() {
			if err := polybar.RunProtonDrive(); err != nil {
				fmt.Fprintf(os.Stderr, "polybar proton-drive error: %v\n", err)
			}
	}
}

func initDriveSyncCommands(cmds map[string]func()) {
	cmds["--proton-drive-sync"] = func() {
			if polybar.IsProtonDriveConfigured() {
				state := polybar.CheckProtonDriveSyncState()
				if state == "synced" {
					notify.SendNotify("proton-drive", "Proton Drive", "Synchronized: "+polybar.GetProtonDriveTooltipText(), "normal", 5000, 0)
					return
				}
				notify.SendNotify("proton-drive", "Proton Drive", "Syncing...", "normal", 2000, 0)
				cmd := `printf "Proton Drive Sync\nFrom: ~/Documents/ProtonSync -> Proton Drive Cloud\n\nWould you like to do a bi-directional Sync or one-way\n  and replace items in the Cloud with local items?\n\n[Y]es for bi-directional sync (or ENTER),\n[O]ne-way for One-Way sync or\n[Q]uit or [N]o ?\n\nChoose your adventure [Y/o/q/n] -> "; read -r ans; case "$ans" in o|O) echo "One-way sync selected..."; rclone copy ~/Documents/ProtonSync proton:ProtonSync --progress; printf "\nDone. Press Enter to close..."; read -r ans ;; [yY]|"") echo "Bi-directional sync selected..."; rclone bisync ~/Documents/ProtonSync proton:ProtonSync --resync --progress; printf "\nDone. Press Enter to close..."; read -r ans ;; *) echo "Canceled."; sleep 1 ;; esac`
				exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
			} else {
				notify.SendNotify("proton-drive", "Proton Drive", "Not configured\nSee OpenRiot.org for setup info", "critical", 5000, 0)
			}
	}
	cmds["--proton-drive-init"] = func() {
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
	cmds["--transmission-toggle"] = func() {
			if rofi.IsTransmissionRunning() {
				exec.Command("pkill", "-INT", "transmission-daemon").Run()
				notify.SendNotify("transmission", "Transmission", "Stopping Transmission...", "normal", 2000, 0)
			} else {
				exec.Command("sh", "-c", "mkdir -p ~/.local/share/transmission ~/.config/transmission && transmission-daemon -f --logfile ~/.local/share/transmission/daemon.log &").Run()
				notify.SendNotify("transmission", "Transmission", "Starting Transmission...", "normal", 2000, 0)
			}
	}
	cmds["--night-light"] = func() {
			nightlight.Toggle()
	}
	cmds["--window-title"] = func() {
			fmt.Print(windowtitle.Get())
	}
	cmds["--wireguard"] = func() {
			if err := wireguard.Toggle(); err != nil {
				fmt.Fprintf(os.Stderr, "WireGuard error: %v\n", err)
				os.Exit(1)
			}
	}
}

func initWindowWorkspaceCommands(cmds map[string]func()) {
	cmds["--window-icon"] = func() {
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: openriot --window-icon <class>")
				os.Exit(1)
			}
			fmt.Print(windowicon.Get(os.Args[2]))
	}
	cmds["--all-window-icons"] = func() {
			windows := windowicon.GetAllWindowIcons()
			for class, icon := range windows {
				fmt.Printf("%s=%s\n", class, icon)
			}
	}
	cmds["--workspace-switch"] = func() {
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
	cmds["--workspace-move"] = func() {
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: openriot --workspace-move <N>")
				os.Exit(1)
			}
			target, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Invalid workspace number")
				os.Exit(1)
			}
			workspace.Move(target)
	}
	cmds["--workspace-icons"] = func() {
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
	cmds["--workspace-icons-all"] = func() {
			fmt.Print(workspaceicons.GetAll())
	}
}

func initLockPowerAppsCommands(cmds map[string]func()) {
	cmds["--lock"] = func() {
			lock.Lock()
	}
	cmds["--smart-lock"] = func() {
			if hasAudioPlaying() {
				return
			}
			players := []string{"firefox", "mpv", "vlc", "mplayer", "chrome", "chromium"}
			for _, p := range players {
				cmd := exec.Command("pgrep", "-x", p)
				if output, _ := cmd.Output(); len(strings.TrimSpace(string(output))) > 0 {
					return
				}
			}
			lock.Lock()
	}
	cmds["--signal-launch"] = func() {
			home, _ := os.UserHomeDir()
			cmd := exec.Command("pgrep", "-f", "gurk")
			output, _ := cmd.Output()
			if len(strings.TrimSpace(string(output))) > 0 {
				notify.SendNotify("signal", "Signal", "Signal already launched", "normal", 2000, 0)
				return
			}
			notify.SendNotify("signal", "Signal", "Starting Signal...", "normal", 2000, 0)
			exec.Command("alacritty", "--class", "gurk", "--title", "Signal", "-e", home+"/.local/share/openriot/config/bin/gurk").Start()
	}
	cmds["--browser"] = func() {
			cmd := exec.Command("pgrep", "-f", "firefox")
			output, _ := cmd.Output()
			if len(strings.TrimSpace(string(output))) > 0 {
				notify.SendNotify("firefox", "Firefox", "Already running", "normal", 2000, 0)
				return
			}
			notify.SendNotify("firefox", "Firefox", "Starting Browser...", "normal", 2000, 0)
			exec.Command("firefox", os.Args[2:]...).Start()
	}
	cmds["--proton"] = func() {
			notify.SendNotify("proton-mail", "Proton Mail", "Opening...", "normal", 2000, 0)
			exec.Command("firefox", "https://mail.proton.me/u/11/inbox").Start()
	}
	cmds["--twitter"] = func() {
			notify.SendNotify("twitter", "X (Twitter)", "Opening...", "normal", 2000, 0)
			exec.Command("firefox", "https://x.com/").Start()
	}
	cmds["--crush"] = func() {
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
	cmds["--suspend"] = func() {
			exec.Command("zzz").Run()
	}
	cmds["--power-menu"] = func() {
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
	cmds["--wallpaper-next"] = func() {
			os.Exit(backgrounds.Next())
	}
	cmds["--wallpaper-prev"] = func() {
			os.Exit(backgrounds.Prev())
	}
	cmds["--wallpaper-load"] = func() {
			os.Exit(backgrounds.Load())
	}
	cmds["--polybar-setup"] = func() {
			os.Exit(polybar.Setup())
	}
	cmds["--dunst-setup"] = func() {
			os.Exit(notify.Setup())
	}
	cmds["--suspend-if-undocked"] = func() {
			detect.SuspendIfUndocked()
	}
	cmds["--screenshot"] = func() {
			selectArea := len(os.Args) >= 3 && os.Args[2] == "select"
			if err := screenshot.Run(selectArea); err != nil {
				fmt.Fprintf(os.Stderr, "Screenshot failed: %v\n", err)
				os.Exit(1)
			}
	}
}

func initNotificationsMetricsCommands(cmds map[string]func()) {
	cmds["--notify-dismiss"] = func() {
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
	cmds["--notify-clear"] = func() {
			if err := notify.Clear(); err != nil {
				fmt.Fprintf(os.Stderr, "notify clear error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
	}
	cmds["--notify-dunst"] = func() {
			if err := notify.Status(); err != nil {
				fmt.Fprintf(os.Stderr, "notify dunst error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
	}
	cmds["--notify-status"] = func() {
			if err := notify.Status(); err != nil {
				fmt.Fprintf(os.Stderr, "notify dunst error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
	}
	cmds["--polybar-metrics"] = func() {
			if err := polybar.RunMetrics(); err != nil {
				fmt.Fprintf(os.Stderr, "polybar metrics error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
	}
	cmds["--polybar-volume"] = func() {
			if err := polybar.RunVolume(); err != nil {
				fmt.Fprintf(os.Stderr, "polybar volume error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
	}
	cmds["--polybar-memory"] = func() {
			ram := polybar.GetRAM()
			ramPct := polybar.GetMemPercent()
			fmt.Printf(" %s\nMemory: %s\n", ram, ramPct)
			os.Exit(0)
	}
	cmds["--cpu-notify"] = func() {
			cpuDetails := polybar.GetCPUDetails()
			notify.SendNotify("cpu", "CPU", cpuDetails, "normal", 5000, 0)
			os.Exit(0)
	}
	cmds["--mem-notify"] = func() {
			memDetails := polybar.GetMemDetails()
			notify.SendNotify("memory", "Memory", memDetails, "normal", 5000, 0)
			os.Exit(0)
	}
	cmds["--crypto-notify"] = func() {
			notify.SendNotify("chart", "Crypto", "Loading...", "normal", 0, 1)
			time.Sleep(1 * time.Second)
			if err := crypto.RunCrypto("NOTIFY_SEND"); err != nil {
				fmt.Fprintf(os.Stderr, "crypto error: %v\n", err)
			}
	}
	cmds["--crypto-refresh"] = func() {
			os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "openriot-crypto.json"))
			os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "openriot-crypto-prev.json"))
			if err := crypto.RunCrypto("ROWML"); err != nil {
				fmt.Fprintf(os.Stderr, "crypto error: %v\n", err)
			}
	}
}

func initCommands() map[string]func() {
	cmds := make(map[string]func())
	initVersionInstallCommands(cmds)
	initPackageManagementCommands(cmds)
	initToolsUpgradesCommands(cmds)
	initVersionCheckImagingCommands(cmds)
	initSystemControlsCommands(cmds)
	initPolybarStatusCommands(cmds)
	initNetworkBatteryCommands(cmds)
	initDriveSyncCommands(cmds)
	initWindowWorkspaceCommands(cmds)
	initLockPowerAppsCommands(cmds)
	initNotificationsMetricsCommands(cmds)
	return cmds
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "openriot %s\n", version)
	fmt.Fprintf(os.Stderr, "Usage: openriot <command>\n")
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  --install          Install OpenRiot (configs, not packages)\n")
	fmt.Fprintf(os.Stderr, "  --install-packages Install packages from packages.yaml\n")
	fmt.Fprintf(os.Stderr, "  --source-builds    Build software from source\n")
	fmt.Fprintf(os.Stderr, "  --packages         List packages from packages.yaml\n")
	fmt.Fprintf(os.Stderr, "  --check-packages   Verify installed packages match yaml\n")
	fmt.Fprintf(os.Stderr, "  --sync-packages    Update packages.yaml to installed versions\n")
	fmt.Fprintf(os.Stderr, "  --check-dependencies Show module dependency order\n")
	fmt.Fprintf(os.Stderr, "  --mirrors          Detect and show fastest OpenBSD mirror\n")
	fmt.Fprintf(os.Stderr, "  --random-mac <subcommand>  Manage MAC address randomization\n")
	fmt.Fprintf(os.Stderr, "  --crush-upgrade    Upgrade crush CLI to latest version\n")
	fmt.Fprintf(os.Stderr, "  --rofi             Show app launcher\n")
	fmt.Fprintf(os.Stderr, "  --lock             Lock the screen\n")
	fmt.Fprintf(os.Stderr, "  --smart-lock       Lock only if no audio playing\n")
	fmt.Fprintf(os.Stderr, "  --suspend          Suspend the system\n")
	fmt.Fprintf(os.Stderr, "  --screenshot [select]  Take screenshot (use 'select' for area)\n")
	fmt.Fprintf(os.Stderr, "  --power-menu       Show power menu\n")
	fmt.Fprintf(os.Stderr, "  --volume <args>    Adjust volume\n")
	fmt.Fprintf(os.Stderr, "  --brightness <args> Adjust brightness\n")
	fmt.Fprintf(os.Stderr, "  --notify \"title\" \"body\" Send notification\n")
	fmt.Fprintf(os.Stderr, "  --polybar-metrics  Show CPU/RAM for polybar\n")
	fmt.Fprintf(os.Stderr, "  --polybar-volume   Show volume for polybar\n")
	fmt.Fprintf(os.Stderr, "  --crypto [BTC|ETH] Show crypto prices\n")
	fmt.Fprintf(os.Stderr, "  --share-log [file] Upload log to ix.io for sharing\n")
	fmt.Fprintf(os.Stderr, "  --make-icon <name> <symbol> Generate icon PNG\n")
	fmt.Fprintf(os.Stderr, "  --version          Show version\n")
	os.Exit(1)
}

// runInstall handles the --install command (runs as USER, no TTY/PTY needed)
// NOTE: Package installation is handled separately via --install-packages
func runInstall() {
	logger.Info("OpenRiot installer starting...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Fail(fmt.Sprintf("Could not determine home directory: %v", err))
		os.Exit(1)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fail(fmt.Sprintf("Failed to load config from %s: %v", configPath, err))
		os.Exit(1)
	}

	// Step 0: Package installation - SKIPPED (use --install-packages separately)
	// This avoids running pkg_add -u twice when called from setup.sh

	// Step 1: Config deployment
	if err := installer.CopyConfigs(repoDir, cfg, testMode); err != nil {
		logger.Warn(fmt.Sprintf("Config deployment skipped: %v", err))
	}

	// Step 2: Command execution
	logger.Info("Running post-install commands...")
	if err := installer.ExecCommands(cfg, testMode); err != nil {
		logger.Warn(fmt.Sprintf("Some commands failed: %v", err))
	}

	// Step 3: Source builds (crush, wlsunset, bibata-cursor, etc.)
	logger.Info("Running source builds...")
	if err := installer.SourceBuilds(cfg, testMode); err != nil {
		logger.Warn(fmt.Sprintf("Source builds: %v", err))
	}

	// Source builds handled above, setup.sh shows completion box
}

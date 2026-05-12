package commands

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
	"openriot/macspoof"
	"openriot/network"
	"openriot/nightlight"
	"openriot/notify"
	"openriot/polybar"
	"openriot/rofi"
	"openriot/roficalc"
	"openriot/screenshot"
	"openriot/screenrec"
	"openriot/settings"
	"openriot/update"
	"openriot/weather"
	"openriot/windowicon"
	"openriot/windowtitle"
	"openriot/wireguard"
	"openriot/workspace"
	"openriot/workspaceicons"
)

// RegisterAll populates the registry with all openriot commands.
func RegisterAll(r *Registry, testMode *bool) {
	// Installation
	r.Register(&Command{
		Name: "--install", Category: "Installation",
		Description: "Install OpenRiot",
		Run: func(args []string) error {
			if len(args) >= 1 {
				return installer.InstallTag(args[0])
			}
			runInstall(testMode)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--install-packages", Category: "Installation",
		Description: "Install packages from packages.yaml",
		Run: func(args []string) error { return installer.RunInstallPackages() },
	})
	r.Register(&Command{
		Name: "--show-release-notes", Category: "Installation",
		Description: "Display release notes for current version",
		Run: func(args []string) error { installer.ShowReleaseNotes(); return nil },
	})
	r.Register(&Command{
		Name: "--source-builds", Category: "Installation",
		Description: "Build software from source",
		Run: func(args []string) error { return installer.RunSourceBuilds(*testMode) },
	})
	r.Register(&Command{
		Name: "--packages", Category: "Installation",
		Description: "List packages from packages.yaml",
		Run: func(args []string) error {
			configPath := config.FindConfigFile()
			if configPath == "" {
				return fmt.Errorf("could not find packages.yaml")
			}
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			for _, pkg := range cfg.GetPackages() {
				fmt.Println(pkg)
			}
			return nil
		},
	})
	r.Register(&Command{
		Name: "--check-packages", Category: "Installation",
		Description: "Verify installed packages match yaml",
		Run: func(args []string) error { return installer.RunCheckPackages() },
	})
	r.Register(&Command{
		Name: "--sync-packages", Category: "Installation",
		Description: "Update packages.yaml to installed versions",
		Run: func(args []string) error { return installer.RunSyncPackages() },
	})
	r.Register(&Command{
		Name: "--check-dependencies", Category: "Installation",
		Description: "Show module dependency order",
		Run: func(args []string) error {
			configPath := config.FindConfigFile()
			if configPath == "" {
				return fmt.Errorf("could not find packages.yaml")
			}
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			refs, err := cfg.GetAllModulesOrdered()
			if err != nil {
				return fmt.Errorf("dependency error: %w", err)
			}
			fmt.Println("Dependency order:")
			for i, ref := range refs {
				deps := ""
				if len(ref.Module.Depends) > 0 {
					deps = fmt.Sprintf("  (depends on: %s)", strings.Join(ref.Module.Depends, ", "))
				}
				fmt.Printf("  %d. %s.%s%s\n", i+1, ref.Category, ref.Name, deps)
			}
			return nil
		},
	})
	r.Register(&Command{
		Name: "--validate-config", Category: "Installation",
		Description: "Validate packages.yaml",
		Run: func(args []string) error {
			configPath := config.FindConfigFile()
			if configPath == "" {
				return fmt.Errorf("could not find packages.yaml")
			}
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if err := config.ValidateConfig(cfg); err != nil {
				return fmt.Errorf("config validation failed: %w", err)
			}
			fmt.Println("[DONE] Config valid")
			return nil
		},
	})
	r.Register(&Command{
		Name: "--mirrors", Category: "Installation",
		Description: "Detect fastest OpenBSD mirror",
		Run: func(args []string) error { return installer.RunMirrors() },
	})

	// Tools & Upgrades
	r.Register(&Command{
		Name: "--random-mac", Category: "Tools & Upgrades",
		Description: "Manage MAC address randomization",
		Run: func(args []string) error { return macspoof.Run(args) },
	})
	r.Register(&Command{
		Name: "--crush-upgrade", Category: "Tools & Upgrades",
		Description: "Upgrade crush CLI to latest version",
		Run: func(args []string) error { return installer.NewCrushUpgrade().Run() },
	})
	r.Register(&Command{
		Name: "--gurk-setup", Category: "Tools & Upgrades",
		Description: "Configure gurk keybindings",
		Run: func(args []string) error { return gurk.Run() },
	})
	r.Register(&Command{
		Name: "--benchmark", Category: "Tools & Upgrades",
		Description: "Run system benchmark",
		Run: func(args []string) error {
			home, _ := os.UserHomeDir()
			cmd := home + "/.local/bin/benchmark; printf \"\\n\\nPress any key to continue...\"; read -r ans"
			exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
			return nil
		},
	})
	r.Register(&Command{
		Name: "--install-asset", Category: "Tools & Upgrades",
		Description: "Install Bibata or Kora assets",
		Run: func(args []string) error { return assets.Run(args) },
	})
	r.Register(&Command{
		Name: "--install-fonts", Category: "Tools & Upgrades",
		Description: "Install bundled Nerd Fonts",
		Run: func(args []string) error { return fonts.Run() },
	})
	r.Register(&Command{
		Name: "--install-rofi-calc", Category: "Tools & Upgrades",
		Description: "Build and install rofi-calc",
		Run: func(args []string) error { return roficalc.Run() },
	})

	// Version & Imaging
	r.Register(&Command{
		Name: "--version-check", Category: "Version & Imaging",
		Description: "Check if remote version is newer",
		Run: func(args []string) error {
			localVer := update.GetLocalVersion()
			remoteVer := update.GetRemoteVersion()
			if localVer == "unknown" || remoteVer == "unknown" {
				return fmt.Errorf("version unknown")
			}
			if update.CompareVersions(localVer, remoteVer) < 0 {
				fmt.Printf("Update available: %s -> %s\n", localVer, remoteVer)
				return nil
			}
			fmt.Printf("Current: %s\n", localVer)
			return fmt.Errorf("no update available")
		},
	})
	r.Register(&Command{
		Name: "--version", Category: "Version & Imaging",
		Description: "Show version",
		Run: func(args []string) error {
			// version is injected at build time; caller prints it
			fmt.Println("openriot", os.Getenv("OPENRIOT_VERSION"))
			return nil
		},
	})
	r.Register(&Command{
		Name: "--make-image", Category: "Version & Imaging",
		Description: "Build custom OpenBSD installer image",
		Run: func(args []string) error { imaging.RunMakeImage(args); return nil },
	})

	// System Controls
	r.Register(&Command{
		Name: "--volume", Category: "System Controls",
		Description: "Adjust volume",
		Run: func(args []string) error { os.Exit(audio.Run(args)); return nil },
	})
	r.Register(&Command{
		Name: "--brightness", Category: "System Controls",
		Description: "Adjust brightness",
		Run: func(args []string) error { os.Exit(display.Run(args)); return nil },
	})
	r.Register(&Command{
		Name: "--notify", Category: "System Controls",
		Description: "Send notification",
		Run: func(args []string) error { return runNotify(args) },
	})
	r.Register(&Command{
		Name: "--crypto", Category: "System Controls",
		Description: "Show crypto prices",
		Run: func(args []string) error {
			mode := "BTC"
			if len(args) >= 1 {
				mode = args[0]
			}
			return crypto.RunCrypto(mode)
		},
	})
	r.Register(&Command{
		Name: "--share-log", Category: "System Controls",
		Description: "Upload log to tmpfiles.org",
		Run: func(args []string) error {
			filename := "setup.log"
			if len(args) >= 1 {
				filename = args[0]
			}
			return installer.ShareLog(filename)
		},
	})
	r.Register(&Command{
		Name: "--make-icon", Category: "System Controls",
		Description: "Generate icon PNG",
		Run: func(args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: openriot --make-icon <name> <symbol>")
			}
			if err := installer.MakeIcon(args[0], args[1]); err != nil {
				return err
			}
			fmt.Printf("Icon created: %s.png\n", args[0])
			return nil
		},
	})

	// Polybar Status
	r.Register(&Command{
		Name: "--wireguard-status", Category: "Polybar Status",
		Description: "Show WireGuard icon",
		Run: func(args []string) error { fmt.Print(wireguard.Status()); return nil },
	})
	r.Register(&Command{
		Name: "--wireguard-notify", Category: "Polybar Status",
		Description: "Toggle or notify WireGuard",
		Run: func(args []string) error {
			if wireguard.Status() != "" {
				notify.SendNotify("wireguard", "WireGuard VPN", "WireGuard VPN is Enabled\nDisable in Settings Menu\nOr Super+Shift+G", "normal", 5000, 0)
			} else {
				wireguard.Toggle()
			}
			return nil
		},
	})
	r.Register(&Command{
		Name: "--wireguard-autostart", Category: "System",
		Description: "Auto-start WireGuard if enabled",
		Run: func(args []string) error { return wireguard.AutoStart() },
	})
	r.Register(&Command{
		Name: "--stealth-status", Category: "Polybar Status",
		Description: "Show stealth mode icon",
		Run: func(args []string) error { fmt.Print(macspoof.StealthStatus()); return nil },
	})
	r.Register(&Command{
		Name: "--stealth", Category: "Polybar Status",
		Description: "Toggle stealth mode",
		Run: func(args []string) error { return runStealthToggle() },
	})
	r.Register(&Command{
		Name: "--stealth-notify", Category: "Polybar Status",
		Description: "Toggle or notify stealth mode",
		Run: func(args []string) error { return runStealthNotify() },
	})
	r.Register(&Command{
		Name: "--update-status", Category: "Polybar Status",
		Description: "Show update icon",
		Run: func(args []string) error { fmt.Print(update.Get()); return nil },
	})
	r.Register(&Command{
		Name: "--update", Category: "Polybar Status",
		Description: "Check for updates",
		Run: func(args []string) error { update.Click(); return nil },
	})
	r.Register(&Command{
		Name: "--rofi", Category: "Polybar Status",
		Description: "Show app launcher",
		Run: func(args []string) error { return rofi.Run() },
	})
	r.Register(&Command{
		Name: "--settings-menu", Category: "Polybar Status",
		Description: "Show settings menu",
		Run: func(args []string) error { settings.RunMenu(); return nil },
	})
	r.Register(&Command{
		Name: "--games-menu", Category: "Polybar Status",
		Description: "Show games menu",
		Run: func(args []string) error { return rofi.RunGames() },
	})
	r.Register(&Command{
		Name: "--weather", Category: "Polybar Status",
		Description: "Show weather",
		Run: func(args []string) error { fmt.Print(weather.Get()); return nil },
	})

	// Network & Battery
	r.Register(&Command{
		Name: "--network-wifi", Category: "Network & Battery",
		Description: "Show WiFi status",
		Run: func(args []string) error { fmt.Print(network.GetWifi()); return nil },
	})
	r.Register(&Command{
		Name: "--network-eth", Category: "Network & Battery",
		Description: "Show ethernet status",
		Run: func(args []string) error { fmt.Print(network.GetEth()); return nil },
	})
	r.Register(&Command{
		Name: "--wifi-info", Category: "Network & Battery",
		Description: "Show WiFi details notification",
		Run: func(args []string) error {
			details := network.GetWifiDetails()
			icon := "wifi.png"
			if !network.IsConnected() {
				icon = "wifi-off.png"
			}
			notify.SendNotify(icon, "WiFi", details, "normal", 5000, 0)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--wifi-reconnect", Category: "Network & Battery",
		Description: "Reconnect WiFi",
		Run: func(args []string) error {
			if network.IsOnline() {
				notify.SendNotify("wifi", "WiFi", "Already connected", "normal", 2000, 0)
				return nil
			}
			if !network.IsConnected() {
				notify.SendNotify("wifi-off", "WiFi", "Not connected", "normal", 2000, 0)
				return nil
			}
			notify.SendNotify("wifi", "WiFi", "Reconnecting...", "normal", 3000, 0)
			if err := network.ReconnectWifi(); err != nil {
				notify.SendNotify("wifi-off", "WiFi", "Reconnect failed: "+err.Error(), "normal", 5000, 0)
				return err
			}
			return nil
		},
	})
	r.Register(&Command{
		Name: "--eth-info", Category: "Network & Battery",
		Description: "Show ethernet details notification",
		Run: func(args []string) error {
			notify.SendNotify("ethernet", "Ethernet", network.GetEthDetails(), "normal", 5000, 0)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--battery", Category: "Network & Battery",
		Description: "Show battery status",
		Run: func(args []string) error { fmt.Print(battery.Get()); return nil },
	})
	r.Register(&Command{
		Name: "--battery-notify", Category: "Network & Battery",
		Description: "Show battery notification",
		Run: func(args []string) error {
			notify.SendNotify("battery", "Battery", battery.GetNotifyDetails(), "normal", 5000, 0)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--night-light-status", Category: "Network & Battery",
		Description: "Show night light status",
		Run: func(args []string) error { fmt.Print(nightlight.Get()); return nil },
	})
	r.Register(&Command{
		Name: "--screenrec-status", Category: "Network & Battery",
		Description: "Show screen recorder status",
		Run: func(args []string) error { fmt.Print(screenrec.Status()); return nil },
	})
	r.Register(&Command{
		Name: "--screenrec-toggle", Category: "Network & Battery",
		Description: "Toggle screen recording",
		Run: func(args []string) error { return screenrec.Toggle() },
	})
	r.Register(&Command{
		Name: "--hdmi", Category: "Network & Battery",
		Description: "Show HDMI/external display icon",
		Run: func(args []string) error { display.RunHDMI(); return nil },
	})
	r.Register(&Command{
		Name: "--hdmi-toggle", Category: "Network & Battery",
		Description: "Toggle HDMI-only / Laptop+HDMI mode",
		Run: func(args []string) error { display.ToggleHDMI(); return nil },
	})
	r.Register(&Command{
		Name: "--polybar-transmission", Category: "Network & Battery",
		Description: "Show transmission icon",
		Run: func(args []string) error {
			if rofi.IsTransmissionRunning() {
				fmt.Print(polybar.Icon("󰐻"))
			}
			return nil
		},
	})
	r.Register(&Command{
		Name: "--polybar-proton-drive", Category: "Network & Battery",
		Description: "Show Proton Drive sync status",
		Run: func(args []string) error { return polybar.RunProtonDrive() },
	})

	// Drive & Sync
	r.Register(&Command{
		Name: "--proton-drive-sync", Category: "Drive & Sync",
		Description: "Sync Proton Drive",
		Run: func(args []string) error { return runProtonDriveSync() },
	})
	r.Register(&Command{
		Name: "--proton-drive-init", Category: "Drive & Sync",
		Description: "Initialize Proton Drive cache",
		Run: func(args []string) error { return runProtonDriveInit() },
	})
	r.Register(&Command{
		Name: "--transmission-toggle", Category: "Drive & Sync",
		Description: "Toggle Transmission",
		Run: func(args []string) error { return runTransmissionToggle() },
	})
	r.Register(&Command{
		Name: "--transmission-notify", Category: "Drive & Sync",
		Description: "Toggle or notify Transmission",
		Run: func(args []string) error { return runTransmissionNotify() },
	})
	r.Register(&Command{
		Name: "--night-light", Category: "Drive & Sync",
		Description: "Toggle night light",
		Run: func(args []string) error { nightlight.Toggle(); return nil },
	})
	r.Register(&Command{
		Name: "--night-light-notify", Category: "Drive & Sync",
		Description: "Toggle or notify night light",
		Run: func(args []string) error { return runNightLightNotify() },
	})
	r.Register(&Command{
		Name: "--window-title", Category: "Drive & Sync",
		Description: "Show focused window title",
		Run: func(args []string) error { fmt.Print(windowtitle.Get()); return nil },
	})
	r.Register(&Command{
		Name: "--wireguard", Category: "Drive & Sync",
		Description: "Toggle WireGuard VPN",
		Run: func(args []string) error { return wireguard.Toggle() },
	})

	// Window & Workspace
	r.Register(&Command{
		Name: "--window-icon", Category: "Window & Workspace",
		Description: "Get icon for window class",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: openriot --window-icon <class>")
			}
			fmt.Print(windowicon.Get(args[0]))
			return nil
		},
	})
	r.Register(&Command{
		Name: "--all-window-icons", Category: "Window & Workspace",
		Description: "List all window icons",
		Run: func(args []string) error {
			for class, icon := range windowicon.GetAllWindowIcons() {
				fmt.Printf("%s=%s\n", class, icon)
			}
			return nil
		},
	})
	r.Register(&Command{
		Name: "--workspace-switch", Category: "Window & Workspace",
		Description: "Switch to workspace N",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: openriot --workspace-switch <N>")
			}
			target, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid workspace number")
			}
			workspace.Switch(target)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--workspace-move", Category: "Window & Workspace",
		Description: "Move window to workspace N",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: openriot --workspace-move <N>")
			}
			target, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid workspace number")
			}
			workspace.Move(target)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--workspace-icons", Category: "Window & Workspace",
		Description: "Show icons for workspace N",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: openriot --workspace-icons <N>")
			}
			target, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid workspace number")
			}
			fmt.Print(workspaceicons.Get(target))
			return nil
		},
	})
	r.Register(&Command{
		Name: "--workspace-icons-all", Category: "Window & Workspace",
		Description: "Show icons for all workspaces",
		Run: func(args []string) error { fmt.Print(workspaceicons.GetAll()); return nil },
	})

	// Lock & Power
	r.Register(&Command{
		Name: "--lock", Category: "Lock & Power",
		Description: "Lock the screen",
		Run: func(args []string) error { lock.Lock(); return nil },
	})
	r.Register(&Command{
		Name: "--smart-lock", Category: "Lock & Power",
		Description: "Lock only if no audio playing",
		Run: func(args []string) error { lock.SmartLock(); return nil },
	})
	r.Register(&Command{
		Name: "--build-lock-cache", Category: "Lock & Power",
		Description: "Pre-convert lock screen images to screen resolution",
		Run: func(args []string) error { return lock.BuildCache() },
	})
	r.Register(&Command{
		Name: "--signal-launch", Category: "Lock & Power",
		Description: "Launch Signal messenger",
		Run: func(args []string) error { return runAppLauncher("signal", "gurk", "alacritty", "--class", "gurk", "--title", "Signal", "-e") },
	})
	r.Register(&Command{
		Name: "--browser", Category: "Lock & Power",
		Description: "Launch Firefox",
		Run: func(args []string) error {
			cmd := exec.Command("pgrep", "-f", "firefox")
			output, _ := cmd.Output()
			if len(strings.TrimSpace(string(output))) > 0 {
				notify.SendNotify("firefox", "Firefox", "Already running", "normal", 2000, 0)
				return nil
			}
			notify.SendNotify("firefox", "Firefox", "Starting Browser...", "normal", 2000, 0)
			exec.Command("firefox", args...).Start()
			return nil
		},
	})
	r.Register(&Command{
		Name: "--proton", Category: "Lock & Power",
		Description: "Open Proton Mail",
		Run: func(args []string) error {
			notify.SendNotify("proton-mail", "Proton Mail", "Opening...", "normal", 2000, 0)
			exec.Command("firefox", "https://mail.proton.me/u/11/inbox").Start()
			return nil
		},
	})
	r.Register(&Command{
		Name: "--twitter", Category: "Lock & Power",
		Description: "Open X (Twitter)",
		Run: func(args []string) error {
			notify.SendNotify("twitter", "X (Twitter)", "Opening...", "normal", 2000, 0)
			exec.Command("firefox", "https://x.com/").Start()
			return nil
		},
	})
	r.Register(&Command{
		Name: "--help-launch", Category: "Lock & Power",
		Description: "Open OpenRiot Help website",
		Run: func(args []string) error {
			notify.SendNotify("help", "Help", "Launching help...", "normal", 2000, 0)
			exec.Command("firefox", "https://openriot.org").Start()
			return nil
		},
	})
	r.Register(&Command{
		Name: "--crush", Category: "Lock & Power",
		Description: "Launch Crush AI",
		Run: func(args []string) error { return runAppLauncher("crush", "crush", "alacritty", "--class", "crush", "--title", "Crush AI", "-e") },
	})
	r.Register(&Command{
		Name: "--suspend", Category: "Lock & Power",
		Description: "Suspend the system",
		Run: func(args []string) error { exec.Command("zzz").Run(); return nil },
	})
	r.Register(&Command{
		Name: "--power-menu", Category: "Lock & Power",
		Description: "Show power menu",
		Run: func(args []string) error { return runPowerMenu() },
	})
	r.Register(&Command{
		Name: "--wallpaper-next", Category: "Lock & Power",
		Description: "Next wallpaper",
		Run: func(args []string) error { os.Exit(backgrounds.Next()); return nil },
	})
	r.Register(&Command{
		Name: "--wallpaper-prev", Category: "Lock & Power",
		Description: "Previous wallpaper",
		Run: func(args []string) error { os.Exit(backgrounds.Prev()); return nil },
	})
	r.Register(&Command{
		Name: "--wallpaper-load", Category: "Lock & Power",
		Description: "Load saved wallpaper",
		Run: func(args []string) error { os.Exit(backgrounds.Load()); return nil },
	})
	r.Register(&Command{
		Name: "--polybar-setup", Category: "Lock & Power",
		Description: "Scale and setup polybar",
		Run: func(args []string) error { os.Exit(polybar.Setup()); return nil },
	})
	r.Register(&Command{
		Name: "--dunst-setup", Category: "Lock & Power",
		Description: "Scale and setup dunst",
		Run: func(args []string) error { os.Exit(notify.Setup()); return nil },
	})
	r.Register(&Command{
		Name: "--suspend-if-undocked", Category: "Lock & Power",
		Description: "Auto-suspend when undocked",
		Run: func(args []string) error { detect.SuspendIfUndocked(); return nil },
	})
	r.Register(&Command{
		Name: "--screenshot", Category: "Lock & Power",
		Description: "Take screenshot",
		Run: func(args []string) error {
			selectArea := len(args) >= 1 && args[0] == "select"
			return screenshot.Run(selectArea)
		},
	})

	// Polybar Metrics
	r.Register(&Command{
		Name: "--notify-dismiss", Category: "Polybar Metrics",
		Description: "Dismiss notification",
		Run: func(args []string) error {
			id := 0
			if len(args) >= 1 {
				fmt.Sscanf(args[0], "%d", &id)
			}
			return notify.Dismiss(id)
		},
	})
	r.Register(&Command{
		Name: "--notify-clear", Category: "Polybar Metrics",
		Description: "Clear all notifications",
		Run: func(args []string) error { return notify.Clear() },
	})
	r.Register(&Command{
		Name: "--notify-status", Category: "Polybar Metrics",
		Description: "Show notification status",
		Run: func(args []string) error { return notify.Status() },
	})
	r.Register(&Command{
		Name: "--polybar-metrics", Category: "Polybar Metrics",
		Description: "Show CPU/RAM for polybar",
		Run: func(args []string) error { return polybar.RunMetrics() },
	})
	r.Register(&Command{
		Name: "--polybar-volume", Category: "Polybar Metrics",
		Description: "Show volume for polybar",
		Run: func(args []string) error { return polybar.RunVolume() },
	})
	r.Register(&Command{
		Name: "--polybar-memory", Category: "Polybar Metrics",
		Description: "Show memory for polybar",
		Run: func(args []string) error {
			ram := polybar.GetRAM()
			ramPct := polybar.GetMemPercent()
			fmt.Printf(" %s\nMemory: %s\n", ram, ramPct)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--polybar-all", Category: "Polybar Metrics",
		Description: "Show all polybar metrics in one call",
		Run: func(args []string) error {
			return polybar.RunAll()
		},
	})
	r.Register(&Command{
		Name: "--cpu-notify", Category: "Polybar Metrics",
		Description: "Show CPU notification",
		Run: func(args []string) error {
			notify.SendNotify("cpu", "CPU", polybar.GetCPUDetails(), "normal", 5000, 0)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--mem-notify", Category: "Polybar Metrics",
		Description: "Show memory notification",
		Run: func(args []string) error {
			notify.SendNotify("memory", "Memory", polybar.GetMemDetails(), "normal", 5000, 0)
			return nil
		},
	})
	r.Register(&Command{
		Name: "--crypto-notify", Category: "Polybar Metrics",
		Description: "Show crypto notification",
		Run: func(args []string) error {
			notify.SendNotify("chart", "Crypto", "Loading...", "normal", 0, 1)
			time.Sleep(1 * time.Second)
			return crypto.RunCrypto("NOTIFY_SEND")
		},
	})
	r.Register(&Command{
		Name: "--crypto-icon", Category: "Polybar Metrics",
		Description: "Show crypto polybar icon if configured",
		Run: func(args []string) error {
			if crypto.ConfigFileExists() {
				fmt.Print("%{T1}󰄨%{T-}%{O2}")
			}
			return nil
		},
	})
	r.Register(&Command{
		Name: "--crypto-refresh", Category: "Polybar Metrics",
		Description: "Clear crypto cache and fetch fresh",
		Run: func(args []string) error {
			os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "openriot", "crypto.json"))
			os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "openriot", "crypto-prev.json"))
			return crypto.RunCrypto("ROWML")
		},
	})
}

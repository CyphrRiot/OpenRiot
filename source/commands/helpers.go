package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"openriot/config"
	"openriot/installer"
	"openriot/lock"
	"openriot/logger"
	"openriot/macspoof"
	"openriot/nightlight"
	"openriot/notify"
	"openriot/paths"
	"openriot/polybar"
	"openriot/rofi"
	"openriot/wireguard"
)

func runInstall(testMode *bool) {
	logger.Info("OpenRiot installer starting...")

	// Pre-install release path check for -current users
	if config.DetectOpenBSDVersion() == "snapshots" {
		res, err := installer.CheckReleasePath(installer.ReleaseDate)
		if err == installer.ErrUpgradeRequired {
			if installer.PrintReleasePathBanner(res, "7.9") {
				fmt.Println()
				logger.Info("Running: doas sysupgrade -R 7.9")
				fmt.Println()
				fmt.Println("The system will reboot into the upgrade kernel.")
				fmt.Println("After boot completes, re-run the installer:")
				fmt.Println("  curl -fsSL https://OpenRiot.org/setup.sh | sh")
				fmt.Println()
				cmd := exec.Command("doas", "sysupgrade", "-R", "7.9")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					logger.Warn(fmt.Sprintf("sysupgrade failed: %v", err))
					fmt.Println("Please run manually: doas sysupgrade -R 7.9")
				}
				os.Exit(0)
			}
			logger.Info("Continuing with -current installation.")
		} else if err == installer.ErrDowngradeRisk {
			logger.Info("Running on post-release snapshot. sysupgrade -R would downgrade.")
			logger.Info("Staying on -current branch.")
		}
	}

	homeDir := paths.HomeDir()
	if homeDir == "" {
		logger.Fail("Could not determine home directory")
		os.Exit(1)
	}

	deployDir := paths.OpenRiotDir()
	configPath := filepath.Join(deployDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fail(fmt.Sprintf("Failed to load config from %s: %v", configPath, err))
		os.Exit(1)
	}

	installGames := installer.GamesPreference()
	if !installGames {
		delete(cfg.Desktop, "games")
		delete(cfg.Desktop, "pacman")
	}

	if err := installer.CopyConfigs(deployDir, cfg, *testMode); err != nil {
		logger.Warn(fmt.Sprintf("Config deployment skipped: %v", err))
	}

	if !installGames {
		installer.StripGamesFromRofi()
	}

	installKate := installer.KatePreference()
	if !installKate {
		delete(cfg.Desktop, "kate")
		installer.StripKateFromRofi()
	} else {
		if err := installer.SetupKateConfig(); err != nil {
			logger.Warn(fmt.Sprintf("Kate config setup: %v", err))
		}
	}

	logger.Info("Running post-install commands...")
	if err := installer.ExecCommands(cfg, *testMode); err != nil {
		logger.Warn(fmt.Sprintf("Some commands failed: %v", err))
	}

	// Render dynamic configs from templates
	openriotBin := filepath.Join(deployDir, "install", "openriot")
	if _, err := os.Stat(openriotBin); err == nil {
		logger.Info("Rendering dynamic configs...")
		for _, flag := range []string{
			"--polybar-setup", "--dunst-setup", "--rofi-setup",
			"--helix-setup",
		} {
			out, err := exec.Command(openriotBin, flag).CombinedOutput()
			if err != nil {
				logger.Warn(fmt.Sprintf("%s failed: %v", flag, err))
			} else if len(out) > 0 {
				logger.Info(strings.TrimSpace(string(out)))
			}
		}
	}

	logger.Info("Running source builds...")
	if err := installer.SourceBuilds(cfg, *testMode); err != nil {
		logger.Warn(fmt.Sprintf("Source builds: %v", err))
	}

	srcDir := filepath.Join(deployDir, "source")
	// Only delete source/ in production installs, not local dev mode
	if os.Getenv("OPENRIOT_LOCAL") != "1" {
		if err := os.RemoveAll(srcDir); err == nil {
			logger.Info("Source files cleaned up")
		}
	} else {
		logger.Info("Local mode: skipping source cleanup")
	}

	if installer.AskShowReleaseNotes() {
		installer.ShowReleaseNotes()
	}

	// Remind -current user to keep base and packages in sync (only if drift detected)
	if config.DetectOpenBSDVersion() == "snapshots" {
		drift, buildDate := hasPackageDrift()
		if drift {
			fmt.Println()
			logger.Warn("You are running OpenBSD -current.")
			logger.Info(fmt.Sprintf("Kernel built: %s", buildDate.Format("Jan 2 2006")))
			logger.Info("To keep base and packages in sync, run:")
			fmt.Println("  doas sysupgrade -s")
			fmt.Println("  (reboot)")
			fmt.Println("  doas pkg_add -u")
		}
	}
}

// hasPackageDrift parses kernel build date from sysctl and returns true if
// the kernel is more than 7 days old (indicating potential package/base drift).
func hasPackageDrift() (bool, time.Time) {
	cmd := exec.Command("sysctl", "-n", "kern.version")
	output, err := cmd.Output()
	if err != nil {
		return false, time.Time{}
	}

	// Parse date from "OpenBSD 7.9-current (...) #475: Thu May 14 12:34:48 MDT 2026"
	line := strings.TrimSpace(string(output))
	idx := strings.Index(line, ": ")
	if idx < 0 {
		return false, time.Time{}
	}
	dateStr := line[idx+2:]
	buildDate, err := time.Parse("Mon Jan 2 15:04:05 MST 2006", dateStr)
	if err != nil {
		return false, time.Time{}
	}

	return time.Since(buildDate) > 7*24*time.Hour, buildDate
}

// checkReleasePath determines if a -current system can safely migrate to the
// 7.9 release. Pre-release snapshots (built before the release date) can use
// sysupgrade -R. Post-release snapshots cannot; fresh install is required.
func checkReleasePath() {
	res, err := installer.CheckReleasePath(installer.ReleaseDate)

	if err != nil && err != installer.ErrUpgradeRequired && err != installer.ErrDowngradeRisk {
		fmt.Printf("Cannot determine system status: %v\n", err)
		return
	}

	switch res.Status {
	case "stable":
		fmt.Println("You are already on a stable/release build.")
		fmt.Println("Run: doas sysupgrade")
		fmt.Println("Then: doas pkg_add -u")
	case "pre-release":
		fmt.Printf("Kernel build date: %s\n", res.BuildDate.Format("2006-01-02"))
		fmt.Printf("7.9 release date:  %s\n", res.ReleaseDate.Format("2006-01-02"))
		fmt.Println()
		fmt.Println("Status: PRE-RELEASE snapshot")
		fmt.Println("You can safely upgrade to the 7.9 release.")
		fmt.Println()
		fmt.Println("Run these commands:")
		fmt.Println("  doas sysupgrade -R 7.9")
		fmt.Println("  (reboot when prompted)")
		fmt.Println("  doas pkg_add -u")
	case "post-release":
		fmt.Printf("Kernel build date: %s\n", res.BuildDate.Format("2006-01-02"))
		fmt.Printf("7.9 release date:  %s\n", res.ReleaseDate.Format("2006-01-02"))
		fmt.Println()
		fmt.Println("Status: POST-RELEASE snapshot")
		fmt.Println("WARNING: sysupgrade -R 7.9 would be a DOWNGRADE.")
		fmt.Println("OpenBSD does not support downgrades. Do NOT run it.")
		fmt.Println()
		fmt.Println("Your options:")
		fmt.Println("  1. Stay on -current: doas sysupgrade (no -R)")
		fmt.Println("  2. Fresh install 7.9 release from install79.img")
	}
}

var (
	bindIPv4Re = regexp.MustCompile(`"bind-address-ipv4"\s*:\s*"[^"]*"`)
	bindIPv6Re = regexp.MustCompile(`"bind-address-ipv6"\s*:\s*"[^"]*"`)
)

func runNotify(args []string) error {
	title, body, urgency, iconPath := "", "", "normal", ""
	expiresIn := 0
	for i := 0; i < len(args); i++ {
		if args[i] == "--urgency" && i+1 < len(args) {
			urgency = args[i+1]
			i++
		} else if args[i] == "--expires-in" && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil {
				expiresIn = n
			}
			i++
		} else if args[i] == "--icon" && i+1 < len(args) {
			iconPath = args[i+1]
			i++
		} else if title == "" {
			title = args[i]
		} else if body == "" {
			body = args[i]
		}
	}
	if title == "" {
		return fmt.Errorf("usage: openriot --notify \"title\" \"body\" [--urgency normal] [--expires-in seconds] [--icon path]")
	}
	cmdArgs := []string{"/usr/local/bin/notify-send"}
	if iconPath != "" {
		cmdArgs = append(cmdArgs, "-i", iconPath)
	}
	if urgency != "normal" {
		cmdArgs = append(cmdArgs, "-u", urgency)
	}
	if expiresIn > 0 {
		cmdArgs = append(cmdArgs, "-t", fmt.Sprintf("%d", expiresIn*1000))
	}
	cmdArgs = append(cmdArgs, title)
	if body != "" {
		cmdArgs = append(cmdArgs, body)
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Run()
	var expiresAt int64
	if expiresIn > 0 {
		expiresAt = time.Now().Unix() + int64(expiresIn)
	}
	notify.Add(title, body, urgency, expiresAt)
	return nil
}

func runStealthToggle() error {
	notify.SendNotify("stealth", "Stealth", " Restarting Networking Services", "normal", 5000, 0)
	if err := macspoof.StealthToggle(); err != nil {
		notify.SendNotify("stealth", "Stealth", "Failed: "+err.Error(), "critical", 5000, 0)
		return err
	}
	enabled := macspoof.IsStealthEnabled()
	if enabled {
		notify.SendNotify("stealth", "Stealth", "Enabled [Stealth]", "normal", 3000, 0)
	} else {
		notify.SendNotify("stealth", "Stealth", "Disabled", "normal", 3000, 0)
	}
	return nil
}

func runStealthNotify() error {
	if macspoof.IsStealthEnabled() {
		notify.SendNotify("stealth", "Stealth Mode", "Stealth Mode is Enabled\nDisable in Settings Menu\nOr Super+Shift+G", "normal", 5000, 0)
		return nil
	}
	return runStealthToggle()
}

func runProtonDriveSync() error {
	if polybar.IsProtonDriveConfigured() {
		state := polybar.CheckProtonDriveSyncState()
		if state == "synced" {
			notify.SendNotify("proton-drive", "Proton Drive", "Synchronized: "+polybar.GetProtonDriveTooltipText(), "normal", 5000, 0)
			return nil
		}
		notify.SendNotify("proton-drive", "Proton Drive", "Syncing...", "normal", 2000, 0)
		cmd := `printf "Proton Drive Sync\nFrom: ~/Documents/ProtonSync -> Proton Drive Cloud\n\nWould you like to do a bi-directional Sync or one-way\n  and replace items in the Cloud with local items?\n\n[Y]es for bi-directional sync (or ENTER),\n[O]ne-way for One-Way sync or\n[Q]uit or [N]o ?\n\nChoose your adventure [Y/o/q/n] -> "; read -r ans; case "$ans" in o|O) echo "One-way sync selected..."; rclone copy ~/Documents/ProtonSync proton:ProtonSync --progress; printf "\nDone. Press Enter to close..."; read -r ans ;; [yY]|"") echo "Bi-directional sync selected..."; rclone bisync ~/Documents/ProtonSync proton:ProtonSync --resync --progress; printf "\nDone. Press Enter to close..."; read -r ans ;; *) echo "Canceled."; sleep 1 ;; esac`
		exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
	} else {
		notify.SendNotify("proton-drive", "Proton Drive", "Not configured\nSee OpenRiot.org for setup info", "critical", 5000, 0)
	}
	return nil
}

func runProtonDriveInit() error {
	if polybar.IsProtonDriveConfigured() {
		if err := polybar.InitProtonDriveCache(); err != nil {
			notify.SendNotify("proton-drive", "Proton Drive", "Failed to init cache", "critical", 5000, 0)
			return err
		}
		notify.SendNotify("proton-drive", "Proton Drive", "Cache initialized", "normal", 2000, 0)
	} else {
		notify.SendNotify("proton-drive", "Proton Drive", "Not configured", "critical", 5000, 0)
	}
	return nil
}

func startTransmission() error {
	if !wireguard.IsRunning() {
		notify.SendNotify("transmission", "Transmission", "Wireguard is NOT running.\nCannot start Transmission without Wireguard.\n(This is a protective measure)", "critical", 5000, 0)
		return nil
	}
	bindTransmissionToWireGuard()
	exec.Command("transmission-gtk").Start()
	notify.SendNotify("transmission", "Transmission", "Starting Transmission...", "normal", 2000, 0)
	return nil
}

func runTransmissionToggle() error {
	if rofi.IsTransmissionRunning() {
		exec.Command("pkill", "transmission-gtk").Run()
		notify.SendNotify("transmission", "Transmission", "Stopping Transmission...", "normal", 2000, 0)
	} else {
		return startTransmission()
	}
	return nil
}

func runTransmissionNotify() error {
	if rofi.IsTransmissionRunning() {
		notify.SendNotify("transmission", "Transmission", "Transmission is Enabled\nDisable in Rofi Menu\nOr Super+Shift+G", "normal", 5000, 0)
	} else {
		return startTransmission()
	}
	return nil
}

func bindTransmissionToWireGuard() {
	// Get the active WireGuard tunnel IP and bind transmission to it
	// so the peer port only listens on the VPN interface, not LAN.
	tunnelIP := wireguard.GetTunnelIP()
	if tunnelIP == "" {
		return
	}
	settingsPath := paths.Join(".config", "transmission", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}
	content := string(data)
	// Replace bind-address-ipv4
	if strings.Contains(content, `"bind-address-ipv4"`) {
		content = bindIPv4Re.ReplaceAllString(content, `"bind-address-ipv4": "`+tunnelIP+`"`)
	} else {
		// Insert after the first property
		content = strings.Replace(content, `"alt-speed-down"`, `"bind-address-ipv4": "`+tunnelIP+`",\n    "alt-speed-down"`, 1)
	}
	// Replace bind-address-ipv6 (empty — let v4 handle it for now)
	if strings.Contains(content, `"bind-address-ipv6"`) {
		content = bindIPv6Re.ReplaceAllString(content, `"bind-address-ipv6": ""`)
	} else {
		content = strings.Replace(content, `"bind-address-ipv4": "`+tunnelIP+`"`, `"bind-address-ipv4": "`+tunnelIP+`",\n    "bind-address-ipv6": ""`, 1)
	}
	if err := os.WriteFile(settingsPath, []byte(content), 0644); err != nil {
		logger.Warn(fmt.Sprintf("failed to write transmission settings: %v", err))
	}
}

func runNightLightNotify() error {
	nightlight.Toggle()
	if nightlight.IsOn() {
		notify.SendNotify("nightlight-on", "Night Light", "Night Light is On", "normal", 3000, 0)
	} else {
		notify.SendNotify("nightlight-off", "Night Light", "Night Light is Off", "normal", 3000, 0)
	}
	return nil
}

func runPowerMenu() error {
	theme := paths.ConfigDir("rofi", "simple-tokyonight.rasi")
	if _, err := os.Stat(theme); os.IsNotExist(err) {
		theme = "simple-tokyonight"
	}
	cmd := exec.Command("rofi", "-dmenu", "-p", "Power: ", "-theme", theme)
	cmd.Stdin = strings.NewReader("󰌾 Lock\n󰒲 Suspend\n󰑐 Reboot\n󰐥 Shutdown\n󰍃 Logout")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	choice := strings.TrimSpace(string(out))
	switch {
	case strings.HasPrefix(choice, "󰌾"):
		lock.Lock()
	case strings.HasPrefix(choice, "󰒲"):
		notify.SendNotify("power", "Power", "Suspending...", "normal", 2000, 0)
		lock.Lock()
		exec.Command("doas", "zzz").Run()
	case strings.HasPrefix(choice, "󰑐"):
		notify.SendNotify("power", "Power", "Rebooting...", "normal", 3000, 0)
		exec.Command("doas", "shutdown", "-r", "now").Run()
	case strings.HasPrefix(choice, "󰐥"):
		notify.SendNotify("power", "Power", "Shutting down...", "normal", 5000, 0)
		exec.Command("doas", "shutdown", "-p", "now").Run()
	case strings.HasPrefix(choice, "󰍃"):
		notify.SendNotify("power", "Power", "Logging out...", "normal", 2000, 0)
		exec.Command("i3-msg", "exit").Run()
	}
	return nil
}

func runAppLauncher(icon, procName string, cmdArgs ...string) error {
	cmd := exec.Command("pgrep", "-f", procName)
	output, _ := cmd.Output()
	if len(strings.TrimSpace(string(output))) > 0 {
		notify.SendNotify(icon, procName, procName+" already launched", "normal", 2000, 0)
		return nil
	}
	home := paths.HomeDir()
	if home == "" {
		return fmt.Errorf("cannot get home dir")
	}
	binPath := paths.OpenRiotDir("config", "bin", procName)
	notify.SendNotify(icon, procName, "Starting "+procName+"...", "normal", 2000, 0)
	exec.Command(cmdArgs[0], append(cmdArgs[1:], binPath)...).Start()
	return nil
}

// fixNzbgetPerms ensures /etc/nzbget.conf and /var/nzbget are owned by the
// current user so nzbget can read/write without permission errors.
// Silently no-op if the paths do not exist.
func fixNzbgetPerms() {
	uid := os.Getuid()
	user := os.Getenv("USER")
	if user == "" {
		out, _ := exec.Command("id", "-un").Output()
		user = strings.TrimSpace(string(out))
	}
	if user == "" {
		return
	}
	for _, path := range []string{"/etc/nzbget.conf", "/var/nzbget"} {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			continue
		}
		if stat.Uid == uint32(uid) {
			continue
		}
		if err := exec.Command("doas", "chown", "-f", user+":"+user, path).Run(); err != nil {
			logger.Warn(fmt.Sprintf("nzbget perms: failed to chown %s: %v", path, err))
		}
	}
}
package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"openriot/config"
	"openriot/installer"
	"openriot/lock"
	"openriot/logger"
	"openriot/macspoof"
	"openriot/nightlight"
	"openriot/notify"
	"openriot/polybar"
	"openriot/rofi"
)

func runInstall(testMode *bool) {
	logger.Info("OpenRiot installer starting...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Fail(fmt.Sprintf("Could not determine home directory: %v", err))
		os.Exit(1)
	}

	deployDir := filepath.Join(homeDir, ".local", "share", "openriot")
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

	logger.Info("Running post-install commands...")
	if err := installer.ExecCommands(cfg, *testMode); err != nil {
		logger.Warn(fmt.Sprintf("Some commands failed: %v", err))
	}

	logger.Info("Running source builds...")
	if err := installer.SourceBuilds(cfg, *testMode); err != nil {
		logger.Warn(fmt.Sprintf("Source builds: %v", err))
	}

	srcDir := filepath.Join(deployDir, "source")
	if err := os.RemoveAll(srcDir); err == nil {
		logger.Info("Source files cleaned up")
	}

	if installer.AskShowReleaseNotes() {
		installer.ShowReleaseNotes()
	}
}

func runNotify(args []string) error {
	title, body, urgency, iconPath := "", "", "normal", ""
	expiresIn := 0
	for i := 0; i < len(args); i++ {
		if args[i] == "--urgency" && i+1 < len(args) {
			urgency = args[i+1]
			i++
		} else if args[i] == "--expires-in" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &expiresIn)
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
	} else {
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
	}
	return nil
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

func runTransmissionToggle() error {
	if rofi.IsTransmissionRunning() {
		exec.Command("pkill", "transmission-gtk").Run()
		notify.SendNotify("transmission", "Transmission", "Stopping Transmission...", "normal", 2000, 0)
	} else {
		exec.Command("transmission-gtk").Start()
		notify.SendNotify("transmission", "Transmission", "Starting Transmission...", "normal", 2000, 0)
	}
	return nil
}

func runTransmissionNotify() error {
	if rofi.IsTransmissionRunning() {
		notify.SendNotify("transmission", "Transmission", "Transmission is Enabled\nDisable in Rofi Menu\nOr Super+Shift+G", "normal", 5000, 0)
	} else {
		exec.Command("transmission-gtk").Start()
		notify.SendNotify("transmission", "Transmission", "Starting Transmission...", "normal", 2000, 0)
	}
	return nil
}

func runNightLightNotify() error {
	if nightlight.Get() != "" {
		notify.SendNotify("nightlight", "Night Light", "Night Light is Enabled\nDisable in Settings Menu\nOr Super+Shift+G", "normal", 5000, 0)
	} else {
		nightlight.Toggle()
	}
	return nil
}

func runPowerMenu() error {
	home, _ := os.UserHomeDir()
	theme := filepath.Join(home, ".local/share/openriot/config/rofi/simple-tokyonight.rasi")
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
	notify.SendNotify(icon, procName, "Starting "+procName+"...", "normal", 2000, 0)
	exec.Command(cmdArgs[0], append(cmdArgs[1:], os.Getenv("HOME")+"/.local/share/openriot/config/bin/"+procName)...).Start()
	return nil
}

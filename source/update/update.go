package update

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"openriot/notify"
	"openriot/polybar"
)

const (
	updateIcon   = "󰋻" // Update available
	noUpdateIcon = "󰚇" // Up to date
	unknownIcon  = "?"
	driftIcon    = "󰀦" // Warning: kernel drift detected
)

// driftThreshold is how old the running kernel must be before we
// consider the base/packages out of sync on -current. Same as the
// helper in source/commands/helpers.go — kept in lockstep.
const driftThreshold = 14 * 24 * time.Hour

// hasKernelDrift returns true if the running kernel is older than
// driftThreshold. Only meaningful on -current (-snapshots); on a
// release branch the kernel date is fixed at install time and drift
// does not apply.
func hasKernelDrift() (bool, time.Time) {
	cmd := exec.Command("sysctl", "-n", "kern.version")
	output, err := cmd.Output()
	if err != nil {
		return false, time.Time{}
	}
	line := strings.TrimSpace(string(output))
	if !strings.Contains(strings.ToLower(line), "current") {
		return false, time.Time{}
	}
	idx := strings.Index(line, ": ")
	if idx < 0 {
		return false, time.Time{}
	}
	dateStr := line[idx+2:]
	buildDate, err := time.Parse("Mon Jan 2 15:04:05 MST 2006", dateStr)
	if err != nil {
		return false, time.Time{}
	}
	return time.Since(buildDate) > driftThreshold, buildDate
}

func Get() string {
	drift, _ := hasKernelDrift()
	if drift {
		return polybar.Icon(driftIcon)
	}
	local := GetLocalVersion()
	remote := GetRemoteVersion()
	return iconForComparison(local, remote)
}

func iconForComparison(local, remote string) string {
	if local == "unknown" {
		return polybar.Icon(unknownIcon)
	}
	if remote == "unknown" {
		return polybar.Icon(noUpdateIcon)
	}
	if CompareVersions(local, remote) < 0 {
		return polybar.Icon(updateIcon)
	}
	return polybar.Icon(noUpdateIcon)
}

func Click() error {
	if drift, buildDate := hasKernelDrift(); drift {
		days := int(time.Since(buildDate).Hours() / 24)
		notify.SendNotify("sysupgrade",
			"System Drift",
			fmt.Sprintf("Kernel is %d days old — run: doas sysupgrade -s", days),
			"critical", 5000, 0)
		dateStr := buildDate.Format("Jan 2 2006")
		script := fmt.Sprintf(
			"printf 'Kernel built: %s (%%d days ago)\\n\\nBase and packages are out of sync.\\n\\nSync now?\\n  [y] doas sysupgrade -s && (reboot) && doas pkg_add -u\\n  [N] Skip\\n\\n[y/N]: ' %d; read -r ans; case \"$ans\" in [yY]) doas sysupgrade -s ;; *) echo Skipped.; sleep 1 ;; esac",
			dateStr, days)
		exec.Command("alacritty", "--class", "openriot_drift", "-e", "sh", "-c", script).Start()
		return nil
	}

	local := GetLocalVersion()
	remote := GetRemoteVersion()

	if remote == "unknown" {
		notify.SendNotify("desktop", "Desktop", "Version check failed...", "normal", 3000, 0)
		return nil
	}

	if CompareVersions(local, remote) < 0 {
		notify.SendNotify("upgrade", "OpenRiot Update", fmt.Sprintf("v%s - Update available!", remote), "normal", 3000, 0)
		cmd := fmt.Sprintf(`printf "You are about to upgrade OpenRiot v%s... are you sure? [Y/n] "; read -r ans; case "$ans" in [yY]|"") curl -fsSL https://openriot.org/setup.sh | sh ;; *) echo "Canceled."; sleep 1 ;; esac`, remote)
		exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
		return nil
	}

	notify.SendNotify("desktop", "OpenRiot Update", fmt.Sprintf("v%s - up to date", local), "normal", 3000, 0)
	return nil
}

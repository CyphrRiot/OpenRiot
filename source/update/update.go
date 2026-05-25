package update

import (
	"fmt"
	"os/exec"

	"openriot/notify"
	"openriot/polybar"
)

const (
	updateIcon   = "󰋻" // Update available
	noUpdateIcon = "󰚇" // Up to date
	unknownIcon  = "?"
)

func Get() string {
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

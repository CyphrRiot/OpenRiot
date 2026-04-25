package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"openriot/notify"
)

var home, _ = os.UserHomeDir()

const (
	updateIcon   = "󰋻" // Update available
	noUpdateIcon = "󰚇" // Up to date
	unknownIcon  = "?"
)

func Get() string {
	local := getLocalVersion()
	remote := getRemoteVersion()
	return iconForComparison(local, remote)
}

func iconForComparison(local, remote string) string {
	if local == "unknown" || remote == "unknown" {
		return unknownIcon
	}
	if CompareVersions(local, remote) < 0 {
		return updateIcon
	}
	return noUpdateIcon
}

func Click() error {
	local := getLocalVersion()
	remote := getRemoteVersion()

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

func getLocalVersion() string {
	path := filepath.Join(home, ".local/share/openriot/VERSION")
	data, err := os.ReadFile(path)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

func getRemoteVersion() string {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://openriot.org/VERSION")
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

// GetWithTimeout returns icon with specified timeout for remote check
func GetWithTimeout(timeout time.Duration) string {
	local := getLocalVersion()
	remote := getRemoteVersion()
	return iconForComparison(local, remote)
}

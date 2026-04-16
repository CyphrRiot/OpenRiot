package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"openriot/notify"
)

const (
	updateIcon   = "󰋻" // Update available
	noUpdateIcon = "󰚇" // Up to date
	unknownIcon  = "?"
)

func getCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache/openriot/remote.version")
}

func readCacheVersion() string {
	data, err := os.ReadFile(getCachePath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeCacheVersion(version string) error {
	cachePath := getCachePath()
	os.MkdirAll(filepath.Dir(cachePath), 0755)
	return os.WriteFile(cachePath, []byte(version), 0600)
}

func Get() string {
	local := getLocalVersion()
	remote := getRemoteVersionWithCache()

	if local == "unknown" || remote == "unknown" {
		return unknownIcon
	}

	if compareVersions(local, remote) < 0 {
		return updateIcon
	}
	return noUpdateIcon
}

func Click() error {
	local := getLocalVersion()
	cached := readCacheVersion()

	if cached == "" {
		// No cached version - user needs to wait for next poll
		notify.SendNotify("desktop", "Desktop", "Version check in progress...", "normal", 3000, 0)
		return nil
	}

	if compareVersions(local, cached) < 0 {
		// Update available - notify then launch upgrade confirmation
		notify.SendNotify("upgrade", "Desktop", fmt.Sprintf("v%s - Update available!", cached), "normal", 3000, 0)
		cmd := `printf "You are about to upgrade OpenRiot... are you sure? [Y/n] "; read -r ans; case "$ans" in [yY]|"") curl -fsSL https://openriot.org/setup.sh | sh ;; *) echo "Canceled."; sleep 1 ;; esac`
		exec.Command("alacritty", "--class", "openriot_upgrade", "-e", "sh", "-c", cmd).Start()
		return nil
	}

	// No update available - show notification
	notify.SendNotify("desktop", "Desktop", fmt.Sprintf("v%s - up to date", local), "normal", 3000, 0)
	return nil
}

func getLocalVersion() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".local/share/openriot/VERSION")
	data, err := os.ReadFile(path)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

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

func getRemoteVersionWithCache() string {
	remote := getRemoteVersion()
	if remote != "unknown" {
		writeCacheVersion(remote)
	}
	return remote
}

// compareVersions compares two semantic versions (a vs b)
// Returns: 1 if a > b, 0 if a == b, -1 if a < b
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
		if vA > vB {
			return 1
		}
		if vA < vB {
			return -1
		}
	}
	return 0
}

// GetWithTimeout returns icon with specified timeout for remote check
func GetWithTimeout(timeout time.Duration) string {
	local := getLocalVersion()
	remote := readCacheVersion()

	if remote != "" {
		if compareVersions(local, remote) < 0 {
			return updateIcon
		}
		return noUpdateIcon
	}

	done := make(chan string, 1)
	go func() {
		done <- getRemoteVersionWithCache()
	}()

	select {
	case remote := <-done:
		if local == "unknown" || remote == "unknown" {
			return unknownIcon
		} else if compareVersions(local, remote) < 0 {
			return updateIcon
		} else {
			return noUpdateIcon
		}
	case <-time.After(timeout):
		return noUpdateIcon
	}
}

package update

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"openriot/installer"
	"openriot/paths"
)

// GetLocalVersion reads the local VERSION file
func GetLocalVersion() string {
	data, err := os.ReadFile(paths.OpenRiotDir("VERSION"))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// GetRemoteVersion fetches VERSION from openriot.org
func GetRemoteVersion() string {
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

// CompareVersions compares two semantic versions (a vs b)
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func CompareVersions(a, b string) int {
	return installer.CompareVersions(a, b)
}
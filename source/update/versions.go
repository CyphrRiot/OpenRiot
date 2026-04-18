package update

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GetLocalVersion reads the local VERSION file
func GetLocalVersion() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "unknown"
	}
	versionPath := filepath.Join(homeDir, ".local", "share", "openriot", "VERSION")
	data, err := os.ReadFile(versionPath)
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
		if vA < vB {
			return -1
		}
		if vA > vB {
			return 1
		}
	}
	return 0
}
package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"openriot/logger"
)

// CrushUpgrade provides crush auto-upgrade functionality
type CrushUpgrade struct{}

// NewCrushUpgrade creates a new CrushUpgrade instance
func NewCrushUpgrade() *CrushUpgrade {
	return &CrushUpgrade{}
}

// Run executes the crush upgrade. Returns error on failure.
func (c *CrushUpgrade) Run() error {
	logger.Info("Checking for crush upgrades...")

	currentVer := c.getCurrentVersion()
	logger.Info(fmt.Sprintf("Current version: v%s", currentVer))

	latestVer, err := c.getLatestVersion()
	if err != nil {
		logger.Warn(fmt.Sprintf("Could not fetch latest version: %v", err))
		return nil
	}

	if latestVer == "" {
		logger.Warn("Could not determine latest version")
		return nil
	}

	logger.Info(fmt.Sprintf("Latest version: v%s", latestVer))

	if currentVer != "" && strings.Compare(currentVer, latestVer) >= 0 {
		logger.Info(fmt.Sprintf("crush v%s is up to date", currentVer))
		return nil
	}

	if err := c.install(latestVer); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}

	logger.Done(fmt.Sprintf("Crush upgraded: v%s → v%s", currentVer, latestVer))
	return nil
}

// getCurrentVersion returns installed crush version or empty string
func (c *CrushUpgrade) getCurrentVersion() string {
	cmd := exec.Command("crush", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	line := strings.TrimSpace(string(output))
	line = strings.TrimPrefix(line, "crush version v")
	line = strings.TrimPrefix(line, "v")
	if idx := strings.IndexAny(line, " "); idx > 0 {
		line = line[:idx]
	}
	return line
}

// getLatestVersion fetches latest version from GitHub API
func (c *CrushUpgrade) getLatestVersion() (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/charmbracelet/crush/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var r struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}

	return strings.TrimPrefix(r.TagName, "v"), nil
}

// install downloads and installs the latest crush binary
func (c *CrushUpgrade) install(version string) error {
	url := fmt.Sprintf("https://github.com/charmbracelet/crush/releases/download/v%s/crush_%s_Openbsd_x86_64.tar.gz", version, version)

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmpFile := filepath.Join(os.TempDir(), "crush.tar.gz")
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpFile)
		return err
	}
	f.Close()
	defer os.Remove(tmpFile)

	cmd := exec.Command("tar", "-xzf", tmpFile, "-C", os.TempDir())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	extractDir := filepath.Join(os.TempDir(), fmt.Sprintf("crush_%s_Openbsd_x86_64", version))
	binPath := filepath.Join(extractDir, "crush")

	cmd = exec.Command("doas", "mv", binPath, "/usr/local/bin/crush")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	cmd = exec.Command("doas", "chmod", "+x", "/usr/local/bin/crush")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	homeDir, _ := os.UserHomeDir()
	compDir := filepath.Join(homeDir, ".config", "fish", "completions")
	os.MkdirAll(compDir, 0755)

	srcComp := filepath.Join(extractDir, "completions", "crush.fish")
	dstComp := filepath.Join(compDir, "crush.fish")
	if _, err := os.Stat(srcComp); err == nil {
		data, _ := os.ReadFile(srcComp)
		os.WriteFile(dstComp, data, 0644)
	}

	os.RemoveAll(extractDir)

	return nil
}
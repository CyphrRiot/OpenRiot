package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/logger"
)

// CrushUpgrade provides crush auto-upgrade functionality
type CrushUpgrade struct{}

// NewCrushUpgrade creates a new CrushUpgrade instance
func NewCrushUpgrade() *CrushUpgrade {
	return &CrushUpgrade{}
}

// Run executes the crush upgrade
func (c *CrushUpgrade) Run() {
	logger.Info("Checking for crush upgrades...")

	currentVer := c.getCurrentVersion()
	logger.Info(fmt.Sprintf("Current version: v%s", currentVer))

	latestVer, err := c.getLatestVersion()
	if err != nil {
		logger.Warn(fmt.Sprintf("Could not fetch latest version: %v", err))
		os.Exit(0)
	}

	if latestVer == "" {
		logger.Warn("Could not determine latest version")
		os.Exit(0)
	}

	logger.Info(fmt.Sprintf("Latest version: v%s", latestVer))

	if currentVer != "" && strings.Compare(currentVer, latestVer) >= 0 {
		logger.Info(fmt.Sprintf("crush v%s is up to date", currentVer))
		os.Exit(0)
	}

	if err := c.install(latestVer); err != nil {
		logger.Fail(fmt.Sprintf("Upgrade failed: %v", err))
		os.Exit(1)
	}

	logger.Done(fmt.Sprintf("Crush upgraded: v%s → v%s", currentVer, latestVer))
	os.Exit(0)
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
	resp, err := http.Get("https://api.github.com/repos/charmbracelet/crush/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	content := string(body)
	idx := strings.Index(content, `"tag_name"`)
	if idx < 0 {
		return "", nil
	}

	start := idx + len(`"tag_name"`)
	for i := start; i < len(content); i++ {
		if content[i] == 'v' && i+1 < len(content) && content[i+1] >= '0' && content[i+1] <= '9' {
			start = i + 1
			break
		}
	}

	end := start
	for i := start; i < len(content); i++ {
		c := content[i]
		if (c >= '0' && c <= '9') || c == '.' {
			end = i + 1
		} else {
			break
		}
	}

	return content[start:end], nil
}

// install downloads and installs the latest crush binary
func (c *CrushUpgrade) install(version string) error {
	url := fmt.Sprintf("https://github.com/charmbracelet/crush/releases/download/v%s/crush_%s_Openbsd_x86_64.tar.gz", version, version)

	resp, err := http.Get(url)
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
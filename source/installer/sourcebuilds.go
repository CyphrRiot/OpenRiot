package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"openriot/config"
	"openriot/logger"
)

// SourceBuilds executes all source build commands from modules with type "Source".
func SourceBuilds(cfg *config.Config, testMode bool) error {
	allModules := cfg.GetAllModules()
	for _, module := range allModules {
		if module.Type != "Source" || len(module.Build) == 0 {
			continue
		}

		logger.LogMessage("INFO", fmt.Sprintf("Building %s from source...", module.Start))

		for _, cmd := range module.Build {
			if testMode {
				logger.LogMessage("INFO", fmt.Sprintf("[DRY-RUN] %s", cmd))
				continue
			}

			// Check for offline tarball first (bundled in ISO at /etc/openriot/)
			// This replaces git clone for wlsunset when network is unavailable
			if strings.Contains(cmd, "git clone") && strings.Contains(cmd, "wlsunset") {
				tarball := "/etc/openriot/wlsunset.tar.gz"
				if _, err := os.Stat(tarball); err == nil {
					tmpDir := "/tmp/openriot-wlsunset-src"
					os.RemoveAll(tmpDir)
					if err := os.MkdirAll(tmpDir, 0755); err == nil {
						extractCmd := exec.Command("tar", "-xzf", tarball, "-C", tmpDir)
						if err := extractCmd.Run(); err == nil {
							logger.LogMessage("INFO", "Extracted wlsunset from offline tarball")
						}
					}
				}
			}

			c := exec.Command("sh", "-c", cmd)
			output, err := c.CombinedOutput()
			if err != nil {
				logger.LogMessage("WARN", fmt.Sprintf("Build cmd '%s' failed:\n%s", cmd, output))
				return fmt.Errorf("source build failed: %w", err)
			}
			logger.LogMessage("SUCCESS", fmt.Sprintf("Built: %s", cmd))
		}
	}
	return nil
}

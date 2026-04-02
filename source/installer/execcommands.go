package installer

import (
	"fmt"
	"os/exec"
	"strings"

	"openriot/config"
	"openriot/logger"
)

// ExecCommands executes commands from packages.yaml modules
func ExecCommands(cfg *config.Config, dryRun bool) error {
	// Get all modules
	modules := cfg.GetAllModules()

	for _, module := range modules {
		for _, cmd := range module.Commands {
			// Skip empty commands
			if strings.TrimSpace(cmd) == "" {
				continue
			}

			// Log the command
			if dryRun {
				logger.LogMessage("INFO", fmt.Sprintf("[DRY RUN] %s", cmd))
				continue
			}

			// Execute the command
			logger.LogMessage("INFO", fmt.Sprintf("Running: %s", cmd))

			// Execute using shell -c for proper parsing
			execCmd := exec.Command("/bin/sh", "-c", cmd)
			output, err := execCmd.CombinedOutput()

			if err != nil {
				logger.LogMessage("WARN", fmt.Sprintf("Command failed: %s - %v", cmd, err))
				// Continue even if a command fails - don't stop the whole install
			} else {
				if len(output) > 0 {
					logger.LogMessage("DEBUG", fmt.Sprintf("Output: %s", strings.TrimSpace(string(output))))
				}
			}
		}
	}

	return nil
}

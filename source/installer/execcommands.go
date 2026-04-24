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
		for _, entry := range module.Commands {
			// Skip empty commands
			if strings.TrimSpace(entry.Cmd) == "" {
				continue
			}

			// Log the command
			if dryRun {
				logger.Info(fmt.Sprintf("[DRY RUN] %s", entry.Desc))
				continue
			}

			// Show command being run
			logger.Info(fmt.Sprintf("Running: %s", entry.Desc))

			// Execute the command
			execCmd := exec.Command("/bin/sh", "-c", entry.Cmd)
			output, err := execCmd.CombinedOutput()

			if err != nil {
				logger.Warn(fmt.Sprintf("Command failed: %s - %v\n%s", entry.Desc, err, string(output)))
				// Continue even if a command fails - don't stop the whole install
			}
		}
	}

	return nil
}

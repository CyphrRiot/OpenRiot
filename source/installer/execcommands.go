package installer

import (
	"fmt"
	"os/exec"
	"strings"

	"openriot/config"
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
				fmt.Printf("[INFO]  [DRY RUN] %s\n", cmd)
				continue
			}

			// Execute the command
			fmt.Printf("[INFO]  Running: %s\n", cmd)

			// Execute using shell -c for proper parsing
			execCmd := exec.Command("/bin/sh", "-c", cmd)
			output, err := execCmd.CombinedOutput()

			if err != nil {
				fmt.Printf("[WARN]  Command failed: %s - %v\n", cmd, err)
				// Continue even if a command fails - don't stop the whole install
			} else {
				if len(output) > 0 {
					fmt.Printf("[DEBUG] Output: %s\n", strings.TrimSpace(string(output)))
				}
			}
		}
	}

	return nil
}

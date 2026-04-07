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
				fmt.Printf("%s[INFO]%s  [DRY RUN] %s\n", Blue, Reset, cmd)
				continue
			}

			// Execute the command
			fmt.Printf("%s[INFO]%s  Running: %s\n", Blue, Reset, cmd)

			// Execute using shell -c for proper parsing
			execCmd := exec.Command("/bin/sh", "-c", cmd)
			output, err := execCmd.CombinedOutput()

			if err != nil {
				fmt.Printf("%s[WARN]%s  Command failed: %s - %v\n", Yellow, Reset, cmd, err)
				// Continue even if a command fails - don't stop the whole install
			} else {
				if len(output) > 0 {
					fmt.Printf("%s[DEBUG]%s Output: %s\n", White, Reset, strings.TrimSpace(string(output)))
				}
			}
		}
	}

	return nil
}

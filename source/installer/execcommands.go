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
		for _, entry := range module.Commands {
			// Skip empty commands
			if strings.TrimSpace(entry.Cmd) == "" {
				continue
			}

			// Log the command
			if dryRun {
				fmt.Printf("%s[INFO]%s [DRY RUN] %s\n", Blue, Reset, entry.Desc)
				continue
			}

			// Execute the command silently
			execCmd := exec.Command("/bin/sh", "-c", entry.Cmd)
			_, err := execCmd.CombinedOutput()

			if err != nil {
				fmt.Printf("%s[WARN]%s Command failed: %s - %v\n", Yellow, Reset, entry.Desc, err)
				// Continue even if a command fails - don't stop the whole install
			}
		}
	}

	return nil
}

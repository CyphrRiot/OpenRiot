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
				fmt.Printf("%s[INFO]%s [DRY RUN] %s\n", Cyan, Reset, entry.Desc)
				continue
			}

			// Show command being run
			fmt.Printf("%s[INFO]%s Running: %s\n", Cyan, Reset, entry.Desc)

			// Execute the command
			execCmd := exec.Command("/bin/sh", "-c", entry.Cmd)
			output, err := execCmd.CombinedOutput()

			if err != nil {
				fmt.Printf("%s[WARN]%s Command failed: %s - %v\n%s%s\n", Yellow, Reset, entry.Desc, err, string(output), Reset)
				// Continue even if a command fails - don't stop the whole install
			}
		}
	}

	return nil
}

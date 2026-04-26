package installer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"openriot/config"
	"openriot/logger"
)

// ExecCommands executes commands from packages.yaml modules
func ExecCommands(cfg *config.Config, dryRun bool) error {
	refs, err := cfg.GetAllModulesOrdered()
	if err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}

	for _, ref := range refs {
		module := ref.Module
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

			// Execute the command with a 15-minute timeout
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			execCmd := exec.CommandContext(ctx, "/bin/sh", "-c", entry.Cmd)

			// Notify if the command runs longer than 2 minutes
			done := make(chan struct{})
			go func() {
				select {
				case <-time.After(2 * time.Minute):
					logger.Info(fmt.Sprintf("Still running (>%dm): %s", 2, entry.Desc))
				case <-done:
				}
			}()

			output, err := execCmd.CombinedOutput()
			close(done)
			cancel()

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					logger.Warn(fmt.Sprintf("Command timed out after 15m: %s", entry.Desc))
				} else {
					logger.Warn(fmt.Sprintf("Command failed: %s - %v\n%s", entry.Desc, err, string(output)))
				}
				// Continue even if a command fails - don't stop the whole install
			}
		}
	}

	return nil
}

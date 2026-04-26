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

// SourceBuilds executes all source build commands from modules with type "Source".
// Commands are run as-is; each command is a separate shell invocation.
func SourceBuilds(cfg *config.Config, testMode bool) error {
	refs, err := cfg.GetAllModulesOrdered()
	if err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}
	for _, ref := range refs {
		module := ref.Module
		if module.Type != "Source" || len(module.Build) == 0 {
			continue
		}

		anySucceeded := false
		for _, rawEntry := range module.Build {
			// Parse the build entry (either simple string or desc/cmd map)
			var cmd string
			var desc string

			switch v := rawEntry.(type) {
			case string:
				cmd = v
				desc = module.Start
			case map[string]any:
				if d, ok := v["desc"].(string); ok {
					desc = d
				}
				if c, ok := v["cmd"].(string); ok {
					cmd = c
				}
				if cmd == "" {
					cmd = desc
				}
			default:
				logger.Warn(fmt.Sprintf("Unknown build entry type: %T", rawEntry))
				continue
			}

			if testMode {
				logger.Info(fmt.Sprintf("[DRY-RUN] %s", desc))
				continue
			}

			logger.Info(fmt.Sprintf("%s...", desc))

			// Execute the command with a 10-minute timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			c := exec.CommandContext(ctx, "/bin/sh", "-c", cmd)

			// Notify if the command runs longer than 2 minutes
			done := make(chan struct{})
			go func() {
				select {
				case <-time.After(2 * time.Minute):
					logger.Info(fmt.Sprintf("Still running (>%dm): %s", 2, desc))
				case <-done:
				}
			}()

			output, err := c.CombinedOutput()
			close(done)

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					logger.Warn(fmt.Sprintf("Build timed out after 10m: %s", desc))
				} else {
					logger.Warn(fmt.Sprintf("%s failed: %v\n%s", desc, err, string(output)))
				}
				continue
			}

			outputStr := string(output)
			if strings.Contains(outputStr, "[SKIP]") {
				continue
			}
			anySucceeded = true
		}

		// Log end message once per module, not per command
		if anySucceeded || testMode {
			logger.Done(module.End)
		}
	}
	return nil
}

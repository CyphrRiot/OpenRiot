package installer

import (
	"fmt"
	"os/exec"

	"openriot/config"
)

// SourceBuilds executes all source build commands from modules with type "Source".
// Commands are run as-is; each command is a separate shell invocation.
func SourceBuilds(cfg *config.Config, testMode bool) error {
	allModules := cfg.GetAllModules()
	for _, module := range allModules {
		if module.Type != "Source" || len(module.Build) == 0 {
			continue
		}

		fmt.Printf("[INFO]  %s...\n", module.Start)

		for _, cmd := range module.Build {
			if testMode {
				fmt.Printf("[INFO]  [DRY-RUN] %s\n", cmd)
				continue
			}

			// Execute each build step as a separate shell invocation
			c := exec.Command("/bin/sh", "-c", cmd)
			output, err := c.CombinedOutput()
			if err != nil {
				fmt.Printf("[WARN]  Build command failed:\n  command: %s\n  error: %v\n  output: %s\n", cmd, err, string(output))
				// Continue on error - don't stop the whole install for one failed source build
				continue
			}
			fmt.Printf("[INFO]  Built: %s\n", cmd)
		}
	}
	return nil
}

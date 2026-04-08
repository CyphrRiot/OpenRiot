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

		fmt.Printf("%s[INFO]%s  %s...\n", Blue, Reset, module.Start)

		for _, cmd := range module.Build {
			if testMode {
				fmt.Printf("%s[INFO]%s  [DRY-RUN] %s\n", Blue, Reset, cmd)
				continue
			}

			// Execute each build step as a separate shell invocation
			c := exec.Command("/bin/sh", "-c", cmd)
			output, err := c.CombinedOutput()
			if err != nil {
				fmt.Printf("%s[WARN]%s  Build command failed:\n  command: %s\n  error: %v\n  output: %s\n", Yellow, Reset, cmd, err, string(output))
				// Continue on error - don't stop the whole install for one failed source build
				continue
			}
			fmt.Printf("%s[DONE]%s  %s\n", Green, Reset, module.End)
		}
	}
	return nil
}

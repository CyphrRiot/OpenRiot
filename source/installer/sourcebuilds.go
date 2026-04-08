package installer

import (
	"fmt"
	"os/exec"
	"strings"

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

		for _, cmdEntry := range module.Build {
			// Parse the build entry (either simple string or desc/cmd)
			var cmd string
			var desc string
			if strings.Contains(cmdEntry, "\n") || strings.HasPrefix(cmdEntry, "desc:") {
				// It's a structured entry (desc/cmd format)
				parts := strings.SplitN(cmdEntry, "\n", 2)
				for _, part := range parts {
					if strings.HasPrefix(part, "cmd:") {
						cmd = strings.TrimPrefix(strings.TrimSpace(part), "cmd:")
					}
					if strings.HasPrefix(part, "desc:") {
						desc = strings.TrimPrefix(strings.TrimSpace(part), "desc:")
					}
				}
				if cmd == "" {
					cmd = cmdEntry
				}
			} else {
				cmd = cmdEntry
				desc = module.Start
			}

			if testMode {
				fmt.Printf("%s[INFO]%s [DRY-RUN] %s\n", Blue, Reset, cmd)
				continue
			}

			// Execute each build step as a separate shell invocation
			c := exec.Command("/bin/sh", "-c", cmd)
			output, err := c.CombinedOutput()
			if err != nil {
				fmt.Printf("%s[WARN]%s %s failed: %v\n", Yellow, Reset, desc, err)
				// Continue on error - don't stop the whole install for one failed source build
				continue
			}

			// Only show DONE if not already installed (skip output contains [SKIP])
			outputStr := string(output)
			if strings.Contains(outputStr, "[SKIP]") {
				continue
			}
			fmt.Printf("%s[DONE]%s %s\n", Green, Reset, desc)
		}
	}
	return nil
}

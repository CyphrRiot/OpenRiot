package installer

import (
	"fmt"
	"os/exec"
	"strings"

	"openriot/config"
	"openriot/logger"
)

// SourceBuilds executes all source build commands from modules with type "Source".
// Commands are run as-is; each command is a separate shell invocation.
func SourceBuilds(cfg *config.Config, testMode bool) error {
	allModules := cfg.GetAllModules()
	for _, module := range allModules {
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
			c := exec.Command("/bin/sh", "-c", cmd)
			output, err := c.CombinedOutput()
			if err != nil {
				logger.Warn(fmt.Sprintf("%s failed: %v\n%s", desc, err, string(output)))
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

package helix

import (
	"fmt"
	"os"
	"path/filepath"

	"openriot/installer"
	"openriot/paths"
)

// Setup renders the Helix theme template with canonical colors.
func Setup() int {
	templatePath := paths.OpenRiotDir("config", "helix", "themes",
		"openriot.toml.tmpl")
	configPath := paths.Join(".config", "helix", "themes",
		"openriot.toml")

	content, _, err := installer.RenderTemplateString(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "helix setup: %v\n", err)
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "helix setup: cannot create dir: %v\n",
			err)
		return 1
	}
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "helix setup: cannot write theme: %v\n",
			err)
		return 1
	}

	fmt.Println("[DONE] Helix theme rendered.")
	return 0
}

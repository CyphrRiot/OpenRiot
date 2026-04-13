package windowicon

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

var mappings map[string]string

func Get(class string) string {
	loadMappings()
	class = strings.ToLower(class)
	if icon, ok := mappings[class]; ok {
		return icon
	}
	return ""
}

func loadMappings() {
	if mappings != nil {
		return
	}
	mappings = make(map[string]string)

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".local/share/openriot/config/window/icons.toml")

	// Try multiple locations for config
	locations := []string{
		configPath,
		filepath.Join(homeDir, "Code/OpenRiot/config/window/icons.toml"),
	}

	var data map[string]interface{}
	for _, path := range locations {
		if _, err := toml.DecodeFile(path, &data); err == nil {
			break
		}
	}

	// Flatten all sections into single map
	flattenMaps(data)
}

func flattenMaps(data map[string]interface{}) {
	for _, sectionData := range data {
		if sm, ok := sectionData.(map[string]interface{}); ok {
			for k, v := range sm {
				if str, ok := v.(string); ok {
					mappings[strings.ToLower(k)] = str
				}
			}
		}
	}
}

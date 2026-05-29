package windowicon

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"openriot/paths"
)

var mappings map[string]string

func Get(class string) string {
	loadMappings()
	class = strings.ToLower(class)
	if icon, ok := mappings[class]; ok {
		return icon
	}
	return ""
}

// GetAllMappings returns all window class→icon mappings
func GetAllMappings() map[string]string {
	loadMappings()
	return mappings
}

// GetAllWindowIcons calls i3-msg once, returns map of class→icon for all windows
func GetAllWindowIcons() map[string]string {
	icons := GetAllMappings()
	result := make(map[string]string)

	cmd := exec.Command("i3-msg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	var tree i3Tree
	if err := json.Unmarshal(output, &tree); err != nil {
		return result
	}

	classes := make(map[string]bool)
	collectWindowClasses(tree.Nodes, classes)
	collectWindowClasses(tree.FloatingNodes, classes)

	for class := range classes {
		icon := ""
		if v, ok := icons[strings.ToLower(class)]; ok {
			icon = v
		}
		result[class] = icon
	}
	return result
}

func IsPrivateFirefox(class, name string) bool {
	return strings.ToLower(class) == "firefox" && strings.Contains(name, "Private Browsing")
}

func collectWindowClasses(nodes []i3Node, classes map[string]bool) {
	for _, n := range nodes {
		if n.Window != 0 && n.WindowProperties.Class != "" {
			cls := n.WindowProperties.Class
			if IsPrivateFirefox(cls, n.Name) {
				classes["firefox-private"] = true
			} else {
				classes[cls] = true
			}
		}
		collectWindowClasses(n.Nodes, classes)
		collectWindowClasses(n.FloatingNodes, classes)
	}
}

type i3Tree struct {
	Nodes         []i3Node `json:"nodes"`
	FloatingNodes []i3Node `json:"floating_nodes"`
}

type i3Node struct {
	Name            string       `json:"name"`
	Window          int          `json:"window"`
	WindowProperties windowProps `json:"window_properties"`
	Nodes           []i3Node     `json:"nodes"`
	FloatingNodes   []i3Node     `json:"floating_nodes"`
}

type windowProps struct {
	Class string `json:"class"`
}

func loadMappings() {
	if mappings != nil {
		return
	}
	mappings = make(map[string]string)

	configPath := paths.OpenRiotDir("config", "window", "icons.toml")

	// Try multiple locations for config
	locations := []string{
		configPath,
		paths.Join("Code", "OpenRiot", "config", "window", "icons.toml"),
	}

	var data map[string]any
	for _, path := range locations {
		if _, err := toml.DecodeFile(path, &data); err == nil {
			break
		}
	}

	// Flatten all sections into single map
	flattenMaps(data)
}

func flattenMaps(data map[string]any) {
	for _, sectionData := range data {
		if sm, ok := sectionData.(map[string]any); ok {
			for k, v := range sm {
				if str, ok := v.(string); ok {
					mappings[strings.ToLower(k)] = str
				}
			}
		}
	}
}

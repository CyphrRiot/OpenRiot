package config

// Config represents the YAML structure from packages.yaml
type Config struct {
	Core    map[string]Module `yaml:"core"`
	System  map[string]Module `yaml:"system"`
	Desktop map[string]Module `yaml:"desktop"`
	Media   map[string]Module `yaml:"media"`
	Fonts   map[string]Module `yaml:"fonts"`
	Themes  map[string]Module `yaml:"themes"`
	Source  map[string]Module `yaml:"source"`
}

// CommandEntry represents a single command with a description
type CommandEntry struct {
	Desc string `yaml:"desc"`
	Cmd  string `yaml:"cmd"`
}

// Module represents a single installation module
type Module struct {
	Packages []string        `yaml:"packages"`
	Configs  []ConfigRule   `yaml:"configs"`
	Commands []CommandEntry `yaml:"commands,omitempty"`
	Depends  []string      `yaml:"depends,omitempty"`
	Start    string        `yaml:"start"`
	End      string        `yaml:"end"`
	Type     string        `yaml:"type"`
	Critical bool          `yaml:"critical,omitempty"`
	Build    []string      `yaml:"build,omitempty"`
}

// ConfigRule represents a configuration copying rule
type ConfigRule struct {
	Pattern          string   `yaml:"pattern"`
	Target           string   `yaml:"target,omitempty"`
	PreserveIfExists []string `yaml:"preserve_if_exists,omitempty"`
}

package config

import (
	"os/exec"
	"strings"
)

// Config represents the YAML structure from packages.yaml
type Config struct {
	OpenBSDVersion string            `yaml:"openbsd_version"`
	Core    map[string]Module `yaml:"core"`
	System  map[string]Module `yaml:"system"`
	Desktop map[string]Module `yaml:"desktop"`
	Media   map[string]Module `yaml:"media"`
	Fonts   map[string]Module `yaml:"fonts"`
	Themes  map[string]Module `yaml:"themes"`
	Source  map[string]Module `yaml:"source"`
	Crypto  map[string]Module `yaml:"crypto"`
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
	Build    []any           `yaml:"build,omitempty"`
}

// ConfigRule represents a configuration copying rule
type ConfigRule struct {
	Pattern          string   `yaml:"pattern"`
	Target           string   `yaml:"target,omitempty"`
	PreserveIfExists []string `yaml:"preserve_if_exists,omitempty"`
}

// DetectOpenBSDVersion detects the running OpenBSD version.
// Returns "snapshots" if running -current/-snap, otherwise returns the version string (e.g., "7.9").
func DetectOpenBSDVersion() string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "snapshots"
	}
	version := strings.TrimSpace(string(output))
	if strings.Contains(version, "current") || strings.Contains(version, "snap") {
		return "snapshots"
	}
	return version
}

// ResolveOpenBSDVersion returns the configured openbsd_version if set,
// otherwise detects the running version automatically.
func (c *Config) ResolveOpenBSDVersion() string {
	if c.OpenBSDVersion != "" {
		return c.OpenBSDVersion
	}
	return DetectOpenBSDVersion()
}

// IsSnapshot returns true if the target is the -snapshots repository.
func (c *Config) IsSnapshot() bool {
	return c.ResolveOpenBSDVersion() == "snapshots"
}

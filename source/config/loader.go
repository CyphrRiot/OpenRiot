package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

// FindConfigFile looks for packages.yaml in common locations
func FindConfigFile() string {
	// Explicit config dir from environment (set by setup.sh)
	if dir := os.Getenv("OPENRIOT_CONFIG_DIR"); dir != "" {
		path := filepath.Join(dir, "packages.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	// Fallback: relative to binary location
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		path := filepath.Join(dir, "packages.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	locations := []string{
		filepath.Join(os.Getenv("HOME"), ".local/share/openriot/install/packages.yaml"),
		filepath.Join("install", "packages.yaml"),
		filepath.Join("..", "install", "packages.yaml"),
	}
	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// LoadConfig reads and parses the YAML configuration
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	return &cfg, nil
}

// ValidateConfig validates that all required fields are present in the YAML
func ValidateConfig(cfg *Config) error {
	// Build set of all valid module IDs for dependency checking
	validIDs := make(map[string]bool)
	addIDs := func(cat string, mods map[string]Module) {
		for name := range mods {
			validIDs[cat+"."+name] = true
		}
	}
	addIDs("core", cfg.Core)
	addIDs("system", cfg.System)
	addIDs("desktop", cfg.Desktop)
	addIDs("media", cfg.Media)
	addIDs("fonts", cfg.Fonts)
	addIDs("themes", cfg.Themes)
	addIDs("source", cfg.Source)

	// Validate all categories
	for _, cat := range []struct {
		name string
		mods map[string]Module
	}{
		{"core", cfg.Core},
		{"system", cfg.System},
		{"desktop", cfg.Desktop},
		{"media", cfg.Media},
		{"fonts", cfg.Fonts},
		{"themes", cfg.Themes},
		{"source", cfg.Source},
	} {
		if err := validateModuleCategory(cat.name, cat.mods, validIDs); err != nil {
			return err
		}
	}

	return nil
}

// validateModuleCategory validates all modules in a category
func validateModuleCategory(category string, modules map[string]Module, validIDs map[string]bool) error {
	for name, module := range modules {
		fullName := fmt.Sprintf("%s.%s", category, name)

		if module.Start == "" {
			return fmt.Errorf("module %s missing required 'start' field", fullName)
		}

		if module.End == "" {
			return fmt.Errorf("module %s missing required 'end' field", fullName)
		}

		if module.Type == "" {
			return fmt.Errorf("module %s missing required 'type' field", fullName)
		}

		// Validate type is one of the allowed values
		validTypes := []string{"Package", "Git", "System", "File", "Source"}
		if !slices.Contains(validTypes, module.Type) {
			return fmt.Errorf("module %s has invalid type '%s', must be one of: %v", fullName, module.Type, validTypes)
		}

		// Validate dependencies exist
		for _, dep := range module.Depends {
			if !validIDs[dep] {
				return fmt.Errorf("module %s depends on unknown module %s", fullName, dep)
			}
		}

		// Type-field coherence checks
		switch module.Type {
		case "Source":
			if len(module.Build) == 0 {
				return fmt.Errorf("module %s type 'Source' requires non-empty 'build'", fullName)
			}
		case "Package":
			if len(module.Packages) == 0 && len(module.Configs) == 0 && len(module.Commands) == 0 {
				return fmt.Errorf("module %s type 'Package' requires at least one of: packages, configs, commands", fullName)
			}
		}
	}

	return nil
}

// GetAllModules returns all modules from all categories in execution order
func (c *Config) GetAllModules() []Module {
	var modules []Module

	// Core modules first
	for _, m := range c.Core {
		modules = append(modules, m)
	}

	// System modules
	for _, m := range c.System {
		modules = append(modules, m)
	}

	// Desktop modules
	for _, m := range c.Desktop {
		modules = append(modules, m)
	}

	// Media modules
	for _, m := range c.Media {
		modules = append(modules, m)
	}

	// Fonts modules
	for _, m := range c.Fonts {
		modules = append(modules, m)
	}

	// Themes modules
	for _, m := range c.Themes {
		modules = append(modules, m)
	}

	// Source modules (built from source)
	for _, m := range c.Source {
		modules = append(modules, m)
	}

	return modules
}

// ModuleRef ties a Module to its category.name identifier.
type ModuleRef struct {
	Category string
	Name     string
	Module   Module
}

// GetAllModulesOrdered returns all modules in dependency order (topological sort).
// Prerequisite modules always come before their dependents.
func (c *Config) GetAllModulesOrdered() ([]ModuleRef, error) {
	var refs []ModuleRef
	idToIndex := make(map[string]int)

	addCategory := func(cat string, mods map[string]Module) {
		for name, mod := range mods {
			id := cat + "." + name
			idToIndex[id] = len(refs)
			refs = append(refs, ModuleRef{Category: cat, Name: name, Module: mod})
		}
	}

	addCategory("core", c.Core)
	addCategory("system", c.System)
	addCategory("desktop", c.Desktop)
	addCategory("media", c.Media)
	addCategory("fonts", c.Fonts)
	addCategory("themes", c.Themes)
	addCategory("source", c.Source)
	addCategory("crypto", c.Crypto)

	n := len(refs)
	adj := make([][]int, n)
	inDegree := make([]int, n)

	for i, ref := range refs {
		for _, dep := range ref.Module.Depends {
			j, ok := idToIndex[dep]
			if !ok {
				return nil, fmt.Errorf("module %s.%s depends on unknown %s",
					ref.Category, ref.Name, dep)
			}
			adj[j] = append(adj[j], i)
			inDegree[i]++
		}
	}

	var queue []int
	for i, d := range inDegree {
		if d == 0 {
			queue = append(queue, i)
		}
	}

	var ordered []ModuleRef
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		ordered = append(ordered, refs[i])
		for _, j := range adj[i] {
			inDegree[j]--
			if inDegree[j] == 0 {
				queue = append(queue, j)
			}
		}
	}

	if len(ordered) != n {
		return nil, fmt.Errorf("circular dependency detected in modules")
	}

	return ordered, nil
}

// GetPackages returns all unique packages from all modules
func (c *Config) GetPackages() []string {
	seen := make(map[string]bool)
	var packages []string

	allModules := c.GetAllModules()
	for _, module := range allModules {
		for _, pkg := range module.Packages {
			if !seen[pkg] {
				seen[pkg] = true
				packages = append(packages, pkg)
			}
		}
	}

	return packages
}

// GetCommands returns all commands from all modules
func (c *Config) GetCommands() []CommandEntry {
	var commands []CommandEntry

	allModules := c.GetAllModules()
	for _, module := range allModules {
		commands = append(commands, module.Commands...)
	}

	return commands
}

// SaveConfig writes the config back to a YAML file
func (c *Config) SaveConfig(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	return os.WriteFile(filename, data, 0600)
}

// GetBaseName extracts base name from package (e.g., "fish-4.6.0" -> "fish")
func GetBaseName(pkg string) string {
	for i := len(pkg) - 1; i >= 0; i-- {
		if pkg[i] == '-' {
			return pkg[:i]
		}
	}
	return pkg
}

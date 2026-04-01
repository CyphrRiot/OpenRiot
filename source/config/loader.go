package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FindConfigFile looks for packages.yaml in common locations
func FindConfigFile() string {
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
	// Validate core modules
	if err := validateModuleCategory("core", cfg.Core); err != nil {
		return err
	}

	// Validate system modules
	if err := validateModuleCategory("system", cfg.System); err != nil {
		return err
	}

	// Validate desktop modules
	if err := validateModuleCategory("desktop", cfg.Desktop); err != nil {
		return err
	}

	// Validate media modules
	if err := validateModuleCategory("media", cfg.Media); err != nil {
		return err
	}

	return nil
}

// validateModuleCategory validates all modules in a category
func validateModuleCategory(category string, modules map[string]Module) error {
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
		typeValid := false
		for _, validType := range validTypes {
			if module.Type == validType {
				typeValid = true
				break
			}
		}
		if !typeValid {
			return fmt.Errorf("module %s has invalid type '%s', must be one of: %v", fullName, module.Type, validTypes)
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
func (c *Config) GetCommands() []string {
	var commands []string

	allModules := c.GetAllModules()
	for _, module := range allModules {
		commands = append(commands, module.Commands...)
	}

	return commands
}

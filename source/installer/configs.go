package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"openriot/config"
	"openriot/logger"
)

// CopyConfigs copies configuration files from the repo to user's home directory
// It reads config rules from the loaded YAML configuration
func CopyConfigs(repoDir string, cfg *config.Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configSourceDir := filepath.Join(repoDir, "config")
	configDir := filepath.Join(homeDir, ".config")

	logger.LogMessage("INFO", fmt.Sprintf("Copying configs from: %s", configSourceDir))

	// Create ~/.config if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Collect all config rules from all modules
	var allRules []config.ConfigRule

	// Get all modules from all categories
	for _, module := range cfg.Core {
		allRules = append(allRules, module.Configs...)
	}
	for _, module := range cfg.Desktop {
		allRules = append(allRules, module.Configs...)
	}
	for _, module := range cfg.System {
		allRules = append(allRules, module.Configs...)
	}
	for _, module := range cfg.Source {
		allRules = append(allRules, module.Configs...)
	}

	// Process each config rule
	for _, rule := range allRules {
		// Skip empty patterns
		if rule.Pattern == "" {
			continue
		}

		// Determine if this is a glob pattern (contains /*)
		isGlob := strings.Contains(rule.Pattern, "/*")

		if isGlob {
			// Glob pattern: copy all files matching the pattern
			globSrc := filepath.Join(configSourceDir, rule.Pattern)
			globDest := filepath.Join(configDir, rule.Pattern)

			// Expand the glob
			matches, err := filepath.Glob(globSrc)
			if err != nil {
				logger.LogMessage("WARN", fmt.Sprintf("Glob failed for %s: %v", rule.Pattern, err))
				continue
			}

			// Determine base directory for destination
			baseDest := globDest
			if rule.Target != "" {
				if strings.HasPrefix(rule.Target, "~/") {
					baseDest = filepath.Join(homeDir, rule.Target[2:])
				} else {
					baseDest = rule.Target
				}
			}

			for _, srcPath := range matches {
				// Skip directories
				info, err := os.Stat(srcPath)
				if err != nil || info.IsDir() {
					continue
				}

				// Get relative path from source base
				relPath, err := filepath.Rel(configSourceDir, srcPath)
				if err != nil {
					continue
				}
				destPath := filepath.Join(baseDest, relPath)

				// Skip if source doesn't exist
				if _, err := os.Stat(srcPath); os.IsNotExist(err) {
					continue
				}

				// Create destination directory
				destDir := filepath.Dir(destPath)
				if err := os.MkdirAll(destDir, 0755); err != nil {
					logger.LogMessage("WARN", fmt.Sprintf("Failed to create directory %s: %v", destDir, err))
					continue
				}

				// Copy file
				if err := copyFile(srcPath, destPath); err != nil {
					logger.LogMessage("WARN", fmt.Sprintf("Failed to copy %s: %v", srcPath, err))
					continue
				}

				logger.LogMessage("INFO", fmt.Sprintf("Copied %s", relPath))
			}
		} else {
			// Single file pattern
			srcPath := filepath.Join(configSourceDir, rule.Pattern)
			destPath := filepath.Join(configDir, rule.Pattern)

			// If custom target specified, use it instead
			if rule.Target != "" {
				if strings.HasPrefix(rule.Target, "~/") {
					destPath = filepath.Join(homeDir, rule.Target[2:])
				} else {
					destPath = rule.Target
				}
			}

			// Skip if source doesn't exist
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				logger.LogMessage("INFO", fmt.Sprintf("Skipping %s (not found)", rule.Pattern))
				continue
			}

			// Create destination directory
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				logger.LogMessage("WARN", fmt.Sprintf("Failed to create directory %s: %v", destDir, err))
				continue
			}

			// Copy file
			if err := copyFile(srcPath, destPath); err != nil {
				logger.LogMessage("WARN", fmt.Sprintf("Failed to copy %s: %v", rule.Pattern, err))
				continue
			}

			logger.LogMessage("INFO", fmt.Sprintf("Copied %s", rule.Pattern))
		}
	}

	// Copy backgrounds to ~/.local/share/openriot/backgrounds
	if err := copyBackgrounds(repoDir, homeDir); err != nil {
		logger.LogMessage("WARN", fmt.Sprintf("Background copy failed: %v", err))
	}

	return nil
}

// copyBackgrounds copies background images to the backgrounds directory
func copyBackgrounds(repoDir, homeDir string) error {
	bgSourceDir := filepath.Join(repoDir, "backgrounds")
	bgDestDir := filepath.Join(homeDir, ".local", "share", "openriot", "backgrounds")

	// Create destination directory
	if err := os.MkdirAll(bgDestDir, 0755); err != nil {
		return fmt.Errorf("creating backgrounds directory: %w", err)
	}

	// Check if source exists
	if _, err := os.Stat(bgSourceDir); os.IsNotExist(err) {
		logger.LogMessage("INFO", "No backgrounds directory found")
		return nil
	}

	// Copy all jpg files
	entries, err := os.ReadDir(bgSourceDir)
	if err != nil {
		return fmt.Errorf("reading backgrounds directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".jpg") {
			continue
		}

		srcPath := filepath.Join(bgSourceDir, name)
		destPath := filepath.Join(bgDestDir, name)

		if err := copyFile(srcPath, destPath); err != nil {
			logger.LogMessage("WARN", fmt.Sprintf("Failed to copy background %s: %v", name, err))
			continue
		}

		logger.LogMessage("INFO", fmt.Sprintf("Copied background %s", name))
	}

	return nil
}

// copyFile copies a single file
func copyFile(source, dest string) error {
	sourceData, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	if err := os.WriteFile(dest, sourceData, 0644); err != nil {
		return fmt.Errorf("writing dest file: %w", err)
	}

	return nil
}

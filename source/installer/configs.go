package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"openriot/config"
)

// CopyConfigs copies configuration files from the repo to user's home directory
// It reads config rules from the loaded YAML configuration
// If dryRun is true, only logs what would be copied without actually copying
func CopyConfigs(repoDir string, cfg *config.Config, dryRun bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configSourceDir := filepath.Join(repoDir, "config")
	configDir := filepath.Join(homeDir, ".config")

	fmt.Printf("[INFO]  Copying configs from: %s\n", configSourceDir)

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
			// Glob pattern: copy all files matching the pattern, recursing into subdirectories
			globSrc := filepath.Join(configSourceDir, rule.Pattern)
			globDest := filepath.Join(configDir, rule.Pattern)

			// Determine base directory for destination
			// For globs like "fastfetch/*", baseDest should be "fastfetch/" (parent dir)
			baseDest := filepath.Dir(globDest)
			if rule.Target != "" {
				if strings.HasPrefix(rule.Target, "~/") {
					baseDest = filepath.Join(homeDir, rule.Target[2:])
				} else {
					baseDest = rule.Target
				}
			}

			// WalkDir recurses into subdirectories unlike filepath.Glob
			err := filepath.WalkDir(globSrc, func(srcPath string, info fs.DirEntry, err error) error {
				if err != nil {
					return nil // skip inaccessible entries
				}
				if info.IsDir() {
					return nil // skip directories, recurse into them automatically
				}

				// Get relative path from source base
				relPath, err := filepath.Rel(configSourceDir, srcPath)
				if err != nil {
					return nil
				}
				destPath := filepath.Join(baseDest, relPath)

				// Create destination directory
				destDir := filepath.Dir(destPath)
				if err := os.MkdirAll(destDir, 0755); err != nil {
					fmt.Printf("[WARN]  Failed to create directory %s: %v\n", destDir, err)
					return nil
				}

				// Copy file
				if dryRun {
					fmt.Printf("[INFO]  [DRY-RUN] Would copy %s -> %s\n", relPath, destPath)
				} else if err := copyFile(srcPath, destPath); err != nil {
					fmt.Printf("[WARN]  Failed to copy %s: %v\n", srcPath, err)
					return nil
				} else {
					fmt.Printf("[INFO]  Copied %s -> %s\n", relPath, destPath)
				}
				return nil
			})
			if err != nil {
				fmt.Printf("[WARN]  WalkDir failed for %s: %v\n", rule.Pattern, err)
				continue
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
				fmt.Printf("[INFO]  Skipping %s (not found)\n", rule.Pattern)
				continue
			}

			// Create destination directory
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				fmt.Printf("[WARN]  Failed to create directory %s: %v\n", destDir, err)
				continue
			}

			// Copy file
			if dryRun {
				fmt.Printf("[INFO]  [DRY-RUN] Would copy %s -> %s\n", rule.Pattern, destPath)
			} else if err := copyFile(srcPath, destPath); err != nil {
				fmt.Printf("[WARN]  Failed to copy %s: %v\n", rule.Pattern, err)
				continue
			} else {
				fmt.Printf("[INFO]  Copied %s -> %s\n", rule.Pattern, destPath)
			}
		}
	}

	// Copy backgrounds to ~/.local/share/openriot/backgrounds
	if dryRun {
		fmt.Println("[INFO]  [DRY-RUN] Would copy backgrounds")
	} else if err := copyBackgrounds(repoDir, homeDir); err != nil {
		fmt.Printf("[WARN]  Background copy failed: %v\n", err)
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
		fmt.Println("[INFO]  No backgrounds directory found")
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
			fmt.Printf("[WARN]  Failed to copy background %s: %v\n", name, err)
			continue
		}

		fmt.Printf("[INFO]  Copied background %s -> %s\n", name, destPath)
	}

	return nil
}

// copyFile copies a single file, preserving source file permissions
func copyFile(source, dest string) error {
	sourceData, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	// Preserve source file permissions instead of hardcoding 0644
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	mode := info.Mode()

	if err := os.WriteFile(dest, sourceData, mode); err != nil {
		return fmt.Errorf("writing dest file: %w", err)
	}

	return nil
}

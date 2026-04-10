package installer

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"openriot/config"
)

// CopyConfigs copies configuration files from the repo to user's home directory.
// It reads config rules from the loaded YAML configuration.
// If dryRun is true, only logs what would be copied without actually copying.
// Files listed in preserve_if_exists are skipped if they already exist at the destination.
func CopyConfigs(repoDir string, cfg *config.Config, dryRun bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configSourceDir := filepath.Join(repoDir, "config")
	configDir := filepath.Join(homeDir, ".config")

	fmt.Printf("%s[INFO]%s Deploying configuration files...\n", Blue, Reset)

	// Create ~/.config if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Track stats per category
	categoryStats := make(map[string]int)

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
			// Get the source directory by removing the /* from the pattern
			patternWithoutGlob := strings.TrimSuffix(rule.Pattern, "/*")
			srcDir := filepath.Join(configSourceDir, patternWithoutGlob)
			globDest := filepath.Join(configDir, rule.Pattern)

			// baseDest is the parent directory (e.g., ~/.config/i3 for pattern "i3/*")
			baseDest := filepath.Dir(globDest)
			if rule.Target != "" {
				if strings.HasPrefix(rule.Target, "~/") {
					baseDest = filepath.Join(homeDir, rule.Target[2:])
				} else {
					baseDest = rule.Target
				}
			}

			// Check if source directory exists
			if _, err := os.Stat(srcDir); os.IsNotExist(err) {
				continue
			}

			// Get category name for stats
			category := patternWithoutGlob

			// Walk the source directory recursively
			err := filepath.WalkDir(srcDir, func(srcPath string, info fs.DirEntry, err error) error {
				if err != nil {
					return nil // skip inaccessible entries
				}
				if info.IsDir() {
					return nil // recurse into directories
				}

				// Get relative path from the pattern directory (not configSourceDir)
				relPath, err := filepath.Rel(srcDir, srcPath)
				if err != nil {
					return nil
				}

				// Compute destination path: baseDest + relative path
				destPath := filepath.Join(baseDest, relPath)

				// Check preservation
				filename := filepath.Base(destPath)
				if shouldPreserve(filename, rule.PreserveIfExists, destPath) {
					return nil
				}

				// Create destination directory
				destDir := filepath.Dir(destPath)
				if err := os.MkdirAll(destDir, 0755); err != nil {
					fmt.Printf("%s[WARN]%s Failed to create directory %s: %v\n", Yellow, Reset, destDir, err)
					return nil
				}

				// Copy file
				if dryRun {
					fmt.Printf("%s[INFO]%s [DRY-RUN] Would copy %s -> %s\n", Blue, Reset, relPath, destPath)
				} else if err := copyFilePreserve(srcPath, destPath); err != nil {
					fmt.Printf("%s[WARN]%s Failed to copy %s: %v\n", Yellow, Reset, srcPath, err)
					return nil
				} else {
					categoryStats[category]++
				}
				return nil
			})
			if err != nil {
				fmt.Printf("%s[WARN]%s WalkDir failed for %s: %v\n", Yellow, Reset, rule.Pattern, err)
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
				continue
			}

			// Check preservation
			filename := filepath.Base(destPath)
			if shouldPreserve(filename, rule.PreserveIfExists, destPath) {
				continue
			}

			// Create destination directory
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				fmt.Printf("%s[WARN]%s Failed to create directory %s: %v\n", Yellow, Reset, destDir, err)
				continue
			}

			// Copy file
			if dryRun {
				fmt.Printf("%s[INFO]%s [DRY-RUN] Would copy %s -> %s\n", Blue, Reset, rule.Pattern, destPath)
			} else if err := copyFilePreserve(srcPath, destPath); err != nil {
				fmt.Printf("%s[WARN]%s Failed to copy %s: %v\n", Yellow, Reset, rule.Pattern, err)
				continue
			} else {
				// Get category from pattern (e.g., "i3/config" -> "i3")
				parts := strings.Split(rule.Pattern, "/")
				category := parts[0]
				categoryStats[category]++
			}
		}
	}

	// Print summary
	for category, count := range categoryStats {
		fmt.Printf("%s[DONE]%s %s: %d files\n", Green, Reset, category, count)
	}
	fmt.Printf("%s[DONE]%s Configuration deployed\n", Green, Reset)

	return nil
}

// shouldPreserve checks if a file should be preserved based on the preserve list
func shouldPreserve(filename string, preserveList []string, destPath string) bool {
	for _, preserve := range preserveList {
		if preserve == filename {
			// File is in preserve list - check if it exists at destination
			if _, err := os.Stat(destPath); err == nil {
				return true
			}
		}
	}
	return false
}

// copyFilePreserve copies a single file, preserving source permissions and skipping identical content
func copyFilePreserve(source, dest string) error {
	sourceData, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	// Skip write when content is identical to prevent spurious reloads
	if existing, err := os.ReadFile(dest); err == nil {
		if bytes.Equal(existing, sourceData) {
			return nil
		}
	}

	// Preserve source file permissions
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

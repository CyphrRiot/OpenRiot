package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"openriot/config"
	"openriot/fsutil"
	"openriot/logger"
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

	logger.Info("Deploying configuration files...")

	// Create ~/.config if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Track stats per category
	categoryStats := make(map[string]int)

	// Collect all config rules from all modules (dependency-resolved order)
	refs, err := cfg.GetAllModulesOrdered()
	if err != nil {
		return fmt.Errorf("resolving module dependencies: %w", err)
	}

	var allRules []config.ConfigRule
	for _, ref := range refs {
		allRules = append(allRules, ref.Module.Configs...)
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
					logger.Warn(fmt.Sprintf("Failed to create directory %s: %v", destDir, err))
					return nil
				}

				// Copy file
				if dryRun {
					logger.Info(fmt.Sprintf("[DRY-RUN] Would copy %s -> %s", relPath, destPath))
				} else if err := fsutil.CopyFile(srcPath, destPath); err != nil {
					logger.Warn(fmt.Sprintf("Failed to copy %s: %v", srcPath, err))
					return nil
				} else {
					categoryStats[category]++
				}
				return nil
			})
			if err != nil {
				logger.Warn(fmt.Sprintf("WalkDir failed for %s: %v", rule.Pattern, err))
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

			// Check if destination exists and is a directory - remove it so file can be written
			if destInfo, err := os.Stat(destPath); err == nil && destInfo.IsDir() {
				if err := os.RemoveAll(destPath); err != nil {
					logger.Warn(fmt.Sprintf("Failed to remove directory %s: %v", destPath, err))
					continue
				}
			}

			// Create destination directory
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				logger.Warn(fmt.Sprintf("Failed to create directory %s: %v", destDir, err))
				continue
			}

			// Copy file
			if dryRun {
				logger.Info(fmt.Sprintf("[DRY-RUN] Would copy %s -> %s", rule.Pattern, destPath))
			} else if err := fsutil.CopyFile(srcPath, destPath); err != nil {
				logger.Warn(fmt.Sprintf("Failed to copy %s: %v", rule.Pattern, err))
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
		logger.Done(fmt.Sprintf("%s: %d files", category, count))
	}
	logger.Done("Configuration deployed")

	return nil
}

// isPreserveFile returns true if filename is in the preserve list
func isPreserveFile(filename string, preserveList []string) bool {
	return slices.Contains(preserveList, filename)
}

// shouldPreserve checks if a file should be preserved based on the preserve list
// If filename is in preserve list AND file exists at destination → skip (preserve user file)
// If filename is in preserve list but file doesn't exist → copy (install default)
func shouldPreserve(filename string, preserveList []string, destPath string) bool {
	if !isPreserveFile(filename, preserveList) {
		return false
	}
	// File is in preserve list - only skip if it already exists
	if _, err := os.Stat(destPath); err == nil {
		return true
	}
	return false
}



package installer

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"openriot/config"
	"openriot/fsutil"
	"openriot/logger"
	"openriot/paths"
	"openriot/theme"
)

// CopyConfigs copies configuration files from the repo to user's home directory.
// It reads config rules from the loaded YAML configuration.
// If dryRun is true, only logs what would be copied without actually copying.
// Files listed in preserve_if_exists are skipped if they already exist at the destination.
func CopyConfigs(repoDir string, cfg *config.Config, dryRun bool) error {
	homeDir := paths.HomeDir()

	configSourceDir := filepath.Join(repoDir, "config")
	configDir := filepath.Join(homeDir, ".config")

	logger.Info("Deploying configuration files...")

	// Create ~/.config if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Track directories already created to avoid redundant syscalls
	createdDirs := make(map[string]bool)
	mkdirOnce := func(dir string) error {
		if createdDirs[dir] {
			return nil
		}
		err := os.MkdirAll(dir, 0755)
		if err == nil {
			createdDirs[dir] = true
		}
		return err
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
				if err := mkdirOnce(destDir); err != nil {
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
			if err := mkdirOnce(destDir); err != nil {
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

// renderColorsMap flattens a theme.ColorPalette into a map usable by
// text/template with keys like BaseBG, AccentFG, SemanticError, etc.
func renderColorsMap() (map[string]string, error) {
	homeDir := paths.HomeDir()
	colorsPath := filepath.Join(homeDir, ".local", "share",
		"openriot", "config", "colors.toml")

	p, err := theme.LoadColors(colorsPath)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"BaseBG":          p.Base.BG,
		"BaseBG2":         p.Base.BG2,
		"BaseFG":          p.Base.FG,
		"BaseFG2":         p.Base.FG2,
		"BaseFG3":         p.Base.FG3,
		"BaseDim":         p.Base.Dim,
		"AccentName":      p.Accent.Name,
		"AccentFG":        p.Accent.FG,
		"AccentFGLight":   p.Accent.FGLight,
		"AccentBG":        p.Accent.BG,
		"SemanticError":   p.Semantic.Error,
		"SemanticWarning": p.Semantic.Warning,
		"SemanticSuccess": p.Semantic.Success,
		"SemanticInfo":    p.Semantic.Info,
		"SemanticCyan":    p.Semantic.Cyan,
		"CompatGreen":     p.Compat.Green,
		"CompatViolet":    p.Compat.Violet,
		"CompatBlue":      p.Compat.Blue,
		"CompatDimGray":   p.Compat.DimGray,
		"CompatWhite":     p.Compat.White,
		"ExtendedTeal":       p.Extended.Teal,
		"ExtendedSky":        p.Extended.Sky,
		"ExtendedElectric":   p.Extended.Electric,
		"ExtendedPurple":     p.Extended.Purple,
		"ExtendedViolet":     p.Extended.Violet,
		"ExtendedOrange":     p.Extended.Orange,
		"ExtendedCyanDim":    p.Extended.CyanDim,
		"ExtendedLauncherFG": p.Extended.LauncherFG,
		"ExtendedSepBG":      p.Extended.SepBG,
		"ExtendedBGDark":     p.Extended.BGDark,
		"ExtendedBGMid":      p.Extended.BGMid,
		"ExtendedFGBright":   p.Extended.FGBright,
		"ExtendedMuted":      p.Extended.Muted,
		"ExtendedSecYellow":  p.Extended.SecYellow,
		"ExtendedSecOrange":  p.Extended.SecOrange,
		"ExtendedAlphaBG":    p.Extended.AlphaBG,
	}, nil
}

// RenderTemplateString reads a .tmpl file and returns the rendered
// content as a string together with the color map used. Callers can
// apply further string replacements before writing to disk.
func RenderTemplateString(srcTmpl string) (string, map[string]string,
	error) {
	colors, err := renderColorsMap()
	if err != nil {
		return "", nil,
			fmt.Errorf("loading colors for template %s: %w",
				srcTmpl, err)
	}

	data, err := os.ReadFile(srcTmpl)
	if err != nil {
		return "", nil,
			fmt.Errorf("reading template %s: %w", srcTmpl, err)
	}

	tmpl, err := template.New(filepath.Base(srcTmpl)).Parse(
		string(data))
	if err != nil {
		return "", nil,
			fmt.Errorf("parsing template %s: %w", srcTmpl, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, colors); err != nil {
		return "", nil,
			fmt.Errorf("executing template %s: %w", srcTmpl, err)
	}

	return buf.String(), colors, nil
}

// RenderTemplate reads a .tmpl file, executes it with the canonical
// palette from colors.toml, and writes the result to dst. An empty
// colors map is returned so callers can extend it with extra
// variables if needed.
func RenderTemplate(srcTmpl, dst string) (map[string]string, error) {
	content, colors, err := RenderTemplateString(srcTmpl)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return nil, fmt.Errorf("creating directory for %s: %w",
			dst, err)
	}

	if err := os.WriteFile(dst, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("writing rendered file %s: %w",
			dst, err)
	}

	return colors, nil
}



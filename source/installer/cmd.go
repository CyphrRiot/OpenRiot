package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/config"
	"openriot/logger"
)

// RunSourceBuilds runs only the source builds phase (used by setup.sh)
func RunSourceBuilds(testMode bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := SourceBuilds(cfg, testMode); err != nil {
		logger.Warn(fmt.Sprintf("Source builds: %v", err))
	}
	logger.Info("Source builds complete!")
}

// RunInstallPackages installs packages from packages.yaml (used by setup.sh)
func RunInstallPackages() {
	// Setup fastest mirror if not already configured
	SetupMirror()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Installing packages (safe one-by-one mode)...")

	packages := cfg.GetPackages()
	if len(packages) == 0 {
		logger.Fail("No packages found in packages.yaml")
		os.Exit(1)
	}

	failed, _ := InstallPackages(cfg, packages)
	if failed > 0 {
		os.Exit(1)
	}

	logger.Done("Package installation complete.")
}

// findPackagesYaml finds packages.yaml: installed location first, then CWD fallback
func findPackagesYaml() string {
	homeDir, _ := os.UserHomeDir()

	// 1. Check installed location first
	installedPath := filepath.Join(homeDir, ".local", "share", "openriot", "install", "packages.yaml")
	if _, err := os.Stat(installedPath); err == nil {
		return installedPath
	}

	// 2. Fallback to CWD (for development)
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "install", "packages.yaml")
}

// RunCheckPackages verifies packages.yaml versions against installed
func RunCheckPackages() {
	configPath := findPackagesYaml()
	logger.Info(fmt.Sprintf("Checking: %s", configPath))

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Get installed packages from pkg_info -a
	installed := GetInstalledPackages()
	if len(installed) == 0 {
		fmt.Fprintf(os.Stderr, "[FAIL] No packages found from pkg_info\n")
		os.Exit(1)
	}

	// Check each yaml package against installed
	yamlPkgs := cfg.GetPackages()
	mismatches := 0

	for _, pkg := range yamlPkgs {
		base := config.GetBaseName(pkg)
		installedVer, exists := installed[base]
		if !exists {
			logger.Info(fmt.Sprintf("[MISS] %s (not installed)", pkg))
			mismatches++
		} else if installedVer != pkg {
			logger.Warn(fmt.Sprintf("%s -> %s", pkg, installedVer))
			mismatches++
		}
	}

	if mismatches > 0 {
		logger.Warn(fmt.Sprintf("%d package version mismatches found", mismatches))
		logger.Info("Run 'openriot --sync-packages' to update packages.yaml")
		os.Exit(1)
	}

	logger.Done("All packages in sync")
	os.Exit(0)
}

// RunSyncPackages updates packages.yaml to latest installed versions
func RunSyncPackages() {
	configPath := findPackagesYaml()
	logger.Info(fmt.Sprintf("Updating: %s", configPath))

	// Get installed packages
	installed := GetInstalledPackages()
	if len(installed) == 0 {
		fmt.Fprintf(os.Stderr, "[FAIL] No packages found from pkg_info\n")
		os.Exit(1)
	}

	// Read yaml file as text to preserve formatting
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Fail(fmt.Sprintf("Failed to read config: %v", err))
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")
	updated := 0

	// Replace only matching package lines
	for i, line := range lines {
		// Find indentation (spaces before "- ")
		indent := ""
		for j, ch := range line {
			if ch == ' ' {
				indent += " "
			} else if ch == '-' && j+1 < len(line) && line[j+1] == ' ' {
				break
			} else {
				indent = ""
				break
			}
		}
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		pkg := strings.TrimPrefix(trimmed, "- ")
		base := config.GetBaseName(pkg)
		if installedVer, exists := installed[base]; exists && installedVer != pkg {
			lines[i] = indent + "- " + installedVer
			updated++
		}
	}

	// Write back (preserves formatting)
	if err := os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Failed to save config: %v\n", err)
		os.Exit(1)
	}

	logger.Done(fmt.Sprintf("Updated %d packages in packages.yaml", updated))
	os.Exit(0)
}

// GetInstalledPackages returns map of base name -> full package version
func GetInstalledPackages() map[string]string {
	cmd := exec.Command("pkg_info", "-a")
	output, err := cmd.Output()
	if err != nil {
		// Return empty map instead of nil to distinguish from "no packages installed"
		// Callers check len() == 0 which works for both, but we lose the error context
		// Use empty map so downstream len() check works consistently
		return make(map[string]string)
	}

	packages := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: package-version description
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			pkg := fields[0]
			// Extract base name from package-version
			if idx := strings.LastIndex(pkg, "-"); idx > 0 {
				base := pkg[:idx]
				packages[base] = pkg
			}
		}
	}
	return packages
}

// RunMirrors tests mirror connectivity and shows results
func RunMirrors() {
	logger.Info("Testing mirrors...")

	mirror, latency, err := SelectFastestMirror()
	if err != nil {
		logger.Fail(fmt.Sprintf("No mirrors responded: %v", err))
		os.Exit(1)
	}

	logger.Done(fmt.Sprintf("Fastest mirror: %s (%dms)", mirror, latency.Milliseconds()))

	if HasInstallurl() {
		current := GetInstallurl()
		logger.Info(fmt.Sprintf("Current /etc/installurl: %s", current))
		if current == mirror {
			logger.Info("Already using fastest mirror")
		} else {
			logger.Warn("Different from detected fastest. Run as root to update.")
		}
	} else {
		logger.Info("No /etc/installurl found. Run as root to create.")
	}
}

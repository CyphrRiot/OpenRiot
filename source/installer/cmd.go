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

// RunSourceBuilds runs only the source builds phase (used by setup.sh).
// Returns an error on failure instead of calling os.Exit, making it testable.
func RunSourceBuilds(testMode bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := SourceBuilds(cfg, testMode); err != nil {
		logger.Warn(fmt.Sprintf("Source builds: %v", err))
	}
	logger.Info("Source builds complete!")
	return nil
}

// RunInstallPackages installs packages from packages.yaml (used by setup.sh).
// Returns an error on failure instead of calling os.Exit, making it testable.
func RunInstallPackages() error {
	// Setup fastest mirror if not already configured
	SetupMirror()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	repoDir := filepath.Join(homeDir, ".local", "share", "openriot")
	configPath := filepath.Join(repoDir, "install", "packages.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check games preference before package installation
	if !GamesPreference() {
		delete(cfg.Desktop, "games")
	}

	logger.Info("Installing packages (safe one-by-one mode)...")

	packages := cfg.GetPackages()
	if len(packages) == 0 {
		return fmt.Errorf("no packages found in packages.yaml")
	}

	failed, _ := InstallPackages(cfg, packages)
	if failed > 0 {
		return fmt.Errorf("%d package(s) failed to install", failed)
	}

	logger.Done("Package installation complete.")
	return nil
}

// findPackagesYaml finds packages.yaml: installed location first, then CWD fallback.
func findPackagesYaml() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// 1. Check installed location first
	installedPath := filepath.Join(homeDir, ".local", "share", "openriot", "install", "packages.yaml")
	if _, err := os.Stat(installedPath); err == nil {
		return installedPath
	}

	// 2. Fallback to CWD (for development)
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "install", "packages.yaml")
}

// RunCheckPackages verifies packages.yaml versions against installed.
// Returns an error if mismatches are found or on any internal failure.
func RunCheckPackages() error {
	configPath := findPackagesYaml()
	logger.Info(fmt.Sprintf("Checking: %s", configPath))

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	installed := GetInstalledPackages()
	if len(installed) == 0 {
		return fmt.Errorf("no packages found from pkg_info")
	}

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
		return fmt.Errorf("%d package version mismatches found", mismatches)
	}

	logger.Done("All packages in sync")
	return nil
}

// RunSyncPackages updates packages.yaml to latest installed versions.
// Returns an error on any failure.
func RunSyncPackages() error {
	configPath := findPackagesYaml()
	logger.Info(fmt.Sprintf("Updating: %s", configPath))

	installed := GetInstalledPackages()
	if len(installed) == 0 {
		return fmt.Errorf("no packages found from pkg_info")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	updated := 0

	for i, line := range lines {
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

	if err := os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	logger.Done(fmt.Sprintf("Updated %d packages in packages.yaml", updated))
	return nil
}

// GetInstalledPackages returns map of base name -> full package version.
func GetInstalledPackages() map[string]string {
	cmd := exec.Command("pkg_info", "-a")
	output, err := cmd.Output()
	if err != nil {
		return make(map[string]string)
	}

	packages := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			pkg := fields[0]
			if idx := strings.LastIndex(pkg, "-"); idx > 0 {
				base := pkg[:idx]
				packages[base] = pkg
			}
		}
	}
	return packages
}

// RunMirrors tests mirror connectivity and shows results.
// Returns an error on failure.
func RunMirrors() error {
	logger.Info("Testing mirrors...")

	mirror, latency, err := SelectFastestMirror()
	if err != nil {
		return fmt.Errorf("no mirrors responded: %w", err)
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
	return nil
}

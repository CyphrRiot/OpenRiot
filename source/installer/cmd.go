package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/config"
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
		fmt.Printf("[WARN] Source builds: %v\n", err)
	}
	fmt.Println("[INFO] Source builds complete!")
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

	fmt.Printf("%s[INFO]%s Installing packages from packages.yaml (safe one-by-one mode)...\n", Cyan, Reset)

	packages := cfg.GetPackages()
	if len(packages) == 0 {
		fmt.Fprintf(os.Stderr, "%s[FAIL]%s No packages found in packages.yaml\n", Red, Reset)
		os.Exit(1)
	}

	failed, _ := InstallPackages(cfg, packages)
	if failed > 0 {
		os.Exit(1)
	}

	fmt.Printf("%s[INFO]%s Updating all packages and dependencies to latest...\n", Cyan, Reset)
	updateCmd := []string{"doas", "pkg_add", "-u"}
	if cfg.IsSnapshot() {
		updateCmd = append(updateCmd, "-D", "snapshot")
	}
	cmd := exec.Command(updateCmd[0], updateCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[WARN]%s Some packages failed to update: %v\n", Yellow, Reset, err)
	}
	fmt.Printf("%s[DONE]%s Package update complete.\n", Green, Reset)
}

// findPackagesYaml finds packages.yaml: CWD first, then installed location
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
	fmt.Printf("[INFO] Checking: %s\n", configPath)

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
		base := GetBaseName(pkg)
		installedVer, exists := installed[base]
		if !exists {
			fmt.Printf("[MISS] %s (not installed)\n", pkg)
			mismatches++
		} else if installedVer != pkg {
			fmt.Printf("[WARN] %s -> %s\n", pkg, installedVer)
			mismatches++
		}
	}

	if mismatches > 0 {
		fmt.Printf("\n[WARN] %d package version mismatches found\n", mismatches)
		fmt.Printf("[INFO] Run 'openriot --sync-packages' to update packages.yaml\n")
		os.Exit(1)
	}

	fmt.Println("[DONE] All packages in sync")
	os.Exit(0)
}

// RunSyncPackages updates packages.yaml to latest installed versions
func RunSyncPackages() {
	configPath := findPackagesYaml()
	fmt.Printf("[INFO] Updating: %s\n", configPath)

	// Get installed packages
	installed := GetInstalledPackages()
	if len(installed) == 0 {
		fmt.Fprintf(os.Stderr, "[FAIL] No packages found from pkg_info\n")
		os.Exit(1)
	}

	// Read yaml file as text to preserve formatting
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Failed to read config: %v\n", err)
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
		base := GetBaseName(pkg)
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

	fmt.Printf("[DONE] Updated %d packages in packages.yaml\n", updated)
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

// GetBaseName extracts base name from package (e.g., "fish-4.6.0" -> "fish")
func GetBaseName(pkg string) string {
	if idx := strings.LastIndex(pkg, "-"); idx > 0 {
		return pkg[:idx]
	}
	return pkg
}

// RunMirrors tests mirror connectivity and shows results
func RunMirrors() {
	fmt.Printf("%s[INFO]%s Testing mirrors...\n", Cyan, Reset)

	mirror, latency, err := SelectFastestMirror()
	if err != nil {
		fmt.Printf("%s[FAIL]%s No mirrors responded: %v\n", Red, Reset, err)
		os.Exit(1)
	}

	fmt.Printf("%s[DONE]%s Fastest mirror: %s (%dms)\n", Green, Reset, mirror, latency.Milliseconds())

	if HasInstallurl() {
		current := GetInstallurl()
		fmt.Printf("[INFO] Current /etc/installurl: %s\n", current)
		if current == mirror {
			fmt.Printf("[INFO] Already using fastest mirror\n")
		} else {
			fmt.Printf("[WARN] Different from detected fastest. Run as root to update.\n")
		}
	} else {
		fmt.Printf("[INFO] No /etc/installurl found. Run as root to create.\n")
	}
}

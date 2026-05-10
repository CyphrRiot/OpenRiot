package imaging

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"openriot/config"
	"openriot/logger"
)

// GetPackageList returns list of packages from config
func GetPackageList() ([]string, error) {
	cfgPath := config.FindConfigFile()
	if cfgPath == "" {
		return nil, fmt.Errorf("could not find packages.yaml")
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg.GetPackages(), nil
}

// Exceptions holds package and module exclusions from exceptions.yaml
type Exceptions struct {
	Packages map[string]bool // package base names to skip
	Modules  map[string]bool // module IDs (e.g. "desktop.games") to skip
}

// LoadExceptions loads excluded packages and modules from Build/exceptions.yaml
func LoadExceptions() (*Exceptions, error) {
	e := &Exceptions{
		Packages: make(map[string]bool),
		Modules:  make(map[string]bool),
	}

	execDir, err := os.Executable()
	if err != nil {
		return e, nil // Return empty, not an error
	}
	repoRoot := filepath.Dir(filepath.Dir(execDir))
	exceptionsPath := filepath.Join(repoRoot, "Build", "exceptions.yaml")

	data, err := os.ReadFile(exceptionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return e, nil // No exceptions file, that's fine
		}
		return nil, err
	}

	var section string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasSuffix(trimmed, ":") {
			section = strings.TrimSuffix(trimmed, ":")
			continue
		}
		if after, ok := strings.CutPrefix(trimmed, "- "); ok {
			item := strings.Trim(after, "\"")
			switch section {
			case "exclude":
				e.Packages[item] = true
			case "modules":
				e.Modules[item] = true
			}
		}
	}

	return e, nil
}

// DownloadPackages downloads all required packages to work dir
func DownloadPackages(cfg *Config) error {
	// Load config
	cfgPath := config.FindConfigFile()
	if cfgPath == "" {
		return fmt.Errorf("could not find packages.yaml")
	}
	cfgFile, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load exceptions
	exceptions, err := LoadExceptions()
	if err != nil {
		return fmt.Errorf("failed to load exceptions: %w", err)
	}

	// Get filtered package list
	packages := cfgFile.GetPackagesExcluding(exceptions.Packages, exceptions.Modules)
	if len(packages) == 0 {
		return fmt.Errorf("no packages to download")
	}

	// Create package directory
	pkgDir := filepath.Join(cfg.WorkDir, "packages", "snapshots", "amd64")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return fmt.Errorf("failed to create package dir: %w", err)
	}

	// Clean stale packages first
	CleanStalePackages(pkgDir, packages, exceptions.Packages)

	// Download each package
	pkgCount := len(packages)
	for i, pkg := range packages {
		displayName := strings.TrimSuffix(pkg, ".tgz")
		displayPadded := fmt.Sprintf("%-35s", displayName)
		fmt.Printf("\r%s[INFO]%s Downloading package %d/%d: %s", logger.Cyan, logger.Reset, i+1, pkgCount, displayPadded)

		pkgPath := filepath.Join(pkgDir, pkg+".tgz")

		// Skip if already exists
		if _, err := os.Stat(pkgPath); err == nil {
			continue
		}

		// Remove old versions of this package
		removeOldVersions(pkgDir, pkg)

		// Download with retry
		err := downloadWithRetry(pkgPath, pkg)
		if err != nil {
			fmt.Println() // newline so WARN doesn't append to the progress line
			logger.Warn(fmt.Sprintf("Failed to download: %s", pkg))
		}
	}

	fmt.Println() // Ensure clean line before next output
	return nil
}

// CleanStalePackages removes packages not in current list or in exceptions
func CleanStalePackages(pkgDir string, currentList []string, exceptions map[string]bool) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return // No directory yet
	}

	// Build set of current packages (without version)
	currentSet := make(map[string]bool)
	for _, pkg := range currentList {
		currentSet[config.GetBaseName(pkg)] = true
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".tgz") {
			continue
		}

		basePkg := config.GetBaseName(strings.TrimSuffix(entry.Name(), ".tgz"))

		// Remove if not in current list OR in exceptions
		if !currentSet[basePkg] || exceptions[basePkg] {
			os.Remove(filepath.Join(pkgDir, entry.Name()))
		}
	}
}

// removeOldVersions removes old versions of a package
func removeOldVersions(pkgDir, pkgName string) {
	basePkg := config.GetBaseName(pkgName)
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".tgz") {
			continue
		}
		// Remove files like go-1.24.4.tgz when downloading go-1.24.5
		if strings.HasPrefix(entry.Name(), basePkg+"-") {
			os.Remove(filepath.Join(pkgDir, entry.Name()))
		}
	}
}

// downloadWithRetry downloads a package with 3 retries
func downloadWithRetry(pkgPath, pkgName string) error {
	url := fmt.Sprintf("https://cdn.openbsd.org/pub/OpenBSD/snapshots/packages/amd64/%s.tgz", pkgName)

	var lastErr error
	for retry := 0; retry < 3; retry++ {
		lastErr = downloadFile(pkgPath, url)
		if lastErr == nil {
			return nil
		}
	}
	return fmt.Errorf("download failed after 3 attempts: %w", lastErr)
}

// downloadFile downloads a single file
func downloadFile(pkgPath, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Create temp file
	tmpPath := pkgPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	f.Close()
	return os.Rename(tmpPath, pkgPath)
}
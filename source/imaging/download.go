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

// LoadExceptions loads excluded packages from Build/exceptions.yaml
func LoadExceptions() (map[string]bool, error) {
	exceptions := make(map[string]bool)

	// Try to find exceptions.yaml relative to the script dir
	execDir, err := os.Executable()
	if err != nil {
		return exceptions, nil // Return empty, not an error
	}
	repoRoot := filepath.Dir(filepath.Dir(execDir))
	exceptionsPath := filepath.Join(repoRoot, "Build", "exceptions.yaml")

	data, err := os.ReadFile(exceptionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return exceptions, nil // No exceptions file, that's fine
		}
		return nil, err
	}

	// Parse: look for lines starting with "  - "
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "- "); ok {
			pkg := strings.Trim(after, "\"") // Remove quotes if present
			exceptions[pkg] = true
		}
	}

	return exceptions, nil
}

// DownloadPackages downloads all required packages to work dir
func DownloadPackages(cfg *Config) error {
	// Get package list from config
	packages, err := GetPackageList()
	if err != nil {
		return fmt.Errorf("failed to get package list: %w", err)
	}

	if len(packages) == 0 {
		return fmt.Errorf("no packages to download")
	}

	// Load exceptions
	exceptions, err := LoadExceptions()
	if err != nil {
		return fmt.Errorf("failed to load exceptions: %w", err)
	}

	// Create package directory
	pkgDir := filepath.Join(cfg.WorkDir, "packages", "snapshots", "amd64")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return fmt.Errorf("failed to create package dir: %w", err)
	}

	// Clean stale packages first
	CleanStalePackages(pkgDir, packages, exceptions)

	// Filter packages by exceptions
	var toDownload []string
	for _, pkg := range packages {
		basePkg := config.GetBaseName(pkg)
		if exceptions[basePkg] {
			continue // Skip excluded packages
		}
		toDownload = append(toDownload, pkg)
	}

	// Download each package
	pkgCount := len(toDownload)
	for i, pkg := range toDownload {
		displayName := strings.TrimSuffix(pkg, ".tgz")
		// Pad to 35 chars to clear any previous longer name
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
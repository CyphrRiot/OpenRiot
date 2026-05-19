package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"openriot/logger"
)

// Config holds imaging configuration
type Config struct {
	BaseImg   string // Path to base OpenBSD image
	OutputImg string // Path for output image
	WorkDir   string // Working directory
	Version   string // OpenBSD version (e.g., "79")
	NoBurn    bool   // Skip interactive burn prompt

	// Computed paths
	OpenriotBin string // Path to openriot binary
	OpenriotTgz string // Path to output tarball

	// Image size (calculated after content injection)
	UsedKB int // Used space in KB after injection
}

// LoadConfig loads imaging config from flags/env
func LoadConfig(args []string) (*Config, error) {
	cfg := &Config{
		BaseImg:   os.Getenv("BASE_IMG"),
		OutputImg: os.Getenv("OUTPUT_IMG"),
		WorkDir:   os.Getenv("WORK_DIR"),
		Version:   "79",
		NoBurn:    false,
	}

	// Default paths
	if cfg.BaseImg == "" {
		cfg.BaseImg = "Build/Images/install79.img"
	}
	if cfg.OutputImg == "" {
		cfg.OutputImg = "Build/Output/openriot.img"
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir = "Build/work"
	}

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--base-img":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--base-img requires a path")
			}
			cfg.BaseImg = args[i+1]
			i++
		case "--output-img":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output-img requires a path")
			}
			cfg.OutputImg = args[i+1]
			i++
		case "--work-dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--work-dir requires a path")
			}
			cfg.WorkDir = args[i+1]
			i++
		case "--version":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--version requires a version number")
			}
			cfg.Version = args[i+1]
			i++
		case "--no-burn":
			cfg.NoBurn = true
		}
	}

	// Compute derived paths
	absWorkDir, err := filepath.Abs(cfg.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve work dir: %w", err)
	}
	cfg.WorkDir = absWorkDir
	outputDir := filepath.Dir(cfg.OutputImg)
	cfg.OpenriotTgz = filepath.Join(outputDir, fmt.Sprintf("site%s.tgz", cfg.Version))

	// Binary is always at install/openriot relative to repo root
	execDir, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to find executable: %w", err)
	}
	repoRoot := filepath.Dir(filepath.Dir(execDir))
	cfg.OpenriotBin = filepath.Join(repoRoot, "install", "openriot")

	return cfg, nil
}

// IsOpenBSD returns true if running on OpenBSD
func IsOpenBSD() bool {
	cmd := exec.Command("uname", "-s")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "OpenBSD"
}

// IsRoot returns true if running as root (uid 0)
func IsRoot() bool {
	cmd := exec.Command("id", "-u")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	uid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	return err == nil && uid == 0
}

// MustRunAsRoot returns error if not running as root
func MustRunAsRoot() error {
	if !IsRoot() {
		return fmt.Errorf("this operation must be run as root (use doas)")
	}
	return nil
}

// CheckPrereqs validates environment before imaging operations
func CheckPrereqs(cfg *Config) error {
	// Must be running on OpenBSD
	if !IsOpenBSD() {
		return fmt.Errorf("this command must be run on OpenBSD")
	}

	// Check base image at configured path (must be non-empty)
	if info, err := os.Stat(cfg.BaseImg); err == nil && info.Size() > 0 {
		return nil
	}

	// Check Build/Images/install79.img relative to repo root
	execDir, _ := os.Executable()
	repoRoot := filepath.Dir(filepath.Dir(execDir))
	fallbackImg := filepath.Join(repoRoot, "Build", "Images", "install79.img")
	if info, err := os.Stat(fallbackImg); err == nil && info.Size() > 0 {
		cfg.BaseImg = fallbackImg
		return nil
	}

	// Check Build/Images/install79.iso and warn
	fallbackIso := filepath.Join(repoRoot, "Build", "Images", "install79.iso")
	if _, err := os.Stat(fallbackIso); err == nil {
		return fmt.Errorf("found %s but image builder requires .img (disk image), not .iso\ndownload the .img from https://cdn.openbsd.org/pub/OpenBSD/%s/amd64/install79.img", fallbackIso, formatVersion(cfg.Version))
	}

	// Download base image
	if err := downloadBaseImage(cfg); err != nil {
		return fmt.Errorf("base image missing and download failed: %w", err)
	}

	return nil
}

// downloadBaseImage fetches install79.img from OpenBSD CDN
func downloadBaseImage(cfg *Config) error {
	dir := filepath.Dir(cfg.BaseImg)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create image dir: %w", err)
	}

	// Remove any stale empty or partial file
	os.Remove(cfg.BaseImg)

	// Default to release CDN; -current users can set BASE_IMG or --base-img
	repoPath := formatVersion(cfg.Version)
	url := fmt.Sprintf("https://cdn.openbsd.org/pub/OpenBSD/%s/amd64/install%s.img", repoPath, cfg.Version)

	logger.Info(fmt.Sprintf("Downloading base image from %s...", url))

	cmd := exec.Command("wget", "-O", cfg.BaseImg, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Remove(cfg.BaseImg)
		return fmt.Errorf("wget failed: %w", err)
	}

	// Verify download is non-empty
	if info, err := os.Stat(cfg.BaseImg); err != nil || info.Size() == 0 {
		os.Remove(cfg.BaseImg)
		return fmt.Errorf("downloaded image is empty")
	}

	logger.Info("Base image downloaded")
	return nil
}
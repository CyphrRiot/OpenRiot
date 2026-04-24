package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
		cfg.OutputImg = "Build/Images/openriot.img"
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
	cfg.OpenriotTgz = filepath.Join(absWorkDir, "openriot.tgz")

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

	// Check base image exists
	if _, err := os.Stat(cfg.BaseImg); os.IsNotExist(err) {
		return fmt.Errorf("base image not found: %s\nlink your image: ln -sf ~/Code/Images/install79.img %s",
			cfg.BaseImg, cfg.BaseImg)
	}

	return nil
}
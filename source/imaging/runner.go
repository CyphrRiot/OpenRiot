package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"openriot/logger"
)

// Mode represents the imaging mode
type Mode string

const (
	ModeFull  Mode = "full"  // Create tarball + image
	ModeSite  Mode = "site"  // Create tarball only
	ModeClean Mode = "clean" // Clean work directory
)

// RunMakeImage is the main entry point for --make-image
func RunMakeImage(args []string) {
	// Parse mode from args
	mode := ModeFull
	for i, arg := range args {
		switch arg {
		case "site":
			mode = ModeSite
			args = append(args[:i], args[i+1:]...)
		case "clean":
			mode = ModeClean
			args = append(args[:i], args[i+1:]...)
		case "help", "--help", "-h":
			printMakeImageHelp()
			return
		}
	}

	// Load config
	cfg, err := LoadConfig(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Failed to load config: %v\n", logger.Red, logger.Reset, err)
		os.Exit(1)
	}

	// Print header
	version := GetOpenriotVersion()
	logger.Info(fmt.Sprintf("OpenRiot Image Builder v%s", version))
	logger.Info(fmt.Sprintf("Building for OpenBSD %s", cfg.Version))

	// Check prerequisites
	if err := CheckPrereqs(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s %v\n", logger.Red, logger.Reset, err)
		os.Exit(1)
	}

	// Run the requested mode
	switch mode {
	case ModeFull:
		runFullBuild(cfg)
	case ModeSite:
		runSiteOnly(cfg)
	case ModeClean:
		runClean(cfg)
	}
}

func runFullBuild(cfg *Config) {
	logger.Info("Mode: Full build (site + image)")

	// Check root for imaging operations
	if err := MustRunAsRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s %v\n", logger.Red, logger.Reset, err)
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Run with doas: doas %s --make-image\n", logger.Red, logger.Reset, cfg.OpenriotBin)
		os.Exit(1)
	}

	// Step 1: Download packages
	logger.Info("Downloading packages...")
	if err := DownloadPackages(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Download failed: %v\n", logger.Red, logger.Reset, err)
		os.Exit(1)
	}

	// Step 2: Create site tarball
	logger.Info("Creating site tarball...")
	if err := CreateSite(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Site creation failed: %v\n", logger.Red, logger.Reset, err)
		os.Exit(1)
	}

	// Step 3: Build image
	logger.Info("Building image...")
	if err := BuildImage(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Build failed: %v\n", logger.Red, logger.Reset, err)
		os.Exit(1)
	}

	// Step 4: Detect drives and offer burn
	logger.Info("Detecting drives...")
	drives, err := DetectDrives()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[WARN]%s Drive detection failed: %v\n", logger.Yellow, logger.Reset, err)
	} else {
		PromptBurn(cfg, drives)
	}

	logger.Done("Build complete!")
}

func runSiteOnly(cfg *Config) {
	logger.Info("Mode: Site only (tarball)")

	// Download packages
	logger.Info("Downloading packages...")
	if err := DownloadPackages(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Download failed: %v\n", logger.Red, logger.Reset, err)
		os.Exit(1)
	}

	// Create tarball
	logger.Info("Creating site tarball...")
	if err := CreateSite(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Site creation failed: %v\n", logger.Red, logger.Reset, err)
		os.Exit(1)
	}

	logger.Done(fmt.Sprintf("Site tarball created: %s", cfg.OpenriotTgz))
}

func runClean(cfg *Config) {
	logger.Info("Mode: Clean")

	if err := Cleanup(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Cleanup failed: %v\n", logger.Red, logger.Reset, err)
		os.Exit(1)
	}

	logger.Done("Cleanup complete")
}

func printMakeImageHelp() {
	fmt.Print(`OpenRiot Image Builder

Usage: openriot --make-image [mode] [flags]

Modes:
  (none)        Full build: create site tarball + image (default)
  site          Create openriot.tgz tarball only
  clean         Clean build artifacts

Flags:
  --base-img PATH    Base OpenBSD image (default: Build/Images/install79.img)
  --output-img PATH  Output image (default: Build/Images/openriot.img)
  --no-burn          Skip drive detection and burn prompt

Examples:
  openriot --make-image            # Full build
  openriot --make-image site       # Create tarball only
  openriot --make-image clean      # Clean artifacts
  openriot --make-image --no-burn  # Build without burning
`)
}

// HasModifications checks if there are uncommitted changes
func HasModifications() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}
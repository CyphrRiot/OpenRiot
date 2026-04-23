package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"openriot/installer"
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
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Failed to load config: %v\n", installer.Red, installer.Reset, err)
		os.Exit(1)
	}

	// Print header
	version := GetOpenriotVersion()
	log("OpenRiot Image Builder v%s", version)
	log("Building for OpenBSD %s", cfg.Version)

	// Check prerequisites
	if err := CheckPrereqs(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s %v\n", installer.Red, installer.Reset, err)
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
	log("Mode: Full build (site + image)")

	// Check root for imaging operations
	if err := MustRunAsRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s %v\n", installer.Red, installer.Reset, err)
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Run with doas: doas %s --make-image\n", installer.Red, installer.Reset, cfg.OpenriotBin)
		os.Exit(1)
	}

	// Step 1: Download packages
	log("Downloading packages...")
	if err := DownloadPackages(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Download failed: %v\n", installer.Red, installer.Reset, err)
		os.Exit(1)
	}

	// Step 2: Create site tarball
	log("Creating site tarball...")
	if err := CreateSite(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Site creation failed: %v\n", installer.Red, installer.Reset, err)
		os.Exit(1)
	}

	// Step 3: Build image
	log("Building image...")
	if err := BuildImage(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Build failed: %v\n", installer.Red, installer.Reset, err)
		os.Exit(1)
	}

	// Step 4: Detect drives and offer burn
	log("Detecting drives...")
	drives, err := DetectDrives()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[WARN]%s Drive detection failed: %v\n", installer.Yellow, installer.Reset, err)
	} else {
		PromptBurn(cfg, drives)
	}

	fmt.Printf("%s[DONE]%s Build complete!\n", installer.Green, installer.Reset)
}

func runSiteOnly(cfg *Config) {
	log("Mode: Site only (tarball)")

	// Download packages
	log("Downloading packages...")
	if err := DownloadPackages(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Download failed: %v\n", installer.Red, installer.Reset, err)
		os.Exit(1)
	}

	// Create tarball
	log("Creating site tarball...")
	if err := CreateSite(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Site creation failed: %v\n", installer.Red, installer.Reset, err)
		os.Exit(1)
	}

	fmt.Printf("%s[DONE]%s Site tarball created: %s\n", installer.Green, installer.Reset, cfg.OpenriotTgz)
}

func runClean(cfg *Config) {
	log("Mode: Clean")

	if err := Cleanup(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s Cleanup failed: %v\n", installer.Red, installer.Reset, err)
		os.Exit(1)
	}

	// Also clean repo cache
	repoCache := getBuildDir() + "/repo-cache"
	if err := os.RemoveAll(repoCache); err == nil {
		log("Repo cache cleaned")
	}

	fmt.Printf("%s[DONE]%s Cleanup complete\n", installer.Green, installer.Reset)
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
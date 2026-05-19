package imaging

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/logger"
)

// CreateSite creates the site79.tgz tarball
func CreateSite(cfg *Config) error {
	workDir := cfg.WorkDir

	// Create work directory
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}

	// Use a fresh temp directory to avoid permission collisions from prior root runs
	siteDir, err := os.MkdirTemp(workDir, "site-")
	if err != nil {
		return fmt.Errorf("create site temp dir: %w", err)
	}
	defer os.RemoveAll(siteDir)

	openriotDir := filepath.Join(siteDir, "openriot")
	if err := os.MkdirAll(openriotDir, 0755); err != nil {
		return fmt.Errorf("create openriot dir: %w", err)
	}

	// Copy MOTD
	if err := copyMotd(siteDir); err != nil {
		return fmt.Errorf("copy motd: %w", err)
	}

	// Copy packages into tarball so install.site can install them
	pkgSrc := filepath.Join(workDir, "packages", formatVersion(cfg.Version), "amd64")
	pkgDst := filepath.Join(openriotDir, "packages", formatVersion(cfg.Version), "amd64")
	if err := copyDir(pkgSrc, pkgDst); err != nil {
		return fmt.Errorf("copy packages: %w", err)
	}

	// Copy firmware into tarball for install.site to install
	fwSrc := filepath.Join(workDir, "firmware")
	if _, err := os.Stat(fwSrc); err == nil {
		fwDst := filepath.Join(openriotDir, "firmware")
		if err := copyDir(fwSrc, fwDst); err != nil {
			return fmt.Errorf("copy firmware: %w", err)
		}
	}

	// Create install.site inside siteDir so it ships in the tarball
	// (extracts to /install.site on the new system and runs via install.site(5))
	if err := createInstallSite(siteDir, formatVersion(cfg.Version)); err != nil {
		return fmt.Errorf("create install.site: %w", err)
	}

	// Create tarball (openriot/ directory, motd, and install.site)
	tgzPath := cfg.OpenriotTgz
	os.MkdirAll(filepath.Dir(tgzPath), 0755) // ensure output dir exists
	os.Remove(tgzPath) // Remove old if exists

	cmd := exec.Command("tar", "czf", tgzPath, "-C", siteDir, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tar failed: %w\n%s", err, out)
	}

	return nil
}

// getBuildDir returns the Build/ directory path
func getBuildDir() string {
	execDir, _ := os.Executable()
	repoRoot := filepath.Dir(filepath.Dir(execDir))
	return filepath.Join(repoRoot, "Build")
}

// copyDir copies a directory recursively from src to dst
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// copyMotd copies the motd file to site directory
func copyMotd(siteDir string) error {
	execDir, _ := os.Executable()
	repoRoot := filepath.Dir(filepath.Dir(execDir))
	motdSrc := filepath.Join(repoRoot, "install", "motd-install")

	if _, err := os.Stat(motdSrc); err != nil {
		return nil // No motd, that's fine
	}

	motdDstDir := filepath.Join(siteDir, "etc")
	// Create directory (may already exist)
	os.MkdirAll(motdDstDir, 0755)

	motdData, err := os.ReadFile(motdSrc)
	if err != nil {
		return err
	}

	// Write to motd file
	motdDst := filepath.Join(motdDstDir, "motd")
	err = os.WriteFile(motdDst, motdData, 0644)
	if err != nil {
		// Try fixing permissions and retry (some files may be root-owned from previous runs)
		os.Chmod(motdDstDir, 0755)
		os.Chmod(motdDst, 0644)
		err = os.WriteFile(motdDst, motdData, 0644)
		if err != nil {
			// Log but don't fail - motd is optional
			logger.Warn(fmt.Sprintf("Could not copy motd: %v", err))
		}
	}
	return nil
}

// createInstallSite writes the install.site script
func createInstallSite(siteDir, repoPath string) error {
	content := `#!/bin/sh
# OpenRiot post-install script — runs at end of OpenBSD install
# Environment: root, chroot at new system root, no TTY, no X

log() { echo "[OPENRIOT] $*"; }

log "OpenRiot post-install starting"

# ------------------------------------------------------------------
# 1. System configuration
# ------------------------------------------------------------------

# doas
if ! [ -f /etc/doas.conf ]; then
	: > /etc/doas.conf
	for homedir in /home/*; do
		[ -d "$homedir" ] || continue
		username="$(basename "$homedir")"
		printf '%s\n' "permit nopass $username" >> /etc/doas.conf
	done
	printf '%s\n' "permit nopass :wheel" >> /etc/doas.conf
	chmod 0440 /etc/doas.conf
	log "doas configured"
fi

# installurl
printf '%s\n' "http://cdn.openbsd.org/pub/OpenBSD" > /etc/installurl
log "installurl configured"

# ------------------------------------------------------------------
# 2. Install packages from local path (offline)
# ------------------------------------------------------------------
PKG_DIR="/openriot/packages/snapshots/amd64"
if [ -d "$PKG_DIR" ]; then
	cd "$PKG_DIR" || { log "Cannot cd to $PKG_DIR"; exit 1; }
	count=$(ls *.tgz 2>/dev/null | wc -l | tr -d ' ')
	if [ "$count" -eq 0 ]; then
		log "No packages found"
		exit 1
	fi
	log "Installing $count packages..."
	pkg_add -D unsigned -I *.tgz
	log "Package installation finished"
else
	log "Package directory not found: $PKG_DIR"
	exit 1
fi

# ------------------------------------------------------------------
# 3. Install non-free firmware from local path (offline)
# ------------------------------------------------------------------
FW_DIR="/openriot/firmware"
if [ -d "$FW_DIR" ] && [ -n "$(ls "$FW_DIR"/*.tgz 2>/dev/null)" ]; then
	log "Installing firmware..."
	for fw in "$FW_DIR"/*.tgz; do
		[ -f "$fw" ] || continue
		tar xzf "$fw" -C / 2>/dev/null && log "Firmware: $(basename "$fw")"
	done
	log "Firmware install complete"
else
	log "No local firmware found, skipping"
fi

log "Post-install complete"
`

	content = strings.Replace(content, "packages/snapshots/amd64", "packages/"+repoPath+"/amd64", 1)
	path := filepath.Join(siteDir, "install.site")
	return os.WriteFile(path, []byte(content), 0755)
}

// PackageInfo holds info about downloaded packages
type PackageInfo struct {
	Path string
	Size string
}

// GetPackageInfo returns info about downloaded packages
func GetPackageInfo(cfg *Config) (*PackageInfo, error) {
	pkgDir := filepath.Join(cfg.WorkDir, "packages", formatVersion(cfg.Version), "amd64")
	if _, err := os.Stat(pkgDir); err != nil {
		return nil, err
	}

	cmd := exec.Command("du", "-sh", pkgDir)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return &PackageInfo{
		Path: pkgDir,
		Size: strings.TrimSpace(string(out)),
	}, nil
}

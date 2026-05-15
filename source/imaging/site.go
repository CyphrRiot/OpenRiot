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
	pkgSrc := filepath.Join(workDir, "packages", "snapshots", "amd64")
	pkgDst := filepath.Join(openriotDir, "packages", "snapshots", "amd64")
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
	if err := createInstallSite(siteDir); err != nil {
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
func createInstallSite(siteDir string) error {
	content := `#!/bin/sh
# OpenRiot post-install script — runs at end of OpenBSD install
# Environment: root, chroot at new system root, no TTY, no X

log() { echo "[OPENRIOT] $*"; }
fail() { echo "[OPENRIOT] FAIL: $*"; }

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
printf '%s\n' "https://cdn.openbsd.org/pub/OpenBSD" > /etc/installurl
log "installurl configured"

# Add fish to /etc/shells
if ! grep -q '^/usr/local/bin/fish$' /etc/shells 2>/dev/null; then
	printf '%s\n' "/usr/local/bin/fish" >> /etc/shells
	log "fish added to /etc/shells"
fi

# Enable critical services (but NOT xenodm — user must start X manually)
rcctl enable apmd 2>/dev/null && rcctl set apmd flags -A 2>/dev/null
rcctl enable sndiod 2>/dev/null
mkdir -p /etc/wireguard
chmod 700 /etc/wireguard
log "services configured"

# ------------------------------------------------------------------
# 2. Install packages from local path (offline)
# ------------------------------------------------------------------
PKG_PATH_LOCAL="/openriot/packages/snapshots/amd64"
if [ -d "$PKG_PATH_LOCAL" ]; then
	log "Installing packages from local path..."
	cd "$PKG_PATH_LOCAL" || fail "cd $PKG_PATH_LOCAL"
	export PKG_PATH="$PKG_PATH_LOCAL"
	pkg_add -D snapshot -I *.tgz > /tmp/pkg_out 2>&1
	pkg_exit=$?
	cat /tmp/pkg_out
	if [ $pkg_exit -ne 0 ]; then
		if grep -qi "signify\|signature\|pubkey\|verification\|can't verify" /tmp/pkg_out 2>/dev/null; then
			echo ""
			echo "ERROR: Signature verification failed."
			echo "OpenBSD base and packages are out of sync."
			echo ""
			echo "To fix, reboot and run:"
			echo "  doas sysupgrade -s"
			echo "  doas pkg_add -D snap -u"
			echo ""
			exit 1
		fi
	fi
	log "Package install complete"

	# Clean up: remove site79.tgz from image to save space
	# rm -f /7.9/amd64/site79.tgz 2>/dev/null
	# rm -rf /openriot/packages/ 2>/dev/null
else
	log "Package directory not found: $PKG_PATH_LOCAL"
fi

# ------------------------------------------------------------------
# 2b. Install non-free firmware from local path (offline)
# ------------------------------------------------------------------
FW_PATH_LOCAL="/openriot/firmware"
if [ -d "$FW_PATH_LOCAL" ] && [ -n "$(ls "$FW_PATH_LOCAL"/*.tgz 2>/dev/null)" ]; then
	log "Installing firmware from local path..."
	for fw in "$FW_PATH_LOCAL"/*.tgz; do
		[ -f "$fw" ] || continue
		fw_name=$(basename "$fw")
		tar xzf "$fw" -C / 2>/dev/null || continue
		log "Firmware installed: $fw_name"
	done
	log "Firmware install complete"
else
	log "No local firmware found, skipping"
fi

log "Post-install complete"
`

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
	pkgDir := filepath.Join(cfg.WorkDir, "packages", "snapshots", "amd64")
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

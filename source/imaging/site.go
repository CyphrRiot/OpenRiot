package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/logger"
)

// CreateSite creates the openriot.tgz tarball
func CreateSite(cfg *Config) error {
	workDir := cfg.WorkDir
	siteDir := filepath.Join(workDir, "site")
	openriotDir := filepath.Join(siteDir, "openriot")

	// Create work directory
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}

	// Clean old site content completely (handles permission issues from previous runs)
	os.RemoveAll(siteDir)
	os.MkdirAll(siteDir, 0755)
	os.MkdirAll(openriotDir, 0755)

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

	// Create install.site inside siteDir so it ships in the tarball
	// (extracts to /install.site on the new system and runs via install.site(5))
	if err := createInstallSite(siteDir); err != nil {
		return fmt.Errorf("create install.site: %w", err)
	}

	// Create tarball (openriot/ directory, motd, and install.site)
	tgzPath := cfg.OpenriotTgz
	os.Remove(tgzPath) // Remove old if exists

	cmd := exec.Command("tar", "czvf", tgzPath, "-C", siteDir, ".")
	cmd.Dir = siteDir
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
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}

// copyMotd copies the motd file to site directory
func copyMotd(siteDir string) error {
	execDir, _ := os.Executable()
	repoRoot := filepath.Dir(filepath.Dir(execDir))
	motdSrc := filepath.Join(repoRoot, "install", "motd")

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
	printf '%s\n' "permit nopass :wheel" > /etc/doas.conf
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
	for pkg in "$PKG_PATH_LOCAL"/*.tgz; do
		[ -f "$pkg" ] || continue
		log "Installing $(basename "$pkg")..."
		PKG_PATH="$PKG_PATH_LOCAL" pkg_add "$pkg" 2>&1 || fail "pkg_add: $(basename "$pkg")"
	done
	log "Package install complete"
else
	fail "Package directory not found: $PKG_PATH_LOCAL"
fi

# ------------------------------------------------------------------
# 3. Per-user setup
# ------------------------------------------------------------------
for homedir in /home/*; do
	[ -d "$homedir" ] || continue
	username="$(basename "$homedir")"

	# Add to wheel
	usermod -G wheel "$username" 2>/dev/null || fail "usermod $username"

	# Set fish as default shell
	chsh -s /usr/local/bin/fish "$username" 2>/dev/null || fail "chsh $username"

	# Create XDG directories
	for dir in Documents Downloads Music Pictures Videos Code Screenshots; do
		mkdir -p "$homedir/$dir"
		chown "$username:$username" "$homedir/$dir"
	done

	# Welcome message
	cat >> "$homedir/.profile" << 'WELCOME'

# OpenRiot — Welcome
echo ""
echo "Welcome to OpenRiot!"
echo ""
echo "All required packages have been pre-installed."
echo "To complete setup and download the latest OpenRiot repository, run:"
echo ""
echo "    curl -fsSL https://OpenRiot.org/setup.sh | sh"
echo ""
echo "This requires a working network or WiFi connection."
echo ""
WELCOME
	chown "$username:$username" "$homedir/.profile"
done

# 4. Skel for future users
cat >> /etc/skel/.profile << 'WELCOME'

# OpenRiot — Welcome
echo ""
echo "Welcome to OpenRiot!"
echo ""
echo "All required packages have been pre-installed."
echo "To complete setup and download the latest OpenRiot repository, run:"
echo ""
echo "    curl -fsSL https://OpenRiot.org/setup.sh | sh"
echo ""
echo "This requires a working network or WiFi connection."
echo ""
WELCOME

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
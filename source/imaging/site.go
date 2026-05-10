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

	// Setup repo
	repoCacheDir := filepath.Join(getBuildDir(), "repo-cache")
	if err := setupRepo(workDir, repoCacheDir); err != nil {
		return fmt.Errorf("setup repo: %w", err)
	}

	// Move repo to site (use cp+rm to avoid cross-device rename issues)
	repoSource := filepath.Join(workDir, "repo")
	repoTarget := filepath.Join(openriotDir, "repo")

	// Remove existing target if present
	os.RemoveAll(repoTarget)

	// Copy repo to site
	if err := copyDir(repoSource, repoTarget); err != nil {
		return fmt.Errorf("copy repo: %w", err)
	}

	// Remove source repo
	os.RemoveAll(repoSource)

	// Copy packages into tarball so install.site can install them
	pkgSrc := filepath.Join(workDir, "packages", "snapshots", "amd64")
	pkgDst := filepath.Join(openriotDir, "packages", "snapshots", "amd64")
	if err := copyDir(pkgSrc, pkgDst); err != nil {
		return fmt.Errorf("copy packages: %w", err)
	}

	// Create install.site outside siteDir so it stays out of the tarball
	if err := createInstallSite(workDir); err != nil {
		return fmt.Errorf("create install.site: %w", err)
	}

	// Create install.conf outside siteDir so it stays out of the tarball
	if err := createInstallConf(workDir); err != nil {
		return fmt.Errorf("create install.conf: %w", err)
	}

	// Create tarball (only openriot/ directory and motd, NOT install.site/conf)
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

// setupRepo clones or updates the OpenRiot repo
func setupRepo(workDir, cacheDir string) error {
	repoDir := filepath.Join(workDir, "repo")
	repoURL := "https://github.com/CyphrRiot/OpenRiot"

	// Use shell rm to ensure directory is removed (handles more edge cases than os.RemoveAll)
	exec.Command("rm", "-rf", repoDir).Run()

	// Check for cached repo
	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); err == nil {
		// Use cached repo - pull latest
		os.MkdirAll(repoDir, 0755)

		// Clone from cache
		cmd := exec.Command("git", "clone", cacheDir, repoDir)
		cmd.Dir = cacheDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone from cache: %w\n%s", err, out)
		}

		// Fetch and reset to latest
		cmd = exec.Command("git", "fetch", "--depth", "1", "origin", "main")
		cmd.Dir = repoDir
		cmd.Run() // Ignore errors - may already be at latest

		cmd = exec.Command("git", "reset", "--hard", "origin/main")
		cmd.Dir = repoDir
		cmd.Run()

		// Update cache for next time (exclude packages/)
		updateCache(repoDir, cacheDir)
		return nil
	}

	// Fresh clone
	os.RemoveAll(repoDir)
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, repoDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %w\n%s", err, out)
	}

	// Create cache for next time
	os.RemoveAll(cacheDir)
	os.MkdirAll(cacheDir, 0755)
	updateCache(repoDir, cacheDir)

	return nil
}

// updateCache copies repo to cache (excluding packages/)
func updateCache(repoDir, cacheDir string) {
	entries, _ := os.ReadDir(repoDir)
	for _, entry := range entries {
		if entry.Name() == "packages" {
			continue
		}
		src := filepath.Join(repoDir, entry.Name())
		dst := filepath.Join(cacheDir, entry.Name())
		os.RemoveAll(dst)
		os.Rename(src, dst) // Move (not copy) for speed
	}
}

// createInstallSite writes the install.site script
func createInstallSite(siteDir string) error {
	content := `#!/bin/sh
# OpenRiot post-install script
# Runs during OpenBSD installer

log() { echo "[OPENRIOT] $*" ; }

log "OpenRiot post-install starting"

# STEP 1: Configure doas (must happen early so user can sudo)
log "Configuring doas..."
echo "permit nopass :wheel" > /etc/doas.conf
chmod 0440 /etc/doas.conf
log "doas configured"

# STEP 2: Configure installurl for the NEW installed system
log "Configuring installurl..."
echo "https://cdn.openbsd.org/pub/OpenBSD" > /etc/installurl
log "installurl configured"

# STEP 3: Install packages from local path
# The installer already extracted site79.tgz as a set, so /openriot/ exists.
PKG_PATH_LOCAL="/openriot/packages/snapshots/amd64"
log "Installing packages from local path..."
if [ -d "$PKG_PATH_LOCAL" ]; then
    for pkg in "$PKG_PATH_LOCAL"/*.tgz; do
        [ -f "$pkg" ] || continue
        pkg_name=$(basename "$pkg" .tgz)
        log "Installing $pkg_name..."
        PKG_PATH="$PKG_PATH_LOCAL" pkg_add "$pkg" 2>&1 || log "Failed: $pkg_name"
    done
    log "Package install complete"
else
    log "Warning: package directory not found"
fi

# STEP 4: Move repo to user's .local/share/openriot
log "Setting up OpenRiot repo..."
for homedir in /home/*; do
    [ -d "$homedir" ] || continue
    username="$(basename "$homedir")"
    # Ensure user can use doas
    usermod -G wheel "$username" 2>/dev/null || log "Warning: could not add $username to wheel"
    target_dir="$homedir/.local/share/openriot"
    mkdir -p "$target_dir"
    if [ -d /openriot/repo ]; then
        rm -rf "$target_dir/repo" 2>/dev/null || true
        cp -r /openriot/repo "$target_dir/repo"
        chown -R "$username:$username" "$target_dir/repo"
        log "Repo installed for $username"
    fi
done

# STEP 5: Add welcome message to skel
log "Adding welcome message..."
if [ -f /etc/skel/.profile ] && ! grep -q openriot-setup-done /etc/skel/.profile 2>/dev/null; then
    cat >> /etc/skel/.profile << 'WELCOME'

# OpenRiot first login
if [ ! -f ~/.openriot-setup-done ]; then
    echo ""
    echo "Welcome to OpenRiot"
    echo ""
    echo "Run the following command to complete setup:"
    echo ""
    echo "    curl -fsSL https://OpenRiot.org/setup.sh | sh"
    echo ""
    touch ~/.openriot-setup-done
fi
WELCOME
fi

log "Post-install complete"
`

	path := filepath.Join(siteDir, "install.site")
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	return nil
}

// createInstallConf writes install.conf autoinstall answers
func createInstallConf(siteDir string) error {
	content := `# OpenRiot autoinstall answers
# User will be prompted for: disk, hostname, passwords, timezone, partition layout

Which disk is the root disk = ask
Use (W)hole disk MBR, whole disk (G)PT, or (E)dit = edit

System hostname = ask
Password for root = ask
Setup a user = ask
Password for user = ask
What timezone are you in = US/Pacific

# Sets come from the install disk (which is already mounted at /)
Location of sets = disk
Is the disk partition already mounted? = yes
Pathname to the sets = /

# Install site79.tgz from the disk
install site79.tgz = yes
`

	path := filepath.Join(siteDir, "install.conf")
	return os.WriteFile(path, []byte(content), 0644)
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
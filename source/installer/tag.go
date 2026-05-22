package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openriot/paths"
)

const repoURL = "https://github.com/CyphrRiot/OpenRiot"

// InstallTag clones a specific tag and runs the installation
func InstallTag(tag string) error {
	repoDir := paths.OpenRiotDir()

	// Remove existing repo to ensure clean slate
	if err := os.RemoveAll(repoDir); err != nil {
		return fmt.Errorf("removing existing repo: %w", err)
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, repoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}

	// Fetch all branches/tags to allow checkout
	cmd = exec.Command("git", "fetch", "--depth", "1", "origin", "refs/tags/*:refs/tags/*")
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Non-fatal: tag might be a branch or main/commit
	}

	// Normalize tag (add 'v' prefix if missing)
	normalizedTag := tag
	if !strings.HasPrefix(tag, "v") && !strings.HasPrefix(tag, "V") {
		normalizedTag = "v" + tag
	}

	// Try to checkout the tag
	cmd = exec.Command("git", "checkout", normalizedTag)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Try without 'v' prefix
		cmd = exec.Command("git", "checkout", tag)
		cmd.Dir = repoDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git checkout %s: %w", tag, err)
		}
	}

	// Run setup.sh in local mode with the tag info
	cmd = exec.Command(filepath.Join(repoDir, "setup.sh"), "--local")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("setup.sh failed: %w", err)
	}

	return nil
}

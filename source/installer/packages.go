package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"openriot/config"
	"openriot/logger"
)

// InstallPackages installs packages using pkg_add with doas.
// On snapshot systems, the installer detects -current and passes the appropriate flag.
// Falls back to base package name (no version) if exact version fails.
// Returns (failedCount, error)
func InstallPackages(cfg *config.Config, packages []string) (int, error) {
	if len(packages) == 0 {
		return 0, nil
	}

	// Filter out already-installed packages using pkg_info
	var toInstall []string
	for _, pkg := range packages {
		if !isPackageInstalled(pkg) {
			toInstall = append(toInstall, pkg)
		}
	}

	if len(toInstall) == 0 {
		logger.Done("All packages already installed")
		return 0, nil
	}

	logger.Info(fmt.Sprintf("Installing %d packages (%d new, %d already installed)...", len(packages), len(toInstall), len(packages)-len(toInstall)))

	failed := 0
	for _, pkg := range toInstall {
		installName := pkg
		logger.Info(fmt.Sprintf("Installing %s...", pkg))


		// Build pkg_add command: use -D snapshot only for snapshot systems
		installCmd := []string{"doas", "pkg_add"}
		if cfg.IsSnapshot() {
			installCmd = append(installCmd, "-D", "snapshot")
		}
		installCmd = append(installCmd, installName)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		cmd := exec.CommandContext(ctx, installCmd[0], installCmd[1:]...)
		output, err := cmd.CombinedOutput()
		cancel()

		if err != nil {
			outputStr := truncateOutput(output)
			if ctx.Err() == context.DeadlineExceeded {
				logger.Warn(fmt.Sprintf("Timed out after 10m: %s", pkg))
			}
			// Retry with base name (without version) on failure
			base := config.GetBaseName(pkg)
			if base != pkg {
				logger.Info(fmt.Sprintf("Retrying %s with latest version...", base))
				installCmd[len(installCmd)-1] = base
				ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
				cmd = exec.CommandContext(ctx, installCmd[0], installCmd[1:]...)
				output, err = cmd.CombinedOutput()
				cancel()
				if err == nil {
					logger.Done(fmt.Sprintf("%s installed (latest version)", base))
					continue
				}
				outputStr = truncateOutput(output)
				if ctx.Err() == context.DeadlineExceeded {
					logger.Warn(fmt.Sprintf("Timed out after 10m: %s", base))
				}
			}
			if cfg.IsSnapshot() && isSignatureError(outputStr) {
				logger.Fail("OpenBSD base and packages are out of sync.")
				fmt.Println()
				fmt.Println("Your snapshot packages cannot be verified against your current base system.")
				fmt.Println("This is a known issue on -current when base and package builds drift.")
				fmt.Println()
				fmt.Println("To fix this, run the following commands:")
				fmt.Println("  doas sysupgrade -s")
				fmt.Println("  (reboot when prompted)")
				fmt.Println("  doas pkg_add -u")
				fmt.Println()
				fmt.Println("Then re-run the OpenRiot installer.")
				os.Exit(1)
			}
			logger.Warn(fmt.Sprintf("Failed to install %s:\n    %s", pkg, outputStr))
			failed++
		} else {
			logger.Done(fmt.Sprintf("%s installed", pkg))
		}
	}

	if failed > 0 {
		logger.Warn(fmt.Sprintf("%d packages failed to install.", failed))
		logger.Warn("You can install remaining ones manually: doas pkg_add <package>")
	} else {
		logger.Done("All packages installed")
	}

	return failed, nil
}

func truncateOutput(output []byte) string {
	s := string(output)
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}

// isPackageInstalled checks if a package is already installed.
// Handles both exact version ("firefox-149.0.2p0") and base name ("firefox").
func isPackageInstalled(pkg string) bool {
	// Check with exact name first
	cmd := exec.Command("pkg_info", "-e", pkg)
	if cmd.Run() == nil {
		return true
	}
	// Check with base name (for packages installed as latest version)
	base := config.GetBaseName(pkg)
	if base != pkg {
		cmd = exec.Command("pkg_info", "-e", base)
		if cmd.Run() == nil {
			return true
		}
	}
	return false
}

// isSignatureError checks whether pkg_add output indicates a base/package
// signature mismatch, which happens when OpenBSD -current snapshots drift.
func isSignatureError(output string) bool {
	s := strings.ToLower(output)
	return strings.Contains(s, "signify") ||
		strings.Contains(s, "signature") ||
		strings.Contains(s, "pubkey") ||
		strings.Contains(s, "verification") ||
		strings.Contains(s, "can't verify") ||
		strings.Contains(s, "gpg")
}

package installer

import (
	"fmt"
	"os/exec"

	"openriot/config"
)

// InstallPackages installs packages using pkg_add with doas.
// On snapshot systems, -D snapshot is passed to pkg_add.
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
		fmt.Printf("%s[DONE]%s All packages already installed\n", Green, Reset)
		return 0, nil
	}

	fmt.Printf("%s[INFO]%s Installing %d packages (%d new, %d already installed)...\n", Cyan, Reset, len(packages), len(toInstall), len(packages)-len(toInstall))

	failed := 0
	for _, pkg := range toInstall {
		installName := pkg
		fmt.Printf("%s[INFO]%s Installing %s...\n", Cyan, Reset, pkg)


		// Build pkg_add command: use -D snapshot only for snapshot systems
		installCmd := []string{"doas", "pkg_add"}
		if cfg.IsSnapshot() {
			installCmd = append(installCmd, "-D", "snapshot")
		}
		installCmd = append(installCmd, installName)

		cmd := exec.Command(installCmd[0], installCmd[1:]...)
		output, err := cmd.CombinedOutput()


		if err != nil {
			outputStr := string(output)
			if len(outputStr) > 300 {
				outputStr = outputStr[:300] + "..."
			}
			// Retry with base name (without version) on failure
			base := config.GetBaseName(pkg)
			if base != pkg {
				fmt.Printf("%s[INFO]%s Retrying %s with latest version...\n", Cyan, Reset, base)
				installCmd[len(installCmd)-1] = base
				cmd = exec.Command(installCmd[0], installCmd[1:]...)
				output, err = cmd.CombinedOutput()
				if err == nil {
					fmt.Printf("%s[DONE]%s %s installed (latest version)\n", Green, Reset, base)
					continue
				}
				outputStr = string(output)
				if len(outputStr) > 300 {
					outputStr = outputStr[:300] + "..."
				}
			}
			fmt.Printf("%s[WARN]%s Failed to install %s:\n    %s\n", Yellow, Reset, pkg, outputStr)
			failed++
		} else {
			fmt.Printf("%s[DONE]%s %s installed\n", Green, Reset, pkg)
		}
	}

	if failed > 0 {
		fmt.Printf("%s[WARN]%s %d packages failed to install.\n", Yellow, Reset, failed)
		fmt.Printf("%s[WARN]%s You can install remaining ones manually: doas pkg_add <package>\n", Yellow, Reset)
	} else {
		fmt.Printf("%s[DONE]%s All packages installed\n", Green, Reset)
	}

	return failed, nil
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

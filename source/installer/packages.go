package installer

import (
	"fmt"
	"os/exec"

	"openriot/config"
)

// InstallPackages installs packages using pkg_add with doas.
// On snapshot systems, -D snapshot is passed to pkg_add.
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
		fmt.Printf("%s[INFO]%s Installing %s...\n", Cyan, Reset, pkg)

		// Build pkg_add command: use -D snapshot only for snapshot systems
		installCmd := []string{"doas", "pkg_add"}
		if cfg.IsSnapshot() {
			installCmd = append(installCmd, "-D", "snapshot")
		}
		installCmd = append(installCmd, pkg)

		cmd := exec.Command(installCmd[0], installCmd[1:]...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			outputStr := string(output)
			if len(outputStr) > 300 {
				outputStr = outputStr[:300] + "..."
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

// isPackageInstalled checks if a package is already installed
func isPackageInstalled(pkg string) bool {
	cmd := exec.Command("pkg_info", "-e", pkg)
	return cmd.Run() == nil
}

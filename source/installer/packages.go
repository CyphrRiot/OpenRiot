package installer

import (
	"fmt"
	"os/exec"
)

// InstallPackages installs packages using pkg_add
func InstallPackages(packages []string) error {
	if len(packages) == 0 {
		fmt.Println("[INFO]  No packages to install")
		return nil
	}

	// Filter out already-installed packages
	var toInstall []string
	for _, pkg := range packages {
		if !isPackageInstalled(pkg) {
			toInstall = append(toInstall, pkg)
		}
	}

	if len(toInstall) == 0 {
		fmt.Println("[INFO]  All packages already installed")
		return nil
	}

	fmt.Printf("[INFO]  Installing %d packages with pkg_add\n", len(toInstall))

	// Install packages one at a time for progress tracking
	for i, pkg := range toInstall {
		fmt.Printf("[INFO]  Installing %s...\n", pkg)

		cmd := exec.Command("pkg_add", pkg)
		output, err := cmd.CombinedOutput()

		if err != nil {
			outputStr := string(output)
			if len(outputStr) > 300 {
				outputStr = outputStr[:300] + "..."
			}
			fmt.Printf("[ERR!]  Failed to install %s: %s\n", pkg, outputStr)
			return fmt.Errorf("pkg_add failed for %s: %w", pkg, err)
		}

		fmt.Printf("[INFO]  Installed %s\n", pkg)

		// Log progress
		fmt.Printf("[INFO]  Progress: %d/%d packages installed\n", i+1, len(toInstall))
	}

	fmt.Printf("[INFO]  Installed %d packages\n", len(toInstall))
	return nil
}

// isPackageInstalled checks if a package is already installed
func isPackageInstalled(pkg string) bool {
	cmd := exec.Command("pkg_info", "-e", pkg)
	return cmd.Run() == nil
}

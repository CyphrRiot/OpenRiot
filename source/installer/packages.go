package installer

import (
	"fmt"
	"os/exec"
)

// InstallPackages installs packages using pkg_add
func InstallPackages(packages []string) error {
	if len(packages) == 0 {
		fmt.Printf("%s[INFO]%s  No packages to install\n", Blue, Reset)
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
		fmt.Printf("%s[INFO]%s  All packages already installed\n", Blue, Reset)
		return nil
	}

	fmt.Printf("%s[INFO]%s  Installing %d packages with pkg_add\n", Blue, Reset, len(toInstall))

	// Install packages one at a time for progress tracking
	for i, pkg := range toInstall {
		fmt.Printf("%s[INFO]%s  Installing %s...\n", Blue, Reset, pkg)

		cmd := exec.Command("pkg_add", pkg)
		output, err := cmd.CombinedOutput()

		if err != nil {
			outputStr := string(output)
			if len(outputStr) > 300 {
				outputStr = outputStr[:300] + "..."
			}
			fmt.Printf("%s[ERR!]%s  Failed to install %s: %s\n", Red, Reset, pkg, outputStr)
			return fmt.Errorf("pkg_add failed for %s: %w", pkg, err)
		}

		fmt.Printf("%s[INFO]%s  Installed %s\n", Green, Reset, pkg)

		// Log progress
		fmt.Printf("%s[INFO]%s  Progress: %d/%d packages installed\n", Blue, Reset, i+1, len(toInstall))
	}

	fmt.Printf("%s[INFO]%s  Installed %d packages\n", Green, Reset, len(toInstall))
	return nil
}

// isPackageInstalled checks if a package is already installed
func isPackageInstalled(pkg string) bool {
	cmd := exec.Command("pkg_info", "-e", pkg)
	return cmd.Run() == nil
}

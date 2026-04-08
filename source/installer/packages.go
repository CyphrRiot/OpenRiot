package installer

import (
	"fmt"
	"os/exec"
)

// InstallPackages installs packages using pkg_add with doas
// Returns (failedCount, error)
func InstallPackages(packages []string) (int, error) {
	if len(packages) == 0 {
		fmt.Printf("%s[INFO]%s  No packages to install\n", Blue, Reset)
		return 0, nil
	}

	// Filter out already-installed packages
	var toInstall []string
	for _, pkg := range packages {
		if !isPackageInstalled(pkg) {
			toInstall = append(toInstall, pkg)
		} else {
			fmt.Printf("%s[INFO]%s  [SKIP] %s already installed\n", Blue, Reset, pkg)
		}
	}

	if len(toInstall) == 0 {
		fmt.Printf("%s[INFO]%s  All packages already installed\n", Green, Reset)
		return 0, nil
	}

	fmt.Printf("%s[INFO]%s  Installing %d packages with pkg_add\n", Blue, Reset, len(toInstall))
	fmt.Printf("%s[WARN]%s  This may take a while (Nerd Fonts are large)...\n", Yellow, Reset)

	failed := 0
	for i, pkg := range toInstall {
		fmt.Printf("%s[INFO]%s  Installing %s...\n", Blue, Reset, pkg)

		// Use doas for root privileges, -D unsigned to allow unsigned packages
		cmd := exec.Command("doas", "pkg_add", "-D", "unsigned", pkg)
		output, err := cmd.CombinedOutput()

		if err != nil {
			outputStr := string(output)
			if len(outputStr) > 300 {
				outputStr = outputStr[:300] + "..."
			}
			fmt.Printf("%s[WARN]%s  Failed to install %s:\n    %s\n", Yellow, Reset, pkg, outputStr)
			failed++
		} else {
			fmt.Printf("%s[INFO]%s  %s installed.\n", Green, Reset, pkg)
		}

		// Log progress
		fmt.Printf("%s[INFO]%s  Progress: %d/%d packages installed\n", Blue, Reset, i+1, len(toInstall))
	}

	if failed > 0 {
		fmt.Printf("%s[WARN]%s  %d packages failed to install.\n", Yellow, Reset, failed)
		fmt.Printf("%s[WARN]%s  You can install remaining ones manually: doas pkg_add <package>\n", Yellow, Reset)
	} else {
		fmt.Printf("%s[INFO]%s  All packages installed successfully!\n", Green, Reset)
	}

	return failed, nil
}

// isPackageInstalled checks if a package is already installed
func isPackageInstalled(pkg string) bool {
	cmd := exec.Command("pkg_info", "-e", pkg)
	return cmd.Run() == nil
}

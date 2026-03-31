package installer

import (
	"fmt"
	"os/exec"

	"openriot/logger"
)

// InstallPackages installs packages using pkg_add
func InstallPackages(packages []string) error {
	if len(packages) == 0 {
		logger.Log("Info", "Package", "Install", "None to install")
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
		logger.LogMessage("INFO", "All packages already installed")
		return nil
	}

	logger.LogMessage("INFO", fmt.Sprintf("Installing %d packages with pkg_add", len(toInstall)))

	// Install all packages at once
	cmd := exec.Command("pkg_add", toInstall...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		outputStr := string(output)
		if len(outputStr) > 300 {
			outputStr = outputStr[:300] + "..."
		}
		logger.LogMessage("ERROR", fmt.Sprintf("pkg_add failed: %s", outputStr))
		return fmt.Errorf("pkg_add failed: %w", err)
	}

	logger.LogMessage("SUCCESS", fmt.Sprintf("✅ Installed %d packages", len(toInstall)))
	return nil
}

// isPackageInstalled checks if a package is already installed
func isPackageInstalled(pkg string) bool {
	cmd := exec.Command("pkg_info", "-e", pkg)
	return cmd.Run() == nil
}

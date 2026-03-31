package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"openriot/installer"
	"openriot/logger"
	"openriot/tui"
)

var version = "0.1"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("openriot", version)
		os.Exit(0)
	}

	// Initialize logger
	if err := logger.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.LogMessage("INFO", "OpenRiot Installer starting...")

	// Install packages
	testPackages := []string{"fish", "sway", "waybar", "foot"}
	if err := installer.InstallPackages(testPackages); err != nil {
		logger.LogMessage("ERROR", fmt.Sprintf("Failed to install packages: %v", err))
		os.Exit(1)
	}
	logger.LogMessage("SUCCESS", "Packages installed successfully!")

	model := tui.NewInstallModel()
	program := tea.NewProgram(model)
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

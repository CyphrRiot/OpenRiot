package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"openriot/tui"
)

var version = "0.1"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("openriot", version)
		os.Exit(0)
	}

	model := tui.NewInstallModel()
	program := tea.NewProgram(model)
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

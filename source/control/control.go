package control

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the Control Center TUI.
func Run() error {
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("control: %w", err)
	}
	return nil
}
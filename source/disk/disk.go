package disk

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the Disk Manager TUI.
func Run() error {
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("disk: %w", err)
	}
	return nil
}

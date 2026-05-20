package resolution

import (
	"fmt"

	"openriot/notify"
	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the Resolution TUI.
func Run() error {
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("resolution: %w", err)
	}
	return nil
}

// Restore applies the last saved resolution on startup.
func Restore() error {
	err := RestoreResolution()
	if err != nil {
		// No state file or corrupted — not fatal, just skip
		return nil
	}
	notify.SendNotify("display", "Resolution", "Restored saved resolution", "normal", 3000, 0)
	return nil
}
package nmtui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the Wi-Fi TUI. It detects the wireless interface and starts the
// Bubble Tea program in alternate-screen mode.
func Run() error {
	iface, err := FindWiFiInterface()
	if err != nil {
		return fmt.Errorf("no wireless interface found: %w", err)
	}

	p := tea.NewProgram(
		NewModel(iface),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("nmtui: %w", err)
	}
	return nil
}

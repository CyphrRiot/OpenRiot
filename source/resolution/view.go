package resolution

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// view renders the TUI based on the current state.
func (m model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	if m.err != nil {
		return m.renderError()
	}

	switch m.state {
	case stateListDisplays:
		return m.renderDisplayList()
	case stateListModes:
		return m.renderModeList()
	case stateListRates:
		return m.renderRateList()
	case stateResult:
		return m.renderResult()
	}

	return ""
}

func (m model) renderHelp() string {
	return lipgloss.NewStyle().
		Padding(2).
		Render(m.help.View(m.keys))
}

func (m model) renderError() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Error"))
	b.WriteString("\n\n")
	b.WriteString(errorStyle.Render(m.err.Error()))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Press q to quit"))
	return b.String()
}

func (m model) renderDisplayList() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Displays"))
	b.WriteString("\n\n")

	if len(m.displays) == 0 {
		b.WriteString(helpStyle.Render("No displays detected."))
		return b.String()
	}

	for i, d := range m.displays {
		name := d.Name
		if d.Primary {
			name += " (primary)"
		}
		if d.Current != "" {
			name += fmt.Sprintf(" [%s]", d.Current)
		}
		if i == m.cursor {
			b.WriteString(cursorStyle.Render("> " + name))
		} else {
			b.WriteString(itemStyle.Render("  " + name))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter: select | q: quit | ?: help"))
	return b.String()
}

func (m model) renderModeList() string {
	var b strings.Builder
	if m.selectedDisp >= 0 && m.selectedDisp < len(m.displays) {
		d := m.displays[m.selectedDisp]
		b.WriteString(titleStyle.Render(fmt.Sprintf("Resolutions: %s", d.Name)))
	}
	b.WriteString("\n\n")

	if m.selectedDisp >= 0 && m.selectedDisp < len(m.displays) {
		modes := m.displays[m.selectedDisp].Modes
		for i, mode := range modes {
			name := mode.Resolution
			if i == m.cursor {
				b.WriteString(cursorStyle.Render("> " + name))
			} else {
				b.WriteString(itemStyle.Render("  " + name))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter: select | esc: back | q: quit"))
	return b.String()
}

func (m model) renderRateList() string {
	var b strings.Builder
	if m.selectedDisp >= 0 && m.selectedDisp < len(m.displays) {
		d := m.displays[m.selectedDisp]
		if m.selectedMode >= 0 && m.selectedMode < len(d.Modes) {
			mode := d.Modes[m.selectedMode]
			b.WriteString(titleStyle.Render(fmt.Sprintf("Refresh Rates: %s @ %s", d.Name, mode.Resolution)))
		}
	}
	b.WriteString("\n\n")

	if m.selectedDisp >= 0 && m.selectedDisp < len(m.displays) {
		modes := m.displays[m.selectedDisp].Modes
		if m.selectedMode >= 0 && m.selectedMode < len(modes) {
			for i, rate := range modes[m.selectedMode].Rates {
				name := rate.String()
				if i == m.cursor {
					b.WriteString(cursorStyle.Render("> " + name))
				} else {
					b.WriteString(itemStyle.Render("  " + name))
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter: apply | esc: back | q: quit"))
	return b.String()
}

func (m model) renderResult() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Result"))
	b.WriteString("\n\n")
	b.WriteString(resultStyle.Render(m.resultMsg))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Press esc or q to return"))
	return b.String()
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7aa2f7"))
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#bb9af7"))
	itemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#a9b1d6"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))
	resultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
)

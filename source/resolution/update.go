package resolution

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update is the Bubble Tea update function.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case displaysMsg:
		m.displays = msg.displays
		return m, nil
	case errMsg:
		m.err = msg.err
		return m, nil
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Quit) {
			m.showHelp = false
			return m, nil
		}
		return m, nil
	}

	// Quit always works
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	// Help
	if key.Matches(msg, m.keys.Help) {
		m.showHelp = !m.showHelp
		return m, nil
	}

	// Back
	if key.Matches(msg, m.keys.Back) {
		switch m.state {
		case stateListModes:
			m.state = stateListDisplays
			m.cursor = m.selectedDisp
		case stateListRates:
			m.state = stateListModes
			m.cursor = m.selectedMode
		case stateResult:
			m.state = stateListDisplays
			m.cursor = 0
		}
		return m, nil
	}

	// Navigation
	switch {
	case key.Matches(msg, m.keys.Up):
		m.cursor--
		if m.cursor < 0 {
			m.cursor = m.maxCursor()
		}
	case key.Matches(msg, m.keys.Down):
		m.cursor++
		if m.cursor > m.maxCursor() {
			m.cursor = 0
		}
	case key.Matches(msg, m.keys.Select):
		return m.handleSelect()
	}
	return m, nil
}

func (m model) maxCursor() int {
	switch m.state {
	case stateListDisplays:
		return len(m.displays) - 1
	case stateListModes:
		if m.selectedDisp >= 0 && m.selectedDisp < len(m.displays) {
			return len(m.displays[m.selectedDisp].Modes) - 1
		}
	case stateListRates:
		if m.selectedDisp >= 0 && m.selectedDisp < len(m.displays) {
			modes := m.displays[m.selectedDisp].Modes
			if m.selectedMode >= 0 && m.selectedMode < len(modes) {
				return len(modes[m.selectedMode].Rates) - 1
			}
		}
	}
	return 0
}

func (m model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateListDisplays:
		if len(m.displays) == 0 {
			return m, nil
		}
		m.selectedDisp = m.cursor
		m.state = stateListModes
		m.cursor = 0
	case stateListModes:
		disp := m.displays[m.selectedDisp]
		if len(disp.Modes) == 0 {
			return m, nil
		}
		m.selectedMode = m.cursor
		m.state = stateListRates
		m.cursor = 0
	case stateListRates:
		disp := m.displays[m.selectedDisp]
		mode := disp.Modes[m.selectedMode]
		if len(mode.Rates) == 0 {
			return m, nil
		}
		m.selectedRate = m.cursor
		rate := mode.Rates[m.selectedRate]
		err := Apply(disp.Name, mode.Resolution, rate.Value)
		if err != nil {
			m.err = err
			m.resultMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.resultMsg = fmt.Sprintf("Set %s to %s @ %.2fHz", disp.Name, mode.Resolution, rate.Value)
		}
		m.state = stateResult
	}
	return m, nil
}

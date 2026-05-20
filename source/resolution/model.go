package resolution

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// viewState drives which screen the TUI renders.
type viewState int

const (
	stateListDisplays viewState = iota
	stateListModes
	stateListRates
	stateResult
)

// keyMap defines every user-actionable key binding.
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
	Quit   key.Binding
	Help   key.Binding
}

// defaultKeys is the canonical key map for the Resolution TUI.
var defaultKeys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// ShortHelp returns the key bindings shown in the mini help bar.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Back, k.Quit}
}

// FullHelp returns the key bindings shown in the full help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select, k.Back, k.Quit},
		{k.Help},
	}
}

// model is the Bubble Tea model for the Resolution TUI.
type model struct {
	displays     []Display
	selectedDisp int
	selectedMode int
	selectedRate int
	cursor       int
	state        viewState
	width        int
	height       int
	err          error
	resultMsg    string
	keys         keyMap
	help         help.Model
	showHelp     bool
}

// NewModel creates a new model.
func NewModel() tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return model{
		state:    stateListDisplays,
		keys:     defaultKeys,
		help:     help.New(),
		cursor:   0,
	}
}

// Init is the Bubble Tea entry command.
func (m model) Init() tea.Cmd {
	return fetchDisplaysCmd
}

// fetchDisplaysCmd loads displays from xrandr.
func fetchDisplaysCmd() tea.Msg {
	displays, err := GetDis()
	if err != nil {
		return errMsg{err: err}
	}
	return displaysMsg{displays: displays}
}

type errMsg struct {
	err error
}

type displaysMsg struct {
	displays []Display
}

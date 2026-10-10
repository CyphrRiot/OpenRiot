package control

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// section is a top-level group of controls in the sidebar.
type section struct {
	icon string
	name string
}

// sections is the ordered sidebar navigation.
var sections = []section{
	{icon: "󰕾", name: "Audio"},
	{icon: "󰹑", name: "Display"},
	{icon: "󱚹", name: "Network"},
	{icon: "󰁹", name: "Power"},
	{icon: "󰌵", name: "Privacy"},
	{icon: "󰙨", name: "System"},
}

// keyMap defines every user-actionable key binding.
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Prev   key.Binding
	Next   key.Binding
	Select key.Binding
	Quit   key.Binding
	Help   key.Binding
}

// defaultKeys is the canonical key map for the Control Center TUI.
var defaultKeys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑", "prev item"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓", "next item"),
	),
	Prev: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("←", "prev section"),
	),
	Next: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("→", "next section"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter", "select"),
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
	return []key.Binding{k.Up, k.Down, k.Prev, k.Next, k.Quit}
}

// FullHelp returns the key bindings shown in the full help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Prev, k.Next, k.Select},
		{k.Help, k.Quit},
	}
}

// model is the Bubble Tea model for the Control Center TUI.
type model struct {
	sections     []section
	section      int
	cursor       int
	width        int
	height       int
	keys         keyMap
	help         help.Model
	showHelp     bool
	volume       int
	muted        bool
	batPercent   int
	batAC        int
	batMinutes   int
	brightness   int
	hdmiHasExt   bool
	hdmiActive   bool
	wifi         bool
	eth          bool
	wg           bool
	night        bool
	stealth      bool
	proton       bool
	recording    bool
	version      string
}

// NewModel creates a new model.
func NewModel() tea.Model {
	return model{
		sections: sections,
		keys:     defaultKeys,
		help:     help.New(),
	}
}

// Init is the Bubble Tea entry command.
func (m model) Init() tea.Cmd {
	return tea.Batch(fetchVolumeCmd, fetchPowerCmd, fetchDisplayCmd, fetchSystemCmd, tickCmd())
}

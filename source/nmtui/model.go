package nmtui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// viewState drives which screen the TUI renders.
type viewState int

const (
	stateList viewState = iota
	statePassword
	stateConnecting
	stateResult
	stateActiveInfo
	stateConfirmDisconnect
)

// keyMap defines every user-actionable key binding.
type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Select     key.Binding
	Back       key.Binding
	Quit       key.Binding
	Refresh    key.Binding
	Help       key.Binding
	Info       key.Binding
	Disconnect key.Binding
}

// defaultKeys is the canonical key map for the Wi-Fi TUI.
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
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Info: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "info"),
	),
	Disconnect: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "disconnect"),
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
		{k.Refresh, k.Help, k.Info, k.Disconnect},
	}
}

// model is the Bubble Tea model for the Wi-Fi TUI.
type model struct {
	iface     string
	aps       []WiFiAP
	cursor    int
	password  textinput.Model
	spinner   spinner.Model
	err       error
	state     viewState
	width     int
	height    int
	conn      *ConnectionInfo
	resultMsg string
	scanning  bool
	keys      keyMap
	help      help.Model
	showHelp  bool
}

// NewModel creates a new model for the given Wi-Fi interface.
func NewModel(iface string) tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter password..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot

	return model{
		iface:     iface,
		state:     stateList,
		scanning:  true,
		password:  ti,
		spinner:   s,
		keys:      defaultKeys,
		help:      help.New(),
	}
}

// Init is the Bubble Tea entry command.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchWifiNetworksCmd(m.iface))
}





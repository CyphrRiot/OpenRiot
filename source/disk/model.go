package disk

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// viewState drives which screen the TUI renders.
type viewState int

const (
	stateMenu viewState = iota
	stateDriveList
	stateConfirm
	statePassword
	stateBenchmarkConfig
	stateRunning
	stateResult
)

// actionType identifies which operation the user selected.
type actionType int

const (
	actionDiscover actionType = iota
	actionMount
	actionUmount
	actionFormat
	actionEncrypt
	actionBenchmark
)

// keyMap defines every user-actionable key binding.
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
	Quit   key.Binding
	Help   key.Binding
	Yes    key.Binding
	No     key.Binding
}

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
	Yes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Back, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select, k.Back, k.Quit},
		{k.Help, k.Yes, k.No},
	}
}

// benchmarkPreset defines test size for Benchmark.
type benchmarkPreset struct {
	name       string
	desc       string
	writeSize  string
	rwSize     string
	duration   time.Duration
}

var benchmarkPresets = []benchmarkPreset{
	{name: "Quick", desc: "15s smoke test — 512MB", writeSize: "512M", rwSize: "256M", duration: 15 * time.Second},
	{name: "Standard", desc: "60s balanced test — 2GB", writeSize: "2G", rwSize: "1G", duration: 60 * time.Second},
	{name: "Thorough", desc: "3min heavy test — 4GB", writeSize: "4G", rwSize: "2G", duration: 180 * time.Second},
}

// model is the Bubble Tea model for the Disk TUI.
type model struct {
	drives         []Drive
	filteredDrives []Drive
	cursor         int
	state          viewState
	action         actionType
	benchCursor    int
	width          int
	height         int
	err            error
	resultMsg      string
	confirmMsg     string
	password       textinput.Model
	spinner        spinner.Model
	keys           keyMap
	help           help.Model
	showHelp       bool
}

// menuItem represents a main menu option.
type menuItem struct {
	name        string
	description string
	action      actionType
	destructive bool
}

var menuItems = []menuItem{
	{name: "Discover Drives", description: "Scan and list all storage devices", action: actionDiscover},
	{name: "Mount Drive", description: "Mount a drive (auto-detects encryption)", action: actionMount},
	{name: "Umount Drive", description: "Unmount and detach encrypted volumes", action: actionUmount},
	{name: "Format Drive", description: "Format with 4.2BSD filesystem ⚠️ DESTRUCTIVE", action: actionFormat, destructive: true},
	{name: "Encrypt Drive", description: "Setup softraid crypto ⚠️ DESTRUCTIVE", action: actionEncrypt, destructive: true},
	{name: "Benchmark Drive", description: "Run fio performance tests", action: actionBenchmark},
}

// NewModel creates a new model.
func NewModel() tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter passphrase..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot

	return model{
		state:       stateMenu,
		keys:        defaultKeys,
		help:        help.New(),
		password:    ti,
		spinner:     s,
		cursor:      0,
		benchCursor: 0,
	}
}

// Init is the Bubble Tea entry command.
func (m model) Init() tea.Cmd {
	return nil
}

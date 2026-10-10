package control

import (
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type volumeMsg struct {
	vol   int
	muted bool
}

type powerMsg struct {
	battery batteryInfo
}

type displayMsg struct {
	brightness int
	hasExt     bool
	extActive  bool
}

type systemMsg struct {
	state systemState
}

type tickMsg time.Time

func fetchVolumeCmd() tea.Msg {
	v, m := volumeState()
	return volumeMsg{vol: v, muted: m}
}

func fetchPowerCmd() tea.Msg {
	b := getBattery()
	return powerMsg{battery: b}
}

func fetchDisplayCmd() tea.Msg {
	b := getBrightness()
	he, ea := hdmiInfo()
	return displayMsg{brightness: b, hasExt: he, extActive: ea}
}

func fetchSystemCmd() tea.Msg {
	return systemMsg{state: fetchSystemState()}
}

func tickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case volumeMsg:
		m.volume = msg.vol
		m.muted = msg.muted
		return m, nil
	case powerMsg:
		m.batPercent = msg.battery.percent
		m.batAC = msg.battery.ac
		m.batMinutes = msg.battery.minutes
		return m, nil
	case displayMsg:
		m.brightness = msg.brightness
		m.hdmiHasExt = msg.hasExt
		m.hdmiActive = msg.extActive
		return m, nil
	case systemMsg:
		m.wifi = msg.state.wifi
		m.eth = msg.state.eth
		m.wg = msg.state.wg
		m.night = msg.state.night
		m.stealth = msg.state.stealth
		m.proton = msg.state.proton
		m.recording = msg.state.recording
		m.version = msg.state.version
		return m, nil
	case tickMsg:
		return m, tea.Batch(fetchVolumeCmd, fetchPowerCmd, fetchSystemCmd, tickCmd())
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		if key.Matches(msg, m.keys.Help) || key.Matches(msg, m.keys.Quit) {
			m.showHelp = false
			return m, nil
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	if key.Matches(msg, m.keys.Help) {
		m.showHelp = !m.showHelp
		return m, nil
	}

	if key.Matches(msg, m.keys.Prev) {
		m.section--
		if m.section < 0 {
			m.section = len(m.sections) - 1
		}
		m.cursor = 0
		return m, nil
	}

	if key.Matches(msg, m.keys.Next) {
		m.section++
		if m.section >= len(m.sections) {
			m.section = 0
		}
		m.cursor = 0
		return m, nil
	}

	switch msg.String() {
	case "+", "=":
		if m.section == 0 {
			return m.handleVolumeAdjust(5)
		}
		if m.section == 1 {
			return m.handleBrightnessAdjust(5)
		}
		return m, nil
	case "-", "_":
		if m.section == 0 {
			return m.handleVolumeAdjust(-5)
		}
		if m.section == 1 {
			return m.handleBrightnessAdjust(-5)
		}
		return m, nil
	case "r", "R":
		if m.section == 1 {
			_ = exec.Command("alacritty", "--class", "openriot_resolution", "-e", "openriot", "--resolution-tui").Start()
		}
		if m.section == 3 && m.cursor == 3 {
			_ = rebootCmd()
		}
		return m, nil
	case "m":
		return m.handleMuteToggle()
	case "p", "P":
		if m.section == 3 && m.cursor == 2 {
			_ = suspendCmd()
		}
		return m, nil
	case "s", "S":
		if m.section == 3 && m.cursor == 4 {
			_ = shutdownCmd()
		}
		return m, nil
	case "l", "L":
		if m.section == 3 && m.cursor == 5 {
			_ = lockCmd()
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.Up) {
		m.cursor--
		if m.cursor < 0 {
			m.cursor = m.itemsInSection() - 1
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.Down) {
		m.cursor++
		if m.cursor >= m.itemsInSection() {
			m.cursor = 0
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.Select) {
		return m.handleSelect()
	}

	return m, nil
}

func (m model) itemsInSection() int {
	switch m.section {
	case 0:
		return 2
	case 1:
		return 3
	case 2:
		return 3
	case 3:
		if m.batPercent == 255 {
			return 1
		}
		return 6
	case 4:
		if m.proton {
			return 3
		}
		return 2
	case 5:
		return 3
	}
	return 1
}

func (m model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.section {
	case 0:
		if m.cursor == 0 {
			return m.handleVolumeAdjust(5)
		}
		return m.handleMuteToggle()
	case 1:
		if m.cursor == 0 {
			return m.handleBrightnessAdjust(5)
		}
		if m.cursor == 1 {
			return m.handleHDMIToggle()
		}
		if m.cursor == 2 {
			_ = exec.Command("alacritty", "--class", "openriot_resolution", "-e", "openriot", "--resolution-tui").Start()
		}
		return m, nil
	case 3:
		switch m.cursor {
		case 2:
			_ = suspendCmd()
		case 3:
			_ = rebootCmd()
		case 4:
			_ = shutdownCmd()
		case 5:
			_ = lockCmd()
		}
		return m, nil
	case 2:
		switch m.cursor {
		case 0:
			wifiAction()
		case 2:
			if err := toggleWireguard(); err == nil {
				m.wg = !m.wg
			}
		}
		return m, nil
	case 4:
		switch m.cursor {
		case 0:
			if err := toggleNightlight(); err == nil {
				m.night = !m.night
			}
		case 1:
			if err := toggleStealth(); err == nil {
				m.stealth = !m.stealth
			}
		case 2:
			if m.proton {
				_ = triggerProtonSync()
			}
		}
		return m, nil
	case 5:
		switch m.cursor {
		case 0:
			_ = runUpdate()
		case 1:
			if err := toggleRecording(); err == nil {
				m.recording = !m.recording
			}
		case 2:
			_ = takeScreenshot()
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleVolumeAdjust(delta int) (tea.Model, tea.Cmd) {
	newVol := m.volume + delta
	if newVol > 100 {
		newVol = 100
	}
	if newVol < 0 {
		newVol = 0
	}
	_ = setVolume(newVol)
	m.volume = newVol
	return m, nil
}

func (m model) handleMuteToggle() (tea.Model, tea.Cmd) {
	_ = toggleMute()
	m.muted = !m.muted
	return m, nil
}

func (m model) handleBrightnessAdjust(delta int) (tea.Model, tea.Cmd) {
	newB := m.brightness + delta
	if newB > 100 {
		newB = 100
	}
	if newB < 0 {
		newB = 0
	}
	_ = setBrightness(newB)
	m.brightness = newB
	return m, nil
}

func (m model) handleHDMIToggle() (tea.Model, tea.Cmd) {
	_ = toggleHDMI()
	m.hdmiActive = !m.hdmiActive
	return m, nil
}
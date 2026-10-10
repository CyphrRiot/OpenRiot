package control

import (
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

type refreshMsg struct{}

// run executes a possibly-blocking action off the UI goroutine, then asks
// for a state refresh so the panel shows the true result.
func run(fn func() error) tea.Cmd {
	return func() tea.Msg {
		_ = fn()
		return refreshMsg{}
	}
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
		return m.clampCursor(), nil
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
		return m.clampCursor(), nil
	case tickMsg:
		return m, tea.Batch(fetchVolumeCmd, fetchPowerCmd, fetchDisplayCmd, fetchSystemCmd, tickCmd())
	case refreshMsg:
		return m, tea.Batch(fetchVolumeCmd, fetchPowerCmd, fetchDisplayCmd, fetchSystemCmd)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m = m.clampCursor()

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

func (m model) clampCursor() model {
	n := m.itemsInSection()
	if n < 1 {
		n = 1
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	return m
}

func (m model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.section {
	case 0:
		if m.cursor == 1 {
			return m.handleMuteToggle()
		}
		return m, nil
	case 1:
		if m.cursor == 0 {
			return m.handleBrightnessAdjust(5)
		}
		if m.cursor == 1 {
			return m, run(toggleHDMI)
		}
		if m.cursor == 2 {
			return m, run(launchResolution)
		}
		return m, nil
	case 3:
		switch m.cursor {
		case 2:
			return m, run(suspendCmd)
		case 3:
			return m, run(rebootCmd)
		case 4:
			return m, run(shutdownCmd)
		case 5:
			return m, run(lockCmd)
		}
		return m, nil
	case 2:
		switch m.cursor {
		case 0:
			wifiAction()
		case 2:
			return m, run(toggleWireguard)
		}
		return m, nil
	case 4:
		switch m.cursor {
		case 0:
			return m, run(toggleNightlight)
		case 1:
			return m, run(toggleStealth)
		case 2:
			if m.proton {
				return m, run(triggerProtonSync)
			}
		}
		return m, nil
	case 5:
		switch m.cursor {
		case 0:
			return m, run(runUpdate)
		case 1:
			return m, run(toggleRecording)
		case 2:
			return m, run(takeScreenshot)
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
	m.volume = newVol
	return m, run(func() error { return setVolume(newVol) })
}

func (m model) handleMuteToggle() (tea.Model, tea.Cmd) {
	m.muted = !m.muted
	return m, run(toggleMute)
}

func (m model) handleBrightnessAdjust(delta int) (tea.Model, tea.Cmd) {
	newB := m.brightness + delta
	if newB > 100 {
		newB = 100
	}
	if newB < 0 {
		newB = 0
	}
	m.brightness = newB
	return m, run(func() error { return setBrightness(newB) })
}
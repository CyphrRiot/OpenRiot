package nmtui

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"openriot/wireguard"
)

// Messages
type wifiListMsg []WiFiAP

type connResultMsg struct {
	err error
	msg string
}

type activeInfoMsg struct {
	conn *ConnectionInfo
	err  error
}

type disconnectMsg struct {
	err error
	msg string
}

// Cmd generators
func fetchWifiNetworksCmd(iface string) tea.Cmd {
	return func() tea.Msg {
		aps, err := ScanWiFi(iface)
		if err != nil {
			return wifiListMsg(nil)
		}
		return wifiListMsg(aps)
	}
}

func connectToWifiCmd(iface, ssid, password string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if password == "" {
			err = ConnectOpen(iface, ssid)
		} else {
			err = ConnectWPA(iface, ssid, password)
		}
		if err != nil {
			return connResultMsg{err: err, msg: ""}
		}
		time.Sleep(3 * time.Second)
		ping := exec.Command("ping", "-c", "2", "-w", "5", "cdn.openbsd.org")
		if ping.Run() != nil {
			return connResultMsg{err: fmt.Errorf("no internet — wrong password or DHCP failure"), msg: ""}
		}
		if wireguard.IsRunning() {
			wireguard.Restart()
		}
		return connResultMsg{err: nil, msg: fmt.Sprintf("Connected to %s", ssid)}
	}
}

func disconnectCmd(iface string) tea.Cmd {
	return func() tea.Msg {
		err := Disconnect(iface)
		if err != nil {
			return disconnectMsg{err: err, msg: ""}
		}
		return disconnectMsg{err: nil, msg: "Disconnected"}
	}
}

func fetchActiveConnInfoCmd(iface string) tea.Cmd {
	return func() tea.Msg {
		conn, err := GetConnectionInfo(iface)
		return activeInfoMsg{conn: conn, err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

// Update handles all incoming messages and returns the next model and command.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.updateList(msg)
		case statePassword:
			return m.updatePassword(msg)
		case stateResult:
			return m.updateResult(msg)
		case stateActiveInfo:
			return m.updateActiveInfo(msg)
		case stateConfirmDisconnect:
			return m.updateConfirmDisconnect(msg)
		case stateConnecting:
			// Block input during connection
		}

	case tickMsg:
		// Auto-refresh active connection info
		cmds = append(cmds, fetchActiveConnInfoCmd(m.iface))

	case wifiListMsg:
		m.aps = msg
		m.scanning = false
		m.cursor = 0
		if len(m.aps) > 0 {
			m.conn = nil // will be refreshed by active info cmd
			cmds = append(cmds, fetchActiveConnInfoCmd(m.iface))
		}

	case connResultMsg:
		m.state = stateResult
		m.err = msg.err
		m.resultMsg = msg.msg
		if msg.err == nil {
			cmds = append(cmds, fetchActiveConnInfoCmd(m.iface))
		}

	case activeInfoMsg:
		if msg.err == nil {
			m.conn = msg.conn
		}

	case disconnectMsg:
		m.state = stateResult
		m.err = msg.err
		m.resultMsg = msg.msg
		if msg.err == nil {
			m.conn = nil
		}
	}

	// Always update spinner
	newModel, cmd := m.spinner.Update(msg)
	m.spinner = newModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Refresh):
		m.scanning = true
		return m, tea.Batch(m.spinner.Tick, fetchWifiNetworksCmd(m.iface))
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		return m, nil
	case key.Matches(msg, m.keys.Info):
		m.state = stateActiveInfo
		return m, fetchActiveConnInfoCmd(m.iface)
	case key.Matches(msg, m.keys.Disconnect):
		if m.conn != nil && m.conn.SSID != "" {
			m.state = stateConfirmDisconnect
		}
		return m, nil
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.aps)-1 {
			m.cursor++
		}
		return m, nil
	case key.Matches(msg, m.keys.Select):
		if len(m.aps) == 0 {
			return m, nil
		}
		ap := m.aps[m.cursor]
		if ap.Security == "open" {
			m.state = stateConnecting
			return m, connectToWifiCmd(m.iface, ap.SSID, "")
		}
		m.state = statePassword
		m.password.SetValue("")
		m.password.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m model) updatePassword(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.state = stateList
		return m, nil
	case key.Matches(msg, m.keys.Select):
		if len(m.aps) == 0 {
			return m, nil
		}
		ap := m.aps[m.cursor]
		pass := m.password.Value()
		m.state = stateConnecting
		return m, connectToWifiCmd(m.iface, ap.SSID, pass)
	}
	var cmd tea.Cmd
	m.password, cmd = m.password.Update(msg)
	return m, cmd
}

func (m model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Select):
		m.err = nil
		m.resultMsg = ""
		m.state = stateList
		return m, tea.Batch(fetchWifiNetworksCmd(m.iface), fetchActiveConnInfoCmd(m.iface))
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateActiveInfo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Select):
		m.state = stateList
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateConfirmDisconnect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.state = stateConnecting
		return m, disconnectCmd(m.iface)
	case "n", "N", "esc":
		m.state = stateList
		return m, nil
	}
	if key.Matches(msg, m.keys.Back) {
		m.state = stateList
		return m, nil
	}
	return m, nil
}

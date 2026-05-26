package disk

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Messages
type drivesMsg struct {
	drives []Drive
	err    error
}

type resultMsg struct {
	msg string
	err error
}

// Cmd generators
func fetchDrivesCmd() tea.Cmd {
	return func() tea.Msg {
		drives, err := DiscoverDrives()
		return drivesMsg{drives: drives, err: err}
	}
}

func mountCmd(device string) tea.Cmd {
	return func() tea.Msg {
		err := MountDrive(device, "")
		if err != nil {
			return resultMsg{err: err}
		}
		return resultMsg{msg: fmt.Sprintf("Mounted %s at /mnt/backup", device)}
	}
}

func umountCmd(device, mountPoint string) tea.Cmd {
	return func() tea.Msg {
		err := UmountDrive(device, mountPoint)
		if err != nil {
			return resultMsg{err: err}
		}
		return resultMsg{msg: fmt.Sprintf("Unmounted %s", device)}
	}
}

func formatCmd(device string) tea.Cmd {
	return func() tea.Msg {
		err := FormatDrive(device)
		if err != nil {
			return resultMsg{err: err}
		}
		return resultMsg{msg: fmt.Sprintf("Formatted %s with 4.2BSD", device)}
	}
}

func encryptCmd(device, passphrase string) tea.Cmd {
	return func() tea.Msg {
		err := EncryptDrive(device, passphrase)
		if err != nil {
			return resultMsg{err: err}
		}
		return resultMsg{msg: fmt.Sprintf("Encrypted %s with softraid", device)}
	}
}

func benchmarkCmd(device, writeSize, rwSize string) tea.Cmd {
	return func() tea.Msg {
		// Find mount point for device
		drives, _ := DiscoverDrives()
		mountPoint := "/mnt/backup"
		for _, d := range drives {
			if d.Device == device && d.MountPoint != "" {
				mountPoint = d.MountPoint
				break
			}
		}
		result, err := BenchmarkDrive(mountPoint, writeSize, rwSize)
		if err != nil {
			return resultMsg{err: err, msg: result}
		}
		return resultMsg{msg: result}
	}
}

// Update handles all incoming messages.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch m.state {
		case stateMenu:
			return m.updateMenu(msg)
		case stateDriveList:
			return m.updateDriveList(msg)
		case stateConfirm:
			return m.updateConfirm(msg)
		case statePassword:
			return m.updatePassword(msg)
		case stateBenchmarkConfig:
			return m.updateBenchmarkConfig(msg)
		case stateResult:
			return m.updateResult(msg)
		case stateRunning:
			// Block input during operation
		}

	case drivesMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateResult
		} else {
			m.drives = msg.drives
			m.filteredDrives = filterDrives(msg.drives, m.action)
			m.cursor = 0
			m.state = stateDriveList
		}

	case resultMsg:
		m.state = stateResult
		m.err = msg.err
		m.resultMsg = msg.msg
	}

	// Always update spinner
	newModel, cmd := m.spinner.Update(msg)
	m.spinner = newModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		return m, nil
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(menuItems)-1 {
			m.cursor++
		}
		return m, nil
	case key.Matches(msg, m.keys.Select):
		item := menuItems[m.cursor]
		m.action = item.action

		if item.action == actionDiscover {
			m.state = stateRunning
			return m, tea.Batch(m.spinner.Tick, fetchDrivesCmd())
		}

		// For all other actions, we need the drive list first
		m.state = stateRunning
		return m, tea.Batch(m.spinner.Tick, fetchDrivesCmd())
	}
	return m, nil
}

func (m model) updateDriveList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Back):
		m.state = stateMenu
		m.cursor = int(m.action)
		return m, nil
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.filteredDrives)-1 {
			m.cursor++
		}
		return m, nil
	case key.Matches(msg, m.keys.Select):
		if len(m.filteredDrives) == 0 || m.action == actionDiscover {
			return m, nil
		}
		d := m.filteredDrives[m.cursor]

		// Safety checks
		if d.IsRoot && m.action != actionBenchmark {
			m.err = fmt.Errorf("cannot %s root drive %s", actionName(m.action), d.Device)
			m.state = stateResult
			return m, nil
		}
		if d.IsChunk {
			m.err = fmt.Errorf("cannot %s softraid chunk %s (use the virtual device)", actionName(m.action), d.Device)
			m.state = stateResult
			return m, nil
		}

		switch m.action {
		case actionMount:
			if d.IsMounted {
				m.err = fmt.Errorf("%s is already mounted at %s", d.Device, d.MountPoint)
				m.state = stateResult
				return m, nil
			}
			m.state = stateRunning
			return m, tea.Batch(m.spinner.Tick, mountCmd(d.Device))

		case actionUmount:
			if !d.IsMounted && !d.IsEncrypted {
				m.err = fmt.Errorf("%s is not mounted", d.Device)
				m.state = stateResult
				return m, nil
			}
			m.state = stateRunning
			return m, tea.Batch(m.spinner.Tick, umountCmd(d.Device, d.MountPoint))

		case actionFormat:
			if d.IsMounted {
				m.err = fmt.Errorf("unmount %s before formatting", d.Device)
				m.state = stateResult
				return m, nil
			}
			m.confirmMsg = fmt.Sprintf("Format %s (%d GB)? All data will be lost.", d.Device, d.SizeGB)
			m.state = stateConfirm
			return m, nil

		case actionEncrypt:
			if d.IsMounted {
				m.err = fmt.Errorf("unmount %s before encrypting", d.Device)
				m.state = stateResult
				return m, nil
			}
			m.confirmMsg = fmt.Sprintf("Encrypt %s (%d GB)? All data will be lost.", d.Device, d.SizeGB)
			m.state = stateConfirm
			return m, nil

		case actionBenchmark:
			if !d.IsMounted {
				m.err = fmt.Errorf("mount %s before benchmarking", d.Device)
				m.state = stateResult
				return m, nil
			}
			m.benchCursor = 0
			m.state = stateBenchmarkConfig
			return m, nil
		}
	}
	return m, nil
}

func (m model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Yes):
		if m.cursor < 0 || m.cursor >= len(m.filteredDrives) {
			return m, nil
		}
		d := m.filteredDrives[m.cursor]

		switch m.action {
		case actionFormat:
			m.state = stateRunning
			return m, tea.Batch(m.spinner.Tick, formatCmd(d.Device))
		case actionEncrypt:
			m.password.SetValue("")
			m.password.Focus()
			m.state = statePassword
			return m, textinput.Blink
		}

	case key.Matches(msg, m.keys.No), key.Matches(msg, m.keys.Back):
		m.state = stateDriveList
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updatePassword(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.state = stateDriveList
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Select):
		passphrase := m.password.Value()
		if passphrase == "" {
			return m, nil
		}
		if m.cursor < 0 || m.cursor >= len(m.filteredDrives) {
			return m, nil
		}
		d := m.filteredDrives[m.cursor]
		m.state = stateRunning
		return m, tea.Batch(m.spinner.Tick, encryptCmd(d.Device, passphrase))
	}

	// Pass through to password input
	var cmd tea.Cmd
	newModel, cmd := m.password.Update(msg)
	m.password = newModel
	return m, cmd
}

func (m model) updateBenchmarkConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Back):
		m.state = stateDriveList
		return m, nil
	case key.Matches(msg, m.keys.Up):
		if m.benchCursor > 0 {
			m.benchCursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if m.benchCursor < len(benchmarkPresets)-1 {
			m.benchCursor++
		}
		return m, nil
	case key.Matches(msg, m.keys.Select):
		if m.cursor < 0 || m.cursor >= len(m.filteredDrives) {
			return m, nil
		}
		d := m.filteredDrives[m.cursor]
		preset := benchmarkPresets[m.benchCursor]
		m.state = stateRunning
		return m, tea.Batch(m.spinner.Tick, benchmarkCmd(d.Device, preset.writeSize, preset.rwSize))
	}
	return m, nil
}

func (m model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.state = stateMenu
		m.cursor = int(m.action)
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func actionName(a actionType) string {
	switch a {
	case actionMount:
		return "mount"
	case actionUmount:
		return "umount"
	case actionFormat:
		return "format"
	case actionEncrypt:
		return "encrypt"
	case actionBenchmark:
		return "benchmark"
	default:
		return "use"
	}
}

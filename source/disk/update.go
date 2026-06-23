package disk

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

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

func mountCmd(device, passphrase string) tea.Cmd {
	return func() tea.Msg {
		var log strings.Builder
		mountPoint := "/mnt/backup"
		exec.Command("doas", "-n", "mkdir", "-p", mountPoint).Run()

		raidDev := findRaidDevice(device)
		hasRAID := hasRAIDPartition(device)

		if raidDev == "" && hasRAID {
			fmt.Fprintln(&log, "Attaching encrypted volume...")
			passFile := "/tmp/.openriot_" + device
			exec.Command("doas", "-n", "sh", "-c",
				fmt.Sprintf("printf '%%s\\n' '%s' > %s && chmod 600 %s",
					passphrase, passFile, passFile)).Run()
			defer exec.Command("doas", "-n", "rm", passFile).Run()
			cmd := exec.Command("doas", "-n", "bioctl", "-c", "C",
				"-p", passFile, "-l", device+"a", "softraid0")
			out, err := cmd.CombinedOutput()
			if err != nil {
				// Chunk may already be attached — find the virtual device
				raidDev = findRaidDevice(device)
				if raidDev == "" {
					return resultMsg{err: fmt.Errorf(
						"bioctl attach failed: %w\n%s", err, string(out))}
				}
				fmt.Fprintf(&log, "Already attached as %s\n", raidDev)
			} else {
				log.Write(out)
				time.Sleep(500 * time.Millisecond)
				for _, tok := range strings.Fields(string(out)) {
					tok = strings.TrimSuffix(tok, ":")
					if strings.HasPrefix(tok, "sd") && tok != device {
						raidDev = tok
						break
					}
				}
				if raidDev == "" {
					raidDev = findRaidDevice(device)
				}
			}
		}

		// If no raid device found and no RAID partition, mount directly
		if raidDev == "" && !hasRAID {
			raidDev = device
		}

		if raidDev == "" {
			return resultMsg{
				err: fmt.Errorf("no mountable device for %s", device)}
		}

		source := "/dev/" + raidDev + "a"
		fmt.Fprintf(&log, "Mounting %s at %s\n", source, mountPoint)
		cmd := exec.Command("doas", "-n", "mount", source, mountPoint)
		if out, err := cmd.CombinedOutput(); err != nil {
			return resultMsg{err: fmt.Errorf("mount failed: %w\n%s",
				err, string(out))}
		}
		return resultMsg{
			msg: fmt.Sprintf("Mounted %s at %s\n%s", device, mountPoint, log.String())}
	}
}

func umountCmd(device, mountPoint string) tea.Cmd {
	return func() tea.Msg {
		var log strings.Builder
		if mountPoint == "" {
			mountPoint = "/mnt/backup"
		}
		if isMountedAt(mountPoint) {
			fmt.Fprintf(&log, "Unmounting %s\n", mountPoint)
			cmd := exec.Command("doas", "-n", "umount", mountPoint)
			if out, err := cmd.CombinedOutput(); err != nil {
				return resultMsg{err: fmt.Errorf("umount failed: %w\n%s",
					err, string(out))}
			}
			exec.Command("doas", "-n", "sync").Run()
		}
		detachSoftraidChunk(device, &log)
		return resultMsg{
			msg: fmt.Sprintf("Unmounted %s\n%s", device, log.String())}
	}
}

func formatCmd(device string) tea.Cmd {
	return func() tea.Msg {
		var log strings.Builder

		// Detach any existing softraid volume for this chunk
		detachSoftraidChunk(device, &log)

		fmt.Fprintf(&log, "Wiping /dev/r%sc ...\n", device)
		cmd := exec.Command("doas", "-n", "dd", "if=/dev/zero",
			"of=/dev/r"+device+"c", "bs=1m", "count=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			return resultMsg{err: fmt.Errorf("dd failed: %w\n%s",
				err, string(out))}
		}
		fmt.Fprint(&log, "done\n")

		fmt.Fprintf(&log, "Creating 4.2BSD partition on %s ...\n", device)
		labelScript := "a a\n\n\n4.2BSD\nw\nq\n"
		cmd = exec.Command("doas", "-n", "disklabel", "-E", device)
		cmd.Stdin = strings.NewReader(labelScript)
		if out, err := cmd.CombinedOutput(); err != nil {
			return resultMsg{err: fmt.Errorf("disklabel failed: %w\n%s",
				err, string(out))}
		}
		fmt.Fprint(&log, "done\n")

		fmt.Fprintf(&log, "newfs /dev/r%sa ...\n", device)
		cmd = exec.Command("doas", "-n", "newfs", "/dev/r"+device+"a")
		if out, err := cmd.CombinedOutput(); err != nil {
			return resultMsg{err: fmt.Errorf("newfs failed: %w\n%s",
				err, string(out))}
		}
		return resultMsg{
			msg: fmt.Sprintf("Formatted %s with 4.2BSD\n%s", device, log.String())}
	}
}

func encryptCmd(device, passphrase string) tea.Cmd {
	return func() tea.Msg {
		var log strings.Builder
		var raidDev string

		// Detach any existing softraid volume for this chunk
		detachSoftraidChunk(device, &log)

		// Step 1: dd
		fmt.Fprintf(&log, "Wiping /dev/r%sc ...\n", device)
		cmd := exec.Command("doas", "-n", "dd", "if=/dev/zero",
			"of=/dev/r"+device+"c", "bs=1m", "count=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			return resultMsg{err: fmt.Errorf("dd failed: %w\n%s",
				err, string(out))}
		}
		fmt.Fprint(&log, "done\n")

		// Step 2: RAID partition
		fmt.Fprintf(&log, "Creating RAID partition on %s ...\n", device)
		labelScript := "a a\n\n\nRAID\nw\nq\n"
		cmd = exec.Command("doas", "-n", "disklabel", "-E", device)
		cmd.Stdin = strings.NewReader(labelScript)
		if out, err := cmd.CombinedOutput(); err != nil {
			return resultMsg{err: fmt.Errorf("disklabel failed: %w\n%s",
				err, string(out))}
		}
		fmt.Fprint(&log, "done\n")

		// Step 3: bioctl crypto
		fmt.Fprintf(&log, "Setting up softraid crypto on %sa ...\n", device)
		passFile := "/tmp/.openriot_" + device
		exec.Command("doas", "-n", "sh", "-c",
			fmt.Sprintf("printf '%%s\\n%%s\\n' '%s' '%s' > %s && chmod 600 %s",
				passphrase, passphrase, passFile, passFile)).Run()
		defer exec.Command("doas", "-n", "rm", "-f", passFile).Run()
		cmd = exec.Command("doas", "-n", "bioctl", "-c", "C",
			"-p", passFile, "-l", device+"a", "softraid0")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return resultMsg{err: fmt.Errorf("bioctl failed: %w\n%s",
				err, string(out))}
		}
		log.Write(out)
		log.WriteString("\n")

		raidDev = ""
		for _, tok := range strings.Fields(string(out)) {
			tok = strings.TrimSuffix(tok, ":")
			if strings.HasPrefix(tok, "sd") && tok != device {
				raidDev = tok
				break
			}
		}
		if raidDev == "" {
			raidDev = findRaidDevice(device)
		}
		if raidDev == "" {
			return resultMsg{err: fmt.Errorf(
				"virtual device not found:\n%s", string(out))}
		}
		fmt.Fprintf(&log, "Virtual device: %s\n", raidDev)
		time.Sleep(1 * time.Second)

		// Step 4: dd virtual device
		fmt.Fprintf(&log, "Wiping /dev/r%sc ...\n", raidDev)
		exec.Command("doas", "-n", "dd", "if=/dev/zero",
			"of=/dev/r"+raidDev+"c", "bs=1m", "count=1").Run()
		fmt.Fprint(&log, "done\n")

		// Step 5: fdisk
		fmt.Fprintf(&log, "Writing partition table on %s ...\n", raidDev)
		exec.Command("doas", "-n", "/sbin/fdisk", "-iy", raidDev).Run()
		fmt.Fprint(&log, "done\n")

		// Step 6: disklabel 4.2BSD on virtual
		fmt.Fprintf(&log, "Creating 4.2BSD partition on %s ...\n", raidDev)
		labelScript = "a a\n\n\n4.2BSD\nw\nq\n"
		cmd = exec.Command("doas", "-n", "disklabel", "-E", raidDev)
		cmd.Stdin = strings.NewReader(labelScript)
		if out, err := cmd.CombinedOutput(); err != nil {
			return resultMsg{err: fmt.Errorf("disklabel failed: %w\n%s",
				err, string(out))}
		}
		fmt.Fprint(&log, "done\n")

		// Step 7: newfs
		fmt.Fprintf(&log, "newfs /dev/r%sa ...\n", raidDev)
		cmd = exec.Command("doas", "-n", "newfs", "/dev/r"+raidDev+"a")
		if out, err := cmd.CombinedOutput(); err != nil {
			return resultMsg{err: fmt.Errorf("newfs failed: %w\n%s",
				err, string(out))}
		}
		return resultMsg{
			msg: fmt.Sprintf("Encrypted %s with softraid\n%s",
				device, log.String())}
	}
}

func benchmarkCmd(device, writeSize, rwSize string) tea.Cmd {
	return func() tea.Msg {
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
		case stateConfirmPassword:
			return m.updateConfirmPassword(msg)
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

		// For all other actions, we need the drive list — use cached drives if available
		if len(m.drives) > 0 {
			m.filteredDrives = filterDrives(m.drives, m.action)
			m.cursor = 0
			m.state = stateDriveList
			return m, nil
		}
		m.state = stateRunning
		m.filteredDrives = nil
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
		if d.IsChunk && !d.IsRemovable && m.action != actionDiscover {
			m.err = fmt.Errorf("cannot %s internal softraid drive %s", actionName(m.action), d.Device)
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
			if hasRAIDPartition(d.Device) && findRaidDevice(d.Device) == "" {
				m.password.SetValue("")
				m.password.Focus()
				m.confirmMsg = fmt.Sprintf("Mount encrypted %s — enter passphrase:", d.Device)
				m.state = statePassword
				return m, textinput.Blink
			}
			m.state = stateRunning
			return m, tea.Batch(m.spinner.Tick, mountCmd(d.Device, ""))

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
		m.password.SetValue("")
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

		if m.action == actionMount {
			m.state = stateRunning
			return m, tea.Batch(m.spinner.Tick, mountCmd(d.Device, passphrase))
		}

		// Encrypt: go to confirmation
		m.confirmPassword.SetValue("")
		m.confirmPassword.Focus()
		m.state = stateConfirmPassword
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	newModel, cmd := m.password.Update(msg)
	m.password = newModel
	return m, cmd
}

func (m model) updateConfirmPassword(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.password.Focus()
		m.state = statePassword
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Select):
		if m.confirmPassword.Value() == "" {
			return m, nil
		}
		if m.password.Value() != m.confirmPassword.Value() {
			m.err = fmt.Errorf("passphrases do not match")
			m.state = stateResult
			return m, nil
		}
		if m.cursor < 0 || m.cursor >= len(m.filteredDrives) {
			return m, nil
		}
		d := m.filteredDrives[m.cursor]
		m.state = stateRunning
		return m, tea.Batch(m.spinner.Tick, encryptCmd(d.Device, m.password.Value()))
	}

	var cmd tea.Cmd
	newModel, cmd := m.confirmPassword.Update(msg)
	m.confirmPassword = newModel
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

// detachSoftraidChunk tries to detach any softraid volume backed by the given
// chunk device. It first tries parseSoftraid, then falls back to scanning
// dmesg for virtual devices at the softraid bus.
func detachSoftraidChunk(device string, log *strings.Builder) {
	// Method 1: try bioctl -d on the virtual device from parseSoftraid
	raidDev := findRaidDevice(device)
	if raidDev != "" {
		fmt.Fprintf(log, "Detaching %s ...\n", raidDev)
		time.Sleep(500 * time.Millisecond)
		for attempt := 0; attempt < 3; attempt++ {
			if out, err := tryDetach(raidDev); err == nil {
				fmt.Fprintf(log, "Detached\n%s\n", out)
				return
			}
			time.Sleep(250 * time.Millisecond)
		}
	}

	// Method 2: scan dmesg for all virtual devices and try each one
	dmesgOut, err := exec.Command("dmesg").Output()
	if err != nil {
		return
	}
	softraidBus := ""
	for _, line := range strings.Split(string(dmesgOut), "\n") {
		if strings.Contains(line, " at softraid") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				softraidBus = fields[0]
				break
			}
		}
	}
	if softraidBus == "" {
		return
	}
	var lastErr error
	found := false
	for _, line := range strings.Split(string(dmesgOut), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "sd") || !strings.Contains(line, " at "+softraidBus) {
			continue
		}
		virtDev := strings.Fields(line)[0]
		if virtDev == "" {
			continue
		}
		found = true
		if out, err := tryDetach(virtDev); err == nil {
			fmt.Fprintf(log, "Detached %s\n%s\n", virtDev, out)
			return
		} else {
			lastErr = err
		}
	}
	if found {
		msg := "Warning: could not detach softraid volume"
		if lastErr != nil {
			msg += ": " + lastErr.Error()
		}
		fmt.Fprintln(log, msg)
	}
}

// tryDetach attempts bioctl -d first without doas (operator group),
// then with doas. Returns combined output on success, error on failure.
func tryDetach(device string) (string, error) {
	cmd := exec.Command("bioctl", "-d", device)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), nil
	}
	cmd = exec.Command("doas", "-n", "bioctl", "-d", device)
	out, err = cmd.CombinedOutput()
	if err == nil {
		return string(out), nil
	}
	return "", err
}

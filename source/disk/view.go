package disk

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"openriot/theme"
)

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	if m.err != nil && m.state == stateResult {
		return m.renderResult()
	}

	switch m.state {
	case stateMenu:
		return m.renderMenu()
	case stateDriveList:
		return m.renderDriveList()
	case stateConfirm:
		return m.renderConfirm()
	case statePassword:
		return m.renderPassword()
	case stateBenchmarkConfig:
		return m.renderBenchmarkConfig()
	case stateRunning:
		return m.renderRunning()
	case stateResult:
		return m.renderResult()
	default:
		return "unknown state"
	}
}

func (m model) renderHelp() string {
	return lipgloss.NewStyle().
		Padding(2).
		Render(m.help.View(m.keys))
}

func (m model) renderMenu() string {
	var b strings.Builder

	headerStyle := theme.Lipgloss.Accent.Header
	dimStyle := theme.Lipgloss.Base.Dim
	helpStyle := theme.Lipgloss.Base.FG3
	selectedStyle := theme.Lipgloss.Accent.Selected
	destructiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Semantic.Error))
	descStyle := theme.Lipgloss.Base.Dim
	nameWidth := 17

	b.WriteString(headerStyle.Render(" 󰋊  Disk Manager "))
	b.WriteString("\n\n")

	windowSize := m.height - 6
	if windowSize < 1 {
		windowSize = 1
	}
	start := m.cursor - windowSize/2
	if start < 0 {
		start = 0
	}
	end := start + windowSize
	if end > len(menuItems) {
		end = len(menuItems)
	}
	if end-start < windowSize && start > 0 {
		start = end - windowSize
		if start < 0 {
			start = 0
		}
	}

	if start > 0 {
		b.WriteString(dimStyle.Render("  ▲ ..."))
		b.WriteString("\n")
	}

	for i := start; i < end; i++ {
		item := menuItems[i]
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
		}

		var nameStr string
		if item.destructive {
			nameStr = destructiveStyle.Render(item.name)
		} else if m.cursor == i {
			nameStr = selectedStyle.Render(item.name)
		} else {
			nameStr = item.name
		}

		// Pad name to fixed visual width using lipgloss
		paddedName := lipgloss.NewStyle().Width(nameWidth).Render(nameStr)

		line := cursor + paddedName + "  " + descStyle.Render(item.description)
		b.WriteString(line)
		b.WriteString("\n")
	}

	if end < len(menuItems) {
		b.WriteString(dimStyle.Render("  ▼ ..."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter: select | q: quit | ?: help"))
	return b.String()
}

func (m model) renderDriveList() string {
	var b strings.Builder

	headerStyle := theme.Lipgloss.Accent.Header
	dimStyle := theme.Lipgloss.Base.Dim
	helpStyle := theme.Lipgloss.Base.FG3
	selectedStyle := theme.Lipgloss.Accent.Selected
	rootStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Semantic.Error))
	mountedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Semantic.Warning))
	encryptedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Semantic.Info))
	chunkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Base.FG3))
	availableStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Semantic.Success))

	actionName := "Select Drive"
	switch m.action {
	case actionMount:
		actionName = "Mount"
	case actionUmount:
		actionName = "Umount"
	case actionFormat:
		actionName = "Format"
	case actionEncrypt:
		actionName = "Encrypt"
	case actionBenchmark:
		actionName = "Benchmark"
	}

	b.WriteString(headerStyle.Render(fmt.Sprintf(" %s ", actionName)))
	b.WriteString("\n\n")

	if len(m.filteredDrives) == 0 {
		b.WriteString(dimStyle.Render("  No eligible drives for this action."))
		return b.String()
	}

	windowSize := m.height - 6
	if windowSize < 1 {
		windowSize = 1
	}
	start := m.cursor - windowSize/2
	if start < 0 {
		start = 0
	}
	end := start + windowSize
	if end > len(m.filteredDrives) {
		end = len(m.filteredDrives)
	}
	if end-start < windowSize && start > 0 {
		start = end - windowSize
		if start < 0 {
			start = 0
		}
	}

	if start > 0 {
		b.WriteString(dimStyle.Render("  ▲ ..."))
		b.WriteString("\n")
	}

	deviceWidth := 8
	statusWidth := 13
	for i := start; i < end; i++ {
		d := m.filteredDrives[i]
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
		}

		var statusStr string
		var statusStyle lipgloss.Style
		switch {
		case d.IsRoot:
			statusStr = "[ROOT]"
			statusStyle = rootStyle
		case d.IsMounted:
			statusStr = "[MOUNTED]"
			statusStyle = mountedStyle
		case d.IsChunk:
			statusStr = "[CHUNK]"
			statusStyle = chunkStyle
		case d.IsEncrypted:
			statusStr = "[ENCRYPTED]"
			statusStyle = encryptedStyle
		default:
			statusStr = "[AVAILABLE]"
			statusStyle = availableStyle
		}

		nameStr := d.Device
		if m.cursor == i {
			nameStr = selectedStyle.Render(d.Device)
		}

		nameCol := lipgloss.NewStyle().Width(deviceWidth).Render(nameStr)
		statusCol := lipgloss.NewStyle().Width(statusWidth).Render(statusStyle.Render(statusStr))
		sizeStr := dimStyle.Render(fmt.Sprintf("%d GB", d.SizeGB))

		var extra string
		if d.MountPoint != "" {
			extra = dimStyle.Render(fmt.Sprintf(" → %s", d.MountPoint))
		}
		if d.IsRemovable {
			extra += dimStyle.Render(" [USB]")
		}

		line := fmt.Sprintf("%s%s %s %s%s", cursor, nameCol, statusCol, sizeStr, extra)
		b.WriteString(line)
		b.WriteString("\n")
	}

	if end < len(m.filteredDrives) {
		b.WriteString(dimStyle.Render("  ▼ ..."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.action == actionDiscover {
		b.WriteString(helpStyle.Render("esc: back | q: quit"))
	} else {
		b.WriteString(helpStyle.Render("enter: select | esc: back | q: quit"))
	}
	return b.String()
}

func (m model) renderConfirm() string {
	var b strings.Builder

	titleStyle := theme.Lipgloss.Accent.Title
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Semantic.Error)).Bold(true)
	dimStyle := theme.Lipgloss.Base.Dim
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)

	b.WriteString(titleStyle.Render("⚠️  Confirm Action"))
	b.WriteString("\n\n")

	var content string
	content += warningStyle.Render("This action is DESTRUCTIVE and cannot be undone.") + "\n\n"
	content += m.confirmMsg + "\n\n"
	content += dimStyle.Render("Press y to confirm, n to cancel")

	centered := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
	b.WriteString(centered)
	return b.String()
}

func (m model) renderPassword() string {
	var b strings.Builder

	titleStyle := theme.Lipgloss.Accent.Title
	dimStyle := theme.Lipgloss.Base.Dim
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)

	b.WriteString(titleStyle.Render("🔐  Encrypt Drive"))
	b.WriteString("\n\n")

	var content string
	content += "Enter passphrase for softraid encryption:\n\n"
	content += m.password.View() + "\n\n"
	content += dimStyle.Render("enter: confirm | esc: cancel")

	centered := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
	b.WriteString(centered)
	return b.String()
}

func (m model) renderBenchmarkConfig() string {
	var b strings.Builder

	headerStyle := theme.Lipgloss.Accent.Header
	selectedStyle := theme.Lipgloss.Accent.Selected
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Base.FG3))
	dimStyle := theme.Lipgloss.Base.Dim

	b.WriteString(headerStyle.Render(" Benchmark Duration "))
	b.WriteString("\n\n")

	for i, preset := range benchmarkPresets {
		cursor := "  "
		var nameStr string
		if m.benchCursor == i {
			cursor = "▸ "
			nameStr = selectedStyle.Render(preset.name)
		} else {
			nameStr = preset.name
		}
		padded := lipgloss.NewStyle().Width(10).Render(nameStr)
		line := cursor + padded + "  " + descStyle.Render(preset.desc)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("enter: select | esc: back | q: quit"))
	return b.String()
}

func (m model) renderRunning() string {
	var b strings.Builder

	titleStyle := theme.Lipgloss.Accent.Title

	b.WriteString(titleStyle.Render("Running..."))
	b.WriteString("\n\n")
	b.WriteString(m.spinner.View())
	b.WriteString(" ")

	var action string
	switch m.action {
	case actionMount:
		action = "Mounting drive"
	case actionUmount:
		action = "Unmounting drive"
	case actionFormat:
		action = "Formatting drive"
	case actionEncrypt:
		action = "Encrypting drive"
	case actionBenchmark:
		action = "Running benchmark"
	default:
		action = "Working"
	}

	b.WriteString(action)
	return b.String()
}

func (m model) renderResult() string {
	var b strings.Builder

	titleStyle := theme.Lipgloss.Accent.Title
	successStyle := theme.Lipgloss.Semantic.Success
	errorStyle := theme.Lipgloss.Semantic.Error
	dimStyle := theme.Lipgloss.Base.Dim

	if m.err != nil {
		b.WriteString(titleStyle.Render("❌ Error"))
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	} else {
		b.WriteString(titleStyle.Render("✅ Success"))
		b.WriteString("\n\n")
		b.WriteString(successStyle.Render(m.resultMsg))
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("esc: back | q: quit"))
	return b.String()
}

package nmtui

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

	switch m.state {
	case stateList:
		return m.listView()
	case statePassword:
		return m.passwordView()
	case stateConnecting:
		return m.connectingView()
	case stateResult:
		return m.resultView()
	case stateActiveInfo:
		return m.activeInfoView()
	case stateConfirmDisconnect:
		return m.confirmDisconnectView()
	default:
		return "unknown state"
	}
}

func (m model) listView() string {
	var b strings.Builder

	// Styles
	dimStyle := theme.Lipgloss.Base.Dim
	successStyle := theme.Lipgloss.Semantic.Success
	headerStyle := theme.Lipgloss.Accent.Header
	helpStyle := theme.Lipgloss.Base.FG3
	itemStyle := lipgloss.NewStyle()
	selectedStyle := theme.Lipgloss.Accent.Selected

	// Header
	header := headerStyle.Render(fmt.Sprintf(" %s ", m.iface))
	if m.conn != nil && m.conn.SSID != "" {
		header += " " + successStyle.Render("● "+m.conn.SSID)
	} else {
		header += " " + dimStyle.Render("○ disconnected")
	}
	if m.scanning {
		header += " " + m.spinner.View()
	}
	b.WriteString(headerStyle.Width(m.width).Render(header))
	b.WriteString("\n\n")

	// Network list
	if len(m.aps) == 0 {
		if m.scanning {
			b.WriteString(dimStyle.Render("  Scanning..."))
		} else {
			b.WriteString(dimStyle.Render("  No networks found. Press r to scan."))
		}
	} else {
		windowSize := m.height - 4
		if windowSize < 1 {
			windowSize = 1
		}
		start := m.cursor - windowSize/2
		if start < 0 {
			start = 0
		}
		end := start + windowSize
		if end > len(m.aps) {
			end = len(m.aps)
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
			ap := m.aps[i]
			cursor := "  "
			style := itemStyle
			if m.cursor == i {
				cursor = "▸ "
				style = selectedStyle
			}

			pct := SignalToPercent(ap.Signal, ap.SignalValid)
			sigBar := renderSignalBar(pct)
			pctStr := "  --%"
			if pct >= 0 {
				pctStr = fmt.Sprintf("%3d%%", pct)
			}

			var secColor string
			switch ap.Security {
			case "wpa3":
				secColor = theme.GetSecColor(0)
			case "wpa2":
				secColor = theme.GetSecColor(1)
			case "wep", "open", "":
				secColor = theme.GetSecColor(2)
			default:
				secColor = theme.GetSecColor(3)
			}
			secLabel := fmt.Sprintf("[%s]", strings.ToUpper(ap.Security))
			secStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(secColor)).Render(secLabel)

			// Fixed-width columns: cursor(2) + SSID(22) + gap(1) + bar(10) + gap(1) + pct(5) + gap(1) + sec(6+)
			ssidCol := style.Render(lipgloss.NewStyle().Width(22).Render(truncateSSID(ap.SSID, 22)))
			line := fmt.Sprintf("%s%s %s %s %s",
				cursor,
				ssidCol,
				dimStyle.Render(sigBar),
				dimStyle.Render(pctStr),
				secStyled,
			)
			b.WriteString(line)
			b.WriteString("\n")
		}

		if end < len(m.aps) {
			b.WriteString(dimStyle.Render("  ▼ ..."))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Help / status bar
	if m.showHelp {
		b.WriteString(m.help.View(m.keys))
	} else {
		b.WriteString(helpStyle.Render("? help  q quit  r refresh  i info  d disconnect"))
	}

	return b.String()
}

func (m model) passwordView() string {
	titleStyle := theme.Lipgloss.Accent.Title
	dimStyle := theme.Lipgloss.Base.Dim
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	ssid := ""
	if m.cursor >= 0 && m.cursor < len(m.aps) {
		ssid = m.aps[m.cursor].SSID
	}

	title := fmt.Sprintf("Connect to %s", ssid)
	content := titleStyle.Render(title) + "\n\n" + m.password.View() + "\n\n" + dimStyle.Render("enter to confirm  esc to cancel")
	centered := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
	b.WriteString(centered)
	return b.String()
}

func (m model) connectingView() string {
	titleStyle := theme.Lipgloss.Accent.Title
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	ssid := ""
	if m.cursor >= 0 && m.cursor < len(m.aps) {
		ssid = m.aps[m.cursor].SSID
	}

	msg := fmt.Sprintf("Connecting to %s...", ssid)
	content := m.spinner.View() + " " + titleStyle.Render(msg)
	centered := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, content)
	b.WriteString(centered)
	return b.String()
}

func (m model) resultView() string {
	errStyle := theme.Lipgloss.Semantic.Error
	successStyle := theme.Lipgloss.Semantic.Success
	dimStyle := theme.Lipgloss.Base.Dim
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	var content string
	if m.err != nil {
		content = errStyle.Render("Error") + "\n\n" + m.err.Error()
	} else {
		content = successStyle.Render("Success") + "\n\n" + m.resultMsg
	}
	content += "\n\n" + dimStyle.Render("esc to continue")
	centered := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
	b.WriteString(centered)
	return b.String()
}

func (m model) activeInfoView() string {
	infoLabelStyle := theme.Lipgloss.Accent.InfoLabel
	infoValueStyle := theme.Lipgloss.Accent.InfoValue
	dimStyle := theme.Lipgloss.Base.Dim
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	if m.conn == nil {
		centered := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, dimStyle.Render("No active connection"))
		b.WriteString(centered)
		b.WriteString("\n" + dimStyle.Render("esc to go back"))
		return b.String()
	}

	rows := []struct {
		label string
		value string
	}{
		{"Device", m.conn.Device},
		{"SSID", m.conn.SSID},
		{"IP", m.conn.IP},
		{"Netmask", m.conn.Netmask},
		{"Gateway", m.conn.Gateway},
		{"MAC", m.conn.MAC},
		{"State", m.conn.State},
	}

	var info strings.Builder
	for _, r := range rows {
		info.WriteString(fmt.Sprintf("%s %s\n", infoLabelStyle.Render(r.label+":"), infoValueStyle.Render(r.value)))
	}
	if len(m.conn.DNS) > 0 {
		info.WriteString(fmt.Sprintf("%s %s\n", infoLabelStyle.Render("DNS:"), infoValueStyle.Render(strings.Join(m.conn.DNS, ", "))))
	}

	centered := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, boxStyle.Render(info.String()))
	b.WriteString(centered)
	b.WriteString("\n" + dimStyle.Render("esc to go back"))
	return b.String()
}

func (m model) confirmDisconnectView() string {
	titleStyle := theme.Lipgloss.Accent.Title
	dimStyle := theme.Lipgloss.Base.Dim
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	ssid := ""
	if m.conn != nil {
		ssid = m.conn.SSID
	}

	msg := fmt.Sprintf("Disconnect from %s?", ssid)
	content := titleStyle.Render(msg) + "\n\n" + dimStyle.Render("y to confirm  n/esc to cancel")
	centered := lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
	b.WriteString(centered)
	return b.String()
}

func (m model) renderHeader() string {
	dimStyle := theme.Lipgloss.Base.Dim
	successStyle := theme.Lipgloss.Semantic.Success
	headerStyle := theme.Lipgloss.Accent.Header

	h := headerStyle.Render(fmt.Sprintf(" %s ", m.iface))
	if m.conn != nil && m.conn.SSID != "" {
		h += " " + successStyle.Render("● "+m.conn.SSID)
	} else {
		h += " " + dimStyle.Render("○ disconnected")
	}
	return headerStyle.Width(m.width).Render(h)
}

func renderSignalBar(percent int) string {
	const total = 10
	if percent < 0 {
		return strings.Repeat("○", total)
	}
	filled := (percent * total) / 100
	if filled < 0 {
		filled = 0
	}
	if filled > total {
		filled = total
	}
	return strings.Repeat("●", filled) + strings.Repeat("○", total-filled)
}

func truncateSSID(s string, max int) string {
	if lipgloss.Width(s) <= max {
		return s
	}
	return lipgloss.NewStyle().MaxWidth(max).Render(s)
}

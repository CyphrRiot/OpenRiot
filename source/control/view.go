package control

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"openriot/theme"
)

var (
	accentColor = theme.Palette.Accent.FG
	dimColor    = theme.Palette.Base.FG3
	helpFg      = theme.Palette.Base.FG3
	sidebarFg   = theme.Palette.Base.FG2
	sidebarBg   = theme.Palette.Base.BG2
	panelBg     = theme.Palette.Base.BG
	sectionW    = 14
	panelW      = 46
	panelHeight = 11

	styleSidebarSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(accentColor)).
				Background(lipgloss.Color(sidebarBg)).
				Padding(0, 1)

	styleSidebarInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color(sidebarFg)).
				Background(lipgloss.Color(sidebarBg)).
				Padding(0, 1)

	styleSidebarContainer = lipgloss.NewStyle().
				Width(sectionW).
				Background(lipgloss.Color(sidebarBg)).
				Height(panelHeight)

	stylePanelContainer = lipgloss.NewStyle().
				Width(panelW).
				Background(lipgloss.Color(panelBg)).
				Height(panelHeight)

	stylePanelTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(accentColor)).
			PaddingLeft(1)

	styleRowSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(accentColor)).
			Background(lipgloss.Color(sidebarBg)).
			PaddingLeft(1)

	styleRowInactive = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dimColor)).
			PaddingLeft(1)

	styleFooter = lipgloss.NewStyle().
			Foreground(lipgloss.Color(helpFg)).
			PaddingLeft(1)

	styleBeta = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color(dimColor)).
			PaddingLeft(1)
)

func (m model) View() string {
	if m.showHelp {
		return "\n" + m.help.View(m.keys) + "\n"
	}

	if m.width == 0 {
		return "Initializing..."
	}

	panel := m.renderPanel()
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styleSidebarContainer.Render(m.renderSidebar()),
		stylePanelContainer.Render(panel),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		body,
		m.renderFooter(),
		styleBeta.Render("* Control Center is Beta *"),
	)
}

func (m model) renderSidebar() string {
	var b strings.Builder
	for i, sec := range m.sections {
		label := fmt.Sprintf("%s %s", sec.icon, sec.name)
		if i == m.section {
			b.WriteString(styleSidebarSelected.Render(label))
		} else {
			b.WriteString(styleSidebarInactive.Render(label))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m model) renderPanel() string {
	sec := m.sections[m.section]
	title := fmt.Sprintf("%s %s", sec.icon, sec.name)

	var b strings.Builder
	b.WriteString(stylePanelTitle.Render(title))
	b.WriteString("\n\n")

	switch m.section {
	case 0:
		b.WriteString(m.renderAudio())
	case 1:
		b.WriteString(m.renderDisplay())
	case 2:
		b.WriteString(m.renderNetwork())
	case 3:
		b.WriteString(m.renderPower())
	case 4:
		b.WriteString(m.renderPrivacy())
	case 5:
		b.WriteString(m.renderSystem())
	}

	return b.String()
}

func onoff(v bool) string {
	if v {
		return "[ON]"
	}
	return "[OFF]"
}

func connState(v bool) string {
	if v {
		return "connected"
	}
	return "disconnected"
}

func (m model) renderNetwork() string {
	var b strings.Builder

	b.WriteString(row(m, 0, fmt.Sprintf("WiFi  %s", connState(m.wifi))))
	b.WriteString("\n")
	b.WriteString(row(m, 1, fmt.Sprintf("Ethernet  %s", connState(m.eth))))
	b.WriteString("\n")
	b.WriteString(row(m, 2, fmt.Sprintf("WireGuard  %s", onoff(m.wg))))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(styleFooter.Render("enter: wifi info / toggle wireguard"))
	return b.String()
}

func (m model) renderPrivacy() string {
	var b strings.Builder

	b.WriteString(row(m, 0, fmt.Sprintf("Night Light  %s", onoff(m.night))))
	b.WriteString("\n")
	b.WriteString(row(m, 1, fmt.Sprintf("Stealth Mode  %s", onoff(m.stealth))))
	b.WriteString("\n")

	if m.proton {
		b.WriteString(row(m, 2, "Proton Drive  sync now"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleFooter.Render("enter toggle / sync"))
	return b.String()
}

func (m model) renderSystem() string {
	var b strings.Builder

	recState := "idle"
	if m.recording {
		recState = "RECORDING"
	}

	version := m.version
	if version == "" {
		version = "unknown"
	}

	b.WriteString(row(m, 0, fmt.Sprintf("OpenRiot  %s  (check updates)", version)))
	b.WriteString("\n")
	b.WriteString(row(m, 1, fmt.Sprintf("Screen Recording  %s", recState)))
	b.WriteString("\n")
	b.WriteString(row(m, 2, "Take screenshot"))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(styleFooter.Render("enter activate"))
	return b.String()
}

func row(m model, idx int, text string) string {
	if m.cursor == idx {
		return styleRowSelected.Render(text)
	}
	return styleRowInactive.Render(text)
}

func (m model) renderAudio() string {
	var b strings.Builder

	bar := renderBar(m.volume, 20, accentColor, dimColor)
	b.WriteString(row(m, 0, fmt.Sprintf("Volume  %s  %d%%", bar, m.volume)))
	b.WriteString("\n")

	muteLabel := "Muted  [ON]"
	if !m.muted {
		muteLabel = "Muted  [OFF]"
	}
	b.WriteString(row(m, 1, muteLabel))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(styleFooter.Render("+/- adjust  enter toggle"))
	return b.String()
}

func (m model) renderDisplay() string {
	var b strings.Builder

	bar := renderBar(m.brightness, 20, accentColor, dimColor)
	b.WriteString(row(m, 0, fmt.Sprintf("Brightness  %s  %d%%", bar, m.brightness)))
	b.WriteString("\n")

	var hdmiStatus string
	if !m.hdmiHasExt {
		hdmiStatus = "HDMI: none"
	} else if m.hdmiActive {
		hdmiStatus = "HDMI: connected"
	} else {
		hdmiStatus = "HDMI: disconnected"
	}
	b.WriteString(row(m, 1, hdmiStatus))
	b.WriteString("\n")

	b.WriteString(row(m, 2, "Change resolution..."))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(styleFooter.Render("+/- adjust  enter toggle/resolution"))
	return b.String()
}

func (m model) renderPower() string {
	var b strings.Builder

	if m.batPercent == 255 {
		b.WriteString(row(m, 0, "No battery detected."))
		return b.String()
	}

	barColor := accentColor
	batIcon := "󰁹"
	switch {
	case m.batPercent >= 70:
		batIcon = "󰁹"
	case m.batPercent >= 50:
		batIcon = "󰁽"
	case m.batPercent >= 25:
		batIcon = "󰁻"
	case m.batPercent >= 10:
		batIcon = "󰁺"
		barColor = theme.Palette.Semantic.Warning
	default:
		batIcon = "󱃍"
		barColor = theme.Palette.Semantic.Error
	}

	bar := renderBar(m.batPercent, 20, barColor, dimColor)
	b.WriteString(row(m, 0, fmt.Sprintf("%s  %s  %d%%", batIcon, bar, m.batPercent)))
	b.WriteString("\n")

	status := "Estimating..."
	if m.batAC == 1 {
		status = "Plugged in"
	} else if m.batMinutes > 0 {
		hours := m.batMinutes / 60
		mins := m.batMinutes % 60
		status = fmt.Sprintf("%d:%02d remaining", hours, mins)
	}
	b.WriteString(row(m, 1, status))
	b.WriteString("\n\n")

	actions := []string{"Suspend", "Reboot", "Shutdown", "Lock"}
	for i, a := range actions {
		b.WriteString(row(m, 2+i, a))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleFooter.Render("enter activate"))
	return b.String()
}

func renderBar(pct, width int, fill, bg string) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := (pct * width) / 100
	var b strings.Builder
	fg := lipgloss.NewStyle().Foreground(lipgloss.Color(fill))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(bg))
	for i := 0; i < width; i++ {
		if i < filled {
			b.WriteString(fg.Render("█"))
		} else {
			b.WriteString(dim.Render("░"))
		}
	}
	return b.String()
}

func (m model) renderFooter() string {
	return styleFooter.Render("↑↓ items   ←→ sections   enter select   q quit")
}
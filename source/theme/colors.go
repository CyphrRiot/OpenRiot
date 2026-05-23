package theme

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/lipgloss"
)

// ColorPalette holds the canonical palette from config/colors.toml.
type ColorPalette struct {
	Base     BaseColors
	Accent   AccentColors
	Semantic SemanticColors
	Compat   CompatColors
}

// BaseColors are the background/foreground foundation.
type BaseColors struct {
	BG  string
	BG2 string
	FG  string
	FG2 string
	FG3 string
	Dim string
}

// AccentColors are the current color wheel position.
type AccentColors struct {
	Name    string
	FG      string
	FGLight string
	BG      string
}

// SemanticColors are status-driven colors.
type SemanticColors struct {
	Error   string
	Warning  string
	Success  string
	Info     string
	Cyan     string
}

// CompatColors are legacy aliases.
type CompatColors struct {
	Green   string
	Violet  string
	Blue    string
	DimGray string
	White   string
}

// ColorStyles wraps lipgloss styles for the color system.
type ColorStyles struct {
	Base     BaseStyles
	Accent   AccentStyles
	Semantic SemanticStyles
}

// BaseStyles are lipgloss styles for base colors.
type BaseStyles struct {
	BG  lipgloss.Style
	BG2 lipgloss.Style
	FG  lipgloss.Style
	FG2 lipgloss.Style
	FG3 lipgloss.Style
	Dim lipgloss.Style
}

// AccentStyles are lipgloss styles for accent colors.
type AccentStyles struct {
	Title     lipgloss.Style
	FG        lipgloss.Style
	FGLight   lipgloss.Style
	BG        lipgloss.Style
	Header    lipgloss.Style
	Selected  lipgloss.Style
	InfoLabel lipgloss.Style
	InfoValue lipgloss.Style
}

// SemanticStyles are lipgloss styles for semantic colors.
type SemanticStyles struct {
	Error   lipgloss.Style
	Warning lipgloss.Style
	Success lipgloss.Style
	Info    lipgloss.Style
	Cyan    lipgloss.Style
}

var (
	// Palette is the parsed color palette.
	Palette ColorPalette

	// Lipgloss are the derived lipgloss styles.
	Lipgloss ColorStyles

	// secColors maps security level index to color hex.
	secColors [5]string
)

func init() {
	loadColors()
	buildStyles()
}

func loadColors() {
	cfgPath := getColorsPath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		setDefaults()
		return
	}

	var raw struct {
		Base     map[string]string `toml:"base"`
		Accent   map[string]string `toml:"accent"`
		Semantic map[string]string `toml:"semantic"`
		Compat   map[string]string `toml:"compat"`
	}
	if _, err := toml.Decode(string(data), &raw); err != nil {
		setDefaults()
		return
	}

	Palette.Base = BaseColors{
		BG:  str(raw.Base, "bg", "#1a1b26"),
		BG2: str(raw.Base, "bg2", "#24283b"),
		FG:  str(raw.Base, "fg", "#c0caf5"),
		FG2: str(raw.Base, "fg2", "#a3acc9"),
		FG3: str(raw.Base, "fg3", "#565f89"),
		Dim: str(raw.Base, "dim", "#3b4261"),
	}

	Palette.Accent = AccentColors{
		Name:    str(raw.Accent, "name", "bondi-green"),
		FG:      str(raw.Accent, "fg", "#9ECE6A"),
		FGLight: str(raw.Accent, "fg-light", "#8BB85A"),
		BG:      str(raw.Accent, "bg", "#2B3A1A"),
	}

	Palette.Semantic = SemanticColors{
		Error:   str(raw.Semantic, "error", "#F7768E"),
		Warning: str(raw.Semantic, "warning", "#E0AF68"),
		Success: str(raw.Semantic, "success", "#04B575"),
		Info:    str(raw.Semantic, "info", "#7AA2F7"),
		Cyan:    str(raw.Semantic, "cyan", "#0DB9D7"),
	}

	Palette.Compat = CompatColors{
		Green:  str(raw.Compat, "green", "#9ECE6A"),
		Violet: str(raw.Compat, "violet", "#7D56F4"),
		Blue:   str(raw.Compat, "blue", "#7AA2F7"),
		DimGray: str(raw.Compat, "dim-gray", "#565F89"),
		White:  str(raw.Compat, "white", "#FAFAFA"),
	}

	secColors[0] = Palette.Compat.Green
	secColors[1] = "#F4D03F"
	secColors[2] = "#FF8844"
	secColors[3] = Palette.Semantic.Error
	secColors[4] = Palette.Base.FG3
}

func str(m map[string]string, key, fallback string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return fallback
}

func buildStyles() {
	Lipgloss.Base = BaseStyles{
		BG:  lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Base.BG)),
		BG2: lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Base.BG2)),
		FG:  lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Base.FG)),
		FG2: lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Base.FG2)),
		FG3: lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Base.FG3)),
		Dim: lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Base.Dim)),
	}

	Lipgloss.Accent = AccentStyles{
		Title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(Palette.Accent.FG)),
		FG:      lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Accent.FG)),
		FGLight: lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Accent.FGLight)),
		BG:      lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Accent.BG)),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(Palette.Base.FG)).
			Background(lipgloss.Color(Palette.Accent.FG)).
			Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Palette.Accent.FG)).
			Bold(true),
		InfoLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Palette.Accent.FG)).
			Bold(true),
		InfoValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Palette.Base.FG)),
	}

	Lipgloss.Semantic = SemanticStyles{
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Semantic.Error)).Bold(true),
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Semantic.Warning)).Bold(true),
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Semantic.Success)).Bold(true),
		Info:    lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Semantic.Info)).Bold(true),
		Cyan:    lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Semantic.Cyan)),
	}
}

func setDefaults() {
	Palette.Base = BaseColors{
		BG: "#1a1b26", BG2: "#24283b",
		FG: "#c0caf5", FG2: "#a3acc9", FG3: "#565f89", Dim: "#3b4261",
	}
	Palette.Accent = AccentColors{Name: "bondi-green", FG: "#9ECE6A", FGLight: "#8BB85A", BG: "#2B3A1A"}
	Palette.Semantic = SemanticColors{
		Error: "#F7768E", Warning: "#E0AF68", Success: "#04B575", Info: "#7AA2F7", Cyan: "#0DB9D7",
	}
	Palette.Compat = CompatColors{
		Green: "#9ECE6A", Violet: "#7D56F4", Blue: "#7AA2F7", DimGray: "#565F89", White: "#FAFAFA",
	}
	secColors = [5]string{"#9ECE6A", "#F4D03F", "#FF8844", "#F7768E", "#565F89"}
	buildStyles()
}

func getColorsPath() string {
	home, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/.local/share/openriot/config/colors.toml", home)
}

// GetAccent returns the current accent color as a hex string.
func GetAccent() string { return Palette.Accent.FG }

// GetSemantic returns a semantic color by name.
func GetSemantic(name string) string {
	switch name {
	case "error":
		return Palette.Semantic.Error
	case "warning":
		return Palette.Semantic.Warning
	case "success":
		return Palette.Semantic.Success
	case "info":
		return Palette.Semantic.Info
	case "cyan":
		return Palette.Semantic.Cyan
	default:
		return Palette.Base.FG
	}
}

// GetCompat returns a compat (legacy alias) color by name.
func GetCompat(name string) string {
	switch name {
	case "green":
		return Palette.Compat.Green
	case "violet":
		return Palette.Compat.Violet
	case "blue":
		return Palette.Compat.Blue
	case "dim-gray", "dimgray":
		return Palette.Compat.DimGray
	case "white":
		return Palette.Compat.White
	default:
		return Palette.Compat.Green
	}
}

// GetSecColor returns a WiFi security level color (0=open, 1=WPA, 2=WPA2, 3=unknown, 4=hidden).
func GetSecColor(idx int) string {
	if idx < 0 || idx >= len(secColors) {
		return Palette.Base.FG3
	}
	return secColors[idx]
}
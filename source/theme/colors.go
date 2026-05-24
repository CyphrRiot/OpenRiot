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
	Extended ExtendedColors
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
	Warning string
	Success string
	Info    string
	Cyan    string
}

// CompatColors are legacy aliases.
type CompatColors struct {
	Green   string
	Violet  string
	Blue    string
	DimGray string
	White   string
}

// ExtendedColors are polybar/rofi specific colors beyond the
// canonical palette.
type ExtendedColors struct {
	Teal       string
	Sky        string
	Electric   string
	Purple     string
	Violet     string
	Orange     string
	CyanDim    string
	LauncherFG string
	SepBG      string
	BGDark     string
	BGMid      string
	FGBright   string
	Muted      string
	SecYellow  string
	SecOrange  string
	AlphaBG    string
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
)

func init() {
	loadColors()
	buildStyles()
}

func loadColors() {
	cfgPath := getColorsPath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"colors: cannot read %s: %v\n", cfgPath, err)
		return
	}

	var raw struct {
		Base     map[string]string `toml:"base"`
		Accent   map[string]string `toml:"accent"`
		Semantic map[string]string `toml:"semantic"`
		Compat   map[string]string `toml:"compat"`
		Extended map[string]string `toml:"extended"`
	}
	if _, err := toml.Decode(string(data), &raw); err != nil {
		fmt.Fprintf(os.Stderr,
			"colors: cannot parse %s: %v\n", cfgPath, err)
		return
	}

	Palette.Base = BaseColors{
		BG:  str(raw.Base, "bg"),
		BG2: str(raw.Base, "bg2"),
		FG:  str(raw.Base, "fg"),
		FG2: str(raw.Base, "fg2"),
		FG3: str(raw.Base, "fg3"),
		Dim: str(raw.Base, "dim"),
	}

	Palette.Accent = AccentColors{
		Name:    str(raw.Accent, "name"),
		FG:      str(raw.Accent, "fg"),
		FGLight: str(raw.Accent, "fg-light"),
		BG:      str(raw.Accent, "bg"),
	}

	Palette.Semantic = SemanticColors{
		Error:   str(raw.Semantic, "error"),
		Warning: str(raw.Semantic, "warning"),
		Success: str(raw.Semantic, "success"),
		Info:    str(raw.Semantic, "info"),
		Cyan:    str(raw.Semantic, "cyan"),
	}

	Palette.Compat = CompatColors{
		Green:   str(raw.Compat, "green"),
		Violet:  str(raw.Compat, "violet"),
		Blue:    str(raw.Compat, "blue"),
		DimGray: str(raw.Compat, "dim-gray"),
		White:   str(raw.Compat, "white"),
	}

	Palette.Extended = ExtendedColors{
		Teal:       str(raw.Extended, "teal"),
		Sky:        str(raw.Extended, "sky"),
		Electric:   str(raw.Extended, "electric"),
		Purple:     str(raw.Extended, "purple"),
		Violet:     str(raw.Extended, "violet"),
		Orange:     str(raw.Extended, "orange"),
		CyanDim:    str(raw.Extended, "cyan-dim"),
		LauncherFG: str(raw.Extended, "launcher-fg"),
		SepBG:      str(raw.Extended, "sep-bg"),
		BGDark:     str(raw.Extended, "bg-dark"),
		BGMid:      str(raw.Extended, "bg-mid"),
		FGBright:   str(raw.Extended, "fg-bright"),
		Muted:      str(raw.Extended, "muted"),
		SecYellow:  str(raw.Extended, "sec-yellow"),
		SecOrange:  str(raw.Extended, "sec-orange"),
		AlphaBG:    alpha(str(raw.Extended, "bg-dark"), 0xB3),
	}

	buildStyles()
}

// str returns m[key] or empty string if missing.
func str(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return ""
}

// alpha prepends an alpha channel to a 6-digit hex color.
func alpha(hex string, a byte) string {
	if len(hex) == 7 && hex[0] == '#' {
		return fmt.Sprintf("#%02X%s", a, hex[1:])
	}
	return hex
}

func buildStyles() {
	Lipgloss.Base = BaseStyles{
		BG: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Base.BG)),
		BG2: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Base.BG2)),
		FG: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Base.FG)),
		FG2: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Base.FG2)),
		FG3: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Base.FG3)),
		Dim: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Base.Dim)),
	}

	Lipgloss.Accent = AccentStyles{
		Title: lipgloss.NewStyle().Bold(true).Foreground(
			lipgloss.Color(Palette.Accent.FG)),
		FG: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Accent.FG)),
		FGLight: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Accent.FGLight)),
		BG: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Accent.BG)),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(Palette.Base.BG)).
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
		Error: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Semantic.Error)).Bold(true),
		Warning: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Semantic.Warning)).Bold(true),
		Success: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Semantic.Success)).Bold(true),
		Info: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Semantic.Info)).Bold(true),
		Cyan: lipgloss.NewStyle().Foreground(
			lipgloss.Color(Palette.Semantic.Cyan)),
	}
}

func getColorsPath() string {
	home, _ := os.UserHomeDir()
	return fmt.Sprintf(
		"%s/.local/share/openriot/config/colors.toml", home)
}

// GetAccent returns the current accent color as a hex string.
func GetAccent() string { return Palette.Accent.FG }

// GetPurple returns the extended purple color as a hex string.
func GetPurple() string { return Palette.Extended.Purple }

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

// GetSecColor returns a WiFi security level color
// (0=open, 1=WPA, 2=WPA2, 3=unknown, 4=hidden).
func GetSecColor(idx int) string {
	switch idx {
	case 0:
		return Palette.Compat.Green
	case 1:
		return Palette.Extended.SecYellow
	case 2:
		return Palette.Extended.SecOrange
	case 3:
		return Palette.Semantic.Error
	case 4:
		return Palette.Base.FG3
	default:
		return Palette.Base.FG3
	}
}

// LoadColors reads a colors.toml file and returns the parsed palette.
// It does not modify global state.
func LoadColors(path string) (ColorPalette, error) {
	var p ColorPalette

	data, err := os.ReadFile(path)
	if err != nil {
		return p, fmt.Errorf("reading colors.toml: %w", err)
	}

	var raw struct {
		Base     map[string]string `toml:"base"`
		Accent   map[string]string `toml:"accent"`
		Semantic map[string]string `toml:"semantic"`
		Compat   map[string]string `toml:"compat"`
		Extended map[string]string `toml:"extended"`
	}
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return p, fmt.Errorf("parsing colors.toml: %w", err)
	}

	p.Base = BaseColors{
		BG:  str(raw.Base, "bg"),
		BG2: str(raw.Base, "bg2"),
		FG:  str(raw.Base, "fg"),
		FG2: str(raw.Base, "fg2"),
		FG3: str(raw.Base, "fg3"),
		Dim: str(raw.Base, "dim"),
	}

	p.Accent = AccentColors{
		Name:    str(raw.Accent, "name"),
		FG:      str(raw.Accent, "fg"),
		FGLight: str(raw.Accent, "fg-light"),
		BG:      str(raw.Accent, "bg"),
	}

	p.Semantic = SemanticColors{
		Error:   str(raw.Semantic, "error"),
		Warning: str(raw.Semantic, "warning"),
		Success: str(raw.Semantic, "success"),
		Info:    str(raw.Semantic, "info"),
		Cyan:    str(raw.Semantic, "cyan"),
	}

	p.Compat = CompatColors{
		Green:   str(raw.Compat, "green"),
		Violet:  str(raw.Compat, "violet"),
		Blue:    str(raw.Compat, "blue"),
		DimGray: str(raw.Compat, "dim-gray"),
		White:   str(raw.Compat, "white"),
	}

	p.Extended = ExtendedColors{
		Teal:       str(raw.Extended, "teal"),
		Sky:        str(raw.Extended, "sky"),
		Electric:   str(raw.Extended, "electric"),
		Purple:     str(raw.Extended, "purple"),
		Violet:     str(raw.Extended, "violet"),
		Orange:     str(raw.Extended, "orange"),
		CyanDim:    str(raw.Extended, "cyan-dim"),
		LauncherFG: str(raw.Extended, "launcher-fg"),
		SepBG:      str(raw.Extended, "sep-bg"),
		BGDark:     str(raw.Extended, "bg-dark"),
		BGMid:      str(raw.Extended, "bg-mid"),
		FGBright:   str(raw.Extended, "fg-bright"),
		Muted:      str(raw.Extended, "muted"),
		SecYellow:  str(raw.Extended, "sec-yellow"),
		SecOrange:  str(raw.Extended, "sec-orange"),
		AlphaBG:    alpha(str(raw.Extended, "bg-dark"), 0xB3),
	}

	return p, nil
}

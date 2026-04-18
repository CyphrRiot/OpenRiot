package installer

// ANSI color codes for output
const (
	Reset  = "\033[0m"
	Red    = "\033[0;31m"
	Green  = "\033[0;32m"
	Yellow = "\033[1;33m"
	Cyan   = "\033[1;36m"
	White  = "\033[1;37m"
)

// Colorize returns a string with ANSI color codes
func Colorize(color string, text string) string {
	return color + text + Reset
}

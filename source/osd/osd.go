// Package osd renders on-screen displays for volume, microphone, and
// brightness changes using osd_cat from the xosd package on X11.
package osd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"openriot/theme"
)

// ErrUnavailable is returned when no on-screen display can be shown.
var ErrUnavailable = fmt.Errorf("osd unavailable")

// Available reports whether an on-screen display can be rendered right
// now: osd_cat must be installed and an X11 display must be present.
func Available() bool {
	if os.Getenv("DISPLAY") == "" {
		return false
	}
	_, err := exec.LookPath("osd_cat")
	return err == nil
}

// KindForIcon maps a notification icon name to an OSD kind.
func KindForIcon(icon string) string {
	switch icon {
	case "speaker":
		return "volume"
	case "speaker-muted":
		return "volume-muted"
	case "mic":
		return "mic"
	case "mic-muted":
		return "mic-muted"
	case "nightlight-on", "brightness":
		return "brightness"
	default:
		return ""
	}
}

// ParsePercent extracts a percentage from text like "Speaker: 42%".
func ParsePercent(s string) (int, bool) {
	end := strings.IndexByte(s, '%')
	if end < 1 {
		return 0, false
	}
	start := end
	for start > 0 && s[start-1] >= '0' && s[start-1] <= '9' {
		start--
	}
	if start == end {
		return 0, false
	}
	pct, err := strconv.Atoi(s[start:end])
	if err != nil {
		return 0, false
	}
	return pct, true
}

// label returns the text shown for a given OSD kind.
func label(kind string) string {
	switch kind {
	case "volume":
		return "Volume"
	case "volume-muted":
		return "Volume Muted"
	case "mic":
		return "Microphone"
	case "mic-muted":
		return "Mic Muted"
	case "brightness":
		return "Brightness"
	default:
		return kind
	}
}

// Show displays an OSD for the given kind and percentage. It returns
// ErrUnavailable when no on-screen display can be rendered.
func Show(kind string, pct int) error {
	if !Available() {
		return ErrUnavailable
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	exec.Command("pkill", "-x", "osd_cat").Run()

	cmd := exec.Command("osd_cat",
		"-b", "percentage",
		"-p", "bottom",
		"-A", "center",
		"-o", "80",
		"-c", theme.GetAccent(),
		"-s", "2",
		"-O", "1",
		"-d", "1",
		"-P", strconv.Itoa(pct),
		"-T", label(kind),
	)
	return cmd.Start()
}

// Run executes the --osd command: openriot --osd <kind> <percent>.
func Run(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: openriot --osd <volume|mic|brightness> <0-100>")
		return 1
	}
	pct, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "osd: percent must be a number, got %q\n", args[1])
		return 1
	}
	if err := Show(args[0], pct); err != nil {
		fmt.Fprintf(os.Stderr, "osd: %v\n", err)
		return 1
	}
	return 0
}

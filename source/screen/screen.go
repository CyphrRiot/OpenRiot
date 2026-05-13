package screen

import (
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var resDimensionRE = regexp.MustCompile(`(\d+)x(\d+)`)

// GetWidth returns the screen width in pixels using xrandr.
// Defaults to 1920 if DISPLAY is not set or xrandr fails.
func GetWidth() int {
	if os.Getenv("DISPLAY") == "" {
		return 1920
	}

	out, err := exec.Command("xrandr").Output()
	if err != nil {
		return 1920
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "connected") {
			matches := resDimensionRE.FindStringSubmatch(line)
			if len(matches) > 1 {
				w, _ := strconv.Atoi(matches[1])
				if w > 0 {
					return w
				}
			}
		}
	}
	return 1920
}

// GetResolution returns the screen width and height in pixels using xrandr.
// Defaults to 1920, 1080 if DISPLAY is not set or xrandr fails.
func GetResolution() (int, int) {
	if os.Getenv("DISPLAY") == "" {
		return 1920, 1080
	}

	out, err := exec.Command("xrandr").Output()
	if err != nil {
		return 1920, 1080
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "connected") {
			matches := resDimensionRE.FindStringSubmatch(line)
			if len(matches) > 2 {
				w, _ := strconv.Atoi(matches[1])
				h, _ := strconv.Atoi(matches[2])
				if w > 0 && h > 0 {
					return w, h
				}
			}
		}
	}
	return 1920, 1080
}

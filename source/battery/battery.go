package battery

import (
	"fmt"
	"os/exec"
	"strconv"
)

func Get() string {
	percent, ac := getBatteryStatus()
	if percent == 255 || percent == 0 {
		return ""
	}

	icon := getBatteryIcon(percent, ac)
	return fmt.Sprintf("%s %d%%", icon, percent)
}

func getBatteryStatus() (int, int) {
	percent := 0
	ac := 0

	cmd := exec.Command("apm", "-l")
	if output, err := cmd.Output(); err == nil {
		percent, _ = strconv.Atoi(fmt.Sprintf("%s", output))
	}

	cmd = exec.Command("apm", "-a")
	if output, err := cmd.Output(); err == nil {
		ac, _ = strconv.Atoi(fmt.Sprintf("%s", output))
	}

	return percent, ac
}

func getBatteryIcon(percent, ac int) string {
	if ac == 1 {
		return "󰂄"
	}

	switch {
	case percent >= 90:
		return "󰂂"
	case percent >= 80:
		return "󰂁"
	case percent >= 70:
		return "󰂀"
	case percent >= 60:
		return "󰁿"
	case percent >= 50:
		return "󰁾"
	case percent >= 40:
		return "󰁽"
	case percent >= 30:
		return "󰁼"
	case percent >= 20:
		return "󰁻"
	case percent >= 10:
		return "󰁺"
	default:
		return "󰁺"
	}
}

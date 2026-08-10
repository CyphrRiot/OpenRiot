package battery

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"openriot/notify"
	"openriot/polybar"
)

var (
	alerted20m bool
	alerted15m bool
	alerted10m bool
	alerted5m  bool
)

// GetNotifyDetails returns formatted battery info for notifications.
func GetNotifyDetails() string {
	percent, ac, minutes := getFullStatus()

	if percent == 255 {
		return "No Battery Installed"
	}

	if ac == 1 {
		if minutes > 0 {
			return fmt.Sprintf("Plugged In at %d%%\nEstimated Time: %s", percent, formatTime(minutes))
		}
		return fmt.Sprintf("Plugged In at %d%%\nOn AC Power", percent)
	}

	if minutes > 0 {
		return fmt.Sprintf("Charged to %d%%\nRemaining Time: %s", percent, formatTime(minutes))
	}
	return fmt.Sprintf("Charged to %d%%", percent)
}

func getFullStatus() (percent, ac, minutes int) {
	percent = 0
	ac = 0
	minutes = 0

	cmd := exec.Command("apm", "-b")
	if output, err := cmd.Output(); err == nil {
		bStatus, _ := strconv.Atoi(strings.TrimSpace(string(output)))
		if bStatus == 4 {
			return 255, 0, 0
		}
	}

	cmd = exec.Command("apm", "-l")
	if output, err := cmd.Output(); err == nil {
		percent, _ = strconv.Atoi(strings.TrimSpace(string(output)))
	}

	cmd = exec.Command("apm", "-a")
	if output, err := cmd.Output(); err == nil {
		ac, _ = strconv.Atoi(strings.TrimSpace(string(output)))
	}

	cmd = exec.Command("apm", "-m")
	if output, err := cmd.Output(); err == nil {
		minutes, _ = strconv.Atoi(strings.TrimSpace(string(output)))
	}

	return percent, ac, minutes
}

func formatTime(minutes int) string {
	if minutes <= 0 {
		return "N/A"
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

func Get() string {
	percent, ac, minutes := getFullStatus()
	if percent == 255 {
		return "" // No battery = no icon in polybar
	}
	if percent == 0 {
		return "" // No battery info
	}

	// Reset alerts when charging or battery has more than 20 minutes
	if ac == 1 || minutes > 20 {
		alerted20m = false
		alerted15m = false
		alerted10m = false
		alerted5m = false
	}

	// Send low-battery notifications only when discharging
	// and percent indicates we're actually low (minutes can be stale
	// during AC-to-battery transitions)
	if ac == 0 && percent < 25 && minutes > 0 {
		if minutes <= 5 && !alerted5m {
			notify.SendNotify("battery", "Battery Critical",
				"Less than 5 minutes remaining!", "critical", 0, 0)
			alerted5m = true
			alerted10m = true
			alerted15m = true
			alerted20m = true
		} else if minutes <= 10 && !alerted10m {
			notify.SendNotify("battery", "Battery Warning",
				"Less than 10 minutes remaining!", "critical", 0, 0)
			alerted10m = true
			alerted15m = true
			alerted20m = true
		} else if minutes <= 15 && !alerted15m {
			notify.SendNotify("battery", "Battery Warning",
				"Less than 15 minutes remaining!", "normal", 5000, 0)
			alerted15m = true
			alerted20m = true
		} else if minutes <= 20 && !alerted20m {
			notify.SendNotify("battery", "Battery Low",
				"Less than 20 minutes remaining!", "normal", 5000, 0)
			alerted20m = true
		}
	}

	icon := polybar.Icon(getBatteryIcon(percent, ac))

	if ac == 1 {
		return fmt.Sprintf("%%{F#0DB9D7}%s%%{F-}", icon)
	}
	switch {
	case percent < 20:
		return fmt.Sprintf("%%{F#F7768E}%s%%{F-}", icon)
	case percent < 25:
		return fmt.Sprintf("%%{F#FF9E64}%s%%{F-}", icon)
	default:
		return fmt.Sprintf("%%{F#9ECE6A}%s%%{F-}", icon)
	}
}

func getBatteryIcon(percent, ac int) string {
	// Arrays indexed by (percent-1)/10 = 0-8
	batteryIcons := []string{"󰁺", "󰁻", "󰁼", "󰁽", "󰁾", "󰁿", "󰂀", "󰂁", "󰂂"}
	chargingIcons := []string{"󰢜", "󰢜", "󰂇", "󰂈", "󰢝", "󰂉", "󰢞", "󰂊", "󰂋"}

	idx := max(0, min((percent-1)/10, 8))

	if ac == 1 {
		if percent >= 100 {
			return "󰁹"
		}
		return chargingIcons[idx]
	}
	return batteryIcons[idx]
}

// TestNotify simulates a battery with given minutes remaining and triggers notification logic.
func TestNotify(minutes int) {
	// Reset state so alerts always fire
	alerted20m = false
	alerted15m = false
	alerted10m = false
	alerted5m = false

	_, _, _ = getFullStatus()

	// Estimate percent from minutes for icon display
	percent := 50
	if minutes <= 5 {
		percent = 5
	} else if minutes <= 10 {
		percent = 10
	} else if minutes <= 15 {
		percent = 15
	} else if minutes <= 20 {
		percent = 20
	}

	icon := polybar.Icon(getBatteryIcon(percent, 0))

	var color string
	switch {
	case percent < 20:
		color = "#F7768E"
	case percent < 25:
		color = "#FF9E64"
	default:
		color = "#9ECE6A"
	}
	fmt.Printf("%%{F%s}%s%%{F-}\n", color, icon)

	if minutes <= 5 {
		notify.SendNotify("battery", "Battery Critical",
			"Less than 5 minutes remaining!", "critical", 0, 0)
		alerted5m = true
		alerted10m = true
		alerted15m = true
		alerted20m = true
	} else if minutes <= 10 {
		notify.SendNotify("battery", "Battery Warning",
			"Less than 10 minutes remaining!", "critical", 0, 0)
		alerted10m = true
		alerted15m = true
		alerted20m = true
	} else if minutes <= 15 {
		notify.SendNotify("battery", "Battery Warning",
			"Less than 15 minutes remaining!", "normal", 5000, 0)
		alerted15m = true
		alerted20m = true
	} else if minutes <= 20 {
		notify.SendNotify("battery", "Battery Low",
			"Less than 20 minutes remaining!", "normal", 5000, 0)
		alerted20m = true
	}
}

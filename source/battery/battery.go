package battery

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// GetNotifyDetails returns formatted battery info for notifications.
func GetNotifyDetails() string {
	percent, ac, minutes := getFullStatus()

	if percent == 255 {
		return "No Battery Installed"
	}

	timeStr := formatTime(minutes)

	if ac == 1 {
		return fmt.Sprintf("Charging at %d%%\nEstimated Time: %s", percent, timeStr)
	}
	return fmt.Sprintf("Charged to %d%%\nRemaining Time: %s", percent, timeStr)
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
	percent, ac, _ := getFullStatus()
	if percent == 255 {
		return "󱉞" // No battery
	}
	if percent == 0 {
		return "" // No battery info
	}

	return getBatteryIcon(percent, ac)
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

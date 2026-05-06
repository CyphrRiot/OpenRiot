package detect

import (
	"fmt"
	"os/exec"
	"strings"

	"openriot/display"
)

// IsUndocked returns true if the system appears to be undocked (no external power detected).
// On laptops this means on battery power.
func IsUndocked() bool {
	// Check if on battery: acpiconf -s shows battery state
	out, err := exec.Command("sh", "-c", "sysctl hw.sensors acpibat0 2>/dev/null | grep -q present && echo yes || echo no").Output()
	if err == nil && strings.TrimSpace(string(out)) == "yes" {
		// Battery present - check if on AC
		ac, _ := exec.Command("sh", "-c", "sysctl hw.sensors acpibat0 | grep -c online || echo 0").Output()
		if strings.TrimSpace(string(ac)) != "0" {
			return false // On AC power
		}
		return true // On battery
	}
	return false // Assume docked/desktop if no battery
}

// SuspendIfUndocked checks and suspends if on battery with no external display.
func SuspendIfUndocked() {
	if !IsUndocked() {
		fmt.Println("Docked or on AC - not suspending")
		return
	}
	if display.HasExternalDisplay() {
		fmt.Println("On battery but external display connected - not suspending")
		return
	}
	fmt.Println("Undocked, suspending...")
	_ = exec.Command("zzz").Run()
}

package detect

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsUndocked returns true if the system appears to be undocked (no external power detected).
// On laptops this means on battery power.
func IsUndocked() bool {
	// Check if on battery: acpiconf -s shows battery state
	out, err := exec.Command("sh", "-c", "sysctl hw.sensors acpibat0 2>/dev/null | grep -q present && echo yes || echo no").Output()
	if err == nil && strings.TrimSpace(string(out)) == "yes" {
		// Battery present — check if on AC
		ac, _ := exec.Command("sh", "-c", "sysctl hw.sensors acpibat0 | grep -c online || echo 0").Output()
		if strings.TrimSpace(string(ac)) != "0" {
			return false // On AC power
		}
		return true // On battery
	}
	return false // Assume docked/desktop if no battery
}

// SuspendIfUndocked checks and suspends if on battery.
func SuspendIfUndocked() {
	if IsUndocked() {
		fmt.Println("Undocked, suspending...")
		_ = exec.Command("zzz").Run()
	} else {
		fmt.Println("Docked or on AC — not suspending")
	}
}

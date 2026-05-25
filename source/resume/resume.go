package resume

import (
	"time"

	"openriot/display"
	"openriot/network"
	"openriot/notify"
	"openriot/wireguard"
)

// Restore brings the system back to a working state after resume from suspend.
// It waits briefly for hardware to settle, then restores displays, WiFi,
// and WireGuard if it was active before suspend.
func Restore() error {
	time.Sleep(2 * time.Second)

	display.RestoreDisplays()

	_ = network.ReconnectWifi()

	if wireguard.IsRunning() {
		_ = wireguard.Restart()
	}

	notify.SendNotify("resume", "System", "Resumed from suspend", "normal", 3000, 0)
	return nil
}

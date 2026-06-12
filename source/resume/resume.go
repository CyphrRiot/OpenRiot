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

	// WiFi driver may still be re-initializing after resume; retry.
	var wifiErr error
	for i := 0; i < 3; i++ {
		wifiErr = network.ReconnectWifi()
		if wifiErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if wifiErr != nil {
		notify.SendNotify("resume", "WiFi", "Failed to reconnect: "+wifiErr.Error(), "critical", 5000, 0)
		return wifiErr
	}

	if wireguard.IsRunning() {
		_ = wireguard.Restart()
	}

	notify.SendNotify("resume", "System", "Resumed from suspend", "normal", 3000, 0)
	return nil
}

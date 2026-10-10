package control

import (
	"os/exec"
	"strconv"
	"strings"

	"openriot/display"
	"openriot/macspoof"
	"openriot/network"
	"openriot/nightlight"
	"openriot/notify"
	"openriot/paths"
	"openriot/polybar"
	"openriot/screenrec"
	"openriot/screenshot"
	"openriot/update"
	"openriot/wireguard"
)

func volumeState() (int, bool) {
	out, err := exec.Command("sndioctl", "-n", "output.level").Output()
	if err != nil {
		return 0, false
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, false
	}
	vol := int(f * 100)

	out, err = exec.Command("sndioctl", "-n", "output.mute").Output()
	muted := false
	if err == nil {
		muted = strings.TrimSpace(string(out)) == "1"
	}
	return vol, muted
}

func setVolume(level int) error {
	val := float64(level) / 100.0
	return exec.Command("sndioctl", "output.level="+strconv.FormatFloat(val, 'f', 2, 64)).Run()
}

func toggleMute() error {
	return exec.Command("sh", "-c", "sndioctl output.mute=!").Run()
}

type batteryInfo struct {
	percent int
	ac      int
	minutes int
}

func getBattery() batteryInfo {
	percent := 0
	ac := 0
	minutes := 0

	out, err := exec.Command("apm", "-b").Output()
	if err == nil {
		bStatus, _ := strconv.Atoi(strings.TrimSpace(string(out)))
		if bStatus == 4 {
			return batteryInfo{percent: 255}
		}
	}

	out, err = exec.Command("apm", "-l").Output()
	if err == nil {
		percent, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	out, err = exec.Command("apm", "-a").Output()
	if err == nil {
		ac, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	out, err = exec.Command("apm", "-m").Output()
	if err == nil {
		minutes, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	return batteryInfo{percent: percent, ac: ac, minutes: minutes}
}

func suspendCmd() error {
	return exec.Command("sh", "-c", "doas /usr/sbin/zzz &").Run()
}

func rebootCmd() error {
	return exec.Command("sh", "-c", "alacritty --class floating_center -e sh -c 'doas /sbin/reboot'").Start()
}

func shutdownCmd() error {
	return exec.Command("sh", "-c", "alacritty --class floating_center -e sh -c 'doas /sbin/shutdown -p now'").Start()
}

func lockCmd() error {
	return exec.Command(paths.OpenRiotDir("install", "openriot"), "--lock").Start()
}

func getBrightness() int {
	return display.GetBrightness()
}

func setBrightness(pct int) error {
	return display.SetBrightness(pct)
}

func hdmiInfo() (hasExternal bool, externalActive bool) {
	return display.HDMIStatus()
}

func toggleHDMI() error {
	return exec.Command(paths.OpenRiotDir("install", "openriot"), "--hdmi-toggle").Start()
}

func launchResolution() error {
	return exec.Command("alacritty", "--class", "openriot_resolution", "-e",
		paths.OpenRiotDir("install", "openriot"), "--resolution-tui").Start()
}

type systemState struct {
	wifi      bool
	eth       bool
	wg        bool
	night     bool
	stealth   bool
	proton    bool
	recording bool
	version   string
}

func fetchSystemState() systemState {
	return systemState{
		wifi:      network.IsConnected(),
		eth:       network.IsEthernetConnected(),
		wg:        wireguard.IsRunning(),
		night:     nightlight.IsOn(),
		stealth:   macspoof.IsStealthEnabled(),
		proton:    polybar.IsProtonDriveConfigured(),
		recording: screenrec.IsRecording(),
		version:   update.GetLocalVersion(),
	}
}

func wifiAction() {
	if network.IsConnected() {
		notify.SendNotify("wifi", "WiFi", network.GetWifiDetails(), "normal", 5000, 0)
		return
	}
	notify.SendNotify("wifi-off", "WiFi", "Not connected", "normal", 3000, 0)
}

func toggleWireguard() error {
	return wireguard.Toggle()
}

func toggleNightlight() error {
	return nightlight.Toggle()
}

func toggleStealth() error {
	return macspoof.StealthToggle()
}

func triggerProtonSync() error {
	return polybar.TriggerSync()
}

func toggleRecording() error {
	return screenrec.Toggle()
}

func runUpdate() error {
	return update.Click()
}

func takeScreenshot() error {
	return screenshot.Run(false)
}
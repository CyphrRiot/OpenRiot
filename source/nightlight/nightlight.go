package nightlight

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"openriot/notify"
	"openriot/paths"
	"openriot/polybar"
)

const (
	stateFile = ".config/openriot/nightlight.state"
	iconOn    = "󰌵"
	iconOff   = ""
)

func Get() string {
	if IsOn() {
		return polybar.Icon(iconOn)
	}
	return polybar.Icon(iconOff)
}

func IsOn() bool {
	return getState() == 1
}

func Toggle() error {
	if IsOn() {
		// Turn off
		exec.Command("sct").Run() // reset
		setState(0)
		sendNotify("Night Light: Off", "nightlight-off")
	} else {
		// Turn on
		exec.Command("sct", "4000").Run()
		setState(1)
		sendNotify("Night Light: On", "nightlight-on")
	}
	return nil
}

func getState() int {
	file := paths.Join(stateFile)
	data, err := os.ReadFile(file)
	if err != nil {
		// Migrate from old dot-file name
		oldFile := paths.Join(".config", "openriot", ".nightlight")
		data, err = os.ReadFile(oldFile)
		if err != nil {
			return 0
		}
		state, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		setState(state)
		os.Remove(oldFile)
		return state
	}
	state, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return state
}

func setState(state int) {
	file := paths.Join(stateFile)
	os.WriteFile(file, []byte(strconv.Itoa(state)), 0600)
}

func sendNotify(message, icon string) {
	notify.SendNotify(icon, "Display", message, "normal", 3000, 0)
}

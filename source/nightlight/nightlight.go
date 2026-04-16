package nightlight

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"openriot/paths"
)

const (
	stateFile    = ".config/openriot/.nightlight"
	iconOn       = "󰌵"
	iconOff      = ""
	defaultLat   = "40.71"
	defaultLon   = "-74.00"
	tempDay      = "6500"
	tempNight    = "4000"
)

var homeDir, _ = os.UserHomeDir()

func Get() string {
	state := getState()
	ensureRedshift(state)
	if state == 1 {
		return iconOn
	}
	return iconOff
}

func Toggle() error {
	currentState := getState()
	if currentState == 1 {
		// Turn off
		exec.Command("pkill", "redshift").Run()
		setState(0)
		notify("Night Light: Off", "nightlight-off.png")
	} else {
		// Turn on
		startRedshift()
		setState(1)
		notify("Night Light: On", "nightlight-on.png")
	}
	return nil
}

func getState() int {
	file := filepath.Join(homeDir, stateFile)
	data, err := os.ReadFile(file)
	if err != nil {
		return 0
	}
	state, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return state
}

func setState(state int) {
	file := filepath.Join(homeDir, stateFile)
	os.WriteFile(file, []byte(strconv.Itoa(state)), 0600)
}

func isRedshiftRunning() bool {
	cmd := exec.Command("pgrep", "-x", "redshift")
	return cmd.Run() == nil
}

func ensureRedshift(state int) {
	running := isRedshiftRunning()
	if state == 1 && !running {
		startRedshift()
	} else if state == 0 && running {
		exec.Command("pkill", "redshift").Run()
	}
}

func startRedshift() {
	exec.Command("sh", "-c", fmt.Sprintf("redshift -l %s:%s -t %s:%s &", defaultLat, defaultLon, tempNight, tempNight)).Run()
}

func notify(message, icon string) {
	exec.Command("/usr/local/bin/notify-send", "-i", paths.GetIconPath(icon), "-t", "3000", "Display", message).Start()
}

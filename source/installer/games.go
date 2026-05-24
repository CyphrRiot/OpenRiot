package installer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"openriot/logger"
	"openriot/paths"
)

// GamesPreference checks whether the user wants games installed.
// It reads ~/.config/openriot/games.cfg. If missing, it prompts
// the user interactively and saves the choice.
// Returns true if games should be installed.
func GamesPreference() bool {
	cfgDir := paths.Join(".config", "openriot")
	cfgPath := filepath.Join(cfgDir, "games.cfg")

	// If config exists, respect it
	if data, err := os.ReadFile(cfgPath); err == nil {
		choice := strings.TrimSpace(strings.ToLower(string(data)))
		answer := choice == "yes" || choice == "y"
		if answer {
			logger.Info("Games will be installed...")
		}
		return answer
	}

	// Prompt user
	logger.Warn("Games are huge. Fun, but massive.")
	logger.Ask("Do you want to install Games (~3GB)? [Y/n] ")
	stdin := os.Stdin
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		stdin = tty
		defer tty.Close()
	}
	reader := bufio.NewReader(stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	answer := input == "yes" || input == "y" || input == ""

	// Save choice
	os.MkdirAll(cfgDir, 0755)
	if answer {
		os.WriteFile(cfgPath, []byte("yes\n"), 0644)
	} else {
		os.WriteFile(cfgPath, []byte("no\n"), 0644)
	}

	if answer {
		logger.Info("Games will be installed...")
	}
	logger.Info(fmt.Sprintf("Games preference saved to %s", cfgPath))
	logger.Info("Delete this file to be asked again.")

	return answer
}

// StripGamesFromRofi removes the Games entry from rofi apps.txt and deletes
// games.txt when the user opted out of games installation.
func StripGamesFromRofi() {
	appsPath := paths.Join(".config", "rofi", "apps.txt")
	gamesPath := paths.Join(".config", "rofi", "games.txt")

	// Remove games submenu file
	os.Remove(gamesPath)

	// Strip Games line from apps.txt
	data, err := os.ReadFile(appsPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	var out []string
	for _, line := range lines {
		if strings.Contains(line, "Games|@submenu:games") {
			continue
		}
		out = append(out, line)
	}

	os.WriteFile(appsPath, []byte(strings.Join(out, "\n")), 0644)
}

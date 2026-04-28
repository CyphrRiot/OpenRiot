package installer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"openriot/logger"
)

// GamesPreference checks whether the user wants games installed.
// It reads ~/.config/openriot/games.cfg. If missing, it prompts
// the user interactively and saves the choice.
// Returns true if games should be installed.
func GamesPreference() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	cfgDir := filepath.Join(home, ".config", "openriot")
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
	logger.Ask("Do you want to install Games (~1.75G)? [Y/n] ")
	reader := bufio.NewReader(os.Stdin)
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
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	appsPath := filepath.Join(home, ".config", "rofi", "apps.txt")
	gamesPath := filepath.Join(home, ".config", "rofi", "games.txt")

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

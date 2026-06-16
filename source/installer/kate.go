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

// KatePreference checks whether the user wants Kate IDE installed.
// It reads ~/.config/openriot/kate.cfg. If missing, it prompts
// the user interactively and saves the choice.
// Returns true if Kate should be installed.
func KatePreference() bool {
	cfgDir := paths.Join(".config", "openriot")
	cfgPath := filepath.Join(cfgDir, "kate.cfg")

	// If config exists, respect it
	if data, err := os.ReadFile(cfgPath); err == nil {
		choice := strings.TrimSpace(strings.ToLower(string(data)))
		answer := choice == "yes" || choice == "y"
		if answer {
			logger.Info("Kate IDE will be installed...")
		}
		return answer
	}

	// Prompt user
	logger.Warn("Kate is a heavy KDE-based code editor (~300MB deps).")
	logger.Ask("Would you like to install Kate (Code IDE)? [Y/n] ")
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
		logger.Info("Kate IDE will be installed...")
	} else {
		os.WriteFile(cfgPath, []byte("no\n"), 0644)
	}

	logger.Info(fmt.Sprintf("Kate preference saved to %s", cfgPath))
	logger.Info("Delete this file to be asked again.")

	return answer
}

// StripKateFromRofi removes the Kate entry from rofi apps.txt.
func StripKateFromRofi() {
	appsPath := paths.Join(".config", "rofi", "apps.txt")

	data, err := os.ReadFile(appsPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	var out []string
	for _, line := range lines {
		if strings.Contains(line, "Kate IDE|kate") {
			continue
		}
		out = append(out, line)
	}

	os.WriteFile(appsPath, []byte(strings.Join(out, "\n")), 0644)
}

// SetupKateConfig writes Kate's katerc with Ayu Dark editor theme
// and BreezeDark UI color scheme.
func SetupKateConfig() error {
	configPath := paths.Join(".config", "katerc")

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	// Read existing katerc into sections
	type sectionMap map[string]string
	sections := make(map[string]sectionMap)
	var sectionOrder []string

	if data, err := os.ReadFile(configPath); err == nil {
		var current string
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
				current = trimmed[1 : len(trimmed)-1]
				if sections[current] == nil {
					sections[current] = make(sectionMap)
					sectionOrder = append(sectionOrder, current)
				}
				continue
			}
			if parts := strings.SplitN(trimmed, "=", 2); len(parts) == 2 && current != "" {
				sections[current][strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// If no existing file, set up default sections in order
	if len(sections) == 0 {
		sectionOrder = []string{
			"General",
			"KTextEditor Renderer",
			"UiSettings",
			"filetree",
			"MainWindow",
		}
		for _, s := range sectionOrder {
			sections[s] = make(sectionMap)
		}
	}

	// Apply our overrides
	if sections["KTextEditor Renderer"] == nil {
		sections["KTextEditor Renderer"] = make(sectionMap)
	}
	sections["KTextEditor Renderer"]["Auto Color Theme Selection"] = "false"
	sections["KTextEditor Renderer"]["Color Theme"] = "ayu Dark"
	sections["KTextEditor Renderer"]["Text Font"] = "FiraCode Nerd Font,10,-1,5,400,0,0,0,0,0,0,0,0,0,0,1,FiraCode Nerd Font"

	if sections["UiSettings"] == nil {
		sections["UiSettings"] = make(sectionMap)
	}
	sections["UiSettings"]["ColorScheme"] = "BreezeDark"

	// Write katerc preserving section order
	var buf strings.Builder
	for _, sec := range sectionOrder {
		kv, ok := sections[sec]
		if !ok || len(kv) == 0 {
			continue
		}
		buf.WriteString(fmt.Sprintf("[%s]\n", sec))
		for k, v := range kv {
			buf.WriteString(fmt.Sprintf("%s=%s\n", k, v))
		}
		buf.WriteString("\n")
	}

	// Append any new sections not in original order
	for sec, kv := range sections {
		found := false
		for _, s := range sectionOrder {
			if s == sec {
				found = true
				break
			}
		}
		if !found && len(kv) > 0 {
			buf.WriteString(fmt.Sprintf("[%s]\n", sec))
			for k, v := range kv {
				buf.WriteString(fmt.Sprintf("%s=%s\n", k, v))
			}
			buf.WriteString("\n")
		}
	}

	if err := os.WriteFile(configPath, []byte(buf.String()), 0600); err != nil {
		return fmt.Errorf("cannot write katerc: %w", err)
	}

	fmt.Println("[DONE] Kate IDE configured with Ayu Dark theme.")
	return nil
}
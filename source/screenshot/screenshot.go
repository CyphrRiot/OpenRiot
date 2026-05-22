package screenshot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"openriot/paths"
)

// Run captures a screenshot using maim and copies to clipboard.
// If select is true, allows area selection. Otherwise captures full screen.
func Run(selectArea bool) error {
	dir := paths.Join("Screenshots")

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create Screenshots dir: %w", err)
	}

	filename := fmt.Sprintf("screenshot_%s.png", time.Now().Format("20060102_150405"))
	path := filepath.Join(dir, filename)

	var cmd *exec.Cmd
	if selectArea {
		cmd = exec.Command("maim", "-s", path)
	} else {
		cmd = exec.Command("maim", path)
	}
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("maim failed: %w", err)
	}

	// Copy to clipboard
	clipCmd := exec.Command("xclip", "-selection", "clipboard", "-t", "image/png", "-i", path)
	if err := clipCmd.Run(); err != nil {
		return fmt.Errorf("xclip failed: %w", err)
	}

	return nil
}

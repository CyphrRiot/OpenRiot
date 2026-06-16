package installer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"openriot/config"
	"openriot/logger"
	"openriot/paths"
)

// InstallPackages installs packages using pkg_add with doas.
// On snapshot systems, the installer detects -current and passes the appropriate flag.
// Falls back to base package name (no version) if exact version fails.
// Returns (failedCount, error)
func InstallPackages(cfg *config.Config, packages []string) (int, error) {
	if len(packages) == 0 {
		return 0, nil
	}

	// Filter out already-installed packages using a single pkg_info -a call
	installed := GetInstalledPackages()
	var toInstall []string
	for _, pkg := range packages {
		base := config.GetBaseName(pkg)
		if installedVer, ok := installed[base]; ok {
			if installedVer != pkg {
				logger.Warn(fmt.Sprintf("Newer version of %s installed", pkg))
			}
			continue
		}
		toInstall = append(toInstall, pkg)
	}

	if len(toInstall) == 0 {
		logger.Done("All packages already installed")
		return 0, nil
	}

	logger.Info(fmt.Sprintf("Installing %d packages (%d new, %d already installed)...", len(packages), len(toInstall), len(packages)-len(toInstall)))

	failed := 0
	for _, pkg := range toInstall {
		// On snapshot systems the exact version from packages.yaml won't exist
		// on the snapshot mirror. Install by base name directly.
		installName := pkg
		if cfg.IsSnapshot() {
			installName = config.GetBaseName(pkg)
		}
		logger.Info(fmt.Sprintf("Installing %s...", installName))

		// Build pkg_add command: use -D snapshot only for snapshot systems
		installCmd := []string{"doas", "pkg_add"}
		if cfg.IsSnapshot() {
			installCmd = append(installCmd, "-D", "snapshot")
		}
		installCmd = append(installCmd, installName)

		// Special handling for Zero A.D. — massive package, requires
		// explicit user confirmation and extended timeout.
		timeout := 30 * time.Minute
		if config.GetBaseName(pkg) == "0ad" {
			if !confirm0ad() {
				logger.Info("Skipping Zero A.D.")
				strip0adFromGames()
				continue
			}
			timeout = 60 * time.Minute
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		cmd := exec.CommandContext(ctx, installCmd[0], installCmd[1:]...)

		// Stream output for massive packages so the user sees real progress
		is0ad := config.GetBaseName(pkg) == "0ad"
		if is0ad {
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		start := time.Now()
		done := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if is0ad {
				select {
				case <-time.After(30 * time.Second):
				case <-done:
					return
				}
				spinner := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				i := 0
				for {
					select {
					case <-ticker.C:
						elapsed := time.Since(start)
						mins := int(elapsed.Minutes())
						secs := int(elapsed.Seconds()) % 60
						fmt.Printf("\r\033[K%s[INFO]%s %s %s (%02dm %02ds elapsed)", logger.Cyan, logger.Reset, string(spinner[i%len(spinner)]), installName, mins, secs)
						i++
					case <-done:
						fmt.Println()
						return
					}
				}
			}
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					elapsed := int(time.Since(start).Minutes())
					logger.Info(fmt.Sprintf("Still running: %s (%dm elapsed)", installName, elapsed))
				case <-done:
					return
				}
			}
		}()

		var output []byte
		var err error
		if is0ad {
			err = cmd.Run()
		} else {
			output, err = cmd.CombinedOutput()
		}
		close(done)
		wg.Wait()
		cancel()

		if err != nil {
			outputStr := truncateOutput(output)
			if ctx.Err() == context.DeadlineExceeded {
				logger.Warn(fmt.Sprintf("Timed out after %dm: %s", int(timeout.Minutes()), installName))
			}
			// Retry with base name if exact version failed
			base := config.GetBaseName(pkg)
			if base != pkg {
				if errOut := tryInstall(installCmd, base, timeout); errOut == "" {
					continue
				} else {
					outputStr = errOut
				}
			}
			logger.Warn(fmt.Sprintf("Failed to install %s:\n    %s", installName, outputStr))
			failed++
		} else {
			logger.Done(fmt.Sprintf("%s installed", installName))
		}
	}

	if failed > 0 {
		logger.Warn(fmt.Sprintf("%d packages failed to install.", failed))
		logger.Warn("You can install remaining ones manually: doas pkg_add <package>")
	} else {
		logger.Done("All packages installed")
	}

	return failed, nil
}

func truncateOutput(output []byte) string {
	s := string(output)
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}

// tryInstall attempts to install a package with two retries:
//   1. Base name with the original install flags
//   2. Base name with -D snap (for -current systems where the
//      exact version doesn't exist on the stable mirror)
//
// Returns the error output if both attempts fail, or empty string on success.
func tryInstall(installCmd []string, base string, timeout time.Duration) string {
	// Attempt 1: base name with original flags
	cmd1 := append([]string{}, installCmd...)
	cmd1[len(cmd1)-1] = base
	if out, err := runInstallAttempt(cmd1, base, timeout); err == nil {
		return ""
	} else {
		_ = out
	}

	// Attempt 2: base name with -D snap for -current
	logger.Info(fmt.Sprintf("Retrying %s with snapshot mode...", base))
	cmd2 := []string{installCmd[0], installCmd[1], "-D", "snap", base}
	out, err := runInstallAttempt(cmd2, base, timeout)
	if err == nil {
		return ""
	}
	return truncateOutput(out)
}

func runInstallAttempt(cmd []string, name string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := int(time.Since(start).Minutes())
				logger.Info(fmt.Sprintf("Still running: %s (%dm elapsed)", name, elapsed))
			case <-done:
				return
			}
		}
	}()

	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	out, err := c.CombinedOutput()
	close(done)

	if err == nil {
		logger.Done(fmt.Sprintf("%s installed (latest version)", name))
	} else if ctx.Err() == context.DeadlineExceeded {
		logger.Warn(fmt.Sprintf("Timed out after %dm: %s", int(timeout.Minutes()), name))
	}
	return out, err
}

// isSignatureError checks whether pkg_add output indicates a base/package
// signature mismatch, which happens when OpenBSD -current snapshots drift.
func isSignatureError(output string) bool {
	s := strings.ToLower(output)
	return strings.Contains(s, "signify") ||
		strings.Contains(s, "signature") ||
		strings.Contains(s, "pubkey") ||
		strings.Contains(s, "verification") ||
		strings.Contains(s, "can't verify") ||
		strings.Contains(s, "gpg")
}

// isLibraryMismatch checks whether pkg_add output indicates a library
// version mismatch, common on -current when Qt or other libs drift.
func isLibraryMismatch(output string) bool {
	s := strings.ToLower(output)
	return strings.Contains(s, "bad major") ||
		strings.Contains(s, "incompatible") ||
		strings.Contains(s, "mismatch") ||
		strings.Contains(s, "depends on") ||
		strings.Contains(s, "missing") ||
		strings.Contains(s, "qt") ||
		strings.Contains(s, "library")
}

// confirm0ad prompts the user twice before installing the massive Zero A.D.
// package. Returns true if the user confirms both prompts.
func confirm0ad() bool {
	logger.Warn("Zero A.D. is massive (~500MB+) and takes 30–45 minutes to install.")
	logger.Ask("Install Zero A.D.? [Y/n] ")

	stdin := os.Stdin
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		stdin = tty
		defer tty.Close()
	}
	reader := bufio.NewReader(stdin)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input != "yes" && input != "y" && input != "" {
		return false
	}

	logger.Ask("Are you sure? This will take a while. [Y/n] ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "yes" || input == "y" || input == ""
}

// strip0adFromGames removes the Zero A.D. entry from the rofi games menu
// when the user opted not to install it.
func strip0adFromGames() {
	gamesPath := paths.Join(".config", "rofi", "games.txt")
	data, err := os.ReadFile(gamesPath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	var out []string
	for _, line := range lines {
		if strings.Contains(line, "0ad") {
			continue
		}
		out = append(out, line)
	}
	os.WriteFile(gamesPath, []byte(strings.Join(out, "\n")), 0644)
}

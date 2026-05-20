package installer

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ReleasePathResult classifies a system's upgrade path.
type ReleasePathResult struct {
	Status      string // "stable", "pre-release", "post-release"
	BuildDate   time.Time
	ReleaseDate time.Time
	KernelLine  string
}

// ErrUpgradeRequired indicates pre-release snapshot that can migrate to stable.
var ErrUpgradeRequired = fmt.Errorf("pre-release snapshot: sysupgrade -R available")

// ErrDowngradeRisk indicates post-release snapshot where -R is unsafe.
var ErrDowngradeRisk = fmt.Errorf("post-release snapshot: sysupgrade -R is a downgrade")

// ReleaseDate is the build date of the current OpenBSD release sets.
// Update this when a new OpenBSD version is released.
var ReleaseDate = time.Date(2026, time.May, 6, 0, 0, 0, 0, time.UTC)
// CheckReleasePath parses the kernel version and determines the migration
// status relative to a known release date.
func CheckReleasePath(releaseDate time.Time) (ReleasePathResult, error) {
	cmd := exec.Command("sysctl", "-n", "kern.version")
	output, err := cmd.Output()
	if err != nil {
		return ReleasePathResult{}, fmt.Errorf("cannot read kern.version: %w", err)
	}

	line := strings.TrimSpace(string(output))
	// sysctl kern.version is multi-line; take only the first line
	if nl := strings.Index(line, "\n"); nl >= 0 {
		line = line[:nl]
	}
	return checkReleasePathFromKernelLine(line, releaseDate)
}

func checkReleasePathFromKernelLine(line string, releaseDate time.Time) (ReleasePathResult, error) {
	if !strings.Contains(strings.ToLower(line), "current") {
		return ReleasePathResult{
			Status:     "stable",
			KernelLine: line,
		}, nil
	}

	idx := strings.Index(line, ": ")
	if idx < 0 {
		return ReleasePathResult{}, fmt.Errorf("cannot parse build date from: %s", line)
	}

	dateStr := line[idx+2:]
	buildDate, err := time.Parse("Mon Jan 2 15:04:05 MST 2006", dateStr)
	if err != nil {
		return ReleasePathResult{}, fmt.Errorf("cannot parse build date %q: %w", dateStr, err)
	}

	res := ReleasePathResult{
		Status:      "pre-release",
		BuildDate:   buildDate,
		ReleaseDate: releaseDate,
		KernelLine:  line,
	}

	if buildDate.After(releaseDate) || buildDate.Equal(releaseDate) {
		res.Status = "post-release"
		return res, ErrDowngradeRisk
	}
	return res, ErrUpgradeRequired
}

// PrintReleasePathBanner prints the decision banner for pre-release snapshots.
// If user confirms, it returns true to indicate sysupgrade should run.
// Does NOT run sysupgrade — the caller decides.
func PrintReleasePathBanner(res ReleasePathResult, version string) bool {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║  PRE-RELEASE SNAPSHOT DETECTED                               ║")
	fmt.Printf("║  Kernel built: %-46s║\n", res.BuildDate.Format("Jan 02, 2006"))
	fmt.Printf("║  %s release:  %-46s║\n", version, res.ReleaseDate.Format("Jan 02, 2006"))
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("You can migrate to the stable release now, or stay on -current.")
	fmt.Println()
	fmt.Printf("[y] Migrate to %s stable — runs: doas sysupgrade -R %s\n", version, version)
	fmt.Println("    (system will reboot into the upgrade kernel)")
	fmt.Println()
	fmt.Println("    After reboot completes, re-run the installer:")
	fmt.Println("    curl -fsSL https://OpenRiot.org/setup.sh | sh")
	fmt.Println()
	fmt.Println("[N] Stay on -current — continue installation normally")
	fmt.Println()
	fmt.Print("Migrate now? [y/N]: ")

	// Open /dev/tty explicitly — stdin may be a pipe (e.g. curl | sh)
	tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
	if err != nil {
		// No TTY available; assume non-interactive and continue
		fmt.Println(" (no terminal detected, continuing...)")
		return false
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

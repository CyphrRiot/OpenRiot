package installer

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
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

// ErrataEntry describes a single security/reliability patch.
type ErrataEntry struct {
	ID          string
	Published   time.Time
	Description string
	Reboot      bool
}

// CheckErrataResult describes what patches are needed.
type CheckErrataResult struct {
	Unapplied []ErrataEntry
	Reboot    bool
	InstallCmd string // e.g. "doas syspatch" or "doas sysupgrade -s"
}

// defaultErrata is the hardcoded patch list for 7.9, used by CheckErrata
// on stable systems (syspatch -l on the local machine).
var defaultErrata = []ErrataEntry{
	{
		ID: "001", Published: time.Date(2026, time.June, 2, 0, 0, 0, 0, time.UTC),
		Description: "X server: dri2, sync, saver, Xkb extensions",
		Reboot:      true,
	},
	{
		ID: "002", Published: time.Date(2026, time.June, 2, 0, 0, 0, 0, time.UTC),
		Description: "smtpd(8): crashing bugs",
		Reboot:      false,
	},
	{
		ID: "003", Published: time.Date(2026, time.June, 2, 0, 0, 0, 0, time.UTC),
		Description: "vmd(8): crashing bugs and -b flag",
		Reboot:      true,
	},
}

// ErrataURL returns the errata page URL for a given OpenBSD version.
func ErrataURL(version string) string {
	v := strings.ReplaceAll(version, ".", "")
	return fmt.Sprintf("https://www.openbsd.org/errata%s.html", v)
}

// FetchErrata fetches and parses the errata page for a given release.
func FetchErrata(version string) ([]ErrataEntry, error) {
	url := ErrataURL(version)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch errata: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read errata: %w", err)
	}

	return ParseErrataHTML(string(body))
}

// ParseErrataHTML parses the OpenBSD errata HTML page into entries.
func ParseErrataHTML(html string) ([]ErrataEntry, error) {
	var entries []ErrataEntry

	chunks := strings.Split(html, `<li id="p`)
	for _, chunk := range chunks[1:] {
		// Extract id from id="p001_xserver"
		idx := strings.Index(chunk, `"`)
		if idx < 0 {
			continue
		}
		idWithPrefix := chunk[:idx]

		// Numeric ID is before the first underscore
		id := idWithPrefix
		if ui := strings.Index(id, "_"); ui >= 0 {
			id = id[:ui]
		}

		// Extract strong tag content: "001: SECURITY FIX: June 2, 2026"
		sIdx := strings.Index(chunk, "<strong>")
		if sIdx < 0 {
			continue
		}
		sEnd := strings.Index(chunk[sIdx:], "</strong>")
		if sEnd < 0 {
			continue
		}
		strongContent := chunk[sIdx+8 : sIdx+sEnd]

		// Date is after the last ": "
		colonIdx := strings.LastIndex(strongContent, ": ")
		if colonIdx < 0 {
			continue
		}
		dateStr := strings.TrimSpace(strongContent[colonIdx+2:])

		published, err := time.Parse("January 2, 2006", dateStr)
		if err != nil {
			published, err = time.Parse("Jan 2, 2006", dateStr)
			if err != nil {
				continue
			}
		}

		// Description: text after the first <br/> following </strong>
		afterStrong := chunk[sIdx+sEnd+7:]
		brIdx := strings.Index(afterStrong, "<br/>")
		if brIdx < 0 {
			continue
		}
		descPart := afterStrong[brIdx+5:]

		// Read until next <br/>, <a, or </li>
		end := len(descPart)
		if n := strings.Index(descPart, "<br/>"); n >= 0 && n < end {
			end = n
		}
		if n := strings.Index(descPart, "<a "); n >= 0 && n < end {
			end = n
		}
		if n := strings.Index(descPart, "</li>"); n >= 0 && n < end {
			end = n
		}

		desc := strings.TrimSpace(descPart[:end])
		desc = strings.Join(strings.Fields(desc), " ")

		// Reboot inference: check ID keywords and description
		reboot := false
		lowerID := strings.ToLower(idWithPrefix)
		if strings.Contains(lowerID, "xserver") ||
			strings.Contains(lowerID, "kernel") ||
			strings.Contains(lowerID, "drm") ||
			strings.Contains(lowerID, "vmd") ||
			strings.Contains(lowerID, "wscons") ||
			strings.Contains(lowerID, "xenocara") {
			reboot = true
		}
		if !reboot {
			lowerDesc := strings.ToLower(desc)
			if strings.Contains(lowerDesc, "x server") ||
				strings.Contains(lowerDesc, "kernel") ||
				strings.Contains(lowerDesc, "xenocara") ||
				strings.Contains(lowerDesc, "drm") ||
				strings.Contains(lowerDesc, "vmd") {
				reboot = true
			}
		}

		entries = append(entries, ErrataEntry{
			ID:          id,
			Published:   published,
			Description: desc,
			Reboot:      reboot,
		})
	}

	return entries, nil
}

// CheckErrata checks for missing security/reliability patches.
// On stable, it checks syspatch -l against the hardcoded errata list.
// On -current, it compares kernel build date against errata publish
// dates — users running a kernel older than a patch need to sysupgrade.
func CheckErrata() (*CheckErrataResult, error) {
	output, err := exec.Command("sysctl", "-n", "kern.version").Output()
	if err != nil {
		return nil, fmt.Errorf("cannot read kern.version: %w", err)
	}
	line := strings.TrimSpace(string(output))
	if nl := strings.Index(line, "\n"); nl >= 0 {
		line = line[:nl]
	}

	if strings.Contains(strings.ToLower(line), "current") {
		return checkErrataCurrent(line)
	}
	return checkErrataStable()
}

// checkErrataCurrent determines which errata apply based on kernel
// build date. A kernel built before a patch's publish date is
// vulnerable and needs sysupgrade -s to get the fix.
func checkErrataCurrent(kernelLine string) (*CheckErrataResult, error) {
	idx := strings.Index(kernelLine, ": ")
	if idx < 0 {
		return &CheckErrataResult{InstallCmd: "doas sysupgrade -s"}, nil
	}
	dateStr := kernelLine[idx+2:]
	buildDate, err := time.Parse("Mon Jan 2 15:04:05 MST 2006", dateStr)
	if err != nil {
		return &CheckErrataResult{InstallCmd: "doas sysupgrade -s"}, nil
	}

	var unapplied []ErrataEntry
	reboot := false
	for _, e := range defaultErrata {
		if buildDate.Before(e.Published) {
			unapplied = append(unapplied, e)
			if e.Reboot {
				reboot = true
			}
		}
	}
	return &CheckErrataResult{Unapplied: unapplied, Reboot: reboot, InstallCmd: "doas sysupgrade -s"}, nil
}

// checkErrataStable checks for unapplied patches via syspatch -l.
func checkErrataStable() (*CheckErrataResult, error) {
	cmd := exec.Command("syspatch", "-l")
	output, err := cmd.Output()
	if err != nil {
		return &CheckErrataResult{}, nil
	}

	applied := make(map[string]bool)
	for _, id := range strings.Fields(string(output)) {
		applied[id] = true
	}

	var unapplied []ErrataEntry
	reboot := false
	for _, e := range defaultErrata {
		if !applied[e.ID] {
			unapplied = append(unapplied, e)
			if e.Reboot {
				reboot = true
			}
		}
	}
	return &CheckErrataResult{Unapplied: unapplied, Reboot: reboot, InstallCmd: "doas syspatch"}, nil
}

// PrintErrataBanner prints a prompt to install missing patches.
// Returns true if the user confirms.
func PrintErrataBanner(res *CheckErrataResult) bool {
	if len(res.Unapplied) == 0 {
		return false
	}

	fmt.Println()
	for _, e := range res.Unapplied {
		fmt.Printf("[INFO] A Security Patch Exists for your system\n")
		fmt.Printf("[INFO] Fixes %s\n", e.Description)
	}
	if res.Reboot {
		fmt.Println("[WARN] Requires a system reboot!")
	}
	fmt.Println()
	fmt.Print("[ASK ] Would you like to install? [y/N]: ")

	tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
	if err != nil {
		fmt.Println(" (no terminal detected, continuing...)")
		return false
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

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

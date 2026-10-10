package control

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"openriot/config"
	"openriot/installer"
	"openriot/update"
)

// status is the outcome of a single health check.
type status int

const (
	statusOK status = iota
	statusWarn
	statusFail
)

// check is one health probe. fix is nil for report-only checks.
type check struct {
	name   string
	status status
	detail string
	fix    func() error
}

func statusLabel(s status) string {
	switch s {
	case statusOK:
		return "OK"
	case statusWarn:
		return "WARN"
	default:
		return "FAIL"
	}
}

// runChecks gathers the current health report.
func runChecks() []check {
	return []check{
		syspatchCheck(),
		diskCheck(),
		pfCheck(),
		doasCheck(),
		swapCheck(),
		driftCheck(),
		packageCheck(),
		crashCheck(),
	}
}

// crashIssue returns a short description of the first panic or
// unclean-filesystem marker in dmesg, or "" if none.
func crashIssue(dmesg string) string {
	low := strings.ToLower(dmesg)
	switch {
	case strings.Contains(low, "panic"), strings.Contains(low, "uvm_fault"):
		return "panic logged"
	case strings.Contains(low, "not properly unmounted"):
		return "unclean filesystem"
	}
	return ""
}

// crashCores counts kernel crash dumps in /var/crash (excluding minfree).
func crashCores() int {
	entries, err := os.ReadDir("/var/crash")
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if e.Name() != "minfree" {
			n++
		}
	}
	return n
}

// crashCheck looks for recent panics (dmesg) and crash dumps. Report-only.
func crashCheck() check {
	c := check{name: "Crashes", status: statusOK, detail: "none"}
	if out, err := exec.Command("dmesg").Output(); err == nil {
		if issue := crashIssue(string(out)); issue != "" {
			c.status = statusFail
			c.detail = issue
			return c
		}
	}
	if n := crashCores(); n > 0 {
		c.status = statusFail
		c.detail = fmt.Sprintf("%d core(s) in /var/crash", n)
	}
	return c
}

// packageDrift counts yaml packages that are missing or at a different
// version when compared against the installed set (base name -> full pkg).
func packageDrift(yamlPkgs []string, installed map[string]string) (missing, mismatch int) {
	for _, pkg := range yamlPkgs {
		got, ok := installed[config.GetBaseName(pkg)]
		if !ok {
			missing++
		} else if got != pkg {
			mismatch++
		}
	}
	return missing, mismatch
}

// packageCheck reports installed packages that don't match packages.yaml.
func packageCheck() check {
	c := check{name: "Packages", status: statusOK, detail: "in sync"}
	path := config.FindConfigFile()
	if path == "" {
		c.status = statusWarn
		c.detail = "no packages.yaml"
		return c
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		c.status = statusWarn
		c.detail = "unreadable"
		return c
	}
	installed := installer.GetInstalledPackages()
	if len(installed) == 0 {
		c.status = statusWarn
		c.detail = "pkg_info unavailable"
		return c
	}
	missing, mismatch := packageDrift(cfg.GetPackages(), installed)
	if missing == 0 && mismatch == 0 {
		return c
	}
	c.status = statusWarn
	c.detail = fmt.Sprintf("%d missing, %d outdated", missing, mismatch)
	c.fix = launchInstallPackages
	return c
}

// driftCheck reports kernel/package drift, a panic risk on -current.
func driftCheck() check {
	c := check{name: "Kernel", status: statusOK, detail: "in sync"}
	drift, buildDate := update.KernelDrift()
	if !drift {
		return c
	}
	c.status = statusFail
	c.detail = fmt.Sprintf("%d days old", int(time.Since(buildDate).Hours()/24))
	c.fix = runUpdate
	return c
}

// swapStatus classifies swap usage by used percentage.
func swapStatus(pct int) status {
	switch {
	case pct >= 75:
		return statusFail
	case pct >= 25:
		return statusWarn
	default:
		return statusOK
	}
}

// parseSwap reads total/used 512-block counts from `swapctl -s` output.
func parseSwap(out string) (total, used int64, ok bool) {
	fields := strings.FieldsFunc(out, func(r rune) bool {
		return r == ' ' || r == ',' || r == ':' || r == '\n' || r == '\t'
	})
	for i, f := range fields {
		if f == "total" && i+1 < len(fields) {
			total, _ = strconv.ParseInt(fields[i+1], 10, 64)
		}
		if f == "used" && i-1 >= 0 {
			used, _ = strconv.ParseInt(fields[i-1], 10, 64)
		}
	}
	return total, used, total > 0
}

// swapCheck reports swap pressure. Report-only.
func swapCheck() check {
	c := check{name: "Swap", status: statusOK, detail: "unused"}
	out, err := exec.Command("swapctl", "-s").Output()
	if err != nil {
		c.status = statusWarn
		c.detail = "unavailable"
		return c
	}
	total, used, ok := parseSwap(string(out))
	if !ok {
		c.detail = "none configured"
		return c
	}
	pct := int(used * 100 / total)
	c.status = swapStatus(pct)
	c.detail = fmt.Sprintf("%d%% used", pct)
	return c
}

// doasConfOK reports whether a doas.conf is owned by root and not
// writable by group or other.
func doasConfOK(perm, owner string) bool {
	if owner != "root" || len(perm) < 10 {
		return false
	}
	return perm[5] != 'w' && perm[8] != 'w'
}

// doasCheck reports whether /etc/doas.conf has safe ownership and
// permissions. Report-only: a broken doas cannot fix itself via doas.
func doasCheck() check {
	c := check{name: "doas", status: statusOK, detail: "safe"}
	out, err := exec.Command("stat", "-f", "%Sp %Su", "/etc/doas.conf").Output()
	if err != nil {
		c.status = statusWarn
		c.detail = "missing"
		return c
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 2 {
		c.status = statusWarn
		c.detail = "unknown"
		return c
	}
	if doasConfOK(fields[0], fields[1]) {
		return c
	}
	c.status = statusFail
	c.detail = fmt.Sprintf("insecure %s %s", fields[0], fields[1])
	return c
}

// pfEnabled reports whether pf is enabled, from `rcctl get pf` output.
func pfEnabled(out string) bool {
	return strings.Contains(strings.ToLower(out), "yes")
}

// pfCheck reports whether the packet filter is enabled.
func pfCheck() check {
	c := check{name: "Firewall", status: statusOK, detail: "pf enabled"}
	out, err := exec.Command("rcctl", "get", "pf").Output()
	if err != nil {
		c.status = statusWarn
		c.detail = "unknown"
		return c
	}
	if pfEnabled(string(out)) {
		return c
	}
	c.status = statusFail
	c.detail = "pf disabled"
	c.fix = enablePf
	return c
}

func enablePf() error {
	if err := exec.Command("doas", "rcctl", "enable", "pf").Run(); err != nil {
		return err
	}
	exec.Command("doas", "pfctl", "-f", "/etc/pf.conf").Run()
	return exec.Command("doas", "pfctl", "-e").Run()
}

// diskStatus classifies a filesystem by its free space in gigabytes.
func diskStatus(freeGB int64) status {
	switch {
	case freeGB < 1:
		return statusFail
	case freeGB < 5:
		return statusWarn
	default:
		return statusOK
	}
}

// diskCheck reports the filesystem with the least free space of
// /, /home and /var.
func diskCheck() check {
	c := check{name: "Disk", status: statusOK, detail: "healthy"}
	out, err := exec.Command("df", "-P", "-k").Output()
	if err != nil {
		c.status = statusWarn
		c.detail = "unavailable"
		return c
	}

	worstFreeKB := int64(-1)
	worstMount := ""
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		mount := fields[len(fields)-1]
		if mount != "/" && mount != "/home" && mount != "/var" {
			continue
		}
		freeKB, err := strconv.ParseInt(fields[len(fields)-3], 10, 64)
		if err != nil {
			continue
		}
		if worstFreeKB < 0 || freeKB < worstFreeKB {
			worstFreeKB = freeKB
			worstMount = mount
		}
	}

	if worstFreeKB < 0 {
		c.status = statusWarn
		c.detail = "unavailable"
		return c
	}

	freeGB := worstFreeKB / (1024 * 1024)
	c.status = diskStatus(freeGB)
	c.detail = fmt.Sprintf("%s %d GB free", worstMount, freeGB)
	return c
}

// syspatchCheck reports pending security patches on stable releases.
func syspatchCheck() check {
	c := check{name: "Syspatch", status: statusOK, detail: "up to date"}
	if config.DetectOpenBSDVersion() == "snapshots" {
		c.detail = "n/a on -current"
		return c
	}
	out, err := exec.Command("doas", "syspatch", "-c").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return c
	}
	c.status = statusFail
	c.detail = fmt.Sprintf("%d pending", len(strings.Fields(string(out))))
	c.fix = applySyspatch
	return c
}

func applySyspatch() error {
	return exec.Command("doas", "syspatch").Run()
}

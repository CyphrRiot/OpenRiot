package nmtui

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// WiFiAP represents a scanned wireless access point.
type WiFiAP struct {
	SSID        string
	BSSID       string
	Signal      int    // Signal strength percentage (0-100), -1 if unknown
	SignalValid bool   // true if Signal was parsed from scan output
	Security    string // "open", "wpa2", "wpa3", "wep", or "unknown"
}

// ConnectionInfo holds active connection details for a Wi-Fi interface.
type ConnectionInfo struct {
	Device  string
	SSID    string
	IP      string
	Netmask string
	Gateway string
	DNS     []string
	MAC     string
	State   string
}

var (
	ifaceLineRE    = regexp.MustCompile(`^([a-z]+[0-9]+):\s+flags=`)
	scanLineRE     = regexp.MustCompile(`^\s+nwid\s+(.+?)(?:\s+|$)`)
	signalRE       = regexp.MustCompile(`(\d+)%`)
	bssidRE        = regexp.MustCompile(`bssid\s+([0-9a-fA-F:]{17})`)
	joinSSIDRE     = regexp.MustCompile(`join\s+("[^"]*"|\S+)`)
	inetRE         = regexp.MustCompile(`\binet\s+(\S+)`)
	lladdrRE       = regexp.MustCompile(`lladdr\s+([0-9a-fA-F:]{17})`)
	statusRE       = regexp.MustCompile(`status:\s+(\S+)`)
	defaultRouteRE = regexp.MustCompile(`^default\s+\S+\s+(\S+)`)
	nwidRE         = regexp.MustCompile(`nwid\s+("[^"]*"|\S+)`)
	hexSSIDRE      = regexp.MustCompile(`^0x([0-9a-fA-F]+)$`)
)

// FindWiFiInterface detects the first wireless interface present on the system,
// even if it is not currently connected.
func FindWiFiInterface() (string, error) {
	out, err := exec.Command("/sbin/ifconfig", "-a").Output()
	if err != nil {
		return "", fmt.Errorf("ifconfig -a failed: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	var current string

	for _, line := range lines {
		if m := ifaceLineRE.FindStringSubmatch(line); m != nil {
			current = m[1]
			continue
		}
		if current != "" && strings.Contains(line, "ieee80211") {
			return current, nil
		}
	}

	return "", fmt.Errorf("no wifi interface found")
}

// ScanWiFi runs `ifconfig <iface> scan` and returns parsed access points.
func ScanWiFi(iface string) ([]WiFiAP, error) {
	out, err := exec.Command("/sbin/ifconfig", iface, "scan").Output()
	if err != nil {
		return nil, fmt.Errorf("ifconfig %s scan failed: %w", iface, err)
	}
	return parseScanOutput(string(out)), nil
}

// IsWiFiConnected returns true if the interface is joined to a network.
func IsWiFiConnected(iface string) bool {
	out, err := exec.Command("/sbin/ifconfig", iface).Output()
	if err != nil {
		return false
	}
	return joinSSIDRE.MatchString(string(out))
}

// SignalToPercent returns the percentage signal strength.
// If valid is false, returns -1. Otherwise returns signal clamped to [0,100].
func SignalToPercent(signal int, valid bool) int {
	if !valid {
		return -1
	}
	if signal >= 100 {
		return 100
	}
	if signal <= 0 {
		return 0
	}
	return signal
}

// ConnectOpen associates with an open network and persists it to
// /etc/hostname.{iface}.
func ConnectOpen(iface, ssid string) error {
	if err := requireDoas(); err != nil {
		return err
	}
	if err := writeHostnameIf(iface, ssid, ""); err != nil {
		return err
	}
	if out, err := exec.Command("doas", "ifconfig", iface, "down").CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig down failed: %w\n%s", err, string(out))
	}
	if out, err := exec.Command("doas", "ifconfig", iface, "up").CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig up failed: %w\n%s", err, string(out))
	}
	cmd := exec.Command("doas", "sh", "/etc/netstart", iface)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("netstart failed: %w\n%s", err, string(out))
	}
	return nil
}

// ConnectWPA associates with a WPA2-PSK network and persists it to
// /etc/hostname.{iface}.
func ConnectWPA(iface, ssid, key string) error {
	if err := requireDoas(); err != nil {
		return err
	}
	if err := writeHostnameIf(iface, ssid, key); err != nil {
		return err
	}
	if out, err := exec.Command("doas", "ifconfig", iface, "down").CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig down failed: %w\n%s", err, string(out))
	}
	if out, err := exec.Command("doas", "ifconfig", iface, "up").CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig up failed: %w\n%s", err, string(out))
	}
	cmd := exec.Command("doas", "sh", "/etc/netstart", iface)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("netstart failed: %w\n%s", err, string(out))
	}
	return nil
}

// Disconnect brings the interface down and clears the nwid line from
// /etc/hostname.{iface}.
func Disconnect(iface string) error {
	if err := requireDoas(); err != nil {
		return err
	}
	if err := clearHostnameIf(iface); err != nil {
		return err
	}
	cmd := exec.Command("doas", "ifconfig", iface, "down")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig down failed: %w\n%s", err, string(out))
	}
	return nil
}

// GetConnectionInfo parses `ifconfig`, `netstat`, and `/etc/resolv.conf` for
// connection details.
func GetConnectionInfo(iface string) (*ConnectionInfo, error) {
	out, err := exec.Command("/sbin/ifconfig", iface).Output()
	if err != nil {
		return nil, fmt.Errorf("ifconfig %s failed: %w", iface, err)
	}
	info := parseIfconfigConnection(string(out))
	info.Device = iface

	// Gateway from netstat
	if g, err := getDefaultGateway(); err == nil {
		info.Gateway = g
	}

	// DNS from resolv.conf
	info.DNS = readResolvDNS()

	return info, nil
}

// GetKnownSSIDs reads /etc/hostname.<iface> and extracts configured SSIDs.
func GetKnownSSIDs(iface string) []string {
	data, err := os.ReadFile("/etc/hostname." + iface)
	if err != nil {
		return nil
	}
	var ssids []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		for _, m := range nwidRE.FindAllStringSubmatch(line, -1) {
			ssid := stripQuotes(m[1])
			if ssid != "" {
				ssids = append(ssids, ssid)
			}
		}
	}
	return ssids
}

// requireDoas returns an error if doas is not available in PATH.
func requireDoas() error {
	if _, err := exec.LookPath("doas"); err != nil {
		return fmt.Errorf("doas is required but not found in PATH")
	}
	return nil
}

// writeHostnameIf writes the nwid line to /etc/hostname.{iface}, preserving
// all other lines (dhcp, inet, etc.). The SSID is always quoted.
func writeHostnameIf(iface, ssid, key string) error {
	path := "/etc/hostname." + iface
	var lines []string

	data, err := os.ReadFile(path)
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "nwid ") {
				continue
			}
			lines = append(lines, line)
		}
	}

	escaped := strings.ReplaceAll(ssid, `"`, `\"`)
	nwidLine := fmt.Sprintf(`nwid "%s"`, escaped)
	if key != "" {
		nwidLine += " wpakey " + key
	}
	lines = append([]string{nwidLine}, lines...)

	content := strings.Join(lines, "\n") + "\n"
	cmd := exec.Command("doas", "sh", "-c", "cat > "+path)
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("writing %s failed: %w\n%s", path, err, string(out))
	}
	return nil
}

// clearHostnameIf removes all nwid lines from /etc/hostname.{iface}. If the
// file becomes empty it is removed entirely.
func clearHostnameIf(iface string) error {
	path := "/etc/hostname." + iface
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // file does not exist, nothing to do
	}

	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "nwid ") {
			continue
		}
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		cmd := exec.Command("doas", "rm", "-f", path)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("removing %s failed: %w\n%s", path, err, string(out))
		}
		return nil
	}

	content := strings.Join(lines, "\n") + "\n"
	cmd := exec.Command("doas", "sh", "-c", "cat > "+path)
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("writing %s failed: %w\n%s", path, err, string(out))
	}
	return nil
}

// --- Parsers ---

func parseScanOutput(raw string) []WiFiAP {
	seen := make(map[string]int) // SSID -> index in aps
	var aps []WiFiAP
	lines := strings.Split(raw, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "nwid") {
			continue
		}

		rest := strings.TrimPrefix(line, "nwid ")
		ssid := extractSSID(rest)
		if ssid == "" {
			continue // skip hidden networks
		}

		ap := WiFiAP{SSID: ssid, Security: "unknown"}

		if m := signalRE.FindStringSubmatch(line); m != nil {
			pct, _ := strconv.Atoi(m[1])
			ap.Signal = pct
			ap.SignalValid = true
		}

		if m := bssidRE.FindStringSubmatch(line); m != nil {
			ap.BSSID = m[1]
		}

		// Security from comma-separated flags at end of line
		switch {
		case strings.Contains(line, "wpa3"):
			ap.Security = "wpa3"
		case strings.Contains(line, "wpa2"):
			ap.Security = "wpa2"
		case strings.Contains(line, "wep"):
			ap.Security = "wep"
		case strings.Contains(line, "privacy"):
			ap.Security = "unknown"
		default:
			ap.Security = "open"
		}

		// Deduplicate by SSID, keep strongest signal
		if idx, ok := seen[ssid]; ok {
			if ap.SignalValid && (!aps[idx].SignalValid || ap.Signal > aps[idx].Signal) {
				aps[idx] = ap
			}
		} else {
			seen[ssid] = len(aps)
			aps = append(aps, ap)
		}
	}

	return aps
}

func extractSSID(rest string) string {
	if strings.HasPrefix(rest, `""`) {
		return ""
	}

	if strings.HasPrefix(rest, `"`) {
		end := strings.Index(rest[1:], `"`)
		if end >= 0 {
			return rest[1 : end+1]
		}
		if len(rest) > 1 {
			return rest[1:]
		}
		return ""
	}

	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return ""
	}
	if isScanKeyword(parts[0]) {
		return ""
	}

	// Hex-encoded SSID: 0x47c3b664656c -> "Gödel"
	if m := hexSSIDRE.FindStringSubmatch(parts[0]); m != nil {
		if decoded, err := hex.DecodeString(m[1]); err == nil {
			if utf8.Valid(decoded) {
				return string(decoded)
			}
		}
	}

	var ssidParts []string
	for _, p := range parts {
		if isScanKeyword(p) || signalRE.MatchString(p) {
			break
		}
		ssidParts = append(ssidParts, p)
	}
	return strings.Join(ssidParts, " ")
}

func isScanKeyword(s string) bool {
	switch s {
	case "bssid", "chan", "media", "wpaprotos", "wpaakms", "wpaciphers",
		"groupcipher", "bgscan", "join", "powersave", "dtimperiod", "closed",
		"adhoc", "nwflag", "pka", "chanlist", "pureg", "puren", "htmode",
		"bintval", "ampdu", "amsdu", "open":
		return true
	}
	return false
}

func parseIfconfigConnection(output string) *ConnectionInfo {
	info := &ConnectionInfo{State: "disconnected"}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if m := joinSSIDRE.FindStringSubmatch(line); m != nil {
			info.SSID = stripQuotes(m[1])
			info.State = "active"
		}

		if m := inetRE.FindStringSubmatch(line); m != nil {
			parts := strings.Split(m[1], "/")
			info.IP = parts[0]
			if len(parts) > 1 {
				info.Netmask = parts[1]
			}
		}

		if m := lladdrRE.FindStringSubmatch(line); m != nil {
			info.MAC = m[1]
		}

		if m := statusRE.FindStringSubmatch(line); m != nil {
			info.State = m[1]
		}
	}

	return info
}

func getDefaultGateway() (string, error) {
	out, err := exec.Command("/usr/bin/netstat", "-rn", "-f", "inet").Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if m := defaultRouteRE.FindStringSubmatch(line); m != nil {
			return m[1], nil
		}
	}
	return "", fmt.Errorf("no default gateway found")
}

func readResolvDNS() []string {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil
	}
	var servers []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver ") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				servers = append(servers, fields[1])
			}
		}
	}
	return servers
}

func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

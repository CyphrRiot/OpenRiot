package macspoof

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"openriot/logger"
)

// Interface represents a network interface with MAC info
type Interface struct {
	Name        string
	Type        string // wifi, ethernet, other
	MAC         string
	LLAddr      string // configured lladdr
	Randomized  bool   // lladdr random configured in hostname file
}

// GetInterfaces returns all detected network interfaces
func GetInterfaces() ([]Interface, error) {
	cmd := exec.Command("/sbin/ifconfig")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ifconfig: %w", err)
	}

	var interfaces []Interface
	var current Interface
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		// New interface line: e.g., iwx0: flags=...
		if matched, _ := regexp.MatchString(`^[a-z]+[0-9]+:`, line); matched {
			// Save previous interface if we have one
			if current.Name != "" {
				interfaces = append(interfaces, current)
			}
			// Parse new interface name
			name := strings.TrimSuffix(strings.SplitN(line, ":", 2)[0], ":")
			current = Interface{Name: name}
			continue
		}

		// Detect interface type
		if strings.Contains(line, "ieee80211") {
			current.Type = "wifi"
		} else if strings.Contains(line, "media: Ethernet") {
			current.Type = "ethernet"
		}

		// Extract current MAC address
		if strings.Contains(line, "lladdr") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "lladdr" && i+1 < len(parts) {
					current.LLAddr = parts[i+1]
					current.MAC = parts[i+1]
					break
				}
			}
		}
	}

	// Don't forget the last interface
	if current.Name != "" {
		interfaces = append(interfaces, current)
	}

	return interfaces, nil
}

// GetNetworkInterfaces returns only wifi and ethernet interfaces
func GetNetworkInterfaces() ([]Interface, error) {
	all, err := GetInterfaces()
	if err != nil {
		return nil, err
	}

	var filtered []Interface
	for _, iface := range all {
		if iface.Type == "wifi" || iface.Type == "ethernet" {
			filtered = append(filtered, iface)
		}
	}
	return filtered, nil
}

// getHostnameConfig checks if lladdr random is configured in hostname file
func getHostnameConfig(ifaceName string) (hasRandom bool, err error) {
	content, err := os.ReadFile("/etc/hostname." + ifaceName)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return strings.Contains(string(content), "lladdr random"), nil
}

// EnableRandomMAC adds 'lladdr random' to the interface's hostname file
func EnableRandomMAC(iface string) error {
	hostnameFile := "/etc/hostname." + iface

	// Check if file exists
	if _, err := os.Stat(hostnameFile); os.IsNotExist(err) {
		return fmt.Errorf("interface %s has no hostname file", iface)
	}

	// Simple shell command: prepend if not already there
	cmd := exec.Command("doas", "sh", "-c",
		fmt.Sprintf(`grep -q "^lladdr random$" %s && exit; (echo "lladdr random"; cat %s) > /tmp/hostname.tmp && mv /tmp/hostname.tmp %s`, hostnameFile, hostnameFile, hostnameFile))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable random MAC: %w", err)
	}

	return nil
}

// DisableRandomMAC removes 'lladdr random' from the interface's hostname file
func DisableRandomMAC(iface string) error {
	hostnameFile := "/etc/hostname." + iface

	// Simple shell command: remove line if it exists
	cmd := exec.Command("doas", "sed", "-i", "/^lladdr random$/d", hostnameFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable random MAC: %w", err)
	}

	return nil
}

// ApplyMAC restarts the interface using netstart to apply changes
func ApplyMAC(iface string) error {
	cmd := exec.Command("doas", "sh", "/etc/netstart", iface)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart interface %s: %w", iface, err)
	}
	return nil
}

// IsStealthEnabled checks if any interface has random MAC configured
func IsStealthEnabled() bool {
	interfaces, err := GetNetworkInterfaces()
	if err != nil || len(interfaces) == 0 {
		return false
	}
	for _, iface := range interfaces {
		hasRandom, _ := getHostnameConfig(iface.Name)
		if hasRandom {
			return true
		}
	}
	return false
}

// StealthStatus returns the appropriate icon for polybar
func StealthStatus() string {
	if IsStealthEnabled() {
		return "󰝴" // Stealth ON
	}
	return "󱊨" // Stealth OFF
}

// StealthToggle enables or disables stealth mode
func StealthToggle() error {
	if IsStealthEnabled() {
		return runDisable()
	}
	return runEnable()
}

// Run is the main entry point for the --random-mac command
func Run(args []string) error {
	if len(args) == 0 {
		return runShow()
	}

	switch args[0] {
	case "show":
		return runShow()
	case "enable":
		return runEnable()
	case "disable":
		return runDisable()
	case "help":
		fmt.Println(usage)
		return nil
	default:
		return fmt.Errorf("unknown subcommand: %s\n%s", args[0], usage)
	}
}

const usage = `Usage: openriot --random-mac <subcommand>
Subcommands:
  show     Show current MAC addresses for all interfaces
  enable   Enable random MAC on detected interfaces
  disable  Remove random MAC configuration`

func runShow() error {
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		return fmt.Errorf("failed to get interfaces: %w", err)
	}

	if len(interfaces) == 0 {
		fmt.Println("No wifi/ethernet interfaces found")
		return nil
	}

	for _, iface := range interfaces {
		// Check if lladdr random is configured in hostname file
		hasRandom, err := getHostnameConfig(iface.Name)
		if err != nil {
			hasRandom = false
		}

		status := ""
		if hasRandom {
			status = " [Stealth]"
		}

		mac := iface.MAC
		if mac == "" {
			mac = "N/A"
		}
		fmt.Printf("%s (%s): %s%s\n", iface.Name, iface.Type, mac, status)
	}

	return nil
}

func runEnable() error {
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		return fmt.Errorf("failed to get interfaces: %w", err)
	}

	enabled := 0
	for _, iface := range interfaces {
		// Check if hostname file exists (or is accessible)
		if _, err := os.Stat("/etc/hostname." + iface.Name); err != nil {
			fmt.Printf("[SKIP] %s: no hostname file\n", iface.Name)
			continue
		}

		// Check if already enabled
		hasRandom, _ := getHostnameConfig(iface.Name)
		if hasRandom {
			fmt.Printf("[SKIP] %s: random MAC already enabled\n", iface.Name)
			continue
		}

		if err := EnableRandomMAC(iface.Name); err != nil {
			return fmt.Errorf("%s: %v", iface.Name, err)
		}
		fmt.Printf("[OK] %s: random MAC enabled\n", iface.Name)
		enabled++

		// Apply immediately
		if err := ApplyMAC(iface.Name); err != nil {
			logger.Warn(fmt.Sprintf("%s: failed to apply (reboot to activate): %v", iface.Name, err))
		}
	}

	if enabled == 0 {
		fmt.Println("No interfaces configured (no hostname files)")
		return nil
	}

	fmt.Printf("\nRandom MAC enabled on %d interface(s)\n", enabled)
	fmt.Println("Changes will take full effect on next reboot")
	return nil
}

func runDisable() error {
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		return fmt.Errorf("failed to get interfaces: %w", err)
	}

	disabled := 0
	for _, iface := range interfaces {
		if err := DisableRandomMAC(iface.Name); err != nil {
			logger.Warn(fmt.Sprintf("%s: %v", iface.Name, err))
			continue
		}
		fmt.Printf("[OK] %s: random MAC disabled\n", iface.Name)
		disabled++

		// Apply immediately
		if err := ApplyMAC(iface.Name); err != nil {
			logger.Warn(fmt.Sprintf("%s: failed to apply (reboot to activate): %v", iface.Name, err))
		}
	}

	if disabled == 0 {
		fmt.Println("No interfaces had random MAC configured")
		return nil
	}

	fmt.Printf("\nRandom MAC disabled on %d interface(s)\n", disabled)
	return nil
}

package imaging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"openriot/logger"
)

// Drive represents a detected drive
type Drive struct {
	Device      string
	SizeGB      int
	IsRemovable bool
	IsProtected bool
	IsBootDrive bool
	Status      string // "ROOT", "WARN", "INFO"
}

// DetectDrives scans for available drives
func DetectDrives() ([]Drive, error) {
	var drives []Drive

	// Get root drive from dmesg
	rootDrive := ""
	cmd := exec.Command("dmesg")
	out, err := cmd.Output()
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "root on ") {
				// Extract sdX from "root on sd1a"
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					device := fields[2] // "sd1a" from "root on sd1a"
					rootDrive = strings.TrimSuffix(device, "a") // Remove partition
				}
				break
			}
		}
	}

	// Get removable drives from dmesg
	removable := make(map[string]bool)
	cmd = exec.Command("dmesg")
	out, _ = cmd.Output()
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "removable") && strings.Contains(line, "sd") {
			// Extract sdX
			fields := strings.Fields(line)
			for _, f := range fields {
				if strings.HasPrefix(f, "sd") && len(f) <= 4 {
					removable[f] = true
				}
			}
		}
	}

	// Get all disks from sysctl
	cmd = exec.Command("sysctl", "-n", "hw.disknames")
	out, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	disknames := strings.TrimSpace(string(out))
	for _, disk := range strings.Split(disknames, ",") {
		disk = strings.TrimSpace(disk)
		if !strings.HasPrefix(disk, "sd") && !strings.HasPrefix(disk, "wd") {
			continue
		}

		// Extract device name (e.g., "sd0" from "sd0:...")
		device := disk
		if idx := strings.Index(device, ":"); idx > 0 {
			device = device[:idx]
		}

		// Skip if not sd* (we only care about SCSI/SATA)
		if !strings.HasPrefix(device, "sd") {
			continue
		}

		// Get disklabel for size and RAID info
		sizeGB, isProtected, err := getDiskInfo(device)
		if err != nil {
			continue
		}

		status := "INFO"
		isRemovable := removable[device]

		if device == rootDrive || isProtected {
			status = "ROOT"
		} else if isRemovable {
			status = "WARN"
		}

		drives = append(drives, Drive{
			Device:      device,
			SizeGB:      sizeGB,
			IsRemovable: isRemovable,
			IsProtected: device == rootDrive || isProtected,
			IsBootDrive: device == rootDrive,
			Status:      status,
		})
	}

	return drives, nil
}

// getDiskInfo returns size in GB and whether it has RAID partitions
func getDiskInfo(device string) (int, bool, error) {
	cmd := exec.Command("doas", "disklabel", device)
	out, err := cmd.Output()
	if err != nil {
		return 0, false, err
	}

	lines := strings.Split(string(out), "\n")
	var bytesPerSec, totalSec int
	hasRAID := false

	for _, line := range lines {
		// Check for RAID partitions BEFORE trimming (shell grep sees untrimmed lines)
		// Pattern: lines like "  a:        ...    RAID"
		if matched, _ := regexp.MatchString(`^  [a-z]:.*RAID`, line); matched {
			hasRAID = true
		}
		// Trim for size parsing
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "bytes/sector:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				bytesPerSec, _ = strconv.Atoi(fields[1])
			}
		}
		if strings.HasPrefix(line, "total sectors:") || strings.HasPrefix(line, "c:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				totalSec, _ = strconv.Atoi(fields[1])
			}
		}
	}

	if bytesPerSec == 0 || totalSec == 0 {
		return 0, false, fmt.Errorf("could not get disk info")
	}

	totalBytes := bytesPerSec * totalSec
	sizeGB := totalBytes / 1073741824

	return sizeGB, hasRAID, nil
}

// PromptBurn asks user to select a drive and burns image
func PromptBurn(cfg *Config, drives []Drive) error {
	if cfg.NoBurn {
		logger.Info("Skipping burn (--no-burn flag)")
		return nil
	}

	// Build eligible drive list
	var eligible []Drive
	for _, d := range drives {
		if !d.IsProtected {
			eligible = append(eligible, d)
		}
	}

	if len(eligible) == 0 {
		logger.Info("No eligible drives for burning")
		return nil
	}

	// Display all drives
	fmt.Println()
	for _, d := range drives {
		var prefix string
		suffix := ""
		switch d.Status {
		case "ROOT":
			prefix = logger.Red + "[ROOT]" + logger.Reset
			if d.IsBootDrive {
				suffix = " [OpenBSD Encrypted]"
			} else {
				suffix = " [OpenBSD]"
			}
		case "WARN":
			prefix = logger.Yellow + "[WARN]" + logger.Reset
			suffix = " [Removable USB]"
		default:
			prefix = logger.Cyan + "[INFO]" + logger.Reset
		}
		fmt.Printf("%s %s - %5d GB%s\n", prefix, d.Device, d.SizeGB, suffix)
	}

	// Build prompt list
	var driveList []string
	for _, d := range eligible {
		driveList = append(driveList, d.Device)
	}

	logger.Done("Available for burn: " + strings.Join(driveList, ", "))
	logger.Warn("THIS WILL ERASE ALL DATA ON THE SELECTED DRIVE.")
	fmt.Printf("\n%s[ASK ]%s Which drive to burn? (%s or press Enter to skip) ", logger.Cyan, logger.Reset, strings.Join(driveList, ", "))

	stdin := os.Stdin
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		stdin = tty
		defer tty.Close()
	}
	reader := bufio.NewReader(stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		logger.Info(fmt.Sprintf("Skipped. Flash %s to USB when ready.", cfg.OutputImg))
		return nil
	}

	// Validate selection
	var selected Drive
	found := false
	for _, d := range eligible {
		if d.Device == input {
			selected = d
			found = true
			break
		}
	}

	if !found {
		logger.Info(fmt.Sprintf("Invalid selection: %s", input))
		return nil
	}

	// Confirmation
	fmt.Printf("\n%s[ASK ]%s You will be erasing %s (%d GB). Are you sure? [y/N] ", logger.Cyan, logger.Reset, selected.Device, selected.SizeGB)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if strings.ToLower(input) != "y" {
		logger.Info(fmt.Sprintf("Aborted. Flash %s to USB when ready.", cfg.OutputImg))
		return nil
	}

	// Burn
	return burnImage(selected.Device, cfg.OutputImg)
}

// burnImage writes image to specified drive
func burnImage(device, imagePath string) error {
	logger.Info(fmt.Sprintf("Burning to /dev/r%s...", device))

	drivePath := fmt.Sprintf("/dev/r%sc", device)

	// Check if drive exists
	if _, err := os.Stat(drivePath); err != nil {
		return fmt.Errorf("drive not found: %s", drivePath)
	}

	// Use pv for progress if available, otherwise plain dd
	pvCmd := exec.Command("which", "pv")
	if pvCmd.Run() == nil {
		return burnWithPV(imagePath, drivePath)
	}

	return burnWithDD(imagePath, drivePath)
}

// burnWithPV pipes image -> pv -> doas dd using Go's io.Pipe chain.
// No shell — safe with paths containing spaces or special characters.
func burnWithPV(imagePath, drivePath string) error {
	// Open the image file
	img, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	defer img.Close()

	info, err := img.Stat()
	if err != nil {
		return fmt.Errorf("stat image: %w", err)
	}
	sizeMB := float64(info.Size()) / (1024 * 1024)
	logger.Info(fmt.Sprintf("Image size: %.1f MB", sizeMB))

	// Chain: img -> pv -> dd
	pvOut, pvIn := io.Pipe()

	pv := exec.Command("pv", "-pterb", "-s", fmt.Sprintf("%d", info.Size()))
	pv.Stdin = img
	pv.Stdout = pvIn
	pv.Stderr = os.Stderr

	dd := exec.Command("doas", "dd", "of="+drivePath, "bs=1M")
	dd.Stdin = pvOut
	dd.Stdout = os.Stdout
	dd.Stderr = os.Stderr

	if err := pv.Start(); err != nil {
		return fmt.Errorf("start pv: %w", err)
	}
	if err := dd.Start(); err != nil {
		_ = pv.Process.Kill()
		return fmt.Errorf("start dd: %w", err)
	}

	// Wait for pv to finish, then close pipe so dd sees EOF
	if err := pv.Wait(); err != nil {
		_ = dd.Process.Kill()
		return fmt.Errorf("pv: %w", err)
	}
	_ = pvIn.Close()

	if err := dd.Wait(); err != nil {
		return fmt.Errorf("dd: %w", err)
	}
	return nil
}

// burnWithDD runs doas dd directly (no shell) with paths as arguments.
func burnWithDD(imagePath, drivePath string) error {
	cmd := exec.Command("doas", "dd", "if="+imagePath, "of="+drivePath, "bs=1M")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Cleanup removes work directory
func Cleanup(cfg *Config) error {
	logger.Info("Cleaning up...")

	// Unmount just in case
	exec.Command("umount", "/mnt").Run()
	exec.Command("vnconfig", "-u", "vnd0").Run()

	// Remove work dir
	workDir := cfg.WorkDir
	if !strings.HasPrefix(workDir, "/") || strings.Contains(workDir, "..") {
		return fmt.Errorf("refusing to clean unsafe work dir: %s", workDir)
	}
	if err := os.RemoveAll(workDir); err != nil {
		return fmt.Errorf("remove work dir: %w", err)
	}

	logger.Done("Cleanup complete")
	return nil
}

// GetOpenriotVersion returns the current OpenRiot version
func GetOpenriotVersion() string {
	data, err := os.ReadFile(filepath.Join(getBuildDir(), "..", "VERSION"))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}
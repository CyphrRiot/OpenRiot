package disk

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Drive represents a detected storage device.
type Drive struct {
	Device      string
	SizeGB      int
	IsRoot      bool
	IsMounted   bool
	MountPoint  string
	HasRAID     bool
	IsRemovable bool
	IsEncrypted bool // part of a softraid crypto setup (chunk or virtual)
	IsChunk     bool // true if this is a physical chunk backing a softraid volume
	BusType     string // NVMe, USB, CRYPTO, etc.
	ModelName   string // "SanDisk Extreme 55DD", etc.
}

// softraidInfo tracks the relationship between physical chunk devices
// and their virtual softraid volumes.
type softraidInfo struct {
	virtualToPhysical map[string]string // sd0 -> sd1 (physical chunk)
	physicalToVirtual map[string]string // sd1 -> sd0 (virtual device)
}

// DiscoverDrives scans the system for all storage devices.
func DiscoverDrives() ([]Drive, error) {
	// Get root drive from mount (preferred) or dmesg (fallback)
	rootDrive := ""
	cmd := exec.Command("mount")
	out, err := cmd.Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(line, " on / ") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 1 {
				continue
			}
			dev := strings.TrimPrefix(fields[0], "/dev/")
			// Strip partition letter
			for len(dev) > 0 {
				c := dev[len(dev)-1]
				if c < 'a' || c > 'z' {
					break
				}
				dev = dev[:len(dev)-1]
			}
			rootDrive = dev
			break
		}
	}
	if rootDrive == "" {
		cmd = exec.Command("dmesg")
		out, err = cmd.Output()
		if err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "root on ") {
					fields := strings.Fields(line)
					if len(fields) >= 3 {
						device := fields[2]
						device = strings.TrimPrefix(device, "/dev/")
						rootDrive = strings.TrimSuffix(device, "a")
					}
					break
				}
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
			fields := strings.Fields(line)
			for _, f := range fields {
				if strings.HasPrefix(f, "sd") && len(f) <= 4 {
				    removable[f] = true
				}
			}
		}
	}

	// Get softraid info (virtual <-> physical mapping)
	sr := parseSoftraid()

	// Get mount info (maps both physical and virtual devices)
	mounts := parseMounts(sr)

	// Get all disks from sysctl
	cmd = exec.Command("sysctl", "-n", "hw.disknames")
	out, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	var drives []Drive
	disknames := strings.TrimSpace(string(out))
	for _, disk := range strings.Split(disknames, ",") {
		disk = strings.TrimSpace(disk)
		if !strings.HasPrefix(disk, "sd") && !strings.HasPrefix(disk, "wd") {
			continue
		}

		device := disk
		if idx := strings.Index(device, ":"); idx > 0 {
			device = device[:idx]
		}

		if !strings.HasPrefix(device, "sd") {
			continue
		}

		sizeGB, hasRAID, err := getDiskInfo(device)
		if err != nil {
			continue
		}

		// Check if this device is mounted
		mountPoint := ""
		isMounted := false
		if mp, ok := mounts[device]; ok {
			isMounted = true
			mountPoint = mp
		}

		// Determine if this is a root drive (physical root OR virtual root)
		isRoot := device == rootDrive
		if !isRoot && sr.physicalToVirtual[device] != "" {
			// Physical device backing a virtual device
			virtualDev := sr.physicalToVirtual[device]
			if virtualDev == rootDrive {
				isRoot = true
			}
		}
		if !isRoot {
			// Virtual device that is the root
			if physicalDev := sr.virtualToPhysical[device]; physicalDev != "" {
				if physicalDev == rootDrive || device == rootDrive {
					isRoot = true
				}
			}
		}

		// Determine roles
		isEncrypted := sr.virtualToPhysical[device] != "" // virtual softraid volume
		isChunk := sr.physicalToVirtual[device] != ""     // physical chunk backing a volume

		// Fallback: if bioctl failed, use disklabel for IsChunk and dmesg for IsEncrypted
		if !isChunk && hasRAID {
			isChunk = true
		}
		if !isEncrypted && isVirtualOnSoftraidBus(device) {
			isEncrypted = true
		}

		drives = append(drives, Drive{
			Device:      device,
			SizeGB:      sizeGB,
			IsRoot:      isRoot,
			IsMounted:   isMounted,
			MountPoint:  mountPoint,
			HasRAID:     hasRAID,
			IsRemovable: removable[device],
			IsEncrypted: isEncrypted,
			IsChunk:     isChunk,
			BusType:     getDeviceBusType(device),
			ModelName:   getDeviceModel(device),
		})
	}

	return drives, nil
}

// parseMounts reads mount output and maps devices to mount points.
// It also resolves softraid virtual devices back to their physical chunks.
func parseMounts(sr softraidInfo) map[string]string {
	result := make(map[string]string)
	cmd := exec.Command("mount")
	out, err := cmd.Output()
	if err != nil {
		return result
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		dev := fields[0]
		if !strings.HasPrefix(dev, "/dev/sd") && !strings.HasPrefix(dev, "/dev/wd") {
			continue
		}

			// Extract sdX from /dev/sdXi (strip all trailing partition letters)
		base := strings.TrimPrefix(dev, "/dev/")
		for len(base) > 0 {
			c := base[len(base)-1]
			if c < 'a' || c > 'z' {
				break
			}
			base = base[:len(base)-1]
		}

		mp := fields[2]
		result[base] = mp

		// Cross-map: if base is a virtual device, mark its physical chunk too
		if phys := sr.virtualToPhysical[base]; phys != "" {
			result[phys] = mp
		}
		// Reverse cross-map: if base is a physical chunk, mark its virtual device too
		if virt := sr.physicalToVirtual[base]; virt != "" {
			result[virt] = mp
		}
	}
	return result
}

// parseSoftraid reads bioctl output and builds virtual <-> physical mappings.
func parseSoftraid() softraidInfo {
	sr := softraidInfo{
		virtualToPhysical: make(map[string]string),
		physicalToVirtual: make(map[string]string),
	}

	// Method 1: bioctl (primary)
	cmd := exec.Command("doas", "/sbin/bioctl", "softraid0")
	out, _ := cmd.CombinedOutput()
	if len(out) > 0 {
		sr = parseBioctlOutput(string(out))
	}
	if len(sr.virtualToPhysical) > 0 {
		return sr
	}

	// Method 2: dmesg fallback — detect softraid device hierarchy
	return discoverSoftraidFromDmesg()
}

func parseBioctlOutput(out string) softraidInfo {
	sr := softraidInfo{
		virtualToPhysical: make(map[string]string),
		physicalToVirtual: make(map[string]string),
	}
	lines := strings.Split(out, "\n")
	var currentVirtual string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Volume line: "Volume sd0 (1.00 TB) RAID 0 CRYPTO"
		// May be prefixed: "softraid0: Volume sd0 ..."
		if strings.Contains(line, "Volume ") {
			idx := strings.Index(line, "Volume ")
			line = line[idx:]
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				currentVirtual = fields[1]
			}
			continue
		}

		// Chunk line: "423 MB chunk sd1a" (indented under volume)
		if currentVirtual != "" && strings.Contains(line, "chunk") {
			re := regexp.MustCompile(`chunk\s+([a-z]+[0-9]+)a?`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				physical := matches[1]
				sr.virtualToPhysical[currentVirtual] = physical
				sr.physicalToVirtual[physical] = currentVirtual
			}
			currentVirtual = "" // reset after we found the chunk
		}
	}
	return sr
}

func discoverSoftraidFromDmesg() softraidInfo {
	// We return empty maps — DiscoverDrives uses dmesg + disklabel
	// independently to set IsChunk/IsEncrypted flags without requiring
	// an exact virtual↔physical pairing (which can't be determined
	// reliably without bioctl).
	return softraidInfo{
		virtualToPhysical: make(map[string]string),
		physicalToVirtual: make(map[string]string),
	}
}

// isVirtualOnSoftraidBus checks dmesg to determine if device is a
// softraid virtual volume (attached to a scsibus owned by softraid).
func isVirtualOnSoftraidBus(device string) bool {
	dmesgOut, err := exec.Command("dmesg").Output()
	if err != nil {
		return false
	}

	// Find all scsibus devices owned by softraid controllers
	softraidBus := ""
	for _, line := range strings.Split(string(dmesgOut), "\n") {
		if strings.Contains(line, " at softraid") {
			// "scsibus3 at softraid0: 256 targets"
			fields := strings.Fields(line)
			if len(fields) > 0 {
				softraidBus = fields[0]
				break
			}
		}
	}
	if softraidBus == "" {
		return false
	}

	// Check if device is attached to that softraid bus
	// "sd1 at scsibus3 targ 1 lun 0: <OPENBSD, SR CRYPTO, 006>"
	for _, line := range strings.Split(string(dmesgOut), "\n") {
		prefix := device + " at " + softraidBus
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}

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
		if matched, _ := regexp.MatchString(`^  [a-z]:.*RAID`, line); matched {
			hasRAID = true
		}
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

// MountDrive mounts a drive. If encrypted, attaches softraid first.
func MountDrive(device, mountPoint string) error {
	if err := guardDrive(device); err != nil {
		return err
	}
	if mountPoint == "" {
		mountPoint = "/mnt/backup"
	}

	_ = exec.Command("doas", "mkdir", "-p", mountPoint).Run()

	if isMountedAt(mountPoint) {
		return fmt.Errorf("already mounted at %s", mountPoint)
	}

	raidDev := findRaidDevice(device)
	if raidDev == "" && hasRAIDPartition(device) {
		cmd := exec.Command("doas", "bioctl", "-c", "C", "-l", device+"a", "softraid0")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("bioctl attach failed: %w\n%s", err, string(out))
		}
		time.Sleep(500 * time.Millisecond)
		raidDev = findRaidDevice(device)
		if raidDev == "" {
			return fmt.Errorf("softraid attached but device not found")
		}
	}

	var source string
	if raidDev != "" {
		source = "/dev/" + raidDev + "a"
	} else {
		source = "/dev/" + device + "a"
	}

	cmd := exec.Command("doas", "mount", source, mountPoint)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mount failed: %w\n%s", err, string(out))
	}

	return nil
}

// UmountDrive unmounts and detaches softraid if present.
func UmountDrive(device, mountPoint string) error {
	if err := guardDrive(device); err != nil {
		return err
	}
	if mountPoint == "" {
		mountPoint = "/mnt/backup"
	}

	if isMountedAt(mountPoint) {
		cmd := exec.Command("doas", "umount", mountPoint)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("umount failed: %w\n%s", err, string(out))
		}
	}

	raidDev := findRaidDevice(device)
	if raidDev != "" {
		cmd := exec.Command("doas", "bioctl", "-d", raidDev)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("bioctl detach failed: %w\n%s", err, string(out))
		}
	}

	return nil
}

// guardDrive prevents operating on root drives or non-removable chunks.
func guardDrive(device string) error {
	drives, err := DiscoverDrives()
	if err != nil {
		return err
	}
	for _, d := range drives {
		if d.Device == device {
			if d.IsRoot {
				return fmt.Errorf("refusing to operate on root drive %s", device)
			}
			if d.IsChunk && !d.IsRemovable {
				return fmt.Errorf("refusing to operate on internal softraid drive %s", device)
			}
			return nil
		}
	}
	return nil
}

// FormatDrive formats a drive with 4.2BSD filesystem.
func FormatDrive(device string) error {
	if err := guardDrive(device); err != nil {
		return err
	}
	cmd := exec.Command("doas", "dd", "if=/dev/zero", "of=/dev/r"+device+"c", "bs=1m", "count=1")
	_ = cmd.Run()

	labelScript := fmt.Sprintf("a a\n\n\n4.2BSD\nw\nq\n")
	cmd = exec.Command("doas", "disklabel", "-E", device)
	cmd.Stdin = strings.NewReader(labelScript)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("disklabel failed: %w\n%s", err, string(out))
	}

	cmd = exec.Command("doas", "newfs", "/dev/r"+device+"a")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("newfs failed: %w\n%s", err, string(out))
	}

	return nil
}

// EncryptDrive sets up softraid crypto on a drive.
func EncryptDrive(device, passphrase string) error {
	if err := guardDrive(device); err != nil {
		return err
	}
	if passphrase == "" {
		return fmt.Errorf("passphrase required")
	}

	cmd := exec.Command("doas", "dd", "if=/dev/zero", "of=/dev/r"+device+"c", "bs=1m", "count=1")
	_ = cmd.Run()

	labelScript := fmt.Sprintf("a a\n\n\nRAID\nw\nq\n")
	cmd = exec.Command("doas", "disklabel", "-E", device)
	cmd.Stdin = strings.NewReader(labelScript)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("disklabel failed: %w\n%s", err, string(out))
	}

	cmd = exec.Command("doas", "bioctl", "-c", "C", "-l", device+"a", "softraid0")
	cmd.Stdin = strings.NewReader(passphrase + "\n" + passphrase + "\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("bioctl failed: %w\n%s", err, string(out))
	}

	time.Sleep(1 * time.Second)
	raidDev := findRaidDevice(device)
	if raidDev == "" {
		return fmt.Errorf("softraid attached but device not found")
	}

	cmd = exec.Command("doas", "dd", "if=/dev/zero", "of=/dev/r"+raidDev+"c", "bs=1m", "count=1")
	_ = cmd.Run()

	cmd = exec.Command("doas", "fdisk", "-iy", raidDev)
	_ = cmd.Run()

	labelScript = fmt.Sprintf("a a\n\n\n4.2BSD\nw\nq\n")
	cmd = exec.Command("doas", "disklabel", "-E", raidDev)
	cmd.Stdin = strings.NewReader(labelScript)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("disklabel on virtual device failed: %w\n%s", err, string(out))
	}

	cmd = exec.Command("doas", "newfs", "/dev/r"+raidDev+"a")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("newfs failed: %w\n%s", err, string(out))
	}

	return nil
}

// BenchmarkDrive runs fio tests on a mounted volume.
func BenchmarkDrive(mountPoint, writeSize, rwSize string) (string, error) {
	if mountPoint == "" {
		mountPoint = "/mnt/backup"
	}

	testFile := mountPoint + "/.disk_benchmark"

	cmd := exec.Command("fio",
		"--name=seqwrite", "--filename="+testFile,
		"--rw=write", "--bs=4m", "--size="+writeSize,
		"--numjobs=1", "--iodepth=1",
		"--group_reporting", "--thread",
	)
	out1, err1 := cmd.CombinedOutput()

	cmd = exec.Command("fio",
		"--name=seqread", "--filename="+testFile,
		"--rw=read", "--bs=4m", "--size="+writeSize,
		"--numjobs=1", "--iodepth=1",
		"--group_reporting", "--thread",
	)
	out2, err2 := cmd.CombinedOutput()

	cmd = exec.Command("fio",
		"--name=randrw", "--filename="+testFile,
		"--rw=randrw", "--bs=4k", "--size="+rwSize,
		"--numjobs=4", "--iodepth=32",
		"--group_reporting", "--thread",
	)
	out3, err3 := cmd.CombinedOutput()

	_ = os.Remove(testFile)

	var result strings.Builder
	result.WriteString("=== Sequential Write ===\n")
	if err1 != nil {
		result.WriteString(fmt.Sprintf("Error: %v\n%s\n", err1, string(out1)))
	} else {
		result.WriteString(extractFioResult(string(out1)))
	}
	result.WriteString("\n=== Sequential Read ===\n")
	if err2 != nil {
		result.WriteString(fmt.Sprintf("Error: %v\n%s\n", err2, string(out2)))
	} else {
		result.WriteString(extractFioResult(string(out2)))
	}
	result.WriteString("\n=== Random 4K ===\n")
	if err3 != nil {
		result.WriteString(fmt.Sprintf("Error: %v\n%s\n", err3, string(out3)))
	} else {
		result.WriteString(extractFioResult(string(out3)))
	}

	return result.String(), nil
}

func extractFioResult(output string) string {
	lines := strings.Split(output, "\n")
	var result strings.Builder
	for _, line := range lines {
		if strings.Contains(line, "read:") || strings.Contains(line, "write:") ||
			strings.Contains(line, "IOPS") || strings.Contains(line, "BW=") ||
			strings.Contains(line, "lat") {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	if result.Len() == 0 {
		return output
	}
	return result.String()
}

// getDeviceModel parses dmesg for the model name of device (e.g. sd2).
func getDeviceModel(device string) string {
	dmesgOut, err := exec.Command("dmesg").Output()
	if err != nil {
		return ""
	}
	prefix := device + " at "
	for _, line := range strings.Split(string(dmesgOut), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		idx := strings.Index(line, "<")
		if idx < 0 {
			return ""
		}
		end := strings.Index(line, ">")
		if end < 0 || end <= idx {
			return ""
		}
		fields := strings.Split(line[idx+1:end], ",")
		if len(fields) >= 2 {
			return strings.TrimSpace(fields[0]) + " " + strings.TrimSpace(fields[1])
		}
		if len(fields) >= 1 {
			return strings.TrimSpace(fields[0])
		}
	}
	return ""
}

// getDeviceBusType determines the bus type (NVMe, USB, CRYPTO, etc.)
// by checking what controller the device's scsibus is attached to.
func getDeviceBusType(device string) string {
	dmesgOut, err := exec.Command("dmesg").Output()
	if err != nil {
		return ""
	}
	dmesg := string(dmesgOut)

	// Find which scsibus this device is on
	deviceBus := ""
	for _, line := range strings.Split(dmesg, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, device+" at scsibus") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				deviceBus = fields[2] // e.g. "scsibus4"
			}
			break
		}
	}
	if deviceBus == "" {
		return ""
	}

	// Find what that scsibus is attached to
	for _, line := range strings.Split(dmesg, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, deviceBus+" at ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				bus := fields[2] // e.g. "umass0", "nvme0", "softraid0"
				switch {
				case strings.HasPrefix(bus, "umass"):
					return "USB"
				case strings.HasPrefix(bus, "nvme"):
					return "NVMe"
				case strings.HasPrefix(bus, "softraid"):
					return "CRYPTO"
				case strings.HasPrefix(bus, "mpath"):
					return "MPATH"
				default:
					return bus
				}
			}
		}
	}
	return ""
}

func isMountedAt(mountPoint string) bool {
	cmd := exec.Command("mount")
	out, _ := cmd.Output()
	return strings.Contains(string(out), " on "+mountPoint+" ")
}

func hasRAIDPartition(device string) bool {
	_, hasRAID, err := getDiskInfo(device)
	return err == nil && hasRAID
}

func findRaidDevice(device string) string {
	sr := parseSoftraid()
	// If device is a physical chunk, return its virtual device
	if v := sr.physicalToVirtual[device]; v != "" {
		return v
	}
	// If device is already a virtual device, return itself
	if sr.virtualToPhysical[device] != "" {
		return device
	}
	return ""
}

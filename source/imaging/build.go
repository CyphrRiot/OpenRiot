package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"openriot/fsutil"
	"openriot/logger"
)

// BuildImage creates the final installer image
func BuildImage(cfg *Config) error {
	// Must be root for all imaging operations
	if err := MustRunAsRoot(); err != nil {
		return err
	}

	logger.Info("Building final image...")

	// Clean up any leftover mounts first
	cleanupMounts()

	// Copy base image first (so we don't modify source)
	logger.Info("Copying base image...")
	if err := fsutil.CopyFile(cfg.BaseImg, cfg.OutputImg); err != nil {
		return fmt.Errorf("copy base: %w", err)
	}

	// Calculate target size: base image + tarball + 350MB buffer, round to 4MB
	baseInfo, err := os.Stat(cfg.BaseImg)
	if err != nil {
		return fmt.Errorf("stat base image: %w", err)
	}
	tgzInfo, err := os.Stat(cfg.OpenriotTgz)
	tgzSize := int64(0)
	if err == nil {
		tgzSize = tgzInfo.Size()
	}
	targetBytes := baseInfo.Size() + tgzSize + 350*1024*1024
	// Round up to nearest 4MB
	targetBytes = ((targetBytes + 4*1024*1024 - 1) / (4 * 1024 * 1024)) * (4 * 1024 * 1024)
	if targetBytes < 512*1024*1024 {
		targetBytes = 512 * 1024 * 1024 // minimum 512MB
	}

	// Expand image to calculated size
	if err := expandImage(cfg, targetBytes); err != nil {
		return fmt.Errorf("expand: %w", err)
	}

	// Inject content
	if err := injectContent(cfg); err != nil {
		return fmt.Errorf("inject: %w", err)
	}

	// Generate checksum
	if err := generateChecksum(cfg); err != nil {
		return fmt.Errorf("checksum: %w", err)
	}

	logger.Info(fmt.Sprintf("Image created: %s", cfg.OutputImg))
	return nil
}

// cleanupMounts releases any mounted filesystems and vnd devices
func cleanupMounts() {
	exec.Command("umount", "/mnt").Run()
	exec.Command("vnconfig", "-u", "vnd0").Run()
	time.Sleep(500 * time.Millisecond)
}

// expandImage resizes the image to targetBytes and expands the partition
func expandImage(cfg *Config, targetBytes int64) error {
	logger.Info(fmt.Sprintf("Expanding image to %dMB...", targetBytes/1024/1024))

	outputImg := cfg.OutputImg

	// Resize image to target
	cmd := exec.Command("truncate", "-s", strconv.FormatInt(targetBytes, 10), outputImg)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("truncate: %w\n%s", err, out)
	}

	// Release any existing vnd
	exec.Command("vnconfig", "-u", "vnd0").Run()

	// Configure vnd device
	cmd = exec.Command("vnconfig", "vnd0", outputImg)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("vnconfig: %w\n%s", err, out)
	}

	// Get current partition info
	rootStart, fstype, err := getPartitionInfo()
	if err != nil {
		return fmt.Errorf("get partition info: %w", err)
	}

	totalSec := getImageSize(outputImg)
	newSize := totalSec - rootStart

	logger.Info(fmt.Sprintf("total=%d start=%d new_size=%d", totalSec, rootStart, newSize))

	// Create new disklabel
	if err := writeDisklabel(rootStart, fstype, totalSec); err != nil {
		return fmt.Errorf("disklabel: %w", err)
	}

	// Grow filesystem
	cmd = exec.Command("growfs", "-y", "/dev/vnd0a")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("growfs: %w\n%s", err, out)
	}

	// Release vnd
	exec.Command("vnconfig", "-u", "vnd0").Run()

	logger.Info(fmt.Sprintf("Image expanded to %dMB", targetBytes/1024/1024))
	return nil
}

// getPartitionInfo returns root partition start sector and filesystem type
func getPartitionInfo() (int, string, error) {
	cmd := exec.Command("disklabel", "vnd0")
	out, err := cmd.Output()
	if err != nil {
		return 0, "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "  a:") {
			// Format: "  a:  1637376  1024  4.2BSD  "
			// fields: [0]=a: [1]=size [2]=offset(start) [3]=fstype
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				offset, _ := strconv.Atoi(fields[2]) // offset/start is in field 2
				fstype := fields[3]
				return offset, fstype, nil
			}
		}
	}
	return 0, "", fmt.Errorf("partition a: not found in disklabel")
}

// getImageSize returns image size in sectors (512 bytes each)
func getImageSize(path string) int {
	info, _ := os.Stat(path)
	return int(info.Size() / 512)
}

// getFileSize returns file size in bytes
func getFileSize(path string) int64 {
	info, _ := os.Stat(path)
	if info == nil {
		return 0
	}
	return info.Size()
}

// writeDisklabel reads the current disklabel, updates the a: and c: partition
// sizes to match the expanded image, and writes it back.
func writeDisklabel(rootStart int, fstype string, totalSec int) error {
	// Read current label
	cmd := exec.Command("disklabel", "vnd0")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("disklabel read: %w", err)
	}

	newSize := totalSec - rootStart
	lines := strings.Split(string(out), "\n")
	var modified []string
	foundA := false
	foundC := false

	for _, line := range lines {
		if strings.HasPrefix(line, "  a:") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				fields[1] = strconv.Itoa(newSize)
				fields[2] = strconv.Itoa(rootStart)
				fields[3] = fstype
				// Rebuild with original spacing, but only reference extra fields if present
				if len(fields) >= 7 {
					modified = append(modified, fmt.Sprintf("  a: %14s %14s  %-8s %s %s %s",
						fields[1], fields[2], fields[3], fields[4], fields[5], fields[6]))
				} else {
					modified = append(modified, fmt.Sprintf("  a: %14s %14s  %-8s",
						fields[1], fields[2], fields[3]))
				}
				foundA = true
				continue
			}
		}
		if strings.HasPrefix(line, "  c:") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				fields[1] = strconv.Itoa(totalSec)
				fields[2] = "0"
				// Rebuild with original spacing
				if len(fields) >= 7 {
					modified = append(modified, fmt.Sprintf("  c: %14s %14s  %-8s %s %s %s",
						fields[1], fields[2], fields[3], fields[4], fields[5], fields[6]))
				} else {
					modified = append(modified, fmt.Sprintf("  c: %14s %14s  %-8s",
						fields[1], fields[2], fields[3]))
				}
				foundC = true
				continue
			}
		}
		modified = append(modified, line)
	}

	if !foundA {
		return fmt.Errorf("partition a: not found in current disklabel")
	}
	if !foundC {
		logger.Warn("partition c: not found in current disklabel")
	}

	// Write modified label to temp file
	tmpPath := "/tmp/newlabel.txt"
	if err := os.WriteFile(tmpPath, []byte(strings.Join(modified, "\n")), 0644); err != nil {
		return err
	}

	cmd = exec.Command("disklabel", "-R", "vnd0", tmpPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("disklabel write: %w\n%s", err, out)
	}
	return nil
}

// injectContent mounts the image and copies tarball + packages
func injectContent(cfg *Config) error {
	logger.Info("Mounting image...")

	// Configure vnd
	cmd := exec.Command("vnconfig", "vnd0", cfg.OutputImg)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("vnconfig: %w\n%s", err, out)
	}

	// Run fsck
	logger.Info("Running fsck...")
	exec.Command("fsck", "-y", "/dev/vnd0a").Run()

	// Mount
	logger.Info("Mounting...")
	cmd = exec.Command("mount", "/dev/vnd0a", "/mnt")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mount: %w\n%s", err, out)
	}

	// Log available space BEFORE injection
	logger.Info("Space before injection:")
	cmd = exec.Command("df", "-h", "/mnt")
	if out, err := cmd.CombinedOutput(); err == nil {
		logger.Info(fmt.Sprintf("\n%s", strings.TrimSpace(string(out))))
	}

	// Inject tarball into the sets directory so the installer can find it
	setsDir := filepath.Join("/mnt", "7.9", "amd64")
	logger.Info(fmt.Sprintf("Creating sets directory %s...", setsDir))
	if err := os.MkdirAll(setsDir, 0755); err != nil {
		return fmt.Errorf("create sets dir: %w", err)
	}

	logger.Info("Injecting site79.tgz...")
	tgzSrc := cfg.OpenriotTgz
	tgzDst := filepath.Join(setsDir, "site79.tgz")
	if err := fsutil.CopyFile(tgzSrc, tgzDst); err != nil {
		return fmt.Errorf("copy tgz: %w", err)
	}

	// index.txt is required for the installer to discover site79.tgz.
	// Append to existing index.txt (base image already lists standard sets).
	logger.Info("Updating index.txt...")
	idxPath := filepath.Join(setsDir, "index.txt")
	idxLine := fmt.Sprintf("-rw-r--r--  1 root  wheel  %d %s site79.tgz\n",
		getFileSize(tgzSrc), time.Now().Format("Jan _2 15:04:05 2006"))
	f, err := os.OpenFile(idxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open index.txt: %w", err)
	}
	if _, err := f.WriteString(idxLine); err != nil {
		f.Close()
		return fmt.Errorf("append index.txt: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close index.txt: %w", err)
	}

	// Log available space AFTER injection
	logger.Info("Space after injection:")
	cmd = exec.Command("df", "-h", "/mnt")
	if out, err := cmd.CombinedOutput(); err == nil {
		logger.Info(fmt.Sprintf("\n%s", strings.TrimSpace(string(out))))
	}

	// Unmount
	exec.Command("umount", "/mnt").Run()
	exec.Command("vnconfig", "-u", "vnd0").Run()

	logger.Info("Content injected")
	return nil
}

// generateChecksum creates SHA256 checksum file
func generateChecksum(cfg *Config) error {
	shaPath := cfg.OutputImg + ".sha256"

	cmd := exec.Command("sha256", "-q", cfg.OutputImg)
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	hash := strings.TrimSpace(string(out))
	return os.WriteFile(shaPath, []byte(hash+"\n"), 0644)
}
package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"openriot/installer"
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
	if err := copyFile(cfg.BaseImg, cfg.OutputImg); err != nil {
		return fmt.Errorf("copy base: %w", err)
	}

	// Expand image from base
	if err := expandImage(cfg); err != nil {
		return fmt.Errorf("expand: %w", err)
	}

	// Inject content
	if err := injectContent(cfg); err != nil {
		return fmt.Errorf("inject: %w", err)
	}

	// Shrink to fit
	if err := shrinkImage(cfg); err != nil {
		return fmt.Errorf("shrink: %w", err)
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

// expandImage creates a 2GB image from base and expands partition
func expandImage(cfg *Config) error {
	logger.Info("Expanding image to 2GB...")

	outputImg := cfg.OutputImg

	// Create 2GB fixed image (truncate expands existing file)
	cmd := exec.Command("truncate", "-s", "2G", outputImg)
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

	logger.Info("Image expanded to 2GB")
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
	return 0, "4.2BSD", nil // Default
}

// getImageSize returns image size in sectors (512 bytes each)
func getImageSize(path string) int {
	info, _ := os.Stat(path)
	return int(info.Size() / 512)
}

// writeDisklabel updates the disklabel to fill the expanded image
func writeDisklabel(rootStart int, fstype string, totalSec int) error {
	label := fmt.Sprintf(`# /dev/rvnd0c:
type: vnd
disk: vnd device
label: fictitious
duid: 2c656feb7ddb57e2
flags:
bytes/sector: 512
sectors/track: 100
tracks/cylinder: 1
sectors/cylinder: 100
cylinders: 31458
total sectors: %d

16 partitions:
#                size           offset  fstype [fsize bsize   cpg]
  a:          %d             %d  %s   2048 16384 16142
  c:          %d                0  unused
  i:              960               64   MSDOS
`, totalSec, totalSec-rootStart, rootStart, fstype, totalSec)

	// Write to temp file
	tmpPath := "/tmp/newlabel.txt"
	if err := os.WriteFile(tmpPath, []byte(label), 0644); err != nil {
		return err
	}

	cmd := exec.Command("disklabel", "-R", "vnd0", tmpPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w\n%s", err, out)
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

	// Inject tarball
	logger.Info("Injecting openriot.tgz...")
	tgzSrc := cfg.OpenriotTgz
	tgzDst := "/mnt/openriot.tgz"
	if err := copyFile(tgzSrc, tgzDst); err != nil {
		return fmt.Errorf("copy tgz: %w", err)
	}

	// Inject packages
	logger.Info("Injecting packages...")
	pkgSrcDir := filepath.Join(cfg.WorkDir, "packages", "snapshots", "amd64")
	pkgDstDir := "/mnt/openriot/packages/snapshots/amd64"
	os.MkdirAll(pkgDstDir, 0755)

	entries, _ := os.ReadDir(pkgSrcDir)
	pkgCount := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tgz") {
			pkgCount++
		}
	}

	pkgDone := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tgz") {
			pkgDone++
			name := strings.TrimSuffix(entry.Name(), ".tgz")
			// Pad to 35 chars to clear any previous longer name
			namePadded := fmt.Sprintf("%-35s", name)
			fmt.Printf("\r%s[INFO]%s Injecting package %d/%d: %s", installer.Cyan, installer.Reset, pkgDone, pkgCount, namePadded)

			src := filepath.Join(pkgSrcDir, entry.Name())
			dst := filepath.Join(pkgDstDir, entry.Name())
			if err := copyFile(src, dst); err != nil {
				fmt.Println()
				logger.Warn(fmt.Sprintf("Failed to copy %s: %v", entry.Name(), err))
			}
		}
	}

	fmt.Println() // Ensure clean line before next output

	// Unmount
	exec.Command("umount", "/mnt").Run()
	exec.Command("vnconfig", "-u", "vnd0").Run()

	logger.Info("Content injected")
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// shrinkImage reduces image to minimum size + buffer
func shrinkImage(cfg *Config) error {
	logger.Info("Shrinking image to fit content...")

	// Configure vnd
	cmd := exec.Command("vnconfig", "vnd0", cfg.OutputImg)
	cmd.Run()

	// Get used space
	dfCmd := exec.Command("df", "-k", "/dev/vnd0a")
	out, _ := dfCmd.Output()
	lines := strings.Split(string(out), "\n")
	var usedKB int
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] != "1K-blocks" {
			usedKB, _ = strconv.Atoi(fields[2])
			break
		}
	}

	// Calculate needed size: used + 10% buffer + 32MB
	neededKB := usedKB*110/100 + 32768
	neededMB := (neededKB + 4095) / 4096 * 4
	if neededMB < 1024 {
		neededMB = 1024
	}

	logger.Info(fmt.Sprintf("Shrinking to %dMB (used: %dKB)...", neededMB, usedKB))

	// Release vnd
	exec.Command("vnconfig", "-u", "vnd0").Run()

	// Truncate
	cmd = exec.Command("truncate", "-s", fmt.Sprintf("%dM", neededMB), cfg.OutputImg)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("truncate: %w\n%s", err, out)
	}

	logger.Info(fmt.Sprintf("Image shrunk to %dMB", neededMB))
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
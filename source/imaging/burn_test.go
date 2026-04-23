package imaging

import (
	"testing"
)

func TestDetectDrives(t *testing.T) {
	drives, err := DetectDrives()
	if err != nil {
		t.Fatalf("DetectDrives failed: %v", err)
	}

	t.Logf("Detected %d drives", len(drives))
	for _, d := range drives {
		t.Logf("  %s: %dGB, removable=%v, protected=%v, status=%s",
			d.Device, d.SizeGB, d.IsRemovable, d.IsProtected, d.Status)
	}
}

func TestGetDiskInfo(t *testing.T) {
	// Try to get info for a disk that might exist
	// This will likely fail in test environment but should handle gracefully
	drives, _ := DetectDrives()
	if len(drives) == 0 {
		t.Skip("No drives detected, skipping disk info test")
	}

	size, protected, err := getDiskInfo(drives[0].Device)
	if err != nil {
		t.Logf("getDiskInfo error (expected in some cases): %v", err)
	} else {
		t.Logf("First drive: %dGB, protected=%v", size, protected)
	}
}
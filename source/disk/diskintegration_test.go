package disk_test

import (
	"testing"

	"openriot/disk"
)

// TestSystemDriveDetection verifies that DiscoverDrives correctly identifies
// sd0 (physical NVMe, softraid chunk) and sd1 (virtual SR CRYPTO, root) on
// this system. This is an integration test that requires the actual hardware
// state — it will report different results on other machines.
func TestSystemDriveDetection(t *testing.T) {
	drives, err := disk.DiscoverDrives()
	if err != nil {
		t.Fatalf("DiscoverDrives failed: %v", err)
	}

	t.Logf("Discovered %d drive(s):", len(drives))
	for _, d := range drives {
		t.Logf("  %s  %d GB  root=%v mounted=%v chunk=%v encrypted=%v raid=%v bus=%s model=%s",
			d.Device, d.SizeGB, d.IsRoot, d.IsMounted, d.IsChunk,
			d.IsEncrypted, d.HasRAID, d.BusType, d.ModelName)
	}

	// Find sd0 and sd1
	var sd0, sd1 *disk.Drive
	for i := range drives {
		switch drives[i].Device {
		case "sd0":
			sd0 = &drives[i]
		case "sd1":
			sd1 = &drives[i]
		}
	}

	if sd0 == nil {
		t.Fatal("sd0 not found — is the NVMe drive present?")
	}
	if sd1 == nil {
		t.Fatal("sd1 not found — is the softraid volume active?")
	}

	// sd0 is the physical NVMe with a RAID partition (softraid chunk)
	if !sd0.HasRAID {
		t.Error("sd0.HasRAID is false — expected true (RAID partition)")
	}
	if !sd0.IsChunk {
		t.Error("sd0.IsChunk is false — expected true (softraid chunk)")
	}

	// sd1 is the virtual softraid crypto volume (root)
	if !sd1.IsEncrypted {
		t.Error("sd1.IsEncrypted is false — expected true (SR CRYPTO)")
	}
	if !sd1.IsRoot {
		t.Error("sd1.IsRoot is false — expected true (root device)")
	}
	if !sd1.IsMounted {
		t.Error("sd1.IsMounted is false — expected true (mounted as /)")
	}

	// Verify filterDrives includes ALL drives (selection guards in
	// updateDriveList block operations on ineligible drives at runtime).
	for _, action := range []disk.ActionType{
		disk.ActionMount, disk.ActionUmount, disk.ActionFormat,
		disk.ActionEncrypt, disk.ActionBenchmark, disk.ActionDiscover,
	} {
		filtered := disk.FilterDrives(drives, action)
		foundSD0, foundSD1 := false, false
		for _, d := range filtered {
			if d.Device == "sd0" {
				foundSD0 = true
			}
			if d.Device == "sd1" {
				foundSD1 = true
			}
		}
		if !foundSD0 {
			t.Errorf("sd0 missing from filter for action %d — should be shown", action)
		}
		if !foundSD1 {
			t.Errorf("sd1 missing from filter for action %d — should be shown", action)
		}
	}

	// Verify Discover filter includes everything (same as above, explicit check)
	discoverFiltered := disk.FilterDrives(drives, disk.ActionDiscover)
	if len(discoverFiltered) != len(drives) {
		t.Errorf("Discover returned %d drives, expected %d", len(discoverFiltered), len(drives))
	}
}

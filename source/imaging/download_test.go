package imaging

import (
	"testing"
)

func TestLoadExceptions(t *testing.T) {
	exceptions, err := LoadExceptions()
	if err != nil {
		t.Fatalf("LoadExceptions failed: %v", err)
	}

	t.Logf("Exceptions loaded: %d packages", len(exceptions))
	for pkg := range exceptions {
		t.Logf("  - %s", pkg)
	}
}

func TestGetPackageList(t *testing.T) {
	packages, err := GetPackageList()
	if err != nil {
		t.Fatalf("GetPackageList failed: %v", err)
	}

	if len(packages) == 0 {
		t.Fatal("expected packages, got none")
	}

	t.Logf("Package count: %d", len(packages))
	// Show first 10
	for i, pkg := range packages {
		if i >= 10 {
			break
		}
		t.Logf("  %d: %s", i+1, pkg)
	}
}
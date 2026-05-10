package imaging

import (
	"testing"

	"openriot/config"
)

func TestLoadExceptions(t *testing.T) {
	exceptions, err := LoadExceptions()
	if err != nil {
		t.Fatalf("LoadExceptions failed: %v", err)
	}

	t.Logf("Package exceptions: %d", len(exceptions.Packages))
	for pkg := range exceptions.Packages {
		t.Logf("  pkg: %s", pkg)
	}

	t.Logf("Module exceptions: %d", len(exceptions.Modules))
	for mod := range exceptions.Modules {
		t.Logf("  mod: %s", mod)
	}
}

func TestGetPackagesExcluding(t *testing.T) {
	cfgPath := config.FindConfigFile()
	if cfgPath == "" {
		t.Skip("no packages.yaml found")
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	all := cfg.GetPackages()
	filtered := cfg.GetPackagesExcluding(map[string]bool{"go": true}, map[string]bool{"desktop.games": true})

	if len(filtered) >= len(all) {
		t.Errorf("expected fewer packages after exclusion, got %d >= %d", len(filtered), len(all))
	}

	// Verify go is excluded
	for _, pkg := range filtered {
		if config.GetBaseName(pkg) == "go" {
			t.Errorf("go should be excluded but found: %s", pkg)
		}
	}

	// Verify desktop.games packages are excluded
	for _, pkg := range filtered {
		base := config.GetBaseName(pkg)
		for _, game := range []string{"gottet", "defendguin", "supertux", "sdlpop", "zelda_roth_se", "endless-sky", "openttd", "cataclysm-dda", "taisei", "angband", "wesnoth", "boswars", "slash", "frozen-bubble"} {
			if base == game {
				t.Errorf("game %s should be excluded but found: %s", game, pkg)
			}
		}
	}

	t.Logf("All: %d, Filtered: %d", len(all), len(filtered))
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
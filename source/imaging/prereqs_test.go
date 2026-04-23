package imaging

import (
	"testing"
)

func TestIsOpenBSD(t *testing.T) {
	result := IsOpenBSD()
	t.Logf("IsOpenBSD: %v", result)
	// On OpenBSD this should be true
}

func TestIsRoot(t *testing.T) {
	result := IsRoot()
	t.Logf("IsRoot: %v", result)
	// Should be false for non-root user
}

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig([]string{})
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	if cfg.BaseImg != "Build/Images/install79.img" {
		t.Errorf("expected default BaseImg, got %s", cfg.BaseImg)
	}
	if cfg.OutputImg != "Build/Images/openriot.img" {
		t.Errorf("expected default OutputImg, got %s", cfg.OutputImg)
	}
	if cfg.WorkDir == "" {
		t.Errorf("expected default WorkDir, got empty")
	}
	if cfg.Version != "79" {
		t.Errorf("expected default Version 79, got %s", cfg.Version)
	}
	if cfg.NoBurn != false {
		t.Errorf("expected NoBurn false, got %v", cfg.NoBurn)
	}
	
	t.Logf("BaseImg: %s", cfg.BaseImg)
	t.Logf("OutputImg: %s", cfg.OutputImg)
	t.Logf("WorkDir: %s", cfg.WorkDir)
	t.Logf("OpenriotBin: %s", cfg.OpenriotBin)
	t.Logf("OpenriotTgz: %s", cfg.OpenriotTgz)
}

func TestLoadConfigWithFlags(t *testing.T) {
	args := []string{
		"--base-img", "/custom/base.img",
		"--output-img", "/custom/output.img",
		"--work-dir", "/custom/work",
		"--version", "80",
		"--no-burn",
	}
	
	cfg, err := LoadConfig(args)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	if cfg.BaseImg != "/custom/base.img" {
		t.Errorf("expected /custom/base.img, got %s", cfg.BaseImg)
	}
	if cfg.OutputImg != "/custom/output.img" {
		t.Errorf("expected /custom/output.img, got %s", cfg.OutputImg)
	}
	if cfg.WorkDir != "/custom/work" {
		t.Errorf("expected /custom/work, got %s", cfg.WorkDir)
	}
	if cfg.Version != "80" {
		t.Errorf("expected version 80, got %s", cfg.Version)
	}
	if !cfg.NoBurn {
		t.Errorf("expected NoBurn true, got %v", cfg.NoBurn)
	}
}

func TestCheckPrereqs(t *testing.T) {
	cfg := &Config{
		BaseImg: "/nonexistent/image.img",
	}
	
	err := CheckPrereqs(cfg)
	if err == nil {
		t.Error("expected error for missing base image")
	} else {
		t.Logf("CheckPrereqs error (expected): %v", err)
	}
}

func TestMustRunAsRoot(t *testing.T) {
	err := MustRunAsRoot()
	if err == nil {
		t.Error("expected error for non-root")
	} else {
		t.Logf("MustRunAsRoot error (expected): %v", err)
	}
}
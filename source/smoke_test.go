package main

import (
	"os"
	"path/filepath"
	"testing"

	"openriot/fonts"
	"openriot/installer"
	"openriot/config"
)

func TestSmoke_FontInstallPath(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	srcDir := filepath.Join(home, ".local", "share", "openriot", "assets", "fonts")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "TestFontNerdFont-Regular.otf"), []byte("fontdata"), 0644)

	if err := fonts.Run(); err != nil {
		t.Fatalf("fonts.Run() failed: %v", err)
	}

	dst := filepath.Join(home, ".local", "share", "fonts", "TestFontNerdFont-Regular.otf")
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("expected font at %s: %v", dst, err)
	}
}

func TestSmoke_ConfigInstallPath(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	deployDir := t.TempDir()
	srcDir := filepath.Join(deployDir, "config", "testapp")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "config.conf"), []byte("testconfig"), 0644)

	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Configs: []config.ConfigRule{
					{Pattern: "testapp/config.conf"},
				},
			},
		},
	}

	if err := installer.CopyConfigs(deployDir, cfg, false); err != nil {
		t.Fatalf("CopyConfigs failed: %v", err)
	}

	dst := filepath.Join(home, ".config", "testapp", "config.conf")
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("expected config at %s: %v", dst, err)
	}
	if string(data) != "testconfig" {
		t.Fatalf("expected 'testconfig', got %q", string(data))
	}
}

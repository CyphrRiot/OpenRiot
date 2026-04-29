package fonts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_HappyPath(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	srcDir := filepath.Join(home, ".local", "share", "openriot", "assets", "fonts")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "test.ttf"), []byte("fontdata"), 0644)

	err := Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dst := filepath.Join(home, ".local", "share", "fonts", "test.ttf")
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("expected copied font at %s: %v", dst, err)
	}
	if string(data) != "fontdata" {
		t.Fatalf("expected 'fontdata', got %q", string(data))
	}
}

func TestRun_MissingSourceDir(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	err := Run()
	if err == nil {
		t.Fatal("expected error for missing source dir")
	}
}

func TestRun_EmptySourceDir(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	srcDir := filepath.Join(home, ".local", "share", "openriot", "assets", "fonts")
	os.MkdirAll(srcDir, 0755)

	err := Run()
	if err == nil {
		t.Fatal("expected error for empty source dir")
	}
}

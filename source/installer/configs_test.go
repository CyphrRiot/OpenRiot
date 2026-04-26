package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openriot/config"
)

func TestCopyConfigs_SingleFile(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Create source file
	srcDir := filepath.Join(repo, "config", "testapp")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "config.conf"), []byte("hello"), 0644)

	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Configs: []config.ConfigRule{
					{Pattern: "testapp/config.conf"},
				},
			},
		},
	}

	err := CopyConfigs(repo, cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dest := filepath.Join(home, ".config", "testapp", "config.conf")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("expected file at %s: %v", dest, err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(data))
	}
}

func TestCopyConfigs_GlobPattern(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	srcDir := filepath.Join(repo, "config", "fastfetch")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "a.json"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(srcDir, "b.json"), []byte("b"), 0644)

	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Configs: []config.ConfigRule{
					{Pattern: "fastfetch/*", Target: "~/.config/fastfetch"},
				},
			},
		},
	}

	err := CopyConfigs(repo, cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, name := range []string{"a.json", "b.json"} {
		dest := filepath.Join(home, ".config", "fastfetch", name)
		if _, err := os.Stat(dest); err != nil {
			t.Fatalf("expected %s to exist: %v", dest, err)
		}
	}
}

func TestCopyConfigs_PreserveIfExists(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	srcDir := filepath.Join(repo, "config", "fish")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "custom_commands.fish"), []byte("new"), 0644)

	// Pre-create destination file
	destDir := filepath.Join(home, ".config", "fish")
	os.MkdirAll(destDir, 0755)
	os.WriteFile(filepath.Join(destDir, "custom_commands.fish"), []byte("old"), 0644)

	cfg := &config.Config{
		Core: map[string]config.Module{
			"shell": {
				Configs: []config.ConfigRule{
					{
						Pattern:          "fish/*",
						PreserveIfExists: []string{"custom_commands.fish"},
					},
				},
			},
		},
	}

	err := CopyConfigs(repo, cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(destDir, "custom_commands.fish"))
	if string(data) != "old" {
		t.Fatalf("expected preserved 'old', got %q", string(data))
	}
}

func TestCopyConfigs_SkipIdenticalContent(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	srcDir := filepath.Join(repo, "config", "testapp")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "same.conf"), []byte("identical"), 0644)

	destDir := filepath.Join(home, ".config", "testapp")
	os.MkdirAll(destDir, 0755)
	os.WriteFile(filepath.Join(destDir, "same.conf"), []byte("identical"), 0644)

	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Configs: []config.ConfigRule{
					{Pattern: "testapp/same.conf"},
				},
			},
		},
	}

	// Verify file exists before
	before, _ := os.Stat(filepath.Join(destDir, "same.conf"))
	beforeTime := before.ModTime()

	err := CopyConfigs(repo, cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was NOT rewritten (modtime unchanged)
	after, _ := os.Stat(filepath.Join(destDir, "same.conf"))
	if !after.ModTime().Equal(beforeTime) {
		t.Fatal("expected file to be skipped (identical content), but modtime changed")
	}
}

func TestCopyConfigs_DryRun(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	srcDir := filepath.Join(repo, "config", "testapp")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "config.conf"), []byte("hello"), 0644)

	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Configs: []config.ConfigRule{
					{Pattern: "testapp/config.conf"},
				},
			},
		},
	}

	err := CopyConfigs(repo, cfg, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dest := filepath.Join(home, ".config", "testapp", "config.conf")
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("expected dry-run to NOT create files")
	}
}

func TestCopyConfigs_TargetOverride(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	srcDir := filepath.Join(repo, "config", "testapp")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "config.conf"), []byte("hello"), 0644)

	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Configs: []config.ConfigRule{
					{Pattern: "testapp/config.conf", Target: "~/custom/target"},
				},
			},
		},
	}

	err := CopyConfigs(repo, cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dest := filepath.Join(home, "custom", "target")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("expected file at %s: %v", dest, err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(data))
	}
}

func TestCopyConfigs_DependencyOrder(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Create source files for two modules
	srcBase := filepath.Join(repo, "config", "base")
	os.MkdirAll(srcBase, 0755)
	os.WriteFile(filepath.Join(srcBase, "a.conf"), []byte("a"), 0644)

	srcShell := filepath.Join(repo, "config", "shell")
	os.MkdirAll(srcShell, 0755)
	os.WriteFile(filepath.Join(srcShell, "b.conf"), []byte("b"), 0644)

	cfg := &config.Config{
		Core: map[string]config.Module{
			"shell": {
				Configs: []config.ConfigRule{{Pattern: "shell/*"}},
				Depends:   []string{"core.base"},
			},
			"base": {
				Configs: []config.ConfigRule{{Pattern: "base/*"}},
			},
		},
	}

	err := CopyConfigs(repo, cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both files should exist regardless of order
	if _, err := os.Stat(filepath.Join(home, ".config", "base", "a.conf")); err != nil {
		t.Fatalf("expected base config: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "shell", "b.conf")); err != nil {
		t.Fatalf("expected shell config: %v", err)
	}
}

func TestCopyConfigs_MissingDependencyError(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	cfg := &config.Config{
		Core: map[string]config.Module{
			"shell": {
				Configs: []config.ConfigRule{},
				Depends:   []string{"core.missing"},
			},
		},
	}

	err := CopyConfigs(repo, cfg, false)
	if err == nil {
		t.Fatal("expected error for missing dependency")
	}
	if !strings.Contains(err.Error(), "resolving module dependencies") {
		t.Fatalf("expected dependency resolution error, got: %v", err)
	}
}

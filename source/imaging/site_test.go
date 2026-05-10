package imaging

import (
	"os"
	"strings"
	"testing"
)

func TestGetBuildDir(t *testing.T) {
	dir := getBuildDir()
	t.Logf("Build dir: %s", dir)
}

func TestCreateInstallSite(t *testing.T) {
	tmpDir := t.TempDir()
	
	err := createInstallSite(tmpDir)
	if err != nil {
		t.Fatalf("createInstallSite failed: %v", err)
	}
	
	path := tmpDir + "/install.site"
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("install.site not created: %v", err)
	}
	
	data, _ := os.ReadFile(path)
	content := string(data)
	
	if !strings.Contains(content, "STEP 1: Configure doas") {
		t.Error("missing STEP 1")
	}
	if !strings.Contains(content, "Configuring doas") {
		t.Error("missing doas config")
	}
	if !strings.Contains(content, "pkg_add") {
		t.Error("missing pkg_add")
	}
	if !strings.Contains(content, "curl -fsSL https://OpenRiot.org/sh") {
		t.Error("missing curl command")
	}
	
	t.Logf("install.site created successfully (%d bytes)", len(content))
}

func TestCreateInstallConf(t *testing.T) {
	tmpDir := t.TempDir()
	
	err := createInstallConf(tmpDir)
	if err != nil {
		t.Fatalf("createInstallConf failed: %v", err)
	}
	
	path := tmpDir + "/install.conf"
	data, _ := os.ReadFile(path)
	content := string(data)
	
	if !strings.Contains(content, "Which disk is the root disk = ask") {
		t.Error("missing disk prompt")
	}
	if !strings.Contains(content, "install openriot.tgz = yes") {
		t.Error("missing tgz install")
	}
	
	t.Logf("install.conf created successfully (%d bytes)", len(content))
}
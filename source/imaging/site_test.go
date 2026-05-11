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

	if !strings.Contains(content, "permit nopass :wheel") {
		t.Error("missing doas config")
	}
	if !strings.Contains(content, "pkg_add") {
		t.Error("missing pkg_add")
	}
	if !strings.Contains(content, "curl -fsSL https://OpenRiot.org/setup.sh") {
		t.Error("missing curl setup.sh welcome message")
	}
	if strings.Contains(content, "rcctl enable xenodm") {
		t.Error("xenodm should not be enabled")
	}
	if strings.Contains(content, "openriot --install") {
		t.Error("openriot --install should not be in first-login script")
	}

	t.Logf("install.site created successfully (%d bytes)", len(content))
}
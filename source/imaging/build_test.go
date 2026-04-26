package imaging

import (
	"os"
	"testing"

	"openriot/fsutil"
)

func TestGetPartitionInfo(t *testing.T) {
	// This will fail without root/vnd configured, which is expected
	// Just verify the function exists and can be called
	t.Skip("Requires root and vnd device - manual test only")
}

func TestGetImageSize(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.img"

	// Create a 1MB test file
	f, _ := os.Create(testFile)
	f.WriteAt(make([]byte, 1024*1024), 0)
	f.Close()

	size := getImageSize(testFile)
	expected := 1024 * 1024 / 512 // 2048 sectors

	if size != expected {
		t.Errorf("expected %d sectors, got %d", expected, size)
	}

	t.Logf("getImageSize test: %d sectors for 1MB file", size)
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := tmpDir + "/source.txt"
	dst := tmpDir + "/dest.txt"

	// Write test content
	content := []byte("hello world test content")
	os.WriteFile(src, content, 0644)

	// Copy
	if err := fsutil.CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify
	data, _ := os.ReadFile(dst)
	if string(data) != string(content) {
		t.Errorf("content mismatch")
	}

	t.Logf("CopyFile works correctly")
}

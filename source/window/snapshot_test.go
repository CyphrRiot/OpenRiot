package window

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestSnapshotJSONRoundTrip(t *testing.T) {
	snap := Snapshot{
		Timestamp: time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
		Windows: []WindowInfo{
			{
				Class:     "Firefox",
				Instance:  "Navigator",
				Workspace: 1,
				Rect:      windowRect{X: 0, Y: 0, Width: 1920, Height: 1080},
				Floating:  false,
				Focused:   true,
			},
			{
				Class:     "Alacritty",
				Instance:  "alacritty",
				Workspace: 2,
				Rect:      windowRect{X: 100, Y: 200, Width: 800, Height: 600},
				Floating:  true,
				Focused:   false,
			},
		},
	}

	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var restored Snapshot
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(restored.Windows) != len(snap.Windows) {
		t.Fatalf("window count: got %d, want %d", len(restored.Windows), len(snap.Windows))
	}
	for i := range snap.Windows {
		if !reflect.DeepEqual(snap.Windows[i], restored.Windows[i]) {
			t.Errorf("window[%d] mismatch:\n  got  %+v\n  want %+v", i, restored.Windows[i], snap.Windows[i])
		}
	}
}

func TestSnapshotFileWriteRead(t *testing.T) {
	snap := Snapshot{
		Timestamp: time.Now(),
		Windows: []WindowInfo{
			{Class: "testapp", Workspace: 1},
		},
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "window-snapshot.json")

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	readData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var restored Snapshot
	if err := json.Unmarshal(readData, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(restored.Windows) != 1 || restored.Windows[0].Class != "testapp" {
		t.Errorf("unexpected content: %+v", restored)
	}
}

func TestSkipClassesFilter(t *testing.T) {
	for cls := range shutdownSkipClasses {
		if !shutdownSkipClasses[cls] {
			t.Errorf("expected %q in skip set", cls)
		}
	}
	nonSkipped := []string{"firefox", "alacritty", "thunar", "mpv"}
	for _, cls := range nonSkipped {
		if shutdownSkipClasses[cls] {
			t.Errorf("unexpected %q in skip set", cls)
		}
	}
}

func TestLoadRestoreCommands(t *testing.T) {
	cmds := loadRestoreCommands()
	if len(cmds) == 0 {
		t.Log("no commands loaded (apps.txt may not exist in test env)")
		return
	}
	for k, v := range cmds {
		if k == "" {
			t.Error("empty key in command map")
		}
		if v == "" {
			t.Errorf("empty command for key %q", k)
		}
	}
}

func TestWindowRectZeroValues(t *testing.T) {
	r := windowRect{}
	if r.X != 0 || r.Y != 0 || r.Width != 0 || r.Height != 0 {
		t.Error("zero-value windowRect should have all fields zero")
	}
}

func TestSnapshotEmptyWindows(t *testing.T) {
	snap := Snapshot{
		Timestamp: time.Now(),
		Windows:   []WindowInfo{},
	}
	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var restored Snapshot
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(restored.Windows) != 0 {
		t.Errorf("expected 0 windows, got %d", len(restored.Windows))
	}
}
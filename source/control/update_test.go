package control

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel() model {
	return NewModel().(model)
}

func TestItemsInSection(t *testing.T) {
	cases := []struct {
		name      string
		section   int
		bat       int
		proton    bool
		recording bool
		want      int
	}{
		{"audio", 0, 0, false, false, 2},
		{"display", 1, 0, false, false, 3},
		{"network", 2, 0, false, false, 3},
		{"power", 3, 80, false, false, 6},
		{"power-no-battery", 3, 255, false, false, 1},
		{"privacy", 4, 0, false, false, 2},
		{"privacy-proton", 4, 0, true, false, 3},
		{"system", 5, 0, false, false, 2},
		{"system-recording", 5, 0, false, true, 3},
		{"health", 6, 0, false, false, 1},
	}
	for _, c := range cases {
		m := newTestModel()
		m.section = c.section
		m.batPercent = c.bat
		m.proton = c.proton
		m.recording = c.recording
		if got := m.itemsInSection(); got != c.want {
			t.Errorf("%s: itemsInSection() = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestCursorWrap(t *testing.T) {
	m := newTestModel()
	m.section = 0 // 2 items

	down, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m = down.(model)
	if m.cursor != 1 {
		t.Fatalf("down: cursor = %d, want 1", m.cursor)
	}

	down, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m = down.(model)
	if m.cursor != 0 {
		t.Fatalf("down past end: cursor = %d, want 0", m.cursor)
	}

	up, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	m = up.(model)
	if m.cursor != 1 {
		t.Fatalf("up from start: cursor = %d, want 1", m.cursor)
	}
}

func TestSectionSwitch(t *testing.T) {
	m := newTestModel()
	m.section = 0
	m.cursor = 1

	right, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRight})
	m = right.(model)
	if m.section != 1 || m.cursor != 0 {
		t.Fatalf("right: section = %d cursor = %d, want 1/0", m.section, m.cursor)
	}

	left, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	m = left.(model)
	if m.section != 0 {
		t.Fatalf("left: section = %d, want 0", m.section)
	}

	left, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	m = left.(model)
	if m.section != len(m.sections)-1 {
		t.Fatalf("left wrap: section = %d, want %d", m.section, len(m.sections)-1)
	}
}

func TestClampCursor(t *testing.T) {
	m := newTestModel()
	m.section = 3
	m.batPercent = 255
	m.cursor = 5
	m = m.clampCursor()
	if m.cursor != 0 {
		t.Fatalf("clamp power-no-battery: cursor = %d, want 0", m.cursor)
	}

	m = newTestModel()
	m.section = 4
	m.proton = false
	m.cursor = 2
	m = m.clampCursor()
	if m.cursor != 1 {
		t.Fatalf("clamp privacy: cursor = %d, want 1", m.cursor)
	}
}

func TestStatusLabel(t *testing.T) {
	cases := map[status]string{
		statusOK:   "OK",
		statusWarn: "WARN",
		statusFail: "FAIL",
	}
	for s, want := range cases {
		if got := statusLabel(s); got != want {
			t.Errorf("statusLabel(%d) = %q, want %q", s, got, want)
		}
	}
}

func TestDiskStatus(t *testing.T) {
	cases := map[int64]status{
		0:   statusFail,
		1:   statusWarn,
		4:   statusWarn,
		5:   statusOK,
		150: statusOK,
	}
	for free, want := range cases {
		if got := diskStatus(free); got != want {
			t.Errorf("diskStatus(%d GB) = %v, want %v", free, got, want)
		}
	}
}

func TestPfEnabled(t *testing.T) {
	cases := map[string]bool{
		"pf=YES": true,
		"pf=No":  false,
		"":       false,
		"pf=yes": true,
	}
	for out, want := range cases {
		if got := pfEnabled(out); got != want {
			t.Errorf("pfEnabled(%q) = %v, want %v", out, got, want)
		}
	}
}

func TestDoasConfOK(t *testing.T) {
	cases := []struct {
		perm, owner string
		want        bool
	}{
		{"-rw-r-----", "root", true},
		{"-rw-r--r--", "root", true},
		{"-rw-rw-rw-", "root", false},
		{"-rw-rw----", "root", false},
		{"-rw-r-----", "grendel", false},
		{"", "root", false},
	}
	for _, c := range cases {
		if got := doasConfOK(c.perm, c.owner); got != c.want {
			t.Errorf("doasConfOK(%q,%q) = %v, want %v", c.perm, c.owner, got, c.want)
		}
	}
}

func TestParseSwap(t *testing.T) {
	total, used, ok := parseSwap("total: 16809354 512-blocks allocated, 0 used, 16809354 available")
	if !ok || total != 16809354 || used != 0 {
		t.Fatalf("parseSwap = (%d,%d,%v), want (16809354,0,true)", total, used, ok)
	}

	if _, _, ok := parseSwap(""); ok {
		t.Fatal("parseSwap(\"\") should not be ok")
	}
}

func TestSwapStatus(t *testing.T) {
	cases := map[int]status{
		0:  statusOK,
		24: statusOK,
		25: statusWarn,
		74: statusWarn,
		75: statusFail,
		99: statusFail,
	}
	for pct, want := range cases {
		if got := swapStatus(pct); got != want {
			t.Errorf("swapStatus(%d%%) = %v, want %v", pct, got, want)
		}
	}
}

func TestPackageDrift(t *testing.T) {
	yaml := []string{"firefox-140.0", "vim-9.0", "zsh-5.9"}
	installed := map[string]string{
		"firefox": "firefox-140.0", // in sync
		"vim":     "vim-8.0",       // outdated
		// zsh missing
	}
	missing, mismatch := packageDrift(yaml, installed)
	if missing != 1 || mismatch != 1 {
		t.Fatalf("packageDrift = (%d,%d), want (1,1)", missing, mismatch)
	}
}

func TestCrashIssue(t *testing.T) {
	cases := map[string]string{
		"OpenBSD 7.9 booting normally":           "",
		"panic: uvm_fault: protection fault":     "panic logged",
		"WARNING: / was not properly unmounted":  "unclean filesystem",
	}
	for dmesg, want := range cases {
		if got := crashIssue(dmesg); got != want {
			t.Errorf("crashIssue(%q) = %q, want %q", dmesg, got, want)
		}
	}
}

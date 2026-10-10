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
		name    string
		section int
		bat     int
		proton  bool
		want    int
	}{
		{"audio", 0, 0, false, 2},
		{"display", 1, 0, false, 3},
		{"network", 2, 0, false, 3},
		{"power", 3, 80, false, 6},
		{"power-no-battery", 3, 255, false, 1},
		{"privacy", 4, 0, false, 2},
		{"privacy-proton", 4, 0, true, 3},
		{"system", 5, 0, false, 3},
	}
	for _, c := range cases {
		m := newTestModel()
		m.section = c.section
		m.batPercent = c.bat
		m.proton = c.proton
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

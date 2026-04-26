package config

import (
	"slices"
	"testing"
)

func TestGetAllModulesOrdered_Empty(t *testing.T) {
	cfg := &Config{}
	ordered, err := cfg.GetAllModulesOrdered()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ordered) != 0 {
		t.Fatalf("expected 0 modules, got %d", len(ordered))
	}
}

func TestGetAllModulesOrdered_NoDeps(t *testing.T) {
	cfg := &Config{
		Core: map[string]Module{
			"base": {Start: "base"},
		},
		System: map[string]Module{
			"tools": {Start: "tools"},
		},
	}
	ordered, err := cfg.GetAllModulesOrdered()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ordered) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(ordered))
	}
	names := []string{ordered[0].Name, ordered[1].Name}
	if !slices.Contains(names, "base") || !slices.Contains(names, "tools") {
		t.Fatalf("expected base and tools, got %v", names)
	}
}

func TestGetAllModulesOrdered_LinearDeps(t *testing.T) {
	cfg := &Config{
		Core: map[string]Module{
			"base":  {Start: "base"},
			"shell": {Start: "shell", Depends: []string{"core.base"}},
		},
	}
	ordered, err := cfg.GetAllModulesOrdered()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ordered) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(ordered))
	}
	if ordered[0].Name != "base" {
		t.Fatalf("expected base first, got %s", ordered[0].Name)
	}
	if ordered[1].Name != "shell" {
		t.Fatalf("expected shell second, got %s", ordered[1].Name)
	}
}

func TestGetAllModulesOrdered_MissingDep(t *testing.T) {
	cfg := &Config{
		Core: map[string]Module{
			"shell": {Start: "shell", Depends: []string{"core.missing"}},
		},
	}
	_, err := cfg.GetAllModulesOrdered()
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
	if !slices.Contains([]string{"depends on unknown", "missing"}, err.Error()) &&
		!containsAny(err.Error(), []string{"depends on unknown", "missing"}) {
		t.Fatalf("expected missing dependency error, got: %v", err)
	}
}

func TestGetAllModulesOrdered_CircularDep(t *testing.T) {
	cfg := &Config{
		Core: map[string]Module{
			"a": {Start: "a", Depends: []string{"core.b"}},
			"b": {Start: "b", Depends: []string{"core.a"}},
		},
	}
	_, err := cfg.GetAllModulesOrdered()
	if err == nil {
		t.Fatal("expected error for circular dependency, got nil")
	}
	if !containsAny(err.Error(), []string{"circular"}) {
		t.Fatalf("expected circular dependency error, got: %v", err)
	}
}

func TestGetAllModulesOrdered_CrossCategory(t *testing.T) {
	cfg := &Config{
		Core: map[string]Module{
			"base": {Start: "base"},
		},
		Desktop: map[string]Module{
			"i3": {Start: "i3", Depends: []string{"core.base"}},
		},
	}
	ordered, err := cfg.GetAllModulesOrdered()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ordered) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(ordered))
	}
	baseIdx := indexOf(ordered, "base")
	i3Idx := indexOf(ordered, "i3")
	if baseIdx == -1 || i3Idx == -1 {
		t.Fatal("expected both base and i3 in output")
	}
	if baseIdx > i3Idx {
		t.Fatalf("expected base before i3, got base@%d i3@%d", baseIdx, i3Idx)
	}
}

func indexOf(refs []ModuleRef, name string) int {
	for i, r := range refs {
		if r.Name == name {
			return i
		}
	}
	return -1
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && sub != "" && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

package installer

import (
	"testing"

	"openriot/config"
)

func TestSourceBuilds_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}
	err := SourceBuilds(cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSourceBuilds_SkipsNonSource(t *testing.T) {
	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Start: "base",
				End:   "base done",
				Type:  "Package",
			},
		},
	}
	err := SourceBuilds(cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSourceBuilds_TestMode(t *testing.T) {
	cfg := &config.Config{
		Core: map[string]config.Module{
			"build": {
				Start: "build",
				End:   "build done",
				Type:  "Source",
				Build: []any{"echo hello"},
			},
		},
	}
	err := SourceBuilds(cfg, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSourceBuilds_SourceWithEmptyBuild(t *testing.T) {
	cfg := &config.Config{
		Core: map[string]config.Module{
			"build": {
				Start: "build",
				End:   "build done",
				Type:  "Source",
			},
		},
	}
	err := SourceBuilds(cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

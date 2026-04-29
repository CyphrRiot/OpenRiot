package installer

import (
	"testing"

	"openriot/config"
)

func TestExecCommands_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}
	err := ExecCommands(cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecCommands_DryRun(t *testing.T) {
	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Start: "base",
				End:   "base done",
				Commands: []config.CommandEntry{
					{Desc: "test cmd", Cmd: "echo hello"},
				},
			},
		},
	}
	err := ExecCommands(cfg, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecCommands_SkipsEmptyCommand(t *testing.T) {
	cfg := &config.Config{
		Core: map[string]config.Module{
			"base": {
				Start: "base",
				End:   "base done",
				Commands: []config.CommandEntry{
					{Desc: "empty", Cmd: "   "},
				},
			},
		},
	}
	err := ExecCommands(cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

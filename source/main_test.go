package main

import (
	"strings"
	"testing"
)

func TestGetUsage_CoversAllCommands(t *testing.T) {
	cmds := initCommands()
	usage := getUsage()

	var missing []string
	for key := range cmds {
		if !strings.Contains(usage, key) {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		t.Fatalf("missing from usage: %v", missing)
	}
}

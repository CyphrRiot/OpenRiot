package main

import (
	"strings"
	"testing"

	"openriot/commands"
)

func TestUsageCoversAllCommands(t *testing.T) {
	registry := commands.NewRegistry()
	commands.RegisterAll(registry, new(bool))

	usage := registry.Usage("test")

	var missing []string
	for _, cat := range registry.Categories() {
		for _, cmd := range registry.CommandsInCategory(cat) {
			if !strings.Contains(usage, cmd.Name) {
				missing = append(missing, cmd.Name)
			}
		}
	}

	if len(missing) > 0 {
		t.Fatalf("missing from usage: %v", missing)
	}
}

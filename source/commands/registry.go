package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Command represents a single CLI command with metadata.
type Command struct {
	Name        string
	Category    string
	Description string
	Run         func(args []string) error
}

// Registry holds all registered commands.
type Registry struct {
	cmds map[string]*Command
}

// NewRegistry creates a new empty command registry.
func NewRegistry() *Registry {
	return &Registry{cmds: make(map[string]*Command)}
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd *Command) {
	r.cmds[cmd.Name] = cmd
}

// Get retrieves a command by name.
func (r *Registry) Get(name string) (*Command, bool) {
	cmd, ok := r.cmds[name]
	return cmd, ok
}

// Categories returns a sorted list of all category names.
func (r *Registry) Categories() []string {
	seen := make(map[string]bool)
	for _, cmd := range r.cmds {
		seen[cmd.Category] = true
	}
	var cats []string
	for cat := range seen {
		cats = append(cats, cat)
	}
	sort.Strings(cats)
	return cats
}

// CommandsInCategory returns commands sorted by name for a given category.
func (r *Registry) CommandsInCategory(category string) []*Command {
	var list []*Command
	for _, cmd := range r.cmds {
		if cmd.Category == category {
			list = append(list, cmd)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}

// Usage generates formatted help text from the registry.
func (r *Registry) Usage(version string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "openriot %s\n", version)
	fmt.Fprintf(&b, "Usage: openriot <command>\n")
	for _, cat := range r.Categories() {
		fmt.Fprintf(&b, "\n%s:\n", cat)
		for _, cmd := range r.CommandsInCategory(cat) {
			fmt.Fprintf(&b, "  %-23s %s\n", cmd.Name, cmd.Description)
		}
	}
	return b.String()
}

// Dispatch looks up and runs a command by name, returning its exit code.
func (r *Registry) Dispatch(name string, args []string) int {
	cmd, ok := r.Get(name)
	if !ok {
		return -1 // not found
	}
	if err := cmd.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%s error: %v\n", strings.TrimPrefix(cmd.Name, "--"), err)
		return 1
	}
	return 0
}

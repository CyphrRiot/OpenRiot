package notify

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Notification represents a single notification
type Notification struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Urgency   string `json:"urgency"` // "low", "normal", "critical"
	Timestamp int64  `json:"timestamp"`
	Expires   int64  `json:"expires,omitempty"` // Unix timestamp, 0 = never expires
}

// State wraps the notification list
type State struct {
	NextID        int            `json:"next-id"`
	Notifications []Notification `json:"notifications"`
}

const stateFile = ".cache/openriot/notifications.json"

func statePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, stateFile)
}

func load() (*State, error) {
	path := statePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{NextID: 1}, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func save(s *State) error {
	path := statePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Add appends a new notification
// expiresAt: Unix timestamp when notification should expire, 0 = never
func Add(title, body, urgency string, expiresAt int64) error {
	s, err := load()
	if err != nil {
		return err
	}
	n := Notification{
		ID:        s.NextID,
		Title:     title,
		Body:      body,
		Urgency:   urgency,
		Timestamp: time.Now().Unix(),
		Expires:   expiresAt,
	}
	s.Notifications = append(s.Notifications, n)
	s.NextID++
	return save(s)
}

// Dismiss removes a notification by ID; if id <= 0, removes oldest
func Dismiss(id int) error {
	s, err := load()
	if err != nil {
		return err
	}
	if len(s.Notifications) == 0 {
		return nil
	}
	if id <= 0 {
		s.Notifications = s.Notifications[1:]
	} else {
		for i, n := range s.Notifications {
			if n.ID == id {
				s.Notifications = append(s.Notifications[:i], s.Notifications[i+1:]...)
				break
			}
		}
	}
	return save(s)
}

// Clear removes all notifications
func Clear() error {
	s, err := load()
	if err != nil {
		return err
	}
	s.Notifications = nil
	return save(s)
}

// List returns all current notifications sorted by timestamp (oldest first)
func List() ([]Notification, error) {
	s, err := load()
	if err != nil {
		return nil, err
	}
	sort.Slice(s.Notifications, func(i, j int) bool {
		return s.Notifications[i].Timestamp < s.Notifications[j].Timestamp
	})
	return s.Notifications, nil
}

// dismissWithoutReload removes a notification by ID without reloading state
func dismissWithoutReload(id int) error {
	s, err := load()
	if err != nil {
		return err
	}
	for i, n := range s.Notifications {
		if n.ID == id {
			s.Notifications = append(s.Notifications[:i], s.Notifications[i+1:]...)
			break
		}
	}
	return save(s)
}

// Status outputs JSON for polybar custom module
// Skips and auto-dismisses any expired notifications
func Status() error {
	notes, err := List()
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		fmt.Println(`{"text": ""}`)
		return nil
	}

	now := time.Now().Unix()

	// First pass: collect expired IDs
	var expiredIDs []int
	for _, n := range notes {
		if n.Expires > 0 && now > n.Expires {
			expiredIDs = append(expiredIDs, n.ID)
		}
	}

	// Dismiss all expired notifications
	for _, id := range expiredIDs {
		dismissWithoutReload(id)
	}

	// Reload to get fresh list after dismissals
	notes, err = List()
	if err != nil || len(notes) == 0 {
		fmt.Println(`{"text": ""}`)
		return nil
	}

	// Find first non-expired notification
	for _, n := range notes {
		if n.Expires > 0 && now > n.Expires {
			continue
		}
		// Found a valid notification - display it
		text := n.Title
		if n.Body != "" {
			text = n.Title + ": " + n.Body
		}

		// Truncate to 40 chars
		if len(text) > 40 {
			text = text[:40] + "..."
		}

		// Urgency icon
		icon := "[BELL]" // normal
		if n.Urgency == "critical" {
			icon = "" // critical
		} else if n.Urgency == "low" {
			icon = "[!]" // warning/low
		}

		// Tooltip (escape quotes)
		tooltip := strings.ReplaceAll(n.Title, `"`, `\"`)
		if n.Body != "" {
			tooltip += `\n` + strings.ReplaceAll(n.Body, `"`, `\"`)
		}

		fmt.Printf(`{"text": "%s %s", "tooltip": "%s", "class": "%s"}`+"\n",
			icon, text, tooltip, n.Urgency)
		return nil
	}

	// All expired or none exist
	fmt.Println(`{"text": ""}`)
	return nil
}

// Setup scales dunstrc based on screen resolution.
// If screen width > 1920, uses larger width (500) and font size (12).
// Otherwise uses defaults from template (400 width, size 10).
func Setup() int {
	home := os.Getenv("HOME")

	// Get screen resolution using xrandr
	width := getScreenWidth()

	// Read template config
	templatePath := filepath.Join(home, ".local/share/openriot/config/dunst/dunstrc")
	configPath := filepath.Join(home, ".config/dunst/dunstrc")

	template, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dunst setup: cannot read template: %v\n", err)
		return 1
	}

	content := string(template)

	// Only scale up for larger screens (>1920)
	if width > 1920 {
		content = strings.ReplaceAll(content, "width = 400", "width = 500")
		content = strings.ReplaceAll(content, "FiraCode Nerd Font 10", "FiraCode Nerd Font 14")
	}

	// Write scaled config
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "dunst setup: cannot create dir: %v\n", err)
		return 1
	}
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "dunst setup: cannot write config: %v\n", err)
		return 1
	}

	return 0
}

// getScreenWidth returns the screen width in pixels using xrandr
func getScreenWidth() int {
	cmd := exec.Command("xrandr")
	output, err := cmd.Output()
	if err != nil {
		return 1920 // default to 1080p
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "connected") {
			re := regexp.MustCompile(`(\d+)x(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				var width int
				fmt.Sscanf(matches[1], "%d", &width)
				return width
			}
		}
	}
	return 1920
}

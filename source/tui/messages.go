package tui

// LogMsg represents a log message
type LogMsg string

// ProgressMsg represents progress update
type ProgressMsg float64

// StepMsg represents a step update
type StepMsg string

// DoneMsg indicates successful completion
type DoneMsg struct{}

// FailureMsg indicates installation failure
type FailureMsg struct {
	Error string
}

// GitUsernameMsg carries git username input
type GitUsernameMsg string

// GitEmailMsg carries git email input
type GitEmailMsg string

// RebootMsg carries reboot decision
type RebootMsg bool

// UpgradeMsg triggers OpenBSD upgrade confirmation prompt
type UpgradeMsg struct{}

// InputRequestMsg requests user input
type InputRequestMsg struct {
	Mode   string // git-username, git-email, git-confirm, etc.
	Prompt string
}

// GitConfirmMsg carries git credential confirmation
type GitConfirmMsg bool

// OpenRouterConfirmMsg carries OpenRouter setup decision
type OpenRouterConfirmMsg bool

// OpenRouterKeyMsg carries OpenRouter API key input
type OpenRouterKeyMsg string

// Helper functions for external packages
var versionGetter func() string
var logPathGetter func() string

// SetVersionGetter sets the function to get version
func SetVersionGetter(fn func() string) {
	versionGetter = fn
}

// SetLogPathGetter sets the function to get log path
func SetLogPathGetter(fn func() string) {
	logPathGetter = fn
}

// GetVersion returns the current version
func GetVersion() string {
	if versionGetter != nil {
		return versionGetter()
	}
	return "unknown"
}

// GetLogPath returns the current log path
func GetLogPath() string {
	if logPathGetter != nil {
		return logPathGetter()
	}
	return "/tmp/openriot-install.log"
}

// Progress and step callbacks — set by main.go to feed TUI
var progressCallback func(float64)
var stepCallback func(string)

// SetProgressCallback sets the callback for progress bar updates (0.0 to 1.0)
func SetProgressCallback(fn func(float64)) {
	progressCallback = fn
}

// SetStepCallback sets the callback for current step name updates
func SetStepCallback(fn func(string)) {
	stepCallback = fn
}

// Git callback functions
var gitCompletionCallback func(bool)
var gitUsernameCallback func(string)
var gitEmailCallback func(string)

// SetGitCallbacks sets the callback functions for git credential handling
func SetGitCallbacks(completion func(bool), username func(string), email func(string)) {
	gitCompletionCallback = completion
	gitUsernameCallback = username
	gitEmailCallback = email
}

// Upgrade callback function (OpenBSD system/pkg upgrade)
var upgradeCompletionCallback func(bool)

// SetUpgradeCallback sets the callback function for upgrade confirmation handling
func SetUpgradeCallback(callback func(bool)) {
	upgradeCompletionCallback = callback
}

// OpenRouter callback function
var openRouterCompletionCallback func(bool)
var openRouterKeyCallback func(string)

// SetOpenRouterCallbacks sets the callback functions for OpenRouter setup handling
func SetOpenRouterCallbacks(completion func(bool), apiKey func(string)) {
	openRouterCompletionCallback = completion
	openRouterKeyCallback = apiKey
}

# Notification Cooldown Mechanism

## Context

OpenRiot has ~45 notification calls spread across multiple source files using `notify-send`. When multiple commands trigger simultaneously (e.g., polybar modules querying at same interval, rapid key presses), dunst can receive overlapping notifications that may cause crashes on OpenBSD.

The icon fallback fix (`GetIconPath` returning `info.png`) addresses malformed icon paths. A cooldown mechanism would address notification spam/overlap.

## Solution

A lightweight cooldown file (`~/.cache/openriot/notify-cooldown`) that tracks last notification timestamp. Any `notify-send` call checks if cooldown is active before proceeding.

## Implementation

### File Location
`~/.cache/openriot/notify-cooldown` (0600 permissions)

### File Format
Single integer: Unix timestamp in nanoseconds of last notification

### Cooldown Duration
500ms (configurable constant)

### New Function: `notify.ShouldNotify() bool`

```go
// source/notify/cooldown.go

const (
    cooldownFile = ".cache/openriot/notify-cooldown"
    cooldownMs   = 500 // milliseconds
)

func shouldNotify() bool {
    home, _ := os.UserHomeDir()
    path := filepath.Join(home, cooldownFile)

    data, err := os.ReadFile(path)
    if err != nil {
        // No cooldown file = allowed
        return true
    }

    lastNs, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
    if err != nil {
        // Corrupted file = reset and allow
        return true
    }

    elapsed := time.Since(time.Unix(0, lastNs))
    return elapsed > time.Duration(cooldownMs)*time.Millisecond
}

func recordNotification() error {
    home, _ := os.UserHomeDir()
    dir := filepath.Join(home, ".cache/openriot")
    os.MkdirAll(dir, 0755)

    path := filepath.Join(dir, cooldownFile)
    return os.WriteFile(path, []byte(strconv.FormatInt(time.Now().UnixNano(), 10)), 0600)
}
```

### Call Sites

Add cooldown check to the `Notify` wrapper function (or each notify-send call):

```go
func notifySend(icon, title, body string, urgency string, timeout int) {
    if !shouldNotify() {
        return
    }
    recordNotification()
    // exec.Command("/usr/local/bin/notify-send", ...).Run()
}
```

## Impact

- **Files to modify:** `source/notify/notify.go` (new cooldown.go), plus update all notify-send call sites
- **Call sites:** ~45 locations in main.go, crypto.go, battery.go, network/, audio/, etc.
- **Risk:** Low - if bug in cooldown, notifications just fire more (not less safe)
- **Side effect:** Some rapid notifications may be dropped (by design)

## Alternative

A full file lock (`~/.cache/openriot/notify.lock`) could prevent concurrent notifications, but adds crash-risk (if holding lock when process dies). Cooldown is simpler and safer.

## Status

Proposed - awaiting crash reports to confirm this is needed after icon fallback fix.
# Notification System - Functional Spec

## Overview

OpenRiot uses a centralized notification system to prevent dunst crashes and provide consistent icon/error handling across all notification call sites.

## Architecture

### Core Function: `notify.SendNotify()`

**Location:** `source/notify/cooldown.go`

**Signature:**
```go
func SendNotify(iconName, title, body, urgency string, timeoutMs int) error
```

**Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `iconName` | string | Icon filename only (e.g., `"cpu.png"`) - NOT full path |
| `title` | string | Notification title |
| `body` | string | Notification body text |
| `urgency` | string | `"normal"` or `"critical"` |
| `timeoutMs` | int | Timeout in milliseconds |

**Behavior:**
1. Checks cooldown - skips if < 500ms since last notification
2. Records current timestamp to cooldown file
3. Resolves icon path via `paths.GetIconPath(iconName)` (includes fallback to `info.png`)
4. Calls `notify-send` with appropriate flags

### Internal Functions

**`ShouldSend() bool`** - Checks if cooldown period has elapsed

**`RecordSend() error`** - Writes current nanosecond timestamp to cooldown file

## Cooldown Mechanism

**File:** `~/.cache/openriot/notify-cooldown` (permissions: 0600)

**Duration:** 500ms

**Purpose:** Prevent dunst crashes from rapid-fire notifications when multiple polybar modules query simultaneously or user spam-triggers keybindings.

## Icon Resolution

`SendNotify` internally calls `paths.GetIconPath(iconName)` which:
- Returns full path to icon in `~/.local/share/openriot/config/icons/`
- Falls back to `info.png` if icon not found (prevents dunst crashes from missing icons)

## Migration Status

### Completed (use `notify.SendNotify`)

All `main.go` notification calls converted.

### Remaining (still use raw `exec.Command`)

- `source/crypto/crypto.go:1007` - crypto notification
- `source/nightlight/nightlight.go:85` - night light toggle
- `source/workspace/workspace.go:47` - workspace switch
- `source/update/update.go:62,70,79` - update checks
- `source/lock/lock.go:46` - screen lock
- `source/rofi/rofi.go:76` - app launcher
- `source/wireguard/wireguard.go:45,54,64` - VPN status
- `source/audio/volume.go:27` - volume changes
- `source/display/display.go:28` - brightness changes

## Examples

**Good (using centralized function):**
```go
notify.SendNotify("cpu.png", "CPU", "Usage: 45%", "normal", 5000)
notify.SendNotify("proton-drive.png", "Proton Drive", "Syncing...", "normal", 2000)
notify.SendNotify("power.png", "Power", "Shutting down...", "critical", 5000)
```

**Bad (bypassing centralized function):**
```go
exec.Command("/usr/local/bin/notify-send", "-i", paths.GetIconPath("cpu.png"), "-t", "5000", "CPU", details).Start()
```

## Future Improvements

Potential additions (not implemented):
- Notification logging to file
- Retry logic on failure
- Configurable cooldown duration
- Rate limiting per category (crypto vs system vs app)
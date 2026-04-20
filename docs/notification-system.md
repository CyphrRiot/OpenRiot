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

All notification call sites have been migrated to use `notify.SendNotify()`.

### Completed (use `notify.SendNotify`)

| Module | Location | Status |
|--------|----------|--------|
| crypto | `source/crypto/crypto.go:1004` | ✓ Migrated |
| nightlight | `source/nightlight/nightlight.go:85` | ✓ Migrated |
| workspace | `source/workspace/workspace.go:47` | ✓ Migrated |
| update | `source/update/update.go:61,67,74` | ✓ Migrated |
| lock | `source/lock/lock.go:47` | ✓ Migrated |
| rofi | `source/rofi/rofi.go` | ✓ Migrated |
| wireguard | `source/wireguard/wireguard.go` | ✓ Migrated |
| audio | `source/audio/volume.go:24` | ✓ Migrated |
| display | `source/display/display.go:28` | ✓ Migrated |

**All modules now use centralized notification system with cooldown protection.**

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
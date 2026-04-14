# Polybar State Management Architecture

## Overview

Polybar modules fall into two categories: **state modules** (user-triggered) and **polling modules** (external state). State modules read from a cached state file instead of spawning the binary on each poll.

---

## State File

**Location:** `~/.cache/openriot/states.toml`

**Contents:**
```toml
[night-light]
enabled = false

[volume]
percent = 75
muted = false
```

---

## Categories

### State Modules (No Polling)

These are user-triggered state changes. Polybar reads state file at startup and on click.

| Module | State File Key | Init on Startup |
|--------|----------------|-----------------|
| volume | `volume.muted`, `volume.percent` | Yes |
| night-light | `night-light.enabled` | Yes |

**Polybar config:**
```ini
[module/volume]
type = custom/script
exec-init = openriot --init-states
exec = grep "^muted" ~/.cache/openriot/states.toml | cut -d= -f2
click-left = openriot --volume toggle
scroll-up = openriot --volume inc
scroll-down = openriot --volume dec

[module/night-light]
type = custom/script
exec-init = openriot --init-states
exec = grep "^enabled" ~/.cache/openriot/states.toml | cut -d= -f2
click-left = openriot --night-light
```

### Polling Modules (Binary Exec)

External or unpredictable state. Binary spawns at interval.

| Module | Interval | Reason |
|--------|----------|--------|
| battery | 60 sec | Hardware state changes |
| wireguard | 30 sec | Network state changes |
| transmission | 30 sec | Daemon state changes |
| weather | 1800 sec | External API |
| update-status | 3600 sec | External API |

---

## Flow

### Startup
1. X session starts
2. Polybar loads, runs `exec-init` → `openriot --init-states`
3. Writes current state to `~/.cache/openriot/states.toml`

### Click Action
1. User clicks module
2. Binary runs `--volume toggle` or `--night-light`
3. Binary updates state file
4. Binary performs action (toggle mute, toggle redshift)
5. Polybar re-reads state file on next poll

---

## Binary Flags

### `--init-states`
Reads current hardware/software state, writes to `~/.cache/openriot/states.toml`.

```go
func initStates() {
    // Read night-light state (check if redshift is running)
    // Read volume state (check sndioctl)
    // Write to states.toml
}
```

### `--volume toggle|inc|dec`
- Updates `volume.muted` or `volume.percent` in state file
- Performs actual action (sndioctl)
- Sends notification

### `--night-light`
- Updates `night-light.enabled` in state file
- Toggles redshift service
- Sends notification

---

## Performance Impact

### Before
| Type | Calls/sec |
|------|-----------|
| State modules (2 modules) | ~0.4/sec |
| Polling modules | ~0.5/sec |
| Workspace/Title | ~2.5/sec |
| **Total** | ~3.4/sec |

### After
| Type | Calls/sec |
|------|-----------|
| State modules | ~0/sec (file read) |
| Polling modules | ~0.1/sec |
| Workspace/Title | ~2.5/sec |
| **Total** | ~2.6/sec |

**Reduction: ~25% overall, ~100% for state modules**

---

## Files Impacted

| File | Change |
|------|---------|
| `source/main.go` | Add `--init-states` flag |
| `source/volume/volume.go` | Add state file write |
| `source/nightlight/nightlight.go` | Add state file write |
| `config/polybar/config.ini` | Update volume/night-light modules |
| `source/polybar/polybar.go` | Remove status-only functions |
| `AGENTS.md` | Document architecture |

---

## Implementation Order

1. Create `source/state/state.go` - TOML read/write utilities
2. Add `--init-states` to main.go
3. Modify volume/nightlight to write state file
4. Update polybar config to use file reads
5. Remove old binary status flags (or deprecate)
6. Update AGENTS.md

---

## Note on WireGuard

WireGuard state can change unexpectedly (network drop, VPN disconnect). It requires polling but interval can be increased to 30s without UX impact since VPN state is relatively stable.

# Network TUI Functional Specification

> "The network is the computer." — John Gage, Sun Microsystems (we're just making it usable.)

## Overview

`openriot --nmtui` is a Bubble Tea-based Wi-Fi network manager TUI for
OpenBSD. It replaces the Linux-centric `nmtui` concept with a backend that
talks exclusively to OpenBSD's native networking stack:
`ifconfig`, `hostname.if`, and `netstart`. No cgo, no dbus, no nmcli.

## Scope (MVP)

| In Scope | Out of Scope |
|----------|-------------|
| Scan Wi-Fi networks | WPA-Enterprise (802.1X) |
| Connect to open networks | Profile CRUD (create/ edit / forget) |
| Connect to WPA2-PSK networks | Hidden SSID manual entry |
| Disconnect active association | Wi-Fi radio toggle |
| View active connection info | Auto-elevation (no self-sudo) |
| Scrollable paginated list | Connection history / logging |

## Architecture

```
┌─────────────────────────────────────┐
│  source/nmtui/                      │
│  ┌─────────┐  ┌─────────┐          │
│  │ backend │  │ model   │          │
│  │  .go    │──│  .go    │          │
│  │(OpenBSD)│  │(BubbleTea state)   │
│  └────┬────┘  └────┬────┘          │
│       │            │                │
│  ┌────▼────────────▼────┐          │
│  │      update.go       │          │
│  │  Cmd / Msg / handlers │          │
│  └────┬─────────────────┘          │
│       │                             │
│  ┌────▼────┐                       │
│  │ view.go │  lipgloss rendering   │
│  └─────────┘                       │
│       │                             │
│  ┌────▼────┐                       │
│  │nmtui.go │  Run() entry point    │
│  └─────────┘                       │
└─────────────────────────────────────┘
```

All mutations (`ifconfig`, `netstart`) require `doas`. If `doas` is not in
`PATH`, the TUI displays an error and returns without attempting elevation.
Read-only operations (scan, view info) run as the invoking user.

## Backend API

Defined in `source/nmtui/backend.go`.

```go
type WiFiAP struct {
    SSID     string
    BSSID    string
    Signal   int    // 0-100 percentage, -1 if unknown
    SignalValid bool
    Security string // "open", "wpa2", "wpa3", "wep", "unknown"
}

type ConnectionInfo struct {
    Device  string
    SSID    string
    IP      string
    Netmask string
    Gateway string
    DNS     []string
    MAC     string
    State   string
}
```

| Function | External Command | Notes |
|----------|-----------------|-------|
| `FindWiFiInterface()` | `ifconfig -a` | Regex `^([a-z]+[0-9]+):\s+flags=` then `ieee80211` search. Tested on `iwx0`. |
| `ScanWiFi(iface)` | `ifconfig <iface> scan` | Runs unprivileged; driver-dependent. |
| `IsWiFiConnected(iface)` | `ifconfig <iface>` | Checks `join "SSID"` presence. |
| `ConnectOpen(iface, ssid)` | `doas ifconfig …` + `doas /etc/netstart` | Two-step: associate then DHCP. |
| `ConnectWPA(iface, ssid, key)` | `doas ifconfig … wpakey …` + `doas /etc/netstart` | No key validation; passes verbatim. |
| `Disconnect(iface)` | `doas ifconfig <iface> down` | Brings interface down. |
| `GetConnectionInfo(iface)` | `ifconfig`, `netstat -rn`, `/etc/resolv.conf` | Parses join/inet/lladdr/status; gateway from default route; DNS from nameserver lines. |
| `GetKnownSSIDs(iface)` | `/etc/hostname.<iface>` | Read-only. Extracts `nwid` values. |
| `SignalToPercent(signal, valid)` | — | Clamps to [0,100]; returns -1 if !valid. |

### Parser Strategy

All `ifconfig` output is parsed with regex field extraction (not column
counting). Fields are space-delimited and order-independent after the initial
keyword. This is defensive against driver differences (iwm, iwn, athn, iwx, etc.)
and minor format changes across OpenBSD releases.

## View States

`source/nmtui/model.go` defines `viewState`:

| State | User Sees | Input |
|-------|-----------|-------|
| `stateList` | Paginated scrollable network list, header, help bar | ↑/↓/j/k navigate, Enter select, r refresh, i info, d disconnect, ? help, q quit |
| `statePassword` | Centered password box with SSID label | Typing (• echo), Enter confirm, Esc cancel |
| `stateConnecting` | Centered spinner + "Connecting to SSID..." | No input (blocking) |
| `stateResult` | Centered success or error message | Esc to return to list |
| `stateActiveInfo` | Boxed connection details (IP, gateway, DNS, MAC) | Esc to return to list |
| `stateConfirmDisconnect` | "Disconnect from SSID?" | y confirm, n/Esc cancel |

## Pagination

The network list is viewport-limited to `m.height - 4` rows (header + footer).
The cursor is always centered in the viewport. Overflow is indicated with
`▲ …` above and `▼ …` below. This prevents the list from pushing the
help bar off-screen on dense scans.

## Key Map

Defined in `defaultKeys` (`model.go`):

| Key(s) | Action |
|--------|--------|
| `↑`, `k` | Move cursor up |
| `↓`, `j` | Move cursor down |
| `Enter` | Select / confirm |
| `Esc` | Back / cancel |
| `q`, `Ctrl+C` | Quit |
| `r` | Refresh scan |
| `i` | Show active connection info |
| `d` | Prompt to disconnect |
| `?` | Toggle full help overlay |

## Files

| File | Lines | Responsibility |
|------|-------|---------------|
| `source/nmtui/backend.go` | ~390 | OpenBSD network primitives + parsers |
| `source/nmtui/backend_test.go` | ~300 | Unit tests for scan parser, SSID extractor, signal mapping, ifconfig connection parser |
| `source/nmtui/model.go` | ~140 | Bubble Tea model, view-state constants, key map, constructor, `Init()` |
| `source/nmtui/update.go` | ~250 | `Update()`, custom `tea.Msg` types, `tea.Cmd` generators, per-state key handlers |
| `source/nmtui/view.go` | ~300 | `View()`, lipgloss styles, pagination, signal bar rendering, per-state layouts |
| `source/nmtui/nmtui.go` | ~30 | `Run()` entry point: detect iface, launch `tea.Program` with alt-screen + mouse |
| `source/commands/commands.go` | +1 registration | `--nmtui` registered under "Network & Battery" category |

## Dependencies

Already present in `source/go.mod`:
- `github.com/charmbracelet/bubbletea v1.3.10`
- `github.com/charmbracelet/lipgloss v1.1.0`

Added during implementation:
- `github.com/charmbracelet/bubbles v1.0.0` (`help`, `key`, `spinner`, `textinput`)

No new system packages required.

## Testing

```
cd source && go test ./nmtui/ -v
```

Backend tests cover:
- `TestFindWiFiInterface` — live detection (skipped gracefully if no Wi-Fi)
- `TestParseScanOutput` — 7 cases: single AP, multiple, hidden, quoted SSID, no signal, empty, mixed with non-`nwid` lines
- `TestExtractSSID` — quoted, unquoted, hidden, keyword boundary, empty input
- `TestSignalToPercent` — boundary and clamping values
- `TestParseIfconfigConnection` — full `iwx0` block with join, inet, lladdr, status
- `TestGetKnownSSIDs` — live read of `/etc/hostname.<iface>`
- `TestIsWiFiConnected` — live check
- `TestUnicodeSSIDs` — hex-encoded SSID decoding (Gödel, Café, etc.)

All 14 tests pass on OpenBSD with `iwx0`.

## Build Status

| Step | Status |
|------|--------|
| 1 — Backend primitives + tests | **Complete** |
| 2 — Connection primitives + tests | **Complete** (all tests pass) |
| 3 — Bubble Tea model (`model.go`) | **Complete** |
| 4 — Views + lipgloss (`view.go`) | **Complete** |
| 5 — Update handlers (`update.go`) | **Complete** |
| 6 — Entry point (`nmtui.go`) | **Complete** |
| 7 — `--nmtui` registration | **Complete** |
| 8 — Build + smoke test | **Complete** — `make` succeeds, binary runs |
| 9 — Pagination fix | **Complete** — auto-scan on launch, viewport scrolling |
| 10 — Docs update | **This document** |

## Known Limitations & Future Work

1. **No profile management.** OpenBSD has no NetworkManager-style profile DB.
   `GetKnownSSIDs()` reads `/etc/hostname.<iface>` but cannot edit it.
   Users must edit `hostname.if` manually or use the displayed `doas` commands.

2. **WPA-Enterprise not supported.** `wpakey` only handles PSK. 802.1X
   would require `wpa_supplicant` configuration files and is out of MVP scope.

3. **`doas` required for all mutations.** The TUI does not self-elevate.
   If `doas` is absent, the user gets a clear error message.

4. **`ifconfig scan` may require root on some drivers.** The backend attempts
   unprivileged scan first; if it fails, the user must elevate externally.

5. **No connection retry or timeout.** `ConnectOpen` / `ConnectWPA` are
   synchronous. A slow DHCP lease will block the spinner.

6. **No auto-connect on launch.** The TUI displays the list and current
   connection status but does not attempt to join a known network automatically.

## Implementation Notes

### Signal Format
`ifconfig scan` outputs signal as a percentage (`83%`), not dBm. The parser
uses `(\d+)%` and stores an `int` 0–100. The `SignalValid` flag distinguishes
"parsed" from "unknown".

### Security from Comma-Separated Flags
OpenBSD scan lines end with comma-separated flags like `privacy,wpa2` or
`privacy,wpa3,wpa2`. The parser detects security in priority order:
wpa3 > wpa2 > wep > privacy > open. The old `wpaprotos` field from Linux
tools does not exist here.

### Hidden Networks
`nwid ""` lines are skipped entirely. They are filtered at parse time
(`ssid == "" → continue`) and never appear in the list.

### Deduplication
`parseScanOutput` deduplicates by SSID, keeping the strongest signal entry.
If the same network appears on multiple channels, only the best BSSID is kept.

### Pagination
The viewport centers on the cursor. Overflow indicators `▲ …` and `▼ …`
appear when the list exceeds `m.height - 4` rows. This prevents the help bar
from being pushed off-screen.

### Hex-Encoded Unicode SSIDs
Some access points broadcast SSIDs with non-ASCII characters. `ifconfig scan`
outputs these as hex strings like `nwid 0x47c3b664656c …` where the hex
decodes to UTF-8 (that example is "Gödel"). The `extractSSID` parser detects
`^0x([0-9a-fA-F]+)$`, hex-decodes the bytes, validates UTF-8, and returns
the decoded string. Invalid UTF-8 falls through to the raw hex string.

---

*Last updated: 2025-05-13*

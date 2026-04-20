# OpenRiot Architecture

Comprehensive reference for OpenRiot project structure and configuration.

---

# 📂 PROJECT STRUCTURE

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          REPOSITORY (~/Code/OpenRiot)                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────────┐  │
│  │   source/    │  │   config/    │  │        install/              │  │
│  │              │  │              │  │                              │  │
│  │ main.go      │  │ i3/          │  │ packages.yaml  ← Package     │  │
│  │ polybar/     │  │ polybar/     │  │   definitions               │  │
│  │ backgrounds/ │  │ rofi/        │  │                              │  │
│  │ wireguard/   │  │ alacritty/   │  │ openriot      ← Built binary │  │
│  │ crypto/      │  │ fastfetch/   │  │ motd                         │  │
│  │ audio/       │  │ dunst/       │  │                              │  │
│  │ display/     │  │ fish/        │  │                              │  │
│  │ installer/   │  │ helix/       │  │                              │  │
│  │ ...         │  │ ...         │  │                              │  │
│  └──────────────┘  └──────────────┘  └──────────────────────────────┘  │
│                                                                         │
│  backgrounds/          config/bin/           install/openriot            │
│  (16 wallpapers)       (scripts)            (binary)                  │
└─────────────────────────────────────────────────────────────────────────┘
```

---

# 📁 KEY LOCATIONS

| What | Location |
|------|---------|
| Repo | `~/Code/OpenRiot/` |
| Installed | `~/.local/share/openriot/` |
| XDG configs | `~/.config/` |
| Binary | `~/.local/share/openriot/install/openriot` |
| Logs | `~/.cache/openriot/` |
| Lock script | `~/.local/share/openriot/config/bin/openriot-lock.sh` |
| Rofi apps | `~/.config/rofi/apps.txt` |
| Crypto config | `~/.config/crypto.toml` |

## Deployment Flow

```
config/fish/*        → ~/.config/fish/
config/i3/*          → ~/.config/i3/
config/polybar/*     → ~/.config/polybar/config.ini (config only)
config/polybar/scripts/* → ~/.local/share/openriot/config/polybar/scripts/
config/bin/*         → ~/.local/share/openriot/config/bin/
config/rofi/*        → ~/.config/rofi/
config/fonts/*       → ~/.local/share/fonts/
config/Xresources    → ~/.Xresources
config/xsession/*    → ~/.xsession (ALWAYS overwrite)
config/picom.conf   → ~/.config/picom.conf
config/alacritty/*  → ~/.config/alacritty/
install/openriot    → ~/.local/share/openriot/install/openriot
```

**Polybar scripts are special:** They deploy to `~/.local/share/openriot/...` NOT `~/.config/polybar/`

---

# 📋 COMPLETE CLI REFERENCE

**Binary location:** `~/.local/share/openriot/install/openriot`

## Installation Commands

| Flag | Description |
|------|-------------|
| `--install` | Deploy configs, run commands, build source (no packages) |
| `--install-packages` | Install packages from packages.yaml (safe one-by-one) |
| `--source-builds` | Build software from source (crush, fonts, cursors) |
| `--packages` | List packages from packages.yaml |
| `--check-packages` | Verify packages.yaml versions against installed/available |
| `--sync-packages` | Update packages.yaml to latest available versions |
| `--version-check` | Check if remote version is newer |

## Package Version Workflow

**After `doas pkg_add -u` (system upgrade):**
```bash
openriot --check-packages   # See what's outdated
openriot --sync-packages    # Update packages.yaml to match installed
git add packages.yaml && git push  # Distribute to users
```

**Users run:**
```bash
./setup.sh  # Calls --install-packages using yaml from repo
```

**Key points:**
- `--sync-packages` only updates `packages.yaml`, NOT the system
- Creates backup: `packages.yaml.bak`
- Uses `pkg_info -a` for fast local lookup (no network needed)

## Wallpaper Commands

| Flag | Description |
|------|-------------|
| `--wallpaper-next` | Rotate to next wallpaper, save to `~/.current-background` |
| `--wallpaper-load` | Load saved wallpaper from `~/.current-background` |

## Power Commands

| Flag | Description |
|------|-------------|
| `--lock` | Lock screen via `openriot-lock.sh` |
| `--suspend` | Suspend system (`zzz`) |
| `--power-menu` | Rofi menu: Lock/Suspend/Reboot/Shutdown/Logout |
| `--suspend-if-undocked` | Auto-suspend when undocked |

## Volume/Brightness Commands

| Flag | Description |
|------|-------------|
| `--volume inc` | Increase volume |
| `--volume dec` | Decrease volume |
| `--volume toggle` | Toggle mute |
| `--brightness [args]` | Adjust brightness |

## Polybar Modules

| Flag | Output | Used By |
|------|--------|---------|
| `--polybar-metrics` | CPU + RAM % | polybar cpu module |
| `--polybar-volume` | Volume icon + % | polybar volume module |
| `--polybar-memory` | RAM icon + % | polybar memory module |
| `--wireguard-status` | 󰛳/󰅛/󰱓 | polybar wireguard module |

## Notification Commands

| Flag | Description |
|------|-------------|
| `--notify "title" "body"` | Send notification |
| `--notify "title" "body" --urgency critical` | Critical notification |
| `--notify-dismiss [id]` | Dismiss notification |
| `--notify-clear` | Clear all notifications |
| `--notify-dunst` | Show dunst status |
| `--mem-notify` | Memory usage notification |
| `--cpu-notify` | CPU usage notification |

## Network Commands

| Flag | Description |
|------|-------------|
| `--wireguard` | Toggle WireGuard VPN |

## Crypto Commands

| Flag | Description |
|------|-------------|
| `--crypto [BTC|ZEC|XMR|LTC]` | Show crypto price |
| `--crypto-refresh` | Clear cache, fetch fresh prices |

## Other Commands

| Flag | Description |
|------|-------------|
| `--share-log [file]` | Upload log to catbox.moe |
| `--version` | Show version |

---

# 🔧 BUILD & TEST

```bash
make build    # DEPRECATED - use 'make' or 'make install'
make         # Dev build (no install)
make install  # Build + deploy to ~/.local/share/openriot/install/
make verify  # Build + smoke test
make test    # Go tests
make deps    # Tidy Go dependencies
```

## Testing Workflow

**Config files:**
```bash
cp repo/path ~/.config/path  # Copy to local
# Restart relevant daemon
pkill -HUP polybar         # Polybar restart
# Test changes
```

**Go binary:**
```bash
make
# Test from install/openriot
./install/openriot --version
# Deploy when ready
make install
```

**NEVER copy files without explicit approval.**

## Version System

Version is in `VERSION` file (e.g., "1.4"). Injected at build time via ldflags:
```
-X main.version=$(cat VERSION)
```

---

# 🛠️ TECHNOLOGY STACK

## Platform
- **OpenBSD** (current/snapshots)
- **i3** window manager (NOT sway/wayland)
- **X11** (NOT Wayland)
- **Minimum resolution: 1920x1080**

## Languages
- **Go** - CLI binary (`source/*.go`)
- **YAML** - Package definitions (`install/packages.yaml`)
- **Fish** - Shell configuration
- **ini** - Polybar, i3, dunst configs

## Key Packages

| Package | Purpose |
|---------|---------|
| `polybar` | Status bar |
| `rofi` | App launcher |
| `i3` + `i3lock` | Window manager |
| `alacritty` | Terminal |
| `dunst` | Notifications |
| `picom` | Compositor |
| `redshift` | Night light |
| `xautolock` | Screen lock trigger |
| `clipmenu` | Clipboard manager |

---

# 📦 PACKAGES.YAML STRUCTURE

```yaml
core:
    base:
        packages: ["pkg-1"]
        configs: [{pattern: "source/*", target: "~/.config/dest"}]
        commands: [{desc: "Desc", cmd: "shell command"}]
        type: "Package"
        critical: true

desktop:
    i3:
        packages: ["i3", "polybar", "rofi"]
        configs:
            - pattern: "i3/*"
              preserve_if_exists: [keybindings.conf, monitors.conf]
        depends: [core.base, fonts.fira-code]
        type: "Package"

source:
    crush:
        type: "Source"
        build: ["curl ... && install"]
        depends: []

fonts:
    fira-code:
        type: "Source"
        build: ["curl ... && unzip"]
        depends: []
```

### Module Categories

| Category | Examples |
|----------|----------|
| `core` | base, shell |
| `desktop` | i3, i3-auto-layout, rofi-calc, apps, media |
| `source` | crush, bibata-cursor, stormy |
| `fonts` | fira-code |
| `system` | tools, services |

### Module Types

| Type | Meaning |
|------|---------|
| `Package` | Install via `pkg_add` |
| `Source` | Build from source |

### Config Pattern Types

| Pattern | Behavior |
|---------|----------|
| `dir/*` | Glob: copy all files in directory |
| `dir/file` | Single file |
| `bin/*` | Glob to `~/.local/bin` |

### Preserve Behavior

- `preserve_if_exists: [file]` = Only deploy if missing
- No preserve = Always overwrite
- Exception: `~/.xsession` ALWAYS overwrites (system config)

---

# 🖥️ POLYBAR MODULE REFERENCE

See also: `docs/polybar-state-architecture.md`, `docs/polybar-performance.md`

## Module Definition Template

```ini
[module/NAME]
type = custom/script
exec = $HOME/.local/share/openriot/install/openriot --flag
format-padding = 1
click-left = $HOME/.local/share/openriot/install/openriot --flag
```

## All Modules (left to right)

```
modules-left = launcher workspaces workspaces2 workspaces3 workspaces4 window-title
modules-center = date
modules-right = crypto night-light cpu memory volume network-wifi network-eth wireguard battery openriot-update power lock
```

## Module Details

| Module | Icon | Click Action | Scroll |
|--------|------|-------------|--------|
| launcher |  | `launcher.sh` | - |
| workspaces | 1-4 | Switch workspace | - |
| window-title | text | Show window name | - |
| date | text | Next wallpaper | - |
| volume | 婢 | Toggle mute | Volume adjust |
| network-wifi | 󰤨/󰤯 | WiFi details | - |
| network-eth | 󰈀/󰌙 | ETH details | - |
| battery |  | - | - |
| crypto |  | crypto-notify.sh | - |
| night-light | /󰌵 | Toggle redshift | - |
| cpu | 龍 | CPU notification | - |
| memory | 펭 | Memory notification | - |
| wireguard | 󰛳/󰅛/󰱓 | Toggle VPN | - |
| openriot-update | 󰟢/󰌷 | Check updates | - |
| power | ⏻ | Power menu | - |
| lock | 󰌾 | Lock screen | - |

## Key Polybar Rules

- **ALL modules-right MUST have `format-padding = 1`**
- Text icons: NO spaces - `"icon"` works, `" icon "` is broken
- Font sizes: T1=10, T2=18
- Date format: `%A • %B %d • %I:%M %p`

---

# 📂 APP LAUNCHER (ROFI)

## apps.txt Format

```
Name|Command|Icon
```

Example:
```
Terminal|alacritty|
Firefox|firefox|󰈹
Telegram|tdesktop.desktop|󰘦
Helix|alacritty -e hx|
Media Player|mpv --player-operation-mode=pseudo-gui|
```

## Desktop Files

Desktop files go to `~/.local/share/applications/`. Commands ending in `.desktop` are executed via:
```bash
grep '^Exec=' ~/.local/share/applications/$CMD | cut -d= -f2-
```

## Launcher Script Flow

1. Read `apps.txt`
2. Filter comments (`#`) and empty lines
3. Build rofi input: `Icon Name`
4. Run rofi with selected theme
5. Get command for selected index
6. Execute command
7. Send "Launching..." notification

---

# 🎨 FASTFETCH CONFIG

## Valid Colors

| Valid | Invalid |
|-------|---------|
| `magenta` | `darkpurple` |
| `cyan` | `lightblue` |
| `green` | `purple` |
| `yellow` | `orange` |
| `blue` | `lightgreen` |
| `white` | |
| `black` | |

## Config Location

`~/.config/fastfetch/config.jsonc`

## Box Drawing

Uses ASCII box-drawing characters for section headers:
```
┌────────────────────Hardware──────────────────────┐
└──────────────────────────────────────────────────┘
```

Key widths must match box width (40 chars inside).

---

# 🔒 LOCK SCREEN

**Script:** `config/bin/openriot-lock.sh`

**Flow:**
1. Find `locked.jpg` in install dir or repo
2. Get screen resolution via `xdpyinfo`
3. Convert JPG to PNG at screen resolution
4. Run `i3lock -i <png>`
5. Delete temp PNG

**i3lock on OpenBSD:**
- Standard i3lock (no color) - uses bsd_auth
- **i3lock-color does NOT work** - requires PAM which OpenBSD doesn't use
- Supports: `-i image.png`, `-c "#rrggbb"`
- Does NOT support: `--circle-color`

---

# 🖥️ XSESSION SETUP

**File:** `config/xsession/openriot-xsession`

## Order Matters

```sh
# 1. Locale
export LC_CTYPE=en_US.UTF-8

# 2. Cursor theme (X11 looks in ~/.icons/, NOT ~/.local/share/icons/)
export XCURSOR_THEME=Bibata-Modern-Ice
export XCURSOR_SIZE=24

# 3. Create dirs
mkdir -p ~/Screenshots

# 4. Load Xresources
[ -f ~/.Xresources ] && xrdb -merge ~/.Xresources

# 5. Disable DPMS
xset -dpms

# 6. Screensaver (10 min)
xset s 600

# 7. xautolock (OpenBSD doesn't have xss-lock)
xautolock -time 10 -locker "$HOME/.local/share/openriot/config/bin/openriot-lock.sh" &
xset s on

# 8. Start i3
exec i3
```

## Screen Lock on OpenBSD

**OpenBSD does NOT have xss-lock** - use `xautolock` instead.

```sh
# Package: xautololock-2.2
xautolock -time 10 -locker "lock-script.sh" &
```

---

# 💡 LESSONS LEARNED

## Process Discipline (CRITICAL)
- **ALWAYS PAUSE AND ASK when confused** - don't run amuck debugging
- Follow the exact workflow: PROPOSE → WAIT FOR CONFIRM → IMPLEMENT → TEST → COPY (only after user tests)
- Don't implement complex solutions when user suggests simpler ones - they're smarter
- If stuck on logic, ask questions instead of implementing wrong approach

## Version Comparison Logic
- `compareVersions(local, remote) < 0` = update available (local < remote)
- `compareVersions(local, remote) <= 0` = local is same or older (wrong!)
- Cache remote version in `~/.cache/openriot/remote.version` for instant click response

## Workspace Switching
- Binary does the switch via `i3-msg` so it knows current before/after
- If `current == target`, exit early (already on that workspace, no-op)
- Don't rely on i3's `workspace_auto_back_and_forth` - control it in binary

## Notifications
- When implementing a feature that needs notifications, ADD notify-send call
- Don't forget: `exec.Command("/usr/local/bin/notify-send", ...).Start()`

## Script Deletion Checklist
- Before deleting scripts, grep polybar config for references
- Update i3 config if it references deleted scripts
- Add cleanup commands to packages.yaml for stale files during upgrades

## WireGuard on OpenBSD
- Config at `/etc/wireguard/wg0.conf` requires root to read
- Use `doas test -f` NOT `os.Stat()` for permission check

## Makefile Changes
- Test Makefile targets after editing - don't break build commands
- Especially `make release` - verify it works before considering done

## fastfetch
- Valid colors: `magenta`, `cyan`, `green`, `yellow`, `blue`, `white`, `black`
- **NOT valid:** `darkpurple`, `lightblue`, etc.
- `mountPoints` may show "/" instead of actual mount point
- Headers need matching width with box-drawing chars

## Memory Reporting
- Use `hw.usermem` NOT `hw.physmem` for accurate user-available percentage
- `hw.physmem` includes reserved kernel memory
- Polybar memory: `--polybar-memory` outputs icon + %
- Memory notification: `GetMemDetails()` shows "X.XX GiB of Y.YY GiB\nTotal Used: Z%"

## Atheros WiFi (athn)
- AR9271 firmware NOT in OpenBSD base (licensing)
- Install via: `fw_update -v athn` or download from `firmware.openbsd.org`
- Error `failed loadfirmware of file athn-open-ar9271 (error 2)` = firmware missing

## Fish Shell Colors
- **ALWAYS omit `#` prefix for hex colors** - fish treats `#` as comment!
- CORRECT: `set_color bb9af7`
- WRONG: `set_color #bb9af7`
- Named colors work fine: `set_color purple`, `set_color cyan`

## Polybar Double Launch (setup.sh)
- Use `pkill polybar` NOT `pkill -HUP polybar`
- `-HUP` only signals reload, but if config changed it may not restart properly
- `pkill` kills and allows clean restart
- **Polybar does NOT auto-start from i3** - must manually restart:
  ```bash
  pkill polybar && nohup polybar 2>/dev/null & disown
  ```

## Polybar Script Paths (CRITICAL)
- **CANONICAL location:** `~/.local/share/openriot/config/polybar/scripts/`
- Config points to: `~/.local/share/openriot/config/polybar/scripts/*.sh`
- **NEVER copy scripts to `~/.config/polybar/scripts/`**
- **NEVER point config to `~/.config/polybar/scripts/`**

**Correct workflow for testing changes:**
1. Edit script in repo: `config/polybar/scripts/workspaces.sh`
2. Copy to canonical: `cp config/polybar/scripts/workspaces.sh ~/.local/share/openriot/config/polybar/scripts/`
3. Restart polybar
4. Test

**If local polybar config has wrong paths (points to `~/.config/polybar/scripts/`):**
1. Fix paths: `sed -i 's|\$HOME/.config/polybar/scripts/|$HOME/.local/share/openriot/config/polybar/scripts/|g' ~/.config/polybar/config.ini`
2. OR copy repo config: `cp config/polybar/config.ini ~/.config/polybar/config.ini`

## Font Names (fontconfig)
- Font names from `fc-list` may differ from filename
- `fc-list | grep Paper` shows: `Paper Mono:style=Regular`
- Using `PaperMono Regular` (filename-style) fails silently

## Polybar Click Coordinates (%lx)
- `%lx` does NOT work for `custom/script` modules - only for `internal/*` modules
- Solution: Pass workspace number directly in config, or use `internal/i3` module
- **OLD (broken):** `click-left = script.sh %lx` - polybar ignores %lx
- **NEW (works):** `click-left = i3-msg workspace 1` - direct i3 command

## notify-send Path
- Always use full path `/usr/local/bin/notify-send` for reliability
- Don't rely on PATH being set correctly in all contexts

## Dunst Restart - MUST Force Kill
- **NEVER use `pkill dunst` (SIGTERM)** - leaves dunst in bad state
- **ALWAYS use `pkill -9 dunst`** - SIGKILL force kills and ensures clean restart
- Stale dunst process causes generic "X" icons even when code is correct

## Git Status Confusion
- Staged files + deleted files can coexist showing "new file" + "deleted"
- Clarify actual file state before making changes
- NEVER run git commands (add, rm, restore, checkout, etc.)

## XSession Setup (CRITICAL)
- Must export `LC_CTYPE=en_US.UTF-8` for proper locale
- Cursor theme: `XCURSOR_THEME=Bibata-Modern-Ice`, `XCURSOR_SIZE=24`
- X11 looks in `~/.icons/` NOT `~/.local/share/icons/` for cursor themes
- DPMS: `xset -dpms` to disable, `xset s 600` for 10min screensaver
- **OpenBSD needs `xautolock`** - xss-lock not available

## Weather Module
- Config at `~/.config/weather.cfg` with `[DEFAULT]` section
- Cache at `~/.cache/openriot-weather.json`
- If no API key in config, weather module hidden automatically
- Polybar format: `weather-font = 10;3` (T1=10, T2=3)

## Proton Drive (rclone)
- Requires `rclone` package + config at `~/.config/rclone/rclone.conf`
- Use `rclone bisync` for two-way sync with local folder
- Example: `rclone bisync Proton:Drive ~/ProtonDrive --progress`

## GTK Theme for Text Editor
- Gnome Text Editor uses GTK3
- Configs: `config/gtk-3.0/gtk.css` + `config/gtk-3.0/settings.ini`
- Deploy to `~/.config/gtk-3.0/`
- Tokyo Night theme works well for dark mode editors

## Polybar Performance
- `exec-interval` controls refresh rate (default often 30s)
- Too frequent updates drain CPU
- `tail -f /dev/null` trick to prevent module from exiting

---

# 🚨 OPENBSD SPECIFIC

| Issue | Solution |
|-------|----------|
| xss-lock | NOT available - use `xautolock` |
| i3lock-color | Doesn't work - requires PAM, OpenBSD uses bsd_auth |
| tar -J | Not supported - use `xz -d` first |
| WireGuard | Use `wg-quick`, NOT manual ifconfig |
| Cursor theme | X11 looks in `~/.icons/`, NOT `~/.local/share/icons/` |
| Bluetooth | NO native support - don't even try |
| NVIDIA | Not supported - use Intel iGPU |
| Secure Boot | Must be disabled in BIOS |
| SATA Mode | Must be AHCI, not RAID |

### WireGuard on OpenBSD

```bash
# Connect
doas wg-quick up /etc/wireguard/wg0.conf

# Disconnect
doas wg-quick down /etc/wireguard/wg0.conf

# Verify
curl https://am.i.mullvad.net/json
```

### WiFi on OpenBSD

Only Intel `iwm` and USB Atheros `athn` work reliably.

```bash
# Check WiFi
ifconfig | grep -E "^iwm"

# Connect via hostname.if
doas vi /etc/hostname.iwn0
# Add: nwid "NetworkName" wpakey "password" dhcp
doas sh /etc/netstart iwn0
```

---

# 🔐 CRYPTO CONFIG

**File:** `~/.config/crypto.toml`

## Format

```toml
api_key = ""  # Optional CoinGecko API key

# Inline array MUST come before inline table
pairs = [
    { sym = "BTC", coin = "bitcoin",   held = 0, entry = 0 },
    { sym = "ZEC", coin = "zcash",     held = 0, entry = 0 },
    { sym = "XMR", coin = "monero",    held = 0, entry = 0 },
    { sym = "LTC", coin = "litecoin",  held = 0, entry = 0 },
]

[indicators]
rsi_period = 14
oversold = 30
overbought = 70

[display]
show_totals = false
```

## Crypto Defaults

- Default coins: BTC, ZEC, XMR, LTC
- CoinGecko IDs: bitcoin, zcash, monero, litecoin
- No API key = rate limited queries

---

# 📋 ICON REFERENCE

See: `docs/Icons.md`

---

# 📋 POLYBAR REFERENCE

See: `docs/polybar-state-architecture.md`, `docs/polybar-performance.md`

---

# 📋 PROTON DRIVE REFERENCE

See: `docs/proton-drive-module.md`

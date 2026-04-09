# OpenRiot — Project Progress

**v0.6** · `https://github.com/CyphrRiot/OpenRiot.git`

**Quick test:** `rm -rf ~/.local/share/openriot ~/.config ~/.cache/openriot && curl -fsSL https://openriot.org/setup.sh | sh`

---

## Active Issues (TESTING NEEDED)

### 1. SUPER+CMD Keybindings (CRITICAL — TESTING IN PROGRESS)

**Status:** CHANGES APPLIED — awaiting user verification

**Changes made:**
- Added `alt:Super` to xkb_options: `ctrl:nocaps,alt:Super`
- Alt key now acts as Super modifier
- Keybindings use `$mod` which is Mod1 (Alt)

**Debug commands:**
```bash
sway -C  # Validate config
ls -la /tmp/sway-ipc.*  # Check SWAYSOCK
xev -event keyboard  # Check key events
```

---

## Recent Changes (April 7, 2026)

### Mirror Fix (April 7, 2026) ✓
- Reverted cloudflare.cdn.openbsd.org → cdn.openbsd.org
- cloudflare.cdn.openbsd.org is invalid

### Network Manager (April 7, 2026) ✓
- Removed neovim (was unused, replaced by Helix)
- Deleted neovim.desktop file
- wifi-selector.sh handles manual network selection via fuzzel
- wifind removed (hangs during install, OpenBSD auto-connects via /etc/hostname.if)

### Fonts (April 7, 2026) ✓
- Added `jetbrains-mono-2.304` to fonts module
- Helps with proper Unicode/emoji rendering

### Fish History (April 7, 2026) ✓
- Deleted disabled `history-fix.fish`
- Added `XDG_DATA_HOME` to config.fish for persistent history

### Sway Config Improvements ✓
- Added `smart_gaps on` and `smart_borders on`
- Added `workspace_auto_back_and_forth yes`
- Fixed Tab keybindings: `focus next/prev`
- Added swayidle auto-lock at 5min, display off at 10min
- Added `exec_always autotiling` (via pip install)
- Added `floating_modifier $mod normal`
- Added blue border colors (client.focused #4c7899)

### Sway Keybindings Cleanup ✓
- Removed duplicate `$mod+W kill`
- Changed `$mod+J split toggle` → `$mod+H split horizontal`, `$mod+V split vertical`
- Changed `$mod+V floating toggle` → `$mod+Z floating toggle`
- Added `$mod+Shift+C reload` and `$mod+Shift+E exit`
- Removed duplicate `$mod+slash` terminal launcher
- Removed duplicate `$mod+space` menu launcher
- Changed `$mod+P layout toggle split` → `layout toggle stacked tabbed split`

### Lock Screen Fixed ✓
- Now uses `openriot-lock.sh` to generate `/tmp/swaylock-bg.png`
- Displays: time, date, crypto prices, user@host, uptime
- Applied to both `$mod+L` and swayidle timeout

### Install Output Clean ✓
- Consistent 1-space indentation for log levels
- Category summaries for config deployment
- Silent command execution

### Terminal: havoc ✓
- Config at `config/havoc/havoc.cfg`
- Version: 0.7.0

### Waybar Scripts: 10 files ✓
- Removed: waybar-cpu.sh, waybar-memory.sh, waybar-temp.sh, wireguard-*.sh, recording-indicator.sh
- Combined CPU+RAM into system-metrics.sh

### Fastfetch ✓
- Logo: openbsd_small

---

## Fixed Issues ✓

| Issue | Fix | Status |
|-------|-----|--------|
| SUPER+CMD keybindings | Added alt:Super xkb option | TESTING |
| Lock screen white | Uses openriot-lock.sh | ✓ |
| Install output verbose | Clean output, silent commands | ✓ |
| Package reinstalling | Use full versions | ✓ |
| smart_gaps/smart_borders | Added to sway config | ✓ |
| swayidle auto-lock | 5min lock, 10min display off | ✓ |
| Tab focus cycling | Fixed to focus next/prev | ✓ |
| Keybinding duplicates | Removed W kill, H/V splits, Z floating | ✓ |
| clipboard history | Skipped (cliphist not in OpenBSD) | — |
| media player controls | Skipped (playerctl not in OpenBSD) | — |
| neovim removed | Replaced with Helix | ✓ |
| wifind removed | Hangs during install | ✓ |
| fish history persist | Added XDG_DATA_HOME | ✓ |
| JetBrains Mono font | Added jetbrains-mono-2.304 | ✓ |
| blue borders | client.focused #4c7899 | ✓ |

---

## All Commits (latest first)

```
6a8b532 Revert cloudflare.cdn.openbsd.org → cdn.openbsd.org
57998e5 Remove wifind (hangs during install)
0e25097 Mirror speed, remove neovim, add wifind, fix fonts
4d59a7f v0.6: Sway config cleanup and improvements
04c3ab5 Clean up install output: fix spacing, remove duplicates
1383d56 Remove unused waybar scripts
124cf44 Fix havoc package version to 0.7.0
```

---

## Waybar Layout

```
Left:   {openbsd-logo} {workspaces} {window title}
Center: {clock} • {weather} {notifications}
Right:  {CPU/RAM} | {volume} | {network} | {battery} | {update} | {⏻} | {🔒}
```

| Position | Module | Backend |
|----------|--------|---------|
| Left | custom/openbsd | fuzzel (app launcher) |
| Left | sway/workspaces | native |
| Left | sway/window | native |
| Center | clock | native |
| Center | custom/weather-emoji | weather-emoji-plain.sh |
| Center | custom/notifications | openriot --notify-waybar |
| Right | custom/system-metrics | system-metrics.sh (CPU+RAM) |
| Right | custom/volume-bar | openriot --volume |
| Right | custom/network | waybar-network |
| Right | custom/battery | waybar-battery.sh |
| Right | custom/openriot-update | openriot-update.sh |
| Right | custom/power | openriot --power-menu |
| Right | custom/lock | swaylock -f |

**RULE:** No separators, no groups, no Hyprland modules, no wireguard, no bluetooth,
no pulseaudio, no wireplumber, no backlight. Only modules that work on OpenBSD.

---

## Platform Info

- **Platform:** OpenBSD 7.9
- **Sway version:** 1.11 with wlroots
- **Terminal:** havoc 0.7.0
- **Shell:** fish 4.6.0
- **Session manager:** seatd

---

## Rules (MUST FOLLOW)

1. **NEVER commit or push without explicit user permission. Show diff first.**
2. **NEVER bump VERSION without explicit user permission.**
3. **ALWAYS run `make` before a commit (even config-only changes).**
4. **Use proposal format: Title / Description / Files / Reason → Continue? [Y/n]**
5. **One issue at a time.**
6. **Waybar layout is APPROVED — do not change without asking.**
7. **Package versions must be full (e.g., havoc-0.7.0 not havoc).**
8. **Keep the pufferfish emoji (🐡) in the fish prompt.**
9. **No Nerd Font icons in waybar custom modules — they break JSON parsing on OpenBSD.**
10. **Verify packages exist at openbsd.app before adding to packages.yaml.**

---

## Debug Commands (on OpenBSD via SSH)

```bash
# Check sway processes
ps aux | grep sway | grep -v grep

# Check if SWAYSOCK exists
ls -la /tmp/sway-ipc.* 2>/dev/null

# Validate sway config
sway -C

# Check deployed config
cat ~/.config/sway/config | head -20

# Check installed packages
pkg_info | grep -E "sway|havoc|waybar"

# Check fastfetch
fastfetch --logo openbsd_small

# Check font loading
fc-list | grep -i "nerd\|paper"
```

---

**Last updated:** Apr 7, 2026 — v0.6, SUPER+CMD testing in progress

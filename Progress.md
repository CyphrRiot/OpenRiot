# OpenRiot — Project Progress

**v0.5** · `https://github.com/CyphrRiot/OpenRiot.git`

**Quick test:** `rm -rf ~/.local/share/openriot ~/.config ~/.cache/openriot && curl -fsSL https://openriot.org/setup.sh | sh`

---

## Recent Changes (April 7, 2026)

### Mirror Change (April 7, 2026) ✓
- Changed cdn.openbsd.org → cloudflare.cdn.openbsd.org
- Applied to: setup.sh, build-iso.sh, README.md

### Network Manager (April 7, 2026) ✓
- Added `wifind-0.7p0` package (TUI wifi manager)
- Removed neovim (was unused, replaced by Helix)
- Deleted neovim.desktop file
- Changed `Mod+N` → `wifind` (was `havoc nvim`)

### Fonts (April 7, 2026) ✓
- Added `jetbrains-mono-2.304` to fonts module
- Helps with proper Unicode/emoji rendering in terminal

### Fish History (April 7, 2026) ✓
- Deleted disabled `history-fix.fish`
- Added `XDG_DATA_HOME` to config.fish for persistent history

### Sway Config Improvements ✓
- Added `smart_gaps on` and `smart_borders on` (cleaner single-window view)
- Added `workspace_auto_back_and_forth yes` (bounce between workspaces)
- Fixed Tab keybindings: now uses `focus next/prev` instead of `focus left/right`
- Added swayidle auto-lock at 5min, display off at 10min
- Added `exec_always autotiling` (via pip install)
- Added `floating_modifier $mod normal` (mouse drag floating windows)

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

### Install Output Now Clean ✓
```
=== OpenRiot v0.5 Setup (OpenBSD 7.9) ===
[INFO] Checking OpenBSD version...
[DONE] OpenBSD 7.9 detected
[INFO] Configuring doas...
[DONE] doas already configured
...
[INFO] Installing packages (1 new, 42 already installed)...
[INFO] Installing havoc-0.7.0...
[DONE] havoc-0.7.0 installed
[DONE] All packages installed
[INFO] Deploying configuration files...
[DONE] fastfetch: 1 files
[DONE] waybar: 10 files
...
[DONE] Configuration deployed
[INFO] Running post-install commands...
[INFO] Running source builds...

+------------------------------------------------------------+
|  OpenRiot v0.5 Installation Complete                       |
+------------------------------------------------------------+
```

### Terminal: havoc (was alacritty)
- Config at `config/havoc/havoc.cfg`
- Havoc version: 0.7.0

### Waybar Scripts: 10 files (was 16)
- Removed: waybar-cpu.sh, waybar-memory.sh, waybar-temp.sh, wireguard-click.sh, wireguard-status.sh, recording-indicator.sh
- Combined CPU+RAM into system-metrics.sh

### Fastfetch: openbsd_small logo
- Fixed config.jsonc "source": "blowfish" → "source": "openbsd_small"

---

## Active Issues

### 1. SUPER+CMD Keybindings (CRITICAL — UNRESOLVED)

**Status:** PARTIALLY RESOLVED — `alt:Super` xkb option added

**Changes made:**
- Added `alt:Super` to xkb_options: `ctrl:nocaps,alt:Super`
- Alt key now acts as Super modifier
- Keybindings use `$mod` which is Mod1 (Alt)

**Still need to verify:**
- Does `sway -C` (config validation) pass?
- Do keybindings work after reboot?
- Is SWAYSOCK created?

---

### 2. Lock Screen Is White (FIXED ✓)

**Status:** FIXED — now uses openriot-lock.sh with crypto, time, date, user@host, uptime

---

## Fixed Issues ✓

| Issue | Fix | Status |
|-------|-----|--------|
| Install output verbose | Clean output: category summaries, silent commands, no [SKIP] spam | ✓ |
| Package reinstalling | Use full versions (havoc-0.7.0 not havoc) | ✓ |
| Duplicate log lines | Removed duplicate "Deploying..." header | ✓ |
| Inconsistent spacing | 1-space for all log levels | ✓ |
| Fastfetch blowfish logo | Changed to openbsd_small | ✓ |
| Terminal (was foot/alacritty) | Switched to havoc | ✓ |
| Unused waybar scripts | Removed 6 scripts | ✓ |
| Screenshot too large | PNG 1.2MB → JPEG 298KB | ✓ |
| Seatd socket permissions | Fixed via rcctl flags | ✓ |
| Fish prompt hanging | Removed __fish_git_prompt | ✓ |
| Lock screen white | Uses openriot-lock.sh with data overlay | ✓ |
| SUPER keybindings | Added alt:Super xkb option | ✓ |
| smart_gaps/smart_borders | Added to sway config | ✓ |
| swayidle auto-lock | 5min lock, 10min display off | ✓ |
| Tab focus cycling | Fixed to focus next/prev | ✓ |
| Keybinding duplicates | Removed W kill, H/V splits, Z floating | ✓ |
| clipboard history | Skipped (cliphist not in OpenBSD) | — |
| media player controls | Skipped (playerctl not in OpenBSD) | — |

---

## All Commits (latest first)

```
04c3ab5 Clean up install output: fix spacing, remove duplicates, silent commands
1383d56 Remove unused waybar scripts (cpu, memory, temp, wireguard, recording)
124cf44 Fix havoc package version to 0.7.0
aa44112 Clean up install output: desc/cmd structure, skip already installed
4de1999 Use openbsd_small logo in fastfetch config.jsonc
493d2ab Compress screenshot: PNG→JPEG, 1.2MB→298KB
215b5bb Reset version to v0.5 and add fastfetch disk/localip modules
701ae96 Refactor package installation to Go binary (v1.6)
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

**Last updated:** Apr 7, 2026 — v0.6, mirror/neovim/wifind cleanup

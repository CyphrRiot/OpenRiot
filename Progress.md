# OpenRiot — Project Progress

**v0.5** · `https://github.com/CyphrRiot/OpenRiot.git`

**Quick test:** `rm -rf ~/.local/share/openriot ~/.config ~/.cache/openriot && curl -fsSL https://openriot.org/setup.sh | sh`

---

## Recent Changes (April 7, 2026)

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

**Status:** UNRESOLVED — this is the #1 blocking issue.

**Symptoms:**
- Sway starts, waybar shows, mouse works
- SUPER+ENTER, SUPER+D, SUPER+Q, Alt+Tab — NONE work
- No SWAYSOCK exists

**Need to investigate:**
- What modifier key does Super actually send on OpenBSD?
- Does `sway -C` (config validation) show errors?
- Could the `include` directive with `~` path be failing silently?

---

### 2. Lock Screen Is White (LOW)

**Status:** Not addressed

Title: Swaylock shows white screen without clock or background

---

## Fixed Issues ✓

| Issue | Fix |
|-------|-----|
| Install output verbose | Clean output: category summaries, silent commands, no [SKIP] spam |
| Package reinstalling | Use full versions (havoc-0.7.0 not havoc) |
| Duplicate log lines | Removed duplicate "Deploying..." header |
| Inconsistent spacing | 1-space for all log levels |
| Fastfetch blowfish logo | Changed to openbsd_small |
| Terminal (was foot/alacritty) | Switched to havoc |
| Unused waybar scripts | Removed 6 scripts |
| Screenshot too large | PNG 1.2MB → JPEG 298KB |
| Seatd socket permissions | Fixed via rcctl flags |
| Fish prompt hanging | Removed __fish_git_prompt |

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
2. **Run `make build` before committing Go source changes.**
3. **Use proposal format: Title / Description / Files / Reason → Continue? [Y/n]**
4. **One issue at a time.**
5. **Waybar layout is APPROVED — do not change without asking.**
6. **Package versions must be full (e.g., havoc-0.7.0 not havoc).**
7. **Keep the pufferfish emoji (🐡) in the fish prompt.**
8. **No Nerd Font icons in waybar custom modules — they break JSON parsing on OpenBSD.**

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

**Last updated:** Apr 7, 2026 — v0.5, install output cleaned up

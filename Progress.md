# OpenRiot — Project Progress

**v0.8 (i3 branch)** · `https://github.com/CyphrRiot/OpenRiot.git`
**Branch:** `i3` — X11/i3 migration (in progress)

**Quick test:** `rm -rf ~/.local/share/openriot ~/.config/fish ~/.config/i3 ~/.config/polybar && curl -fsSL https://openriot.org/setup.sh | sh`

---

## Active Issue: i3 Migration (AUDIT COMPLETE)

### Status: FILE AUDIT COMPLETE — READY FOR TESTING

| Step | Task | Status |
|------|------|--------|
| 1 | Update packages.yaml (sway→i3) | ✅ DONE |
| 2 | Create i3 config directory | ✅ DONE |
| 3 | Create i3/config | ✅ DONE |
| 4 | Create polybar config | ✅ DONE |
| 5 | Create rofi config | ✅ DONE |
| 6 | Delete Xresources (urxvt not installed) | ✅ DONE |
| 7 | File audit (all configs) | ✅ DONE |
| 8 | Lock screen with blur/metrics | ✅ DONE |
| 9 | Build and test | ✅ DONE |

### File Audit Results

| File | Change |
|------|--------|
| `config/bin/openriot-lock.sh` | **Rewritten** — blur background, time/date/crypto/user metrics |
| `config/i3/keybindings.conf` | Lock → `openriot --lock`, rofi-calc, clipmenu |
| `source/main.go` | `--lock` handler generates background first |
| `config/lf/lfrc` | `wayland`→`x11`, `wl-copy`→`xclip` |
| Wayland .desktop files | **Deleted** swaybar, swaybg, swaynag, wl-clipboard, wf-recorder, openriot-screenrecord |

### Why i3?

OpenBSD's wskbd doesn't support XKB options to remap Alt→Super. SUPER keybindings don't fire in Sway/Wayland. i3 on X11 has proven modifier key support.

### Components Selected

| Component | Choice |
|-----------|--------|
| Window Manager | i3 |
| Status Bar | polybar |
| App Launcher | rofi |
| Lock Screen | i3lock |
| Terminal | rxvt-unicode |

## Issues Discovered During April 9, 2026 Session

| Issue | Root Cause | Fix | Status |
|-------|-----------|-----|--------|
| Package check always reinstalls | `pkg_info -e pkg` fails with bare names | Use exact versions (`foot-1.26.1`, `havoc-0.7.0`, `mpv-0.41.0p0`) | FIXED |
| Desktop files not in fuzzel | Shell copy command failed | Added CopyConfigs rule for applications/* | FIXED |
| Foot hangs | `fish_greeting` with fastfetch hangs | Removed fish_greeting | FIXED |
| Fish config overwrites user settings | `config.fish` not in preserve list | Added to preserve list | FIXED |
| conf.d files not preserved | Only `conf.d/*.fish` pattern used | Listed files explicitly | FIXED |

---

## Package Version Fixes (CRITICAL)

OpenBSD's `pkg_add` requires **exact versions** for `pkg_info -e` to work:

| Package | Version |
|---------|---------|
| foot | 1.26.1 |
| havoc | 0.7.0 |
| mpv | 0.41.0p0 |

**Rule:** Always use exact versions in packages.yaml. Bare names cause `pkg_info -e` to fail.

---

## Changes Today (April 9, 2026)

### v0.7: i3 Migration ✓
- Migrated from Sway/Wayland to i3/X11 (Mod4 works on X11)
- Switched terminal: foot remains primary (works on X11 via Xwayland)
- Added rofi for app launching
- Removed openriot-welcome
- Changed SUPER key from Mod1 to Mod4

### Installer Fixes ✓
- Fixed package check: `pkg_info -e pkg` with exact versions
- Desktop files now copied via CopyConfigs (not shell command)
- Fish config.fish now preserved
- Fish conf.d/*.fish files now preserved (custom_commands.fish, ssh-agent.fish)

### Config Fixes ✓
- rofi: replaced fuzzel
- lf: removed invalid `setfiler` command
- Added havoc.desktop to applications (both foot and havoc available)
- foot works on i3 via Xwayland

---

## Recent Commits

```
e9cdad7 fish: remove fastfetch greeting to prevent foot hang
5f8b930 packages: preserve fish conf.d files
de2411b packages: preserve fish config.fish to protect user settings
98b280a packages: use exact version havoc-0.7.0
8eada9b installer: copy desktop files via CopyConfigs instead of shell command
a1d3fbb installer: revert to pkg_info -e (works with exact versions)
bb0e762 installer: copy desktop files via CopyConfigs instead of shell command
```

---

## Polybar Layout

```
Left:   {openbsd-logo} {workspaces} {window title}
Center: {clock} • {weather} {notifications}
Right:  {CPU/RAM} | {volume} | {network} | {battery} | {update} | {⏻} | {🔒}
```

| Position | Module | Backend |
|----------|--------|--------|
| Left | i3/workspaces | native |
| Left | custom/openbsd | rofi (app launcher) |
| Center | clock | native |
| Center | custom/weather-emoji | weather-emoji-plain.sh |
| Right | custom/system-metrics | system-metrics.sh (CPU+RAM) |
| Right | custom/volume | openriot --volume |
| Right | custom/network | network.sh |
| Right | custom/battery | battery.sh |
| Right | custom/openriot-update | openriot-update.sh |
| Right | custom/power | openriot --power-menu |
| Right | custom/lock | i3lock |

---

## Platform Info

- **Platform:** OpenBSD 7.9
- **Window Manager:** i3 (X11)
- **Terminal:** rxvt-unicode 9.31p2
- **Shell:** fish 4.6.0
- **Status Bar:** polybar 3.7.2
- **Launcher:** rofi 1.7.9.1

---

## Desktop Files (23 total)

Location: `config/applications/*.desktop` → `~/.local/share/applications/`

| File | Name | Purpose |
|------|------|---------|
| bluesky.desktop | Bluesky | Web app |
| btop.desktop | Btop System Monitor | System monitor |
| crush.desktop | Crush (AI) | CLI assistant |
| firefox.desktop | Firefox | Browser |
| flare.desktop | Flare Messenger | IM client |
| foot.desktop | Foot Terminal | Wayland terminal |
| fuzzel.desktop | 🔍 Fuzzel | App launcher (Wayland only) |
| havoc.desktop | 🫱 Terminal | Terminal (backup) |
| helix.desktop | Helix | Text editor |
| htop.desktop | Htop System Monitor | System monitor |
| lf.desktop | LF File Manager | File manager |
| tdesktop.desktop | Telegram Desktop | IM client |
| terminal.desktop | 🫱 Terminal | Generic terminal |

---

## Rules (MUST FOLLOW)

1. **NEVER commit or push without explicit user permission. Show diff first.**
2. **NEVER bump VERSION without explicit user permission.**
3. **ALWAYS run `make` before a commit (even config-only changes).**
4. **Use proposal format: Title / Description / Files / Reason → Continue? [Y/n]**
5. **One issue at a time.**
6. **Package versions must be EXACT (e.g., `foot-1.26.1` not `foot`).**
7. **Keep the pufferfish emoji (🐡) in the fish prompt.**
8. **Use Nerd Font icons in polybar custom modules — they work fine on X11.**
9. **Test end-to-end with `curl -fsSL https://openriot.org/setup.sh | sh`.**
10. **Fish config.fish and conf.d files must be preserved.**

---

## Debug Commands (on OpenBSD)

```bash
# Check package versions
pkg_info | grep -E "foot|havoc|mpv"

# Validate i3 config
i3 -C

# Check key events
wev

# Check desktop files
ls ~/.local/share/applications/*.desktop | wc -l

# Check fish config loaded
fish -c "set -S LANG"

# Re-run install
openriot --install

# Clean test
rm -rf ~/.local/share/openriot ~/.config/fish ~/.config/i3 ~/.config/polybar ~/.config/rofi ~/.Xresources
curl -fsSL https://openriot.org/setup.sh | sh
```

---

**Last updated:** Apr 9, 2026 — i3 branch, migration in progress

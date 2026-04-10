# OpenRiot — Project Progress

**v0.8** · `https://github.com/CyphrRiot/OpenRiot.git`
**Branch:** `main` — i3/X11 migration COMPLETE

**Quick test:** `rm -rf ~/.local/share/openriot ~/.config/fish ~/.config/i3 ~/.config/polybar && curl -fsSL https://openriot.org/setup.sh | sh`

---

## Completed: i3/X11 Migration

### What Changed

| Component | Before (Wayland) | After (X11) |
|-----------|------------------|--------------|
| Window Manager | Sway | i3 |
| Status Bar | waybar | polybar |
| App Launcher | fuzzel | rofi (filtered) |
| Lock Screen | hyprlock | i3lock |
| Night Light | wlsunset | redshift |
| Terminal | foot | foot |

### Why i3?

OpenBSD's wskbd doesn't support XKB options to remap Alt→Super. SUPER keybindings don't fire in Sway/Wayland. i3 on X11 has proven Mod4 (Super) key support.

---

## Desktop Applications (14 apps)

| Desktop File | App Name | Polybar Icon | Notes |
|-------------|----------|--------------|-------|
| firefox.desktop | Firefox | 󰈹 | Browser |
| flare.desktop | Flare | 󰇢 | Matrix messenger |
| tdesktop.desktop | Telegram | 󰘦 | Messaging |
| terminal.desktop | Terminal | 󰽒 | Foot emulator |
| helix.desktop | Helix | 󰛞 | Text editor |
| thunar.desktop | File Manager | 󰝰 | Thunar browser |
| btop.desktop | System Monitor | 󰍹 | btop |
| htop.desktop | Htop | 󰍹 | Process viewer |
| crush.desktop | Crush | 󰚩 | AI CLI assistant |
| mpv.desktop | Media Player | 󰕼 | Video player |
| abiword.desktop | Word Processor | 󰈙 | Document editor |
| xfce4-settings.desktop | Settings | 󰒓 | System settings |
| transmission-start.desktop | Transmission | 󰇚 | BitTorrent |
| protonmail.desktop | Proton Mail | 󰊫 | Email web app |

**Removed:** lf, havoc, alacritty, bluesky, fuzzel, foot duplicates, sway (full Wayland stack)

---

## Polybar Workspace Icons

Shows all 4 workspaces with app icons:

```
󰀻  ● 󰽒 󰈹   ○   ◉ 󰝰   ○
```

- `󰀻` = App launcher button (click to open rofi)
- `●` = focused workspace
- `◉` = unfocused with windows
- `○` = empty workspace
- Icons show running apps: 󰽒 foot, 󰈹 firefox, 󰝰 thunar, etc.

### How Workspace Icons Work

1. User opens app → i3 registers window with **window class** (e.g., "firefox")
2. `workspaces.sh` runs every 2 seconds → queries i3 tree → extracts window classes
3. `window-icon.sh` maps class → Nerd Font icon (e.g., "firefox" → 󰈹)
4. Polybar displays: `● 󰈹`

### Rofi vs Polybar Icons

| Source | Icon Type | Example |
|--------|-----------|---------|
| Rofi menu | System icon theme (Adwaita) | `firefox` |
| Polybar | Nerd Font | `󰈹` |

---

## Fish Shell Prompt

```
🐡 ~/Code/OpenRiot (main) →●☡ ❯ 
```

- `🐡` = pufferfish logo
- `(main)` = git branch in yellow
- `→●☡` = git status (staged, dirty, untracked)
- `❯` = prompt

### Git Status Icons

| Icon | Meaning |
|------|---------|
| `●` | Dirty (unsaved changes) |
| `→` | Staged (ready to commit) |
| `☡` | Untracked files |
| `↩` | Stashed |
| `+` | Ahead of remote |
| `-` | Behind remote |

---

## Package Version Rules (CRITICAL)

OpenBSD's `pkg_add` requires **exact versions** for `pkg_info -e` to work:

| Package | Version |
|---------|---------|
| foot | 1.26.1 |
| btop | 1.4.6 |
| htop | 3.4.1 |
| firefox | 149.0 |
| tdesktop | 6.6.4p0 |
| helix | 25.07.1 |
| mpv | 0.41.0p0 |
| redshift | 1.12p11 |

**Rule:** Always use exact versions in packages.yaml. Bare names cause `pkg_info -e` to fail.

---

## Known Issues Fixed

| Issue | Fix |
|-------|-----|
| Package check always reinstalls | Use exact versions in packages.yaml |
| Desktop files not copied | Added CopyConfigs rule for applications/* |
| Foot hangs | Removed fish_greeting |
| Fish config overwrites user | Added to preserve list |
| conf.d files not preserved | Listed files explicitly |
| `pkg_info -e pkg` fails | Use exact version (e.g., `foot-1.26.1`) |

---

## File Audit (Apr 9, 2026)

Completed full codebase audit:

| Category | Files | Status |
|----------|-------|--------|
| Root scripts | 5 | ✅ DONE |
| config/bin/ | 5 | ✅ DONE |
| polybar/scripts/ | 3 | ✅ DONE |
| Source code | 22 Go files | ✅ DONE |
| packages.yaml | 1 | ✅ DONE |

**Key fixes during audit:**
- `build-iso.sh` — Updated Wayland→X11 comments
- `openriot-lock.sh` — hyprlock→i3lock
- wlsunset → redshift (Wayland→X11)
- Added scrot, age packages

---

## AGENTS.md Rules

See `AGENTS.md` for full agent guidelines. Key rules:

1. **NEVER commit or push without explicit user permission**
2. **NEVER bump VERSION without explicit user permission**
3. **Show diff first, wait for confirmation**
4. **Test end-to-end before claiming complete**
5. **One issue at a time**

---

## Debug Commands (on OpenBSD)

```bash
# Check package versions
pkg_info | grep -E "foot|btop|mpv"

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

**Last updated:** Apr 10, 2026 — v0.8, i3/X11 migration complete

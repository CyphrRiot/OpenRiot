# OpenRiot — Project Progress

**v1.1** · commit 654f3f7 · `https://github.com/CyphrRiot/OpenRiot.git`

**Quick test:** `rm -rf ~/.local/share/openriot && curl -fsSL https://openriot.org/setup.sh | sh`

---

## Completed

| Feature | Status | Files |
|---------|--------|-------|
| ISO builder (BIOS+UEFI) | ✅ | `build-iso.sh` |
| Autoinstall config | ✅ | `install/autoinstall.conf` |
| Package installation | ✅ | `source/installer/packages.go` |
| Source builds | ✅ | `install/packages.yaml` (source module) |
| Config deployment | ✅ | `source/installer/configs.go` |
| CLI commands | ✅ | `source/main.go` |
| Sway + Waybar | ✅ | `config/sway/`, `config/waybar/` |
| Fish shell autostart | ✅ | `config/fish/config.fish` |
| Log sharing | ✅ | `--share-log` flag |
| JetBrainsMono Nerd Font | ✅ | `setup.sh`, `config/foot/`, `config/sway/` |

---

## Issues Fixed

| # | Issue | Fix |
|---|-------|-----|
| 1 | NULL byte error from binary | Fallback grep parser in setup.sh |
| 2 | python3-3.11.0 wrong pkg name | Updated to python-3.13.12 |
| 3 | Tree-sitter glob `[` syntax error | `pkg_info -m` only, no glob |
| 4 | Backgrounds to wrong location | Disabled in configs.go |
| 5 | nvim configs overwritten | Added `preserve_if_exists` |
| 6 | No log sharing | Added `--share-log` |
| 7 | setup.sh not executable | chmod +x via git |
| 8 | Misleading debug messages | "New installation..." |
| 9 | Duplicate $mod+F keybinding | Removed duplicate |
| 10 | Sway opacity (unsupported) | Commented out all opacity |
| 11 | Grep parsed commands as packages | `grep -E '^ +- [a-zA-Z]'` + filter quotes |
| 12 | Fish emojis garbled on console | ASCII alternatives |
| 13 | Fastfetch Nerd Font broken | ASCII labels |
| 14 | `pkg_info -m` always fails skip check | Changed to `pkg_info -e` |
| 15 | Grep fallback included YAML commands | Replaced with yq + Python YAML parser |

---

## TODO

### TODO 1 — Sway Autostart on Reboot
**What:** Sway must start automatically on TTY1 after login.

**Flow:** TTY1 login → fish shell → `exec sway` → Sway desktop

**Test:**
```sh
# After reboot, log in at TTY1 — sway should start automatically
sway -d 2>&1 | head -100   # debug if broken
echo $SHELL                  # should be /usr/local/bin/fish
```

**File:** `config/fish/config.fish`
**Symptom if broken:** Black screen or command prompt on TTY1 after login.

---

### TODO 2 — Package Skip Detection
**What:** Re-running setup.sh should skip already-installed packages fast (not re-run pkg_add on each one).

**Test:**
```sh
# Run setup.sh twice. Second run should skip packages with "[SKIP]" not "[INFO] Installing"
tail -20 ~/.cache/openriot/setup.log
```

**Files:** `setup.sh` (lines 248-271)
**Root cause:** Was using `pkg_info -m` (maintainer lookup, always fails). Fixed to `pkg_info -e` (installed check).
**Fix applied:** commit 654f3f7 — needs hardware test verification.

---

### TODO 3 — Nerd Font Rendering in Foot/Waybar ✅ DONE
**What:** JetBrainsMono Nerd Font v3.4.0 installed for glyph/icon rendering in foot, lsd, waybar.

**Installed:** `~/.local/share/fonts/JetBrainsMono/` (downloaded from GitHub, fc-cache updated)

**Configured:** `config/foot/cypherriot.ini` → `font=JetBrainsMono NF:size=11`

**Configured:** `config/sway/config` → `font pango:JetBrainsMono NF 10`

**Test:**
1. Start Sway desktop
2. Open foot terminal
3. Run `fastfetch` — icons should appear
4. Run `lsd -l` — file icons should appear

---

### TODO 4 — Upgrade Flow
**What:** Re-running setup.sh on existing install should git pull updates and rebuild.

**Test:**
```sh
# Modify something, push to git
# Run setup.sh again
# Verify: binary rebuilt, configs updated, packages skipped
```

**File:** `setup.sh` (setup_repository function)

---

### TODO 5 — Source Build Reliability
**What:** wlsunset, crush, bibata-cursor compile without errors.

**Test:**
```sh
which wlsunset
which crush
ls ~/.icons/*/cursors/   # bibata should exist
```

**Dependencies:** Require `system.tools` module (go-1.26.1, meson, ninja)

---

### TODO 6 — Waybar Scripts
**What:** Waybar battery/volume scripts render Nerd Font icons correctly.

**Files:** `config/waybar/scripts/`
**Test:** Check waybar on running Sway desktop.

---

## Debug Commands

```sh
# Share logs for analysis
openriot --share-log

# Manual sway test
sway -d 2>&1 | head -100

# Check fish as default shell
echo $SHELL   # should be fish

# Check installed packages
pkg_info -e   # no args = list all installed
```

---

**Last updated:** Apr 6, 2026

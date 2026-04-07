# OpenRiot — Project Progress

> **OpenRiot** transforms a fresh OpenBSD installation into a fully-configured Sway desktop — in one command.
> It is the OpenBSD counterpart to [ArchRiot](https://archriot.org).

---

## 🚨 CURRENT STATUS (Apr 6, 2026)

**Version:** 1.1 (latest commit: be26283)  
**Testing:** Hardware end-to-end test in progress  
**Git:** `https://github.com/CyphrRiot/OpenRiot.git` (commit: 2429c9d)

### Quick Test Command
```sh
rm -rf ~/.local/share/openriot && curl -fsSL https://openriot.org/setup.sh | sh
```

---

## Installation Flow

```
1. Boot ISO → OpenBSD installer (autoinstall, no interaction needed)
2. install.site runs → configures doas, enables apmd/sndiod
3. REBOOT
4. User logs in, runs: curl -fsSL https://openriot.org/setup.sh | sh
5. setup.sh → packages, source builds, configs, fish shell
6. REBOOT → Sway auto-starts on TTY1
```

---

## ✅ COMPLETED FEATURES

- ISO builder with BIOS+UEFI boot
- Autoinstall configuration
- Package installation (pkg_add)
- Source builds (wlsunset, crush, bibata cursor)
- Config deployment with preserve logic
- CLI commands (--lock, --suspend, --power-menu, --volume, --brightness, etc.)
- Sway Wayland compositor
- Waybar status bar
- Fish shell with auto-start
- Background management

---

## 🔴 ISSUES FIXED (Apr 6, 2026)

| # | Issue | Fix |
|---|-------|-----|
| 1 | NULL byte error from openriot binary | Fallback grep parser in setup.sh |
| 2 | python3-3.11.0 wrong package name | Changed to python-3.11.0 |
| 3 | Tree-sitter glob pattern `[` syntax error | Removed glob, use pkg_info -m only |
| 4 | Backgrounds copying to wrong location | Disabled in configs.go |
| 5 | crypto.toml/nvim configs overwritten | Added preserve_if_exists |
| 6 | Log sharing missing | Added --share-log |
| 7 | setup.sh not executable | chmod +x via git |
| 8 | Misleading debug messages | Fixed to "New installation..." |
| 9 | Duplicate $mod+F keybinding | Removed duplicate |
| 10 | Sway opacity commands (unsupported) | Commented out all opacity |
| 11 | Grep fallback parsed commands as packages | Fixed to `grep -E '^ +- [a-zA-Z]'` |
| 12 | Fish shell emojis garbled on console | Replaced with ASCII ($ instead of λ, etc) |
| 13 | Fastfetch Nerd Font icons broken | Replaced with ASCII labels |

---

## 🔴 REMAINING ISSUES / TODO

### Critical (Blocking)
1. **Sway autostart verification** — Need to confirm sway starts on reboot
   - Flow: TTY1 login → fish → exec sway
   - Requires: fish as default shell, console login (not SSH)

2. **Package skip detection unreliable**
   - `pkg_info -m` check may not work correctly
   - Packages reinstalled on every run unnecessarily

### High Priority
3. **Foot terminal emoji support** — Need to verify emojis render in foot (Wayland)
   - Console uses ASCII (correct behavior)
   - Foot should render emojis with Nerd Font

4. **Waybar scripts with emojis** — Battery/volume scripts use Nerd Font icons
   - Should work in foot/waybar but need verification

### Medium Priority
5. **Source build reliability** — Commands sometimes parsed as packages
   - Fixed grep pattern, needs testing

6. **Upgrade flow** — Need to verify git pull works for updates

---

## 📁 KEY FILES

| File | Purpose |
|------|---------|
| `setup.sh` | Bootstrap: packages, builds, configs, shell setup |
| `install/packages.yaml` | All packages, configs, commands, source builds |
| `config/sway/config` | Main sway config (sources other configs) |
| `config/sway/keybindings.conf` | Keybindings (was duplicate $mod+F) |
| `config/sway/windowrules.conf` | Window rules (removed opacity commands) |
| `config/fish/config.fish` | Fish shell (auto-starts sway on TTY1) |
| `config/fastfetch/config.jsonc` | System info (ASCII icons for console) |
| `source/main.go` | Go binary CLI commands |
| `source/installer/configs.go` | Config deployment (backgrounds disabled) |

---

## 🔧 BUILD COMMANDS

```sh
make build     # Cross-compile for OpenBSD
make dev       # Native build (testing)
make verify    # Smoke test
make iso       # Build bootable ISO
make isotest   # Test ISO in QEMU
```

---

## 📝 SWAY CONFIGURATION NOTES

### What Sway DOES Support
- Standard keybindings (bindsym)
- Window rules (for_window)
- Output/monitor config
- Workspace management
- Floating/split layouts

### What Sway Does NOT Support
- **Per-window opacity** — removed all opacity rules
- Per-window corner radius
- Some Hyprland-specific features

### Sway Autostart Flow
```fish
# config/fish/config.fish
if status is-login && test (tty) = /dev/tty1
    exec sway
end
```

---

## 📦 SOURCE BUILDS (packages.yaml)

| Package | Method |
|---------|--------|
| wlsunset | git clone + meson build |
| crush | go install + move to ~/.local/bin |
| bibata-cursor | curl release + extract to icons |

Dependencies: All require `system.tools` module (includes go-1.26.1)

---

## 🐛 DEBUGGING

### Share logs from OpenBSD:
```sh
~/.local/share/openriot/install/openriot --share-log
# OR
openriot --share-log
```

### Check sway config:
```sh
sway -d 2>&1 | head -100
```

### Verify fish is default:
```sh
echo $SHELL  # Should show fish
```

---

## 📌 SESSION SUMMARY (Apr 6, 2026)

**Working on:** Getting Sway to autostart on reboot

**Root causes fixed:**
1. Duplicate keybinding prevented sway from loading
2. Unsupported opacity commands would cause parse errors
3. Package parser was trying to install YAML commands as packages
4. Console (TTY) can't render emojis — using ASCII instead

**Next steps:**
1. User tests fresh install on OpenBSD hardware
2. Verify sway auto-starts on reboot
3. Confirm emojis render in foot terminal
4. Fix any remaining issues

---

**Last updated:** Apr 6, 2026  
**Git commit:** 2429c9d

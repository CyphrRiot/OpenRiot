# OpenRiot — Project Progress

**v1.4** · `https://github.com/CyphrRiot/OpenRiot.git`

**Quick test:** `rm -rf ~/.local/share/openriot && curl -fsSL https://openriot.org/setup.sh | sh`

---

## Session Context (Apr 7, 2026 — resumed mid-task)

Sway keyboard input and Super+Enter terminal launch broken on OpenBSD. User has been debugging 10+ hours.

### ROOT CAUSE (from sway.log 32531881)

**seatd socket permissions are STILL BROKEN.** The sway log shows:

```
[ERROR] [libseat] Could not connect to socket /var/run/seatd.sock: Permission denied
[INFO] [libseat] Backend 'seatd' failed, skipping
[INFO] [libseat] Seat opened with backend 'noop'
```

The `noop` backend means sway CANNOT properly manage input devices. PS/2 devices (wskbd0, wskbd1, wsmouse0) show up because OpenBSD kernel provides them, but there's no proper input event processing.

**The user previously fixed this manually** (`doas chgrp _seatd /var/run/seatd.sock && doas chmod 770 /var/run/seatd.sock`) but it was reset (likely seatd service restart recreates the socket).

### This explains EVERYTHING:
- Mouse cursor appears (PS/2 device visible) but real input events don't process
- Super+Enter doesn't open terminal (keyboard events not processed)
- `focus_follows_mouse` fix in repo is irrelevant — input doesn't work at all with noop backend
- The sway log shows `focus_follows_mouse yes` (OLD config still deployed), but even with `no` it wouldn't help

### The Fix Needed on OpenBSD

```bash
# Check seatd is running
ps aux | grep seatd

# Fix socket permissions (must persist across reboots)
doas chgrp _seatd /var/run/seatd.sock
doas chmod 770 /var/run/seatd.sock

# Or: ensure grendel is in _seatd group
doas usermod -G _seatd grendel
# Then log out and back in

# Verify
ls -la /var/run/seatd.sock
# Should show: root:_seatd 770
```

**The proper fix is to ensure seatd creates the socket with the right group.** Check `/etc/rc.d/seatd` or `/etc/rc.conf.local` for seatd configuration.

### Waybar Config Issue (FIXED in repo)

The waybar config was `[{...}]` (array wrapper). waybar 0.15.0 expects `{...}` (plain object).
Also changed `$HOME` to `~` in include paths (waybar doesn't expand $HOME in JSON).

**Fixed:** Changed `config/waybar/config` from array to object, `$HOME` to `~` in includes.

### Other Issues from Sway Log

- **swaybg**: `Could not find config for output HDMI-A-1` — needs `output * bg ...` or specific swaybg config
- **Many "Permission denied" on /dev/wskbd*, /dev/wsmouse*** — all due to noop backend

---

## Files Modified (keyboard focus / sway / waybar fixes)

| File | Changes | Commit |
|------|---------|--------|
| `config/waybar/config` | Removed `//` comments | 34c68be |
| `config/waybar/Modules` | Removed trailing commas | 351bcf1 |
| `config/waybar/ModulesCustom` | Removed trailing commas | 351bcf1 |
| `config/waybar/ModulesVertical` | Removed trailing commas, converted `//` to `/* */` | 351bcf1 |
| `config/waybar/ModulesWorkspaces` | Removed `//` section headers | 34c68be |
| `config/waybar/UserModules` | Replaced `//` commented blocks with empty object | 34c68be |
| `config/waybar/UserWorkspaces` | Removed `//` section headers, fixed trailing comma | 34c68be, 351bcf1 |
| `config/waybar/config` | Array→Object, `$HOME`→`~` in includes | NOT COMMITTED |
| `config/sway/config` | Added 1920x1080 output, keyboard/pointer input, `focus_follows_mouse no` | 2b47d29, 351bcf1, 3bf12d4 |

---

## Git Commits (latest sway/waybar fixes)

```
3bf12d4 Fix keyboard focus: disable focus_follows_mouse
351bcf1 Add 1920x1080 output config, fix waybar trailing commas
2b47d29 Add keyboard and pointer input configuration for OpenBSD
34c68be Remove // style comments from waybar JSON configs
```

---

## Current Todo List

- [completed] Fix waybar JSON: remove invalid `//` comments
- [completed] Fix waybar JSON: remove trailing commas
- [completed] Fix waybar config: change array [{...}] to object {...}
- [completed] Fix waybar includes: $HOME → ~
- [completed] Add sway output config for 1920x1080
- [completed] Add sway keyboard/pointer input configuration
- [completed] Fix focus_follows_mouse (changed to no)
- [pending] COMMIT waybar config fix (user permission needed)
- [in_progress] Fix seatd socket permissions on OpenBSD (the REAL root cause)
- [pending] Install quirks package — fixes shutdown crash
- [pending] Fix swaybg config for HDMI-A-1

---

## Remaining Issues

1. **seatd socket permissions** — THE ROOT CAUSE. Socket must be `root:_seatd` 770. Currently `root:wheel`. This breaks ALL input on sway.

2. **Waybar config** — Fixed in repo (array→object, $HOME→~). Needs commit + deploy.

3. **swaybg** — `Could not find config for output HDMI-A-1`. Needs background config.

4. **Shutdown crash** — `quirks_context_unref` undefined symbol. Needs `doas pkg_add quirks`.

---

## Platform Info

- **Platform:** OpenBSD 7.9 on mini.openriot.org (AMD64, Intel ADL-N GPU)
- **Sway version:** 1.11 with wlroots 0.19.2
- **Terminal:** foot (Wayland native)
- **Session manager:** seatd (BROKEN — noop fallback)
- **Waybar:** 0.15.0p0

---

## Critical Rule

**NEVER commit or push without explicit user permission.** Always show diffs first.

---

**Last updated:** Apr 7, 2026

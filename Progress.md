# OpenRiot — Project Progress

**v1.5** · `https://github.com/CyphrRiot/OpenRiot.git`

**Quick test:** `rm -rf ~/.local/share/openriot ~/.config ~/.cache/openriot && curl -fsSL https://openriot.org/setup.sh | sh`

---

## CRITICAL: SUPER+CMD Keybindings Don't Work

**Status:** UNRESOLVED — this is the #1 blocking issue.

**Symptoms:**
- Sway starts, waybar shows, mouse works (can click waybar icons)
- Can type in fuzzel (when opened by clicking waybar icon)
- Can type in Firefox
- SUPER+ENTER, SUPER+D, SUPER+Q, Alt+Tab — NONE work
- No SWAYSOCK exists (`/tmp/grendel-runtime/sway-ipc.*` empty)
- `focus_follows_mouse no` confirmed in deployed config
- `sway -d 2>/tmp/sway.log` captures waybar stderr, not sway debug
- keybindings.conf was loaded in earlier logs ("Bound Mod4+Return")

**Tried and failed:**
- Fixed `focus_follows_mouse` to `no`
- Simplified sway config to minimal keybindings
- Fixed seatd socket permissions (`rcctl set seatd flags '-g _seatd'`)
- Cleaned keybindings.conf (removed Hyprland leftovers)
- Added `include ~/.config/sway/windowrules.conf`
- Verified 3 sway processes is normal (compositor + helpers)

**Need to investigate:**
- What modifier key does Super actually send on OpenBSD?
- Is there a sway IPC socket at all? (SWAYSOCK missing)
- Does `sway -C` (config validation) show errors?
- Could the `include` directive with `~` path be failing silently?

---

## Active Issues — Priority Order

### 1. SUPER+CMD Keybindings (CRITICAL — BLOCKING)
```
Title: SUPER+{CMD} keybindings don't work in Sway
Description: No Super key combination triggers sway commands. Mouse works, 
             keyboard works in fuzzel/Firefox, but sway keybindings don't 
             intercept SUPER key events.
Files: config/sway/config, config/sway/keybindings.conf
Status: UNRESOLVED
```

### 2. Terminal Doesn't Accept Input (HIGH)
```
Title: Terminal (was foot, now alacritty) can't receive keyboard input
Description: Foot showed "fastfetch then nothing" — no cursor, no prompt. 
             Switched to alacritty. Not yet tested since alacritty isn't 
             installed (needs setup.sh run).
Files: config/sway/config ($terminal alacritty), install/packages.yaml 
       (added alacritty), config/alacritty/alacritty.toml (new)
Status: Awaiting test after fresh install
```

### 3. Waybar Clock Click Changes Format AND Background (MEDIUM)
```
Title: Clicking clock changes date format AND background
Description: User wants left-click to ONLY change background. Clock module 
             was cycling between format and format-alt on click.
Files: config/waybar/config (removed format-alt, removed calendar sub-object)
Status: Fixed in repo — awaiting test
```

### 4. Waybar Lock Icon Shows Yellow (LOW)
```
Title: Lock icon in waybar shows yellow instead of purple
Description: CSS defines @lock_color (#9c7ce8) but icon not rendering 
             correctly. Nerd Font glyph may not load.
Files: config/waybar/config, config/waybar/style.css, config/waybar/colors.css
Status: Need to check font loading (fc-list | grep nerd)
```

### 5. Fuzzel Shows Too Many Items (MEDIUM)
```
Title: Fuzzel shows system desktop files instead of 15 curated apps
Description: System packages install desktop files to /usr/local/share/applications/. 
             Fuzzel shows all of them. Created NoDisplay overrides for 
             foot-server, footclient, swaybg, swaynag, swaybar, wl-clipboard, 
             wf-recorder.
Files: config/applications/foot-server.desktop, footclient.desktop, 
       swaybg.desktop, swaynag.desktop, swaybar.desktop, wl-clipboard.desktop, 
       wf-recorder.desktop (all with NoDisplay=true)
Status: Committed — need to verify overrides deploy to ~/.local/share/applications/
```

### 6. Lock Screen Is White (LOW)
```
Title: Swaylock shows white screen without clock or background
Description: swaylock -f with no image config shows white. Need to either 
             pass -i image or create swaylock config. openriot-lock.sh exists 
             but isn't called from keybinding.
Files: config/sway/keybindings.conf (bindsym $mod+L), config/bin/openriot-lock.sh
Status: Not yet addressed — need to wire openriot-lock.sh into swaylock binding
```

### 7. Seatd Socket Permissions (FIXED)
```
Title: Seatd socket resets to root:wheel on restart
Description: Socket owned by root:wheel instead of root:_seatd. Fixed 
             permanently via rcctl flags.
Files: install/packages.yaml (added rcctl set seatd flags '-g _seatd')
Status: FIXED
```

### 8. Fastfetch Display Issues (FIXED)
```
Title: Fastfetch shows broken modules on OpenBSD
Description: Packages module timed out, DE module empty, icons showed "kora", 
             disk showed unlabeled devices, RAM had no emoji key.
Files: config/fastfetch/config.jsonc
Status: FIXED — removed broken modules, simplified keys, added uptime
```

### 9. Fish History Corruption (FIXED)
```
Title: "ignoring corrupted history entry" on shell startup
Description: Fish history entries missing blank line separators. Python 
             repair script was too slow for every startup — disabled.
Files: config/fish/conf.d/history-fix.fish (disabled with exit 0)
Status: FIXED — disabled auto-repair, manual fix needed if corruption recurs
```

### 10. Fish Prompt Hanging (FIXED)
```
Title: Fish shell potentially hanging after fastfetch in foot
Description: Removed __fish_git_prompt (ran git status on every prompt). 
             Removed echo from custom_commands loader. Removed Python 
             history-fix from startup.
Files: config/fish/config.fish, config/fish/conf.d/custom_commands.fish, 
       config/fish/conf.d/history-fix.fish
Status: FIXED — fish startup is now minimal
```

---

## All Commits (latest first)

```
bc55e6d Fix fastfetch config for OpenBSD — remove broken modules
c52f18b Fix desktop files foot→alacritty, add helix, add alacritty config
68c066c Fix clock "argument not found" error
ec84761 Switch terminal from foot to alacritty, fix fuzzel entries
ff32a13 Fix fish startup, waybar scripts, clock, and icons
f9876bc Fix seatd socket permissions permanently via rcctl
8c35d75 Fix sway config with variables, autostart, keyboard settings
a190287 Clean waybar config to OpenBSD-compatible modules only
074cc74 Flatten waybar config into single file
d5fdce1 Fix waybar config format and seatd socket permissions
71b6ec4 Use shallow clone (--depth 1) to skip 484MB git history
503887d Skip curl install in bootstrap
380dd5c Add fish history corruption auto-repair on shell startup
24cc342 Bump version to v1.5
```

---

## Waybar Layout (APPROVED — do not change)

```
Left:   {openbsd-logo} {workspaces} {window title}
Center: {clock} {weather} {notifications}
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

- **Platform:** OpenBSD 7.9 on mini.openriot.org (AMD64, Intel ADL-N GPU)
- **Sway version:** 1.11 with wlroots 0.19.2
- **Terminal:** alacritty (replaced foot)
- **Session manager:** seatd (fixed with `-g _seatd` flag)
- **Waybar:** 0.15.0p0
- **Shell:** fish

---

## Rules (MUST FOLLOW)

1. **NEVER commit or push without explicit user permission. Show diff first.**
2. **Run `make build` before committing Go source changes.**
3. **Use proposal format: Title / Description / Files / Reason → Continue? [Y/n]**
4. **One issue at a time.**
5. **Waybar layout is APPROVED — do not change without asking.**
6. **Keep the pufferfish emoji (🐡) in the fish prompt.**
7. **No Nerd Font icons in waybar custom modules — they break JSON parsing on OpenBSD.**

---

## Debug Commands (on OpenBSD via SSH)

```bash
# Check seatd socket
ls -la /var/run/seatd.sock

# Check sway processes
ps aux | grep sway | grep -v grep

# Check deployed config
cat ~/.config/sway/config | head -10

# Check if SWAYSOCK exists (may be in /tmp/grendel-runtime/ on OpenBSD)
ls -la /tmp/grendel-runtime/sway-ipc.* 2>/dev/null
ls -la /tmp/sway-ipc.* 2>/dev/null

# Check fuzzel desktop files
ls ~/.local/share/applications/

# Check installed packages
pkg_info -e alacritty
pkg_info -e foot

# Check font loading
fc-list | grep -i "nerd\|paper"

# Check fastfetch
fastfetch --logo-width 20
```

---

## What a New Session Needs to Know

1. **SUPER+CMD is the #1 unsolved issue.** Everything else is secondary.
2. Terminal was switched from foot to alacritty (needs testing).
3. Waybar errors are mostly fixed (clock, battery, update scripts).
4. Fuzzel needs NoDisplay overrides verified on deploy.
5. Lock screen needs openriot-lock.sh wired up.
6. The user is frustrated — follow the format, one issue at a time, never commit without permission.

---

**Last updated:** Apr 7, 2026 — Active debugging session, fresh install in progress

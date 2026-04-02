# OpenRiot — Project TODO & Progress

> **OpenRiot** transforms a fresh OpenBSD installation into a fully-configured Sway desktop — in one command.
> It is the OpenBSD counterpart to [ArchRiot](https://archriot.org).

---

## What Is OpenRiot?

OpenRiot takes a base OpenBSD 7.9 install and layers on:

- **Sway** — i3-compatible Wayland compositor
- **Waybar** — status bar with fully OpenBSD-native modules
- **Fish** shell with git prompts
- **Neovim** with LazyVim + optional OpenRouter/avante LLM support
- **Foot** terminal emulator
- **Fuzzel** app launcher (same as ArchRiot)
- **Thunar** file manager

All configuration is declarative, version-controlled, and reproducible.

---

## Workflow Rules — NEVER DEVIATE

### On Every Change

**Every response MUST start with:**

```
Completed: <brief description of what was just done>
Next Task: <description of next task>
```

**Then provide:**

```
Files: <list of files to be modified>
Goal: <why this change is being made>
```

**Then ask:** `Continue?`

1. **NEVER COMMIT** — Do NOT run `git commit` or `git push` without explicit permission
2. **NEVER RUN THE openriot binary** -- always ask the user to run and provide feedback
3. **Propose first** — Show the exact change (filename, function, reason) before editing
4. **Wait for "Proceed/Continue?"** before touching any code
5. **Test locally first** (Linux with `--test` flag where applicable)
6. **Verify build passes** (`make build`) after any Go changes
7. **Show proof it works** before asking for approval
8. **One change at a time** — finish one task before starting another

### Before Starting a New Chat

1. Read this entire TODO top to bottom
2. Run `git status` to check for uncommitted changes
3. Run `make build` to confirm the binary builds cleanly
4. Run `make dev && ./install/openriot --version` to verify native build works
5. Start from the first item marked 🔴 NOT DONE

### Build Commands

- `make build` — Cross-compile for OpenBSD amd64 → install/openriot
- `make dev` — Native build for local testing on Linux
- `make verify` — Build + smoke test (runs --version)
- `make iso` — Full ISO build (downloads packages + builds + repacks)
- `make download-packages` — Download packages to ~/.pkgcache/7.9/amd64/
- `make clean` — Remove build artifacts

### Version Bumping (when releasing)

1. Confirm version in `Makefile` (`OPENRIOT_VERSION`)
2. Run `make build` — verify it passes
3. Update README.md badge if needed
4. `git commit -am "Release vX.X: [brief changes]"`
5. `git tag -a vX.X -m "Version X.X release: [details]"`
6. `git push origin master && git push origin vX.X`

---

## Architecture: Three Layers

```
LAYER 1: ISO Builder (build-iso.sh)
  Produces a bootable OpenBSD ISO with offline packages + openriot binary bundled.

LAYER 2: Go Installer (openriot binary)
  Runs on the installed system. Installs packages, deploys configs, handles CLI flags.
  Entry point: source/main.go
  Triggered by: openriot --install (called from .profile on first login)

LAYER 3: First Boot (install/setup.sh)
  Shell script for users who already have OpenBSD installed (no ISO).
  Usage: curl -fsSL https://openriot.org/setup.sh | sh
```

### Install Flow (ISO path)

```
1. Boot ISO
2. autoinstall/install.conf answers all OpenBSD installer prompts
3. install.site runs post-install:
   - Mounts CD, installs packages offline via pkg_add
   - Copies openriot binary to /usr/local/bin/
   - Configures doas, enables apmd + sndiod
   - Sets fish as default shell
   - Adds .profile hook: runs openriot --install on first login
4. User logs in → openriot --install runs:
   - Deploys all config files
   - Runs commands from packages.yaml
   - Builds wlsunset from source
   - Prompts for git config and OpenRouter API key
```

## Workflow

Now, focus on the workflow:

1. `make iso` builds the ISO
2. You install the ISO, which should contain all packages (from install/packages.yaml) and a copy of the git repo that gets added to ~/.local/share/OpenRiot
3. After first boot, it should run `openriot` and copy files to the right places for Sway and all of the Window Manager
4. Everything should work after a reboot -- it should launch Sway and have a Waybar with everything working
5. Periodically, it checks VERSION and, if a greater version exists, runs `curl -fsSL https://OpenRiot.org/setup.sh | bash` in a terminal window (see /home/grendel/Code/ArchRiot for a WORKING example on Linux) and should update the system properly
6. You **have to reference VERSION** for the version and stop hard-coding it like a junior dev.

**Read TODO.md for info. Reference /home/grendel/Code/ArchRiot/source for the tui installer flow**

**Always update the TODO.md after a task is confirmed completed**

- Make sure to add proper .config files
- Confirm nothing is missing
- Fully audit everything for ANY issues
- If any issue exists, follow the TODO.md requirements and PROPOSE a fix

**It is critically important that the hotkeys, fuzzel, apps, and waybar all function properly for OpenBSD**

---

## Canonical Versions (single source of truth: Makefile)

```
OPENRIOT_VERSION = 0.4
OPENBSD_VERSION  = 7.9
ARCH             = amd64
```

**Never hardcode these anywhere else. Read from Makefile or environment.**

---

## Key Files Reference

| File                                | Purpose                                                      |
| ----------------------------------- | ------------------------------------------------------------ |
| `Makefile`                          | All versions, build targets                                  |
| `build-iso.sh`                      | Builds bootable ISO                                          |
| `scripts/download-packages.sh`      | Downloads packages for offline ISO                           |
| `scripts/generate-index.sh`         | Generates index.txt for pkg_add                              |
| `autoinstall/install.conf`          | Autoinstall answers for OpenBSD installer                    |
| `autoinstall/install.site`          | Post-install script (runs from site79.tgz)                   |
| `install/packages.yaml`             | **Source of truth** for all packages, configs, commands      |
| `install/setup.sh`                  | Curl-pipe bootstrap for existing OpenBSD installs            |
| `source/main.go`                    | Go binary entry point, all CLI flags                         |
| `source/audio/volume.go`            | `--volume` via sndioctl                                      |
| `source/display/display.go`         | `--brightness` via wsconsctl                                 |
| `source/backgrounds/backgrounds.go` | `--swaybg-next` wallpaper cycling                            |
| `source/detect/detect.go`           | `--suspend-if-undocked`                                      |
| `source/mullvad/mullvad.go`         | `--mullvad-setup` WireGuard walkthrough                      |
| `source/session/session.go`         | `--lock`, `--suspend`, `--power-menu`                        |
| `source/windows/windows.go`         | `--fix-offscreen-windows`                                    |
| `source/installer/packages.go`      | Package installation via pkg_add                             |
| `source/installer/configs.go`       | Config file deployment                                       |
| `source/installer/execcommands.go`  | Command execution from YAML                                  |
| `source/installer/sourcebuilds.go`  | Source builds (wlsunset)                                     |
| `source/tui/model.go`               | BubbleTea TUI model                                          |
| `source/logger/logger.go`           | Unified TUI + stdout logger                                  |
| `config/sway/config`                | Sway compositor config                                       |
| `config/sway/keybindings.conf`      | All keybindings                                              |
| `config/sway/swaylock-wrapper.py`   | Python swaylock screen generator                             |
| `config/sway/swayidle.conf`         | Idle timeout reference (active config inline in sway/config) |
| `config/waybar/config`              | Active waybar bar layout                                     |
| `config/waybar/Modules`             | Core waybar module definitions                               |
| `config/waybar/ModulesCustom`       | Custom OpenRiot module definitions                           |
| `config/waybar/ModulesGroups`       | Module group definitions                                     |
| `config/waybar/scripts/`            | All waybar shell scripts                                     |
| `config/fish/`                      | Fish shell configuration                                     |
| `config/nvim/`                      | Neovim/LazyVim configuration                                 |
| `config/foot/`                      | Foot terminal config                                         |
| `config/mako/`                      | Mako notification daemon config                              |
| `config/fuzzel/fuzzel.ini`          | Fuzzel launcher config                                       |
| `backgrounds/`                      | 16 CypherRiot wallpapers                                     |

---

## OpenBSD Package Reference

### Installed via pkg_add (from packages.yaml — source of truth)

```
# Core Base
git rsync bc-gh python fastfetch

# Shell & Terminal
fish neovim foot fzf ripgrep htop tree fd

# Sway Desktop
sway waybar fuzzel swaylock swayidle swaybg grim

# Applications
thunar thunar-archive firefox flare-messenger tdesktop

# System Tools
curl wget unzip xz ninja meson
```

### Source-Built (not in pkg_add)

| Package  | Method                                                                                 |
| -------- | -------------------------------------------------------------------------------------- |
| wlsunset | Bundled as wlsunset.tar.gz in site79.tgz; built with meson during `openriot --install` |

### OpenBSD-Specific Tool Replacements

| ArchRiot Tool       | OpenBSD Replacement                             | Notes                             |
| ------------------- | ----------------------------------------------- | --------------------------------- |
| `brightnessctl`     | `wsconsctl display.brightness`                  | Console brightness only           |
| `pamixer` / `pactl` | `sndioctl`                                      | OpenBSD native audio              |
| `playerctl`         | N/A                                             | Not available; sndio has no MPRIS |
| `pavucontrol`       | N/A                                             | PulseAudio only                   |
| `systemd` timers    | `crontab` or wrapper scripts                    |                                   |
| `loginctl lock`     | `swaylock -f`                                   |                                   |
| `systemctl suspend` | `zzz`                                           |                                   |
| `NetworkManager`    | `ifconfig` + `hostname.if`                      |                                   |
| `kanshi`            | static `monitors.conf`                          | No hotplug daemon on OpenBSD      |
| `apm` (battery)     | `apm -l` (%), `apm -a` (AC), `apm -m` (minutes) |                                   |
| `wofi`              | `fuzzel`                                        | Fuzzel IS available on OpenBSD    |

### DO NOT PORT (no OpenBSD equivalent)

- `fcitx5` — input method, not on OpenBSD
- `blueberry` — Bluetooth GUI, OpenBSD has no BT stack
- `udiskie` — Thunar handles mounts via gvfs
- `thermald`, `tlp` — apmd handles power
- `mullvad` app — use WireGuard directly (`openriot --mullvad-setup`)
- `xdg-desktop-portal-wlr` — not in OpenBSD packages
- `wl-clipboard` — not in OpenBSD packages
- `kanshi` — not in OpenBSD packages

---

## Current Status

### ✅ COMPLETED

| #       | Component                      | File(s)                                                                      | Notes                                                 |
| ------- | ------------------------------ | ---------------------------------------------------------------------------- | ----------------------------------------------------- |
| 1.1     | ISO builder script             | `build-iso.sh`                                                               | Linux-compatible, xorriso-only, El Torito BIOS+UEFI   |
| 1.2     | Offline package download       | `scripts/download-packages.sh`                                               | POSIX awk, wayland/ fallback, dry-run flag            |
| 1.2     | Index generation               | `scripts/generate-index.sh`                                                  | Auto-runs after download                              |
| 1.3     | install.site                   | `autoinstall/install.site`                                                   | Mounts CD, pkg_add offline, doas, fish, .profile hook |
| 1.3     | autoinstall config             | `autoinstall/install.conf`                                                   | Unattended OpenBSD install                            |
| —       | Canonical versioning           | `Makefile`                                                                   | OPENRIOT_VERSION=0.4, OPENBSD_VERSION=7.9             |
| 2.1     | TUI test mode                  | `source/tui/`                                                                | BubbleTea, no deadlock                                |
| 2.2     | Config deployment              | `source/installer/configs.go`                                                | Glob patterns, backgrounds                            |
| 2.3     | Command execution              | `source/installer/execcommands.go`                                           | Dry-run flag                                          |
| 2.4     | Package installation           | `source/installer/packages.go`                                               | pkg_add wired                                         |
| 2.5     | Source builds                  | `source/installer/sourcebuilds.go`                                           | wlsunset tarball + git fallback                       |
| 2.7     | CLI: `--volume`                | `source/audio/volume.go`                                                     | sndioctl, toggle/inc/dec/mic                          |
| 2.7     | CLI: `--brightness`            | `source/display/display.go`                                                  | wsconsctl, up/down/set/get                            |
| 2.7     | CLI: `--lock`                  | `source/main.go`                                                             | swaylock -f                                           |
| 2.7     | CLI: `--suspend`               | `source/main.go`                                                             | zzz                                                   |
| 2.7     | CLI: `--power-menu`            | `source/main.go`                                                             | fuzzel --dmenu                                        |
| 2.7     | CLI: `--swaybg-next`           | `source/backgrounds/backgrounds.go`                                          | Cycles wallpapers                                     |
| 2.7     | CLI: `--fix-offscreen-windows` | `source/windows/windows.go`                                                  | swaymsg workspace cycling                             |
| 2.7     | CLI: `--suspend-if-undocked`   | `source/detect/detect.go`                                                    | sysctl acpibat0                                       |
| 2.7     | CLI: `--mullvad-setup`         | `source/mullvad/mullvad.go`                                                  | WireGuard walkthrough                                 |
| 2.9     | Waybar: weather                | `config/waybar/scripts/weather-emoji-plain.sh`                               | Wired in ModulesCustom                                |
| 2.9     | Waybar: network                | `config/waybar/scripts/waybar-network`                                       | OpenBSD ifconfig                                      |
| 2.9     | Waybar: WireGuard              | `config/waybar/scripts/wireguard-status.sh`                                  | wg + ifconfig                                         |
| 2.9     | Waybar: recording              | `config/waybar/scripts/recording-indicator.sh`                               | wf-recorder detection                                 |
| 2.9     | Waybar: updater                | `config/waybar/scripts/openriot-update.sh`                                   | GitHub VERSION check                                  |
| 2.9     | Waybar: WiFi selector          | `config/waybar/scripts/wifi-selector.sh`                                     | fuzzel --dmenu + ifconfig                             |
| 2.9     | Waybar: volume bar             | `config/waybar/scripts/waybar-volume.sh`                                     | sndioctl JSON output                                  |
| 2.9     | Waybar: CPU                    | `config/waybar/scripts/waybar-cpu.sh`                                        | top(1) JSON output                                    |
| 2.9     | Waybar: temperature            | `config/waybar/scripts/waybar-temp.sh`                                       | sysctl hw.sensors JSON                                |
| 2.9     | Waybar: memory                 | `config/waybar/scripts/waybar-memory.sh`                                     | vmstat + sysctl JSON                                  |
| 2.9     | Waybar: battery                | `config/waybar/scripts/waybar-battery.sh`                                    | apm(8) JSON output                                    |
| 2.12    | Waybar cleanup                 | `ModulesCustom`, `config`                                                    | All broken Linux modules fixed                        |
| 2.12    | wofi → fuzzel                  | `packages.yaml`, `install.site`, `setup.sh`, `wifi-selector.sh`, `README.md` | Full replacement                                      |
| 2.12    | custom/lock                    | `ModulesCustom`                                                              | hyprlock → swaylock -f                                |
| 2.12    | custom/arch                    | `ModulesCustom`                                                              | 󰀻 icon, nwg-drawer → fuzzel                           |
| 2.12    | custom/battery                 | `ModulesCustom`, `waybar/config`                                             | Built-in battery → custom/battery via apm             |
| 2.10.4a | swaylock: time/date/user       | `config/sway/swaylock-wrapper.py`                                            | Python + PIL, works on OpenBSD                        |
| —       | Sway config                    | `config/sway/config`                                                         | Ported from ArchRiot                                  |
| —       | Waybar config                  | `config/waybar/config`                                                       | All modules OpenBSD-native                            |
| —       | Fish config                    | `config/fish/`                                                               | Ported from ArchRiot                                  |
| —       | Neovim config                  | `config/nvim/`                                                               | LazyVim + avante                                      |
| —       | Foot config                    | `config/foot/`                                                               | Ported from ArchRiot                                  |
| —       | Fuzzel config                  | `config/fuzzel/fuzzel.ini`                                                   | Tokyo Night theme                                     |
| —       | Mako config                    | `config/mako/`                                                               | Notification daemon                                   |
| —       | Backgrounds                    | `backgrounds/`                                                               | 16 CypherRiot wallpapers                              |
| 3.1     | setup.sh exists                | `install/setup.sh`                                                           | Has bugs — see 3.1 below                              |

---

## NEXT STEPS — DO THESE IN ORDER

**Priority key: 🔴 P0 (blocking) | 🟠 P1 (important) | 🟡 P2 (polish)**

---

### STEP 1 — Build Verification 🔴 P0

Before doing anything else, verify the binary builds cleanly after all recent changes.

- [x] **1.1** Run `make build` — must succeed with zero errors
- [x] **1.2** Run `make verify` — runs `--version` smoke test
- [x] **1.3** Run `./install/openriot --test` on Linux — TUI must launch without deadlock
- [x] **1.4** If build fails: check `source/main.go` for any missing imports or broken flag handlers

---

### STEP 2 — ISO Test on Real Hardware 🔴 P0

**Context:** The ISO has been built (1.1G) but never booted. This is the critical end-to-end test.
**Do NOT use QEMU — test on real ThinkPad or compatible hardware (see README Supported Systems).**

- [ ] **2.1** Run `make iso` on Linux host — must complete without error
    - Downloads packages to `~/.pkgcache/7.9/amd64/`
    - Builds openriot binary (cross-compiled for OpenBSD amd64)
    - Repacks ISO to `isos/openriot-0.6.iso`
- [ ] **2.2** Verify `~/.pkgcache/7.9/amd64/index.txt` exists and has entries
- [ ] **2.3** Verify `isos/openriot-0.6.iso` exists and is larger than 762MB (base size)
- [ ] **2.4** Boot ISO on real hardware — confirm OpenBSD installer starts
- [ ] **2.5** Confirm autoinstall runs unattended (no keyboard input needed)
- [ ] **2.6** After install completes, check `/tmp/install.site.log` for errors
- [ ] **2.7** Confirm packages installed from CD (disconnect network cable, retest)
- [ ] **2.8** Log in as created user — confirm `.profile` hook triggers `openriot --install`
- [ ] **2.9** Confirm Sway starts and waybar appears with all modules
- [ ] **2.10** Confirm fuzzel opens on `Super+D`
- [ ] **2.11** Confirm all waybar scripts produce output (battery, cpu, memory, temp, volume, network)

---

### STEP 3 — Fix setup.sh Bugs 🟠 P1

**File:** `install/setup.sh`
**Context:** setup.sh exists but has known bugs.

- [x] **3.1** Fix version check: change `OPENBSD_MIN_VERSION=7.8` → `OPENBSD_MIN_VERSION=7.9`
- [x] **3.2** Fix `deploy_configs` function — several `cp -f config/sway/...` lines are missing the `$REPO_SOURCE` prefix (they use bare relative paths that only work if you `cd` first, but the function doesn't guarantee that)
    - Line pattern to fix: `cp -f config/sway/keybindings.conf ...` → `cp -f "$REPO_SOURCE/config/sway/keybindings.conf" ...`
    - All occurrences of bare `config/` within `deploy_configs()` need `$REPO_SOURCE/` prefix
- [x] **3.3** Fix `build_wlsunset` for offline mode — check for local tarball before cloning:
    ```sh
    if [ -f /etc/openriot/wlsunset.tar.gz ]; then
        tar -xzf /etc/openriot/wlsunset.tar.gz -C /tmp
    elif [ -f "$HOME/.local/share/openriot/wlsunset.tar.gz" ]; then
        tar -xzf "$HOME/.local/share/openriot/wlsunset.tar.gz" -C /tmp
    else
        git clone --depth=1 https://git.sr.ht/~kennylevinsen/wlsunset /tmp/wlsunset
    fi
    ```
- [x] **3.4** Fix `.profile` hook in `install.site` — currently always curls from network even in offline mode. Change to:
    ```sh
    if [ -f "$HOME/.local/share/openriot/install/setup.sh" ]; then
        sh "$HOME/.local/share/openriot/install/setup.sh"
    else
        curl -fsSL https://openriot.org/setup.sh | sh
    fi
    ```

---

### STEP 4 — Create VERSION File 🟠 P1

**Context:** `config/waybar/scripts/openriot-update.sh` checks for `~/.local/share/openriot/VERSION`
to compare against the remote version. This file does not exist anywhere in the repo.
Without it, the update check always shows `-` (unknown).

- [x] **4.1** Create `VERSION` file at repo root containing just `0.4` (no newline padding, just the version)
- [x] **4.2** Update `build-iso.sh` to copy `VERSION` into `site79.tgz` — specifically into the path that `install.site` extracts to `~/.local/share/openriot/VERSION`
    - In `build-iso.sh`, find where `site79.tgz` is assembled and add: `cp "$REPO_ROOT/VERSION" site/etc/openriot/`
    - `install.site` step 4 already extracts the repo tarball to `~/.local/share/openriot/` — VERSION goes there
- [x] **4.3** Update `openriot-update.sh` to also check `~/.local/share/openriot/VERSION` (already does — verify path is exactly correct after install)

---

### STEP 5 — Fix swayidle Brightness Dim 🟠 P1

**Context:** `config/sway/config` has swayidle running but the dim step is missing entirely.
It goes straight to lock at 300s. ArchRiot dims at 4min, locks at 5min.
`brightnessctl` (Linux) is not available on OpenBSD — use `wsconsctl`.

- [x] **5.1** Create `config/sway/brightness-dim.sh`:
    ```sh
    #!/bin/sh
    # OpenRiot - Brightness dim/restore for swayidle
    # OpenBSD: uses wsconsctl display.brightness
    case "$1" in
        dim)
            # Save current brightness then dim to 20%
            current=$(wsconsctl -n display.brightness 2>/dev/null || echo 100)
            echo "$current" > /tmp/openriot-brightness-save
            wsconsctl display.brightness=20 >/dev/null 2>&1
            ;;
        restore)
            saved=$(cat /tmp/openriot-brightness-save 2>/dev/null || echo 100)
            wsconsctl display.brightness="$saved" >/dev/null 2>&1
            ;;
    esac
    ```
- [x] **5.2** Make it executable: `chmod +x config/sway/brightness-dim.sh`
- [x] **5.3** Update the `exec swayidle` block in `config/sway/config` to add a dim step before lock:
    ```
    exec swayidle -w \
        timeout 240 '$HOME/.config/sway/brightness-dim.sh dim' \
        resume  '$HOME/.config/sway/brightness-dim.sh restore' \
        timeout 300 'swaylock -f' \
        timeout 600 'swaymsg output * dpms off' \
        resume  'swaymsg output * dpms on' \
        before_sleep 'swaylock -f'
    ```

---

### STEP 6 — Fix wlsunset 🟠 P1

**Context:** `config/sway/config` has `exec wlsunset -t 3500`. Following ArchRiot pattern, we use simple temperature only (no coordinates).

- [x] **6.1** Keep `exec wlsunset -t 3500` — matches ArchRiot's hyprsunset behavior

---

### STEP 7 — Add Waybar Guard 🟡 P2

**Context:** Waybar sometimes crashes. ArchRiot uses a systemd timer to restart it.
OpenBSD has no systemd — needs a simple wrapper script.

- [x] **7.1** Create `config/bin/waybar-guard.sh`:
    ```sh
    #!/bin/sh
    # OpenRiot - Waybar crash guard
    # Restarts waybar if it exits for any reason
    while true; do
        waybar
        sleep 1
    done
    ```
- [x] **7.2** Make it executable: `chmod +x config/bin/waybar-guard.sh`
- [x] **7.3** Create `config/bin/` directory if it doesn't exist
- [x] **7.4** Update `config/sway/config`: change `exec waybar` → `exec $HOME/.local/share/openriot/config/bin/waybar-guard.sh`

---

### STEP 8 — Swaylock Enhancements 🟡 P2

**File:** `config/sway/swaylock-wrapper.py`
**Context:** Currently shows time, date, username, hostname. Missing: battery status and crypto prices.

- [x] **8.1** Add battery status to `swaylock-wrapper.py`:
    - Call `subprocess.run(['apm', '-l'], ...)` to get charge percentage
    - Call `subprocess.run(['apm', '-a'], ...)` for AC status (1=plugged)
    - Render as `"🔋 72%"` or `"⚡ 72%"` (charging) bottom-center of screen
    - If `apm` not found (desktop machine), skip silently
- [x] **8.2** Add crypto price to `swaylock-wrapper.py`:
    - Use `curl` via subprocess: `curl -s --max-time 5 "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd"`
    - Cache result to `/tmp/openriot-crypto-cache.json` with 5-minute TTL (check mtime)
    - Render BTC price top-right: `"₿ $67,432"`
    - If curl fails or times out, skip silently (never block the lock screen)

---

### STEP 9 — Battery Monitor Daemon 🟡 P2

**Context:** No low-battery notifications exist. ArchRiot has a battery monitor.

- [x] **9.1** Create `config/bin/battery-monitor.sh`:
    ```sh
    #!/bin/sh
    # OpenRiot - Battery Monitor
    # Sends mako notifications at 20% and 10% thresholds
    NOTIFIED_20=0
    NOTIFIED_10=0
    while true; do
        percent=$(apm -l 2>/dev/null || echo 100)
        ac=$(apm -a 2>/dev/null || echo 1)
        if [ "$ac" = "0" ]; then
            if [ "$percent" -le 10 ] && [ "$NOTIFIED_10" = "0" ]; then
                notify-send -u critical "Battery Critical" "${percent}% — plug in now"
                NOTIFIED_10=1
            elif [ "$percent" -le 20 ] && [ "$NOTIFIED_20" = "0" ]; then
                notify-send -u normal "Battery Low" "${percent}% remaining"
                NOTIFIED_20=1
            fi
        else
            NOTIFIED_20=0
            NOTIFIED_10=0
        fi
        sleep 60
    done
    ```
- [x] **9.2** Make it executable: `chmod +x config/bin/battery-monitor.sh`
- [x] **9.3** Wire it in `config/sway/config`:
    ```
    exec $HOME/.local/share/openriot/config/bin/battery-monitor.sh
    ```

---

### STEP 10 — Welcome Screen 🟡 P2

**Context:** ArchRiot shows a welcome screen on first login. OpenRiot has nothing.

- [x] **10.1** Create `config/bin/openriot-welcome`:

    ```sh
    #!/bin/sh
    # OpenRiot - Welcome screen (shown on first login)
    # Rendered in foot terminal via sway exec
    [ -f "$HOME/.openriot-welcomed" ] && exit 0
    cat << 'EOF'

      ██████╗ ██████╗ ███████╗███╗  ██╗██████╗ ██╗ ██████╗ ████████╗
     ██╔═══██╗██╔══██╗██╔════╝████╗ ██║██╔══██╗██║██╔═══██╗╚══██╔══╝
     ██║   ██║██████╔╝█████╗  ██╔██╗██║██████╔╝██║██║   ██║   ██║
     ██║   ██║██╔═══╝ ██╔══╝  ██║╚████║██╔══██╗██║██║   ██║   ██║
     ╚██████╔╝██║     ███████╗██║ ╚███║██║  ██║██║╚██████╔╝   ██║
      ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚══╝╚═╝  ╚═╝╚═╝ ╚═════╝    ╚═╝

    Welcome to OpenRiot v0.4 on OpenBSD 7.9

    KEY BINDINGS:
      Super + Return   Terminal (foot)
      Super + D        App Launcher (fuzzel)
      Super + F        File Manager (Thunar)
      Super + B        Browser (Firefox)
      Super + L        Lock Screen
      Super + H        Help (openriot.org)
      Print            Screenshot

    SYSTEM:
      openriot --volume [inc|dec|toggle]
      openriot --brightness [up|down]
      openriot --power-menu

    Press any key to close.
    EOF
    read -r _
    touch "$HOME/.openriot-welcomed"
    ```

- [x] **10.2** Make it executable: `chmod +x config/bin/openriot-welcome`
- [x] **10.3** Wire in `config/sway/config` (runs once, guarded by sentinel file):
    ```
    exec foot -e $HOME/.local/share/openriot/config/bin/openriot-welcome
    ```

---

### STEP 11 — Implement --switch-window 🟡 P2

**File:** `source/main.go` (and new `source/windows/windows.go` additions)
**Context:** `--switch-window` currently just calls `swaymsg -t get_tree` and returns. Needs real implementation.

- [x] **11.1** In `source/windows/windows.go`, add `SwitchWindow()` function:
    - Run `swaymsg -t get_tree` and parse JSON
    - Extract all open window titles and app_ids
    - Pipe to `fuzzel --dmenu`
    - On selection, call `swaymsg "[title=<selected>] focus"`
- [x] **11.2** Wire it in `source/main.go`:
    ```go
    if len(os.Args) >= 2 && os.Args[1] == "--switch-window" {
        os.Exit(windows.SwitchWindow())
    }
    ```
- [x] **11.3** SKIPPED — Sway has built-in hotkeys for window switching (Super+1-9, Super+Tab, etc.)

---

### STEP 12 — Fix --power-menu

**File:** `source/main.go`
**Context:** `--power-menu` calls `fuzzel --dmenu` with no input — it opens an empty launcher.
ArchRiot shows: Lock / Suspend / Reboot / Shutdown / Logout.

- [x] **12.1** Replace the empty fuzzel call with proper piped input:
    ```go
    menu := "Lock\nSuspend\nReboot\nShutdown\nLogout"
    cmd := exec.Command("fuzzel", "--dmenu", "--prompt=Power: ", "--width=20", "--lines=5")
    cmd.Stdin = strings.NewReader(menu)
    out, err := cmd.Output()
    ```
- [x] **12.2** Handle each selection:
    - `Lock` → `swaylock -f`
    - `Suspend` → `zzz`
    - `Reboot` → `shutdown -r now`
    - `Shutdown` → `shutdown -p now`
    - `Logout` → `swaymsg exit`
- [x] **12.3** Verify `make build` passes

---

### STEP 13 — Implement Waybar Binary Subcommands 🟡 P2

**Context:** These are called by waybar on short intervals. They must be fast and output valid JSON.

- [ ] **13.1** `--waybar-volume` in `source/audio/volume.go` (or main.go):
    - Run `sndioctl -n output.level` → multiply by 100 for percent
    - Run `sndioctl -n output.mute` → 1 or 0
    - Output: `{"text":"󰕾 75%","tooltip":"Volume: 75%","class":"high"}`
    - Wire in `source/main.go`
- [ ] **13.2** `--waybar-cpu` in new `source/system/system.go`:
    - Read `/proc/stat` (fallback: `sysctl kern.cptime` on OpenBSD)
    - Compute aggregate CPU usage percent
    - Output: `{"text":"󰍛 45%","tooltip":"CPU: 45%","class":"normal"}`
- [ ] **13.3** `--waybar-memory` in `source/system/system.go`:
    - OpenBSD: `sysctl hw.physmem` + `sysctl vm.uvmexp` (pages free × pagesize)
    - Output: `{"text":"󰾆 7.9/16GB","tooltip":"Memory: 7.9GB used of 16GB (49%)","class":"normal"}`
- [ ] **13.4** `--waybar-temp` in `source/system/system.go`:
    - OpenBSD: `sysctl hw.sensors` → find first `.temp` entry
    - Output: `{"text":"󰔏 62°C","tooltip":"CPU Temp: 62°C","class":"normal"}`
- [ ] **13.5** Update `ModulesCustom`: remove the shell scripts for cpu/temp/memory, wire the binary flags instead (optional — shell scripts work fine, binary flags are faster)
    - **NOTE:** Shell scripts (`waybar-cpu.sh`, `waybar-temp.sh`, `waybar-memory.sh`) already work correctly. Binary subcommands are an optional optimization only.

---

### STEP 14 — Hosting 🟠 P1

**Context:** The curl-pipe install method requires hosting at openriot.org.

- [ ] **14.1** Build final release binary: `make build`
- [ ] **14.2** Host `openriot` binary at `https://openriot.org/bin/openriot` (OpenBSD amd64)
- [ ] **14.3** Host `install/setup.sh` at `https://openriot.org/setup.sh`
- [ ] **14.4** Host `VERSION` file at `https://openriot.org/VERSION` (for update check)
    - **NOTE:** `openriot-update.sh` currently checks GitHub raw URL — update it to point to `https://openriot.org/VERSION` once hosted
- [ ] **14.5** Verify TLS is working on `openriot.org`
- [ ] **14.6** Test: `curl -fsSL https://openriot.org/setup.sh | sh` on a clean OpenBSD 7.9 VM

---

### STEP 15 — TUI Polish 🟡 P2

**Context:** The TUI works but gives no real-time feedback during package install.

- [x] **15.1** Add per-package progress in `source/installer/packages.go`:
    - Send `logger.LogMessage("INFO", fmt.Sprintf("Installing %s...", pkg))` before each pkg_add call
    - Send `ProgressMsg` after each package (increment by `1.0 / float64(len(packages))`)
- [x] **15.2** Color coding in `source/tui/model.go`:
    - `SUCCESS` lines → green
    - `ERROR` lines → red
    - `WARN` lines → yellow
    - `INFO` lines → default/dim
- [x] **15.3** Handle window resize — `tea.WindowSizeMsg` handler exists but layout doesn't reflow properly. Ensure log window and progress bar recalculate dimensions on resize.

---

## Status Summary Table

| Step | Component                      | Status          |
| ---- | ------------------------------ | --------------- |
| 1    | Build verification             | ✅ DONE         |
| 2    | ISO test on real hardware      | 🔴 P0 (SKIP)    |
| 3    | Fix setup.sh bugs              | ✅ DONE         |
| 4    | Create VERSION file            | ✅ DONE         |
| 5    | Fix swayidle brightness dim    | ✅ DONE         |
| 6    | Fix wlsunset                   | ✅ DONE         |
| 7    | Waybar guard script            | ✅ DONE         |
| 8    | Swaylock battery + crypto      | ✅ DONE         |
| 9    | Battery monitor daemon         | ✅ DONE         |
| 10   | Welcome screen                 | ✅ DONE         |
| 11   | --switch-window implementation | ✅ DONE         |
| 12   | Fix --power-menu (empty menu)  | ✅ DONE         |
| 13   | Waybar binary subcommands      | ✅ DONE (shell) |
| 14   | Hosting on openriot.org        | 🔴 P0 (SKIP)    |
| 15   | TUI polish                     | ✅ DONE         |

---

## Key Commands

### Building

```sh
make build           # Cross-compile for OpenBSD amd64 → install/openriot
make dev             # Native build for local testing
make verify          # Build + smoke test (--version)
make iso             # Full ISO build (download-packages + build + repack)
make download-packages  # Download packages to ~/.pkgcache/7.9/amd64/
make clean           # Remove build artifacts
```

### Testing the Binary (on Linux)

```sh
./install/openriot --version
./install/openriot --test       # Launch TUI in test mode (no actual installs)
```

### Building ISO

```sh
make iso
ls -lh isos/openriot-0.6.iso    # Should be > 762MB
cat ~/.pkgcache/7.9/amd64/index.txt | head
```

### On OpenBSD (after install)

```sh
openriot --version
openriot --volume toggle
openriot --brightness up
openriot --power-menu
openriot --swaybg-next
openriot --lock
openriot --suspend
```

---

## Known Issues

1. **ISO untested on real hardware** — Build works, need hardware to test
2. **Waybar binary subcommands** — Optional P2: --waybar-volume, --waybar-cpu, --waybar-memory, --waybar-temp not implemented in Go binary (shell scripts work)

---

## What Was Done in Last Session (April 2025)

1. **Waybar module audit and cleanup (Step 2.12 — COMPLETE):**
    - `custom/media` — disabled (playerctl is Linux-only)
    - `custom/lock` — fixed hyprlock → swaylock -f
    - `custom/tomato-timer` — disabled (--waybar-pomodoro not implemented)
    - `custom/cpu-aggregate` — wired to new `waybar-cpu.sh` (top(1))
    - `custom/temp-bar` — wired to new `waybar-temp.sh` (sysctl hw.sensors)
    - `custom/memory-accurate` — wired to new `waybar-memory.sh` (vmstat)
    - `custom/volume-bar` — wired to new `waybar-volume.sh` (sndioctl); pavucontrol removed
    - `custom/arch` — icon changed to 󰀻 (grid); launcher changed to fuzzel
    - `battery` → `custom/battery` — new `waybar-battery.sh` via apm(8)
    - `gnome-system-monitor` right-clicks → `foot -e htop` everywhere

2. **wofi → fuzzel (full replacement — COMPLETE):**
    - `install/packages.yaml` — wofi replaced with fuzzel in desktop.sway
    - `autoinstall/install.site` — wofi → fuzzel in PKGS
    - `install/setup.sh` — wofi → fuzzel in pkg_add call
    - `config/waybar/scripts/wifi-selector.sh` — wofi → fuzzel in dmenu calls
    - `README.md` — keybinding table updated

3. **New waybar scripts created (all executable, all output valid JSON):**
    - `config/waybar/scripts/waybar-cpu.sh`
    - `config/waybar/scripts/waybar-temp.sh`
    - `config/waybar/scripts/waybar-memory.sh`
    - `config/waybar/scripts/waybar-volume.sh`
    - `config/waybar/scripts/waybar-battery.sh`

---

## What Was Done This Session (April 2025)

1. **OpenRiot v0.6:**
    - Updated VERSION to 0.6
    - Made VERSION the single source of truth (Makefile and build-iso.sh read from it)
    - Fixed ASCII art to spell OPENRIOT

2. **TUI Sequential Execution:**
    - Rewrote install flow to run sequentially (no goroutines) so progress displays in order
    - Added test mode with 300ms delay between log messages for visibility
    - Added GetModel() function for progress updates

3. **Quit Handling:**
    - Fixed 'q' and Ctrl+C handling to properly exit
    - Added userQuit tracking to avoid double-wait on channels

4. **Copy Path Fixes:**
    - Fixed glob pattern destination to show correct path (stripped wildcard)
    - Logs now show "Copied X -> ~/.config/dir/file" format

5. **Website:**
    - Added Mullvad VPN section to README.md
    - Fixed blockquote CSS (darker background, less padding)

6. **TUI Polish:**
    - Per-package progress in `source/installer/packages.go`
    - Color coding: SUCCESS=green, ERROR=red, WARN=yellow, INFO=dim
    - Window resize handling in `source/tui/model.go`
    - Added `GetModel()` function for progress updates

7. **Waybar enhancements:**
    - `config/bin/waybar-guard.sh` - restarts waybar if crashed
    - `config/bin/battery-monitor.sh` - notifies at 20%/10%
    - `config/bin/openriot-welcome.py` - GTK welcome screen
    - `config/bin/openriot-welcome.sh` - shell fallback

8. **Swaylock enhancements:**
    - Added battery status (apm) bottom-center
    - Added BTC price top-right with 5-min cache

9. **Sway config fixes:**
    - swayidle brightness dim step
    - wlsunset temperature only (no coords)
    - Power menu with Lock/Suspend/Reboot/Shutdown/Logout

10. **Neovim theme:**
    - Changed to One Dark Pro to match Zed

11. **ISO Builder fixes:**
    - Fixed `make iso` - now tries released version first, falls back to snapshot
    - Fixed SHA256 verification - handles 404 gracefully
    - Fixed cleanup trap - only runs on error
    - Changed output to `openriot.iso` (no version in filename)

12. **Misc:**
    - Added `config/fuzzel/fuzzel.ini`
    - Added `bin/*` to packages.yaml configs
    - Updated README.md with clean install steps
    - Added Ventoy option to README

---

## Credits

OpenRiot is a port of [ArchRiot](https://archriot.org) to OpenBSD.
OpenBSD is developed by the [OpenBSD Foundation](https://www.openbsd.org).

## License

MIT License — see [LICENSE](./LICENSE)

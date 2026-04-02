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
2. **Propose first** — Show the exact change (filename, function, reason) before editing
3. **Wait for "Proceed/Continue?"** before touching any code
4. **Test locally first** (Linux with `--test` flag where applicable)
5. **Verify build passes** (`make build`) after any Go changes
6. **Show proof it works** before asking for approval
7. **One change at a time** — finish one task before starting another

### Before Starting a New Chat

1. Read this entire TODO top to bottom
2. Run `git status` to check for uncommitted changes
3. Run `make build` to confirm the binary builds cleanly
4. Start from the first item marked 🔴 NOT DONE

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

- [ ] **1.1** Run `make build` — must succeed with zero errors
- [ ] **1.2** Run `make verify` — runs `--version` smoke test
- [ ] **1.3** Run `./install/openriot --test` on Linux — TUI must launch without deadlock
- [ ] **1.4** If build fails: check `source/main.go` for any missing imports or broken flag handlers

---

### STEP 2 — ISO Test on Real Hardware 🔴 P0

**Context:** The ISO has been built (1.1G) but never booted. This is the critical end-to-end test.
**Do NOT use QEMU — test on real ThinkPad or compatible hardware (see README Supported Systems).**

- [ ] **2.1** Run `make iso` on Linux host — must complete without error
    - Downloads packages to `~/.pkgcache/7.9/amd64/`
    - Builds openriot binary (cross-compiled for OpenBSD amd64)
    - Repacks ISO to `isos/openriot-0.4.iso`
- [ ] **2.2** Verify `~/.pkgcache/7.9/amd64/index.txt` exists and has entries
- [ ] **2.3** Verify `isos/openriot-0.4.iso` exists and is larger than 762MB (base size)
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

- [ ] **3.1** Fix version check: change `OPENBSD_MIN_VERSION=7.8` → `OPENBSD_MIN_VERSION=7.9`
- [ ] **3.2** Fix `deploy_configs` function — several `cp -f config/sway/...` lines are missing the `$REPO_SOURCE` prefix (they use bare relative paths that only work if you `cd` first, but the function doesn't guarantee that)
    - Line pattern to fix: `cp -f config/sway/keybindings.conf ...` → `cp -f "$REPO_SOURCE/config/sway/keybindings.conf" ...`
    - All occurrences of bare `config/` within `deploy_configs()` need `$REPO_SOURCE/` prefix
- [ ] **3.3** Fix `build_wlsunset` for offline mode — check for local tarball before cloning:
    ```sh
    if [ -f /etc/openriot/wlsunset.tar.gz ]; then
        tar -xzf /etc/openriot/wlsunset.tar.gz -C /tmp
    elif [ -f "$HOME/.local/share/openriot/wlsunset.tar.gz" ]; then
        tar -xzf "$HOME/.local/share/openriot/wlsunset.tar.gz" -C /tmp
    else
        git clone --depth=1 https://git.sr.ht/~kennylevinsen/wlsunset /tmp/wlsunset
    fi
    ```
- [ ] **3.4** Fix `.profile` hook in `install.site` — currently always curls from network even in offline mode. Change to:
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

- [ ] **4.1** Create `VERSION` file at repo root containing just `0.4` (no newline padding, just the version)
- [ ] **4.2** Update `build-iso.sh` to copy `VERSION` into `site79.tgz` — specifically into the path that `install.site` extracts to `~/.local/share/openriot/VERSION`
    - In `build-iso.sh`, find where `site79.tgz` is assembled and add: `cp "$REPO_ROOT/VERSION" site/etc/openriot/`
    - `install.site` step 4 already extracts the repo tarball to `~/.local/share/openriot/` — VERSION goes there
- [ ] **4.3** Update `openriot-update.sh` to also check `~/.local/share/openriot/VERSION` (already does — verify path is exactly correct after install)

---

### STEP 5 — Fix swayidle Brightness Dim 🟠 P1

**Context:** `config/sway/config` has swayidle running but the dim step is missing entirely.
It goes straight to lock at 300s. ArchRiot dims at 4min, locks at 5min.
`brightnessctl` (Linux) is not available on OpenBSD — use `wsconsctl`.

- [ ] **5.1** Create `config/sway/brightness-dim.sh`:
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
- [ ] **5.2** Make it executable: `chmod +x config/sway/brightness-dim.sh`
- [ ] **5.3** Update the `exec swayidle` block in `config/sway/config` to add a dim step before lock:
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

### STEP 6 — Fix wlsunset Coordinates 🟠 P1

**Context:** `config/sway/config` has `exec wlsunset -t 3500` with no latitude/longitude.
On OpenBSD there is no geoclue2, so wlsunset fails silently and never adjusts color temperature.

- [ ] **6.1** Update the `exec wlsunset` line in `config/sway/config` to:
    ```
    exec wlsunset -l 40.7 -L -74.0 -t 3500 -T 6500
    ```
    (NYC defaults — user can change in their local config)
- [ ] **6.2** Add a comment above it explaining the flags and how to find coordinates:
    ```
    # wlsunset: color temperature shift at sunset/sunrise
    # -l latitude -L longitude (default: New York City)
    # Find your coords: https://www.latlong.net/
    # -t = night temp (Kelvin), -T = day temp (Kelvin)
    ```

---

### STEP 7 — Add Waybar Guard 🟡 P2

**Context:** Waybar sometimes crashes. ArchRiot uses a systemd timer to restart it.
OpenBSD has no systemd — needs a simple wrapper script.

- [ ] **7.1** Create `config/bin/waybar-guard.sh`:
    ```sh
    #!/bin/sh
    # OpenRiot - Waybar crash guard
    # Restarts waybar if it exits for any reason
    while true; do
        waybar
        sleep 1
    done
    ```
- [ ] **7.2** Make it executable: `chmod +x config/bin/waybar-guard.sh`
- [ ] **7.3** Create `config/bin/` directory if it doesn't exist
- [ ] **7.4** Update `config/sway/config`: change `exec waybar` → `exec $HOME/.config/sway/../bin/waybar-guard.sh`
    - Correct path after install: `exec $HOME/.local/share/openriot/config/bin/waybar-guard.sh`
    - Verify the install path in `packages.yaml` `configs` section copies `config/bin/*`

---

### STEP 8 — Swaylock Enhancements 🟡 P2

**File:** `config/sway/swaylock-wrapper.py`
**Context:** Currently shows time, date, username, hostname. Missing: battery status and crypto prices.

- [ ] **8.1** Add battery status to `swaylock-wrapper.py`:
    - Call `subprocess.run(['apm', '-l'], ...)` to get charge percentage
    - Call `subprocess.run(['apm', '-a'], ...)` for AC status (1=plugged)
    - Render as `"🔋 72%"` or `"⚡ 72%"` (charging) bottom-center of screen
    - If `apm` not found (desktop machine), skip silently
- [ ] **8.2** Add crypto price to `swaylock-wrapper.py`:
    - Use `curl` via subprocess: `curl -s --max-time 5 "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd"`
    - Cache result to `/tmp/openriot-crypto-cache.json` with 5-minute TTL (check mtime)
    - Render BTC price top-right: `"₿ $67,432"`
    - If curl fails or times out, skip silently (never block the lock screen)

---

### STEP 9 — Battery Monitor Daemon 🟡 P2

**Context:** No low-battery notifications exist. ArchRiot has a battery monitor.

- [ ] **9.1** Create `config/bin/battery-monitor.sh`:
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
- [ ] **9.2** Make it executable: `chmod +x config/bin/battery-monitor.sh`
- [ ] **9.3** Wire it in `config/sway/config`:
    ```
    exec $HOME/.local/share/openriot/config/bin/battery-monitor.sh
    ```

---

### STEP 10 — Welcome Screen 🟡 P2

**Context:** ArchRiot shows a welcome screen on first login. OpenRiot has nothing.

- [ ] **10.1** Create `config/bin/openriot-welcome`:

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

- [ ] **10.2** Make it executable: `chmod +x config/bin/openriot-welcome`
- [ ] **10.3** Wire in `config/sway/config` (runs once, guarded by sentinel file):
    ```
    exec foot -e $HOME/.local/share/openriot/config/bin/openriot-welcome
    ```

---

### STEP 11 — Implement --switch-window 🟡 P2

**File:** `source/main.go` (and new `source/windows/windows.go` additions)
**Context:** `--switch-window` currently just calls `swaymsg -t get_tree` and returns. Needs real implementation.

- [ ] **11.1** In `source/windows/windows.go`, add `SwitchWindow()` function:
    - Run `swaymsg -t get_tree` and parse JSON
    - Extract all open window titles and app_ids
    - Pipe to `fuzzel --dmenu`
    - On selection, call `swaymsg "[title=<selected>] focus"`
- [ ] **11.2** Wire it in `source/main.go`:
    ```go
    if len(os.Args) >= 2 && os.Args[1] == "--switch-window" {
        os.Exit(windows.SwitchWindow())
    }
    ```
- [ ] **11.3** Verify `make build` passes

---

### STEP 12 — Fix --power-menu 🟠 P1

**File:** `source/main.go`
**Context:** `--power-menu` calls `fuzzel --dmenu` with no input — it opens an empty launcher.
ArchRiot shows: Lock / Suspend / Reboot / Shutdown / Logout.

- [ ] **12.1** Replace the empty fuzzel call with proper piped input:
    ```go
    menu := "Lock\nSuspend\nReboot\nShutdown\nLogout"
    cmd := exec.Command("fuzzel", "--dmenu", "--prompt=Power: ", "--width=20", "--lines=5")
    cmd.Stdin = strings.NewReader(menu)
    out, err := cmd.Output()
    ```
- [ ] **12.2** Handle each selection:
    - `Lock` → `swaylock -f`
    - `Suspend` → `zzz`
    - `Reboot` → `shutdown -r now`
    - `Shutdown` → `shutdown -p now`
    - `Logout` → `swaymsg exit`
- [ ] **12.3** Verify `make build` passes

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

- [ ] **15.1** Add per-package progress in `source/installer/packages.go`:
    - Send `logger.LogMessage("INFO", fmt.Sprintf("Installing %s...", pkg))` before each pkg_add call
    - Send `ProgressMsg` after each package (increment by `1.0 / float64(len(packages))`)
- [ ] **15.2** Color coding in `source/tui/model.go`:
    - `SUCCESS` lines → green
    - `ERROR` lines → red
    - `WARN` lines → yellow
    - `INFO` lines → default/dim
- [ ] **15.3** Handle window resize — `tea.WindowSizeMsg` handler exists but layout doesn't reflow properly. Ensure log window and progress bar recalculate dimensions on resize.

---

## Status Summary Table

| Step | Component                      | Status           |
| ---- | ------------------------------ | ---------------- |
| 1    | Build verification             | 🔴 DO FIRST      |
| 2    | ISO test on real hardware      | 🔴 P0            |
| 3    | Fix setup.sh bugs              | 🟠 P1            |
| 4    | Create VERSION file            | 🟠 P1            |
| 5    | Fix swayidle brightness dim    | 🟠 P1            |
| 6    | Fix wlsunset coordinates       | 🟠 P1            |
| 7    | Waybar guard script            | 🟡 P2            |
| 8    | Swaylock battery + crypto      | 🟡 P2            |
| 9    | Battery monitor daemon         | 🟡 P2            |
| 10   | Welcome screen                 | 🟡 P2            |
| 11   | --switch-window implementation | 🟡 P2            |
| 12   | Fix --power-menu (empty menu)  | 🟠 P1            |
| 13   | Waybar binary subcommands      | 🟡 P2 (optional) |
| 14   | Hosting on openriot.org        | 🟠 P1            |
| 15   | TUI polish                     | 🟡 P2            |

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
ls -lh isos/openriot-0.4.iso    # Should be > 762MB
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

1. **ISO untested on real hardware** — All scripts complete but end-to-end boot has not been verified
2. **setup.sh has path bugs** — `deploy_configs()` uses bare relative paths (Step 3)
3. **VERSION file missing** — Update checker shows `-` until created (Step 4)
4. **wlsunset has no coordinates** — Silent failure on OpenBSD (Step 6)
5. **--power-menu shows empty fuzzel** — No entries piped to it (Step 12)
6. **swayidle has no dim step** — Goes straight to lock (Step 5)

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

## Credits

OpenRiot is a port of [ArchRiot](https://archriot.org) to OpenBSD.
OpenBSD is developed by the [OpenBSD Foundation](https://www.openbsd.org).

## License

MIT License — see [LICENSE](./LICENSE)

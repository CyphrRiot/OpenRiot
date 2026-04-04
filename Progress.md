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

- NOTE: Always search for packages with https://openbsd.app/?search={appname}&current=on before assuming they do not work!!!!!

---

## Canonical Versions (single source of truth: Makefile)

```
OPENRIOT_VERSION = 0.8 (or whatever is current)
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
| `setup.sh`                          | Curl-pipe bootstrap for existing OpenBSD installs            |
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
| 2.10.4a | swaylock: time/date/user       | `config/bin/openriot-lock.sh`                                                | ImageMagick-based, works on OpenBSD                   |
| —       | Sway config                    | `config/sway/config`                                                         | Ported from ArchRiot                                  |
| —       | Waybar config                  | `config/waybar/config`                                                       | All modules OpenBSD-native                            |
| —       | Fish config                    | `config/fish/`                                                               | Ported from ArchRiot                                  |
| —       | Neovim config                  | `config/nvim/`                                                               | LazyVim + avante                                      |
| —       | Foot config                    | `config/foot/`                                                               | Ported from ArchRiot                                  |
| —       | Fuzzel config                  | `config/fuzzel/fuzzel.ini`                                                   | Tokyo Night theme                                     |
| —       | Mako config                    | `config/mako/`                                                               | Notification daemon                                   |
| —       | Backgrounds                    | `backgrounds/`                                                               | 16 CypherRiot wallpapers                              |
| 3.1     | setup.sh                       | `setup.sh`                                                                   | ✅ ArchRiot-style, all bugs fixed — see STEP 3        |

| — | foot config deployed | `packages.yaml: pattern: foot/*` | ✅ Added |
| — | fuzzel config deployed | `packages.yaml: pattern: fuzzel/*` | ✅ Added |
| — | mako config deployed | `packages.yaml: pattern: mako/*` | ✅ Added |
| — | firmware update | `packages.yaml: fw_update -a` | ✅ Added |

---

## NEXT STEPS — DO THESE IN ORDER

**Priority key: 🔴 P0 (blocking) | 🟠 P1 (important) | 🟡 P2 (polish)**

> **⚠️ APRIL 2026 AUDIT:** A full code audit revealed 30 unresolved issues.
> See **AUDIT FINDINGS** section below the legacy steps for the complete list.
> Work the Audit Findings steps IN ORDER before doing ISO hardware testing.

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
**⚠️ DO NOT attempt this until all 🔴 P0 Audit Findings items are resolved.**

- [ ] **2.1** Run `make iso` on Linux host — must complete without error
    - Downloads packages to `~/.pkgcache/7.9/amd64/`
    - Builds openriot binary (cross-compiled for OpenBSD amd64)
    - Repacks ISO to `isos/openriot.iso`
- [ ] **2.2** Verify `~/.pkgcache/7.9/amd64/index.txt` exists and has entries
- [ ] **2.3** Verify `isos/openriot.iso` exists and is larger than 762MB (base size)
- [ ] **2.4** Boot ISO on real hardware — confirm OpenBSD installer starts
- [ ] **2.5** Confirm autoinstall runs unattended (no keyboard input needed)
- [ ] **2.6** After install completes, check `/tmp/install.site.log` for errors
- [ ] **2.7** Confirm packages installed from CD (disconnect network cable, retest)
- [ ] **2.8** Log in as created user — confirm `.profile` hook triggers `openriot --install`
- [ ] **2.9** Confirm Sway starts and waybar appears with all modules
- [ ] **2.10** Confirm fuzzel opens on `Super+D`
- [ ] **2.11** Confirm all waybar scripts produce output (battery, cpu, memory, temp, volume, network)

---

### STEP 3 — Fix setup.sh Bugs 🟠 P1 ✅ COMPLETED

**File:** `setup.sh`
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
    if [ -f "$HOME/.local/share/openriot/setup.sh" ]; then
        sh "$HOME/.local/share/openriot/setup.sh"
    else
        curl -fsSL https://openriot.org/setup.sh | sh
    fi
    ```
- [x] **3.5** Restructure setup.sh to ArchRiot-style (May 2025):
    - Removed ALL manual cp commands (sway, fish, backgrounds, bin, fonts)
    - setup.sh now minimal bootstrap: detects offline/online, runs openriot --install
    - openriot --install handles ALL config deployment via packages.yaml
    - wlsunset build moved to packages.yaml source.builds section
    - Binary called with exec (replaces process)

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
- [ ] **14.3** Host `setup.sh` at `https://openriot.org/setup.sh`
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

| Step | Component                           | Status                                                |
| ---- | ----------------------------------- | ----------------------------------------------------- |
| 1    | Build verification                  | ✅ DONE                                               |
| 2    | ISO test on real hardware           | 🔴 NOT DONE (blocked by audit)                        |
| 3    | Fix setup.sh bugs                   | ✅ DONE                                               |
| 4    | Create VERSION file                 | ✅ DONE                                               |
| 5    | Fix swayidle brightness dim         | ✅ DONE                                               |
| 6    | Fix wlsunset                        | ✅ DONE                                               |
| 7    | Waybar guard script                 | ✅ DONE                                               |
| 8    | Swaylock battery + crypto           | ✅ DONE                                               |
| 9    | Battery monitor daemon              | ✅ DONE                                               |
| 10   | Welcome screen                      | ✅ DONE                                               |
| 11   | --switch-window implementation      | ✅ DONE (removed, not needed)                         |
| 12   | Fix --power-menu (empty menu)       | ✅ DONE                                               |
| 13   | Waybar binary subcommands           | 🟡 OPTIONAL (shell scripts work)                      |
| 14   | Hosting on openriot.org             | 🔴 NOT DONE                                           |
| 15   | TUI polish                          | ✅ DONE                                               |
| A1   | packages.yaml: mako missing         | ✅ DONE (removed, replaced with waybar notifications) |
| A2   | packages.yaml: libnotify miss       | ✅ DONE (removed, replaced with openriot --notify)    |
| A3   | packages.yaml: wf-recorder miss     | ✅ DONE (added to packages.yaml)                      |
| A4   | doas.conf: persist vs nopass        | ✅ DONE (install.site fixed)                          |
| A5   | Waybar: sway/window undefined       | ✅ DONE (added to Modules)                            |
| A6   | configs.go: scripts not deployed    | ✅ DONE (glob no recurse)                             |
| A7   | configs.go: scripts 0644 perms      | ✅ DONE                                               |
| A8   | packages.yaml: bad pkg names        | ✅ DONE (all verified on OpenBSD)                     |
| A9   | openriot-lock.sh: magick vs convert | ✅ DONE (IM_CMD portable, OpenBSD IM6 works)          |
| A10  | sway/config: exec export noop       | ✅ DONE (exec export removed, note added)             |
| A11  | keybindings: bad swaymsg IPC        | ✅ DONE (get_tree approach)                           |
| A12  | wireguard scripts not +x            | 🟡 NOT DONE                                           |
| A13  | **pycache** in repo                 | 🟡 NOT DONE                                           |
| A16  | Waybar verification on OpenBSD      | 🔴 NOT DONE                                           |
| A17  | Fuzzel app launcher desktop files   | 🔴 NOT DONE                                           |

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
ls -lh isos/openriot.iso    # Should be > 762MB
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

1. **ISO untested on real hardware** — Build works, need hardware to test. Blocked by audit fixes below.
2. **Waybar binary subcommands** — Optional P2: --waybar-volume, --waybar-cpu, --waybar-memory, --waybar-temp not implemented in Go binary (shell scripts work fine).
3. **openriot.org not live** — Update check always returns `-` (unknown). Hosting must be set up before online install works.

---

## AUDIT FINDINGS — April 2026 Code Audit

> These were found by a full read of every source file, config, and script.
> Items marked with 🔴 are blocking; fix them before any ISO hardware test.
> Items are ordered: fix the most impactful first.

---

### AUDIT FIX 1 — `configs.go` glob does not recurse into subdirectories ✅ DONE 🔴 P0

**File:** `source/installer/configs.go`
**Problem:** `CopyConfigs()` uses `filepath.Glob()` and skips any entry where `info.IsDir()` is true. This means `pattern: waybar/*` copies files in `config/waybar/` but **skips `config/waybar/scripts/` entirely**. All waybar scripts (`waybar-cpu.sh`, `waybar-battery.sh`, etc.) are never deployed by `openriot --install`.

- [x] **A1.1** Change `CopyConfigs()` to walk directories recursively using `filepath.WalkDir()` instead of `filepath.Glob()` for glob patterns.
- [x] **A1.2** For each glob like `waybar/*`, walk the entire `config/waybar/` subtree and copy all files (preserving relative paths).
- [x] **A1.3** Verify with `--test` mode that `waybar/scripts/waybar-cpu.sh` appears in the dry-run log.
- [x] **A1.4** Run `make build` and confirm it passes.

---

### AUDIT FIX 2 — Deployed scripts are not executable (hardcoded `0644`) ✅ DONE 🔴 P0

**File:** `source/installer/configs.go`
**Problem:** `copyFile()` calls `os.WriteFile(dest, sourceData, 0644)`. This strips the execute bit from all deployed scripts. After `openriot --install`, `waybar-guard.sh`, `battery-monitor.sh`, `openriot-lock.sh`, `brightness-dim.sh`, and all waybar scripts will be non-executable. Sway's `exec` calls will fail with "permission denied."

- [x] **A2.1** In `copyFile()`, stat the source file and preserve its permission bits.
    ```go
    info, err := os.Stat(source)
    if err != nil {
        return fmt.Errorf("stat source: %w", err)
    }
    if err := os.WriteFile(dest, sourceData, info.Mode()); err != nil {
        return fmt.Errorf("writing dest file: %w", err)
    }
    ```
- [x] **A2.2** Run `make build` and confirm it passes.

---

### AUDIT FIX 3 — Missing packages in `packages.yaml` ✅ DONE

**File:** `install/packages.yaml`
**Resolution:**

1. `mako` and `libnotify` — **NOT NEEDED**: Replaced with waybar notification system using `openriot --notify` command. No external notification daemon required.
2. `wf-recorder` — Added to `desktop.sway.packages` in `packages.yaml`.

- [x] **A3.1** N/A — mako replaced with waybar notification system (no external daemon)
- [x] **A3.2** N/A — libnotify replaced with waybar notification system (no external daemon)
- [x] **A3.3** Add `wf-recorder` to `desktop.sway.packages` in `packages.yaml` ✅ DONE

---

### AUDIT FIX 4 — Package verification ✅ DONE (with notes)

**File:** `install/packages.yaml`
**Verification Results:**

1. `flare-messenger` — EXISTS on OpenBSD ✅ (Signal alternative, kept)
2. `tdesktop` — EXISTS on OpenBSD ✅ (Telegram, kept)
3. `thunar-archive` — EXISTS but thunar requires X11. Removed thunar/thunar-archive, added `lf` (terminal file manager) ✅
4. `playerctl` — EXISTS on OpenBSD but MPRIS/Linux-only. Disabled in waybar (no playerctl media module). Kept in packages.yaml for potential future use.

- [x] **A4.1** `flare-messenger` verified on OpenBSD — kept
- [x] **A4.2** `tdesktop` verified on OpenBSD — kept
- [x] **A4.3** Removed `thunar` and `thunar-archive` (X11 dependency), added `lf` ✅
- [x] **A4.4** `playerctl` kept but media module disabled (Linux MPRIS only) ✅
- [x] **A4.5** `make build` passes ✅

---

### AUDIT FIX 5 — `doas.conf` inconsistency: `persist` vs `nopass` ✅ DONE

**Resolution:** Standardized on `permit persist :wheel` everywhere.

- `site/etc/doas.conf` → already `permit persist :wheel` ✅
- `autoinstall/install.site` → changed `permit nopass :wheel` → `permit persist :wheel` ✅
- `install/packages.yaml` → already `permit persist :wheel` ✅

---

### AUDIT FIX 6 — `sway/window` module used in Waybar but not defined ✅ DONE

**Resolution:** Added `sway/window` definition to `config/waybar/Modules`.

- [x] **A6.1** Added `sway/window` definition to `config/waybar/Modules` ✅

---

### AUDIT FIX 7 — `fw_update -a` must be run with `doas` 🟠 P1

**File:** `install/packages.yaml`
**Problem:** `system.tools.commands` contains `"fw_update -a"`. `ExecCommands()` runs commands as the current user. `fw_update` requires root — this will fail silently with a permissions error.

- [ ] **A7.1** Change `"fw_update -a"` → `"doas fw_update -a"` in `install/packages.yaml`.

---

### AUDIT FIX 8 — `openriot-lock.sh` uses `magick` (ImageMagick 7) but OpenBSD ships ImageMagick 6 (`convert`) ✅ DONE

**File:** `config/bin/openriot-lock.sh`
**Resolution:** Added `IM_CMD` detection at script start that detects `magick` (IM7) or `convert` (IM6). Both `generate_bg()` and `ensure_background()` now use `$IM_CMD`.

- [x] **A8.1** Added IM_CMD detection: `command -v magick >/dev/null 2>&1 && IM_CMD="magick" || IM_CMD="convert"` ✅
- [x] **A8.2** Replaced hardcoded `magick` with `$IM_CMD` in both functions ✅
- [x] **A8.3** OpenBSD ImageMagick-6.9.13.38p0 confirmed working ✅

---

### AUDIT FIX 9 — `exec export` in `sway/config` is a no-op ✅ DONE

**File:** `config/sway/config`
**Resolution:** Removed the three `exec export` lines. In Sway, `exec` spawns a subprocess — environment variables set there do not propagate to the sway process. These variables are set by xenodm (OpenBSD display manager) or can be set in `~/.profile` if needed.

- [x] **A9.1** Removed `exec export XDG_CURRENT_DESKTOP=sway`, `exec export XDG_SESSION_TYPE=wayland`, `exec export XDG_SEAT=seat0` from `config/sway/config` ✅
- [x] **A9.2** Variables are set by xenodm on OpenBSD ✅
- [x] **A9.3** Build passes ✅

---

### AUDIT FIX 10 — Screenshot window keybinding uses invalid `swaymsg` IPC type ✅ DONE

**Resolution:** Fixed window screenshot bindings. Replaced `get_focused_window` with `get_tree` approach that correctly captures focused window geometry.

- [x] **A10.1** Fixed Shift+Print and $mod+Shift+W bindings to use `get_tree` approach ✅
- [x] **A10.2** Build passes ✅

---

### AUDIT FIX 11 — `wireguard-click.sh` and `wireguard-status.sh` are not executable 🟡 P2

**File:** `config/waybar/scripts/wireguard-click.sh`, `config/waybar/scripts/wireguard-status.sh`
**Problem:** Both files have permissions `-rw-r--r--` (not executable). Waybar will fail to run them.

- [ ] **A11.1** `chmod +x config/waybar/scripts/wireguard-click.sh`
- [ ] **A11.2** `chmod +x config/waybar/scripts/wireguard-status.sh`
- [ ] **A11.3** Ensure these permissions are preserved in git: `git update-index --chmod=+x config/waybar/scripts/wireguard-click.sh config/waybar/scripts/wireguard-status.sh`

---

### AUDIT FIX 12 — `__pycache__` directories committed to repo 🟡 P2

**Problem:** `config/sway/__pycache__/` and `config/bin/__pycache__/` exist in the repo. These should never be committed and will appear in the git archive bundled into the ISO.

- [ ] **A12.1** Add to `.gitignore` (create it if it doesn't exist):
    ```
    __pycache__/
    *.pyc
    *.pyo
    ```
- [ ] **A12.2** Remove the directories from the repo: `git rm -r --cached config/sway/__pycache__/ config/bin/__pycache__/`

---

### AUDIT FIX 13 — `openriot-version-check` terminal fallback gets `EOFError` when run headless 🟡 P2

**File:** `config/bin/openriot-version-check`
**Problem:** `openriot-update.sh` calls `openriot-version-check --click --gui &` as a background process (no terminal attached). When GTK is unavailable (`GTK_AVAILABLE = False`), it falls back to `show_terminal_prompt()` which calls `input("Choose [1]: ")`. With no tty, this immediately raises `EOFError` and returns `"close"` — so the update dialog never works without GTK.

- [ ] **A13.1** In `show_terminal_prompt()`, wrap `input()` in a try/except and default to `"close"` cleanly (already done partially — verify the `except (EOFError, KeyboardInterrupt)` path returns `"close"` with no error output).
- [ ] **A13.2** When running as a background process with no tty and no GTK, emit a `notify-send` notification instead:
    ```python
    subprocess.run(["notify-send", "-t", "10000", "OpenRiot Update Available",
                    f"v{remote_ver} available. Run openriot-version-check to upgrade."])
    ```

---

### AUDIT FIX 14 — `py3-gobject3` not in `packages.yaml` (GTK welcome screen non-functional) 🟡 P2

**File:** `install/packages.yaml`
**Problem:** `config/bin/openriot-welcome.py` requires `gi` (PyGObject / GTK3 Python bindings). The OpenBSD package is `py3-gobject3`. Without it, the GTK welcome screen always falls back to the shell version.

- [ ] **A14.1** Add `py3-gobject3` to `desktop.apps.packages` in `packages.yaml`.
- [ ] **A14.2** Verify the package name is correct for OpenBSD 7.9 (`pkg_add py3-gobject3`).

---

### AUDIT FIX 15 — Hosting on openriot.org 🟠 P1

**Context:** The curl-pipe install method requires hosting at openriot.org. Update check always returns `-` until this is live.

- [ ] **A15.1** Build final release binary: `make build`
- [ ] **A15.2** Host `openriot` binary at `https://openriot.org/bin/openriot` (OpenBSD amd64)
- [ ] **A15.3** Host `setup.sh` at `https://openriot.org/setup.sh`
- [ ] **A15.4** Host `VERSION` file at `https://openriot.org/VERSION`
- [ ] **A15.5** Verify TLS is working on `openriot.org`
- [ ] **A15.6** Test: `curl -fsSL https://openriot.org/setup.sh | sh` on a clean OpenBSD 7.9 VM

---

### AUDIT FIX 16 — Waybar verification on OpenBSD 🔴 P1

**Context:** Waybar modules were written for ArchRiot and ported over, but never verified on OpenBSD. The `weather-emoji` module uses `stormy` which does not exist on OpenBSD (gracefully degrades to empty). Need to audit every waybar module and script to confirm they work with OpenBSD tools.

- [ ] **A16.1** Audit all waybar scripts in `config/waybar/scripts/` for OpenBSD compatibility:
    - `waybar-cpu.sh` — uses `top(1)`, verify OpenBSD output format matches
    - `waybar-temp.sh` — uses `sysctl hw.sensors`, verify format
    - `waybar-memory.sh` — uses `vmstat`, verify format
    - `waybar-battery.sh` — uses `apm(8)`, verify format
    - `waybar-volume.sh` — uses `sndioctl`, verify format
    - `waybar-network` — uses `ifconfig`, verify format
    - `recording-indicator.sh` — uses `wf-recorder`, verify (needs wf-recorder package)
    - `openriot-update.sh` — uses `curl`, verify
    - `wifi-selector.sh` — uses `ifconfig`, verify
- [ ] **A16.2** Run `make iso` and test on real hardware
- [ ] **A16.3** Confirm waybar launches without errors on OpenBSD

---

### AUDIT FIX 17 — Fuzzel app launcher desktop files 🔴 P1

**Context:** ArchRiot uses wofi/wofi --show=drun for app launching. OpenRiot uses fuzzel (OpenBSD native) which reads XDG `.desktop` files from `~/.local/share/applications/`. Need to port all relevant desktop files from ArchRiot and add OpenBSD-specific apps.

- [ ] **A17.1** Port all OpenBSD-relevant `.desktop` files from ArchRiot `config/applications/` to OpenRiot `config/applications/`:
    - `thunar.desktop` — file manager (OpenBSD has thunar)
    - `firefox.desktop` — browser (already in packages.yaml)
    - `btop.desktop` — system monitor (already in packages.yaml)
    - `fish.desktop` — shell
    - Any others that have OpenBSD equivalents
- [ ] **A17.2** Remove Linux-only desktop files (blueman, blueberry, etc.)
- [ ] **A17.3** Add OpenBSD-specific desktop files:
    - `openriot-screenrecord.desktop` — screen recording (NEW)
    - `openriot-welcome.desktop` — welcome screen
- [ ] **A17.4** Verify `config/applications/` is deployed to `~/.local/share/applications/` via packages.yaml
- [ ] **A17.5** Test fuzzel app launcher on OpenBSD and confirm apps appear

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
    - `setup.sh` — wofi → fuzzel in pkg_add call
    - `config/waybar/scripts/wifi-selector.sh` — wofi → fuzzel in dmenu calls
    - `README.md` — keybinding table updated

3. **New waybar scripts created (all executable, all output valid JSON):**
    - `config/waybar/scripts/waybar-cpu.sh`
    - `config/waybar/scripts/waybar-temp.sh`
    - `config/waybar/scripts/waybar-memory.sh`
    - `config/waybar/scripts/waybar-volume.sh`
    - `config/waybar/scripts/waybar-battery.sh`

---

## What Was Done This Session (May 2025)

### All 24 Bugs from Prompt.md FIXED ✅

1. **BUG 1** — Makefile fallback version: 0.6 → 0.7
2. **BUG 2** — packages.yaml header: 7.8 → 7.9
3. **BUG 3** — packages.yaml: Added btop, slurp, wl-clipboard, playerctl, gnome-text-editor, desktop.media section
4. **BUG 4** — setup.sh: Fixed bare relative paths for fish configs
5. **BUG 5** — setup.sh: Replaced dead swaylock refs with openriot-lock.sh; Added config/bin/ and fonts deployment
6. **BUG 6** — setup.sh: Added openriot binary invocation with Offline/Online distinction
7. **BUG 7** — main.go: Fixed repoDir to use ~/.local/share/openriot with execPath fallback
8. **BUG 8** — monitors.conf: Removed Hyprland env= syntax
9. **BUG 9** — sway/config: Fixed exec_always if/then to sh -c wrapper
10. **BUG 10** — sway/config: Added exec swaybg for wallpaper
11. **BUG 11** — sway/config: Added resume for brightness restore after dim
12. **BUG 12** — swaylock %% format: Reverted — was working on OpenBSD
13. **BUG 13** — keybindings.conf: Ghostty → foot --app-id=floating_foot
14. **BUG 14-17** — Already fixed by BUG 3 (packages added)
15. **BUG 18** — openriot-welcome.sh: Hardcoded v0.4 → dynamic VERSION
16. **BUG 19** — openriot-welcome.py: Hardcoded v0.4 → dynamic VERSION
17. **BUG 20** — openriot-version-check: /proc/PID → os.kill(pid, 0) for OpenBSD
18. **BUG 21** — ModulesWorkspaces: All Hyprland → Sway conversion (single clean default)
19. **BUG 22** — Modules: gnome-system-monitor → foot -e btop (3 places)
20. **BUG 23** — install.site: Removed misleading curl instruction
21. **BUG 24** — TODO.md: Fixed all install/setup.sh → setup.sh references

### Architecture Restructured (ArchRiot-style) ✅

- **setup.sh**: Now minimal bootstrap only
    - Removed ALL manual cp commands (sway, fish, backgrounds, bin, fonts)
    - Detects offline (ISO) vs online (curl) mode
    - Runs openriot --install via exec (replaces process)
    - wlsunset build removed (now in packages.yaml source.builds)

- **openriot --install**: Handles ALL config deployment via packages.yaml
    - CopyConfigs reads packages.yaml patterns
    - Deploys sway/_, waybar/_, fish/_, foot/_, fuzzel/_, mako/_, backgrounds/_, bin/_, fonts/_, btop/themes/_

### New Config Patterns Added to packages.yaml ✅

- `pattern: foot/*`
- `pattern: fuzzel/*`
- `pattern: mako/*`
- `pattern: bin/*` (already existed, confirmed)

### Other Fixes ✅

- **fw_update -a**: Added to packages.yaml system.tools.commands
- **ModulesWorkspaces**: Replaced with clean single sway/workspaces default with running-app icons
- **Brightness resume**: Added to swayidle after dim timeout
- **openriot-lock.sh**: Fixed to deploy to ~/.local/share/openriot/config/bin/ (not ~/.config/sway/)

### Build Verification ✅

- `make dev`: Linux native build passes
- `make verify`: Cross-compile passes
- `./install/openriot --version`: Reports 0.7
- `./install/openriot --crypto ROWML`: Crypto output correct

---

## What Was Done This Session (April 2025)

1. **ISO Build Path Fixes:**
    - `build-iso.sh`: Fixed `git archive --prefix=openriot/` so repo tarball extracts to `~/.local/share/openriot/` (not flat)
    - `autoinstall/install.site`: Removed dead `openriot-HEAD` rename code
    - `autoinstall/install.site`: Replaced hardcoded PKGS with awk parser reading from `/etc/openriot/packages.yaml`
    - `autoinstall/install.site`: Removed hardcoded `OPENRIOT_VERSION`, banner now says "OpenRiot post-install starting"

2. **Dead Code Cleanup:**
    - Removed `source/mullvad/` and `source/windows/` (not needed on OpenBSD/Sway)
    - Removed `--signal`, `--wallet`, `--pomodoro-click`, `--switch-window`, `--fix-offscreen-windows`, `--mullvad-setup` from main.go
    - Removed dead keybindings from `config/sway/keybindings.conf`
    - Deleted `config/sway/swaylock-wrapper.py`, `config/sway/swaylock-wrapper.sh`, `config/sway/swaylock.conf` (replaced by openriot-lock.sh)

3. **New Source Files (was untracked):**
    - `source/audio/volume.go` — `--volume` for media keys
    - `source/backgrounds/backgrounds.go` — `--swaybg-next` for wallpaper cycling
    - `source/detect/detect.go` — dock/undock detection
    - `source/display/display.go` — `--brightness` for brightness keys
    - `source/crypto/crypto.go` — `--crypto`, `--crypto-refresh` (ported from ArchRiot)
    - `source/crypto/trading.go` — trading helpers
    - `config/crypto.toml` — crypto portfolio config
    - `source/go.mod`: Added `github.com/BurntSushi/toml` dependency

4. **Lock Screen (NEW):**
    - Created `config/bin/openriot-lock.sh` — ImageMagick-based swaylock background generator
    - Uses PaperMono font (added to `config/fonts/`) for consistent aesthetic
    - Auto-detects screen resolution via `swaymsg` or `xrandr`
    - Generates time (12hr format), date, crypto holdings, user@host, uptime
    - Crypto refreshes every 30 mins via background loop
    - `swaylock --clock` shows live clock on top of static background
    - Updated `config/sway/config` swayidle to use openriot-lock.sh + swaylock --clock
    - Added PaperMono font deployment to `packages.yaml` (desktop.sway configs + mkdir)
    - Added `ImageMagick` package to `packages.yaml` for OpenBSD

5. **OpenRiot v0.6:**
    - Updated VERSION to 0.6
    - Made VERSION the single source of truth (Makefile and build-iso.sh read from it)
    - Fixed ASCII art to spell OPENRIOT

6. **TUI Sequential Execution:**
    - Rewrote install flow to run sequentially (no goroutines) so progress displays in order
    - Added test mode with 300ms delay between log messages for visibility
    - Added GetModel() function for progress updates

7. **Quit Handling:**
    - Fixed 'q' and Ctrl+C handling to properly exit
    - Added userQuit tracking to avoid double-wait on channels

8. **Copy Path Fixes:**
    - Fixed glob pattern destination to show correct path (stripped wildcard)
    - Logs now show "Copied X -> ~/.config/dir/file" format

9. **Website:**
    - Added Mullvad VPN section to README.md
    - Fixed blockquote CSS (darker background, less padding)

10. **TUI Polish:**
    - Per-package progress in `source/installer/packages.go`
    - Color coding: SUCCESS=green, ERROR=red, WARN=yellow, INFO=dim
    - Window resize handling in `source/tui/model.go`
    - Added `GetModel()` function for progress updates

11. **Waybar enhancements:**
    - `config/bin/waybar-guard.sh` - restarts waybar if crashed
    - `config/bin/battery-monitor.sh` - notifies at 20%/10%
    - `config/bin/openriot-welcome.py` - GTK welcome screen
    - `config/bin/openriot-welcome.sh` - shell fallback

12. **Swaylock enhancements:**
    - Added battery status (apm) bottom-center
    - Added BTC price top-right with 5-min cache

13. **Sway config fixes:**
    - swayidle brightness dim step
    - wlsunset temperature only (no coords)
    - Power menu with Lock/Suspend/Reboot/Shutdown/Logout

14. **Neovim theme:**
    - Changed to One Dark Pro to match Zed

15. **ISO Builder fixes:**
    - Fixed `make iso` - now tries released version first, falls back to snapshot
    - Fixed SHA256 verification - handles 404 gracefully
    - Fixed cleanup trap - only runs on error
    - Changed output to `openriot.iso` (no version in filename)

16. **Misc:**
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

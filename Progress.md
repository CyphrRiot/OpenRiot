# OpenRiot — Project TODO & Progress

> **OpenRiot** transforms a fresh OpenBSD installation into a fully-configured Sway desktop — in one command.
> It is the OpenBSD counterpart to [ArchRiot](https://archriot.org).

---

## Installation Flow (Current Architecture)

```
1. Boot ISO → OpenBSD installer (autoinstall, no interaction needed)
2. install.site (from site79.tgz) runs:
   - Configures doas.conf (permit persist :wheel)
   - Enables services (apmd, sndiod)
   - Writes welcome message to .profile
3. REBOOT
4. User logs in, runs:
   doas pkg_add curl git
   curl -fsSL https://openriot.org/setup.sh | sh
5. setup.sh runs (as user):
   - Configures installurl
   - Updates doas.conf (permit nopass :wheel)
   - Installs curl and git
   - Clones OpenRiot repo to ~/.local/share/openriot/
   - Installs ALL packages via pkg_add
   - Runs openriot --source-builds (builds wlsunset from packages.yaml)
   - Runs openriot --install (configs + commands from packages.yaml)
   - Sets fish as default shell
   - Configures sway autostart in fish config
6. REBOOT → Sway starts automatically
```

---

## Canonical Versions

```
OPENRIOT_VERSION = (from VERSION file, currently 0.98)
OPENBSD_VERSION  = 7.9
ARCH             = amd64
```

**Never hardcode these anywhere. Read from Makefile or VERSION file.**

---

## Key Files

| File                                     | Purpose                                                 |
| ---------------------------------------- | ------------------------------------------------------- |
| `Makefile`                               | Build targets, version info                             |
| `build-iso.sh`                           | Builds bootable ISO                                     |
| `autoinstall/install.conf`               | Autoinstall answers for OpenBSD installer               |
| `autoinstall/install.site`               | Post-install script (runs from site79.tgz)              |
| `autoinstall/autopartitionning.template` | Disk partitioning template                              |
| `install/packages.yaml`                  | **Source of truth** for all packages, configs, commands |
| `install/openriot`                       | Compiled Go binary                                      |
| `setup.sh`                               | Bootstrap script — curl-pipe install                    |
| `source/main.go`                         | Go binary entry point, all CLI flags                    |
| `source/installer/*.go`                  | Package install, config deploy, source builds           |
| `source/config/loader.go`                | YAML parsing                                            |
| `config/`                                | Sway, Waybar, Fish, Foot, Fuzzel configs                |
| `config/backgrounds/`                    | Wallpaper images                                        |
| `site/`                                  | Files extracted to / on target system                   |

---

## Files Deleted / Deprecated

| File                           | Reason                                |
| ------------------------------ | ------------------------------------- |
| `scripts/download-packages.sh` | Packages not bundled in ISO anymore   |
| `scripts/generate-index.sh`    | Not needed — no offline package cache |

---

## Log Format

All output uses consistent 5-character bracket format:

```
[INFO]  Informational message
[OKAY]  Success message
[WARN]  Warning (non-fatal)
[ERR!]  Error (fatal)
```

Used by both `openriot --install` and `setup.sh`.

---

## Log Locations

| Stage                | Log File                        | Description       |
| -------------------- | ------------------------------- | ----------------- |
| `setup.sh`           | `~/.cache/openriot/setup.log`   | All setup output  |
| `openriot --install` | `~/.cache/openriot/install.log` | Config deployment |

Logs are NOT written to `~/.local/share/openriot` — always `~/.cache/openriot`.

---

## Package List (from packages.yaml)

**Core Base:**
`git rsync bc-gh python3 fastfetch jq`

**Shell:**
`fish neovim foot fzf ripgrep htop btop tree fd gnupg meson ninja`

**Sway Desktop:**
`sway waybar fuzzel swaylock swayidle swaybg grim slurp wl-clipboard ImageMagick wf-recorder`

**Applications:**
`firefox flare-messenger tdesktop helix lsd lf`

**Media:**
`playerctl transmission`

**System Tools:**
`curl wget unzip xz isc-dhcp-client`

---

## Source-Built

| Package  | Method                                                   |
| -------- | -------------------------------------------------------- |
| wlsunset | Built by `openriot --source-builds` (from packages.yaml) |

---

## OpenBSD-Specific Tool Replacements

| ArchRiot Tool     | OpenBSD Replacement        | Notes                                |
| ----------------- | -------------------------- | ------------------------------------ |
| `brightnessctl`   | `wsconsctl`                | Console brightness only              |
| `pactl`           | `sndioctl`                 | OpenBSD native audio                 |
| `systemd suspend` | `zzz`                      |                                      |
| `loginctl lock`   | `swaylock -f`              |                                      |
| `NetworkManager`  | `networkmanager` + `nmtui` | Used for WiFi management in OpenRiot |
| `kanshi`          | static `monitors.conf`     | No hotplug on OpenBSD                |
| `wofi`            | `fuzzel`                   | Fuzzel IS available on OpenBSD       |
| `apm` (battery)   | `apm -l` / `-a` / `-m`     |                                      |

---

## NOT PORTED (no OpenBSD equivalent)

- `xdg-desktop-portal-wlr` — not in OpenBSD packages
- `pipewire` / `wireplumber` — not in OpenBSD
- `fcitx5` — input method, not on OpenBSD
- `blueberry` — OpenBSD has no BT stack
- `kanshi` — not in OpenBSD packages

---

## Build Commands

```sh
# Development build (Linux)
make dev

# Cross-compile for OpenBSD
make build

# Smoke test
make verify

# Build ISO
make iso

# Build and test in QEMU
make isotest
```

---

## Testing on OpenBSD

```sh
# After base install and reboot
doas pkg_add curl git
curl -fsSL https://openriot.org/setup.sh | sh
# Reboot — Sway should start automatically
```

---

## openriot CLI Commands

| Command                            | Description                                       |
| ---------------------------------- | ------------------------------------------------- | ------------------ |
| `openriot --install`               | Deploy configs to ~/.config (no packages, no TUI) |
| `openriot --version`               | Show version                                      |
| `openriot --lock`                  | Lock screen (swaylock -f)                         |
| `openriot --suspend`               | Suspend (zzz)                                     |
| `openriot --power-menu`            | Show power menu (fuzzel dmenu)                    |
| `openriot --volume [args]`         | Adjust volume (sndioctl)                          |
| `openriot --brightness [args]`     | Adjust brightness (wsconsctl)                     |
| `openriot --notify "title" "body"` | Send notification                                 |
| `openriot --crypto [BTC            | ETH]`                                             | Show crypto prices |
| `openriot --swaybg-next`           | Cycle wallpaper                                   |

---

## ✅ COMPLETED

- [x] ISO builder (`build-iso.sh` — Linux-compatible, xorriso-only, BIOS+UEFI boot)
- [x] install.site (`autoinstall/install.site` — configures doas, enables services)
- [x] autoinstall config (`autoinstall/install.conf` — unattended OpenBSD install)
- [x] setup.sh (`setup.sh` — orchestrates all root ops, --install/--upgrade modes, version check, preserve logic)
- [x] openriot --install (`source/main.go`, `source/installer/*.go` — config-only, runs as USER)
- [x] Package installation (`setup.sh` — one-by-one pkg_add with -D unsigned)
- [x] Config deployment (`source/installer/configs.go` — glob patterns, preserve_if_exists, identical skip)
- [x] Source builds (`setup.sh` — wlsunset via git clone + meson)
- [x] Canonical versioning (`Makefile`, `VERSION` — single source of truth)
- [x] CLI commands (`source/main.go` — --lock, --suspend, --power-menu, --volume, --brightness, etc.)
- [x] Logging (`setup.sh`, `source/installer/*.go` — logs to ~/.cache/openriot/)
- [x] Disk space check (`setup.sh` — checks ~1GB free before installing packages)
- [x] Keyboard repeat rate (`config/sway/config` — rate 60, delay 120ms)
- [x] Helix config (`config/helix/config.toml` — x remapped to delete_char_forward)
- [x] Background management (`config/backgrounds/` — riot_XX.jpg naming, copyBackgrounds())
- [x] Clock click to cycle backgrounds (`config/waybar/Modules` — on-click to --swaybg-next)
- [x] NetworkManager/nmtui (`packages.yaml` — networkmanager pkg, rcctl enable/start, waybar on-click)
- [x] Upgrade flow (`setup.sh` — git pull if newer version, skip packages if same version)
- [x] Preserve user configs (`packages.yaml` — preserve_if_exists lists for sway, waybar, fish, helix)
- [x] All 18 audit fixes applied

---

## 🔄 IN PROGRESS

### Real Hardware End-to-End Testing

**Issue:** Full ISO → install → `setup.sh` → Sway flow needs validation on actual hardware.

**Files impacted:** `build-iso.sh`, `autoinstall/install.conf`, `autoinstall/install.site`, `setup.sh`

**Status:** Awaiting hardware test results.

---

## Architecture Notes

**ISO (~757MB):**

- OpenBSD base sets
- site79.tgz containing: install.site + packages.yaml + VERSION
- NO packages bundled (downloaded fresh after install)

**Why this architecture:**

- Smaller ISO (~757MB vs 1.1GB+)
- Fresh packages always match current OpenBSD version
- setup.sh handles all complexity after internet is available

**Key design decisions:**

- `openriot --install` runs as USER (not root) — writes to ~/.config without PTY issues
- `setup.sh` handles all root operations via doas
- `doas nopass` means no password prompts after initial setup
- wlsunset built from source by setup.sh (internet available)

---

**Last updated:** May 2025
**Git commit:** ac8fb8f

# OpenRiot — Project TODO & Progress

> **OpenRiot** transforms a fresh OpenBSD installation into a fully-configured Sway desktop — in one command.
> It is the OpenBSD counterpart to [ArchRiot](https://archriot.org).

---

## Installation Flow (Current Architecture)

```
1. Boot ISO → OpenBSD installer (autoinstall, no interaction needed)
2. install.site (from site79.tgz) runs:
   - Extracts repo.tar.gz to ~/.local/share/openriot/
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
   - Removes root-owned openriot directory if present
   - Clones OpenRiot repo to ~/.local/share/openriot/
   - Installs ALL packages via pkg_add
   - Runs setup commands (git config, mkdir, etc.)
   - Builds wlsunset from source
   - Runs openriot --install (config deployment)
   - Sets fish as default shell
   - Configures sway autostart in fish config
6. REBOOT → Sway starts automatically
```

---

## Canonical Versions

```
OPENRIOT_VERSION = (from VERSION file, currently 0.9)
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
| `backgrounds/`                           | Wallpaper images                                        |
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

## Package List (from packages.yaml)

**Core Base:**
`git rsync bc-gh python fastfetch jq`

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

| Package  | Method                                    |
| -------- | ----------------------------------------- |
| wlsunset | Built by `setup.sh` via git clone + meson |

---

## OpenBSD-Specific Tool Replacements

| ArchRiot Tool     | OpenBSD Replacement        | Notes                          |
| ----------------- | -------------------------- | ------------------------------ |
| `brightnessctl`   | `wsconsctl`                | Console brightness only        |
| `pactl`           | `sndioctl`                 | OpenBSD native audio           |
| `systemd suspend` | `zzz`                      |                                |
| `loginctl lock`   | `swaylock -f`              |                                |
| `NetworkManager`  | `ifconfig` + `hostname.if` |                                |
| `kanshi`          | static `monitors.conf`     | No hotplug on OpenBSD          |
| `wofi`            | `fuzzel`                   | Fuzzel IS available on OpenBSD |
| `apm` (battery)   | `apm -l` / `-a` / `-m`     |                                |

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

## Current Status

### ✅ COMPLETED

| Component            | File(s)                                   | Notes                                                         |
| -------------------- | ----------------------------------------- | ------------------------------------------------------------- |
| ISO builder          | `build-iso.sh`                            | Linux-compatible, xorriso-only, BIOS+UEFI boot                |
| install.site         | `autoinstall/install.site`                | Extracts repo, configures doas, enables services              |
| autoinstall config   | `autoinstall/install.conf`                | Unattended OpenBSD install                                    |
| setup.sh             | `setup.sh`                                | Orchestrates all root ops, calls openriot --install as USER   |
| openriot --install   | `source/main.go`, `source/installer/*.go` | Config-only, no TUI, simple fmt.Printf logging                |
| Package installation | `setup.sh` (not openriot)                 | pkg_add with awk-parsed package list                          |
| Config deployment    | `source/installer/configs.go`             | Glob patterns, permission preservation                        |
| Source builds        | `setup.sh`                                | wlsunset via git clone + meson                                |
| Canonical versioning | `Makefile`, `VERSION`                     | Single source of truth                                        |
| CLI commands         | `source/main.go`                          | --lock, --suspend, --power-menu, --volume, --brightness, etc. |

---

## Known Issues

1. **`install.conf` interactive mode** — The OpenBSD installer `I` (interactive) mode may not use `install.conf` the same way autoinstall does. Testing needed.
2. **Real hardware end-to-end testing** — Full ISO → install → `setup.sh` → Sway flow not yet tested on actual hardware.

---

## Audit Fixes Applied (April 2026)

| #   | Issue                             | Status                                  |
| --- | --------------------------------- | --------------------------------------- |
| 1   | configs.go glob recursion         | ✅ Fixed — uses filepath.WalkDir        |
| 2   | Script permissions (0644)         | ✅ Fixed — preserves source permissions |
| 3   | Missing packages in packages.yaml | ✅ Fixed                                |
| 4   | Package verification              | ✅ Fixed                                |
| 5   | doas persist vs nopass            | ✅ Fixed — setup.sh uses nopass         |
| 6   | sway/window module undefined      | ✅ Fixed                                |
| 7   | fw_update -a with doas            | ✅ Removed from packages.yaml           |
| 8   | ImageMagick 6 vs 7                | ✅ Uses `convert` not `magick`          |
| 9   | exec export in sway/config        | ✅ Fixed                                |
| 10  | Screenshot keybinding             | ✅ Fixed                                |
| 11  | wireguard scripts not executable  | ✅ Fixed                                |
| 12  | **pycache** committed             | ✅ Removed                              |
| 13  | Version check headless            | ✅ Fixed                                |
| 14  | py3-gobject3 missing              | ✅ Removed GTK welcome screen           |

---

## Architecture Notes

**ISO (~757MB):**

- OpenBSD base sets
- site79.tgz (~9MB) containing: install.site + repo.tar.gz + packages.yaml + VERSION
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

**Last updated:** April 2025 (post major refactor)
**Git commit:** c0142b0 + ab4baa5

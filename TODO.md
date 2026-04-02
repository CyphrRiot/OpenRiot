# OpenRiot — Project Progress

> **OpenRiot** transforms a fresh OpenBSD installation into a fully-configured Sway desktop — in one command.

---

## What Is OpenRiot?

OpenRiot is the OpenBSD counterpart to [ArchRiot](https://archriot.org). It takes a base OpenBSD install and layers on:

- **Sway** (i3-compatible Wayland compositor)
- **Fish** shell with git prompts
- **Waybar** status bar with modules
- **Neovim** with LazyVim configuration (with optional avante LLM support)
- **Foot** terminal emulator
- **Thunar** file manager

All configuration is declarative, version-controlled, and reproducible.

---

## Workflow Rules — **NEVER DEVIATE**

Do not skip any steps. Reference prior context. Output a numbered plan first, then implement. For every function, include error handling with context.Context and at least two table-driven tests.

### On Every Change

**Every response MUST start with:**

```
Completed: <brief description of what was just done>
Next Task: <description of next task>
```

**Then provide details:**

```
Files: <list of files to be modified>
Goal: <why this change is being made>
```

**Then ask:** `Continue?`

1. **NEVER COMMIT** — Do NOT run `git commit` or `git push` without explicit permission
2. **Propose first** — Show the exact change (filename, function, reason) before editing
3. **Wait for "Proceed/Continue?"** before touching any code
4. **Test locally first** (Linux with `--test` flag)
5. **Verify build passes** (`make build`)
6. **Show proof it works** before asking for approval
7. **One change at a time** — finish one task before starting another

### Before Starting a New Chat

- Re-read relevant section of this TODO
- Check git status for uncommitted changes
- Verify no pending work from previous session

### Version Bumping

1. Update version in `source/main.go` (`var version = "x.x"`)
2. Update README.md badge
3. Commit with `git commit -am "Release vX.X: [brief changes]"`
4. Tag with `git tag -a vX.X -m "Version X.X release: [details]"`
5. Push and push tag

### Git Branches

- `master` — stable, always buildable
- Feature work happens on branches or direct commits with good messages

---

## Architecture: Three Layers

OpenRiot operates in three distinct phases:

### Layer 1: ISO Builder (`build-iso.sh`)

Builds a bootable OpenBSD ISO with:

- Base OpenBSD 7.8 install sets (offline, bundled)
- OpenRiot autoinstall answers pre-filled
- `openriot` binary + `packages.yaml` injected
- `setup.sh` bootstrap script

**Key difference from ArchRiot:** OpenBSD uses an **autoinstall** mechanism (similar to Linux autoinstall/cloud-init) rather than a live ISO with a wizard. The ISO boots, runs install.conf answers automatically, then runs siteXX.tgz scripts.

### Layer 2: Base OpenBSD Install

The ISO's autoinstall flow:

1. Partition and format disk
2. Install base system from bundled sets
3. Extract `site79.tgz` (contains OpenRiot binary + configs)
4. Run `install.site` script for post-install hooks
5. Reboot into fresh OpenBSD with OpenRiot files in place

### Layer 3: Post-Install Setup

On first boot (or via `curl | sh`), the OpenRiot binary:

1. Reads `packages.yaml` for package list
2. Runs `pkg_add` for all packages
3. Copies configs from repo to `~/.config/`
4. Sets Fish as default shell
5. Configures doas
6. Starts Sway

---

## Design Decisions

### OpenBSD 7.8 Stable (NOT -current)

- OpenBSD releases are **stable** — 7.8 is a proper release with 2-year support
- `-current` is the development branch; not suitable for daily driver installs
- The ISO builder should pull from `pub/OpenBSD/7.8/` not `snapshots/`
- Security patches via `syspatch`, package updates via `pkg_add -u`

### Offline Package Bundling

Unlike ArchRiot's ISO (which downloads packages during install), OpenRiot's ISO should:

1. Pre-download all packages from `packages.yaml` to a local directory
2. Serve them via `file://` or include them in the ISO
3. Configure `pkg_add` to use the local repo first, then mirror as fallback

This allows installation on isolated networks (air-gapped, high-security environments).

### OpenBSD Install Flow (vs ArchRiot)

| Step            | ArchRiot              | OpenRiot                  |
| --------------- | --------------------- | ------------------------- |
| Boot            | Live ISO with wizard  | Base ISO with autoinstall |
| Package install | pacman during install | pkg_add on first boot     |
| Config deploy   | During install        | Via openriot binary       |
| Network         | NetworkManager        | iwx/iwm (Intel)           |
| Init            | systemd               | rc.d                      |

### What's NOT Ported from ArchRiot

These ArchRiot features have no OpenBSD equivalent or are unnecessary:

| ArchRiot                    | OpenBSD Reality                      |
| --------------------------- | ------------------------------------ |
| Secure Boot (sbctl/mokutil) | OpenBSD has its own boot security    |
| LUKS encryption             | softraid(4) for full-disk encryption |
| Plymouth                    | OpenBSD boot is simple, no splash    |
| brightnessctl               | No hardware brightness control tool  |
| PipeWire                    | sndiod(1) handles audio natively     |
| NetworkManager              | iwx/iwm + simple config              |
| AUR (paru/yay)              | pkg_add is sufficient                |
| systemd                     | rc.d / rcctl                         |

---

## Current Status

### ✅ COMPLETED

| Component             | File(s)                        | Notes                                                              |
| --------------------- | ------------------------------ | ------------------------------------------------------------------ |
| build-iso.sh          | `build-iso.sh`                 | Linux-compatible; xorriso-only (no sudo); bootable El Torito flags |
| autoinstall config    | `autoinstall/install.conf`     | Works with OpenBSD 7.9                                             |
| install.site          | `autoinstall/install.site`     | Mounts CD, pkg_add offline, doas, rcctl, fish shell, .profile hook |
| packages.yaml         | `install/packages.yaml`        | Source of truth; removed `man` (base system pkg)                   |
| download-packages.sh  | `scripts/download-packages.sh` | POSIX-safe; wayland/ fallback; fetches mirror index once           |
| generate-index.sh     | `scripts/generate-index.sh`    | Auto-run at end of download-packages.sh                            |
| Canonical versioning  | `Makefile`                     | OPENRIOT_VERSION=0.4, OPENBSD_VERSION=7.9 — single source of truth |
| Version ldflags       | `source/main.go`               | Version injected at build time via `-X main.version`               |
| config loader         | `source/config/`               | Reads packages.yaml                                                |
| Go installer skeleton | `source/main.go`               | Handles --version, --test                                          |
| TUI model             | `source/tui/`                  | BubbleTea-based progress display                                   |
| Sway config           | `config/sway/`                 | Ported from ArchRiot                                               |
| Waybar config         | `config/waybar/`               | Ported from ArchRiot                                               |
| Fish config           | `config/fish/`                 | Ported from ArchRiot                                               |
| Neovim config         | `config/nvim/`                 | Ported from ArchRiot                                               |
| Foot config           | `config/foot/`                 | Ported from ArchRiot                                               |
| Backgrounds           | `backgrounds/`                 | 16 CypherRiot backgrounds                                          |

### 🔴 NOT YET STARTED

| Component                     | Priority | Blocking                                            |
| ----------------------------- | -------- | --------------------------------------------------- |
| **Run download-packages.sh**  | **P0**   | Packages must be cached before ISO build            |
| **Build and verify ISO**      | **P0**   | End-to-end test: boot in QEMU, confirm offline pkgs |
| Verify autoinstall runs       | P0       | Requires test ISO boot                              |
| Verify install.site executes  | P0       | Requires test ISO boot                              |
| Verify offline pkg install    | P0       | Disconnect network, retest                          |
| Fix main.go deadlock in test  | P1       | TUI blocks on startup                               |
| Host setup.sh at openriot.org | P1       | Curl install needs hosting                          |
| Test on real OpenBSD hardware | P1       | VM or real hardware                                 |
| wlsunset source build         | P2       | Not in pkg_add                                      |
| Waybar modules (idle, tray)   | P2       | Partially done                                      |

---

## Task Hierarchy

### LAYER 1: ISO Builder

**Goal:** Produce a bootable OpenBSD ISO that installs completely offline with OpenRiot desktop.

**Correct Architecture:**

```
ISO Structure:
  / (root)
    install79.iso contents
    openriot/
      packages/7.9/amd64/     <- Offline packages + index.txt
      site79.tgz              <- install.site + configs + openriot binary
```

**Install Flow:**

1. Boot ISO → autoinstall runs base system install from network
2. After base install → `install.site` runs from `site79.tgz`
3. `install.site` mounts CD and uses `PKG_PATH=/mnt/iso/openriot/packages/7.9/amd64/`
4. Packages install offline → no network needed
5. OpenRiot desktop configured

#### 1.1 ISO Builder Script (DONE ✅)

- [x] 1.1.1 Uses `OPENBSD_VERSION=7.9` (snapshots/current)
- [x] 1.1.2 Uses `snapshots` mirror path
- [x] 1.1.3 ISO_NAME derived from version (`install79.iso`) — no hardcoding
- [x] 1.1.4 Site tarball name derived from version (`site79.tgz`) — no hardcoding
- [x] 1.1.5 SHA256 verified — OS-aware (`sha256sum` Linux, `sha256 -q` OpenBSD)
- [x] 1.1.6 ISO extracted with `xorriso -osirrox` — **no sudo, works on Linux**
- [x] 1.1.7 `install.conf` injected at ISO root
- [x] 1.1.8 `install.site` always injected into `site79.tgz` from `autoinstall/`
- [x] 1.1.9 `site79.tgz` placed at `7.9/amd64/site79.tgz` (correct installer path)
- [x] 1.1.10 Repacked with correct El Torito boot flags (BIOS + UEFI bootable)
- [x] 1.1.11 ISO output: `isos/openriot-0.4.iso`
- [x] 1.1.12 Preflight checks: xorriso, curl, tar, pkg cache, install.conf, install.site

#### 1.2 Offline Package Bundling (DONE ✅)

- [x] 1.2.1 `scripts/download-packages.sh` — complete rewrite
    - Parses `install/packages.yaml` with POSIX awk (no grep -P)
    - Downloads to `~/.pkgcache/7.9/amd64/` (matches build-iso.sh expectation)
    - Fetches mirror HTML index **once**, reuses for all lookups (no per-pkg HTTP)
    - Searches `wayland/` subdirectory as fallback (waybar, etc.)
    - `--dry-run` flag supported
    - Reads `OPENBSD_VERSION`/`ARCH` from env (set by `make`), falls back to defaults
    - Auto-runs `generate-index.sh` on completion (step 4)

- [x] 1.2.2 `scripts/generate-index.sh` — generates `index.txt` from `.tgz` files
    - Required for `pkg_add` dependency resolution
    - Reads `OPENBSD_VERSION`/`ARCH` from env, falls back to defaults

- [x] 1.2.3 Index generation integrated into `download-packages.sh` (auto-runs)

- [x] 1.2.4 `build-iso.sh` copies packages to ISO
    - Cache: `~/.pkgcache/7.9/amd64/*.tgz` + `index.txt`
    - Destination in ISO: `openriot/packages/7.9/amd64/`
    - Preflight fails clearly if cache missing or empty

#### 1.3 install.site Script (DONE ✅)

- [x] 1.3.1 Mounts `/dev/cd0a` → `/mnt/iso`, validates package dir, dies cleanly if missing
- [x] 1.3.2 `PKG_PATH=/mnt/iso/openriot/packages/7.9/amd64` — correct offline path
- [x] 1.3.3 Installs all packages from `packages.yaml` via `pkg_add -v`
- [x] 1.3.4 Unmounts CD cleanly after install
- [x] 1.3.5 Copies `openriot` binary from `/etc/openriot/` (where site79.tgz extracts)
- [x] 1.3.6 Writes `/etc/doas.conf` — `permit nopass :wheel`
- [x] 1.3.7 Enables `apmd` + `sndiod` via `rcctl`
- [x] 1.3.8 Sets fish as default shell for all `/home/*` users
- [x] 1.3.9 Adds `.profile` hook — `openriot --install` runs on first login
- [ ] 1.3.10 **Verify `install.site` actually executes** — requires ISO boot test (1.4)

#### 1.4 Test ISO Build (NEXT — P0 🔴)

**This is the immediate next task. Do not start a new session without completing this.**

**Step-by-step:**

```sh
# 1. Build everything — downloads packages, builds binary, repacks ISO
make iso

# 2. Confirm cache was populated
ls ~/.pkgcache/7.9/amd64/*.tgz | wc -l   # expect ~27+
cat ~/.pkgcache/7.9/amd64/index.txt       # must exist

# 3. Confirm ISO output
ls -lh isos/openriot-0.4.iso              # expect ~900MB+ (762MB base + packages)

# 4. Test boot in QEMU (with network)
qemu-system-x86_64 -cdrom isos/openriot-0.4.iso -m 2G -enable-kvm

# 5. Test offline (most important — -nic none disables all networking)
qemu-system-x86_64 -cdrom isos/openriot-0.4.iso -m 2G -enable-kvm -nic none
```

- [ ] 1.4.1 `make iso` completes — downloads packages, builds binary, repacks ISO in one command
- [ ] 1.4.2 `~/.pkgcache/7.9/amd64/index.txt` exists and has entries
- [ ] 1.4.3 `isos/openriot-0.4.iso` produced
- [ ] 1.4.4 ISO size is larger than base (762MB) — confirms packages were injected
- [ ] 1.4.5 ISO boots in QEMU (BIOS and/or UEFI)
- [ ] 1.4.6 Autoinstall runs unattended (no keyboard input needed)
- [ ] 1.4.7 `install.site` executes post-install (check `/tmp/install.site.log` or serial)
- [ ] 1.4.8 Packages install from CD — **no network required**
- [ ] 1.4.9 Test with `-nic none` in QEMU — full offline install succeeds

---

### LAYER 2: Go Installer (openriot binary)

**Goal:** `openriot` binary handles package install, config deployment, commands from `packages.yaml`.

**Current state:** Basic TUI works in test mode, configs deploy from YAML, commands execute dry-run

#### 2.1 TUI Test Mode (DONE ✅)

- [x] 2.1.1 Logger integrated with TUI (`logger.SetProgram()`, `SetProgramReady()`)
- [x] 2.1.2 TUI renders without deadlock
- [x] 2.1.3 ASCII header "OpenRiot Installer v0.1" displays
- [x] 2.1.4 Logs go to TUI window when program ready
- [x] 2.1.5 Pre-TUI logs go to stdout (correct)

#### 2.2 Config Deployment from YAML (DONE ✅)

- [x] 2.2.1 `source/config/types.go` has `ConfigRule` struct
- [x] 2.2.2 `source/installer/configs.go` reads `module.Configs`
- [x] 2.2.3 Glob patterns work (`fish/*`, `sway/*`, etc.) ]2.2.4`CopyConfigs()`-[xhandles single files and globs
- [x] 2.2.5 Backgrounds copy to `~/.local/share/openriot/backgrounds/`

#### 2.3 Command Execution from YAML (DONE ✅)

- [x] 2.3.1 `source/installer/execcommands.go` created
- [x] 2.3.2 Reads `module.Commands` from all modules
- [x] 2.3.3 `dryRun` flag echoes instead of executing
- [x] 2.3.4 Commands from YAML execute in test mode (dry-run)

#### 2.4 Package Installation from YAML (DONE ✅)

- [x] 2.4.1 `source/config/loader.go` parses `packages.yaml`
- [x] 2.4.2 `config.GetPackages()` returns package list
- [x] 2.4.3 `source/installer/packages.go` uses `pkg_add`
- [x] 2.4.4 `cfg.GetPackages()` wired in `main.go`

#### 2.5 Source Builds (PENDING)

- [ ] 2.5.1 Read `module.Build` commands from Source modules
- [ ] 2.5.2 `wlsunset` requires source build (meson)
- [ ] 2.5.3 Implement in `source/installer/sourcebuild.go`

#### 2.6 TUI Polish (PENDING)

- [ ] 2.6.1 Progress reporting for package install
- [ ] 2.6.2 Log window shows installation steps
- [ ] 2.6.3 Color coding (success/error/warning)
- [ ] 2.6.4 Handle window resize

---

### LAYER 3: First Boot / setup.sh

**Goal:** `curl -fsSL https://openriot.org/setup.sh | sh` for existing OpenBSD systems

#### 3.1 setup.sh (PENDING)

- [ ] 3.1.1 Create `install/setup.sh`
- [ ] 3.1.2 Check OpenBSD version (require 7.9+)
- [ ] 3.1.3 Download `openriot` binary
- [ ] 3.1.4 Download `packages.yaml`
- [ ] 3.1.5 Run `openriot --install`

#### 3.2 Hosting (PENDING)

- [ ] 3.2.1 Host `openriot` binary at `openriot.org/bin/`
- [ ] 3.2.2 Host `setup.sh` at `openriot.org/setup.sh`
- [ ] 3.2.3 TLS configured

---

### Current Status Summary

| Layer | Component                     | Status      |
| ----- | ----------------------------- | ----------- |
| 1.1   | build-iso.sh                  | ✅ DONE     |
| 1.2.1 | download-packages.sh          | ✅ DONE     |
| 1.2.2 | generate-index.sh             | ✅ DONE     |
| 1.2.3 | Integrate index into download | ✅ DONE     |
| 1.2.4 | Copy packages to ISO          | ✅ DONE     |
| 1.3   | install.site                  | ✅ DONE     |
| 1.4   | **Test ISO build**            | 🔴 **NEXT** |
| —     | Canonical versioning          | ✅ DONE     |
| 2.1   | TUI test mode                 | ✅ DONE     |
| 2.2   | Config deployment             | ✅ DONE     |
| 2.3   | Command execution             | ✅ DONE     |
| 2.4   | Package install               | ✅ DONE     |
| 2.5   | Source builds                 | ⬜ TODO     |
| 2.6   | TUI polish                    | ⬜ TODO     |
| 3.1   | setup.sh                      | ⬜ TODO     |
| 3.2   | Hosting                       | ⬜ TODO     |

### ⚠️ Session Notes (important context for next chat)

**What was done this session:**

- `autoinstall/install.site` — full rewrite: mounts CD, offline `pkg_add`, doas, rcctl, fish shell
- `scripts/download-packages.sh` — full rewrite: POSIX-safe, wayland/ fallback, fetches index once, auto-runs generate-index.sh
- `scripts/generate-index.sh` — reads OPENBSD_VERSION from env
- `build-iso.sh` — full Linux-compatible rewrite: `xorriso -osirrox` extraction (no sudo), OS-aware SHA256, correct El Torito BIOS+UEFI boot flags, package injection at Step 7
- `Makefile` — canonical `OPENRIOT_VERSION=0.4`, `OPENBSD_VERSION=7.9`; version injected into Go binary via `-X main.version` ldflags; all scripts read from env with fallback
- `source/main.go` — `var version = "dev"` (injected at build); `var openbsdVersion = "7.9"` added
- `install/packages.yaml` — removed `man` (OpenBSD base system pkg, not installable via pkg_add)

**Known remaining risk:**

- `download-packages.sh` has NOT been run against the live mirror yet — waybar wayland/ fallback and POSIX sed parsing are untested against real mirror HTML
- ISO boot is untested — El Torito flags are correct per `xorriso -report_el_torito` but need QEMU confirmation
- `install.site` path `/etc/openriot/openriot` assumes binary is bundled in `site/etc/openriot/` — verify `site/` dir has correct structure before building ISO

**Before starting next chat — check:**

```sh
git status                    # see what's uncommitted
make verify                   # confirm binary builds and reports version 0.4
ls site/                      # confirm site/ structure for site79.tgz
```

**Resuming next chat — the one command to run:**

```sh
make iso   # self-contained: downloads packages → builds binary → repacks ISO
```

`make iso` dependency chain: `build` → `download-packages` → `build-iso.sh`
`download-packages.sh` is idempotent — already-cached `.tgz` files are skipped.

---

### Key Files

| File                               | Purpose                                       |
| ---------------------------------- | --------------------------------------------- |
| `build-iso.sh`                     | Build bootable ISO                            |
| `scripts/download-packages.sh`     | Download packages for offline                 |
| `scripts/generate-index.sh`        | Generate `index.txt` for repo                 |
| `autoinstall/install.conf`         | Autoinstall answers                           |
| `autoinstall/install.site`         | Post-install script (runs from site79.tgz)    |
| `site/`                            | Files to include in site79.tgz                |
| `install/packages.yaml`            | Source of truth for packages/configs/commands |
| `source/main.go`                   | Go installer entry point                      |
| `source/installer/packages.go`     | Package installation                          |
| `source/installer/configs.go`      | Config deployment                             |
| `source/installer/execcommands.go` | Command execution                             |
| `source/logger/logger.go`          | TUI logging                                   |
| `source/tui/model.go`              | BubbleTea TUI model                           |

---

## OpenBSD Package Reference

### From packages.yaml (source of truth)

```
# Core Base
git rsync bc python3 fastfetch

# Shell & Terminal
fish neovim foot fzf ripgrep wl-clipboard man less htop tree fd lsd

# Sway Desktop
sway waybar wofi swaylock swayidle swaybg grim kanshi xdg-desktop-portal-wlr

# Applications
thunar thunar-archive thunar-volman firefox

# System Tools
doas curl wget unzip xz

# Wireless Firmware
iwx-firmware urtwn-firmware
```

### Source-Built

| Package  | Build Steps                                                                                                                                                            |
| -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| wlsunset | `git clone https://git.sr.ht/~kennylevinsen/wlsunset && cd wlsunset && meson setup build --prefix=/usr/local && meson compile -C build && doas meson install -C build` |

### NOT Available (use alternatives)

| ArchRiot       | OpenBSD Alternative        |
| -------------- | -------------------------- |
| brightnessctl  | None (no hardware control) |
| NetworkManager | iwx/iwm + simple config    |
| PipeWire       | sndiod (in base)           |
| systemd        | rc.d / rcctl               |
| AUR (paru/yay) | pkg_add sufficient         |
| pamixer        | sndctl                     |

---

## File Reference

### ISO Builder

| File                           | Purpose                                         |
| ------------------------------ | ----------------------------------------------- |
| `build-iso.sh`                 | Downloads OpenBSD ISO, injects configs, repacks |
| `autoinstall/install.conf`     | Autoinstall answers for base OpenBSD            |
| `autoinstall/install.site`     | Post-install script (run after base install)    |
| `scripts/download-packages.sh` | Pre-download packages for offline install       |

### Go Installer

| File                           | Purpose                                    |
| ------------------------------ | ------------------------------------------ |
| `source/main.go`               | Entry point, argument parsing, TUI startup |
| `source/installer/packages.go` | Package installation via pkg_add           |
| `source/config/types.go`       | YAML structure types                       |
| `source/config/loader.go`      | YAML loading and validation                |
| `source/tui/model.go`          | BubbleTea TUI model                        |
| `source/tui/messages.go`       | TUI message types                          |
| `source/logger/`               | Logging utilities                          |
| `source/git/`                  | Git configuration helpers                  |

### Configuration

| File                     | Purpose                                          |
| ------------------------ | ------------------------------------------------ |
| `install/packages.yaml`  | **Source of truth** for all packages and configs |
| `install/openriot`       | Compiled binary                                  |
| `config/sway/*`          | Sway compositor config                           |
| `config/waybar/*`        | Waybar status bar config                         |
| `config/fish/*`          | Fish shell config                                |
| `config/nvim/*`          | Neovim/LazyVim config                            |
| `config/foot/*`          | Foot terminal config                             |
| `config/environment.d/*` | Environment variables                            |
| `config/gtk-3.0/*`       | GTK3 theme                                       |
| `config/gtk-4.0/*`       | GTK4 theme                                       |
| `config/mako/*`          | Notification daemon                              |
| `backgrounds/*`          | Wallpaper images                                 |

### Build

| File            | Purpose                     |
| --------------- | --------------------------- |
| `Makefile`      | Build system (`make build`) |
| `source/go.mod` | Go module definition        |
| `source/go.sum` | Go dependencies             |

---

## Step-by-Step Build Plan

### Phase 0: ISO

Infrastructure (IN PROGRESS 🔶)

- [ ] 1.1 Rewrite build-iso.sh for OpenBSD 7.8 stable
- [ ] 1.2 Implement offline package download
- [ ] 1.3 Create install.site post-install script
- [ ] 1.4 Test ISO build and boot

### Phase 1: Go Installer Core (IN PROGRESS 🔶)

- [ ] 2.1 Fix test mode deadlock
- [ ] 2.2 Implement package installation from YAML
- [ ] 2.3 Implement config deployment from YAML
- [ ] 2.4 Implement command execution from YAML
- [ ] 2.5 Verify build passes (`make build`)

### Phase 2: TUI Polish

- [ ] 2.6 Add progress and log display
- [ ] 2.7 Handle resize and input properly

### Phase 3: First Boot Integration

- [ ] 3.1 Create and host setup.sh
- [ ] 3.2 Configure post-install services

### Phase 4: Testing & Polish

- [ ] Test on real OpenBSD hardware
- [ ] Verify WiFi works
- [ ] Test offline install (no network)

### Phase 5: Release

- [ ] Build final ISO
- [ ] Host setup.sh and binary
- [ ] Announce

---

## Key Commands

### Building the OpenRiot Binary

```sh
make build          # Production build
make dev            # Development build (faster)
make release        # Release build
```

### Testing

```sh
./install/openriot --test          # Test mode (Linux)
./install/openriot --version       # Check version
```

### Building ISO

```sh
./build-iso.sh                      # Build ISO (requires OpenBSD or Linux)
```

### Manual Package Download (for offline ISO)

```sh
./scripts/download-packages.sh      # Download all packages to .pkgcache/
```

---

## Known Issues

1. **ISO untested** — Scripts are complete but ISO has not been built and booted end-to-end
2. **download-packages.sh untested against live mirror** — POSIX sed parsing and wayland/ fallback need real-world confirmation
3. **install.site binary path unverified** — Assumes `site/etc/openriot/openriot` exists; `site/` directory may be empty
4. **Test mode deadlock** — TUI blocks on startup (P1, not blocking ISO work)
5. **setup.sh not hosted** — Curl install not available (P1, post-ISO)

---

## Credits

OpenRiot is a port of [ArchRiot](https://archriot.org) to OpenBSD.
OpenBSD is developed by the [OpenBSD Foundation](https://www.openbsd.org).

## License

MIT License — see [LICENSE](./LICENSE)

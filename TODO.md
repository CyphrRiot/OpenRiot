# OpenRiot — Project Progress

> **OpenRiot** transforms a fresh OpenBSD installation into a fully-configured Sway desktop — in one command.

---

## What Is OpenRiot?

OpenRiot is the OpenBSD counterpart to [ArchRiot](https://archriot.org). It takes a base OpenBSD install and layers on:

- **Sway** (i3-compatible Wayland compositor)
- **Fish** shell with git prompts
- **Waybar** status bar with modules
- **Neovim** with LazyVim configuration
- **Foot** terminal emulator
- **Thunar** file manager

All configuration is declarative, version-controlled, and reproducible.

---

## Workflow Rules — _NEVER DEVIATE_

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

1. **Propose first** — Show the exact change (filename, function, reason) before editing
2. **Wait for "Proceed/Continue?"** before touching any code
3. **Test locally first** (Linux with `--test` flag)
4. **Verify build passes** (`make build`)
5. **Show proof it works** before asking for approval
6. **One change at a time** — finish one task before starting another

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

| Component             | File(s)                    | Notes                              |
| --------------------- | -------------------------- | ---------------------------------- |
| build-iso.sh skeleton | `build-iso.sh`             | Downloads ISO, injects autoinstall |
| autoinstall config    | `autoinstall/install.conf` | Works with OpenBSD 7.8             |
| packages.yaml         | `install/packages.yaml`    | Source of truth for pkg_add        |
| config loader         | `source/config/`           | Reads packages.yaml                |
| Go installer skeleton | `source/main.go`           | Handles --version, --test          |
| TUI model             | `source/tui/`              | BubbleTea-based progress display   |
| Sway config           | `config/sway/`             | Ported from ArchRiot               |
| Waybar config         | `config/waybar/`           | Ported from ArchRiot               |
| Fish config           | `config/fish/`             | Ported from ArchRiot               |
| Neovim config         | `config/nvim/`             | Ported from ArchRiot               |
| Foot config           | `config/foot/`             | Ported from ArchRiot               |
| Backgrounds           | `backgrounds/`             | 16 CypherRiot backgrounds          |
| Makefile              | `Makefile`                 | Builds openriot binary             |

### 🔴 NOT YET STARTED

| Component                                           | Priority | Blocking                      |
| --------------------------------------------------- | -------- | ----------------------------- |
| Rewrite build-iso.sh for offline packages           | P0       | Need stable 7.8, not -current |
| Fix build-iso.sh to use OpenBSD 7.8 (not snapshots) | P0       | Current script pulls -current |
| Implement offline package download                  | P0       | ISO must work offline         |
| Port install.site script                            | P0       | Post-install hooks            |
| Fix main.go deadlock in test mode                   | P1       | TUI blocks on startup         |
| Host setup.sh at openriot.org                       | P1       | Curl install needs hosting    |
| Test on real OpenBSD hardware                       | P1       | VM or real hardware           |
| OpenBSD package verification                        | P2       | Ensure all packages exist     |
| wlsunset source build                               | P2       | Not in pkg_add                |
| Waybar modules (idle, tray)                         | P2       | Partially done                |

---

## Task Hierarchy

### LAYER 1: ISO Builder

#### 1.1 Rewrite build-iso.sh for OpenBSD 7.8 Stable

**Problem:** Current `build-iso.sh` pulls from `snapshots/` (OpenBSD -current).

**Sub-tasks:**

- [ ] 1.1.1 Change `OPENBSD_VERSION` from `7.9` to `7.8`
- [ ] 1.1.2 Change `MIRROR` from `cdn.openbsd.org/pub/OpenBSD/snapshots` to `cdn.openbsd.org/pub/OpenBSD/7.8`
- [ ] 1.1.3 Update ISO_NAME from `install79.iso` to `install78.iso`
- [ ] 1.1.4 Update site tarball from `site79.tgz` to `site78.tgz`
- [ ] 1.1.5 Verify SHA256 format for 7.8 release matches expected

#### 1.2 Implement Offline Package Bundling

**Problem:** OpenRiot ISO should work offline (like ArchRiot's).

**Sub-tasks:**

- [ ] 1.2.1 Create `scripts/download-packages.sh`
    - Read packages from `install/packages.yaml`
    - Download each package from OpenBSD mirror to `.pkgcache/`
    - Create local `packages.txt` index file
    - Output: `.pkgcache/` directory + `packages.txt` manifest

- [ ] 1.2.2 Modify `build-iso.sh` to include packages
    - Copy `.pkgcache/` to `ISO_CONTENTS/opt/openriot-packages/`
    - Create `pkg_add.conf` pointing to local repo first
    - Ensure sets list includes local package path

- [ ] 1.2.3 Test offline install in VM
    - Disconnect network in VM
    - Boot ISO
    - Verify packages install from local cache

#### 1.3 Autoinstall Enhancement

**Sub-tasks:**

- [ ] 1.3.1 Update `autoinstall/install.conf` for 7.8
    - Set correct HTTP server path for 7.8 sets
    - Add any 7.8-specific answers

- [ ] 1.3.2 Create `autoinstall/install.site` (post-install script)
    - Extract openriot binary to `/usr/local/bin/`
    - Copy packages.yaml to `/opt/openriot/`
    - Set up initial doas configuration
    - Create `~/.config/` skeleton if needed
    - Install packages via `/usr/local/bin/openriot --install`

- [ ] 1.3.3 Inject install.site into ISO
    - Place in ISO root as `install.site`
    - OpenBSD autoinstall runs this after base install

#### 1.4 ISO Build Testing

**Sub-tasks:**

- [ ] 1.4.1 Test build-iso.sh completes without errors
- [ ] 1.4.2 Verify ISO boots in QEMU/VM
- [ ] 1.4.3 Verify autoinstall runs without prompts
- [ ] 1.4.4 Verify post-install script runs
- [ ] 1.4.5 Test with Intel WiFi (iwx) firmware

---

### LAYER 2: Go Installer (openriot binary)

#### 2.1 Fix Test Mode Deadlock

**Problem:** `main.go` deadlocks in test mode because `program.Send()` is called after `program.Run()`.

**Sub-tasks:**

- [ ] 2.1.1 Remove `program.Send(tui.OpenRouterConfirmMsg(true))` after `program.Run()`
- [ ] 2.1.2 Ensure TUI starts in initial confirmation state
- [ ] 2.1.3 Verify `--test` flag works without deadlock

#### 2.2 Implement Package Installation from YAML

**Problem:** `main.go` has hardcoded package list instead of reading from `packages.yaml`.

**Sub-tasks:**

- [ ] 2.2.1 Already partially done: `config.FindConfigFile()` and `config.LoadConfig()` exist
- [ ] 2.2.2 Verify `cfg.GetPackages()` returns correct list
- [ ] 2.2.3 Connect package installation to orchestrator flow
- [ ] 2.2.4 Add progress reporting for package install

#### 2.3 Implement Config Deployment from YAML

**Problem:** Config deployment hardcodes paths instead of reading from `packages.yaml`.

**Sub-tasks:**

- [ ] 2.3.1 Read `module.Configs` from each module in `packages.yaml`
- [ ] 2.3.2 Implement `CopyConfigs()` similar to ArchRiot's `configs.go`
- [ ] 2.3.3 Support `preserveIfExists` for user configs
- [ ] 2.3.4 Support custom `target` paths

#### 2.4 Implement Command Execution from YAML

**Problem:** Some modules have `commands` (e.g., doas config, chsh).

**Sub-tasks:**

- [ ] 2.4.1 Read `module.Commands` from each module
- [ ] 2.4.2 Execute commands in dependency order
- [ ] 2.4.3 Handle command failures gracefully (log but continue)
- [ ] 2.4.4 Skip commands that require root in test mode

#### 2.5 Implement Source Builds from YAML

**Problem:** `wlsunset` requires building from source.

**Sub-tasks:**

- [ ] 2.5.1 Read `module.Build` commands from Source modules
- [ ] 2.5.2 Implement source build runner
- [ ] 2.5.3 Clone, build, install wlsunset
- [ ] 2.5.4 Report build progress to TUI

#### 2.6 Port TUI Enhancements

**Sub-tasks:**

- [ ] 2.6.1 Add step progress (like ArchRiot's "Installing desktop...")
- [ ] 2.6.2 Add log window scrolling
- [ ] 2.6.3 Add color coding (success green, error red)
- [ ] 2.6.4 Handle window resize
- [ ] 2.6.5 Add confirmation prompt on start

#### 2.7 Port Waybar Integration

**Sub-tasks:**

- [ ] 2.7.1 Port ArchRiot's waybar module protocols
- [ ] 2.7.2 Implement waybar restart command
- [ ] 2.7.3 Implement idle detection module
- [ ] 2.7.4 Implement tray module

---

### LAYER 3: First Boot Experience

#### 3.1 setup.sh Bootstrap

**Problem:** Need `curl | sh` installation for existing OpenBSD systems.

**Sub-tasks:**

- [ ] 3.1.1 Create `install/setup.sh`
    - Check OpenBSD version (require 7.8+)
    - Detect architecture (amd64)
    - Download appropriate `openriot` binary
    - Download `packages.yaml`
    - Run `openriot --install`

- [ ] 3.1.2 Host `openriot` binary at `https://openriot.org/bin/openriot-0.1-openbsd-amd64`
- [ ] 3.1.3 Host `setup.sh` at `https://openriot.org/setup.sh`
- [ ] 3.1.4 Test curl install on existing OpenBSD

#### 3.2 Post-Install Services

**Sub-tasks:**

- [ ] 3.2.1 Configure apmd (`rcctl enable apmd`)
- [ ] 3.2.2 Configure sndiod (usually already default)
- [ ] 3.2.3 Configure libui for waybar
- [ ] 3.2.4 Set up waybar startup in Sway

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

1. **build-iso.sh uses -current** — Must rewrite to use 7.8 stable
2. **Offline packages not implemented** — ISO requires network
3. **Test mode deadlock** — TUI blocks on startup
4. **install.site not created** — Post-install script missing
5. **setup.sh not hosted** — Curl install not available

---

## Credits

OpenRiot is a port of [ArchRiot](https://archriot.org) to OpenBSD.
OpenBSD is developed by the [OpenBSD Foundation](https://www.openbsd.org).

## License

MIT License — see [LICENSE](./LICENSE)

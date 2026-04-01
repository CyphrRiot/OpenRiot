# OpenRiot — Project Progress

> **Status: IN PROGRESS — Not yet installable**
> **Version: 0.1**
> **Last Updated: 2026-03-31**

---

## What Is OpenRiot?

OpenRiot is a polished, opinionated **OpenBSD desktop system** built in the spirit of [ArchRiot](https://github.com/CyphrRiot/ArchRiot). The goal is a one-command post-install setup (`curl -fsSL https://openriot.org/setup.sh | sh`) that transforms a base OpenBSD installation into a fully-configured Sway + Waybar desktop — same CypherRiot aesthetics, same keybindings, same workflow — running on OpenBSD's audited, pledge/unveil-secured base.

**Philosophy:**

- ArchRiot = rolling release, bleeding-edge Hyprland, Arch Linux base
- OpenRiot = stable OpenBSD base, Sway (i3-compatible Wayland compositor), same rice

**Target hardware:** Any desktop or laptop, especially the ThinkPad X1 Carbon Gen 13
**Target users:** Intermediate Linux/BSD users who want OpenBSD's security with ArchRiot's workflow
**Time to desktop:** ~15 minutes after base install

---

## Workflow Rules

### On Every Change

1. **Propose before executing** — use this exact format:

    Completed: {completed task}
    Next Task: {description}
    Files: {list of files we will touch}
    Goal: {why we are doing it}

    Continue? [Y/n]

2. **Wait for confirmation** — do not run commands without approval
3. **One step at a time** — never combine or skip steps
4. Update `TODO.md` — reflect exactly what changed
5. Update `README.md` — if the change affects user-facing instructions
6. **Verify against ArchRiot/local system** — before adding packages or dependencies, check `~/Code/ArchRiot/` and the local running system to confirm actual need; don't install based on documentation alone

**OpenBSD-specific rules:**

- WiFi adapters: Only use adapters confirmed working in README.md (iwx for Intel AX, urtwn for RTL8188CU/EU, athn for Atheros AR9271)
- No native Bluetooth on OpenBSD — document workarounds
- No starship — ArchRiot has it installed but doesn't use it (custom fish_prompt used instead)
- Use Firefox as primary browser (not Chromium/Ungoogled unless specifically needed)

7. NEVER, EVER, NEVER COMMIT WITHOUT ASKING!!!!!

### Before Starting a New Chat

Copy `TODO.md` into the new context so work can resume immediately.

### Version Bumping

- Version format: `0.1`, `0.2`, ..., `0.9`, `1.0`, `1.1`, etc.
- Bump when: Phase 2 complete, Phase 3 complete, etc.
- Update `OPENRIOT_VERSION` in `build-iso.sh` AND `README.md` badges

### Git Branches

- `main` — stable, tested work
- Branch for: Phase 2 (setup.sh), Phase 3 (Go port), etc.

---

## Architecture: Three Layers

```
Layer 1: ISO Builder        → Produces bootable OpenBSD + autoinstall ISO
Layer 2: Base Install      → OpenBSD autoinstall (unattended or guided, disk always prompts)
Layer 3: setup.sh + Go CLI → Packages, Sway, Waybar, Fish, dotfiles
```

### Layer 1: ISO Builder (`build-iso.sh`)

**Location:** `build-iso.sh`

Downloads the OpenBSD install ISO, injects an `install.conf` (autoinstall answers) and optionally a `site79.tgz` (custom files overlay), then repacks into a bootable ISO using `xorriso` or `mkisofs`.

**Flow:**

```
OpenBSD snapshots/amd64/install79.iso
    ↓ (download, verify SHA256 against OpenBSD's published hash)
    ↓ (inject autoinstall/install.conf)
    ↓ (optional: build site79.tgz from site/ dir and inject)
    ↓ (xorriso repack)
isos/openriot-VERSION-OPENBSD_VERSION.iso
```

**Key behavior:**

- Downloads from `https://cdn.openbsd.org/pub/OpenBSD/snapshots/amd64/install79.iso` (rolling -current)
- SHA256 verification parses OpenBSD's `SHA256 (file) = hash` format correctly
- Caches downloaded ISO to `.work/dl/`
- Injects `site79.tgz` only if `site/` directory is non-empty (skips silently if empty)
- Cleanup trap removes `.work/` dir on exit

### Layer 2: Base OpenBSD Install

The ISO boots into the standard OpenBSD installer. When `(A)utoinstall` is selected, the installer reads `install.conf` from the CD root and pre-fills all answers. Disk selection **always prompts** to prevent accidental data loss.

**What the base install gets you:**

- OpenBSD -current (7.9 snapshots) base system
- All sets (base, comp, games, man, xbase, xfont, xserv, xshare)
- User account with wheel group
- Network configured for HTTP install from cdn.openbsd.org

### Layer 3: Post-Install Setup

After base OpenBSD is installed and rebooted, run:

```sh
curl -fsSL https://openriot.org/setup.sh | sh
```

This installs and configures all desktop packages and dotfiles.

---

## Pending Questions (Answered ✅)

1. ✅ SoftRAID encryption - possible post-install via bioctl but recommended at installation time
2. ✅ LLM/OpenRouter.ai - added avante.nvim to neovim config with OpenRouter provider
3. ✅ Waybar vs alternatives - Keep Waybar (native for Sway, already configured)
4. ✅ Bundling packages - Not needed; pkg_add fetches from OpenBSD mirrors
5. ✅ openbsd.app search - Added to README.md with pkg_info -Q instructions

## Pending Tasks

- [ ] **setup.sh**: Add optional OpenRouter API Key prompt during install. If user says "Yes", prompt for API key and add `export OPENROUTER_API_KEY="..."` to `~/.config/fish/config.env` or similar fish environment file.

## Current Status

### ✅ COMPLETED — Phase 0 and Phase 1

| Item                | File(s)                                             | Notes                                                     |
| ------------------- | --------------------------------------------------- | --------------------------------------------------------- |
| ISO builder         | `build-iso.sh`                                      | Tested, working, SHA256 verified                          |
| Autoinstall config  | `autoinstall/install.conf`                          | Asks hostname/password/user/timezone; disk always prompts |
| site79.tgz support  | `site/` dir + build-iso.sh step 4                   | Skips when dir empty; ready for custom files              |
| Sway config         | `config/sway/`                                      | Ported from `~/.config/sway` with OpenBSD fixes           |
| swaylock-wrapper.py | `config/sway/swaylock-wrapper.py`                   | Standalone rewrite; no ArchRiot dependency                |
| Backgrounds         | `backgrounds/`                                      | 13 riot\_\*.jpg files from ArchRiot                       |
| Jekyll site         | `_layouts/`, `_config.yml`,                         |
| `assets/css/`       | Midnight theme, CypherRiot CSS                      |
| README              | `README.md`                                         | Badges, IN PROGRESS warning, install methods, TOC         |
| CNAME               | `CNAME`                                             | openriot.org                                              |
| Blowfish emoji      | `_layouts/default.html`, `README.md`, `_config.yml` | Replaced 🎭 with 🐡                                       |
| TODO.md             | `TODO.md`                                           | This document                                             |
| ISO output          | `isos/`                                             | `openriot-V.v.iso`                                        |

### ✅ COMPLETED — Phase 2

| Item                             | Priority | Blocking | Notes                              |
| -------------------------------- | -------- | -------- | ---------------------------------- |
| `setup.sh` script                | ✅ DONE  | No       | See `install/setup.sh`             |
| Host `setup.sh` at openriot.org  | P0       | Yes      | Domain not yet serving the script  |
| Test `setup.sh` on real OpenBSD  | P0       | No       | Needed before first release        |
| Fix any package name differences | P1       | No       | Found during real hardware testing |

### 🔴 NOT YET STARTED — Phase 3 (Go Installer Port)

| Item                  | Priority | Blocking | Notes                                  |
| --------------------- | -------- | -------- | -------------------------------------- |
| Go Installer Port     | P0       | Yes      | Port ArchRiot Go CLI to OpenBSD        |
| `install/packages.go` | P1       | No       | Verify pkg_add integration works       |
| Waybar modules        | P1       | No       | Port cpu, memory, volume from ArchRiot |
| Go control panel      | P2       | No       | GTK4 app not yet ported                |

### ✅ COMPLETED — Phase 4

| Item                        | Priority | Blocking | Notes                                         |
| --------------------------- | -------- | -------- | --------------------------------------------- |
| Fish shell config           | ✅ DONE  | No       | See `config/fish/`                            |
| packages.yaml (OpenBSD)     | ✅ DONE  | No       | See `install/packages.yaml`                   |
| Neovim config               | ✅ DONE  | No       | See `config/nvim/`                            |
| btop config                 | ✅ DONE  | No       | See `config/btop/`                            |
| fastfetch config            | ✅ DONE  | No       | See `config/fastfetch/`                       |
| waybar modules              | ✅ DONE  | No       | See `config/waybar/`                          |
| mako config                 | ✅ DONE  | No       | See `config/mako/`                            |
| GTK themes (gtk-3.0/4.0)    | ✅ DONE  | No       | See `config/gtk-3.0/` and `config/gtk-4.0/`   |
| environment.d               | ✅ DONE  | No       | See `config/environment.d/`                   |
| Thunar config               | ✅ DONE  | No       | See `config/Thunar/`                          |
| `site/` populated files     | P1       | No       | doas.conf, pkg_add.conf, hostname, etc.       |
| Swaylock dynamic wallpaper  | P2       | No       | Requires backgrounds + openriot binary        |
| OpenRiot wallpapers package | P2       | No       | Full CypherRiot backgrounds not yet assembled |

---

## File Reference

### ISO Builder

| File                       | Purpose                                                                                                    |
| -------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `build-iso.sh`             | Main build script; downloads ISO, verifies SHA256, injects install.conf + site79.tgz, repacks with xorriso |
| `autoinstall/install.conf` | Autoinstall answer file; hostname/password/user/timezone = ask; sets location = http from cdn.openbsd.org  |
| `site/`                    | Directory for custom files to overlay on target system; populate before building ISO                       |
| `isos/`                    | Output directory for built ISOs                                                                            |

### Sway Config (ported from `~/.config/sway`)

| File                              | Purpose                                                                                                    |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `config/sway/config`              | Main Sway config; sources monitors, windowrules, keybindings; sets $terminal=foot, $archriot=openriot path |
| `config/sway/keybindings.conf`    | All keybindings (bindsym); ArchRiot binary paths replaced with openriot paths                              |
| `config/sway/monitors.conf`       | Monitor/workspace config (empty defaults)                                                                  |
| `config/sway/windowrules.conf`    | for_window rules (largely compatible with Sway)                                                            |
| `config/sway/swayidle.conf`       | Idle/lock configuration (reference; active config is inline in config)                                     |
| `config/sway/swaylock.conf`       | Swaylock config (uses /tmp/swaylock-bg.png)                                                                |
| `config/sway/swaylock-wrapper.sh` | Calls swaylock-wrapper.py to generate lock background                                                      |
| `config/sway/swaylock-wrapper.py` | Generates lock screen image with time/date/user/host; standalone, no ArchRiot dependency                   |

**OpenBSD fixes applied to Sway config:**

| Original                                                 | Problem                      | Fix                                                       |
| -------------------------------------------------------- | ---------------------------- | --------------------------------------------------------- |
| `$terminal = ghostty`                                    | ghostty not in OpenBSD pkg   | Changed to `foot`                                         |
| `$browser = archriot-brave`                              | Brave not in OpenBSD pkg     | ✅ Replaced with Firefox (firefox pkg_add)                |
| `$archriot $HOME/.local/share/archriot/install/archriot` | Wrong path                   | Changed to `$HOME/.local/share/openriot/install/openriot` |
| `exec gsettings set org.gnome.desktop.interface...`      | No gsettings on OpenBSD      | Commented out                                             |
| `exec-once = i3-autotiling`                              | Not in OpenBSD pkg repo      | Commented out                                             |
| `swayidle: brightnessctl`                                | brightnessctl not on OpenBSD | Removed                                                   |
| `swayidle: hyprlock`                                     | Not on OpenBSD               | Replaced with `swaylock -f`                               |
| `swayidle: loginctl lock-session`                        | Not on OpenBSD               | Removed                                                   |

### Backgrounds

| File                                      | Purpose                                                   |
| ----------------------------------------- | --------------------------------------------------------- |
| `backgrounds/riot_01.jpg` – `riot_13.jpg` | CypherRiot-themed wallpapers; used by swaylock-wrapper.py |

### Jekyll Site

| File                    | Purpose                                                       |
| ----------------------- | ------------------------------------------------------------- |
| `_config.yml`           | Jekyll config; Midnight theme, OpenRiot title/description     |
| `_layouts/default.html` | Page layout; Midnight CSS overrides, footer (🐡 blowfish), JS |
| `assets/css/style.scss` | Custom CSS; imports Midnight theme then CypherRiot overrides  |
| `README.md`             | GitHub Pages landing page                                     |
| `CNAME`                 | Domain: openriot.org                                          |

### Go Installer

| File                     | Purpose                                                                       |
| ------------------------ | ----------------------------------------------------------------------------- |
| `source/`                | Go source code (bubbletea TUI framework, minimal port from ArchRiot)          |
| `source/tui/model.go`    | Main TUI model — progress display, input handling, log window                 |
| `source/tui/messages.go` | TUI message types (Init, Tick, Done, Error)                                   |
| `source/main.go`         | Entry point — handles --version flag, starts TUI program                      |
| `source/go.mod`          | Go module definition; requires bubbletea + lipgloss                           |
| `install/openriot`       | Compiled OpenBSD binary (built via `make`)                                    |
| `Makefile`               | Build system — `make build` compiles source/ → install/openriot, GOOS=openbsd |

---

## Step-by-Step Build Plan

### Phase 0: Infrastructure (DONE ✅)

- [x] Write `build-iso.sh`
- [x] Fix SHA256 verification (parse OpenBSD's `SHA256 (file) = hash` format)
- [x] Create `autoinstall/install.conf`
- [x] Add `site79.tgz` injection step (skips when site/ empty)
- [x] Test build: `isos/openriot-V.v.iso` produced
- [x] Version set to `0.1`

### Phase 1: Sway Config (DONE ✅)

- [x] Copy `~/.config/sway/` → `config/sway/`
- [x] Fix ghostty → foot
- [x] Fix ArchRiot binary paths → openriot paths
- [x] Comment out gsettings, i3-autotiling, brightnessctl, hyprlock, loginctl
- [x] Fix swayidle exec line (use swaylock instead of hyprlock)
- [x] Fix swaylock-wrapper.py (standalone, no ArchRiot dependency)
- [x] Copy 13 backgrounds from ArchRiot → `backgrounds/`
- [x] Commit all Sway config files

### Phase 2: setup.sh Bootstrap (DONE ✅)

- [x] Write `setup.sh` (the `curl | sh` bootstrap) — see `install/setup.sh`
    - Check OpenBSD version (require 7.8+)
    - Install packages via `pkg_add` (verified against ArchRiot; no starship)
    - Clone dotfiles or link from repo
    - Deploy Sway + Waybar configs
    - Set Fish as default shell
    - Configure doas (passwordless wheel)
    - Start Sway
- [ ] Host `setup.sh` at `https://openriot.org/setup.sh` ⬜
- [ ] Test `setup.sh` on real OpenBSD installation ⬜
- [ ] Fix any package name differences discovered ⬜

### Phase 3: Go Installer Port (in progress 🔶)

Based on ArchRiot analysis (~84 Go files → only 5 exist in OpenRiot). Port only what's relevant for OpenBSD.

**Priorities:**

- [ ] `install/packages.go` — Already exists; verify pkg_add integration works
- [ ] `tui/model.go` — Already exists; enhance for OpenBSD-specific workflow
- [ ] `source/waybar/` — Port relevant modules (cpu, memory, volume)
- [ ] `source/tools/` — Port basic diagnostics (OpenBSD has no brightnessctl)
- [ ] `source/backgrounds/` — Port swaybg wallpaper cycling
- [ ] `source/display/` — Skip (kanshi handles this natively)

**Not needed for OpenBSD:**

- Secure Boot, LUKS encryption (OpenBSD has its own mechanisms)
- Signal, Telegram, Trezor (not in OpenBSD packages)
- Crypto wallets (not in scope)
- Plymouth (boot loader, OpenBSD different)

**Commands to support:**

- `--version` / `-v` ✅ exists
- `--install` ✅ basic package install exists
- `--waybar-restart` — useful for Sway
- `--idle-diagnostics` — could port
- `--crypto-refresh` — ❌ skip (no crypto module needed)

### Phase 4: Full Desktop Integration (P1) ✅

- [x] Write `install/packages.yaml` (pkg_add package list) — see `install/packages.yaml`
- [x] Copy/configure Waybar modules from ArchRiot — see `config/waybar/`
- [x] Copy/configure Fish shell config from ArchRiot — see `config/fish/`
- [x] Copy/configure Neovim config from ArchRiot — see `config/nvim/`
- [x] Copy/configure btop config from ArchRiot — see `config/btop/`
- [x] Copy/configure fastfetch config from ArchRiot — see `config/fastfetch/`
- [x] Copy/configure mako config from ArchRiot — see `config/mako/`
- [x] Copy/configure GTK themes from ArchRiot — see `config/gtk-3.0/`, `config/gtk-4.0/`
- [x] Copy/configure Thunar config from ArchRiot — see `config/Thunar/`
- [x] Copy/configure environment.d from ArchRiot — see `config/environment.d/`
- [ ] Populate `site/` with useful files (doas.conf, pkg_add.conf, hostname)

### Phase 5: Testing & Polish (P2)

- [ ] Test on real hardware (ThinkPad X1 Carbon Gen 13)
- [ ] Test WiFi (AX211 working; BE201 fallback)
- [ ] Port Go control panel to OpenBSD
- [ ] Swaylock dynamic wallpaper with backgrounds
- [ ] Create OpenRiot wallpapers package

### Phase 6: Release (P3)

- [ ] Build final ISO with setup.sh included
- [ ] Create openriot.org/releases page
- [ ] Announce on misc@openbsd.org

---

## OpenBSD Package List

Packages to install via `pkg_add`. Derived from ArchRiot's packages.yaml, translated to OpenBSD equivalents.

### Core Base

| Package     | Description        | ArchRiot Equivalent |
| ----------- | ------------------ | ------------------- |
| `git`       | Version control    | git                 |
| `rsync`     | File sync          | rsync               |
| `bc`        | Calculator         | bc                  |
| `python`    | Python interpreter | python              |
| `fastfetch` | System info tool   | fastfetch           |

### Shell & Terminal

| Package        | Description      | ArchRiot Equivalent |
| -------------- | ---------------- | ------------------- |
| `fish`         | Fish shell       | fish                |
| `neovim`       | Text editor      | neovim              |
| `foot`         | Wayland terminal | kitty               |
| `fd`           | File finder      | fd                  |
| `fzf`          | Fuzzy finder     | fzf                 |
| `ripgrep`      | Search tool      | ripgrep             |
| `wl-clipboard` | Clipboard        | wl-clipboard        |
| `man`          | Manual pages     | man                 |
| `less`         | Pager            | less                |
| `htop`         | Process viewer   | htop                |
| `tree`         | Dir listing      | tree                |

### Desktop (Sway)

| Package                      | Description        | ArchRiot Equivalent         |
| ---------------------------- | ------------------ | --------------------------- |
| `sway`                       | Wayland compositor | hyprland                    |
| `waybar`                     | Status bar         | waybar                      |
| `wofi`                       | App launcher       | fuzzel                      |
| `swaylock`                   | Screen lock        | hyprlock                    |
| `swayidle`                   | Idle daemon        | hypridle                    |
| `wlsunset`                   | Blue light reducer | hyprsunset                  |
| `swaybg`                     | Wallpaper          | swaybg                      |
| `grim`                       | Screenshot tool    | grim                        |
| `kanshi`                     | Display config     | kanshi                      |
| `xdg-desktop-portal`         | Portal             | xdg-desktop-portal          |
| `xdg-desktop-portal-wlroots` | Portal backend     | xdg-desktop-portal-hyprland |

### Applications

| Package                 | Description     | ArchRiot Equivalent   |
| ----------------------- | --------------- | --------------------- |
| `thunar`                | File manager    | thunar                |
| `thunar-archive-plugin` | Archive support | thunar-archive-plugin |

### System Tools

| Package | Description      | ArchRiot Equivalent |
| ------- | ---------------- | ------------------- |
| `doas`  | Sudo replacement | sudo                |
| `curl`  | HTTP client      | curl                |
| `wget`  | Download tool    | wget                |
| `unzip` | Zip extraction   | unzip               |
| `xz`    | Compression      | `xz`                |

### Wireless Firmware

| Package          | Description                   | Driver  |
| ---------------- | ----------------------------- | ------- |
| `iwx-firmware`   | Intel AX200/AX201/AX210/AX211 | `iwx`   |
| `urtwn-firmware` | Realtek RTL8188EU USB WiFi    | `urtwn` |

**Note:** OpenBSD packages are installed via `pkg_add`. Run `pkg_add -l` to list installed packages.

### Source-Built Packages

Some packages are not available in OpenBSD's package repository and must be compiled from source:

| Package    | Build Method        | URL                                       |
| ---------- | ------------------- | ----------------------------------------- |
| `wlsunset` | `git clone` + meson | https://git.sr.ht/~kennylevinsen/wlsunset |

**wlsunset build commands:**

```sh
git clone https://git.sr.ht/~kennylevinsen/wlsunset
cd wlsunset
meson setup build --prefix=/usr/local --buildtype=release
meson compile -C build
doas meson install -C build
```

---

## OpenBSD Reference

### Package Management

```sh
# Install packages
pkg_add -v package1 package2

# Update packages
pkg_add -u

# Search
pkg_info -Q searchterm

# List installed
pkg_info

# Cleanup after remove
pkg_delete -a  # removes orphaned libraries
```

**Key desktop packages:**

```
sway waybar foot fish neovim git curl
grim slurp wl-clipboard mako wofi
firefox btop transmission-gtk thunar
python3  # for swaylock-wrapper.py
swaylock swayidle
```

### Services (rc.d)

```sh
rcctl enable service
rcctl start service
rcctl disable service
```

**Key desktop services:**
| Service | Purpose |
|---------|---------|
| apmd | Power management (battery, suspend, lid) |
| sndiod | Audio (disable with `sndiod_flags=NO`) |
| sshd | SSH (enabled by default) |

### Network / WiFi

OpenBSD uses `iwx` for Intel WiFi 6 (AX211). Configure in `/etc/hostname.if`:

```sh
# /etc/hostname.iwx0
nwid YOUR_SSID wpakey YOUR_PASSWORD
inet autoconf
```

Then: `sh /etc/netstart`

**WiFi hardware status:**

- Intel AX211 (Wi-Fi 6E) → ✅ Fully supported; install `iwx-firmware`
- Intel BE201 (Wi-Fi 7) → ❌ NOT supported in OpenBSD 7.9
- Realtek RTL8188EU → ✅ Supported via `urtwn` driver; install `urtwn-firmware`
- Realtek RTL8811AU/RTL8812AU → ❌ NOT supported

### System Updates

```sh
# Patch and upgrade
syspatch -a && sysupgrade

# Package updates
pkg_add -u
```

### Doas (sudo replacement)

```sh
# /etc/doas.conf
permit nopass :wheel
```

### Disk Encryption

OpenBSD supports softraid full-disk encryption. During install, choose `(E)ncrypt` when prompted for the root disk. Passphrase prompt appears at boot.

---

## OpenBSD vs Linux: Key Differences for the Installer

| Topic           | ArchRiot (Linux)     | OpenRiot (OpenBSD)    |
| --------------- | -------------------- | --------------------- |
| Package manager | pacman + yay (AUR)   | pkg_add               |
| Init system     | systemd              | rc.d                  |
| WiFi            | networkmanager / iwd | iwx (Intel)           |
| Brightness      | brightnessctl        | N/A                   |
| Lock screen     | hyprlock             | swaylock              |
| Idle            | hypridle             | swayidle              |
| Notifications   | mako                 | mako (same)           |
| Screenshots     | grim + slurp         | grim + slurp (same)   |
| Launcher        | fuzzel               | wofi                  |
| Terminal        | ghostty              | foot                  |
| Font rendering  | fontconfig           | fontconfig (same)     |
| Idle dim        | brightnessctl        | N/A                   |
| GPU drivers     | auto-detected        | inteldrm mostly works |

---

## Known Issues

1. **`brightnessctl` not available on OpenBSD** — removed from swayidle; no standard brightness control tool exists
2. **`gsettings` not available on OpenBSD** — commented out in sway/config
3. **`i3-autotiling` not in pkg repo** — commented out; Sway's default dwindle handles most tiling
4. **`hyprlock` not available** — replaced with `swaylock`
5. **`loginctl` not available on OpenBSD** — removed from swayidle

6. **`archriot-control-panel`** — GTK4 control panel not yet ported
7. **`setup.sh`** — not yet written; critical next step
8. **Go installer** — not yet ported; setup.sh will be shell-only initially
9. **`openriot.org/setup.sh`** — not yet hosted at domain
10. **`$HOME/.local/share/openriot/`** — directory structure not yet created by installer

---

## Credits

- **ArchRiot** — The project OpenRiot is derived from. See [github.com/CyphrRiot/ArchRiot](https://github.com/CyphrRiot/ArchRiot)
- **OpenBSD** — The operating system base. See [openbsd.org](https://openbsd.org)
- **Sway** — The Wayland compositor. See [swaywm.org](https://swaywm.org)
- **Midnight Theme** — GitHub Pages Jekyll theme. See [pages-themes/midnight](https://github.com/pages-themes/midnight)

---

## License

MIT License — same as ArchRiot

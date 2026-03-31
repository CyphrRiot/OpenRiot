# OpenRiot — Project Progress

> **Status: IN PROGRESS** — Not yet installable
> **Version: 0.1**
> **Last Updated: 2026-03-31**

---

## What Is OpenRiot?

OpenRiot is a polished, opinionated **OpenBSD desktop system** built in the spirit of [ArchRiot](https://github.com/CyphrRiot/ArchRiot). The goal is a one-command post-install setup (`curl -fsSL https://openriot.org/setup.sh | sh`) that transforms a base OpenBSD installation into a fully-configured Sway + Waybar desktop — same CypherRiot aesthetics, same keybindings, same workflow — but running on OpenBSD's audited, pledge/unveil-secured base.

**Philosophy:**
- ArchRiot = rolling release, bleeding-edge Hyprland, Arch Linux base
- OpenRiot = stable OpenBSD base, Sway (i3-compatible Wayland compositor), same rice

---

## Architecture Overview

OpenRiot has three layers:

```
Layer 1: ISO Builder         → Produces bootable OpenBSD + autoinstall ISO
Layer 2: Base Install       → OpenBSD autoinstall ( unattended or guided)
Layer 3: setup.sh + Go CLI → Packages, Sway, Waybar, Fish, dotfiles
```

### Layer 1: ISO Builder (`build-iso.sh`)

**Location:** `~/Code/OpenRiot/build-iso.sh`

Downloads the OpenBSD install ISO, injects an `install.conf` (autoinstall answers) and optionally a `site79.tgz` (custom files overlay), then repacks into a bootable ISO.

**Flow:**
```
OpenBSD snapshots/amd64/install79.iso
    ↓ (download, verify SHA256)
.inject(install.conf)
    ↓ (optional)
.inject(site79.tgz)
    ↓ (xorriso repack)
isos/openriot-VERSION-OPENBSD_VERSION.iso
```

**Key files:**
- `autoinstall/install.conf` — autoinstall answer file (hostname/root password/user/password/timezone all = ask; sets location = http from cdn.openbsd.org)
- `site/` directory — custom files overlaid on target system (empty for now; populated in later steps)
- `is/os/` directory — output for built ISOs

**Status:** ✅ Working, tested, committed

---

### Layer 2: Base OpenBSD Install

The ISO boots into the standard OpenBSD installer. When `(A)utoinstall` is selected, the installer reads `install.conf` from the CD root and pre-fills all answers except disk selection (which always prompts to prevent accidental data loss).

**What the base install gets you:**
- OpenBSD -current (7.9 snapshots) base system
- All sets (base, comp, games, man, xbase, xfont, xserv, xshare)
- User account with wheel sudo/doas access
- Network configured for HTTP install from cdn.openbsd.org

**Status:** ✅ Works via autoinstall.manual disk selection.

---

### Layer 3: Post-Install Setup (`setup.sh` + Go CLI)

After base OpenBSD is installed and rebooted, run:

```sh
curl -fsSL https://openriot.org/setup.sh | sh
```

This installs and configures:

| Component | Package/Source | Notes |
|-----------|--------------|-------|
| Sway | `pkg_add sway` | Wayland compositor (i3-compatible) |
| Waybar | `pkg_add waybar` | Status bar |
| Foot | `pkg_add foot` | Terminal (replaces ghostty) |
| Fish | `pkg_add fish` | Shell |
| Neovim | `pkg_add neovim` | Editor |
| Firefox | `pkg_add firefox` | Browser |
| Mako | `pkg_add mako` | Notifications |
| Wofi | `pkg_add wofi` | App launcher |
| Grim/Slurp | `pkg_add grim slurp` | Screenshots |
| wl-clipboard | `pkg_add wl-clipboard` | Clipboard |
| doas | OpenBSD built-in | sudo replacement |
| Swaylock | `pkg_add swaylock` | Lock screen |
| swayidle | `pkg_add swayidle` | Idle management |
| wlsunset | `pkg_add wlsunset` | Blue light filter |
| bcrypt | OpenBSD base | Password hashing |

**Dotfiles** are cloned from `~/.dotfiles` (user's own repo) and symlinked.

**Status:** 🔴 Not yet written — this is the next major task (Step D in the plan below).

---

## Current Project State

### ✅ Completed

| Item | Location | Notes |
|------|----------|-------|
| ISO builder | `build-iso.sh` | Downloads, verifies, injects, repacks |
| Autoinstall config | `autoinstall/install.conf` | Asks hostname/password/user/timezone; disk prompts |
| site79.tgz support | `site/` dir + build-iso.sh step | Empty for now; ready for custom files |
| Sway config | `config/sway/` | Ported from `~/.config/sway`; OpenBSD fixes applied |
| Jekyll site | `_layouts/`, `_config.yml`, `assets/css/style.scss` | Midnight theme, CypherRiot CSS |
| README | `README.md` | Badges, IN PROGRESS warning, install methods |
| CNAME | `CNAME` | openriot.org |
| Blowfish emoji | `_layouts/default.html`, `README.md`, `_config.yml` | Replaced 🎭 with 🐡 |

### 🔴 Not Yet Done

| Item | Priority | Blocking |
|------|----------|----------|
| `setup.sh` script | Critical | Blocks end-to-end install |
| Go installer port | Critical | setup.sh depends on it |
| Package list (packages.yaml) | Critical | Need pkg_add equivalent of ArchRiot's packages.yaml |
| OpenBSD dotfiles structure | High | How to organize installed configs |
| Control panel (Go) | Medium | Nice to have; not blocking |
| `site79.tgz` populated files | Medium | doas.conf, pkg.conf, hostname |
| Backgrounds/wallpapers | Low | CypherRiot backgrounds not yet available |
| Swaylock dynamic wallpaper | Low | Requires backgrounds + openriot binary |

---

## Step-by-Step Build Plan

### Phase 1: ISO Builder (DONE ✅)

- [x] `build-iso.sh` written and tested
- [x] SHA256 verification working (correctly parses OpenBSD's `SHA256 (file) = hash` format)
- [x] `install.conf` created (autoinstall answers)
- [x] `site79.tgz` injection working (skips when dir empty)
- [x] ISO output: `isos/openriot-0.1-7.9.iso`
- [x] Version bumped to 0.1

### Phase 2: Post-Install Setup Script (NEXT 🔴)

- [ ] Write `setup.sh` (the `curl | sh` bootstrap)
  - Detect OpenBSD version
  - Install packages via `pkg_add`
  - Clone/fetch dotfiles
  - Deploy Sway + Waybar configs
  - Set Fish as default shell
  - Configure doas
  - Start Sway

**Proposed `setup.sh` structure:**
```sh
#!/bin/sh
set -e

echo "=== OpenRiot Setup ==="

# 1. Check we're on OpenBSD
if [ "$(uname)" != "OpenBSD" ]; then
    echo "ERROR: OpenRiot requires OpenBSD"
    exit 1
fi

# 2. Detect architecture and install packages
pkg_add -v sway waybar foot fish neovim \
    git curl doas \
    grim slurp wl-clipboard mako wofi \
    firefox btop transmission-gtk \
    python3

# 3. Fish as default shell for user
chsh -s /usr/local/bin/fish

# 4. Clone dotfiles
if [ -d "$HOME/.dotfiles" ]; then
    echo "Dotfiles already exist, skipping clone"
else
    git clone https://github.com/CyphrRiot/dotfiles.git "$HOME/.dotfiles"
fi

# 5. Deploy Sway config
mkdir -p "$HOME/.config/sway"
cp "$HOME/.dotfiles/sway/config" "$HOME/.config/sway/config" 2>/dev/null || true
cp "$HOME/.dotfiles/sway/keybindings.conf" "$HOME/.config/sway/keybindings.conf" 2>/dev/null || true
cp "$HOME/.dotfiles/sway/windowrules.conf" "$HOME/.config/sway/windowrules.conf" 2>/dev/null || true

# 6. Deploy Waybar config
mkdir -p "$HOME/.config/waybar"
cp -r "$HOME/.dotfiles/waybar/"*. "$HOME/.config/waybar/" 2>/dev/null || true

# 7. Deploy Fish config
cp -r "$HOME/.dotfiles/fish/"*. "$HOME/.config/fish/" 2>/dev/null || true

# 8. Doas config (passwordless for wheel)
echo "permit nopass :wheel" > /etc/doas.conf

# 9. Waybar autostart (already in Sway config)

echo "=== OpenRiot ready! Run 'sway' from tty to start ==="
```

### Phase 3: Go Installer Port (later 🔴)

Port the ArchRiot Go CLI to OpenBSD. Key differences from ArchRiot:

| ArchRiot | OpenRiot |
|----------|----------|
| `pacman` + `yay` (AUR) | `pkg_add` |
| `systemd` user services | `rc.d` / `doas` |
| Hyprland | Sway |
| `hyprlock` | `swaylock` |
| `hypridle` | `swayidle` |
| `brightnessctl` | N/A (no brightness control on OpenBSD) |
| `gsettings` | N/A (no GNOME deps) |
| `i3-autotiling` (AUR) | Not available |
| `$HOME/.local/share/archriot/` | `$HOME/.local/share/openriot/` |
| ArchRiot Hyprland modules | Sway/Waybar equivalents |

**Go packages to port/rewrite:**
- `source/main.go` — entry point, CLI flags
- `source/installer/` — pacman/yay → pkg_add
- `source/session/` — systemd → rc.d / direct commands
- `source/waybar/` — largely compatible (Waybar works on Sway)
- `source/theming/` — largely compatible
- `source/upgradeguard/` — pacman db → pkg info
- `source/upgrunner/` — pacman → pkg_add

**Go packages likely reusable without changes:**
- `source/cli/`
- `source/logger/`
- `source/config/` (loader, types, dependency validator)
- `source/executor/`
- `source/crypto/`
- `source/display/`
- `source/windows/`

---

## File Reference

### ISO Builder

| File | Purpose |
|------|---------|
| `build-iso.sh` | Main build script; downloads ISO, injects files, repacks |
| `autoinstall/install.conf` | Autoinstall answer file |
| `site/` | Directory for custom files to overlay on target system |
| `is/os/` | Output directory for built ISOs |

### Sway Config (ported from `~/.config/sway`)

| File | Purpose |
|------|---------|
| `config/sway/config` | Main Sway config; sources other files |
| `config/sway/keybindings.conf` | All keybindings (bindsym) |
| `config/sway/monitors.conf` | Monitor/workspace config |
| `config/sway/windowrules.conf` | for_window rules |
| `config/sway/swayidle.conf` | Idle/lock configuration (reference) |
| `config/sway/swaylock.conf` | Swaylock config |
| `config/sway/swaylock-wrapper.sh` | Lock screen background generator (calls Python script) |
| `config/sway/swaylock-wrapper.py` | Generates dynamic lock screen with time/date/crypto |

### Jekyll Site

| File | Purpose |
|------|---------|
| `_config.yml` | Jekyll config; Midnight theme, OpenRiot title/description |
| `_layouts/default.html` | Page layout; Midnight CSS overrides, footer, JS |
| `assets/css/style.scss` | Custom CSS (imports Midnight, then CypherRiot overrides) |
| `README.md` | GitHub Pages landing page |
| `CNAME` | Domain: openriot.org |

---

## OpenBSD-Specific Reference

### Package Management

```sh
# Install packages
pkg_add -v package1 package2

# Update packages
pkg_add -u

# Search packages
pkg_info -Q searchterm

# List installed
pkg_info
```

### Services (rc.d)

```sh
# Enable service
rcctl enable service

# Start now
rcctl start service

# Disable
rcctl disable service
```

### Key Services for Desktop

| Service | Purpose |
|---------|---------|
| `apmd` | Power management (battery, suspend, lid) |
| `sndiod` | Audio (disable with `sndiod_flags=NO`) |
| `sshd` | SSH (enabled by default on install) |

### Network/WiFi

OpenBSD uses `iwx` for Intel WiFi 6 (AX211). Configure in `/etc/hostname.if`:

```sh
# /etc/hostname.iwx0
nwid YOUR_SSID wpakey YOUR_PASSWORD
inet autoconf
```

Then: `sh /etc/netstart`

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

OpenBSD supports softraid encryption (full-disk encryption). During install, choose `(E)ncrpyt` when prompted for the root disk. This sets up a passphrase prompt at boot.

---

## OpenBSD vs Linux: Key Differences for the Installer

| Topic | ArchRiot (Linux) | OpenRiot (OpenBSD) |
|--------|------------------|-------------------|
| Package manager | pacman + yay (AUR) | pkg_add |
| Init system | systemd | rc.d (OpenBSD base) |
| Compiler toolchain | base + AUR | base + comp79.tgz |
| WiFi driver | networkmanager/iwd | iwx (Intel), urtwn, athn |
| Brightness | brightnessctl | N/A (no standard tool) |
| Lock screen | hyprlock | swaylock |
| Idle | hypridle | swayidle |
| Notifications | mako | mako |
| Screenshots | grim + slurp | grim + slurp |
| Font rendering | fontconfig | fontconfig |
| Desktop entry | .desktop files | same |
| XDG dirs | same | same |
| User systemd | systemctl --user | N/A (use raw commands) |
| Session lock | hyprlock | swaylock |
| Idle dim | brightnessctl | N/A |

---

## Known Issues

1. **`brightnessctl` not available on OpenBSD** — removed from swayidle in config/sway/config. No standard brightness control tool on OpenBSD.
2. **`gsettings` not available on OpenBSD** — commented out in sway/config. GTK theming handled differently.
3. **`i3-autotiling` not available on OpenBSD pkg repo** — commented out. Sway's default dwindle layout handles most tiling.
4. **`hyprlock` not available on OpenBSD** — replaced with `swaylock` in swayidle.
5. **Control panel** — ArchRiot's GTK4 control panel has not been ported.
6. **`setup.sh`** — not yet written; this is the critical next step.
7. **Go installer** — not yet ported; `setup.sh` will be shell-only until Go port is done.
8. **Wallpapers/backgrounds** — not yet available in OpenRiot; swaylock-wrapper.py references `~/.local/share/openriot/backgrounds/` which doesn't exist yet.
9. **`openriot.org`** site — currently serves Jekyll site only; `setup.sh` is not yet hosted at the domain.

---

## TODO List

### P0 — Critical (must have before first install)

- [ ] Write `setup.sh` bootstrap script
- [ ] Host `setup.sh` at `openriot.org/setup.sh`
- [ ] Test `setup.sh` on a real OpenBSD installation
- [ ] Fix any package name differences

### P1 — Important (full desktop experience)

- [ ] Write `install/packages.yaml` (pkg_add equivalent of ArchRiot's packages.yaml)
- [ ] Port Go installer (or significant subset) to OpenBSD
- [ ] Populate `site/` with useful files (doas.conf, pkg_add.conf, etc.)
- [ ] Copy/configure Waybar modules from ArchRiot
- [ ] Copy/configure Fish shell config from ArchRiot
- [ ] Copy/configure Neovim config from ArchRiot

### P2 — Nice to have

- [ ] Port Go control panel to OpenBSD (GTK4 or Go TUI)
- [ ] Swaylock dynamic wallpaper with OpenBSD backgrounds
- [ ] OpenRiot wallpapers (CypherRiot theme ported)
- [ ] Test on real hardware (ThinkPad X1 Carbon Gen 13)
- [ ] Test WiFi (AX211 vs BE201)

### P3 — Future

- [ ] Release ISO with pre-baked setup.sh included
- [ ] Create OpenRiot.org release page with ISO downloads
- [ ] Announce on misc@openbsd.org

---

## Credits

- **ArchRiot** — The project OpenRiot is derived from. See [github.com/CyphrRiot/ArchRiot](https://github.com/CyphrRiot/ArchRiot)
- **OpenBSD** — The operating system base. See [openbsd.org](https://www.openbsd.org)
- **Sway** — The Wayland compositor. See [swaywm.org](https://swaywm.org)
- **Midnight Theme** — GitHub Pages Jekyll theme. See [pages-themes/midnight](https://github.com/pages-themes/midnight)

---

## License

MIT License — same as ArchRiot
```

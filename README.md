<div align="center">

<img src="OpenRiot.png" alt="OpenRiot" width="200"/>

# :: 𝕆𝕡𝕖𝕟ℝ𝕚𝕠𝕥 ::

## One command. Complete OpenBSD desktop. Zero compromises.

![Version](https://img.shields.io/badge/version-0.9-blue?labelColor=0052cc)
![License](https://img.shields.io/github/license/CyphrRiot/OpenRiot?color=4338ca&labelColor=3730a3)
![Platform](https://img.shields.io/badge/platform-OpenBSD-4338ca?logo=openbsd&logoColor=white&labelColor=3730a3)
![Sway](https://img.shields.io/badge/Sway-Wayland-312e81?logo=wayland&logoColor=a855f7&labelColor=1e1b4b)

![Last Commit](https://img.shields.io/github/last-commit/CyphrRiot/OpenRiot?color=5b21b6&labelColor=4c1d95)
![Code Size](https://img.shields.io/github/languages/code-size/CyphrRiot/OpenRiot?color=4338ca&labelColor=3730a3)
![Code](https://img.shields.io/badge/human-coded-blue?logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNmZmZmZmYiIHN0cm9rZS13aWR0aD0iMiIgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIiBzdHJva2UtbGluZWpvaW49InJvdW5kIiBjbGFzcz0ibHVjaWRlIGx1Y2lkZS1wZXJzb24tc3RhbmRpbmctaWNvbiBsdWNpZGUtcGVyc29uLXN0YW5kaW5nIj48Y2lyY2xlIGN4PSIxMiIgY3k9IjUiIHI9IjEiLz48cGF0aCBkPSJtOSAyMCAzLTYgMyA2Ii8+PHBhdGggZD0ibTYgOCA2IDIgNi0yIi8+PHBhdGggZD0iTTEyIDEwdjQiLz48L3N2Zz4=&logoColor=a855f7&labelColor=1e1b4b)
![Language](https://img.shields.io/badge/language-Go-4338ca?logo=go&logoColor=c7d2fe&labelColor=3730a3)
![Language](https://img.shields.io/badge/language-YAML-5b21b6?logo=yaml&logoColor=e0e7ff&labelColor=4c1d95)

</div>

OpenRiot is the answer to every time you've thought "Why can't an OpenBSD installation just work correctly from the start and give me a fully functional desktop environment that's actually usable?" Built on the same principles as [ArchRiot](https://ArchRiot.org) and by the same creator. If you liked ArchRiot, you'll love OpenRiot.

- Read the [Post on X](https://x.com/CyphrRiot/status/2039409143891837297?s=20) to understand why OpenRiot was created and more about the creator's vision for the system.

### **Curated to be correct:**

- **🪟 Sway Tiling** — Wayland-native tiling that actually gets it right
- **⚡ Robust Binary** — Atomic operations, run-time, instant rollbacks, zero dependency hell
- **🛡️ Privacy** — Zero telemetry, tracking, zero data harvesting, zero ID requirements
- **🎨 Aesthetics** — Carefully crafted dark themes that work at any hour
- **💻 Development** — Helix, shell enhancements, and other upgrades
- **💎 OpenBSD** — The most security-audited OS on the planet

_Built on OpenBSD, because compromises are for other operating systems. This isn't maintained by committee or corporate roadmap — it's maintained by someone with an obsessive, singular focus on getting it right the first time, because crappy computing environments are an insult to what they should be._

> "Linux has never been about quality. There are so many parts of the system that are just these cheap little hacks, and it happens to run." -Theo de Raadt

## 📚 Navigate This Guide

- [🚀 Choose Your OpenRiot Experience](#choose-your-openriot-experience)
- [⌨️ Master Your OpenRiot Desktop](#master-your-openriot-desktop)
- [📝 Using Helix (Editor)](#using-helix)
- [🔄 System Management](#system-management)
- [🧰 Advanced Usage](#advanced-usage)
- [🔧 Troubleshooting](#troubleshooting)
- [📄 License](#license)
- [📋 Progress](./Progress.md)

## ✅ Supported Systems

### Highly Recommended ThinkPad Series

These ThinkPads have excellent OpenBSD support for WiFi, trackpoints, and suspend/resume:

| Model                 | CPU               | WiFi                         | Notes                                                                              |
| --------------------- | ----------------- | ---------------------------- | ---------------------------------------------------------------------------------- |
| **T14s Gen 1+** (AMD) | Ryzen 3 PRO 4450U | ⭐⭐⭐ `iwm` (AX200 adapter) | Best OpenBSD laptop experience ([buy ~$300](https://www.amazon.com/dp/B086MD6LTM)) |
| **T490**              | Intel i5-8265U    | ⭐⭐ `iwm` (Intel 9560)      | Good experience overall                                                            |
| **T480**              | Intel i5-8350U    | ⭐⭐ `iwm` (Intel 8265)      | Works well, slightly older                                                         |
| **X1 Carbon Gen 7**   | Intel i7-8665U    | ⭐⭐ `iwm` (Intel 9560)      | Premium build, good Linux/OpenBSD support                                          |
| **X270**              | Intel i5-6300U    | ⭐ `iwm` (Intel 8265)        | Small, portable, older but solid                                                   |

### Other Well-Supported Laptops

| Model                   | CPU             | WiFi                    | Notes                                        |
| ----------------------- | --------------- | ----------------------- | -------------------------------------------- |
| **Lenovo V14**          | Ryzen 5 3500U   | ⭐⭐⭐ `iwm` (AX200)    | Budget option, excellent OpenBSD support     |
| **Framework Laptop 13** | Intel i5-1240P  | ⭐⭐⭐ `iwm` (AX211)    | Modular, user-repairable, OpenBSD works well |
| **Dell XPS 13 9300**    | Intel i7-1065G7 | ⭐⭐ `iwm` (Intel 9560) | Beautiful screen, good Linux/OpenBSD support |

### Avoid or Use Caution

| Model             | Reason                                                                   |
| ----------------- | ------------------------------------------------------------------------ |
| **Any MacBook**   | Broadcom WiFi requires proprietary firmware; OpenBSD does not support it |
| **Lenovo Flex 3** | Very new hardware may not be recognized                                  |
| **HP Envy x360**  | Some models have unsupported AMD WiFi                                    |

### Key Components for OpenBSD

- **WiFi**: Use Intel `iwm` or USB Atheros adapters only. See the full supported list below.
- **CPU**: Intel and AMD Ryzen are both well-supported. ARM support is experimental.
- **GPU**: Intel integrated graphics are best-supported. AMD Radeon works but with varying feature support. NVIDIA is
  not supported on Wayland/Sway.
- **Trackpoint**: All ThinkPad trackpoints work. Some USB trackpoints may require additional configuration.

## ✅ Supported Network Hardware

#### **⚠️ OpenBSD is very selective about WiFi adapters. Only use adapters from this list:**

### Built-in WiFi (PCIe/M.2)

| Adapter                      | Chip   | OpenBSD Driver | Support Level        | Buy                                                         |
| ---------------------------- | ------ | -------------- | -------------------- | ----------------------------------------------------------- |
| **Intel Wi-Fi 6 AX200**      | `iwm`  | `iwm(4)`       | ⭐⭐⭐ Excellent     | [Check ThinkPad T14s](https://www.amazon.com/dp/B086MD6LTM) |
| **Intel Wi-Fi 6 AX201**      | `iwm`  | `iwm(4)`       | ⭐⭐⭐ Excellent     | Common in 10th-gen+ ThinkPads                               |
| **Intel Wireless 8265**      | `iwm`  | `iwm(4)`       | ⭐⭐ Good            | Found in T470, X270, others                                 |
| **Intel Wireless 8260**      | `iwm`  | `iwm(4)`       | ⭐⭐ Good            | Older but well-supported                                    |
| **Intel Wireless 3165**      | `iwm`  | `iwm(4)`       | ⭐ Good              | Older, 802.11ac only                                        |
| **Intel Wireless 7265**      | `iwm`  | `iwm(4)`       | ⭐⭐ Good            | Found                                                       |
| in T450, X250                |
| **Qualcomm Atheros QCA6174** | `athn` | `athn(4)`      | ⭐⭐ Good            | Found in some ThinkPads                                     |
| **Broadcom BCM4360**         | `brcm` | `brcm(4)`      | ⚠️ Requires firmware | Avoid if possible                                           |

### USB WiFi Adapters (Nano/Compact)

| Adapter                  | Chip      | OpenBSD Driver | Support Level    | Buy                                                 |
| ------------------------ | --------- | -------------- | ---------------- | --------------------------------------------------- |
| **ASUS USB-AC56**        | `urtwn`   | `urtwn(4)`     | ⭐⭐⭐ Excellent | [Check price](https://www.amazon.com/dp/B00PB5VR1G) |
| **TP-Link Archer T3U**   | `urtwn`   | `urtwn(4)`     | ⭐⭐ Good        | Budget option                                       |
| **Netgear A6200**        | `urtwn`   | `urtwn(4)`     | ⭐ Good          | Older but supported                                 |
| **TP-Link TL-WN722N v3** | `urtwn`   | `urtwn(4)`     | ⭐⭐ Good        | Very cheap, 802.11n only                            |
| **Alfa AWUS036NHA**      | `athn`    | `athn(4)`      | ⭐⭐⭐ Excellent | High gain, excellent range, 802.11n                 |
| **Alfa AWUS036ACS**      | `rtl88au` | `rsu(4)`       | ⭐⭐ Good        | Long range, 802.11ac                                |

### NOT Supported (Do Not Buy)

| Adapter                            | Chip        | Reason                                                  |
| ---------------------------------- | ----------- | ------------------------------------------------------- |
| **Any Broadcom** (e.g., BCM94352Z) | `brcmfmac`  | Requires proprietary firmware; OpenBSD will not load it |
| **Realtek 8812AU/8821AU**          | `rtl8812au` | No OpenBSD driver exists                                |
| **MediaTek MT7921**                | `mt7921u`   | No OpenBSD driver                                       |
| **Any 802.11ax (WiFi 6E/7) USB**   | various     | Generally not supported                                 |

## ⚠️ UEFI/BIOS Settings

Before installing OpenBSD (and therefore OpenRiot), you need to make some BIOS/UEFI adjustments to ensure everything works correctly. Most hardware ships with settings that assume you're running Windows or macOS — we need to fix that.

### How to Enter BIOS

- **ThinkPads**: Press `Enter` during boot to interrupt, then `F1` for BIOS. Or press `F12` for boot menu and look for BIOS setup.
- **Other brands**: Press `F2`, `F10`, or `Del` during boot.

### Recommended UEFI/BIOS Settings

1. **Disable Secure Boot** — OpenBSD does not support Secure Boot. You must disable it in BIOS.
    - Navigate to `Security` → `Secure Boot` → Set to **Disabled**
    - If there's a "Microsoft Windows" Secure Boot key, you may need to clear it first

2. **Set Boot Mode to "UEFI Only" (or "UEFI and Legacy" if available)**
    - Navigate to `Boot` → `Boot Mode` → Select **UEFI Only** (or **UEFI + Legacy**)
    - Avoid "Legacy Only" as OpenBSD prefers UEFI

3. **Disable Fast Boot / Fast Startup** (if available)
    - This can prevent the boot menu from appearing

    - Navigate to `Power` → `Fast Startup` → **Disabled**

4. **Enable "USB Boot"** (if available)
    - Ensures you can boot from USB drives

5. **Set boot order to prioritize your USB/ISO device**
    - Navigate to `Boot` → `Boot Order` → Place your USB drive first

6. **Disable Intel VTD** (if you encounter Sway/wlroots issues)
    - Navigate to `Security` → `Intel VT-d` or `AMD-Vi` → **Disabled**
    - Note: This is only needed in rare cases. Try with it enabled first.

7. **Set SATA mode to AHCI** (not RAID/Intel RST)
    - Navigate to `Storage` → `SATA Mode` → **AHCI**
    - RAID mode can cause OpenBSD to not see the disk

### Pre-Installation Checklist

Before booting the OpenRiot ISO:

- [ ] USB drive created with OpenRiot ISO (see above)
- [ ] Secure Boot disabled in BIOS
- [ ] Boot mode set to UEFI
- [ ] USB boot enabled
- [ ] SATA mode set to AHCI
- [ ] BIOS defaults loaded if you made many changes
- [ ] CMOS battery healthy (or laptop plugged in) to preserve settings

### Why This Matters for OpenBSD

OpenBSD is more conservative than Linux about hardware defaults. It assumes a clean, standards-compliant UEFI environment. Secure Boot, fast boot, and RAID modes are all Microsoft/Intel/AMD-specific optimizations that OpenBSD doesn't use — they can cause boot failures, disk recognition issues, or prevent Sway from starting.

## 🔊 Bluetooth

#### **⚠️ OpenBSD has NO native Bluetooth support.** The Bluetooth stack was removed years ago and has not been reinstated.

This means:

- **No Bluetooth audio** (no AirPods, no Bluetooth headphones, no Bluetooth speakers)
- **No Bluetooth mice or keyboards** (pairing will fail)
- **No file transfer** (no OBEX)

### What Doesn't Work

- AirPods, Beats, or any Bluetooth audio device
- Bluetooth mice or keyboards (Logitech MX Master, Apple Magic Mouse, etc.)
- Any device that requires Bluetooth pairing

### Bluetooth Audio Workaround

The best workaround is to use USB audio or a USB Bluetooth adapter that presents itself as a wired audio device. Options:

1. **USB Speaker** — Just plug and play. No Bluetooth needed.
2. **USB DAC + Wired Headphones** — Better audio quality anyway.
3. **AirPods via USB-C cable** — Use them as wired earbuds (yes, really)
4. **USB Bluetooth adapter that works as audio** — Some adapters present A2DP profile as USB audio (very rare)

### Bluetooth Mouse/Keyboard Workaround

1. **Use a USB mouse** — Any basic USB mouse works perfectly
2. **Use a 2.4GHz wireless mouse** — Logitech Unifying Receiver (uses a separate USB dongle, not Bluetooth)
3. **Use a wired mouse or keyboard** — Works 100% of the time

### Recommended Input Setup

For the best OpenBSD + Sway experience:

| Device    | Recommendation                                           |
| --------- | -------------------------------------------------------- |
| **Mouse** | Basic USB mouse (2.4GHz wireless with dongle also works) |

> "You are absolutely deluded, if not stupid, if you think that a worldwide collection of software engineers who can't write operating systems or applications without security holes, can then turn around and suddenly write virtualization layers without security holes." — Theo de Raadt
> | **Keyboard** | Any USB keyboard; ThinkPad keyboards work perfectly |
> | **TrackPoint** | Works natively on ThinkPads — no configuration needed |
> | **Graphics** | Intel iGPU preferred; AMD Radeon works; NVIDIA not supported |

> "My favorite part of the 'many eyes' argument is how few bugs were found by the two eyes of Eric (the originator of the statement). All the many eyes are apparently attached to a lot of hands that type lots of words about many eyes, and never actually audit code." — Theo de Raadt

## 🚀 Choose Your OpenRiot Experience

### 🔥 Method 1: Install Script

**You already have OpenBSD installed**

If you already have a working OpenBSD system and just want the OpenRiot desktop experience:

```bash
curl -fsSL https://openriot.org/setup.sh | sh
```

This will:

- Install all required packages
- Deploy Sway, Waybar, Fish, Helix, foot, fuzzel configs
- Set up themes, fonts, keybindings
- Configure your desktop automatically

**Perfect for:**

- 🏠 Existing OpenBSD installations (7.8+)
- 🎨 Upgrading from a manual Sway setup
- ⚡ Quick setup on a fresh OpenBSD install

### ⚡ Method 2: OpenRiot ISO

**You do NOT have OpenBSD installed**

Download the OpenRiot ISO and boot it:

1. **Download OpenRiot ISO** — Get it from [openriot.org](https://openriot.org) (or the GitHub releases page)
2. **Create bootable USB** — Use `dd` or [Etcher](https://etcher.balena.io/) to write to USB
3. **Boot from USB** — Disable Secure Boot, set UEFI boot order
4. **Run autoinstall** — The ISO will automatically partition, install, and configure everything

**Perfect for:**

- 🖥️ Fresh hardware / new builds
- 🚀 Instant desktop in minutes
- 💀 Complete system replacement
- 🎯 Zero configuration required

### 🔧 Method 3: Interactive Install

**You want control over partitioning**

If you want to manually control your disk layout:

1. Download OpenRiot ISO
2. Boot from USB
3. At the boot prompt, type `I` for interactive install (instead of `A` for autoinstall)
4. Follow the guided prompts — most answers are pre-filled
5. When asked about sets, just press Enter to accept defaults (site79.tgz is pre-selected)

#### Boot and Install

1. Boot from USB (disable Secure Boot first!)
2. At the `boot>` prompt, type `I` and press Enter
3. The installer will start in interactive mode

#### Interactive Prompts

Most prompts are pre-answered. You only need to:

| Prompt              | Action                                                                             |
| ------------------- | ---------------------------------------------------------------------------------- |
| Which disk          | Press `Enter` for default (usually `sd0`)                                          |
| Use (W)hole disk    | Press `Enter` for default (GPT)                                                    |
| Root disk           | Press `Enter` for default                                                          |
| Partition layout    | Press `Enter` to accept `c` (custom layout shown)                                  |
| System hostname     | Type your hostname or press `Enter` for default                                    |
| Password for root   | Type and confirm                                                                   |
| Setup a user        | Type username and password                                                         |
| Timezone            | Press `Enter` for `US/Pacific` or type yours                                       |
| Do you expect X     | Type `no` (we use Wayland, not X)                                                  |
| Location of sets    | Press `Enter` for `cd0`                                                            |
| Set name            | **IMPORTANT:** Type `site79.tgz` to select the OpenRiot packages, then type `done` |
| SHA256 verification | Type `yes` to continue without verification                                        |

#### Partition Layout (choose `c`)

When asked for partition layout, choose `c` for custom:

```
Partition layout: c
```

This gives you:

```
/           2G
swap        1G
/home       *   (rest of disk)
```

This is correct for most users. Adjust only if you know what you're doing.

### Quick-Start Install Reference

| Prompt             | Answer                   |
| ------------------ | ------------------------ |
| Network interfaces | `done` (offline)         |
| X Window System    | **`no`**                 |
| Sets location      | `cd0`                    |
| Set name(s)        | `*` (all sets + site79)  |
| SSH                | `none` (offline install) |

#### After Install

After the base OpenBSD install completes, the system will:

1. Extract `site79.tgz` (contains all OpenRiot packages and configs)
2. Run `install.site` to configure everything
3. Reboot

After reboot:

1. Log in as your user
2. Type `fish` if bash is still default
3. Run `openriot --install` if configs don't deploy automatically
4. Type `sway` to start the desktop

## ⌨️ Master Your OpenRiot Desktop

_This section is being actively documented. For now, the essential bindings are documented in [📝 Using Helix](#using-helix). A full Sway keybinding reference is coming._

### Essential Keybindings

| Key                   | Action                     |
| --------------------- | -------------------------- |
| `Super + Return`      | Open terminal (foot)       |
| `Super + D`           | Open app launcher (fuzzel) |
| `Super + Q`           | Close window               |
| `Super + E`           | Open file manager          |
| `Super + L`           | Lock screen                |
| `Super + V`           | Toggle floating            |
| `Super + J`           | Toggle split               |
| `Super + 1-4`         | Switch workspace           |
| `Super + Shift + 1-4` | Move window to workspace   |
| `Super + Shift + Q`   | Force close window         |
| `Super + Shift + R`   | Reload Sway config         |
| `Super + Escape`      | Open power menu            |
| `Super + F`           | File Manager (Thunar)      |
| `Super + B`           | Browser                    |
| `Super + O`           | Open Helix editor          |
| `Print`               | Screenshot (region)        |
| `Super + Shift + H`   | Keybindings Help           |
| `Super + Shift + S`   | Screenshot (region)        |
| `Super + Shift + W`   | Screenshot (window)        |
| `Super + Shift + F`   | Screenshot (full)          |

### Waybar Modules

Waybar is your status bar. Click on modules for more:

| Module      | Click Action     |
| ----------- | ---------------- |
| Workspace   | Click to switch  |
| CPU         | Shows usage      |
| Memory      | Shows usage      |
| Temperature | Shows temp       |
| Battery     | Shows percentage |
| Network     | Click for nmtui  |
| Volume      | Click for mixer  |
| Clock       | Shows date/time  |

<a id="using-helix"></a>

## 📝 Using Helix — The Default Editor

OpenRiot ships with **Helix** as the default terminal editor instead of Neovim.

Helix is a modern, fast, and highly polished modal text editor written in Rust. It was chosen for OpenRiot because it perfectly aligns with the project's core philosophy: **simplicity, correctness, excellent defaults, and minimal maintenance overhead**.

### Why Helix Was Chosen Over Neovim

- **Sane defaults out of the box** — Built-in LSP support, Tree-sitter syntax highlighting, multi-cursor editing, fuzzy finding, and diagnostics work immediately with zero configuration.
- **Minimal configuration** — A single, readable `config.toml` file (usually under 100 lines) replaces hundreds of lines of Lua plugins and init scripts.
- **Performance** — Extremely fast startup time and low memory usage, which feels especially good on OpenBSD.
- **Simpler maintenance** — Much easier to include and keep consistent across OpenRiot installs and future OpenBSD releases.
- **Modern editing model** — Selection-first workflow (select then act) is consistent and reduces cognitive load once learned.
- **Better security & auditability** — Written in Rust with memory safety, aligning with OpenBSD's values.

Helix gives you a powerful, modern editing experience while staying lightweight and "correct" — exactly what OpenRiot aims for.

### Getting Started with Helix

Launch Helix with:

- `Super + O` — Open Helix (default keybinding in OpenRiot)
- Or simply run `hx` in any terminal

Helix starts in **Normal mode** by default. Here are the most important commands to get you productive quickly:

#### Basic Movement & Modes

| Key         | Action                                |
| ----------- | ------------------------------------- |
| `i`         | Enter **Insert mode** (type normally) |
| `Escape`    | Return to **Normal mode**             |
| `h j k l`   | Move left / down / up / right         |
| `w / b / e` | Jump word forward / backward / to end |
| `gg / G`    | Go to top / bottom of file            |
| `0 / $`     | Go to start / end of line             |

#### Editing

| Key     | Action                              |
| ------- | ----------------------------------- |
| `x`     | Select current line                 |
| `y`     | Yank (copy) selection               |
| `p / P` | Paste after / before cursor         |
| `d`     | Delete selection                    |
| `c`     | Change (delete + enter Insert mode) |
| `> / <` | Indent / unindent selection         |
| `u / U` | Undo / Redo                         |

#### Advanced & Useful

| Key            | Action                                     |
| -------------- | ------------------------------------------ |
| `Space + f`    | Open file picker (fuzzy finder)            |
| `Space + b`    | Switch between open buffers                |
| `Space + s`    | Symbol picker (functions, variables, etc.) |
| `/`            | Search forward                             |
| `:`            | Command mode (`:w`, `:q`, `:wq`, etc.)     |
| `gd`           | Go to definition (via LSP)                 |
| `Ctrl+w v / s` | Split window vertically / horizontally     |

### Helix on OpenBSD & OpenRiot

Helix works **beautifully** on OpenBSD:

- Excellent performance on ThinkPads and Framework laptops
- Native OpenBSD packaging (`pkg_add helix`)
- Full Tree-sitter and LSP support for Go, Rust, Python, Lua, YAML, TOML, and many other languages
- No plugin manager headaches — everything just works
- Plays perfectly with Sway, foot terminal, and fish shell

**Pro tip:** Helix has one of the best default dark themes available. It looks right at home with OpenRiot's dark aesthetic.

For the complete keymap and configuration options, visit the official documentation:  
[https://docs.helix-editor.com/](https://docs.helix-editor.com/)

**Tutorial Video:** [Helix Editor Crash Course](https://www.youtube.com/watch?v=HcuDmSb-JBU)

<a id="system-management"></a>

## 🔄 System Management

OpenRiot uses `pkg_add` for package management. Packages are pre-configured in `/etc/installurl` to use OpenBSD's official CDN.

### Finding Packages

```bash
# Search for a package
pkg_info -Q <package-name>

# List all installed packages
pkg_info -m

# Check for updates (OpenBSD doesn't have a rolling update model)
# Fresh install is always the current release
```

### Updating the System

OpenBSD doesn't use `apt update` or `pacman -Syu`. To update:

1. Download and boot the **new** OpenBSD ISO
2. Run `Upgrade` instead of `Install`
3. Your `/home` partition is preserved
4. All packages are refreshed from the new ISO

For packages between releases:

```bash
# Install a new package
pkg_add <package-name>

# Remove a package
pkg_delete <package-name>
```

## 🧰 Advanced Usage

### Mullvad VPN on OpenBSD

OpenRiot supports Mullvad VPN with WireGuard. Here's how to set it up:

#### 1. Install WireGuard Tools

```bash
pkg_add wireguard-tools
```

#### 2. Generate Mullvad Config

1. Log into your [Mullvad account](https://mullvad.net/)
2. Go to **Account** → **WireGuard keys**
3. Generate a new WireGuard key
4. Download the WireGuard config file

#### 3. Place the Config

```bash
# Save the Mullvad config
doas mv ~/Downloads/mullvad.conf /etc/wireguard/wg0.conf
```

#### 4. Connect

```bash
doas rcctl enable wg-quickwg0
doas rcctl start wg-quickwg0
```

#### 5. Verify

```bash
# Check if tunnel is up
ifconfig wg0

# Verify traffic goes through VPN
curl https://am.i.mullvad.net/json
```

The output should show `"mullvad_exit_ip": true`

#### Disconnect

```bash
doas rcctl stop wg-quickwg0
```

#### Auto-start at Boot (Optional)

```bash
doas rcctl enable wg-quickwg0
```

#### DNS Leaks

Mullvad config includes their DNS servers by default. To verify no DNS leaks:

```bash
# Check DNS
cat /etc/resolv.conf

# Should show Mullvad DNS (10.64.0.1 or similar)
```

### Environment Variables

OpenRiot sets sensible defaults. Key environment variables:

```bash
# Wayland display (usually set automatically)
echo $WAYLAND_DISPLAY

# XDG directories (usually correct by default)
echo $XDG_CONFIG_HOME
echo $XDG_DATA_HOME

# Fish is the default shell
echo $SHELL  # Should show /usr/local/bin/fish
```

### Keybindings Customization

Keybindings are in `~/.config/sway/keybindings.conf`.

Edit this file to customize. After saving, press `Super + Shift + R` to reload Sway.

### Waybar Modules

Waybar modules are in `~/.config/waybar/config`.

Each module has its own config section. Common modules:

| Module      | Config Section        |
| ----------- | --------------------- |
| Workspaces  | `hyprland/workspaces` |
| CPU         | `cpu`                 |
| Memory      | `memory`              |
| Temperature | `temperature`         |
| Battery     | `battery`             |
| Network     | `network`             |
| Volume      | `volume`              |
| Clock       | `clock`               |

## 🔧 Troubleshooting

### WiFi not working

1. **Check if WiFi is recognized:**

    ```bash
    ifconfig | grep -E "^iwm[0-9]"
    ```

2. **If no WiFi device shows:**
    - Your adapter may not be supported (see hardware list above)
    - Try a USB WiFi adapter from the supported list
    - Check `dmesg` for hardware errors

3. **Connect to WiFi:**

    ```bash
    # Use nmtui for a curses-based network manager
    nmtui

    # Or use iwnctl directly
    doas iwnctl
    ```

4. **After connecting:**
    ```bash
    # Verify connection
    ifconfig iwm0
    ping -c 3 openbsd.org
    ```

### Sway won't start

1. **Check for errors:**

    ```bash
    sway 2>&1 | head -50
    ```

2. **Common fixes:**
    - Missing seatd: `doas rcctl enable seatd && doas rcctl start seatd`
    - Graphics driver issue: Try `WLR_BACKENDS=headless sway` to test
    - XWayland missing: `pkg_add xwayland`

3. **Check dmesg for hardware issues:**
    ```bash
    dmesg | grep -E "error|failed|intel|amd|nvidia"
    ```

### Package missing

If `pkg_add` fails:

1. **Verify installurl is set:**

    ```bash
    cat /etc/installurl
    # Should show: https://cdn.openbsd.org/pub/OpenBSD
    ```

2. **Set it if missing:**

    ```bash
    echo "https://cdn.openbsd.org/pub/OpenBSD" | doas tee /etc/installurl
    ```

3. **Try again:**
    ```bash
    pkg_add -v <package-name>
    ```

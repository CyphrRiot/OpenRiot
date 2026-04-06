<div align="center">

<img src="OpenRiot.png" alt="OpenRiot" width="200"/>

# :: 𝕆𝕡𝕖𝕟ℝ𝕚𝕠𝕥 ::

## One command. Complete OpenBSD desktop. Zero compromises.

![Version](https://img.shields.io/badge/version-1.0-blue?labelColor=0052cc)
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

### **Curated to be correct**

- **🪟 Sway Tiling** — Wayland-native tiling that actually gets it right
- **⚡ Robust Binary** — Atomic operations, run-time, instant rollbacks, zero dependency hell
- **🛡️ Privacy** — Zero telemetry, tracking, zero data harvesting, zero ID requirements
- **🎨 Aesthetics** — Carefully crafted dark themes that work at any hour
- **💻 Development** — Helix, shell enhancements, and other upgrades
- **💎 OpenBSD** — The most security-audited OS on the planet

#### Built on OpenBSD.

**Because compromises belong on other operating systems.**

This isn’t shaped by committees, corporate roadmaps, or quarterly deliverables. It’s built and maintained by one person with an obsessive focus on doing it right the first time — because a mediocre computing environment isn’t just inconvenient. It’s an insult to what computers should be.

> "Linux has never been about quality. There are so many parts of the system that are just these cheap little hacks, and it happens to run." -Theo de Raadt

---

![OpenRiot Desktop](assets/screenshot.png)

## ⚠️ **NOT READY FOR PRODUCTION USE** ⚠️

OpenRiot is under active development.

The ISO install is functional but has known limitations:

- You have to answer "yes" to fix many configuration issues
- Package installation on first boot will require manual intervention
- Some features are still being developed and tested
- **DO NOT use on production systems**

**Current status:** ISO installs and boots, some packages install via install.site, but many are missing (like `curl`) and there are serious bugs in the installer.

---

## 📚 Navigate This Guide

- [🚀 Installing OpenRiot]
- [⌨️ Master Your OpenRiot Desktop](#master-your-openriot-desktop)
- [📝 Using Helix (Editor)](#using-helix)
- [🔄 System Management](#system-management)
- [🧰 Advanced Usage](#advanced-usage)
    - [🔄 Environment Variables](#environment-variables)
    - [⌨️ Keybindings Customization](#keybindings-customization)
    - [📊 Waybar Modules](#waybar-modules)
    - [🔐 Crypto Config](#crypto-config)
    - [🔒 Mullvad VPN](#mullvad-vpn-on-openbsd)
- [🔧 Troubleshooting](#troubleshooting)
- [🦊 Browser & Data Transfer](#browser--data-transfer)
- [📄 License](#license)
- [📋 Progress](./Progress.md)

## ✅ Supported Systems

### Highly Recommended ThinkPad Series

These ThinkPads have excellent OpenBSD support for WiFi, trackpoints, and suspend/resume:

| Model                 | CPU               | WiFi                         | Notes                                     |
| --------------------- | ----------------- | ---------------------------- | ----------------------------------------- |
| **T14s Gen 1+** (AMD) | Ryzen 3 PRO 4450U | ⭐⭐⭐ `iwm` (AX200 adapter) | Best OpenBSD laptop                       |
| **T490**              | Intel i5-8265U    | ⭐⭐ `iwm` (Intel 9560)      | Good experience overall                   |
| **T480**              | Intel i5-8350U    | ⭐⭐ `iwm` (Intel 8265)      | Works well, slightly older                |
| **X1 Carbon Gen 7**   | Intel i7-8665U    | ⭐⭐ `iwm` (Intel 9560)      | Premium build, good Linux/OpenBSD support |
| **X270**              | Intel i5-6300U    | ⭐ `iwm` (Intel 8265)        | Small, portable, older but solid          |

You can buy a T14s Gen 1 for ~$300 USD at experience [Amazon](https://www.amazon.com/dp/B086MD6LTM).

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

- USB drive created with OpenRiot ISO (see above)
- Secure Boot disabled in BIOS
- Boot mode set to UEFI
- USB boot enabled
- SATA mode set to AHCI
- BIOS defaults loaded if you made many changes
- CMOS battery healthy (or laptop plugged in) to preserve settings

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

| Device         | Recommendation                                               |
| -------------- | ------------------------------------------------------------ |
| **Mouse**      | Basic USB mouse (2.4GHz wireless with dongle also works)     |
| **Keyboard**   | Any USB keyboard; ThinkPad keyboards work perfectly          |
| **TrackPoint** | Works natively on ThinkPads — no configuration needed        |
| **Graphics**   | Intel iGPU preferred; AMD Radeon works; NVIDIA not supported |

<a id="choose-your-openriot-experience"></a>

## 🚀 Installing OpenRiot

The OpenRiot ISO is the OpenBSD installer — it installs the base system AND configures Sway, Waybar, Fish, and everything else automatically. No interaction needed.

1. **Download OpenRiot ISO** — Get it from the [Release Page](https://github.com/CyphrRiot/OpenRiot/releases/tag/v1.0) or download directly: [openriot.iso](https://github.com/CyphrRiot/OpenRiot/releases/download/v1.0/openriot.iso) (~757MB)
2. **Create bootable USB** — Use `dd` or [Etcher](https://etcher.balena.io/) to write to USB
3. **Boot from USB** — Disable Secure Boot, set UEFI boot order
4. **Walk away** — The installer runs completely unattended

After the install finishes and the system reboots:

```bash
doas pkg_add curl git
curl -fsSL https://openriot.org/setup.sh | sh
# Reboot — Sway starts automatically
```

**Perfect for:**

- 🖥️ Fresh hardware / new builds
- 🚀 Instant desktop in minutes
- 💀 Complete system replacement
- 🎯 Zero configuration required

#### Boot and Install

1. Boot from USB (disable Secure Boot first!)
2. After the `boot>` prompt, type `I` and press Enter
3. The installer will start in interactive mode

#### Interactive Prompts

Most prompts are pre-answered. You only need to:

| Prompt               | Action                                                  |
| -------------------- | ------------------------------------------------------- |
| Keyboard layout      | Press `Enter` (use default)                             |
| System hostname      | Type `openriot` (or your preferred hostname) → Enter    |
| Network interface    | Type `done` → Enter                                     |
| IPv4 autoconf        | Press `Enter` (accept default)                          |
| IPv6                 | Type `none` → Enter                                     |
| Root password        | Type and confirm strong password                        |
| Start sshd           | Press `Enter` (yes is fine)                             |
| X Window System      | Type `no` → Enter                                       |
| Setup a user         | Type your username → Enter, then set password           |
| Which disk           | Type `sd1` (USB boot: sd0=USB, sd1=target)           |
| Use (W)hole disk MBR | Press `Enter` (GPT)                                     |
| Encrypt disk         | Type `p` for passphrase or `no` for no encryption       |
| Partition layout     | Type `c` for custom                                     |
| Label editor         | `z` → `a /` → size → `a swap` → `a /home` → `w` → `q`   |
| Location of sets     | Type `disk` (USB is auto-mounted)                       |
| Set name(s)          | Press `Enter` to select all sets including `site79.tgz` |
| SHA256 verification  | Type `yes` → Enter                                      |

#### Partition Layout (choose `c`)

When asked for partition layout, choose `c` for custom:

```
Partition layout: c
```

This gives you:

```
/           50G (or more, as needed)
swap        2G  (or more, as needed)
/home       *   (rest of disk)
```

This is correct for most users. Adjust only if you know what you're doing.

### Quick-Start Install Reference

| Prompt             | Answer                   |
| ------------------ | ------------------------ |
| Network interfaces | `done` (offline)         |
| X Window System    | `**no**`                |
| Sets location      | `disk`           |
| Set name(s)        | `*` (all sets + site79)  |
| SSH                | `none` (offline install) |

### Log Locations

If something goes wrong, check these logs:

| Stage                | Log File                        | Description       |
| -------------------- | ------------------------------- | ----------------- |
| `setup.sh`           | `~/.cache/openriot/setup.log`   | All setup output  |
| `openriot --install` | `~/.cache/openriot/install.log` | Config deployment |

To view logs:

```bash
cat ~/.cache/openriot/setup.log
cat ~/.cache/openriot/install.log
```

#### After Install

After the base OpenBSD install completes, the system will:

1. Extract `site79.tgz` (contains all OpenRiot packages and configs)
2. Run `install.site` to configure everything
3. Reboot

After reboot:

1. Log in as your user
2. Type `fish` if ksh is still default
3. Run `openriot --install` if configs don't deploy automatically
4. Type `sway` to start the desktop

<a id="master-your-openriot-desktop"></a>

## ⌨️ Master Your OpenRiot Desktop

_This section is being actively documented. For now, the essential bindings are documented in [📝 Using Helix](#using-helix). A full Sway keybinding reference is coming._

### Essential Keybindings

| Key                   | Action                     |
| --------------------- | -------------------------- |
| `Super + Return`      | Open terminal              |
| `Super + D`           | Open app launcher (fuzzel) |
| `Super + Q`           | Close window               |
| `Super + E`           | Proton Mail (web app)      |
| `Super + L`           | Lock screen                |
| `Super + V`           | Toggle floating            |
| `Super + J`           | Toggle split               |
| `Super + 1-4`         | Switch workspace           |
| `Super + Shift + 1-4` | Move window to workspace   |
| `Super + Shift + Q`   | Force close window         |
| `Super + Escape`      | Open power menu            |
| `Super + F`           | File Manager (lf)          |
| `Super + B`           | Browser                     |
| `Super + P`           | Toggle pseudo tiling        |
| `Super + O`           | Open Helix (Documents)      |
| `Super + N`           | Open NeoVim                 |
| `Super + C`           | Open Crush AI               |
| `Super + T`           | Open system monitor (btop) |
| `Super + G`           | Telegram                    |
| `Super + M`           | Google Messages             |
| `Super + X`           | X (Twitter)                 |
| `Super + Shift + Return` | Floating terminal        |
| `Print`               | Screenshot (region)         |
| `Mod1 + Tab`          | Cycle windows               |
| `Super + Shift + H`   | OpenRiot Help (website)     |
| `Super + F`           | File Manager (lf)           |

### Waybar Modules

Waybar is your status bar. Click on modules for more:

| Module      | Click Action                     |
| ----------- | -------------------------------- |
| Workspace   | Click to switch                  |
| CPU         | Shows usage                      |
| Memory      | Shows usage                      |
| Temperature | Shows temp                       |
| Battery     | Shows percentage                 |
| Network     | Click for nmtui (NetworkManager) |
| Volume      | Click for mixer                  |
| Clock       | Shows date/time                  |

## Shell Aliases & Quick Reference

Fish comes pre-configured with useful aliases:

| Alias | Command   | Description             |
| ----- | --------- | ----------------------- |
| `ls`  | `lsd`      | Default listing with icons  |
| `ll`  | `lsd -l`   | Long listing with icons     |
| `la`  | `lsd -la`  | Show hidden files           |

### lf File Manager Shortcuts

| Key       | Action                           |
| --------- | -------------------------------- |
| `j/k`     | Navigate down/up                 |
| `h/l`     | Go back / Open file or directory |
| `gh`      | Go to home `~`                   |
| `g/`      | Go to root `/`                   |
| `gg`      | Go to top of listing             |
| `G`       | Go to bottom of listing          |
| `a.`      | Toggle hidden files              |
| `o`       | Open file with default handler   |
| `<enter>` | Open file (same as `o`)          |
| `E`       | Edit file in `$EDITOR` (Helix)   |
| `yy`      | Copy (yank) file path            |
| `yd`      | Copy directory path to clipboard |
| `y.`      | Copy filename to clipboard       |
| `dd`      | Cut / trash file                 |
| `p`       | Paste                            |
| `nf`      | New directory                    |
| `n.`      | New file                         |
| `r`       | Rename (inline edit)             |
| `bc`      | Bulk rename selected files       |
| `za`      | Create archive from selection    |
| `zx`      | Extract archive                  |
| `gs`      | Show git status                  |
| `ai`      | Sort by size                     |
| `at`      | Sort by size + time              |
| `an`      | Sort by name                     |
| `q`       | Quit                             |
| `Q`       | Quit all lf instances            |

**Note:** Press `?` in lf for full help. File previews shown inline (images via chafa, text via bat).

### Tutorial Video

**Tutorial Video:** [How to Set Up and Configure LF (The Best Terminal File Manager)](https://www.youtube.com/watch?v=2oWqD3JCXuI) by Eric Murphy (~16 min)

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

### Vim to Helix Quick Reference

If you know Vim/Neovim, here's how the same tasks work in Helix:

| Task                       | Vim/Neovim                 | Helix Equivalent            | Notes / Nuances in Helix                                                                                                                                   |
| -------------------------- | -------------------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Go to top of document      | `gg`                       | `gg`                        | Same as Vim. Also works with a count (e.g., `5gg` for line 5).                                                                                             |
| Go to bottom of document   | `G`                        | `ge`                        | Different from Vim. `G` alone does nothing useful by default.                                                                                              |
| Delete character           | `x`                        | `x`                         | **OpenRiot remaps `x` to `delete_char_forward`** — now works exactly like Vim! `X` deletes backward.                                                       |
| Delete line                | `dd`                       | `dl` or `x` then `d`        | Use `dl` (delete line under cursor). Or `x` to select, `d` to delete.                                                                                      |
| Go to end of line          | `$`                        | `gl`                        | `gl` = goto line end. Very common.                                                                                                                         |
| Go to start of line        | `0` or `^`                 | `gh`                        | `gh` = goto home (start of line). Use `gs` if you want the first non-whitespace character (like Vim's `^`).                                                |
| Copy line (yank line)      | `yy`                       | `yl`                        | `yl` yanks the current line.                                                                                                                               |
| Paste line                 | `p` (below) or `P` (above) | `p` (after) or `P` (before) | Works similarly, but Helix pastes after/before the current selection (or cursor position). For a full line paste, the behavior is usually what you expect. |
| Copy text (yank selection) | `y` (after selecting)      | `y`                         | Same letter, but you select first (e.g., `w` for word, `gl` for to end of line, or visual movements).                                                      |
| Paste text                 | `p` or `P`                 | `p` or `P`                  | Same as above. Helix also supports system clipboard via `<space>p` / `<space>y` (or configure defaults).                                                   |

> **Note:** OpenRiot's Helix config remaps `x` to `delete_char_forward` and `X` to `delete_char_backward` — so `x` now works like Vim instead of Helix's default (select entire line).

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

_See the [helix-cheat-sheet](https://github.com/stevenhoy/helix-cheat-sheet) project for a visual keybinding reference._

**Tutorial Video:** [Helix Editor Crash Course](https://www.youtube.com/watch?v=HcuDmSb-JBU)

### AI Integration with OpenRouter

OpenRiot bundles **Crush** for AI-assisted coding. Crush is a modern, lightweight, Go-based terminal AI coding agent with excellent OpenBSD support. It is built automatically during setup and installed to `~/.local/bin/crush`.

![Crush AI in action](assets/crush.png)

#### Configure Crush

Create `~/.config/crush/config.yaml`:

```yaml
provider: openrouter
model: minimax/minimax-m2.7
api_key: sk-or-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

Replace `sk-or-XXXXXXXX...` with your actual OpenRouter API key from https://openrouter.ai/settings

#### How to Use

Run Crush in a terminal:

```fish
crush
```

For a Zed-like experience, run Helix and Crush side-by-side in Zellij:

1. Start Zellij with a vertical split
2. Left pane: `hx`
3. Right pane: `crush`

Select code in Helix (`y` to yank), paste into Crush, and ask questions.

<a id="browser--data-transfer"></a>

## 🦊 Browser & Data Transfer

#### **OpenBSD has no Brave browser.** Chromium-based browsers are limited — only Ungoogled Chromium is available, and Firefox is the recommended default.

This means:

- **No Brave, no Chrome, no Edge** — these Chromium derivatives are not ported
- **Firefox is the recommended browser** — available as `firefox` package
- **Ungoogled Chromium** — available as `ungoogled-chromium` for those who prefer Chromium

### Why Firefox?

| Browser                | OpenBSD Support  | Notes                         |
| ---------------------- | ---------------- | ----------------------------- |
| **Firefox**            | ✅ Full          | `pkg_add firefox`             |
| **Ungoogled Chromium** | ✅ Available     | `pkg_add ungoogled-chromium`  |
| **Brave/Chrome/Edge**  | ❌ Not available | Chromium derivatives, no port |

Firefox is open source, actively maintained, privacy-respecting by default, and has excellent OpenBSD support.

---

### Transferring Your Data from Brave

If you're moving from Arch/Brave to OpenBSD/Firefox, here's how to migrate your data.

#### Bookmarks (Easy ✅)

Brave and Firefox both support standard HTML bookmark export/import:

```bash
# 1. In Brave
Navigate to brave://bookmarks/ → click ⋮ → Export Bookmarks

# 2. In Firefox
Bookmarks → Show All Bookmarks → Import and Backup → Import → Choose HTML file
```

#### Extensions

Unfortunately, extensions must be **manually reinstalled** in Firefox. There is no bulk export when moving to a different system.

```bash
# Visit Firefox Add-ons and reinstall each one:
about:addons
```

#### Passwords (Moderate 🔧)

**Option 1: CSV Export (Quick)**

```bash
# In Brave
brave://settings/passwords → Export Passwords → CSV

# In Firefox
about:logins → Import → CSV
```

> ⚠️ CSV is unencrypted — only do this on a trusted machine.

**Option 2: Just re-login** — skip the export entirely for security.

#### History (Difficult ⚠️)

Firefox and Chromium use incompatible SQLite schemas. Full history transfer requires third-party tools:

```bash
# Export Brave history to JSON
pip install browser-history
browser-history --browser brave -f json > brave_history.json

# Import to Firefox (limited tool support)
```

For most users, **accepting the loss of browsing history** and starting fresh is the pragmatic choice.

---

### Recommended Workflow

1. **Export bookmarks** from Brave → import to Firefox
2. **Transfer passwords** using CSV or just re-login as needed
3. **Accept** that history won't transfer cleanly

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

### Updating OpenRiot

OpenRiot upgrades are handled automatically. When a new version is released, Waybar will notify you. Click the update indicator to upgrade.

#### How Upgrades Work

| Scenario              | What Happens                                                       |
| --------------------- | ------------------------------------------------------------------ |
| **Fresh install**     | Clones repo, installs packages, builds source, deploys configs     |
| **Version available** | Pulls latest from git, re-runs package install, re-deploys configs |
| **Same version**      | Re-deploys configs only (preserves existing settings)              |

#### Upgrade Paths

**Automatic (Waybar):**

1. Waybar shows update indicator when new version available
2. Click the indicator → confirmation dialog
3. Confirm → upgrade runs in terminal

**Manual:**

```bash
# Same command works for fresh install and upgrade
curl -fsSL https://openriot.org/setup.sh | sh
```

The script automatically detects:

- No existing install → fresh install
- Older version → upgrade (git pull + re-run)
- Same version → config refresh only

All package installation uses `pkg_add -D unsigned` — fresh packages matching the current OpenBSD release are always fetched from the CDN.

<a id="advanced-usage"></a>

## 🧰 Advanced Usage

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

| Module      | Config Section    |
| ----------- | ----------------- |
| Workspaces  | `sway/workspaces` |
| CPU         | `cpu`             |
| Memory      | `memory`          |
| Temperature | `temperature`     |
| Battery     | `battery`         |
| Network     | `network`         |
| Volume      | `volume`          |
| Clock       | `clock`           |

### 🔐 Crypto Config

**Weather (Waybar)**

- Requires: `stormy` package (auto-installed)
- Disable: `touch ~/.config/openriot/disable-weather`
- Enable: `rm ~/.config/openriot/disable-weather`
- Location config (optional):
    - `~/.config/waybar/weather.conf` or `~/.config/openriot/weather.conf`
    - Format: `LOCATION="City, CC"`

**Crypto on Lock Screen**

- Config file: `~/.config/crypto.toml` (copied on first install)
- Shows prices and optional P/L on swaylock background

#### Configuration

```toml
# ~/.config/crypto.toml
api_key = ""

[indicators]
rsi_period = 14
oversold = 30
overbought = 70
bb_period = 16
bb_std = 2.0

# Set held=0 to show price only
pairs = [
    { sym = "XMR", coin = "monero",    held = 0, entry = 0 },
    { sym = "ZEC", coin = "zcash",     held = 0, entry = 0 },
    { sym = "BTC", coin = "bitcoin",    held = 0, entry = 0 },
]

[display]
show_totals = false
max_pairs = 6
```

**Quick rules:**

- Add coin: `{ sym = "SYM", coin = "coin-gecko-id", held = 0, entry = 0 }`
- Show P/L: set `held > 0` AND `entry > 0`
- Show totals: `show_totals = true`

### 🔒 Mullvad VPN on OpenBSD

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

## 🔧 Troubleshooting

### Upload Logs for Support

If you need to share logs with someone for debugging:

```bash
# Upload a log file and get a shareable link
curl -F "file=@~/.cache/openriot/setup.log" https://urlz.li/upload
```

This uses `curl` (available by default on OpenRiot and via pkg_add on OpenBSD) and returns a short URL you can share.

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

    OpenRiot uses **NetworkManager** (`nmtui`) for WiFi management — a simple TUI that works great in foot.

    Click the **network icon in Waybar** or run `nmtui` in a terminal to connect, manage saved networks, and enter passwords.

    ```bash
    # Install NetworkManager (done automatically by OpenRiot setup)
    doas pkg_add networkmanager
    doas rcctl enable networkmanager
    doas rcctl start networkmanager

    # Open the network manager UI
    nmtui

    # For manual OpenBSD-style config instead, edit:
    doas vi /etc/hostname.iwn0
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

> "You are absolutely deluded, if not stupid, if you think that a worldwide collection of software engineers who can't write operating systems or applications without security holes, can then turn around and suddenly write virtualization layers without security holes." — Theo de Raadt

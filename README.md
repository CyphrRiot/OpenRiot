<div align="center">

<img src="OpenRiot.png" alt="OpenRiot" width="100%"/>


## One command. Complete OpenBSD desktop. Zero compromises.

![Version](https://img.shields.io/badge/version-6.8-blue?labelColor=0052cc)
![License](https://img.shields.io/github/license/CyphrRiot/OpenRiot?color=4338ca&labelColor=3730a3)
![Platform](https://img.shields.io/badge/platform-OpenBSD-4338ca?logo=openbsd&logoColor=white&labelColor=3730a3)
![i3](https://img.shields.io/badge/i3-X11-312e81?logo=x11&logoColor=a855f7&labelColor=1e1b4b)

![Last Commit](https://img.shields.io/github/last-commit/CyphrRiot/OpenRiot?color=5b21b6&labelColor=4c1d95)
![Code Size](https://img.shields.io/github/languages/code-size/CyphrRiot/OpenRiot?color=4338ca&labelColor=3730a3)
![Code](https://img.shields.io/badge/human-coded-blue?logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNmZmZmZmYiIHN0cm9rZS13aWR0aD0iMiIgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIiBzdHJva2UtbGluZWpvaW49InJvdW5kIiBjbGFzcz0ibHVjaWRlIGx1Y2lkZS1wZXJzb24tc3RhbmRpbmctaWNvbiBsdWNpZGUtcGVyc29uLXN0YW5kaW5nIj48Y2lyY2xlIGN4PSIxMiIgY3k9IjUiIHI9IjEiLz48cGF0aCBkPSJtOSAyMCAzLTYgMyA2Ii8+PHBhdGggZD0ibTYgOCA2IDIgNi0yIi8+PHBhdGggZD0iTTEyIDEwdjQiLz48L3N2Zz4=&logoColor=a855f7&labelColor=1e1b4b)
![Language](https://img.shields.io/badge/language-Go-4338ca?logo=go&logoColor=c7d2fe&labelColor=3730a3)
![Language](https://img.shields.io/badge/language-YAML-5b21b6?logo=yaml&logoColor=e0e7ff&labelColor=4c1d95)

</div>

---
> “I just want OpenBSD to install and actually feel like a usable desktop without 100 hours of pain.”

![OpenRiot Desktop](assets/screenshot.png)

---

## 𝕆𝕡𝕖𝕟ℝ𝕚𝕠𝕥

OpenRiot is a clean, minimal, ridiculously polished i3-based OpenBSD 7.9 setup with fish, Helix, and Polybar — all tuned so things just work. No more config drama, no more obscure package fights, no more “works on Linux” copium.

Install base OpenBSD 7.9 → run one script → get a fast, dark, beautiful desktop that respects OpenBSD’s strengths instead of fighting them and adds all the workflow you need to be productive as a user and a developer.

- Read the [original Post on X](https://x.com/CyphrRiot/status/2043140621398127012?s=20)

### **Curated to be correct**

- **🪟 i3 Tiling** — X11-native tiling that actually gets it right ([tutorial](https://www.youtube.com/watch?v=Wx0eNaGzAZU))
- **⚡ Robust Binary** — Atomic operations, run-time, rollbacks, no dependency hell
- **🛡️ Privacy** — Zero telemetry, tracking, data harvesting, or ID requirements
- **🎨 Aesthetics** — Carefully crafted dark themes that work at any hour
- **💻 Development** — Helix, shell enhancements, crush, and other upgrades
- **💳 Monero** — Private by default, GUI wallet with polybar integration
- **💎 OpenBSD** — The most security-audited OS on the planet


#### Built on OpenBSD.

**Because compromises belong on other operating systems.**

Built with the same principles as [ArchRiot](https://ArchRiot.org) and by the same creator. If you liked ArchRiot, you'll love OpenRiot.

---

## ⚠️ **The Usual Free Software Warning** ⚠️

OpenRiot is under active development. It may not work as expected. Some features might be broken. Use at your own risk. Blah blah.

**Hardware Diversity:** Every system is unique — different network cards, WiFi chipsets, video cards, storage controllers, and countless other components. We've done our best to handle every possible configuration, but it's simply impossible to be completely comprehensive.

**Found an issue?** [Open an issue on GitHub](https://github.com/CyphrRiot/OpenRiot/issues) and we'll work through it together.

**Repository:** [github.com/CyphrRiot/OpenRiot](https://github.com/CyphrRiot/OpenRiot)

### System Requirements

| Requirement | Notes |
| --- | --- | --- |
| **Resolution** | 1920x1080 minimum | OpenRiot's User Interface requires this |
| **RAM** | 4GB+ minimum | 8GB+ Optimal |
| **Disk** | 25GB+ recommended | 100GB+ Optimal |

> "Linux has never been about quality. There are so many parts of the system that are just these cheap little hacks, and it happens to run." -Theo de Raadt

#### Xenocara's Hardening (OpenBSD's Custom X11 Server)

> "Why it is acceptable to move from Wayland and Sway to a fcking X11 desktop when everyone knows X11 is complete shit."

Xenocara is not vanilla X.Org. It is OpenBSD's integrated, heavily patched build of the X server with these security features:

- **Privilege separation**: The server runs with minimal privileges; input and rendering are isolated.
- **Pledge(2) and unveil(2)**: The X server itself and many clients are sandboxed.
- **No unnecessary setuid root**: Modern Xenocara drops privileges aggressively.
- **Stronger default configuration**: Fewer extensions enabled by default, audited for local attacks.

This makes the underlying X server far more resistant to client-side abuse than stock Xorg on Linux. Xenocara users generally consider it one of the more secure X11 implementations available.

---


## 📚 Navigate This Guide

- [🚀 Installing OpenRiot](#installing-openriot)
- [⌨️ Master Your OpenRiot Desktop](#master-your-openriot-desktop)
- [📝 Using Helix (Editor)](#using-helix)
- [🔄 System Management](#system-management)
- [🧰 Advanced Usage](#advanced-usage)
    - [🔄 Environment Variables](#environment-variables)
    - [⌨️ Keybindings Customization](#keybindings-customization)
    - [🔒 MAC Address Randomization](#mac-address-randomization)
    - [📊 Polybar Modules](#polybar-modules)
    - [🌤 Weather Module](#-weather-module-polybar)
    - [🔐 Crypto Config](#-crypto-config)
    - [🔒 WireGuard VPN](#-wireguard-vpn)
    - [📥 Transmission](#-transmission-bittorrent-client)
    - [📂 Proton Drive](#-proton-drive-sync)
    - [💳 Monero Wallet](#-monero-wallet)
- [🔧 Troubleshooting](#troubleshooting)
- [🦊 Browser & Data Transfer](#browser--data-transfer)

## ✅ Supported Systems

### Highly Recommended ThinkPad Series

These ThinkPads have excellent OpenBSD support for WiFi, trackpoints, and suspend/resume:

| Model                 | CPU               | WiFi                         | Notes                                     |
| --------------------- | ----------------- | ---------------------------- | ----------------------------------------- |
| **T14s Gen 1+** | Intel i5/i7 or AMD Ryzen | ⭐⭐⭐ `iwm` (AX200 adapter) | Best OpenBSD laptop                       |
| **T490**              | Intel i5-8265U    | ⭐⭐ `iwm` (Intel 9560)      | Good experience overall                   |
| **T480**              | Intel i5-8350U    | ⭐⭐ `iwm` (Intel 8265)      | Works well, slightly older                |
| **X1 Carbon Gen 7**   | Intel i7-8665U    | ⭐⭐ `iwm` (Intel 9560)      | Premium build, good Linux/OpenBSD support |
| **X270**              | Intel i5-6300U    | ⭐ `iwm` (Intel 8265)        | Small, portable, older but solid          |

You can buy a T14s Gen 1 for ~$300 USD at [Amazon](https://www.amazon.com/dp/B086MD6LTM). You can also buy a [T14s Gen 1](https://www.amazon.com/Lenovo-Thinkpad-T14s-Laptop-1-8Ghz-4-9GHz/dp/B0DKT19DQJ) for around the same price. Both are rock-solid, tested, and work perfectly out of the box.

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
- **GPU**: Intel integrated graphics are best-supported. AMD Radeon works but with varying feature support. NVIDIA is not supported on OpenBSD.
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

If you run into trouble, there's a [video on installing OpenBSD](https://www.youtube.com/watch?v=uTrOIPIx7pY) that walks through much of what you need to know.

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

6. **Disable Intel VTD** (if you encounter i3/X11 issues)
    - Navigate to `Security` → `Intel VT-d` or `AMD-Vi` → **Disabled**
    - Note: This is only needed in rare cases. Try with it enabled first.

7. **Set SATA mode to AHCI** (not RAID/Intel RST)
    - Navigate to `Storage` → `SATA Mode` → **AHCI**
    - RAID mode can cause OpenBSD to not see the disk

### Pre-Installation Checklist

Before booting the OpenBSD ISO:

- USB drive created with OpenBSD ISO (see above)
- Secure Boot disabled in BIOS
- Boot mode set to UEFI
- USB boot enabled
- SATA mode set to AHCI
- BIOS defaults loaded if you made many changes
- CMOS battery healthy (or laptop plugged in) to preserve settings

### Why This Matters for OpenBSD

OpenBSD is more conservative than Linux about hardware defaults. It assumes a clean, standards-compliant UEFI environment. Secure Boot, fast boot, and RAID modes are all Microsoft/Intel/AMD-specific optimizations that OpenBSD doesn't use — they can cause boot failures, disk recognition issues, or prevent i3 from starting.

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

For the best OpenBSD + i3 experience:

| Device         | Recommendation                                               |
| -------------- | ------------------------------------------------------------ |
| **Mouse**      | Basic USB mouse (2.4GHz wireless with dongle also works)     |
| **Keyboard**   | Any USB keyboard; ThinkPad keyboards work perfectly          |
| **TrackPoint** | Works natively on ThinkPads — no configuration needed        |
| **Graphics**   | Intel iGPU preferred; AMD Radeon works; NVIDIA not supported |

<a id="choose-your-openriot-experience"></a>

<a id="installing-openriot"></a>

## 🚀 Installing OpenRiot

There are two ways to install OpenRiot on a fresh machine.

**Primary path (recommended):** Install standard OpenBSD from the official
`install79.img`, then run one command to fetch and configure OpenRiot.
Requires internet during setup.

**Experimental path (in testing):** Flash a custom `openriot.img` that
includes all OpenRiot packages pre-bundled. The base OpenBSD install and
package installation happen offline, but you still run `setup.sh` after
first login to fetch the latest OpenRiot repository and configs.

> Typical time: **5 minutes** standard path with fast internet; **10–15
minutes** offline image path (no package downloads, but repo fetch still
requires network).

### 1. Download

For the standard path, download the official OpenBSD installer:

#### Disk Image (Best for USB)

The `.img` file is pre-configured for USB boot. **Recommended for most users.**

Download: [install79.img](https://cdn.openbsd.org/pub/OpenBSD/snapshots/amd64/install79.img)

#### OpenRiot Installer Image (Experimental)

A custom `openriot.img` is under development. It bundles `install79.img`
with pre-downloaded OpenRiot packages as a custom install set. This
eliminates the `pkg_add` download phase during installation, but still
requires `curl setup.sh` after first boot to deploy the OpenRiot
repository, configs, and desktop environment.

Not yet available for download. Build from source with `doas make img` if
you want to test it.

### 2. Flash to USB

> ⚠️ **Works for any `.img` file** — `install79.img` (standard) or `openriot.img` (experimental).

**For Linux:**
```bash
dd if=install79.img of=/dev/sdX bs=4M status=progress oflag=sync
```
(Replace `/dev/sdX` with your actual USB device)

**For OpenBSD:**
```bash
doas dd if=install79.img of=/dev/rsdXc bs=1M
sync
```
(Use raw disk device `/dev/rsdXc`, find with `dmesg | grep ^sd`)

Or, for progress on OpenBSD:
```bash
dmesg | grep ^sd   # get your usb sdX number
doas pkg_add pv    # if not installed already
pv -tpreb install79.img | doas dd of=/dev/rsdXc bs=1M
sync
```

---

### Method 1: Standard OpenBSD Install (Recommended)

Boot from the flashed USB and install OpenBSD normally. Then run
`setup.sh` to install OpenRiot.

1. Disable Secure Boot, set USB first in boot order
2. At `boot>` prompt, type `I` and press Enter
3. Follow the interactive prompts:

| Prompt               | Action                                                |
| -------------------- | ----------------------------------------------------- |
| Keyboard layout      | Press `Enter`                                         |
| System hostname      | Type `openriot` → Enter                               |
| Network interface    | Type `done` → Enter                                   |
| IPv4 autoconf        | Press `Enter`                                         |
| IPv6                 | Type `none` → Enter                                   |
| Root password        | Type and confirm password                             |
| Start sshd           | Press `Enter`                                         |
| X Window System      | Type `no` → Enter                                     |
| Setup a user         | Type username → Enter, then set password              |
| Which disk           | Type `sd1` (USB boot: sd0=USB, sd1=target)         |
| Use (W)hole disk MBR | ⚠️ **Choose `G` for GPT** (MBR won't boot)        |
| Encrypt disk         | Type `p` or `no`                                      |
| Partition layout     | Type `c` for custom                                   |
| Label editor         | `z` → `a /` → size → `a swap` → `a /home` → `w` → `q` |
| Location of sets     | **Recommended:** Type `disk` → Select your USB device    |
|                     | *(Sets are on the install image — no download needed.)*  |
|                     | **Alternative:** Type `http` → Use `http` or `httpcd`    |
|                     | *(Requires snapshot path since 7.9 is -current.)*        |
| Set name(s)          | Press `Enter` (all sets) or type specific sets           |
| SHA256 verification | Type `yes` → Enter                                   |

**Partition layout (choose `c`):**
```
/       50G (or more)
/home   * (rest of disk)
swap    2G (or more)
```

#### Reboot

```bash
reboot
```

#### After reboot — configure doas and run setup

Log in as **root** first:

```bash
# Configure doas (passwordless sudo for OpenBSD)
vi /etc/doas.conf

# Or create with:
tee /etc/doas.conf << 'EOF'
# OpenRiot-style: your user + wheel group, no password
permit nopass $USER
permit nopass :wheel

# Optional: keep your environment (PATH, etc.)
permit nopass keepenv $USER
permit nopass keepenv :wheel
EOF
```

Now log in as **your user** and run:

```bash
doas pkg_add -D snapshot curl
curl -fsSL https://openriot.org/setup.sh | sh
```

> **🌍 Auto Mirror Selection:** OpenRiot automatically detects and uses the fastest OpenBSD mirror during installation — no manual configuration needed. Run `openriot --mirrors` anytime to check or update your mirror selection.

The desktop will start automatically after setup completes. No need to run `startx`.

---

#### Log Locations

If something goes wrong:

| Stage                | Log File                        |
| -------------------- | ------------------------------- |
| `setup.sh`           | `~/.cache/openriot/setup.log`   |
| `openriot --install`  | `~/.cache/openriot/install.log` |

```bash
cat ~/.cache/openriot/setup.log
cat ~/.cache/openriot/install.log
```

---

### Method 2: OpenRiot Installer Image (Experimental)

Download the pre-built `openriot.img` from OpenRiot.org. It bundles OpenBSD
7.9 with all OpenRiot packages pre-loaded as a custom install set — no
`pkg_add` downloads during installation.

After the OpenBSD base install completes, packages install automatically from
the USB media. The system reboots to a console login. Log in as your user
and run:

```bash
curl -fsSL https://openriot.org/setup.sh | sh
```

This fetches the latest OpenRiot repository, deploys configs, and completes
the setup. Requires network during `setup.sh` execution. X11 does **not**
start automatically on first boot — run `startx` after setup completes.

#### Install Steps

1. Download `openriot.img` *Coming Soon*
2. Flash to USB:
   **Linux:** `dd if=openriot.img of=/dev/sdX bs=4M status=progress`
   **OpenBSD:** `doas dd if=openriot.img of=/dev/rsdXc bs=1M`
3. Boot from USB, type `I` at the `boot>` prompt
4. Answer the standard OpenBSD prompts (disk, hostname, user, password)
5. When prompted for installation sets, select `site79.tgz` (or press Enter
   to include all sets)
6. Reboot when finished
7. Log in as your user and run `curl -fsSL https://openriot.org/setup.sh | sh`

---

### WiFi Setup

If you need to configure WiFi on OpenBSD (interface names like `iwx0`, `athn0`, `urtwn0`):

```bash
# Create hostname file
doas vi /etc/hostname.iwx0

# Add these lines:
join "MyNetwork" wpakey "Password"
autoconf

# Save and exit, then start the interface:
doas sh /etc/netstart iwx0
```

---

<a id="master-your-openriot-desktop"></a>

![OpenRiot Desktop](assets/menu.png)

## ⌨️ Master Your OpenRiot Desktop

_We use Helix instead of `vi` or `vim`. The essential bindings are documented in [📝 Using Helix](#using-helix). A full OpenRiot keybindings reference is coming._

### Essential Keybindings

| Key                          | Action                           |
| ---------------------------- | -------------------------------- |
| `Super + Return`             | Open terminal                    |
| `Super + Shift + Return`     | Floating terminal                |
| `Super + D`                  | Open app launcher (rofi)         |
| `Super + Q`                  | Close window                     |
| `Super + E`                  | Proton Mail (web app)            |
| `Super + L`                  | Lock screen                      |
| `Super + Z`                  | Toggle floating                  |
| `Super + H`                  | Split horizontal                 |
| `Super + P`                  | Toggle layout                    |
| `Super + Shift + F`          | Toggle fullscreen                |
| `Super + 1-4`               | Switch workspace                 |
| `Super + Shift + 1-4`        | Move window to workspace         |
| `Super + Shift + E`          | Exit i3                          |
| `Super + F`                  | File Manager (Thunar)            |
| `Super + B`                  | Browser (Firefox)                |
| `Super + O`                  | Gnome Text Editor                |
| `Super + C`                  | Open Crush AI                    |
| `Super + T`                  | Open system monitor (btop)      |
| `Super + G`                  | Telegram                         |
| `Super + S`                  | Signal (gurk CLI)                |
| `Super + M`                  | Google Messages                  |
| `Super + X`                  | X (Twitter)                      |
| `Super + W`                  | Next wallpaper                   |
| `Super + Shift + W`          | Previous wallpaper               |
| `Super + Shift + S`          | Screenshot (region)              |
| `Super + Shift + R`          | Screen recording toggle          |
| `Super + Shift + V`          | Clipboard manager                 |
| `Super + Shift + G`          | Open settings menu               |
| `Super + Shift + J`          | Open games menu                  |
| `Super + Shift + H`          | OpenRiot Help (website)          |
| `Super + Escape`             | Power menu                       |
| `Super + =`                  | Calculator (rofi)                |
| `Super + [ / ]`              | Resize: shrink/grow width       |
| `Super + Shift + [ / ]`      | Resize: shrink/grow height      |
| `Super + -`                  | Show scratchpad                  |
| `Super + Shift + -`          | Move to scratchpad               |
| `Super + Shift + C`          | Reload i3 config                 |
| `Super + Shift + R`          | Restart i3                       |
| `Super + Tab`                | Focus next window                |
| `Super + Shift + Tab`        | Focus previous window             |
| `Super + Arrow keys`         | Focus window direction           |
| `Super + Shift + Arrow`     | Move window to direction        |
| `Super + button4/5`         | Scroll workspaces                |
| `Print`                     | Screenshot (fullscreen + clipboard) |
| `Shift + Print`             | Screenshot (fullscreen + clipboard) |
| `Ctrl + Print`              | Screenshot (fullscreen + clipboard) |
| `Super + Shift + X`         | Compose tweet                    |
| `Super + Shift + space`     | Refresh polybar                  |


### Media Keys

| Key                    | Action                           |
| ---------------------- | -------------------------------- |
| `Volume +/-`           | Adjust volume                    |
| `Mute`                 | Toggle mute                      |
| `Mic Mute`             | Toggle microphone mute           |
| `Brightness +/-`       | Adjust screen brightness         |


### System Hotkeys

Control polybar modules and system features directly from the keyboard.

| Key                    | Action                           |
| ---------------------- | -------------------------------- |
| `Super + N`            | Toggle night light (redshift)   |
| `Super + V`            | Toggle audio mute                |
| `Super + U`            | Check for OpenRiot updates       |
| `Super + Y`            | Show crypto prices notification  |
| `Super + A`            | Show battery status notification |
| `Super + I`            | Toggle WireGuard VPN             |
| `Super + J`            | Sync Proton Drive               |
| `Super + Shift + P`    | Show CPU notification            |
| `Super + Shift + M`    | Show memory notification         |
| `Super + Shift + N`    | Show WiFi info                   |
| `Super + Shift + G`    | Open settings menu               |
| `Super + Shift + J`    | Open games menu                 |


### App Launcher (Rofi)

Press `Super + D` to open the app launcher. Only curated apps are shown — no system clutter.

| App              | Icon | Description              |
| ---------------- | ---- | ------------------------ |
| Terminal         |    | Alacritty terminal       |
| File Manager     | 󰝰   | Thunar file browser      |
| Firefox          |    | Web browser              |
| Firefox (Private)|    | Private browsing         |
| Text Editor      | 󰷉   | GNOME text editor        |
| Helix            |    | Text editor              |
| Word Processor   | 󰈙   | LibreOffice Writer       |
| Media Player     |    | mpv video player         |
| Proton Mail      | 󰊫   | Email (web app)          |
| Signal           | 󰬚   | Signal messenger (gurk)  |
| System Monitor   | 󰍹   | btop resource monitor    |
| Telegram         | 󰭹   | Messaging app            |
| Monero Wallet    |    | Monero GUI wallet        |
| Transmission     | 󰐻   | BitTorrent client        |
| Crush AI         | 󰚩   | AI CLI assistant         |
| Settings         | 󰒓   | XFCE settings manager    |
| Games            | 󰊗   | Games sub-menu           |

![OpenRiot Terminal](assets/terminal.png)

### Top Menu (Polybar)

Polybar is your status bar. Click on modules for more:

| Module | Click Action |
| ------ | ------------- |
|  Launcher | Opens app launcher |
| 󰎤󰎧󰎪 Workspaces 1-3 | Click to switch workspace |
| Window Title | Shows focused window name |
| 󰃭 Date | Shows date/time |
| weather | Shows current temp + conditions (OpenWeatherMap) |

| No. | Module | Icon | Meaning |
|-----|--------|------|---------|
| 1 | crypto | 󰄨 | Crypto prices (hidden when no config) |
| 2 | proton-drive | 󱥾 / 󰴋 /  | Synced / Syncing / Not configured |
| 3 | transmission | 󰐻 | Running (auto-hides when stopped) |
| 4 | night-light | 󰌵 /  | Night light ON / OFF |
| 5 | screenrec |  /  | Idle / Recording |
| 6 | cpu | 󰡳→󰡵→󰊚→󰡴 | CPU load tier |
| 7 | memory | 󰢿→󰢼→󰢽→󰢾 | RAM usage tier |
| 8 | wireguard | 󰱓 / 󰅛 | VPN up / Down (auto-hides when not configured) |
| 9 | stealth | 󰝴 / 󱊨 | MAC randomization ON / OFF |
| 10 | network-wifi | 󰤨→󰤥→󰤢→󰤟→󰤯 / 󱛅 | Signal bars / No internet |
| 11 | network-eth | 󰈀 / 󰌙 / (empty) | Carrier / No carrier / No eth |
| 12 | openriot-update | 󰋻 / 󰚇 / ? | Update available / Up to date / Unknown |
| 13 | settings |  | Settings menu |
| 14 | laptop-monitor | 󰌢 / 󰛧 | Laptop monitor enabled / disabled (auto-hides) |
| 15 | volume | //󱄠/󰕾/ | Muted / Very low / Low / Medium / High |
| 16 | battery | 󰁺→󰂂 / 󰁹 | Level / Full+charging |
| 17 | power | ⏻ | Power menu |
| 18 | lock | 󰌾 | Lock screen |

**Workspace Bar:** Shows all 3 workspaces with indicators and app icons.

```
 󰞷 󰈹       󰝰
```
- `` focused workspace
- `` urgent workspace
- `` unfocused with windows
- `` empty workspace
- Icons show running apps: `` Alacritty, `󰈹` Firefox, `󰝰` Thunar, etc.

## Shell Aliases & Quick Reference

Fish comes pre-configured with useful aliases:

| Alias | Command   | Description                    |
| ----- | --------- | ------------------------------ |
| `ls`  | `lsd`     | Default listing with icons     |
| `ll`  | `lsd -l`  | Long listing with icons        |
| `la`  | `lsd -la` | Show hidden files              |
| `vi`  | `hx`      | Open Helix editor              |
| `vim` | `hx`      | Open Helix editor              |

### File Manager

OpenRiot uses **Thunar** (xfce file manager) as the primary file manager.

### Password Management with Glyphriot

OpenRiot includes **[Glyphriot](https://github.com/CyphrRiot/glyphriot)**, a secure password manager that uses a memorable seed phrase and optional glyph to derive your master password. Run with:

```fish
glyphriot --prompt
```

This prompts for your seed and optional glyph, then derives the master password using Argon2id.

**Key features:**
- Seed + glyph → master password (never stored)
- Supports multiple services
- Encrypted storage with age
- Master password derived on-demand

**Security notes:**
- Your seed is never stored — only the derived hash
- Use a strong, unique seed you can remember
- Add a glyph for extra security (optional but recommended)

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
| **Ungoogled Chromium** | ✅ Available     | `pkg_add ungoogled-chromium` (not installed by default) |
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

OpenBSD doesn't use `apt update` or `pacman -Syu`. Use `pkg_add -u` to update packages:

```bash
# Update all packages
doas pkg_add -u

# Install a new package
doas pkg_add <package-name>

# Remove a package
doas pkg_delete <package-name>
```

### Updating OpenRiot

OpenRiot upgrades are handled automatically. When a new version is released, Polybar will notify you. Click the update indicator to upgrade.

#### How Upgrades Work

| Scenario              | What Happens                                                       |
| --------------------- | ------------------------------------------------------------------ |
| **Fresh install**     | Clones repo, installs packages, builds source, deploys configs     |
| **Version available** | Pulls latest from git, re-runs package install, re-deploys configs |
| **Same version**      | Re-deploys configs only (preserves existing settings)              |

#### 󰋻 Upgrade Paths

**Automatic (Polybar):**

1. Polybar shows update indicator when new version available
2. Click the indicator → confirmation dialog
3. Confirm → upgrade runs in terminal

**Manual:**

```bash
# Same command works for fresh install and upgrade
curl -fsSL https://openriot.org/setup.sh | sh
```

**Install a specific version (tag):**

```bash
openriot --install v4.5
```

The script automatically detects:

- No existing install → fresh install
- Older version → upgrade (git pull + re-run)
- Same version → config refresh only

> **Note:** Upgrade also updates OpenBSD packages to the latest stable versions via `pkg_add -u`. This applies to both the Polybar upgrade click and `curl -fsSL https://openriot.org/setup.sh | sh`.

Package installation uses `pkg_add -D snapshot` only when running OpenBSD -current (snapshots). For stable releases (e.g., 7.9, 8.0), packages are fetched from the release CDN without the snapshot flag.

<a id="advanced-usage"></a>

## 🧰 Advanced Usage

### Environment Variables

OpenRiot sets sensible defaults. Key environment variables:

```bash
# Check OpenRiot version
openriot --version

# XDG directories (usually correct by default)
echo $XDG_CONFIG_HOME
echo $XDG_DATA_HOME

# Fish is the default shell
echo $SHELL  # Should show /usr/local/bin/fish
```

### Keybindings Customization

Keybindings are in `~/.config/i3/keybindings.conf`.

Edit this file to customize. After saving, press `Super + Shift + R` to reload i3.

<a id="mac-address-randomization"></a>

### 🔒 MAC Address Randomization

OpenRiot includes automatic MAC address randomization for network interfaces. This prevents network operators and observers from tracking your device across different networks by using a randomly generated MAC address on each connection.

![Stealth Mode](assets/stealth.png)

#### How It Works

The system uses `ifconfig` to spoof MAC addresses on both WiFi and Ethernet interfaces. Each time you connect to a network, a new random MAC is generated, making device fingerprinting significantly more difficult.

#### Enable/Disable

```bash
openriot --random-mac enable    # Enable MAC randomization
openriot --random-mac disable   # Disable (use real MAC)
openriot --random-mac show      # Check current status
```

#### Privacy Benefits

- **Network Tracking Prevention** — Your real MAC is never exposed on public networks
- **[Stealth]** — The polybar network module shows `[Stealth]` when enabled
- **Automatic** — Takes effect on every connection

#### Compatibility

Works with all supported network interfaces in OpenRiot. Some networks with MAC authentication (captive portals, corporate 802.1X) may require disabling randomization.

### Top Menu (Polybar)

Polybar modules are in `~/.config/polybar/config`.

Each module is a custom script that outputs icon + info for display. Modules update automatically and respond to clicks.

| Module | Icon | States | Click Action | Scroll |
|--------|------|-------|-------------|--------|
| **launcher** |  | - | Open app launcher (Rofi) | - |
| **workspaces** | 󰎤󰎧󰎪 | - | Switch to workspace | - |
| **window-title** | text | - | - | - |
| **date** | 󰃭 | - | Show date/time | - |
| **stealth** | 󰝴 | ON | Toggle MAC randomization | - |
| | 󱊨 | OFF | | |
| **network-wifi** | 󰤨 | Excellent (70%+) | WiFi info | - |
| | 󰤥 | Good (50-69%) | | |
| | 󰤢 | Fair (30-49%) | | |
| | 󰤟 | Poor (20-29%) | | |
| | 󰤯 | Very poor / no signal | | |
| | 󱛅 | No internet | | |
| **network-eth** | 󰈀 | Connected | Ethernet info | - |
| **wireguard** | 󰱓 | Connected | Toggle VPN | - |
| | 󰅛 | Disconnected | | |
| **volume** |  | High (76%+) | Toggle mute | Volume adjust |
| | 󰕾 | Medium (46-75%) | | |
| | 󱄠 | Low (11-45%) | | |
| |  | Very low (1-10%) | | |
| |  | Muted | | |
| **battery** | 󰁺-󰂂 | Discharging | Battery notification | - |
| | 󰢜-󰂋 | Charging | | |
| **crypto** | 󰄨 | - | Show crypto prices (hidden when no config) | - |
| **night-light** |  | OFF | Toggle redshift | - |
| | 󰌵 | ON | | |
| **cpu** | 󰡳→󰡵→󰊚→󰡴 | - | CPU notification | - |
| **memory** | 󰢿→󰢼→󰢽→󰢾 | - | Memory notification | - |
| **proton-drive** | 󱥾 | Synced | Sync Proton Drive | - |
| | 󰴋 | Syncing | | |
| |  | Not configured | | |
| **transmission** | 󰐻 | Running | Toggle GTK client | - |
| **openriot-update** | 󰋻 | Update available | Check for updates | - |
| | 󰚇 | Up to date | | |
| **settings** |  | - | Open settings menu | - |
| **laptop-monitor** | 󰌢 / 󰛧 | Enabled / Disabled | Toggle laptop monitor | - |
| **weather** | varies | - | - | - |
| **power** | ⏻ | - | Open power menu | - |
| **lock** | 󰌾 | - | Lock screen | - |

### 🌤 Weather Module (Polybar)

The weather module shows current temperature and conditions in the polybar status bar using OpenWeatherMap API.

**Configuration:**

1. Create weather config at `~/.config/weather.cfg`:

```ini
location=Las Vegas
units=imperial
api=85a4e3c55b73909f42c6a23ec35b7147
```

- `location` - City name (required)
- `units` - `imperial` (°F) or `metric` (°C)
- `api` - OpenWeatherMap API key (optional, uses built-in key if omitted)

2. Restart polybar: `Super + Shift + space`

**If no config exists**, the weather module is hidden automatically.

**Weather Icons:**

| Code | Condition | Icon |
|------|-----------|------|
| 01x | Clear sky | 󰖕 |
| 02x | Few clouds |  |
| 03x/04x | Scattered/broken |  |
| 09x | Drizzle |  |
| 10x | Rain |  |
| 11x | Thunderstorm |  |
| 13x | Snow |  |
| 50x | Mist/Fog | 󰖑 |
| default | Unknown | 󰨹 |

---

### 🔐 Crypto Config

**Crypto on Lock Screen**

- Config file: `~/.config/crypto.toml` (copied on first install)
- Shows prices and optional P/L on i3lock background

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
```

**Quick rules:**

- Add coin: `{ sym = "SYM", coin = "coin-gecko-id", held = 0, entry = 0 }`
- Show P/L: set `held > 0` AND `entry > 0`
- Show totals: `show_totals = true`

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

- `Super + O` — Open GNOME Text Editor
- Run `hx` in any terminal to open Helix

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
| Go to bottom of document   | `G`                        | `G`                        | `G` goes to bottom of document (OpenRiot remaps this).                                                                             |
| Delete character           | `x`                        | `x`                        | Selects entire line in Helix. Use `dl` to delete line, or `d` after selection.                                                      |
| Delete line                | `dd`                       | `dl`                       | Use `dl` (delete line under cursor).                                                                                               |
| Go to end of line          | `$`                        | `gl`                        | `gl` = goto line end. Very common.                                                                                                                         |
| Go to start of line        | `0` or `^`                 | `gh`                        | `gh` = goto home (start of line). Use `gs` if you want the first non-whitespace character (like Vim's `^`).                                                |
| Copy line (yank line)      | `yy`                       | `yl`                        | `yl` yanks the current line.                                                                                                                               |
| Paste line                 | `p` (below) or `P` (above) | `p` (after) or `P` (before) | Works similarly, but Helix pastes after/before the current selection (or cursor position). For a full line paste, the behavior is usually what you expect. |
| Copy text (yank selection) | `y` (after selecting)      | `y`                         | Same letter, but you select first (e.g., `w` for word, `gl` for to end of line, or visual movements).                                                      |
| Paste text                 | `p` or `P`                 | `p` or `P`                  | Same as above. Helix also supports system clipboard via `<space>p` / `<space>y` (or configure defaults).                                                   |

### Helix on OpenBSD & OpenRiot

Helix works **beautifully** on OpenBSD:

- Excellent performance on ThinkPads and Framework laptops
- Native OpenBSD packaging (`pkg_add helix`)
- Full Tree-sitter and LSP support for Go, Rust, Python, Lua, YAML, TOML, and many other languages
- No plugin manager headaches — everything just works
- Plays perfectly with i3, Alacritty terminal, and fish shell

**Pro tip:** Helix has one of the best default dark themes available. It looks right at home with OpenRiot's dark aesthetic.

For the complete keymap and configuration options, visit the official documentation:  
[https://docs.helix-editor.com/](https://docs.helix-editor.com/)

_See the [helix-cheat-sheet](https://github.com/stevenhoy/helix-cheat-sheet) project for a visual keybinding reference._

**Tutorial Video:** [Helix Editor Crash Course](https://www.youtube.com/watch?v=HcuDmSb-JBU)

### AI Integration with OpenRouter

OpenRiot bundles **Crush** for AI-assisted coding. Crush is a modern, lightweight, Go-based terminal AI coding agent with excellent OpenBSD support. It is built automatically during setup and installed to `/usr/local/bin/crush`.

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

### 🔒 WireGuard VPN

OpenRiot includes a polybar module to toggle WireGuard VPN with a single click.

#### Prerequisites

1. Install WireGuard tools:
```bash
pkg_add wireguard-tools
```

2. Create the config directory:
```bash
doas mkdir -p /etc/wireguard
```

#### Setting Up Mullvad VPN

1. **Generate your config:**
   - Go to [mullvad.net/en/account/wireguard-config](https://mullvad.net/en/account/wireguard-config)
   - Select **Linux** platform (WireGuard works the same on OpenBSD)
   - Click **Generate Key**
   - Choose a server location (Country/City)
   - Download the config file

2. **Install the config:**
```bash
# Move the downloaded config to WireGuard directory
doas mv ~/Downloads/mullvad.conf /etc/wireguard/wg0.conf
```

#### Using the VPN

**Polybar Module:**

| Icon | Meaning |
---------------|
| 󰛳 | No config file installed |
| 󰅛 | Config exists, VPN disconnected |
| 󰱓 | VPN connected |

Click the icon to toggle. You'll see notifications for:
- "Starting WireGuard..."
- "Stopping WireGuard..."
- "WireGuard is not configured. Go to OpenRiot.org Read directions." (if no config)

#### Manual Commands

```bash
# Connect
doas wg-quick up /etc/wireguard/wg0.conf

# Disconnect
doas wg-quick down /etc/wireguard/wg0.conf

# Verify connection
curl https://am.i.mullvad.net/json
```

The output should show `"mullvad_exit_ip": true` when connected.

#### Auto-start at Boot

WireGuard automatically starts on boot if it was enabled when you last shut down.

**How it works:**
- Enable WireGuard via `Super+I` or the polybar icon → autostart flag is set
- Disable WireGuard → autostart flag is cleared
- On next i3 startup, WireGuard comes up automatically after network is ready

No manual configuration required. The state is stored at `~/.config/openriot/wireguard.enabled`.

#### Troubleshooting

**VPN won't connect:**
```bash
# Check if config exists
ls -la /etc/wireguard/wg0.conf

# Check interface
ifconfig wg0
```

**Slow speeds:**
- Try a different Mullvad server location
- Some Mullvad servers may have limited bandwidth

## 📥 Transmission BitTorrent Client

OpenRiot includes the Transmission GTK client with polybar and rofi integration.

### ⚠️ IMPORTANT: Use with VPN

**Always run Transmission behind a VPN!** Your ISP can see BitTorrent traffic, and you can receive copyright infringement notices (or worse) if you download copyrighted material. 😉

Click the **VPN icon** 󰱓 in polybar to connect before downloading anything.

### Polybar Module

The Transmission module auto-hides when the client is not running.

| Icon | Meaning |
---------------|
| 󰐻 | Transmission running |
| (hidden) | Not running |

Click the icon to toggle. Notifications confirm state changes.

### Rofi Menu

The app launcher (Rofi) also has a Transmission entry that dynamically shows:
- **Transmission** 󱧝 — Click to stop (running)
- **Transmission** 󰐻 — Click to start (stopped)

### Default Settings

- **Download directory:** `~/Downloads`
- **Blocklist:** Enabled (courtesy of [BT BlockLists](https://github.com/Naunter/BT_BlockLists))

### Manual Commands

```bash
# Check if running
pgrep transmission-gtk
```

## 📂 Proton Drive Sync

OpenBSD has no native Proton Drive client. OpenRiot includes **rclone** for end-to-end encrypted bidirectional file syncing.

### Complete Setup Guide

### 1. Create Proton Drive Folder

1. Log into [drive.proton.me](https://drive.proton.me)
2. Create a new folder named **`ProtonSync`** (case-sensitive)

### 2. Configure rclone

```bash
rclone config
```

| Prompt | Action |
----------------|
| `n` | New remote |
| Name | `ProtonSync` |
| Storage | `protondrive` |
| Proton email | Your email |
| Proton password | Your password |
| 2FA | Code if enabled |

### 3. Create Local Sync Folder

```bash
mkdir -p ~/Documents/ProtonSync
```

### 4. Initial Sync (dry-run first)

```bash
rclone bisync ~/Documents/ProtonSync proton:ProtonSync --dry-run --resync
```

If output looks correct, remove `--dry-run --resync` to sync.

### 5. Set Up Automatic Sync

Edit your crontab:
```bash
doas crontab -e
```

Add this line (replace `username` with your actual username):
```cron
*/15 * * * * /usr/local/bin/rclone bisync /home/username/Documents/ProtonSync proton:ProtonSync --fast-list >> /var/log/rclone.log 2>&1
```

### 6. Secure Your Config

```bash
chmod 600 ~/.config/rclone/rclone.conf
```

### How It Works

- **Polybar icon** 󱥾 synced, 󰴋 needs sync,  not configured
- **Click the icon** to sync (auto-init cache on first click)
- Files are encrypted client-side before transit (end-to-end encryption)

### Sync Between Multiple Systems

`rclone bisync` is bidirectional — it syncs both ways:
- Local changes → Proton Drive
- Proton Drive changes (from other systems) → Local

If both systems edit the same file, rclone creates a conflict file (`.sync_orig`) that you can manually resolve.

### Security Notes

- Your files remain encrypted end-to-end — Proton never sees unencrypted data
- rclone never sees your actual file contents
- Keep `rclone.conf` permissions at 600
- Run rclone as your normal user, never root

## 📨 Signal with Gurk

**Quick Overview:** [The Ultimate Signal Messenger Terminal Client Guide](https://www.blog.brightcoding.dev/2025/11/30/gurk-the-ultimate-signal-messenger-terminal-client-for-privacy-savvy-power-users-2025-guide/) — great introduction by Bright Coding.

A pure-Rust Signal messenger TUI — zero Java, zero GTK/libsecret. Built for OpenBSD.

### First Run

1. Launch from the app launcher (SUPER+D) or run `~/.local/bin/gurk`
2. On first launch it will prompt for a passphrase — **select "Store it in config"**, not "prompt" (prompt mode causes issues)
3. Open Signal on your phone → Linked Devices → add a new linked device → scan the QR code
4. Wait 2–3 minutes, then press `ctrl+p` to open the channel list

**Note:** Gurk does not remember channels or messages on startup — it starts clean and only updates when you receive messages. If the channel list stays empty, press `ctrl+p` to force the popup, wait 30–60 seconds, or send yourself a test message from your phone.

### Daily Workflow

| What | How |
|------|-----|
| Open channel popup | `ctrl+p` (most important key) |
| Switch channels | `ctrl+j` / `ctrl+k` or Up/Down |
| Read messages | Scroll with `alt+Up` / `alt+Down` or `PgUp` / `PgDn` |
| Select a message | `PgUp` / `PgDn` |
| Reply | Type your message + `Enter` |
| Open a link | `Enter` with empty input |
| View attachment | `Enter` on selected message |
| Multi-line input | `alt+Enter` |
| Send a file | `alt+Enter` then `file:///home/{user}/{path}` |
| Attach clipboard image | `alt+Enter` then `file://clip` |
| React to a message | Select it → type emoji → `tab` |
| Copy message | `alt+y` |
| Open help | `f1` |
| Deselect message | `ESC` |
| Mouse support | Click Edit field or Channel (not messages) |

Gurk supports the same emojis as Github. It's a little daunting at first, but here's a pretty comprehensive (and searchable) [Emoji Cheat Sheet](https://phw198.github.io/github-emoji-cheatsheet/) that makes it much easier.

### Configuration

Create or edit `~/.config/gurk/gurk.toml` to customize:

**Recommended keybindings** (add right above `[keybindings]`):
```toml
[keybindings.message_selected]
alt-e = "edit_message"           # Edit selected message
alt-y = "copy_message selected"  # Copy selected message
ctrl-t = "react :thumbsup:"      # React with 👍
ctrl-f = "react 🔥"              # React with 🔥
ctrl+h = "react :purple_heart:" # React with 💜
```

**Popular settings:**
```toml
[keybindings.normal]
ctrl-n = "select_channel next"   # Next channel
ctrl-p = "select_channel previous" # Previous channel
```

### Reactions

React to messages by selecting them (PgUp/PgDn), typing an emoji shortcode, then pressing Tab:

| Shortcode | Emoji | Meaning |
|-----------|-------|---------|
| `:thumbsup:` or `:+1:` | 👍 | Like / Agree |
| `:heart:` | ❤️ | Love |
| `:laughing:` or `:joy:` | 😂 | Funny |
| `:fire:` | 🔥 | Awesome |
| `:rocket:` | 🚀 | Great idea |
| `:eyes:` | 👀 | Watching |
| `:clap:` | 👏 | Well done |
| `:100:` | 💯 | Perfect |
| `:ok_hand:` | 👌 | OK |
| `:thinking:` | 🤔 | Thinking |
| `:sparkles:` | ✨ | New feature |
| `:bug:` | 🐛 | Bug |
| `:wrench:` | 🔧 | Fix |
| `:memo:` | 📝 | Docs |
| `:white_check_mark:` | ✅ | Done |
| `:construction:` | 🚧 | WIP |
| `:tada:` | 🎉 | Celebration |
| `:heavy_check_mark:` | ✔️ | Verified |

Or just paste Unicode emoji with `ctrl+Shift+V`.

## 💳 Monero Wallet

OpenRiot includes the Monero GUI wallet pre-built for OpenBSD with full desktop integration.

![Monero Wallet](assets/xmr.png)

### Launch

- **Rofi**: Press `Super + D` → select "Monero Wallet"
- **Desktop entry**: Available in the app launcher with  icon

### Prerequisites (installed automatically)

The following Qt5 QML runtime packages are included in `packages.yaml`:
- `qtquickcontrols`
- `qtquickcontrols2`
- `qtgraphicaleffects`
- `qtdeclarative`

### Runtime Environment

The desktop entry sets `QML2_IMPORT_PATH=/usr/local/lib/qt5/qml` automatically so Monero GUI finds Qt Quick modules.

`~/.local/bin` is added to PATH in both `~/.xinitrc` and `~/.xsession` so rofi finds the binary immediately after install — no re-login required.

`XDG_RUNTIME_DIR` permissions are set to `0700` at session startup to satisfy Qt's runtime checks.

### Features

- Full node or remote node wallet support
- Integrated with polybar crypto module (shows  when `~/.config/crypto.toml` exists)
- Window icon mapping via `config/window/icons.toml`

## 🎬 Screen Recording

Click the **** icon in polybar to start recording your screen. Click again to stop. No terminal needed.

### How It Works

- **Idle:** `` (teal, next to night-light)
- **Recording:** `` (red)
- **Output:** `~/Videos/Recordings/YYYYMMDD-HHMM.mp4`
- **Video:** H.264, 30fps, native resolution, `ultrafast` preset
- **Audio:** AAC 192kbps via `snd/0.mon` (when available)

### Dynamic Audio Monitoring

Instead of permanently reconfiguring `sndiod` at install time, the recorder temporarily enables audio monitoring:

1. Before recording: saves your current `sndiod` flags, sets monitor flags, restarts `sndiod`
2. After stopping: restores your original flags, restarts `sndiod`

Your music and mic stay clean when not recording. The monitor only exists while you're actually capturing.

### Notifications

- **Start:** "Screen Recorder / Recording is starting..."
- **Stop:** "Screen Recorder / Recording is stopping... / Saved to: ~/Videos/Recordings/..."

### Keyboard Shortcut

| Key | Action |
|-----|--------|
| `Super + Shift + R` | Toggle screen recording |

## 🔧 Troubleshooting

### Upload Logs for Support

If something goes wrong, upload your log file for debugging:

```bash
# Setup log location
cat ~/.cache/openriot/setup.log

# Install log location
cat ~/.cache/openriot/install.log

# Share install log
~/.local/share/openriot/install/openriot --share-log install.log
```

This will upload the log to tmpfiles.org and give you a URL to share.

### Hostname shows as `x.my.domain`

If the hostname prompt was left blank during install, OpenBSD sets a default domain of `my.domain`, making your hostname look like `openriot.my.domain`.

**Fix:**
```bash
doas vi /etc/myname
# Change: openriot.my.domain
# To:     openriot
# Then reboot.
```

### Multiple wsmouse devices in settings

You may see `wsmouse`, `wsmouse0`, `wsmouse2`, `wsmouse3`, etc. in mouse settings. This is normal OpenBSD kernel behavior.

- `wsmouse` — kernel mux device
- `wsmouse0` — trackpad touch surface
- `wsmouse2` — TrackPoint (red nub)
- `wsmouse3` — physical button / clickpad protocol layer

The gap at `wsmouse1` is skipped enumeration. Nothing is broken.

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

    OpenRiot uses **ifconfig** for WiFi management on OpenBSD.

    Click the **network icon in Polybar** or run `ifconfig iwm0 up` + `fw_update` to set up WiFi.

    ```bash
    # Connect manually via hostname.if(5):
    doas vi /etc/hostname.iwm0
    # Add: nwid "YourNetworkName" wpakey "YourPassword" dhcp
    # Add: inet autoconf
    # Add: mode 11g
    # Then: doas sh /etc/netstart iwm0
    ```

4. **After connecting:**
    ```bash
    # Verify connection
    ifconfig iwm0
    ping -c 3 cdn.openbsd.org
    ```

### Mouse freezes or jitters

If the mouse hangs after idle, then jitters/glitches when moved:

1. **Check for USB conflicts:**
    ```bash
    # List all USB devices and their controllers
    usbdevs -v

    # Check if mouse and other devices share the same USB bus
    dmesg | grep -E '^(uhub|usb).*port'

    # View detailed USB tree
    usbdevs -d
    ```

2. **Common fixes:**
    - Move devices to different USB ports/controllers
    - Avoid using the same USB hub for mouse + network adapter
    - Check `dmesg` for USB errors: `dmesg | grep -i usb`

3. **Verify mouse detection:**
    ```bash
    wsconsctl -m
    ```

### Firefox can't see Downloads folder

If Firefox shows a blank path instead of `~/Downloads`:

```bash
# Fix the user-dirs.dirs file
cat > ~/.config/user-dirs.dirs << 'EOF'
XDG_DOWNLOAD_DIR="$HOME/Downloads"
EOF

# Make sure Downloads exists
mkdir -p ~/Downloads

# Restart Firefox
pkill -9 firefox
firefox &
```

### Firefox tabs crash on heavy sites (Grok, X, Canvas apps)

OpenBSD's GPU stack frequently triggers WebRender crashes on WebGL/Canvas-heavy sites. If you see `Gah. Your tab just crashed.`, disable hardware acceleration:

**via `about:config`:**
```
about:config → gfx.webrender.software → true
```

**via Settings:**
Settings → Performance → uncheck "Use recommended performance settings" → uncheck "Use hardware acceleration when available"

> Why: Xenocara's GPU drivers and Firefox's WebRender compositor don't always play well together on OpenBSD. Software rendering is stable and fast enough for desktop work.

### i3 won't start

1. **Check for errors:**

    ```bash
    i3 2>&1 | head -50
    ```

2. **Common fixes:**
    - Graphics driver issue: Check X11 logs at `/var/log/Xorg.0.log`
    - Verify DISPLAY is set: `echo $DISPLAY`

3. **Check dmesg for hardware issues:**
    ```bash
    dmesg | grep -E "error|failed|intel|amd|nvidia"
    ```

### Package missing

If `pkg_add` fails:

1. **Verify installurl is set:**

    ```bash
    cat /etc/installurl
    # Should show a mirror URL
    ```

2. **Auto-select fastest mirror:**

    ```bash
    doas openriot --mirrors
    ```

3. **Or set it manually:**

    ```bash
    echo https://cdn.openbsd.org/pub/OpenBSD | doas tee /etc/installurl
    ```

4. **Try again:**
    ```bash
    pkg_add -v <package-name>
    ```

### System Benchmark

Run `benchmark` to test CPU, memory, and disk performance. Requires `sysbench`:

```bash
doas pkg_add sysbench
benchmark
```

Run via terminal or app launcher:

```bash
openriot --benchmark
```

Results go to `~/.benchmark/<hostname>-YYYYMMDD-N.log` with full system specs.

| Metric | mini (N150, 4c) | t14 (i5-10310U, 8c) | Winner |
|--------|-----------------|---------------------|--------|
| CPU (prime 20k) | 0.0029s | 0.0020s | t14 (31% faster) |
| Memory ops/sec | 4,018,316 | 4,743,669 | t14 (18% faster) |
| Memory MB/s | 3,924 | 4,632 | t14 (18% faster) |
| FileIO (rnd) | 2,206 req/s | 3,561 req/s | t14 (61% faster) |
| dd write | 236 MB/s | 335 MB/s | t14 (42% faster) |
| dd read | 169 MB/s | 2,584 MB/s | t14 (15x faster) |

> "You are absolutely deluded, if not stupid, if you think that a worldwide collection of software engineers who can't write operating systems or applications without security holes, can then turn around and suddenly write virtualization layers without security holes." — Theo de Raadt

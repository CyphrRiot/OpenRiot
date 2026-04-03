<div align="center">

<img src="OpenRiot.png" alt="OpenRiot" width="200"/>

# :: 𝕆𝕡𝕖𝕟ℝ𝕚𝕠𝕥 ::

## One command. Complete OpenBSD desktop. Zero compromises.

![Version](https://img.shields.io/badge/version-0.8-blue?labelColor=0052cc)
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

- **🪟 Sway Tiling** — i3-compatible Wayland compositor with OpenBSD's legendary stability
- **⚡ Robust Binary** — Atomic operations, pledge/unveil security, zero dependency hell
- **🛡️ Privacy** — Zero telemetry, zero tracking, zero data harvesting
- **🎨 Aesthetics** — Carefully crafted dark themes that work at any hour
- **💎 OpenBSD** — The most security-audited OS on the planet

_Built on OpenBSD with Sway, because security and aesthetics shouldn't be mutually exclusive._

## 📚 Navigate This Guide

- [✅ Supported Systems](#supported-systems) — **Read first!**
- [✅ Supported Network Hardware](#supported-network-hardware) — **Read first!**
- [⚠️ UEFI/BIOS Settings](#uefibios-settings) — **Required before install**
- [🔊 Bluetooth](#bluetooth) — **No native support; see workarounds**
- [🚀 Choose Your OpenRiot Experience](#choose-your-openriot-experience)
    - [🔥 Method 1: Install Script](#method-1-install-script)
    - [⚡ Method 2: OpenRiot ISO](#method-2-openriot-iso)
- [⌨️ Master Your OpenRiot Desktop](#master-your-openriot-desktop)
- [🔄 System Management](#system-management)
- [🧰 Advanced Usage](#advanced-usage)
    - [Mullvad VPN on OpenBSD](#mullvad-vpn-on-openbsd)
- [🔧 Troubleshooting](#troubleshooting)
- [📄 License](#license)
- [📋 Progress](#progress) — Project status, plan, and architecture

> "Linux has never been about quality. There are so many parts of the system that are just these cheap little hacks, and it happens to run." -Theo de Raadt

<a id="supported-systems"></a>

## ✅ Supported Systems

**Best and Most Reliable Laptops for OpenBSD 7.8+**

Lenovo ThinkPads remain the strongest and most recommended choice. OpenBSD developers and many users heavily favor them because of their straightforward hardware, excellent keyboards, durability, and long-term support.

### Highly Recommended ThinkPad Series

| Series        | Examples                        | Notes                            |
| ------------- | ------------------------------- | -------------------------------- |
| **X1 Carbon** | Gen 5–7 and later               | Intel WiFi (iwm/iwx) works great |
| **X1 Nano**   | 1st Gen and later               | Lightweight, excellent support   |
| **T series**  | T480, T14, T14s, T420/T430/T61  | Business workhorses              |
| **X series**  | X230, X250, X270, X280, X1 Nano | Compact classics                 |
| **P series**  | P50, P14s Gen 5 Intel           | Workstation power                |

### Other Well-Supported Laptops

| Laptop                                | Notes                        |
| ------------------------------------- | ---------------------------- |
| **Framework Laptop** (11th Gen Intel) | Modular design, good support |
| **Huawei MateBook X** (2017–2020)     | Quiet and reliable           |
| **Dell Latitude/XPS** (older Intel)   | Business class, Intel WiFi   |

### Avoid or Use Caution

- ❌ **NVIDIA discrete GPUs** — Poor support; use Intel iGPU
- ❌ **Killer Wi-Fi** — Replace with Intel card
- ❌ **Realtek/MediaTek Wi-Fi 6/6E/7** — Often unsupported
- ⚠️ **AMD laptops** — Improving but more variable than Intel

### Key Components for OpenBSD

| Component    | Best Choice        | Notes                   |
| ------------ | ------------------ | ----------------------- |
| **Wi-Fi**    | Intel (iwm/iwx)    | Avoid RTL88xxAU, Killer |
| **Graphics** | Intel (inteldrm)   | Best Wayland support    |
| **Audio**    | azalia             | Works on most Intel/AMD |
| **Trackpad** | ThinkPad synaptics | Excellent support       |

For full hardware details, see the [OpenBSD Hardware Compatibility List](https://www.openbsd.org/hardware.html).

<a id="supported-network-hardware"></a>

## ✅ Supported Network Hardware

#### **⚠️ OpenBSD is very selective about WiFi adapters. Only use adapters from this list:**

### Built-in WiFi (PCIe/M.2)

| Driver | Chipsets                                           | Notes                              |
| ------ | -------------------------------------------------- | ---------------------------------- |
| `iwx`  | Intel AX200 / AX201 / AX210 / AX211 (Wi-Fi 6)      | **Best choice for modern laptops** |
| `iwm`  | Intel 7260, 7265, 3160, 3165, 3168, 8260, 8265     | Older Intel cards                  |
| `iwn`  | Intel 4965, 5100, 5300, 5350                       | Legacy Intel                       |
| `athn` | Atheros AR5008 → AR9287 (802.11n)                  | Good range, 2.4GHz only            |
| `bwfm` | Broadcom BCM43xx series                            | Improved WPA in 7.8                |
| `qwx`  | Qualcomm/Atheros 802.11a/ac/ax                     | 802.11n/HT improvements in 7.8     |
| `rtwn` | Realtek RTL8188CE, RTL8188EE, RTL8192CE, RTL8723AE | PCIe cards                         |

### USB WiFi Adapters (Nano/Compact)

| Adapter                       | Chipset      | Driver  | Notes                         |
| ----------------------------- | ------------ | ------- | ----------------------------- |
| **Edimax EW-7811Un** (and v2) | RTL8188CU/EU | `urtwn` | ✅ **Your best bet for USB**  |
| **Asus USB-N10 NANO**         | RTL8188CU    | `urtwn` | Tiny nano adapter             |
| **TP-Link TL-WN725N v2**      | RTL8188EU    | `urtwn` | Very small                    |
| **TP-Link TL-WN722N v1**      | AR9271       | `athn`  | Excellent range (avoid v2/v3) |
| **D-Link DWA-121, DWA-131**   | RTL8188EU    | `urtwn` | Various revisions work        |
| **Alfa AWUS036NHA**           | AR9271       | `athn`  | High gain, great range        |

**All USB adapters are 2.4GHz 802.11n only (~50-100 Mbps real-world).**

- [Amazon link to Edimax](https://www.amazon.com/dp/B08F2ZNC6J)

### NOT Supported (Do Not Buy)

- ❌ Intel BE201 (Wi-Fi 7)
- ❌ Realtek RTL8811AU / RTL8812AU / RTL8812AU
- ❌ MediaTek WiFi chips
- ❌ Most Qualcomm Wi-Fi 6E/7 chips

For full compatibility, see [iwx(4)](https://man.openbsd.org/iwx.4), [urtwn(4)](https://man.openbsd.org/urtwn.4), and [athn(4)](https://man.openbsd.org/athn.4) man pages.

<a id="uefibios-settings"></a>

> "My favorite part of the 'many eyes' argument is how few bugs were found by the two eyes of Eric (the originator of the statement). All the many eyes are apparently attached to a lot of hands that type lots of words about many eyes, and never actually audit code." — Theo de Raadt

## ⚠️ UEFI/BIOS Settings

There are several **UEFI/BIOS settings** you should verify or adjust before installing **OpenBSD**. These ensure the installer boots reliably, hardware is detected correctly, and post-install features like suspend/resume and WiFi work as expected.

OpenBSD boots in **pure UEFI mode** and does **not** support Secure Boot. Most modern hardware is compatible once these settings are correct.

### How to Enter BIOS

Power on your machine and tap the appropriate key during boot (commonly **F1**, **F2**, **Del**, or **Esc** — consult your manufacturer's docs). You can also tap **F12** (or equivalent) for a one-time boot menu to select your install USB.

### Recommended UEFI/BIOS Settings

The following settings are broadly applicable across modern laptops and desktops. Not all will be present on every system — locate the equivalents for your hardware.

| Category           | Setting                               | Recommended Value                  | Why It Matters for OpenBSD                                                                                                       |
| ------------------ | ------------------------------------- | ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| **Boot**           | Secure Boot                           | **Disabled**                       | **Critical.** OpenBSD's bootloader is not signed with Microsoft's keys. Enabled Secure Boot prevents the installer from booting. |
| **Boot**           | UEFI/Legacy Boot                      | **UEFI Only**                      | OpenBSD works best in pure UEFI mode. Avoid "Legacy" or "Both."                                                                  |
| **Boot**           | CSM (Compatibility Support Module)    | **Disabled**                       | No longer needed on modern UEFI systems; can cause boot issues.                                                                  |
| **Boot**           | Fast Boot / Quick Boot                | **Disabled**                       | Can skip necessary hardware initialization, causing boot hangs or installer failures.                                            |
| **Power**          | Sleep State                           | **Linux** or **S3** (if available) | Improves suspend/resume compatibility. "Windows" or "Modern Standby" mode can break `apmd`/`zzz`.                                |
| **Power**          | Modern Standby                        | **Disabled** or **Legacy S3**      | Set to legacy S3-style sleep for reliable suspend/resume on OpenBSD.                                                             |
| **Security**       | TPM / PTT                             | Enabled (or "Firmware TPM")        | Helps with suspend/resume and disk encryption (softraid). Safe to leave on.                                                      |
| **Security**       | Intel SGX                             | **Disabled** (if present)          | Reduces attack surface; some systems behave more reliably with this off.                                                         |
| **USB**            | USB UEFI BIOS Support / Always On USB | Default or As needed               | Usually fine to leave as default unless you encounter boot issues.                                                               |
| **Thunderbolt**    | Thunderbolt BIOS Assist Mode          | **Enabled** (if present)           | Improves USB-C/Thunderbolt stability and performance.                                                                            |
| **Thunderbolt**    | Wake by Thunderbolt                   | **Disabled** (if present)          | Prevents unwanted wake events.                                                                                                   |
| **Input**          | TrackPoint / TouchPad                 | **Enabled** (both)                 | OpenBSD has excellent TrackPoint support. Touchpad is optional based on preference.                                              |
| **Virtualization** | VT-x / VT-d                           | **Enabled** (if present)           | Required for `vmm(4)` if you plan to run virtual machines.                                                                       |

### Pre-Installation Checklist

1. **Update firmware** — Run the latest BIOS/UEFI from your manufacturer's support site before installing OpenBSD.
2. **Enter BIOS** (F1 or equivalent) and apply the settings above.
3. **Save & Exit**.
4. **Boot your install media** (use one-time boot menu if needed).
5. If the installer doesn't appear, double-check **Secure Boot is off** and boot mode is **UEFI Only**.
6. After a successful install and first boot:
    - Run `fw_update -a` immediately.
    - Enable `apmd` with `-A` or `-L` for power management.
    - Test suspend with `zzz` (S3 sleep) or `ZZZ` (hibernation).

### Why This Matters for OpenRiot

These BIOS changes are a **one-time setup**. Disabling Secure Boot and setting UEFI-only mode give OpenBSD full control of your hardware — exactly what a sovereignty-focused OS needs. Once these are in place, OpenBSD handles the rest with its legendary stability and security auditing.

If you encounter boot issues (e.g., "no bootable device" or hangs at logo), the most common culprits are **Secure Boot still enabled** or **CSM interference**. Re-check those first.

<a id="bluetooth"></a>

## 🔊 Bluetooth

#### **⚠️ OpenBSD has NO native Bluetooth support.** The Bluetooth stack was removed years ago and has not been reinstated.

> "You are absolutely deluded, if not stupid, if you think that a worldwide collection of software engineers who can't write operating systems or applications without security holes, can then turn around and suddenly write virtualization layers without security holes." — Theo de Raadt

### What Doesn't Work

- ❌ Built-in laptop Bluetooth
- ❌ USB Bluetooth dongles
- ❌ Pairing Bluetooth headphones, keyboards, mice

### Bluetooth Audio Workaround

Use a dedicated USB Bluetooth audio transceiver that handles Bluetooth internally and appears as standard USB audio:

| Transceiver        | Type  | Notes               |
| ------------------ | ----- | ------------------- |
| **Creative BT-W3** | USB-C | ✅ Most recommended |
| **Creative BT-W2** | USB-A | Good alternative    |
| **UGREEN BT501**   | USB-C | Budget option       |

Once paired (via button on dongle), switch audio output using `sndioctl`.

### Bluetooth Mouse/Keyboard Workaround

**Logitech MX Anywhere 3S:**

- Bluetooth mode → ❌ Will not work
- Logi Bolt USB receiver → ✅ Basic support (cursor, clicks, scroll)
- Wired USB-C mode → ✅ Full support

**Logi Bolt Receiver:** ~$15 on Amazon (USB-A or USB-C)

### Recommended Input Setup

| Device              | Solution                          | What Works                   |
| ------------------- | --------------------------------- | ---------------------------- |
| Mouse               | Logi Bolt receiver or wired       | Cursor, clicks, basic scroll |
| Keyboard            | 2.4 GHz USB dongle or wired       | Basic typing                 |
| Advanced mouse feat | Not possible (gestures, MagSpeed) | N/A                          |

**Most reliable:** Wired USB keyboard + mouse for critical work.

## 🚀 Choose Your OpenRiot Experience

<a id="method-1-install-script"></a>

### 🔥 Method 1: Install Script

#### You already have OpenBSD installed

**Transform your current OpenBSD system into OpenRiot**

```bash
curl -fsSL https://openriot.org/setup.sh | sh
```

**Perfect for:**

- 🏠 **System preservation** — Keep your data and configs intact
- 🔧 **OpenBSD variants** — Any OpenBSD 7.x installation
- 🎨 **Desktop upgrade** — Transform just your desktop environment
- ⚡ **Quick wins** — Get OpenRiot's features without starting over

**What you get:**

- OpenRiot Sway desktop environment and apps
- CypherRiot themes and customizations
- Waybar with custom modules
- Fish shell with git prompts

<a id="method-2-openriot-iso"></a>

### ⚡ Method 2: OpenRiot ISO

#### You do NOT have OpenBSD installed

1. **📥 Download ISO**

    ```bash
    curl -fsSL https://github.com/CyphrRiot/OpenRiot/releases/latest/download/openriot.iso -o openriot.iso
    ```

    Or download directly at [OpenRiot ISO](https://OpenRiot.org/isos/openriot.iso)

2. **🔧 Create Bootable USB**

    **Option A — dd (Linux/macOS):**

    ```bash
    sudo dd if=openriot.iso of=/dev/sdX bs=1M status=progress
    ```

    **Option B — Ventoy (recommended for multi-boot):**
    - Install Ventoy on USB: https://www.ventoy.net/
    - Copy `.iso` to Ventoy partition
    - Boot and select from Ventoy menu

3. **🚀 Install OpenBSD**
    - Boot from USB
    - Choose `(I)nstall` — fully automated
    - Enter disk encryption passphrase when prompted
    - Create your user account
    - After install completes, log in

4. **✅ First Boot**
   The installer runs automatically on first login:

    ```bash
    # Just log in — everything installs automatically
    ```

    Or run manually:

    ```bash
    curl -fsSL https://openriot.org/setup.sh | sh
    ```

5. **🔄 Reboot**
    - Log out and back in
    - Type `sway` to start your desktop

<a id="master-your-openriot-desktop"></a>

## ⌨️ Master Your OpenRiot Desktop

OpenRiot uses **Sway** (i3-compatible Wayland compositor). Keybindings mirror ArchRiot:

| Key                   | Action                   |
| --------------------- | ------------------------ |
| `Super + Return`      | Terminal (foot)          |
| `Super + D`           | App Launcher (fuzzel)    |
| `Super + F`           | File Manager (Thunar)    |
| `Super + B`           | Browser                  |
| `Super + L`           | Lock Screen              |
| `Super + 1-6`         | Switch Workspace         |
| `Super + Shift + 1-6` | Move Window to Workspace |
| `Print`               | Screenshot (region)      |
| `Super + Shift + H`   | Keybindings Help         |

<a id="system-management"></a>

## 🔄 System Management

OpenBSD update commands:

```bash
# Update packages
pkg_add -u

# Update system
syspatch -a && sysupgrade -n && syspatch -a && sysupgrade

# Rebuild packages after major version upgrade
pkg_add -u
```

### Finding Packages

Search for OpenBSD packages at **[openbsd.app](https://openbsd.app)** or via command line:

```bash
# Search for packages
pkg_info -Q package-name

# List all available packages
pkg_add -l | less
```

<a id="advanced-usage"></a>

## 🧰 Advanced Usage

### Mullvad VPN on OpenBSD

The simplest way to run Mullvad VPN on OpenBSD using WireGuard:

#### 1. Install WireGuard Tools

```bash
doas pkg_add wireguard-tools
```

#### 2. Generate Mullvad Config

1. Log in at: https://mullvad.net/account/wireguard-config
2. Select **Linux** as platform
3. Generate a new key pair
4. Choose a single location (avoid multihop for first test)
5. Download the `.conf` file

#### 3. Place the Config

```bash
doas mkdir -p /etc/wireguard
doas mv ~/Downloads/mullvad-*.conf /etc/wireguard/wg0.conf
```

#### 4. Connect

```bash
doas wg-quick up wg0
```

#### 5. Verify

- Visit: https://mullvad.net/check (should show green "You are connected")
- Or run: `curl -4 ifconfig.me` (should show Mullvad IP)

#### Disconnect

```bash
doas wg-quick down wg0
```

#### Auto-start at Boot (Optional)

```bash
doas rcctl enable wg
doas rcctl start wg
```

#### DNS Leaks

wg-quick handles DNS automatically. If you see leaks, manually set in `/etc/resolv.conf`:

```
nameserver 193.138.218.74
lookup file bind
```

### Environment Variables

OpenRiot sets sensible defaults in `~/.config/environment.d/`:

```sh
XDG_CURRENT_DESKTOP=sway
XDG_SESSION_TYPE=wayland
XCURSOR_THEME=Bibata-Modern-Ice
```

### Keybindings Customization

Edit `~/.config/sway/keybindings.conf` and reload:

```bash
killall sway && sway
```

### Waybar Modules

Waybar modules are in `~/.config/waybar/`. See ArchRiot config for reference.

<a id="troubleshooting"></a>

## 🔧 Troubleshooting

### WiFi not working

OpenBSD uses `iwx` for Intel WiFi 6 (AX211). Run:

```bash
fw_update -a -v
reboot
```

### Sway won't start

Check logs:

```bash
sway -d 2>&1 | less
```

> "Linux people do what they do because they hate Microsoft." — Theo de Raadt

### Package missing

Search OpenBSD packages:

```bash
pkg_info -Q package-name
```

<a id="license"></a>

## 📄 License

MIT License — see [LICENSE](./LICENSE)

---

<a id="progress"></a>

> "It's terrible. Everyone is using it, and they don't realize how bad it is. And the Linux people will just stick with it and add to it rather than stepping back and saying, 'This is garbage and we should fix it.'" — Theo de Raadt

## 📋 Progress

See [Progress.md](./Progress.md) for full project status, architecture, build plan, and TODO list.

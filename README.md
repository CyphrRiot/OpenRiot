<div align="center">

# :: 𝕆𝕡𝕖𝕟ℝ𝕚𝕠𝕥 ::

#### **⚠️ WARNING: OpenRiot is IN PROGRESS and will not (yet) install! ⚠️**

### **One Command. Complete Environment. Zero Compromises.**

![Version](https://img.shields.io/badge/version-1.0.0-blue?labelColor=0052cc)
![License](https://img.shields.io/github/license/CyphrRiot/OpenRiot?color=4338ca&labelColor=3730a3)
![Platform](https://img.shields.io/badge/platform-OpenBSD-4338ca?logo=openbsd&logoColor=white&labelColor=3730a3)
![OpenBSD](https://img.shields.io/badge/OpenBSD_7.9-1e1b4b?logo=openbsd&logoColor=a855f7&labelColor=0f172a)
![Sway](https://img.shields.io/badge/Sway-Wayland-312e81?logo=wayland&logoColor=a855f7&labelColor=1e1b4b)
![OpenBSD-current](https://img.shields.io/badge/OpenBSD_-current_7.9-4338ca?labelColor=3730a3)

![Last Commit](https://img.shields.io/github/last-commit/CyphrRiot/OpenRiot?color=5b21b6&labelColor=4c1d95)
![Code Size](https://img.shields.io/github/languages/code-size/CyphrRiot/OpenRiot?color=4338ca&labelColor=3730a3)
![Code](https://img.shields.io/badge/human-coded-blue?logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNmZmZmZmYiIHN0cm9rZS13aWR0aD0iMiIgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIiBzdHJva2UtbGluZWpvaW49InJvdW5kIiBjbGFzcz0ibHVjaWRlIGx1Y2lkZS1wZXJzb24tc3RhbmRpbmctaWNvbiBsdWNpZGUtcGVyc29uLXN0YW5kaW5nIj48Y2lyY2xlIGN4PSIxMiIgY3k9IjUiIHI9IjEiLz48cGF0aCBkPSJtOSAyMCAzLTYgMyA2Ii8+PHBhdGggZD0ibTYgOCA2IDIgNi0yIi8+PHBhdGggZD0iTTEyIDEwdjQiLz48L3N2Zz4=&logoColor=a855f7&labelColor=1e1b4b)

![Language](https://img.shields.io/badge/language-Go-4338ca?logo=go&logoColor=c7d2fe&labelColor=3730a3)
![Language](https://img.shields.io/badge/language-YAML-5b21b6?logo=yaml&logoColor=e0e7ff&labelColor=4c1d95)
![Language](https://img.shields.io/badge/Python-312e81?label=language&logo=python&logoColor=c7d2fe&labelColor=1e1b4b&color=312e81)

</div>

## **OpenRiot: The OpenBSD System You've Always Wanted**

OpenRiot is the answer to every time you've thought "Why can't an OpenBSD installation just work correctly from the start?" Built on the same principles as ArchRiot — perfect the details once, get it right every time.

**Curated to be correct:**

- **🪟 Sway Tiling** — i3-compatible Wayland compositor with OpenBSD's legendary stability
- **⚡ Robust Binary** — Atomic operations, pledge/unveil security, zero dependency hell
- **🛡️ Privacy** — Zero telemetry, zero tracking, zero data harvesting
- **🎨 Aesthetics** — Carefully crafted dark themes that work at any hour
- **💎 OpenBSD** — The most security-audited OS on the planet

_Built on OpenBSD with Sway, because security and aesthetics shouldn't be mutually exclusive._

## 📚 Navigate This Guide

- [🚀 Choose Your OpenRiot Experience](#choose-your-openriot-experience)
    - [🔥 Method 1: Install Script](#method-1-install-script)
    - [⚡ Method 2: OpenRiot ISO](#method-2-openriot-iso)
- [⌨️ Master Your OpenRiot Desktop](#master-your-openriot-desktop)
- [🔄 System Management](#system-management)
- [🧰 Advanced Usage](#advanced-usage)
- [🔧 Troubleshooting](#troubleshooting)
- [📄 License](#license)

<a id="choose-your-openriot-experience"></a>

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

⚠️ **Warning: ISO will replace a drive with OpenBSD + OpenRiot. ⚠️**

1. **📥 Download OpenRiot ISO**
    - **[OpenRiot ISO](https://github.com/CyphrRiot/OpenRiot/releases)**
    - Verify the SHA256 checksum before flashing

2. **🔧 Create bootable USB**

    ```bash
    dd if=openriot-*.iso of=/dev/sdX bs=1M status=progress
    ```

3. **🚀 Boot and install**
    - Boot from USB
    - Choose `(I)nstall`
    - Answer the prompts (autoinstall answers are pre-filled)
    - After base install, log in and run:

    ```bash
    curl -fsSL https://openriot.org/setup.sh | sh
    ```

<a id="master-your-openriot-desktop"></a>

## ⌨️ Master Your OpenRiot Desktop

OpenRiot uses **Sway** (i3-compatible Wayland compositor). Keybindings mirror ArchRiot:

| Key                   | Action                   |
| --------------------- | ------------------------ |
| `Super + Return`      | Terminal (foot)          |
| `Super + D`           | App Launcher (wofi)      |
| `Super + F`           | File Manager (Thunar)    |
| `Super + B`           | Browser                  |
| `Super + L`           | Lock Screen              |
| `Super + 1-6`         | Switch Workspace         |
| `Super + Shift + 1-6` | Move Window to Workspace |
| `Super + Shift + L`   | Lock Screen              |
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

<a id="advanced-usage"></a>

## 🧰 Advanced Usage

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

### Package missing

Search OpenBSD packages:

```bash
pkg_info -Q package-name
```

<a id="license"></a>

## 📄 License

MIT License — see [LICENSE](./LICENSE)

---

**OpenRiot 🐡** — An opinionated OpenBSD desktop system  
Created by [CyphrRiot](https://github.com/CyphrRiot)

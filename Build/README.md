# OpenRiot Image Builder

Builds `openriot.img` - a custom OpenBSD 7.9 installer with OpenRiot configs baked in.

## Prerequisites

- OpenBSD 7.9 (required for vnconfig/mount)
- Root access (`doas`)
- Base image linked: `Images/install79.img`

## Snapshot Build Steps (7.9 pre-release)

While 7.9 is in development/snapshots, package versions change frequently. Run these before building:

```bash
# 1. Update system packages to get current versions
doas pkg_add -D snapshot -u

# 2. Sync packages.yaml with installed versions
openriot --sync-packages

# 3. Build the image
doas ./make-img.sh
```

This ensures the bundled packages match what actually exists on the CDN.

**After 7.9 stable release**: Skip steps 1-2, packages are frozen by version number.

## Quick Start

```bash
# 1. Link base image (if not present)
mkdir -p Images && ln -sf ~/Code/Images/install79.img Images/install79.img

# 2. Build
doas ./make-img.sh

# 3. Flash to USB
dd if=Images/openriot.img of=/dev/rsd2c bs=1M
```

## What Gets Built

### openriot.img
- Base OpenBSD installer with `openriot.tgz` injected at `/`
- During install, `install.site` runs to configure the system
- User is prompted for: disk, hostname, root/user passwords, timezone

### openriot.tgz (in Build/work/)
Contains:
- `/etc/openriot/openriot` — OpenRiot binary
- `/etc/openriot/packages.yaml` — Package definitions
- `/usr/local/share/openriot/` — All configs
- `/root/.local/share/fonts/` — FiraCode Nerd Font
- `/install.site` — Post-install script
- `/install.conf` — Installer answers

## Commands

```bash
./make-img.sh          # Full build (default)
./make-img.sh site     # Create openriot.tgz only
./make-img.sh clean    # Clean Build/work/
./make-img.sh help     # Show help
```

## How It Works

1. **Create openriot.tgz** — bundles binary, configs, install.site
2. **Mount base image** — via vnconfig
3. **Inject openriot.tgz** — into `/` of image
4. **Output to Images/openriot.img**

## Flow

```
install79.img + configs → openriot.tgz → inject → openriot.img
```

## Output

- `Images/openriot.img` — Ready to flash to USB
- `Build/work/openriot.tgz` — Also injected into image

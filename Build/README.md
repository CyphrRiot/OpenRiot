# OpenRiot Image Builder

Builds `openriot.img` - a custom OpenBSD installer with OpenRiot configs and packages baked in.

## Prerequisites

- OpenBSD (current/snapshots)
- Root access (`doas`)
- Base image linked: `Images/install79.img`

## Quick Start

```bash
# 1. Link base image (if not present)
mkdir -p Images && ln -sf ~/Code/Images/install79.img Images/install79.img

# 2. Build
make img

# 3. Flash to USB (when prompted, select drive)
```

## What Gets Built

### openriot.img
- Base OpenBSD installer expanded to 2GB
- `openriot.tgz` injected at `/`
- `packages/` directory with all .tgz files
- During install, `install.site` runs to configure the system
- User prompted for: disk, hostname, root/user passwords, timezone
- **Shrinks to fit actual content** after building

### Image Size
- Starts at 2GB fixed
- Shrinks after content injection based on filesystem usage
- Final size ~1.8-2GB depending on packages

## How It Works

```
1. Clean old artifacts (openriot.img, openriot.tgz)
2. Create openriot.tgz with repo + configs
3. Download packages (skipping exceptions.yaml)
4. Expand base image to 2GB
5. Inject openriot.tgz + packages
6. Shrink image to fit content
7. Prompt to burn to removable drive
```

## Package Management

### exceptions.yaml
Packages to exclude from the image:
```yaml
exclude:
  - abiword    # Not needed
  - btop       # User install only
  - go         # Not needed on install
```

Old packages are auto-cleaned before downloading fresh ones.

### packages.yaml sync
For snapshots, sync packages before building:
```bash
openriot --sync-packages
make img
```

## Drive Detection

The builder detects drives and protects system disks:

```
[ROOT] sd0 -   238 GB [OpenBSD]           # Protected: softraid parent
[ROOT] sd1 -   238 GB [OpenBSD Encrypted] # Protected: boot drive (root on sd1a)
[WARN] sd3 -    29 GB [Removable USB]     # Burnable: removable drive

[DONE] Available for burn: sd3
```

**Protection logic:**
- `[ROOT]` = Boot drive (parsed from `root on sd1a` in dmesg) + softraid parents (drives with RAID partitions)
- `[WARN]` = Removable drives (from `removable` in dmesg)
- `[INFO]` = Internal empty drives

Only non-protected drives are offered for burning.

## Commands

```bash
make img      # Full build (site + image)
make site     # Create openriot.tgz only
make clean    # Clean Build/work/
```

## Output

- `Images/openriot.img` — Ready to flash to USB
- `Images/openriot.sha256` — SHA256 checksum
- `Build/work/openriot.tgz` — Also injected into image
- `Build/work/packages/` — Downloaded .tgz files

## Snapshot Updates

While using OpenBSD snapshots, package versions change frequently:

```bash
# Update system packages to get current versions
doas pkg_add -D snapshot -u

# Sync packages.yaml with installed versions
openriot --sync-packages

# Build the image
make img
```

**After stable release**: Skip steps 1-2, packages are frozen by version number.

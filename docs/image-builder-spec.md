# Functional Specification: `openriot --make-image`

## Overview

Convert the shell script `Build/make-img.sh` to a native Go command integrated into the `openriot` binary. The command builds a bootable OpenRiot installer image from an OpenBSD base image.

## Requirements

### Platform
- **Must run on OpenBSD** (current or snapshots)
- **Must run as root** (required for `vnconfig`, mounting, burning)
- Binary must exist at `install/openriot` (from prior `make` build)

### Flags and Arguments

```
openriot --make-image [mode] [flags]

Modes:
  (none)        Full build: create site tarball + image (default)
  site          Create openriot.tgz tarball only
  clean         Clean build artifacts (work/, cached repos)
  help          Show help

Flags:
  --base-img PATH    Base OpenBSD image (default: ./Build/Images/install79.img)
  --output-img PATH  Output image path (default: ./Build/Images/openriot.img)
  --work-dir PATH    Working directory (default: ./Build/work)
  --version X.Y      OpenBSD version to target (default: 79)
  --no-burn          Skip interactive burn prompt
```

### Environment Variables (fallback)

| Variable    | Purpose                          | Default |
|-------------|----------------------------------|---------|
| `BASE_IMG`  | Path to base OpenBSD image      | See flags |
| `OUTPUT_IMG`| Path for output image           | See flags |
| `WORK_DIR`  | Working directory               | See flags |

## Functional Modules

### 1. Prerequisites Check

**Location**: New file `source/imaging/prereqs.go`

- Verify running on OpenBSD (`uname -s == OpenBSD`)
- Verify root user (`id -u == 0`)
- Verify base image exists at `--base-img` or `BASE_IMG`
- Verify binary exists at `install/openriot` (from `make`)
- Verify `--packages` flag works

### 2. Package Download

**Location**: New file `source/imaging/download.go`

- Run `openriot --packages` to get package list
- Read `Build/exceptions.yaml` for excluded packages
- Download from `https://cdn.openbsd.org/pub/OpenBSD/snapshots/packages/amd64/{pkg}.tgz`
- Clean stale packages (not in current list, or in exceptions)
- Progress indicator: `Downloading package N/N: pkgname`
- Retry logic: 3 attempts per package

### 3. Site Tarball Creation

**Location**: New file `source/imaging/site.go`

Create `openriot.tgz` containing:

```
site/
├── etc/
│   └── motd                    # From install/motd
├── openriot/
│   └── repo/                   # OpenRiot Git repo (cloned/fetched)
├── install.site                # Post-install script (inline)
└── install.conf                # Autoinstall answers (inline)
```

#### install.site (inline, embedded in Go)

Runs during OpenBSD install with these steps:
1. Extract `openriot.tgz`
2. Configure `doas` (permit wheel group)
3. Configure `installurl`
4. Install packages from local path via `pkg_add`
5. Copy repo to `~/.local/share/openriot`
6. Add welcome message to `/etc/skel/.profile`

#### install.conf (inline, embedded in OpenRiot repo)

Autoinstall answers for interactive prompts:
- Disk selection: `ask` → `edit`
- Hostname, passwords: `ask`
- Timezone: `US/Pacific`
- Sets location: `disk` (already mounted)
- Install `openriot.tgz`: `yes`

### 4. Image Building

**Location**: New file `source/imaging/build.go`

#### Expand Phase
- Create 2GB fixed-size image from base
- Configure `vnd0` device
- Modify disklabel to fill partition
- Run `growfs` on expanded partition

#### Injection Phase
- Mount filesystem at `/mnt`
- Copy `openriot.tgz` to `/mnt/openriot.tgz`
- Copy packages to `/mnt/openriot/packages/snapshots/amd64/`
- Unmount

#### Shrink Phase
- Calculate used space: `df -k /dev/vnd0a`
- Add 10% buffer + 32MB for filesystem metadata
- Align to 4MB boundary
- Minimum 1GB
- Truncate image file

### 5. Drive Detection & Burning

**Location**: New file `source/imaging/burn.go`

#### Detection Logic

| Drive Status | Label      | Color   | Suffix                | Burn Eligible |
|--------------|------------|---------|-----------------------|---------------|
| Root drive   | `[ROOT]`   | RED     | [OpenBSD Encrypted]   | No            |
| Softraid     | `[ROOT]`   | RED     | [OpenBSD]             | No            |
| Removable    | `[WARN]`   | YELLOW  | [Removable USB]       | Yes           |
| Internal     | `[INFO]`   | CYAN    | (none)                | Yes           |

**Algorithm**:
1. Parse `dmesg` for `root on sdXa` → extract root drive
2. Parse `dmesg` for `removable` → collect removable drives
3. Parse `sysctl -n hw.disknames` for all disks
4. For each disk, run `disklabel`:
   - Check for RAID partitions → add to protected
   - Calculate size (bytes/sector × total sectors)
5. Build drive list with status

#### Burn Prompt
```
[ROOT] sd0 -  512 GB [OpenBSD Encrypted]
[INFO] sd1 -  128 GB
[WARN] sd2 -   64 GB [Removable USB]

[DONE] Available for burn: sd1, sd2

[WARN] THIS WILL ERASE ALL DATA ON THE SELECTED DRIVE.

[ASK ] Which drive to burn? (sd1, sd2 or press Enter to skip)
> sd2

[WARN] You will be erasing sd2 (64 GB).
[ASK ] Are you sure? [y/N]
> y
Burning to /dev/rsd2c...
[DONE] Burn complete!
```

#### Write Command
```go
cat imagePath | pv -pterb | doas dd of=/dev/r{DRIVE}c bs=1M
```

### 6. Cleanup

**Location**: New file `source/imaging/cleanup.go`

- Unmount `/mnt` if mounted
- Release `vnd0` device
- Remove `WORK_DIR` contents (packages, site, repo cache)

## Output Artifacts

| File                     | Location                    | Description                    |
|--------------------------|-----------------------------|--------------------------------|
| `openriot.tgz`           | `WORK_DIR/openriot.tgz`    | Site tarball for installer     |
| `openriot.img`           | `OUTPUT_IMG`               | Bootable installer image       |
| `openriot.sha256`        | Same dir as img            | SHA256 checksum                |
| Package cache            | `WORK_DIR/packages/`       | Downloaded .tgz files          |
| Repo cache               | `Build/repo-cache/`        | Cached Git repo for faster builds |

## Error Handling

| Error                           | Action                                      |
|---------------------------------|---------------------------------------------|
| Not running on OpenBSD          | Exit with error, show required OS          |
| Not root                        | Exit with error, require sudo/doas         |
| Base image missing              | Exit with error, show expected path        |
| Binary missing                  | Exit with error, run `make` first          |
| Package download fails          | Warn, continue with remaining              |
| Package list empty              | Exit with error                             |
| Disklabel/growfs fails          | Exit with error                             |
| Mount fails                     | Exit with error                             |
| Burn fails                      | Show error, leave image file                |

## Integration with CLI

Add to `source/main.go` command map:

```go
"--make-image": func() {
    imaging.RunMakeImage(os.Args[2:])
},
```

Parse flags using `flag` package or manual parsing (matching existing `--packages` pattern).

## Dependencies

### External Commands (via exec)
- `vnconfig` - Configure vnd device
- `disklabel` - Read/modify disk label
- `growfs` - Expand filesystem
- `fsck` - Filesystem check
- `mount` / `umount` - Mount operations
- `sha256` - Checksum generation
- `dmesg` - Drive detection
- `sysctl` - System info
- `doas` - Privilege escalation for burn

### Go Standard Library
- `os` / `os/exec` - File operations, subprocesses
- `io` - Copy operations
- `crypto/sha256` - (optional, prefer `sha256` binary)

## Testing

- Unit test each module with mocked exec calls
- Integration test on OpenBSD VM (QEMU)
- Test drive detection with various configurations
- Test burn on small test image

## Future Considerations

- Support custom base image versions
- Parallel package downloads
- Resume interrupted downloads
- Image signing/verification
# OpenRiot Image Builder - Functional Specification

## Command

```bash
openriot --make-image [site|clean|help]
make image          # Wrapper that runs as root
```

## Purpose

Build a bootable OpenBSD installer image with OpenRiot packages pre-bundled. The image contains:
- OpenBSD base installer (`install79.img`)
- `site79.tgz` (custom set with 81 packages + `install.site` post-install script)
- Modified `index.txt` for installer set discovery

Target size: **< 2.0GB**. Actual: **~1932MB**.

## Architecture

### Components

| Component | Location | Description |
|-----------|----------|-------------|
| `openriot --make-image` | `source/imaging/*.go` | Main image builder logic |
| Base image | `Build/Images/install79.img` | OpenBSD 7.9 installer (~801MB) |
| Output image | `Build/Images/openriot.img` | Final bootable image (~1932MB) |
| Work dir | `Build/work/` | Temporary build artifacts |
| Site tarball | `Build/work/site79.tgz` | Custom set (~781MB) |

### Directory Structure

```
Build/
├── Images/
│   ├── install79.img          # Base OpenBSD installer (source, read-only)
│   └── openriot.img           # Output image (bootable)
├── work/
│   ├── site79.tgz             # Custom installer set
│   └── site-XXXXXX/           # Temp staging dir (auto-removed)
└── packages/snapshots/amd64/  # Downloaded .tgz files (~781MB)
```

### Source Files

| File | Purpose |
|------|---------|
| `source/imaging/runner.go` | Entry point, mode dispatch, config loading |
| `source/imaging/prereqs.go` | Host checks, base image validation/download |
| `source/imaging/download.go` | Package downloading from CDN |
| `source/imaging/site.go` | `site79.tgz` creation, `install.site` generation |
| `source/imaging/build.go` | Image expansion, content injection, checksum |
| `source/imaging/burn.go` | USB drive detection and burning |
| `source/fsutil/copy.go` | File copy utility (`io.Copy` for large files) |

## Build Flow

### 1. Prerequisites Check (`CheckPrereqs`)

- Must be OpenBSD
- Must be root (for `vnconfig`, `mount`, `disklabel`, `growfs`)
- Base image exists at `Build/Images/install79.img` (auto-downloaded if missing)
- Binary exists at `install/openriot`

### 2. Download Packages (`DownloadPackages`)

- Reads package list from `install/packages.yaml`
- Downloads to `Build/work/packages/snapshots/amd64/`
- 2-minute HTTP timeout per file
- 81 packages, ~781MB total

### 3. Create Site Tarball (`CreateSite`)

- Uses `os.MkdirTemp` for clean staging directory
- Copies packages into `openriot/packages/snapshots/amd64/`
- Copies `install/motd` to `etc/motd`
- Generates `install.site` post-install script:
  - Configures `doas.conf`, `installurl`, `/etc/shells`
  - Installs all packages from local path via `pkg_add`
  - Adds users to `wheel`, sets fish shell
  - Adds welcome message to `.profile`
- Creates `site79.tgz` via `tar czf`

### 4. Build Image (`BuildImage`)

1. **Copy base:** `fsutil.CopyFile(install79.img, openriot.img)`
2. **Expand:**
   - `truncate -s` to `base + tarball + 350MB`
   - `vnconfig vnd0 openriot.img`
   - Parse `disklabel vnd0` for partition info
   - `disklabel -R` to update `a:` and `c:` sizes
   - `growfs -y /dev/vnd0a`
3. **Inject:**
   - `vnconfig vnd0 openriot.img`
   - `mount /dev/vnd0a /mnt`
   - Create `7.9/amd64/` directory
   - Copy `site79.tgz` into sets directory
   - Update `index.txt` with `ls -lT` output
   - `sync` before unmount
   - `umount /mnt`, `vnconfig -u vnd0`
4. **Checksum:** Generate SHA256

## Known Issues

### Critical: Unvalidated End-to-End

**The image has been built but NEVER successfully booted to completion.**

| Issue | Status | Detail |
|-------|--------|--------|
| Image builds without errors | ✅ Working | `make image` completes |
| Image size under 2.0GB | ✅ Working | ~1932MB |
| `site79.tgz` injects into `7.9/amd64/` | ✅ Working | Verified in build logs |
| `index.txt` contains `site79.tgz` entry | ✅ Working | Uses real `ls -lT` output |
| **Boot test** | ❌ **NOT DONE** | Image has never been booted |
| **`site79.tgz` discovered by installer** | ❌ **NOT VALIDATED** | May require manual `*` selection |
| **`install.site` execution** | ❌ **NOT VALIDATED** | Script generated, never observed running |
| **Package installation** | ❌ **NOT VALIDATED** | `pkg_add` from local path untested |
| **First-login welcome** | ❌ **NOT VALIDATED** | `.profile` append untested |

### Installer UX (Interactive Install)

The OpenBSD interactive installer has **intrinsic behaviors** that cannot be changed without modifying `bsd.rd` or `install.sub`:

1. **Sets location defaults to HTTP** when network interface detected
   - User must manually select `disk`
   - Cannot be overridden from media filesystem

2. **`site79.tgz` is deselected by default**
   - Custom `site*.tgz` sets are not in `DEFAULTSETS`
   - User must add with `*`
   - Cannot be pre-selected without modifying installer

3. **SHA256.sig verification fails**
   - `site79.tgz` is not signed by OpenBSD release key
   - Installer prompts "Continue without verification? [no]"
   - User must answer `yes`
   - Cannot be signed without private release key

### Historical Bugs (Fixed)

| Bug | Fix | Commit |
|-----|-----|--------|
| Permission collisions from root-owned `site/` dir | `os.MkdirTemp` + `defer os.RemoveAll` | Earlier |
| `args` slice mutation in `RunMakeImage` | Build `filtered` slice instead of mutating | Earlier |
| Indefinite HTTP hangs | `http.Client{Timeout: 2*time.Minute}` | Earlier |
| Leaked `vnd0` devices on error paths | `defer` cleanup blocks | Earlier |
| Wrong fallback image path | `Build/Images/install79.img` | Earlier |
| **I/O error at 154MB during install** | Removed `fsck -y` from `injectContent()` | Current session |
| Copy verification | Size check + `sync` after injection | Current session |
| Memory pressure on 781MB tarball | `io.Copy` instead of `os.ReadFile`/`WriteFile` | Current session |

### Abandoned: Autoinstall via bsd.rd

Attempted to patch `bsd.rd` ramdisk with `auto_install.conf` using `rdsetroot`.
**Reverted** — the approach broke the build (`rdsetroot: not an elf`) and introduced complexity without proven benefit.

The `install.conf` / autoinstall approach was used in v1.0–v1.2's ISO builder (`build-iso.sh`), which was deleted in commit `4e217db`. That workflow embedded response files differently and is not applicable to disk images.

## Size Budget

| Component | Size |
|---|---|
| Base `install79.img` | ~801MB |
| `site79.tgz` (81 packages) | ~781MB |
| FFS metadata + 5% minfree | ~350MB |
| **Total** | **~1932MB** |

FFS overhead is significant: a 1932MB image file only exposes ~1.7GB usable FFS space. After injection, ~218MB remains free.

## Security

1. **Root required** for `vnconfig`, `mount`, `disklabel`
2. **Work directory** must be absolute path (rejects relative/`..`)
3. **Cleanup** failures are returned, not swallowed

## Next Steps

### Immediate (Before Next Release)

1. **Boot the image on live hardware**
   - Flash `Build/Images/openriot.img` to USB
   - Boot on OpenBSD-compatible machine
   - Document exact manual steps required

2. **Validate installer discovery**
   - Confirm `site79.tgz` appears in sets list
   - If not: debug `index.txt` format against standard entries

3. **Validate `install.site` execution**
   - Confirm post-install script runs after base install
   - Check `/var/log/` or add `set -x` for debugging

4. **Validate package installation**
   - After install: `which fish`, `which alacritty`
   - `rcctl check xenodm` should show disabled
   - `doas whoami` works

### Future (Post-Validation)

5. **Automate installer UX** (if desired)
   - Options:
     a. Modify `bsd.rd` to embed `auto_install.conf` in ramdisk (requires `rdsetroot` expertise)
     b. Serve `install.conf` via HTTP during netboot
     c. Accept manual steps and document them clearly

6. **Add `openriot.zip` for offline install**
   - Requires ~350MB additional space
   - Options: trim 3-4 large packages, or relax 2GB target

7. **Shrink image**
   - Truncating FFS after write corrupts superblock — not viable
   - Only option is reducing package count

## Files Changed (This Session)

| File | Change |
|------|--------|
| `source/imaging/build.go` | Removed `fsck`, added copy verification + `sync`, reverted `patchBsdRd` |
| `source/imaging/site.go` | Added `createAutoInstallConf()` (unused but kept for reference) |
| `source/fsutil/copy.go` | `io.Copy` for large files |
| `Makefile` | Cleanup `rm -f` updated for new artifacts |
| `docs/Image-Builder-Spec.md` | Updated with current architecture (see `docs/`) |

## Reference

- `docs/Image-Builder-Spec.md` — Canonical spec in repo
- `docs/v6.5-Release-Notes.md` — Release notes mentioning image builder
- `autoinstall(8)` — OpenBSD autoinstall mechanism
- `install.site(5)` — Post-install script documentation
- `install.sub` — OpenBSD installer source (`/distrib/miniroot/`)

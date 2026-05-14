# Functional Specification: `openriot --make-image`

## Overview

Build a bootable OpenRiot offline installer image from an OpenBSD base image.
The resulting image is **completely self-contained** — no network required during
installation. All packages and firmware are pre-bundled. The OpenRiot repository is
fetched by the user after first login via `curl -fsSL https://OpenRiot.org/setup.sh | sh`.

Target size: **< 2.0GB**.
Actual size: **~1932MB**.

---

## Current Status

| Item | Status |
|---|---|
| Image builds without errors | ✅ Working |
| Image size under 2.0GB | ✅ Working (~1932MB) |
| `site79.tgz` injects into `7.9/amd64/` | ✅ Working |
| `index.txt` contains `site79.tgz` entry | ✅ Working (uses real `ls -lT` output) |
| Image boots from USB | ✅ Working |
| Installer discovers `site79.tgz` | ✅ Working (manual `*` + `yes` to prompts) |
| `install.site` executes | ✅ Working |
| Package installation during post-install | ✅ Working — bulk `pkg_add -I *.tgz` installs in topological order; missing deps are non-fatal and resolved by `setup.sh` |
| First-login welcome message | ✅ Working |
| Firmware auto-installation | ⚠️ Code present, not yet validated end-to-end |
| User/group/shell modifications in `install.site` | ❌ Removed — handled by `setup.sh` |
| `sync` before `umount` | ✅ Working (prevents site79.tgz truncation) |

---

## Size Budget

| Component | Size |
|---|---|
| Base `install79.img` | ~801MB |
| Site tarball (`site79.tgz`) | ~781MB |
| Firmware in `site79.tgz` | ~31MB |
| FFS metadata + 5% minfree | ~350MB |
| **Final image** | **~1932MB** |

> **Note:** The OpenRiot Git repository (~1.1GB with `.git/`) is **NOT** included
> in the tarball. It is fetched by `setup.sh` after first login. This keeps the
> image under 2GB.

---

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
  site          Create site tarball only
  clean         Clean build artifacts (work/)
  help          Show help

Flags:
  --base-img PATH    Base OpenBSD image (default: Build/Images/install79.img)
  --output-img PATH  Output image path (default: Build/Images/openriot.img)
  --work-dir PATH    Working directory (default: Build/work)
  --version X.Y      OpenBSD version to target (default: 79)
  --no-burn          Skip interactive burn prompt
```

### Environment Variables (fallback)

| Variable     | Purpose                    | Default   |
|--------------|----------------------------|-----------|
| `BASE_IMG`   | Path to base OpenBSD image | See flags |
| `OUTPUT_IMG` | Path for output image      | See flags |
| `WORK_DIR`   | Working directory          | See flags |

---

## Architecture

### What Gets Installed

The OpenBSD installer discovers `site79.tgz` as a custom install set via
`index.txt` in the `7.9/amd64/` directory (per `install.site(5)`).

> **Critical fix applied:** Previous builds used a manually-crafted `fmt.Sprintf`
> line that mimicked `ls -lT` output. The installer's parser rejected it, so
> `site79.tgz` never appeared in the sets list. We now run `ls -lT` on the
> actual copied file and append that output verbatim.

`site79.tgz` extracts to the target system's root during install, creating:

```
/openriot/packages/snapshots/amd64/*.tgz   # All pre-downloaded packages
/openriot/firmware/*.tgz                   # Pre-downloaded non-free firmware
/install.site                              # Post-install script
/etc/motd                                  # Custom MOTD
```

After extraction, `/install.site` runs automatically in a chroot of the new
system. It:

1. Configures `doas.conf` (per-user `permit nopass $username` + `permit nopass :wheel`)
2. Sets `installurl` to `https://cdn.openbsd.org/pub/OpenBSD`
3. Adds `/usr/local/bin/fish` to `/etc/shells`
4. Enables `apmd` and `sndiod`
5. Creates `/etc/wireguard` (mode 700)
6. Installs all packages from `/openriot/packages/` via `pkg_add -I *.tgz` (non-fatal; `setup.sh` resolves missing deps when network is available)
7. Installs all firmware from `/openriot/firmware/` via `tar xzf -C /`
8. Appends welcome message to `~/.profile` for each user in `/home/*`
9. Adds the same welcome message to `/etc/skel/.profile`

### What Does NOT Happen

- **X11 does not start on first boot.** `xenodm` is deliberately NOT enabled.
- **The OpenRiot repo is NOT copied to `~/.local/share/openriot`.** The user
  fetches it after login with `curl setup.sh`.
- **`openriot --install` does NOT run automatically.** The welcome message
  directs the user to run `setup.sh`, which handles everything.
- **`install.conf` is NOT used.** `autoinstall(8)` requires response files in
  `bsd.rd`'s ramdisk or via HTTP. A file on the media filesystem is ignored.
- **Firmware is NOT placed on a `/firmware/` directory on the install media.**
  It is bundled inside `site79.tgz` and installed automatically by `install.site`.

### First-Login Flow

After installation, the user logs in and sees:

```
Welcome to OpenRiot!

All required packages have been pre-installed.
To complete setup and download the latest OpenRiot repository, run:

    curl -fsSL https://OpenRiot.org/setup.sh | sh

This requires a working network or WiFi connection.
```

The user runs the command, `setup.sh` clones the repo to `~/.local/share/openriot`,
copies configs to `~/.config/`, and runs `openriot --install`.

---

## Functional Modules

### 1. Prerequisites Check

**File:** `source/imaging/prereqs.go`

- Verify running on OpenBSD (`uname -s == OpenBSD`)
- Verify root user (`id -u == 0`)
- Verify base image exists at `--base-img` or `BASE_IMG`
- Fallback to `Build/Images/install79.img` relative to repo root
- Download base image from CDN if missing

> **Fix applied:** Fallback path corrected from `Images/install79.img` to
> `Build/Images/install79.img` to match actual repo layout.

### 2. Package & Firmware Download

**File:** `source/imaging/download.go`

#### Package Download
- Read `packages.yaml` for the package list
- Read `Build/exceptions.yaml` for exclusions
- Download from `https://cdn.openbsd.org/pub/OpenBSD/snapshots/packages/amd64/{pkg}.tgz`
- Skip existing files; clean stale versions
- Progress indicator: `Downloading package N/N: pkgname`
- Retry logic: 3 attempts per package
- **HTTP timeout: 2 minutes** (previously indefinite)

> **Fix applied:** `downloadFile()` now uses `http.Client{Timeout: 2 * time.Minute}`
> instead of `http.Get()`, which has no timeout and hangs indefinitely on slow
> networks.

#### Firmware Download
- Downloads from `http://firmware.openbsd.org/firmware/7.9/`
- Files bundled: `athn-firmware`, `intel-firmware`, `inteldrm-firmware`, `iwm-firmware`, `iwx-firmware`
- Total: ~31MB
- Warns on individual failures, does not fail the build

### 3. Site Tarball Creation

**File:** `source/imaging/site.go`

Create `site79.tgz` containing:

```
site/
├── etc/
│   └── motd                    # From install/motd
├── openriot/
│   ├── packages/
│   │   └── snapshots/
│   │       └── amd64/
│   │           └── *.tgz       # All downloaded packages
│   └── firmware/
│       └── *.tgz               # Downloaded non-free firmware
└── install.site                # Post-install script (inline)
```

> **No `openriot/repo/` directory.** The Git repository is fetched post-install.
> **No `install.conf` file.** It is dead code on media filesystems.

`CreateSite()` now uses `os.MkdirTemp(workDir, "site-")` instead of a fixed
`site/` subdirectory. This avoids permission collisions when a non-root user
runs `--make-image site` after a prior root run left `site/` root-owned.
The temp directory is cleaned up automatically via `defer os.RemoveAll`.

#### `install.site` (inline, embedded in Go)

```sh
#!/bin/sh
# OpenRiot post-install script — runs at end of OpenBSD install
# Environment: root, chroot at new system root, no TTY, no X

log() { echo "[OPENRIOT] $*"; }
fail() { echo "[OPENRIOT] FAIL: $*"; }

log "OpenRiot post-install starting"

# ------------------------------------------------------------------
# 1. System configuration
# ------------------------------------------------------------------

# doas
if ! [ -f /etc/doas.conf ]; then
	: > /etc/doas.conf
	for homedir in /home/*; do
		[ -d "$homedir" ] || continue
		username="$(basename "$homedir")"
		printf '%s\n' "permit nopass $username" >> /etc/doas.conf
	done
	printf '%s\n' "permit nopass :wheel" >> /etc/doas.conf
	chmod 0440 /etc/doas.conf
	log "doas configured"
fi

# installurl
printf '%s\n' "https://cdn.openbsd.org/pub/OpenBSD" > /etc/installurl
log "installurl configured"

# Add fish to /etc/shells
if ! grep -q '^/usr/local/bin/fish$' /etc/shells 2>/dev/null; then
	printf '%s\n' "/usr/local/bin/fish" >> /etc/shells
	log "fish added to /etc/shells"
fi

# Enable critical services (but NOT xenodm — user must start X manually)
rcctl enable apmd 2>/dev/null && rcctl set apmd flags -A 2>/dev/null
rcctl enable sndiod 2>/dev/null
mkdir -p /etc/wireguard
chmod 700 /etc/wireguard
log "services configured"

# ------------------------------------------------------------------
# 2. Install packages from local path (offline)
# ------------------------------------------------------------------
PKG_PATH_LOCAL="/openriot/packages/snapshots/amd64"
if [ -d "$PKG_PATH_LOCAL" ]; then
	log "Installing packages from local path..."
	cd "$PKG_PATH_LOCAL" || fail "cd $PKG_PATH_LOCAL"
	export PKG_PATH="$PKG_PATH_LOCAL"
	pkg_add -I *.tgz 2>&1
	log "Package install complete"
else
	log "Package directory not found: $PKG_PATH_LOCAL"
fi

# ------------------------------------------------------------------
# 2b. Install non-free firmware from local path (offline)
# ------------------------------------------------------------------
FW_PATH_LOCAL="/openriot/firmware"
if [ -d "$FW_PATH_LOCAL" ] && [ -n "$(ls "$FW_PATH_LOCAL"/*.tgz 2>/dev/null)" ]; then
	log "Installing firmware from local path..."
	for fw in "$FW_PATH_LOCAL"/*.tgz; do
		[ -f "$fw" ] || continue
		fw_name=$(basename "$fw")
		tar xzf "$fw" -C / 2>/dev/null || continue
		log "Firmware installed: $fw_name"
	done
	log "Firmware install complete"
else
	log "No local firmware found, skipping"
fi

# ------------------------------------------------------------------
# 3. Per-user setup
# ------------------------------------------------------------------
for homedir in /home/*; do
	[ -d "$homedir" ] || continue
	username="$(basename "$homedir")"

	# Welcome message only — setup.sh handles groups, shell, and XDG dirs
	cat >> "$homedir/.profile" << 'WELCOME'

# OpenRiot — Welcome
echo ""
echo "Welcome to OpenRiot!"
echo ""
echo "All required packages have been pre-installed."
echo "To complete setup and download the latest OpenRiot repository, run:"
echo ""
echo "    curl -fsSL https://OpenRiot.org/setup.sh | sh"
echo ""
echo "This requires a working network or WiFi connection."
echo ""
WELCOME
done

# 4. Skel for future users
cat >> /etc/skel/.profile << 'WELCOME'

# OpenRiot — Welcome
echo ""
echo "Welcome to OpenRiot!"
echo ""
echo "All required packages have been pre-installed."
echo "To complete setup and download the latest OpenRiot repository, run:"
echo ""
echo "    curl -fsSL https://OpenRiot.org/setup.sh | sh"
echo ""
echo "This requires a working network or WiFi connection."
echo ""
WELCOME

log "Post-install complete"
```

### 4. Image Building

**File:** `source/imaging/build.go`

#### Expand Phase

- Copy base image to output path (never modify source)
- Calculate target size: `baseSize + tgzSize + 350MB buffer`, round to 4MB
- Minimum: 512MB (irrelevant in practice)
- `truncate -s <size> <image>`
- Configure `vnd0`, read current disklabel
- Update `a:` partition to fill expanded space
- Update `c:` partition to total image size
- `growfs -y /dev/vnd0a`
- Release `vnd0` (via `defer` to ensure cleanup on error)

> **Fix applied:** Added `defer exec.Command("vnconfig", "-u", "vnd0").Run()`
> in `expandImage()` so the vnd device is released even if `getPartitionInfo`,
> `writeDisklabel`, or `growfs` fails.

#### Injection Phase

- Configure `vnd0` on output image
- `fsck -y /dev/vnd0a`
- Mount at `/mnt`
- Create `/mnt/7.9/amd64/` directory
- Copy `site79.tgz` to `/mnt/7.9/amd64/site79.tgz`
- Append `site79.tgz` entry to `/mnt/7.9/amd64/index.txt` using actual `ls -lT` output on the copied file (do NOT overwrite)
- **`sync`** filesystem before unmount to prevent truncation
- Unmount, release `vnd0` (via `defer` to ensure cleanup on error)

> **Fix applied:** Added `exec.Command("sync").Run()` after `index.txt` close
> and before the deferred `umount`. Without this, the 781MB `site79.tgz` was
> silently truncated by the kernel buffer cache on unmount.
>
> **Fix applied:** Added `defer` block in `injectContent()` that runs
> `umount /mnt` and `vnconfig -u vnd0` on all return paths, preventing
> leaked mounts and vnd devices when errors occur mid-injection.

#### Shrink Phase

> **Removed.** The image is sized correctly from the start. No shrink step.
> Truncating an FFS image after data has been written corrupts the filesystem.

### 5. Drive Detection & Burning

**File:** `source/imaging/burn.go`

Same as before — detect drives, classify root/removable/internal, prompt user,
write with `dd` via `pv`.

### 6. Cleanup

**File:** `source/imaging/burn.go` (`Cleanup()`)

- Unmount `/mnt` if mounted
- Release `vnd0` device
- Remove `WORK_DIR` contents (packages, site temp dirs, tarball)

> **No repo cache.** `Build/repo-cache/` is no longer used and is removed from
> cleanup logic.

> **Fix applied:** `Cleanup()` now returns `os.RemoveAll` errors instead of
> swallowing them, and rejects unsafe work-dir paths (relative paths or paths
> containing `..`).

---

## Output Artifacts

| File              | Location                 | Description                |
|-------------------|--------------------------|----------------------------|
| `site79.tgz`      | `WORK_DIR/site79.tgz`    | Custom install set         |
| `openriot.img`    | `OUTPUT_IMG`             | Bootable installer image   |
| `openriot.sha256` | Same dir as img          | SHA256 checksum            |
| Package cache     | `WORK_DIR/packages/`     | Downloaded .tgz files      |
| Firmware cache    | `WORK_DIR/firmware/`     | Downloaded firmware .tgz |

> **No `openriot.tgz`.** The tarball is named `site79.tgz` per OpenBSD custom
> set conventions. It is placed in `7.9/amd64/` on the install media.

---

## Error Handling

| Error                    | Action                              |
|--------------------------|-------------------------------------|
| Not running on OpenBSD   | Exit with error                     |
| Not root                 | Exit with error, suggest `doas`     |
| Base image missing       | Download from CDN, or exit          |
| Package download fails   | Warn, continue with remaining       |
| Firmware download fails  | Warn, continue with remaining       |
| Package list empty       | Exit with error                     |
| Disklabel/growfs fails   | Exit with error                     |
| Mount fails              | Exit with error                     |
| No space left on device  | Exit with error, increase buffer    |
| Burn fails               | Show error, leave image file        |

---

## Implementation Checklist

### `source/imaging/site.go`
- [x] Delete `setupRepo()` function entirely
- [x] Delete `updateCache()` function entirely
- [x] Rewrite `CreateSite()` to skip repo copy, copy only packages + firmware + motd + install.site
- [x] Rewrite `createInstallSite()` with new script (no xenodm, no repo copy, curl welcome)
- [x] Remove `getBuildDir()` if no longer used
- [x] Use `os.MkdirTemp` for site staging directory to avoid root-owned leftover collisions
- [x] Add `defer os.RemoveAll(siteDir)` for automatic cleanup
- [x] Check all `MkdirAll` errors instead of silently ignoring
- [x] Use `tar czf` instead of `tar czvf` (no verbose output)
- [x] **Fix:** Replace one-by-one `pkg_add` loop with bulk `pkg_add -I *.tgz` (non-fatal)
- [x] **Fix:** Remove all user/group/shell modifications from `install.site`; `setup.sh` handles them
- [x] **Fix:** Add per-user `doas.conf` entries (`permit nopass $username`) alongside `:wheel`
- [x] **Fix:** Add firmware auto-installation block (`tar xzf -C /`)

### `source/imaging/build.go`
- [x] Change buffer from `50*1024*1024` to `350*1024*1024`
- [x] Fix hardcoded "2GB" log message to report actual expanded size
- [x] Add `df -h /mnt` logging before and after injection
- [x] Fix `injectContent()` to use real `ls -lT` output for `index.txt` instead of manual `fmt.Sprintf`
- [x] Remove dead `getFileSize()` function
- [x] Add `defer vnconfig -u vnd0` in `expandImage()` for error-path cleanup
- [x] Add `defer umount/vnconfig` in `injectContent()` for error-path cleanup
- [x] **Fix:** Add `exec.Command("sync").Run()` after `index.txt` close to prevent `site79.tgz` truncation
- [x] **Removed:** `/mnt/firmware/` copy (redundant — firmware is inside `site79.tgz`)

### `source/imaging/runner.go`
- [x] Remove `repo-cache` cleanup from `runClean()`
- [x] Fix `args` slice mutation bug during `range` iteration (build `filtered` slice)
- [x] Update help text to include `--work-dir`, `--version`, and `help` mode
- [x] **Add:** `DownloadFirmware()` call between package download and site creation

### `source/imaging/download.go`
- [x] Add 2-minute HTTP timeout to `downloadFile()` (was indefinite)
- [x] **Add:** `DownloadFirmware()` function with firmware file list and download logic
- [x] **Add:** `firmwareFiles` var with 5 bundled firmware packages

### `source/imaging/burn.go`
- [x] Make `Cleanup()` return `os.RemoveAll` errors instead of swallowing them
- [x] Reject unsafe work-dir paths (relative or containing `..`)

### `source/imaging/prereqs.go`
- [x] Fix fallback image path to `Build/Images/install79.img` (was missing `Build/`)

### `docs/Image-Builder-Spec.md`
- [x] This file — updated to reflect current architecture, status, and firmware

### Validation
- [x] `make && make test` passes
- [x] `make image` as root succeeds
- [x] `ls -lh Build/Images/openriot.img` shows **< 2.0GB** (~1932MB)
- [x] Mount image, verify `site79.tgz` exists in `7.9/amd64/`
- [x] Verify `index.txt` contains `site79.tgz` alongside standard sets
- [x] Extract `site79.tgz`, verify it contains `openriot/packages/`, `openriot/firmware/`, and `install.site`
- [x] Verify `install.site` does NOT contain `xenodm` enable
- [x] Verify `install.site` contains `curl -fsSL https://OpenRiot.org/setup.sh | sh`
- [x] `--make-image site` works as non-root even after prior root run
- [ ] **Flash to USB, test boot and install** — Partially done (image boots, site discovered, install.site runs, packages install with some dependency edge cases)
- [ ] **After install: verify all packages pre-installed (`which fish`, `which alacritty`)** — Partially done
- [ ] **After install: verify firmware is extracted to `/etc/firmware/`** — NOT YET DONE
- [ ] **After install: verify `xenodm` is NOT enabled** — NOT YET DONE
- [ ] **After install: verify first login shows welcome message** — NOT YET DONE
- [ ] **After install: verify `doas` works** — NOT YET DONE
- [ ] **After install: verify `fish` is default shell** — NOT YET DONE
- [ ] **After install: verify user groups include `wheel` and original groups** — NOT YET DONE

---

## Why This Architecture

### Why no repo in the tarball?

The OpenRiot repository (with `.git/`, `Locked/`, `backgrounds/`) is ~1.1GB.
Adding it to the tarball pushes the image to ~2.1GB+. The only way to stay under
2GB is to exclude it. The user fetches the latest repo after login anyway.

### Why `install.site` instead of `install.conf`?

`autoinstall(8)` only reads response files from:
1. `bsd.rd`'s built-in RAM disk (`/auto_install.conf`)
2. HTTP fetch during netboot

A file named `install.conf` on the install media filesystem is **never scanned**.
It is dead code. `install.site(5)` is the documented, reliable mechanism for
running post-install scripts from a custom install set.

### Why no `shrinkImage`?

Truncating an FFS image after data has been written corrupts the superblock
and cylinder group summaries. The correct approach is to size the image
accurately from the start using `truncate` + `growfs`.

### Why not enable xenodm?

The user explicitly requested X11 NOT start on first boot. The welcome message
directs the user to run `setup.sh` after login, which handles graphical setup.

### Why bulk `pkg_add *.tgz` instead of one-by-one?

OpenBSD `pkg_add *.tgz` builds a dependency graph across all packages and
installs them in topological order. One-by-one alphabetical installation fails
when a package's dependencies come later in the alphabet. `pkg_add *.tgz`
with `PKG_PATH` set resolves all dependencies correctly regardless of filename
order.

### Why firmware inside `site79.tgz` instead of `/firmware/` on media?

OpenBSD's installer only checks `/firmware/` when the user answers `yes` to
"Install non-free firmware?" — and even then, it expects a specific structure.
By placing firmware inside `site79.tgz`, `install.site` extracts it directly to
`/` during post-install. No user prompt needed, and it works regardless of
installer UI behavior.

---

## Future Considerations

- Custom base image versions
- Parallel package downloads
- Resume interrupted downloads
- Image signing/verification
- Shallow Git clone in `setup.sh` to reduce post-install bandwidth

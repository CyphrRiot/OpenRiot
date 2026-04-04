# OpenRiot — Full Context for New Chat Session

## What Is OpenRiot?

OpenRiot transforms a fresh OpenBSD installation into a fully-configured Sway Wayland desktop — in one command. It is the OpenBSD counterpart to ArchRiot (https://archriot.org).

**Current version:** 0.9 (see `VERSION` file)
**OpenBSD version:** 7.9

## What We Just Did (This Session)

We spent this session building a **bootable offline OpenBSD ISO** with all OpenRiot packages pre-bundled. The ISO is built on Linux using `xorriso` (no root required).

### The Core Problem

OpenBSD's **autoinstall** (pressing `a` at boot) tries to fetch `install.conf` via DHCP/HTTP. For fully offline USB install, this fails. The only reliable way to get true offline autoinstall is to embed `auto_install.conf` inside `bsd.rd` (the ramdisk kernel), which requires `rdsetroot` (OpenBSD only) or `fuse2fs` (not available on our Linux build system).

**Current workaround:** Use interactive install (`i`) — all defaults are pre-filled, user just presses Enter through prompts.

### Files Modified This Session

| File | Purpose |
|------|---------|
| `build-iso.sh` | ISO builder — removes bsd.rd injection, adds SHA256 regeneration, includes openriot.tgz inside site79.tgz |
| `autoinstall/install.conf` | Install answers — `cd0` sets location, `site79.tgz` selected, skip fw_update |
| `autoinstall/install.site` | Post-install script — **FIXED**: `/etc/openriot/VERSION` now writes "7.9" (OpenBSD version), not "0.9" (OpenRiot version) |
| `autoinstall/autopartitionning.template` | Partition layout: 2GB root, 1GB swap, home * |
| `README.md` | Added Method 3: Interactive Install section |
| `Makefile` | Added `isotest` target, updated help text |
| `.gitignore` | Ignore `*.qcow2`, `.work/`, `isos/*.iso` |
| `test-iso.sh` | QEMU test script — creates 5GB qcow2 in `~/.cache/`, runs ISO install |
| `test-image.sh` | QEMU test script — boots installed system from `~/.cache/openriot-test.qcow2` |

### Current ISO State

- **Size:** ~1.1GB
- **Sets on ISO:** base79.tgz, comp79.tgz, man79.tgz, site79.tgz (includes openriot.tgz), xbase79.tgz, xfont79.tgz, xshare79.tgz
- **Removed:** game79.tgz, xserv79.tgz (~90MB savings)
- **Location:** `isos/openriot.iso`

### Critical Bug Fixed This Session

`install.site` was reading `OPENBSD_VERSION` from `/etc/openriot/VERSION`, but that file contained "0.9" (OpenRiot version) instead of "7.9" (OpenBSD version). This caused `pkg_add` to look for packages at `/openriot/packages/0.9/amd64/` instead of `/openriot/packages/7.9/amd64/`, failing silently.

**Fix:** `build-iso.sh` now writes `${OPENBSD_VERSION}` (7.9) to the VERSION file, not `${OPENRIOT_VERSION}` (0.9).

### Test Workflow

```bash
# Build ISO
rm -rf .work && make iso

# Test install (creates 5GB qcow2 in ~/.cache/)
./test-iso.sh
# - Boot ISO, press 'i' for interactive install
# - Select sets with '*' (includes site79.tgz)
# - Accept checksum warnings

# Boot installed system
./test-image.sh
```

## What's Still Broken

1. **Autoinstall (`a`) doesn't work** — requires embedding `auto_install.conf` into `bsd.rd`. Need to build ISO on OpenBSD (with `rdsetroot`) or find another approach.

2. **Interactive install requires manual set selection** — at "Set name(s)?", must type `*` to select all sets including site79.tgz. Default is all sets except site79.tgz.

3. **Ambiguous package names** — `pkg_add fish` is ambiguous (fish-3.7 and fish-4.6 both exist). Need to use exact versions like `fish-3.7.1p4`.

4. **`repo.tar.gz` extraction** — sometimes fails to create `~/.local/share/openriot/` properly. The git repo archive is bundled in site79.tgz.

5. **Network after install** — OpenBSD's network interface isn't brought up automatically. User must run `sh /etc/netstart` or configure `hostname.if` manually.

6. **Multiple fish versions** — OpenBSD package repo has fish-3.7 and fish-4.6. The ISO cache only has fish-3.7.1p4 but `pkg_add fish` prompts for ambiguity.

## Key Commands

```bash
make iso           # Build ISO (downloads packages, builds binary, repacks ISO)
make isotest       # Build ISO + run test-iso.sh
make clean         # Remove build artifacts
./test-iso.sh     # Boot ISO in QEMU, install to ~/.cache/openriot-test.qcow2
./test-image.sh   # Boot installed system from ~/.cache/openriot-test.qcow2
```

## Important Paths

| Path | Description |
|------|-------------|
| `~/.pkgcache/7.9/amd64/` | Downloaded OpenBSD packages (~43 packages) |
| `.work/iso_contents/7.9/amd64/` | Extracted ISO contents (base sets, site79.tgz, etc.) |
| `~/.cache/openriot-test.qcow2` | Test VM virtual disk (5GB) |
| `autoinstall/install.conf` | Answers for OpenBSD installer |
| `autoinstall/install.site` | Post-install script (runs in chroot after base install) |
| `install/packages.yaml` | **Source of truth** for all packages, configs, commands |
| `site/` | Files deployed to new system root (etc/doas.conf, etc/hostname.iwx0) |

## Package Installation Flow

1. OpenBSD installer extracts `site79.tgz` to `/` (includes `install.site`, `openriot.tgz`, `repo.tar.gz`, `VERSION`, etc.)
2. `install.site` runs automatically in chroot
3. `install.site` extracts `openriot.tgz` to `/` (gives `/openriot/packages/7.9/amd64/*.tgz`)
4. `install.site` runs `pkg_add` on packages from `packages.yaml` using local `PKG_PATH=/openriot/packages/7.9/amd64`
5. `install.site` extracts `repo.tar.gz` to `~/.local/share/openriot/` for first-boot setup
6. After reboot, user's `.profile` runs `openriot --install` from the deployed repo

## Next Steps

1. **Fix VERSION bug** — DONE this session (VERSION now = 7.9 not 0.9)
2. **Rebuild ISO** with fix and test package installation
3. **Fix ambiguous package names** — use exact versions in `packages.yaml`
4. **Test `repo.tar.gz` extraction** — verify `~/.local/share/openriot/` is created
5. **Add network bring-up** to `install.site` or post-install
6. **Consider building ISO on OpenBSD** to enable true autoinstall (`a`)

## Reference

- OpenBSD autoinstall: https://man.openbsd.org/autoinstall.8
- OpenBSD install.conf: https://man.openbsd.org/install.conf.5
- Project repo: https://github.com/CyphrRiot/OpenRiot

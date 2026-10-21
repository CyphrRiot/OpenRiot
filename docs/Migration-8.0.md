# OpenRiot Migration Plan: OpenBSD 7.9 to 8.0

Status: proposed
Target: OpenBSD 8.0 amd64
Scope: source, installer, image builder, packages, setup script, art assets

## 1. Summary

OpenBSD 8.0 amd64 sets are published on the CDN. The release page is
dated October 20, 2026, and the sets were built on October 6, 2026. The
formal announcement date does not block preparation: `install80.img`,
`base80.tgz`, and the 8.0 package tree are already downloadable.

OpenRiot currently assumes 7.9 in the migration path, the image builder,
the package pins, and the setup script. The desktop stack itself needs
no configuration rewrite. The work is version plumbing, package
re-pinning, media regeneration, and a full art refresh.

OpenRiot release numbering is independent of the OpenBSD target. This
plan does not change `VERSION`, the README badge, or existing release
notes.

Two supported execution paths are defined in section 12: a clean 8.0
install followed by `setup.sh`, and a 7.9 to 8.0 upgrade followed by
`setup.sh`.

## 2. Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Target | OpenBSD 8.0 amd64 | Current release line |
| Upgrade method | `sysupgrade -R 8.0 -n`, then reboot | User controls reboot |
| Release date constant | October 6, 2026 | Convention uses sets build date |
| Desktop | Keep i3 and X11 | 8.0 breaks nothing used by OpenRiot |
| Wayland | No migration in this plan | No functional need |
| Package source | 8.0 stable pins | Avoid snapshot drift |
| Art | Replace all slots | User requirement |
| OpenRiot version | Separate release process | Do not bump here |

The release date constant follows existing behavior. OpenBSD 7.9 uses
May 6, 2026 in `installer.ReleaseDate`, which is the sets build date,
while the announcement was May 19. OpenBSD 8.0 sets were built on
October 6, 2026, so the constant becomes October 6.

## 3. Current 7.9 Assumptions

### 3.1 Setup script

`setup.sh` contains:

```sh
OPENBSD_MIN_VERSION="7.9"
OPENBSD_BRANCH="7.9"
CDN_BRANCH="$OPENBSD_BRANCH"
```

`check_openbsd_version()` accepts 8.0 because the comparison is
greater-or-equal. The failure is in `install_x11_sets()`:

```sh
AMD64_PATH="https://cdn.openbsd.org/pub/OpenBSD/${CDN_BRANCH}/amd64"
VER_NUM=$(uname -r | sed 's/\.//' | sed 's/-.*//')
```

On 8.0, `VER_NUM` is `80`, but `CDN_BRANCH` is still `7.9`. The script
requests `xbase80.tgz` from the 7.9 directory and fails if X11 is
missing. The banner also prints the 7.9 minimum.

This file is frozen by `AGENTS.md`. A migration exception is required.

### 3.2 Build and migration commands

| Location | Assumption |
|----------|------------|
| `Makefile` | `OPENBSD_VERSION = 7.9` |
| `Makefile` image target | Removes `Build/work/site79.tgz` |
| `source/main.go` | Default `openbsdVersion = "7.9"` |
| `source/installer/version.go` | `ReleaseDate` is May 6, 2026 |
| `source/commands/helpers.go` | Prints `7.9`, runs `sysupgrade -R 7.9` |
| `source/commands/commands.go` | `--pkg-update` forces the 7.9 package URL |
| `source/imaging/prereqs.go` | Defaults to `install79.img`, version `79` |
| `source/imaging/build.go` | Mounts `/mnt/7.9/amd64`, creates `site79.tgz` |
| `source/imaging/runner.go` | Help text names `install79.img` |
| `install/packages.yaml` | `openbsd_version: "7.9"` and 180 exact pins |
| `Build/README.md` | Links `Images/install79.img` |

Tests also encode 7.9:

| Location | Assumption |
|----------|------------|
| `source/installer/version_test.go` | 7.9 kernel fixtures and May dates |
| `source/imaging/prereqs_test.go` | Default version `79` |

### 3.3 Package pins

`install/packages.yaml` pins exact package revisions from the 7.9
repository and five firmware archives. Exact revisions do not survive a
release bump. When an exact pin fails, `InstallPackages()` retries the
base name. If that fails, `tryInstall()` retries with `-D snap`, even on
a release host. That fallback can pull snapshot packages onto a stable
system.

## 4. Phase 0: Preflight and Optional Backup

Run on the 7.9 machine before any migration work:

1. Confirm architecture and release:

   ```sh
   uname -m
   uname -r
   ```

2. Confirm `/usr` has at least 1.1G free:

   ```sh
   df -h /usr
   ```

3. Create a full backup (optional):

   ```sh
   openriot --backup
   ```

   Skipping this removes the only rollback path.

4. Record installed packages and custom configuration:

   ```sh
   pkg_info -a > ~/pkg_info-7.9.txt
   doas syspatch -c
   ```

5. Verify the upgrade kernel before booting it:

   ```sh
   signify -C -p /etc/signify/openbsd-80-base.pub -x SHA256.sig bsd.rd
   ```

6. Do not start the upgrade with less than 1.1G free on `/usr`.
   `sysupgrade` also refuses to proceed when `/usr` is over 90 percent
   full, but the preflight check catches this before a reboot.

## 5. Phase 1: Source Migration

Make these changes in order. Run `make test` after each edit.

### 5.1 Version constants

| File | Change |
|------|--------|
| `Makefile` | `OPENBSD_VERSION = 8.0` |
| `Makefile` | Image messages say 8.0 |
| `Makefile` image target | Remove version-derived site tarball |
| `source/main.go` | Default `openbsdVersion = "8.0"` |
| `source/installer/version.go` | `ReleaseDate` to October 6, 2026 |

The image target should remove `Build/work/site80.tgz`. If the target
continues to hardcode a name, change it explicitly. A derived name is
preferred so a future release does not repeat this migration.

### 5.2 Migration commands

Replace every release string and `sysupgrade -R` argument in
`source/commands/helpers.go` and `source/commands/commands.go` with one
shared target-release constant. At minimum:

- `runInstall()` banner and `sysupgrade -R` call
- `checkReleasePath()` stable, pre-release, and post-release messages
- `--check-release-path` description
- `--pkg-update` description, status text, and `PKG_PATH`
- Fresh-install message that names `install79.img`

The target release string must be defined once. Hardcoding `8.0` in
multiple functions repeats the defect this plan exists to fix.

### 5.3 Image builder paths

| File | Change |
|------|--------|
| `source/imaging/prereqs.go` | Derive default base image from `cfg.Version` |
| `source/imaging/prereqs.go` | Derive fallback `.img` and `.iso` names |
| `source/imaging/build.go` | Derive `/mnt/<version>/amd64` |
| `source/imaging/build.go` | Derive `site<version>.tgz` names |
| `source/imaging/runner.go` | Update help text |
| `source/imaging/prereqs_test.go` | Expect version `80` and `install80.img` |
| `source/installer/version_test.go` | Replace 7.9 fixtures with 8.0 |

Do not replace `79` with `80` in place. Derive the path from
`cfg.Version` so `--version 79` still works when the 7.9 package pins
are available.

### 5.4 Package fallback hardening

Change `tryInstall()` so `-D snap` is attempted only when the target is
a snapshot repository. On a release host the order should be:

1. exact pin
2. base package name from the running release
3. fail with the package name and output

This prevents a transient mirror error from installing snapshot
packages on 8.0 stable.

### 5.5 Comments and documentation strings

Update comments that claim 7.9 is the supported release. Do not edit
historical release notes. Do not rewrite unrelated comments.

## 6. Phase 2: Packages and Firmware

`install/packages.yaml` has 180 package entries across core, desktop,
crypto, media, source, fonts, and system sections.

1. Upgrade a representative 8.0 host and install the full package set.
2. Regenerate exact pins:

   ```sh
   openriot --sync-packages
   ```

3. Set the version field:

   ```yaml
   openbsd_version: "8.0"
   ```

4. Re-pin the firmware block against the 8.0 firmware tree:

   ```sh
   ftp -o - https://firmware.openbsd.org/firmware/8.0/
   ```

   Verify each listed archive exists before accepting it. Firmware
   revision names change between releases.

5. Verify pins:

   ```sh
   openriot --check-packages
   ```

6. Test a clean install on a machine that has none of the pinned
   packages. Exact pins must resolve without the base-name fallback.

Known package changes in 8.0 that affect pins:

| Package | 7.9 pin | 8.0 expectation |
|---------|---------|-----------------|
| Go | 1.26.x | 1.27.1 |
| Python | 3.13.13 | 3.14.7 |
| FFmpeg | 8.0.1p1v1 | 9.0.2 |
| Rust | 1.94.x | 1.98.1 |
| LLVM | 21.1.8 | 21.1.8 and 22.1.8 |

The OpenRiot module declares Go 1.24.2 with toolchain 1.24.5. It builds
with the 8.0 Go package; no module change is required.

## 7. Phase 3: Setup Script

Preferred change, if the exception is granted:

```sh
OPENBSD_MIN_VERSION="8.0"
OPENBSD_BRANCH="8.0"
```

If dual 7.9 and 8.0 installs are required, derive the branch from the
running release and stop pinning `CDN_BRANCH`:

```sh
OPENBSD_BRANCH=$(uname -r | sed 's/-.*//')
```

This keeps X11 set downloads aligned with the running kernel. The
minimum version check still applies.

Without either change, fresh 8.0 installs without X11 are unsupported.
Document this limitation in the release notes rather than shipping a
script that requests files from the wrong directory.

## 8. Phase 4: Image Builder and Media

The image builder mounts the base image and injects a site tarball into
the sets directory. Both the mount path and the tarball name are
currently fixed at 7.9.

Required result:

| Item | Value |
|------|-------|
| Base image | `Build/Images/install80.img` |
| Sets directory | `/mnt/8.0/amd64` |
| Site tarball | `site80.tgz` |
| Index entry | `ls -lT site80.tgz` |
| Output | `Build/Output/openriot.img` |

Build steps on an 8.0 host:

```sh
make
doas make image
```

Validation:

1. Confirm `Build/Images/install80.img` is non-empty.
2. Confirm `site80.tgz` appears in `index.txt`.
3. Boot the image in a VM.
4. Complete an unattended install with no network.
5. Confirm `install.site` runs and installs local packages.
6. Confirm X11 starts and `openriot --version` reports the build.

Fresh media is not a substitute for testing the upgrade path. Both are
required.

## 9. Phase 5: Full Image Refresh

All existing art is custom OpenRiot work. The 8.0 release artwork does
not replace it automatically. This phase replaces every slot with
8.0-branded art supplied as source masters.

### 9.1 Inventory

| Directory | Count | Current size | Format |
|-----------|-------|--------------|--------|
| `backgrounds/` | 50 | 16.5M | WebP |
| `Locked/` | 42 | 26.0M | WebP |

Names are `01.webp` through the final slot in each directory. Preserve
both counts and names unless a deliberate set-size change is approved.

### 9.2 Inputs

Required from the user before this phase starts:

- 50 background masters, PNG, at least 2560x1440
- 42 lock screen masters, PNG, at least 2560x1440
- Confirmed 8.0 branding text and palette
- Confirmed OpenRiot version string for lock screens

Lock screen requirements:

- Large, legible "Locked"
- "Enter your password" below it
- "OpenBSD 8.0" and "OpenRiot" in the frame
- OpenRiot version number in the frame
- No date, taskbar, or unrelated UI
- Nothing that becomes unreadable at 1920x1080

Background requirements:

- No text that conflicts with the lock screen overlay
- Safe center composition for ultrawide and 16:9 crops
- Consistent palette across the set

### 9.3 Conversion

1. Place masters in `backgrounds/` and `Locked/` with their final
   numeric names.
2. Convert at quality 95:

   ```sh
   openriot --convert
   ```

   `make convert` is equivalent. The converter writes `.webp` and
   removes the source `.png`.
3. Verify no PNGs remain:

   ```sh
   ls backgrounds/*.png Locked/*.png
   ```

   The command must report no matches.

### 9.4 Lock cache invalidation

Existing users keep a rendered lock cache at:

```text
~/.local/share/openriot/Locked/.cache/<width>x<height>/
```

New source images are not used while the old cache exists. After the
asset update:

```sh
rm -rf ~/.local/share/openriot/Locked/.cache
openriot --build-lock-cache
```

The next lock rebuilds the cache automatically, but the explicit step
ensures the new art is validated before a user locks the screen.

### 9.5 Quality gates

| Gate | Requirement |
|------|-------------|
| Count | 50 backgrounds, 42 lock screens |
| Names | Numeric, contiguous, no gaps |
| Format | WebP only in both directories |
| Decode | Every file opens and reports expected dimensions |
| Branding | No 7.9 text anywhere |
| Readability | "Locked" legible at 1920x1080 |
| Cropping | Lock art fills 16:9 without dead bars |
| Size | Backgrounds and locks stay near current totals |
| Visual | Full set reviewed at 100 percent |

Do not prune git history or rewrite asset blobs as part of this phase.
That is a separate operation with its own approval.

## 10. Phase 6: Validation Matrix

| Scenario | Steps | Pass condition |
|----------|-------|----------------|
| Path B | `sysupgrade`, `sysmerge`, `pkg_add -u` | Boots, X, pins OK |
| Path A setup | Fresh 8.0 plus setup | Desktop works |
| 7.9 current | Re-run installer | Offers 8.0, no downgrade |
| Post-release | Kernel newer than sets | Warns, no `-R` |
| Fresh install | Boot `install80.img` offline | `install.site` completes |
| Image build | `doas make image` | `site80.tgz` injected |
| Package pins | Clean host, install all | No exact-pin failures |
| Lock cache | Replace art, purge cache | New art appears |
| Backgrounds | Next and previous shortcuts | All 50 slots reachable |
| Syspatch | `openriot --syspatch-check` | Applies 8.0 errata |
| Update check | `openriot --check-update` | Reports current release |
| Mirror | `openriot --mirrors` | Fastest mirror selected |

Tests to run on every source edit:

```sh
make test
```

Targeted checks:

```sh
make test-img
./install/openriot --check-release-path
```

## 11. Phase 7: Release and Documentation

1. Read `Specs/Release-Style-Guide.md`.
2. Write release notes for the next OpenRiot version. Do not edit
   existing notes.
3. Update the README release notes TOC, newest first, ten entries
   maximum.
4. Update `Build/README.md` to `install80.img`.
5. Update `assets/WhySeries/` claims only after 8.0 is tested. The
   current documents name 7.9 explicitly.
6. Leave historical release notes untouched.
7. Rebuild `install/openriot` with the 8.0 build flags and include it
   in the release. Setup users receive behavior from this tracked
   binary, not from source edits alone.

Release notes must state:

- OpenBSD 8.0 target
- Upgrade command
- Package re-pin requirement
- Setup script status
- Art refresh status
- Path A and Path B support status
- Any unsupported path

## 12. Execution Paths

Two paths are supported. Both must pass before release.

### 12.1 Prerequisites

| Item | Required state |
|------|----------------|
| `setup.sh` | Branch derived from `uname -r` or set to 8.0 |
| `install/packages.yaml` | 8.0 pins |
| `install/openriot` | Rebuilt and shipped |
| Media | `install80.img` tested for Path A |

If `setup.sh` stays frozen at 7.9, Path A is unsupported when X11 sets
are not selected during install.

### 12.2 Path A: Clean 8.0 Install

1. Install OpenBSD 8.0 amd64. Select X sets or leave them out.
2. Boot the new system and confirm network access.
3. Run the installer:

   ```sh
   curl -fsSL https://OpenRiot.org/setup.sh | sh
   ```

4. Setup must complete these steps:

   - Configure `/etc/installurl` and `/etc/pkg_add.conf`.
   - Install `git` as the only bootstrap package.
   - Clone the OpenRiot repository to `~/.local/share/openriot`.
   - Install the 8.0 package pins.
   - Download 8.0 X sets when X11 is missing.
   - Copy OpenRiot configuration and assets.

5. Reboot into X and verify:

   ```sh
   openriot --version
   openriot --check-packages
   openriot --check-update
   ```

6. Lock the screen once and confirm the new lock art appears.

Acceptance:

- Setup exits without a failed package.
- `pkg_info` shows the pinned 8.0 revisions.
- X starts and i3 loads.
- All 50 backgrounds are present.
- The lock cache builds from all 42 lock screens.

### 12.3 Path B: 7.9 to 8.0 Upgrade

1. Complete the Phase 0 preflight and backup.
2. Upgrade the base system:

   ```sh
   doas sysupgrade -R 8.0 -n
   ```

3. Reboot explicitly:

   ```sh
   doas reboot
   ```
4. Merge configuration and update firmware:

   ```sh
   doas sysmerge      # recommended: merge base config changes
   doas fw_update     # optional: only if firmware updates are available
   ```

   Setup runs `pkg_add -Iu` before installing pins. You do not need to
   run `pkg_add -u` separately unless setup cannot complete.

5. Re-run the installer to apply the OpenRiot release:

   ```sh
   curl -fsSL https://OpenRiot.org/setup.sh | sh
   ```

   This step re-applies OpenRiot settings after `sysmerge`. OpenRiot
   edits `/etc/pf.conf`, `/etc/sysctl.conf`, `/etc/rc.conf.local`, and
   `/etc/X11/xenodm/Xsetup_0`. Those files can be merged or replaced
   during the upgrade.

6. Verify:

   ```sh
   openriot --version
   openriot --check-packages
   openriot --check-release-path
   ```

7. Lock the screen and confirm the new art appears.

Acceptance:

- The system boots 8.0 without a downgrade warning.
- `sysmerge` leaves no unresolved conflict.
- `pkg_add -u` completes.
- Package pins resolve on 8.0.
- X, audio, network, and the lock screen work.
- The art refresh is visible.

Direct upgrade from 7.8 to 8.0 is unsupported. OpenBSD tests only one
release step.

### 12.4 sysupgrade Behavior

`sysupgrade` on a 7.9 release selects the next release, so it can find
and install 8.0 without `-R`. Use `-R 8.0` to pin the target and make
the intent explicit.

`sysupgrade` downloads the sets, verifies them with `signify`, writes
`/bsd.upgrade`, and reboots by default. There is no check-only mode.
`-n` fetches and verifies the sets and prepares `/bsd.upgrade` without
rebooting. It is not a lightweight availability query.

The check depends on `/etc/installurl`. If that mirror has not synced
8.0, the upgrade fails even when another mirror has the sets.

### 12.5 Upgrade Guide Review

Reference: https://www.openbsd.org/faq/upgrade80.html

| 8.0 change | OpenRiot impact |
|------------|-----------------|
| bgpd enforces 4-byte AS | None, bgpd is not used |
| Kea 3.2 config migration | None, Kea is not used |
| libmaxminddb databases removed | None, no GeoIP dependency |
| stalwart-mail 0.16 migration | None, not used |
| Files to remove | None |

The guide's package step is `pkg_add -u`. Run it before re-running
setup so OpenRiot installs pins against an updated package base.

### 12.6 Acceptance Summary

| Path | Required result |
|------|-----------------|
| A | Clean 8.0 plus setup yields a working desktop |
| B | 7.9 upgrades to 8.0 and setup restores OpenRiot |
| Both | 8.0 pins, 8.0 art, no 7.9 runtime strings |

## 13. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| X11 sets | Wrong CDN branch | Exception or derive branch |
| Package pins | Fallback installs | Sync and verify |
| Snapshot fallback | Mixed package base | Gate `-D snap` |
| Image paths | Inject into 7.9 path | Derive from `cfg.Version` |
| Stale lock cache | Old art stays visible | Purge and rebuild |
| Registry | Signature mismatch | 8.0 installurl and re-pin |
| Downgrade | Unsupported state | Never run `-R` post-release |
| Art | Poor first impression | Full-set visual review |

## 14. Rollback

OpenBSD does not support downgrades. There is no safe `sysupgrade`
path from 8.0 back to 7.9.

Rollback options:

1. Restore the Phase 0 backup to new media.
2. Reinstall 7.9 from `install79.img`.
3. Keep the 7.9 machine available until 8.0 validation is complete.

Do not promise an in-place rollback in release notes.

## 15. Definition of Done

- All source references to 7.9 in runtime paths are removed or derived.
- `make test` passes.
- `make image` produces a bootable 8.0 image.
- Path A passes: clean 8.0 install plus setup yields a desktop.
- Path B passes: 7.9 upgrades to 8.0 and setup restores OpenRiot.
- A 7.9 system upgrades with `-n` and a controlled reboot.
- A fresh offline install completes.
- `install/openriot` is rebuilt and shipped with the release.
- Package pins resolve on a clean 8.0 host.
- Firmware pins resolve from the 8.0 tree.
- All 50 backgrounds and 42 lock screens are 8.0-branded WebP.
- Lock cache rebuild shows the new art.
- Release notes and README TOC are updated.
- No existing release notes, `VERSION`, or README badge were edited by
  the migration.

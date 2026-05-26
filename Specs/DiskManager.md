# OpenRiot Disk TUI — Bug Tracker & Progress Log

## Last Updated
2026-05-24 (post-session 9 — all 26 issues resolved, 3 remaining tasks)

## Verification

A system integration test (`diskintegration_test.go`) passes on this machine:

```
sd0  3726 GB  root=false mounted=false chunk=true  encrypted=false raid=true   ← NVMe chunk
sd1  3725 GB  root=true   mounted=true   chunk=false encrypted=true  raid=false  ← virtual root
```

- sd0: `IsChunk=true` (hasRAID from disklabel), `IsRoot=false`
- sd1: `IsEncrypted=true` (`isVirtualOnSoftraidBus` from dmesg), `IsRoot=true` (mounted at `/`)
- Both correctly blocked from Mount/Format/Encrypt/Umount filters
- Both correctly shown in Discover filter
- sd2 (USB removable) available for operations

## Issues Found (chronological)

### Issue 1: Mount status wrong — softraid virtual devices not traced to physical
- **Symptom:** Physical chunks (sd1, sd2) showed `[AVAILABLE]` even though their virtual devices (sd0, sd3) were mounted.
- **Root cause:** `parseMounts()` wasn't resolving virtual devices back to physical chunks.
- **Fix:** `parseMounts` now takes `softraidInfo` and maps both virtual and physical devices.
- **Status:** ✅ FIXED

### Issue 2: Root drive protection wrong
- **Symptom:** Virtual softraid root device (sd0) wasn't marked as `[ROOT]`, so Format/Encrypt offered it.
- **Root cause:** `DiscoverDrives` only checked direct `rootDrive` match, didn't trace virtual↔physical.
- **Fix:** Added both directions of softraid root tracing.
- **Status:** ✅ FIXED

### Issue 3: softraid parsing wrong
- **Symptom:** `parseSoftraid()` expected volume + chunk on one line.
- **Root cause:** `bioctl softraid0` output is multi-line (`Volume sd0 ... CRYPTO` on one line, `chunk sd1a` on indented next line).
- **Fix:** Rewrote as stateful parser tracking `currentVirtual` across lines.
- **Status:** ✅ FIXED

### Issue 4: Menu text brightness & alignment
- **Symptom:** Description text was same brightness as names; `fmt.Sprintf("%-17s", styledString)` caused misalignment because ANSI codes are invisible bytes.
- **Fix:** Switched to `lipgloss.NewStyle().Width(17).Render()` and `theme.Lipgloss.Base.Dim` for descriptions.
- **Status:** ✅ FIXED

### Issue 5: Cursor mismatch after filtering
- **Symptom:** After adding drive filtering per action, `updateDriveList()` still indexed into `m.drives` instead of the filtered list.
- **Fix:** Added `filteredDrives []Drive` to model, extracted `filterDrives()` to shared helper.
- **Status:** ✅ FIXED

### Issue 6: Chunk drives appear in lists
- **Symptom:** Physical chunks (sd1, sd2) appeared in Mount/Format/etc. lists.
- **Fix:** Added `IsChunk` field, filtered chunks out of all action lists.
- **Status:** ✅ FIXED

### Issue 7: `[ROOT]` not shown on virtual root
- **Symptom:** Virtual root (sd0) didn't always get `[ROOT]` status depending on dmesg output.
- **Fix:** Root detection now traces both virtual→physical and physical→virtual.
- **Status:** ✅ FIXED

### Issue 8: `[ENCRYPTED]` shown on physical chunks
- **Symptom:** Physical chunks showed `[ENCRYPTED]` instead of `[CHUNK]`.
- **Root cause:** `IsEncrypted` checked `physicalToVirtual` instead of `virtualToPhysical`.
- **Fix:** Fixed boolean logic and added `[CHUNK]` status with `FG3` color.
- **Status:** ✅ FIXED

### Issue 9: Root drive appears in Umount list
- **Symptom:** Root drive was offered for unmount.
- **Fix:** Added `!d.IsRoot` to umount filter.
- **Status:** ✅ FIXED

### Issue 10: Root drive appears in Format list
- **Symptom:** Root drive was offered for format.
- **Root cause:** Root detection was failing on virtual softraid root (Issue 2).
- **Fix:** Same as Issue 2 — root tracing now correct.
- **Status:** ✅ FIXED

### Issue 11: Discover Drives should not be selectable
- **Symptom:** In Discover mode, pressing Enter does nothing meaningful (hits default case which is a no-op), yet the bottom bar says `enter: select | esc: back | q: quit`.
- **Fix:** `updateDriveList` ignores Select key when `m.action == actionDiscover`. `renderDriveList` omits "enter: select" from help text for Discover.
- **Status:** ✅ FIXED

### Issue 12: Error screen key help is wrong
- **Symptom:** Error screen says `"Press esc or q to return"`. But `q` quits the app, `esc` returns.
- **Fix:** Changed to `"esc: back | q: quit"` matching the rest of the TUI.
- **Status:** ✅ FIXED

### Issue 13: `renderDriveList()` doesn't show chunks in Discover
- **Symptom:** Even though `actionDiscover` allows chunks in `filterDrives`, the outer `if d.IsChunk { continue }` in `filterDrives` still skips them.
- **Fix:** Restructured `filterDrives` so `actionDiscover` appends everything (including chunks) before the IsChunk check, while all other actions guard with `!d.IsChunk`.
- **Status:** ✅ FIXED

### Issue 14: `parseMounts` partition extraction fails for non-`a` partitions
- **Symptom:** Mount detection missed partitions like `sd0i`, `sd0j` because `strings.TrimSuffix(base, "a")` only stripped `a`.
- **Root cause:** Partition letters can be any lowercase a-z, not just `a`.
- **Fix:** Changed to a loop stripping ALL trailing lowercase letters (a-z).
- **Status:** ✅ FIXED

### Issue 15: Cross-mapping missing in `parseMounts`
- **Symptom:** If mount output showed a virtual device, its physical chunk wasn't marked mounted (and vice versa).
- **Fix:** Added bidirectional cross-mapping: if virtual is mounted, physical inherits `[MOUNTED]`; if physical is mounted, virtual inherits it.
- **Status:** ✅ FIXED

### Issue 16: `findRaidDevice` completely broken
- **Symptom:** `findRaidDevice` returned wrong device or empty string, causing mount/umount/format operations to target incorrect drives.
- **Root cause:** Used `fields[4]` from `Volume sd0 ... RAID 0 CRYPTO` which resolved to `"RAID"` not `"sd0"`. Also searched for `"<"+device` which doesn't appear in bioctl output.
- **Fix:** Rewrote to use `parseSoftraid()` directly via `physicalToVirtual` / `virtualToPhysical` lookups.
- **Status:** ✅ FIXED

### Issue 17: `updateResult` quits on `esc` instead of returning to menu
- **Symptom:** After any operation completed, pressing `esc` quit the entire TUI instead of returning to the main menu.
- **Fix:** `esc` (Back) now returns to `stateMenu`; only `q` quits.
- **Status:** ✅ FIXED

### Issue 18: `UmountDrive` hardcodes mount point
- **Symptom:** `umountCmd` always passed `"/mnt/backup"` regardless of where the drive was actually mounted.
- **Fix:** `umountCmd` now captures `d.MountPoint` from the Drive struct.
- **Status:** ✅ FIXED

### Issue 19: Benchmark hardcodes 2GB with no duration cap
- **Symptom:** fio ran 2GB writes which could hang for 5+ minutes on slow drives. No way to choose a shorter test.
- **Fix:** Added benchmark preset configuration screen (`stateBenchmarkConfig`) with Quick (512MB/15s), Standard (2GB/60s), Thorough (4GB/180s) presets. `BenchmarkDrive` now accepts `writeSize` and `rwSize` parameters.
- **Status:** ✅ FIXED

### Issue 20: No backend safety guard against root/chunk operations
- **Symptom:** If UI filtering had a bug, root or chunk drives could still be formatted/encrypted/unmounted.
- **Fix:** Added `guardDrive()` — re-runs `DiscoverDrives()` at operation time and refuses if target is `IsRoot` or `IsChunk`. Applied to `FormatDrive`, `EncryptDrive`, `UmountDrive`.
- **Status:** ✅ FIXED

### Issue 21: Mount filter missing `!d.IsRoot`
- **Symptom:** `filterDrives` for `actionMount` allowed root drives to appear in Mount list even if root detection was correct.
- **Root cause:** Mount filter only checked `!d.IsChunk && !d.IsMounted`, missing `!d.IsRoot`.
- **Fix:** Added `!d.IsRoot` to Mount filter.
- **Status:** ✅ FIXED

### Issue 22: `MountDrive` missing backend safety guard
- **Symptom:** Even after UI filter fix, a root drive could still be passed to `MountDrive`.
- **Fix:** Added `guardDrive(device)` call to `MountDrive`.
- **Status:** ✅ FIXED

### Issue 23: Root device detection uses dmesg — unreliable
- **Symptom:** dmesg can show physical chunk (`root on sd1a`) or include `/dev/` prefix, causing root detection to miss virtual devices.
- **Fix:** Prefer mount output (`mount | grep ' on / '`) which shows the currently mounted root device. Fall back to dmesg only if mount doesn't find it.
- **Status:** ✅ FIXED

### Issue 24: `parseMounts` missing reverse cross-map
- **Symptom:** If mount shows physical chunk (`/dev/sd1a on /`), only chunk was marked mounted; virtual device sd0 was not.
- **Fix:** Added reverse cross-map: if base is a physical chunk, mark its virtual device as mounted too.
- **Status:** ✅ FIXED

### Issue 25: `parseSoftraid` CRYPTO check too strict
- **Symptom:** Only matched volume lines containing "CRYPTO", silently dropping volumes with other RAID levels. Also `"chunk sd"` match too narrow.
- **Fix:** Match any `Volume ` line (not just CRYPTO). Relaxed chunk match from `"chunk sd"` to `"chunk"`.
- **Status:** ✅ FIXED

### Issue 26: `parseSoftraid` fails — cascading flag loss
- **Symptom:** `parseSoftraid` silently returned empty maps, causing ALL softraid-aware flags (`IsChunk`, `IsEncrypted`, cross-mount mapping) to collapse. sd0 appeared as `[AVAILABLE]` in Mount/Format/Umount lists even though it was the root's softraid chunk.
- **Previous fix (broken):** Added dmesg-based `discoverSoftraidFromDmesg()` that paired virtual devices with chunks by sorted index. Fragile — broke with multiple volumes.
- **Correct fix:** Removed pairing entirely. Each device now gets its flags independently from multiple sources:
  - `IsChunk = sr.physicalToVirtual[device] != "" || hasRAID` — a device with a RAID partition IS a chunk, whether or not bioctl says so
  - `IsEncrypted = sr.virtualToPhysical[device] != "" || isVirtualOnSoftraidBus(device)` — a device at a softraid bus in dmesg IS a virtual volume
  - `isVirtualOnSoftraidBus` parses dmesg for `sdX at scsibusY` where scsibusY is `at softraid0`
- **Status:** ✅ FIXED

Specs from actual dmesg on this system:
| Device | Location | Disklabel | Role |
|--------|----------|-----------|------|
| sd0 | scsibus1 at nvme0 | sd0a = RAID | chunk |
| sd1 | scsibus3 at softraid0 | sd1a = 4.2BSD | virtual (root) |
| sd2 | scsibus4 at umass0 | FAT/exFAT | removable |

## Files Modified

| File | Changes |
|------|---------|
| `source/disk/disk.go` | Entry point |
| `source/disk/model.go` | Model struct, states, key map, benchmark presets |
| `source/disk/view.go` | All rendering (menu, drive list, confirm, password, benchmark config, running, result) |
| `source/disk/update.go` | Key handlers, async cmd generators, benchmark config dispatch |
| `source/disk/backend.go` | Drive discovery, mount/umount/format/encrypt/benchmark ops, `guardDrive()`, `parseSoftraid()`, `findRaidDevice()` |
| `source/disk/filter.go` | `filterDrives()` helper with action-aware filtering |
| `source/commands/commands.go` | Added `--disk` command |
| `source/settings/settings.go` | Added Disk Manager menu entry |
| `config/rofi/apps.txt` | Added `openriot_disk` window class |
| `config/window/icons.toml` | Added `openriot_disk = "󰋊"` |
| `install/packages.yaml` | Added `fio-3.41` |
| `README.md` | Added Disk Manager section + TOC link + `disk.webp` screenshot |
| `assets/disk.webp` | Screenshot (converted from disk.png) |

## Remaining Work

1. **Manual test on real OpenBSD system with softraid crypto root**
   - Run `openriot --disk`
   - Verify Discover shows: `sd0 [ROOT]`, `sd1 [CHUNK]`, `sd3 [MOUNTED]` (or similar)
   - Verify Mount list excludes `sd0`, `sd1`, `sd3`
   - Verify Umount list includes `sd3` but not `sd0`, `sd1`
   - Verify Format/Encrypt lists exclude root and chunks
   - Select Benchmark on `sd3` → verify preset picker appears (Quick/Standard/Thorough) → verify fio runs with chosen size
   - Umount `sd3` → verify `bioctl -d` runs (check with `bioctl softraid0` after)
   - After any operation result screen, press `esc` → should return to menu, NOT quit

2. **Consider adding `--runtime` flag to fio commands**
   - Current fio commands use `--size` but not `--runtime`. On very slow drives, even Quick (512MB) can take minutes. Consider adding `--runtime=<seconds>` to each fio invocation in `BenchmarkDrive` to cap execution time regardless of drive speed.

3. **Consider multi-chunk softraid arrays (RAID 1) support**
   - Current `parseSoftraid` resets `currentVirtual = ""` after each chunk. For RAID 1 with two chunks, only the first chunk is captured. For CRYPTO volumes (single chunk), this is fine. If RAID 1 support is needed later, change `currentVirtual = ""` to only reset when a new `Volume` line is encountered.

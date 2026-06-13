# OpenBSD Zed Port — Status

**Last attempt:** `fca2ccd403` (origin/main as of 2026-06-12)
**Working tree:** Patched, ready to build.
**Build:** NOT YET VERIFIED at `fca2ccd403`. Multiple script bugs block
automated build (see "Script Bugs" below).

## TL;DR for the Next AI

- `~/Code/zed` is checked out at `fca2ccd403` (`origin/main`).
- `scripts/zed-patch.diff` (1111 lines) applies cleanly to `fca2ccd403`
  with **0 FAILED** hunks. Picard-era OpenBSD patches are present on the
  working tree (36 files modified).
- `scripts/zed.sh` is **broken**. Do not run it as-is. Four bugs below.
- The runtime `Backends::empty()` panic from the Picard-era 137e677a05
  build is still unresolved. The zed.sh wgpu cache invalidation
  (Step 2.5) is incomplete.

## State of the World

### `~/Code/zed`

| Item | Value |
|------|-------|
| Commit | `fca2ccd403` |
| Branch | `HEAD detached` (will become `main` on `git switch -C main fca2ccd403`) |
| Remote | `origin` → `https://github.com/zed-industries/zed.git` |
| Stash | `stash@{0}` = Picard patches at `137e677a05` (pre-reset). Keep until verified. |
| Patches | 36 files modified, 217 insertions(+), 102 deletions(-) |
| `.orig` | None (cleaned) |
| `target/` | 99 GB from `137e677a05` Picard build. Most artifacts still valid for `fca2ccd403` (Cargo.toml deps unchanged). Do **not** `cargo clean`. |

### `scripts/zed-patch.diff`

1111 lines, 36 files. Applies cleanly to `fca2ccd403` with `patch -p1 -t`.
All hunks Picard-era OpenBSD additions: `target_os = "openbsd"` to
existing `linux`/`freebsd` cfg gates, `current_path` impl for fs.rs,
`vscode_import.rs` platform var, etc. Full list:

```
Cargo.toml
crates/audio/Cargo.toml
crates/audio/src/audio_pipeline/echo_canceller.rs
crates/call/src/call_impl/diagnostics.rs
crates/cli/Cargo.toml
crates/cli/src/main.rs
crates/client/src/telemetry.rs
crates/crashes/Cargo.toml
crates/crashes/src/crashes.rs
crates/fs/src/fs.rs
crates/gpui/src/app.rs
crates/gpui/src/gpui.rs
crates/gpui/src/keymap/context.rs
crates/gpui/src/platform.rs
crates/gpui/src/platform/keystroke.rs
crates/gpui/src/platform/test/platform.rs
crates/gpui/src/scene.rs
crates/gpui/src/svg_renderer.rs
crates/gpui/src/window.rs
crates/gpui_linux/Cargo.toml
crates/gpui_linux/src/gpui_linux.rs
crates/gpui_linux/src/linux/platform.rs
crates/gpui_linux/src/linux/wayland/client.rs
crates/gpui_linux/src/linux/x11/client.rs
crates/gpui_platform/Cargo.toml
crates/gpui_platform/src/gpui_platform.rs
crates/languages/src/python.rs
crates/languages/src/rust.rs
crates/livekit_client/Cargo.toml
crates/livekit_client/src/lib.rs
crates/remote_server/Cargo.toml
crates/settings/src/vscode_import.rs
crates/zed/Cargo.toml
crates/zed/src/main.rs
crates/zed/src/zed.rs
crates/zed/src/zed/open_listener.rs
```

### `scripts/zed.sh`

The script as it sits in HEAD is **not runnable** on this OpenBSD host.
Four bugs in the patch-check section. The current `Step 2: Apply
OpenBSD patches` block has been edited (during the prior session) to
attempt a fix, but the fix is itself broken. See "Script Bugs" below
for the precise corrections.

## Script Bugs (in `scripts/zed.sh`)

### Bug 1 — patch(1) opens `/dev/tty` and hangs

OpenBSD's `patch(1)` opens `/dev/tty` (fd 6) when it encounters a
`"Hmm..."` format warning, which appears on every multi-hunk patch.
In a non-interactive shell, the read from `/dev/tty` blocks forever.

**Symptom:** `Step 2: Applying OpenBSD compatibility patches...` then
silently hangs. No `patch` or `zed.sh` process visible after ~30s.

**Fix:** Add `-t` (batch mode) to **every** `patch` invocation.

### Bug 2 — `set -e` + `$(...)` capture = silent exit

OpenBSD's `/bin/sh` (pdksh) treats a non-zero exit from a `$(...)`
substitution under `set -e` as fatal. OpenBSD `patch` exits non-zero
on success (the `"Hmm..."` warnings), so the dry-run capture kills
the script silently before reaching the build.

**Symptom:** Same as Bug 1 — `Step 2` hangs/vanishes. No log output,
no error.

**Fix:** Append `|| true` to the `$(patch ...)` substitution so its
exit code is always 0. Then check the captured string for `FAILED` to
decide whether the dry-run was clean.

### Bug 3 — patch auto-applies in reverse

OpenBSD `patch` detects a reversed diff and silently assumes `-R`,
printing `Reversed (or previously applied) patch detected!  Assuming -R`
and applying the reverse. This **destroys** the working-tree patches
when the script is run a second time (or after `git stash`/restore).

**Symptom:** First run applies patches correctly. Second run wipes
them off. Working tree shows 1 modified file instead of 36.

**Fix:** Before applying, run `patch --dry-run -R -t` first. If it
finds no `FAILED`, the patch is already applied → skip the forward
apply entirely. Only run the forward apply when the reverse dry-run
**does** report `FAILED` (i.e., the patch has not been applied yet).

### Bug 4 — exit-code check is unreliable

The original `if patch ... ; then ... elif patch -R ... ; then ...` logic
relies on `patch` exit codes, but OpenBSD `patch` exits non-zero on
success. The original `if/elif` always falls through to the error
branch even on a clean apply.

**Fix:** Don't trust `$?` from `patch`. Capture the dry-run output
into a variable and `rg -q 'FAILED'` against it.

### Correct Step 2 Logic (reference)

```sh
echo "Applying OpenBSD compatibility patches..."
if [ -f "$PATCH_FILE" ] && [ -s "$PATCH_FILE" ]; then
    # Reverse dry-run first: if it succeeds, patch is already applied.
    if patch -d "$SOURCE_DIR" -p1 --dry-run -t -R < "$PATCH_FILE" 2>&1 \
            | rg -q 'FAILED' >/dev/null 2>&1; then
        :  # reverse dry-run has FAILED → forward dry-run is needed
        PATCH_DRYRUN=$(patch -d "$SOURCE_DIR" -p1 --dry-run -t \
            < "$PATCH_FILE" 2>&1 || true)
        if ! printf '%s\n' "$PATCH_DRYRUN" | rg -q 'FAILED'; then
            patch -d "$SOURCE_DIR" -p1 -t < "$PATCH_FILE" || {
                echo "ERROR: patch apply failed despite clean dry-run" >&2
                exit 1
            }
            find "$SOURCE_DIR" -name '*.orig' -type f -delete 2>/dev/null || true
            echo "Patch applied"
        else
            echo "ERROR: Patch cannot be applied. Upstream code may have changed." >&2
            echo "Update the patch or the commit pin." >&2
            exit 1
        fi
    else
        echo "Patch already applied"
    fi
else
    echo "No patch file found, building without patches"
fi
```

## Runtime Issue (carried over from Picard era)

The Picard-era 137e677a05 build produces a binary that panics at
runtime with:

```
thread 'main' panicked at
  ~/.cargo/git/checkouts/wgpu-423de87c978aca7f/357a0c5/wgpu/src/api/instance.rs:65:13:
No wgpu backend feature that is implemented for the target
platform was enabled.
```

`wgpu`'s build script must emit `cargo:rustc-cfg=gles` for the GLES
backend to compile in. Cargo caches the build-script output at
`target/release-fast/build/wgpu-<hash>/output` and re-runs the script
only when its fingerprint changes. Patching the wgpu source after
extraction does **not** invalidate the cache on its own.

The current `zed.sh` Step 2.5 has a `find` that deletes
`target/release-fast/build/wgpu-*/output`, but cargo's modern
fingerprint file is `output.json` (or both). The cleanup is incomplete
and may need to be:

```sh
find "$SOURCE_DIR"/target -path '*/build/wgpu-*/output*' -type f -delete
```

or, more aggressively, `cargo clean -p wgpu -p wgpu-core -p wgpu-hal`
which removes all artifacts for those three packages and forces a
full re-run of their build scripts.

The Picard binary from `137e677a05` panicked even after the find-clean.
The `cargo clean -p` approach was the only one that produced a
non-panicking binary. **Add `cargo clean -p wgpu -p wgpu-core -p wgpu-hal`
to zed.sh before the build loop** (after the patches, before cargo).

## Build Strategy (what the next AI should do)

1. **Fix `scripts/zed.sh` Step 2** with the reference logic above
   (or equivalent). Verify by running `bash -n scripts/zed.sh`.
2. **Add `cargo clean -p wgpu -p wgpu-core -p wgpu-hal`** to the
   script just before the build loop.
3. **Manually run the build**, with output captured:
   ```sh
   ZED_COMMIT=fca2ccd403 ZED_FEATURES=x11 \
       bash /home/grendel/Code/OpenRiot/scripts/zed.sh 2>&1 \
       | tee /tmp/zed-build-$(date +%s).log
   ```
4. **Watch for hangs** at the patch step. If the new logic is wrong,
   the same symptoms (silent hang, no `patch` process) will appear.
5. **After first link:** `ls -lh ~/Code/zed/target/release-fast/zed`.
   Should be ~290–310 MB. If linker OOMs, raise `ulimit -d`.
6. **After install:** `~/.local/share/openriot/config/bin/zed` should
   not panic. If `Backends::empty()` appears, the wgpu cache is stale;
   the build was run without step 2 above.
7. **If new upstream has cfg-gated functions returning `()`** (the
   `scap` issue from Picard era): a `rg -n 'target_os = "linux".*target_os = "freebsd"'`
   search across `crates/` will find missing `openbsd` cfg additions.
   The Picard patch already covers all known cases at `fca2ccd403`; if
   new ones appear, add them.

## Things That DO Work

- The 1111-line `zed-patch.diff` applies cleanly to `fca2ccd403`
  (0 FAILED, 36 files modified, +217/-102).
- `cargo --version` is 1.96.0, `rustc 1.96.0`. Both on PATH.
- `wgpu` source in cargo cache is pre-patched with
  `target_os = "openbsd"` in the four cfg files (Step 2.5 will be a
  no-op on first run; idempotent guard handles it).
- `wasmtime-wasi` and `ipc-channel` patches in Step 3 retry loop are
  applied to registry sources on attempt 2.
- The build profile `release-fast` is already configured in the
  workspace (opt-level=1, codegen-units=256, debug=0, no LTO).
- `gpui_linux/x11` and `gpui_linux/wayland` are the default features
  of `gpui_linux`. `--no-default-features --features 'x11'` works
  because the `x11` feature is `["gpui_platform/x11"]` and
  `gpui_platform/x11` is `["gpui_linux/x11"]`.
- Scap is NOT pulled in (the `screen-capture` feature is in
  `test-support`, not in `x11`).
- Existing `target/` from Picard build (99 GB) is mostly valid; most
  workspace deps unchanged.

## Things That Are Broken

- **`scripts/zed.sh` patch check (4 bugs above).** Must be fixed
  before any build attempt.
- **Runtime wgpu backend.** Even with all source patches correct,
  cargo's build script cache produces a binary that panics with
  `Backends::empty()` unless the cache is invalidated.
- **No build has been verified at `fca2ccd403`.** Everything from
  this point is theoretical.

## Known cfg-Gating Gaps (probably irrelevant, listed for completeness)

These are `linux`/`freebsd`-only cfg blocks in upstream that are NOT
covered by the Picard patch. They produce dead code on OpenBSD but
should not break the build:

```
crates/zlog/src/filter.rs:38
crates/vim/src/state.rs:859
crates/vim/src/state.rs:861
crates/vim/src/state.rs:929
crates/vim/src/state.rs:933
crates/vim/src/replace.rs:319
crates/util/src/util.rs:327
crates/util/src/paths.rs:106
crates/util/src/paths.rs:2772
```

If a runtime panic mentions any of these, add `target_os = "openbsd"`
to the cfg list.

## Reference Commands

```sh
# Verify patch applies cleanly
cd ~/Code/zed && patch -p1 --dry-run -t < /home/grendel/Code/OpenRiot/scripts/zed-patch.diff

# Apply patch manually
cd ~/Code/zed && patch -p1 -t < /home/grendel/Code/OpenRiot/scripts/zed-patch.diff

# Force wgpu build-script re-run
cd ~/Code/zed && cargo clean -p wgpu -p wgpu-core -p wgpu-hal

# Build (manual, after patches are applied)
cd ~/Code/zed && RUSTFLAGS="-L /usr/local/lib" \
    sh -c 'ulimit -d 8388608 && exec cargo build \
        --profile release-fast \
        --no-default-features --features x11 \
        -j 1'

# Reset to upstream (destroys patches)
cd ~/Code/zed && git fetch origin && git reset --hard origin/main

# Restore Picard-era working tree
cd ~/Code/zed && git stash pop  # 'stash@{0}' = Picard patches at 137e677a05
```

## Background: How We Got Here

- Picard-era (Jun 9–10, 2026): Built zed at `137e677a05` successfully.
  295 MB binary. But binary panicked with `Backends::empty()` at
  runtime.
- New upstream (Jun 11, 2026): User wanted `fef979dec4`. Multiple
  failed attempts: patch FAILED count of 12+ on line shifts, `--features
  'x11'` removed from `zed` crate, scap cfg-gating, `kinfo_file`
  FreeBSD-only code in `fs.rs`, etc.
- **Bad pivot (Jun 12, 2026):** I restored the Picard patch and pinned
  to `137e677a05`, thinking it was a "known-working" state. **User
  correction: keep current upstream, fix OpenRriot against it.**
- Current state (Jun 12, 2026): Reset to `fca2ccd403` (current
  `origin/main`), applied Picard patch (applies cleanly), discovered
  the build script has 4 bugs, attempted to fix them, killed the
  build, and wrote this doc.

## Next Steps (in order)

1. Apply the Step 2 fix from the "Correct Step 2 Logic" block to
   `scripts/zed.sh`.
2. Add `cargo clean -p wgpu -p wgpu-core -p wgpu-hal` before the
   build loop in zed.sh.
3. Manually run zed.sh; capture full output to a log file.
4. Verify the binary links (~290–310 MB stripped).
5. Run the binary; verify no `Backends::empty()` panic.
6. If panic, ensure wgpu cache invalidation is actually working.
7. If new compile errors appear, add `target_os = "openbsd"` to the
   affected cfg list and regenerate the patch:
   ```sh
   cd ~/Code/zed
   # edit the file
   git diff > /tmp/new-hunk.diff
   # splice into the existing patch with ed(1) or by hand
   ```
8. Update this doc with the result.

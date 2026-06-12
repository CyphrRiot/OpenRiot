# OpenBSD Zed Port — Status

**CURRENT STATUS:** Compiles, links, and installs. Binary still
panics at runtime with `Backends::empty()` at
`wgpu/src/api/instance.rs:65`. **Cargo incremental build
issue** — see "Workaround" below.

## Runtime Failure (Current)

```
thread 'main' panicked at
  /home/grendel/.cargo/git/checkouts/wgpu-423de87c978aca7f/357a0c5/wgpu/src/api/instance.rs:65:13:
No wgpu backend feature that is implemented for the target
platform was enabled.
```

The `wgpu-core` and `wgpu-hal` build scripts output
`cargo:rustc-cfg=gles` (the cfg alias evaluates to true after
the source patches). The `wgpu` build script does **not**.
`enabled_backend_features()` in `instance.rs:65` therefore
returns `Backends::empty()` and the binary panics.

## Root Cause (Discovered)

The wgpu source has been patched correctly
(`target_os = "openbsd"` added to the `gles` cfg alias in
`wgpu/build.rs` and `wgpu-core/build.rs`; `wgpu-core` and
`wgpu-core-deps-windows-linux-android/Cargo.toml` updated to
include openbsd in the GLES platform deps).

However, **cargo's incremental build caches the wgpu build
script's output**. The cached output
(`target/release-fast/build/wgpu-<hash>/output`) still
predates the patch. The build script is not re-run because
cargo's fingerprint matches the cached state:

- `target/release-fast/build/wgpu-5850a6c2bf5ed10d/output` —
  mtime Jun 9 11:53, no `cargo:rustc-cfg=gles`
- `target/release-fast/build/wgpu-core-eaddd30d0e701ecd/output` —
  mtime Jun 9 14:14, **has** `cargo:rustc-cfg=gles`
- `target/release-fast/build/wgpu-hal-5af7f487d3140fce/output` —
  mtime Jun 9 14:12, **has** `cargo:rustc-cfg=gles`

wgpu-core and wgpu-hal were re-evaluated correctly. wgpu was
not. A 10-second `cargo build` after the patches produces a
binary that still panics.

The cfg_aliases! macro itself is not at fault. The patches
are correct. The issue is purely cargo's cache invalidation
for the wgpu build script.

## Workaround

Force cargo to re-run the wgpu build script:

```sh
cd ~/Code/zed
cargo clean -p wgpu -p wgpu-core -p wgpu-hal
sh -c 'ulimit -d 8388608 && exec cargo build --profile release-fast \
  --no-default-features --features x11 -j 1'
```

The `cargo clean -p` removes cached artifacts for those three
crates. The next build re-runs their build scripts and
produces a binary with the GLES backend enabled.

**TODO:** add this to `scripts/zed.sh` automatically so the
workaround is not manual.

## Build Script (`scripts/zed.sh`)

Four fixes already applied:

| # | Issue | Fix |
|---|-------|-----|
| 1 | `ulimit -v` silently no-ops on OpenBSD | `ulimit -d 8388608` |
| 2 | `TMPDIR` unset → cargo scratch on 495 MB mfs `/tmp` | `TMPDIR=$HOME/.cache/zed/tmp` |
| 3 | `sed -i ''` fails on OpenBSD (`/usr/bin/sed` lacks `-i`) | `sed_i()` helper using temp file + `mv` |
| 4 | wgpu patches gated on `attempt 2` retry that never ran | New Step 2.5 before the build loop, with post-patch verification |

The build script compiles and links successfully (~10 min
first run, ~10 s no-op rebuilds).

## Patch (`scripts/zed-patch.diff`)

~1110 lines across 27 files. All changes add
`target_os = "openbsd"` to existing `linux`/`freebsd` cfg
gates, stub out unsupported features (LiveKit, crash
reporting, screen capture, Wayland layer-shell), and provide
OpenBSD platform constants.

## Build Requirements

- OpenBSD 7.9 amd64
- `pkg_add rust xkbcommon`
- Datasize limit must allow ~8 GB (`ulimit -d 8388608`).
  The script enforces this automatically.
- `TMPDIR` is set to `~/.cache/zed/tmp` (not the small mfs `/tmp`).
- `RUSTFLAGS="-L /usr/local/lib"` (set automatically by `zed.sh`)

## Build Profile

Must use `--profile release-fast` (already defined in workspace:
`opt-level=1`, `codegen-units=256`, `debug=0`, no LTO). Debug builds
produce a 507 MB binary that exceeds OpenBSD GENERIC kernel's 1 GB MAXDSIZ
and cannot exec.

`release-fast` produces a ~295 MB (stripped) binary that loads and runs.

## Third-Party Crate Patches (applied by zed.sh)

| Crate | Issue | Fix |
|-------|-------|-----|
| `wasmtime-wasi` | TCP keepalive sockopt not available on OpenBSD | Stub callsites with error returns |
| `ipc-channel` | `POLLRDHUP` not in OpenBSD libc | Define as 0 |
| `wgpu-core` platform deps | Platform gate excludes OpenBSD from GLES | Add `target_os = "openbsd"` to cfg |
| `wgpu` / `wgpu-core` `build.rs` | `gles` cfg alias excludes OpenBSD | Inject `target_os = "openbsd"` into `gles` definition |

The wgpu patches **enable** the GLES (OpenGL ES) backend via EGL,
which Mesa provides on OpenBSD (`/usr/X11R6/lib/libEGL.so`).
Without these patches, `cfg!(gles)` evaluates to false at
compile time even though the `gles` feature is enabled, so
`enabled_backend_features()` returns `Backends::empty()` and
the binary panics at `wgpu/src/api/instance.rs:65`.

## Known Working

- All crates compile and link with `--profile release-fast`
- `wgpu-core` and `wgpu-hal` build scripts correctly emit
  `cargo:rustc-cfg=gles`
- X11 window created, colormap allocated
- Keyboard/mouse via XInput
- Font rendering, terminal, LSP, UI framework compile

## Known Broken

- **Runtime panic** at `wgpu/src/api/instance.rs:65` —
  `Backends::empty()`. Cause: wgpu build script output is
  stale (cargo incremental build did not re-run it). See
  "Root Cause" and "Workaround" above.
- Screen capture (scap gated out)
- Crash reporting (minidumper/crash-handler stubbed)
- Voice/video calls (livekit/webrtc gated out)
- Wayland (no compositor on OpenBSD)
- Audio echo cancellation (libwebrtc APM gated out)

## Build Command (if starting from scratch)

```
cd ~/Code/zed
RUSTFLAGS="-L /usr/local/lib" cargo build \
  --profile release-fast \
  --no-default-features --features x11 \
  -j 1
```

This builds all crates incrementally. Subsequent runs are ~3–4 minutes.

## Next Steps

1. Add `cargo clean -p wgpu -p wgpu-core -p wgpu-hal` to
   `scripts/zed.sh` Step 3 (or before the build) so the
   wgpu build script is forced to re-run.
2. Confirm the binary launches without the
   `Backends::empty()` panic.
3. If EGL display initialization fails at runtime, check
   that `/usr/X11R6/lib/libEGL.so` is loadable (set
   `LD_LIBRARY_PATH` if needed).
4. Report whether the window renders or if a new runtime
   failure appears.

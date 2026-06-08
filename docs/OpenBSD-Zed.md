# OpenBSD Zed Port — Status

**CURRENT STATUS:** Compiles and links with `--profile release-fast`. Binary
runs `--help`. Runtime fix pending — the `wgpu` GLES backend cfg was missing
`target_os = "openbsd"` in `build.rs`. Patches for this are now applied by
`scripts/zed.sh`.

## Patch

`scripts/zed-patch.diff` — ~1110 lines across 27 files. All changes add
`target_os = "openbsd"` to existing `linux`/`freebsd` cfg gates, stub out
unsupported features (LiveKit, crash reporting, screen capture, Wayland
layer-shell), and provide OpenBSD platform constants.

`scripts/zed.sh` — Build script. Applies patch, patches third-party
crates in the cargo cache at build time, then builds.

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

The wgpu patches enable the GLES (OpenGL ES) backend via EGL, which Mesa
provides on OpenBSD (`/usr/X11R6/lib/libEGL.so`). Without these patches,
`cfg!(gles)` evaluates to false at compile time even though the `gles`
feature is enabled, so `enabled_backend_features()` returns
`Backends::empty()` and the binary panics at `wgpu/src/api/instance.rs:65`.

## Runtime Status

The binary starts, initializes X11, creates a window. With the `build.rs`
patches, the GLES backend is now compiled in and `enabled_backend_features()`
should include `Backends::GL`.

**Root cause (fixed):** `wgpu/build.rs` and `wgpu-core/build.rs` defined the
`gles` backend alias as
`any(windows, target_os = "linux", target_os = "android", target_os = "freebsd")`,
so OpenBSD was never considered a GLES-capable platform. The crate compiled
(because `wgpu-hal/Cargo.toml` gates `khronos-egl` on `unix`, which includes
OpenBSD), but the compile-time cfg was false.

## Next Steps

1. Re-run `scripts/zed.sh` on a checkout with the current patches.
2. Verify the binary no longer panics at `wgpu/src/api/instance.rs`.
3. If EGL display initialization fails at runtime, check that
   `/usr/X11R6/lib/libEGL.so` is loadable (set `LD_LIBRARY_PATH` if needed).
4. Report whether the window renders or if a new runtime failure appears.

## Known Working

- All crates compile and link with `--profile release-fast`
- X11 window created, colormap allocated
- Keyboard/mouse via XInput
- Font rendering, terminal, LSP, UI framework compile

## Known Broken

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

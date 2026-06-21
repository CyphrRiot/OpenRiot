# OpenBSD Zed Port — Status

**Last attempt:** `fca2ccd403` (origin/main, pinned)  
**Date:** 2026-06-20  
**Working tree:** Patched, compiles, installs.  
**Build:** 410 MB unstripped, `release-fast` profile, GLES backend.  
**Runtime:** All crash sources resolved (Rounds 1-9). Zed launches, window
renders, GPU works, extension host loads. UI glitches remain (missing
title bar on i3, stray "p" character, broken file dialog). Awaiting
rebuild with latest fixes.

## TL;DR for the Next AI

- `~/Code/zed` checked out at `fca2ccd403`. `scripts/zed-patch.diff`
  applies cleanly. `scripts/zed.sh` builds+installs.
- Full rebuild ~90 min; incremental ~1 min when only zed workspace crates
  change; ~60 min when wasmtime registry crates change.
- All crash sources **RESOLVED**:
  - Rounds 1-5: retguard SIGILL and IBT endbr64 in hand-written assembly
  - Round 6-7: extension-load SIGILL via `-z nobtcfi` linker flag
  - Round 8: MAP_STACK SIGSEGV (four patches: fiber, continuation,
    sigaltstack, stacker)
  - Round 9: PROT_NONE + MAP_STACK EINVAL in sigaltstack allocation
- **Script flags**: `CFLAGS="-fno-ret-protector"`; `RUSTFLAGS="-C
  link-arg=-Wl,-z,nobtcfi -L /usr/local/lib ..."`.
- **Critical bug fixed**: all `find` commands used `-maxdepth 1` which
  never found crate dirs at depth 2 inside `index.crates.io-*`. Now
  recursive (no `-maxdepth`), matching the working `aws-lc-sys` pattern.
- **Cargo tarball extraction trap**: registry patches modify `src/` but
  cargo extracts build scripts from original `.crate` tarballs in
  `cache/`. The `aws-lc-sys` cc_builder fix patches both `src/` AND any
  extracted `target/build/` copies.
- **Zed-patch.diff additions** (Round 10):
  - `gpui_linux/.../x11/client.rs`: force `client_side_decorations_supported
    = true` on OpenBSD (i3 has no compositor, needs CSD title bar)
  - `gpui/src/platform.rs`: `PlatformScreenCaptureFrame = ()` on
    OpenBSD (`scap` crate is Linux/Windows only)
  - `crates/zed/Cargo.toml`: `x11` feature re-export
- **Status**: all fixes staged in `scripts/zed.sh` and
  `scripts/zed-patch.diff`. Awaiting rebuild with aws-lc-sys build-dir
  patch. After rebuild: test CSD title bar, file dialog, and investigate
  stray "p" character.

## Crash Progression

| Round | Result | Crash Location | Root Cause |
|-------|--------|----------------|------------|
| 1 | ❌ | `migrate_settings` | tree-sitter-json retguard |
| 2 | ❌ | X11 client init | More retguard in other -sys crates |
| 3 | ❌ | Window creation | Missing freetype-sys, libsqlite3-sys, etc. |
| 4 | ❌ | After GPU + window | psm/ring assembly lacks endbr64 |
| 5 | ❌ | tree_sitter_json | `-fcf-protection=none` broke C code endbr64 |
| 6 | ✅ | Startup complete | `-fno-ret-protector` + IBT asm patches |
| 7 | ✅ | Extension load | `-z nobtcfi` (PT_OPENBSD_NOBTCFI) |
| 8 | ✅ | Post-startup SIGSEGV | Fiber/continuation/signal stacks lack MAP_STACK |
| 9 | ✅ | `allocate_sigaltstack` EINVAL | PROT_NONE + MAP_STACK rejected by OpenBSD |
| 10 | 🔄 | UI: missing title bar | i3 has no compositor, CSD disabled |

## CPU Protection Issues — Resolved

### 1. Retguard (`-ret-protector`) — **DISABLED**

Emits `__retguard_XXXX` assembly prologues into every C function. On PIC
code, the canary address can land on an unmapped page — SIGILL before
`main()` even starts.

**Fix:** `export CFLAGS="-fno-ret-protector"`

### 2. IBT / Branch Tracking (`-fcf-protection=branch`) — **KEPT ENABLED**

Requires `endbr64` at every indirect call target. Intel Tiger Lake (11th
gen) enforces this in hardware. Hand-written assembly in `psm` and `ring`
crates lacks `endbr64`; patched manually.

### 3. Linker IBT opt-out (`-z nobtcfi`) — **ENABLED**

rustc stable has no `+ibt` target-feature. wasmtime Cranelift JIT makes
indirect calls into Rust hostcalls that lack `endbr64`. Opt the binary
out of IBT enforcement at link time via `PT_OPENBSD_NOBTCFI`.

Complete flags:
```sh
export CFLAGS="-fno-ret-protector"
export CXXFLAGS="-fno-ret-protector"
export RUSTFLAGS="-C link-arg=-Wl,-z,nobtcfi -L /usr/local/lib ${RUSTFLAGS:-}"
```

## MAP_STACK Enforcement — Resolved

OpenBSD enforces `MAP_STACK` on every mmap'd region used as a stack.
Without it, the kernel delivers SIGSEGV. Four allocation sites needed
fixing:

| Crate | File | Allocation |
|-------|------|-------------|
| `wasmtime-internal-fiber` | `src/unix.rs:184` | Fiber execution stacks |
| `wasmtime` | `.../stack_switching/stack/unix.rs:101` | Continuation stacks |
| `wasmtime` | `.../sys/unix/signals.rs:517` | sigaltstack |
| `stacker` | `src/mmap_stack_restore_guard.rs:31` | Recursion guard stacks |

All four patched in `scripts/zed.sh` via `sed`/`ed`. Each patch is
independently idempotent with its own cache invalidation.

### PROT_NONE + MAP_STACK EINVAL

OpenBSD rejects `PROT_NONE | MAP_STACK` with EINVAL. The sigaltstack
code allocated the region with PROT_NONE then mprotected the usable
portion. Fix: allocate with PROT_READ|PROT_WRITE + MAP_STACK, then
mprotect the guard page to PROT_NONE.

## UI / Rendering Issues

### Missing title bar (Round 10)

i3 has no compositor (`compositor_present = false`). Zed checks this and
disables client-side decorations (`window.rs:1750-1752`), expecting the
WM to draw title bars. i3 only draws thin pixel borders — no title bar.

**Fix:** `gpui_linux/.../x11/client.rs:377` —
`client_side_decorations_supported` is now always `true` on OpenBSD,
forcing zed to draw its own title bar regardless of compositor status.

### Stray "p" character in editor

A "p" appears in the empty editor buffer at startup. Likely a key event
leaking through X11 input handling during initialization. Needs
investigation — may be related to the "activate not implemented on Linux"
log message or XIM input method initialization.

### File dialog broken

Cannot open folders or files properly. May be related to the missing
title bar (dialog windows inherit decoration behavior) or to fd
exhaustion. The wrapper now sets `ulimit -n 1024`.

### Cursor icons missing

Several X11 cursor names not found in OpenBSD's cursor theme:
`col-resize`, `sb_h_double_arrow`, `text`, `xterm`. Non-fatal — falls
back to `left_ptr`.

### Wasm extension loading

The `html` extension fails to load with EINVAL (os error 22). Likely the
same PROT_NONE + MAP_STACK issue in wasmtime's instance instantiation.
Check `wasmtime/src/runtime/vm/sys/unix/mmap.rs` for similar mmap
patterns.

### D-Bus error

Non-fatal. OpenBSD has no D-Bus session bus. Zed gracefully degrades.

## Known Issues

| Issue | Status | Notes |
|-------|--------|-------|
| SIGILL crashes | ✅ Resolved | Rounds 1-7 |
| MAP_STACK SIGSEGV | ✅ Resolved | Round 8 |
| PROT_NONE EINVAL | ✅ Resolved | Round 9 |
| Missing title bar | 🔄 Round 10 | CSD forced; awaiting rebuild |
| "p" character | 🔄 Pending | Key event leak at startup |
| File dialog broken | 🔄 Pending | Likely CSD/decorations related |
| Cursor icons missing | ℹ️ Cosmetic | Falls back to left_ptr |
| Wasm extension load | ❌ | EINVAL, likely same PROT_NONE |
| D-Bus | ℹ️ Non-fatal | No D-Bus on OpenBSD |
| `os-release` missing | ℹ️ Non-fatal | OpenBSD has no os-release |
| Unstripped binary | ℹ️ 410 MB | Debug symbols |
| `datasize` limit | ⚠️ | `ulimit -d 8388608` in wrapper |
| FD exhaustion | ⚠️ | `ulimit -n 1024` in wrapper |

## Build Cache Invalidation

Cargo does not detect source changes in registry crates (`~/.cargo/
registry/src/`). Direct deletion of `target/.../build/` and
`target/.../fingerprint/` directories is required after patching.
CFLAGS/RUSTFLAGS changes also require cache invalidation — tracked via
a sentinel file (`target/release-fast/.cflags`).

**Critical trap:** Cargo extracts build scripts from original `.crate`
tarballs in `~/.cargo/registry/cache/`, NOT from the patched `src/`
directory. The `aws-lc-sys` fix patches both locations.

## Background Timeline

- **Jun 9-10:** Initial build at `137e677a05`. `Backends::empty()` panic.
- **Jun 12-16:** Cherry-picked to `fca2ccd403`. Fixed 4 build bugs, wgpu
  GLES cfg. GPU detected.
- **Jun 16:** SIGILL crash (tree_sitter_json). Added retguard fix.
- **Jun 17-18:** Discovered IBT + retguard interaction. Tried
  `-fcf-protection=none` — broke C code. Settled on keeping IBT enabled
  and patching hand-written assembly.
- **Jun 19 (Round 6-7):** Fixed IBT asm patches + `-z nobtcfi` linker
  flag. Window renders, GPU works, extensions load.
- **Jun 19 (Round 8):** MAP_STACK SIGSEGV after first frame. Four
  patches. `find -maxdepth 1` bug caused three wasted 78-min rebuilds.
  Fixed: all `find` commands now recursive.
- **Jun 20 (Round 9):** Sigaltstack panics with EINVAL. PROT_NONE +
  MAP_STACK rejected. Fixed with RW allocation + guard page mprotect.
  Zed launches successfully. UI issues discovered.
- **Jun 20 (Round 10):** Title bar missing on i3. Forced CSD via
  `client_side_decorations_supported = true`. `PlatformScreenCaptureFrame`
  compile error: `scap` crate is Linux/Windows only — OpenBSD uses `()`
  fallback. aws-lc-sys panic moved from `cmake_builder.rs` to
  `cc_builder.rs`; added second patch targeting extracted build
  directory (cargo extracts from tarball, not patched src/).

## Next Steps

1. **Rebuild:** Run `./scripts/zed.sh`. The aws-lc-sys build-dir patch
   will apply on retry, cargo will continue the incremental rebuild.
   Expected: clean compile, zed launches with CSD title bar.

2. **Test title bar:** Verify the window has a proper title bar with
   menus, close/minimize/maximize buttons. Try opening a file.

3. **Investigate "p" character:** Check if it's a key event leak.
   Possible causes: XIM input method, clipboard initialization, or
   focus handling during startup.

4. **If file dialog still broken:** Increase `ulimit -n` further or
   check if it's a CSD window-decorations issue propagating to dialog
   windows.

5. **After UI stability:** Test LSP, terminal, git integration.
# OpenBSD Zed Port — Status

**Last attempt:** `fca2ccd403` (origin/main, pinned)
**Date:** 2026-06-18 (UPDATED: rebuild optimization + IBT fix)
**Working tree:** Patched, compiles, installs.
**Build:** 410 MB unstripped, `release-fast` profile, GLES backend.
**Runtime:** Crashes with SIGILL in `psm` stack assembly after GPU detection.

## TL;DR for the Next AI

- `~/Code/zed` is checked out at `fca2ccd403` (`origin/main`).
- `scripts/zed-patch.diff` applies cleanly (36 files, 0 FAILED).
- `scripts/zed.sh` builds+installs. Full rebuild ~90 min, incremental
  ~10s (when CFLAGS sentinel matches).
- The runtime `Backends::empty()` panic is **FIXED**. GLES backend
  compiles, GPU detected (`Mesa Intel Iris Xe Graphics`).
- The SIGILL crash has **TWO causes** (both must be disabled):
  1. **`-ret-protector`**: emits `__retguard` canaries (fixed by `-fno-ret-protector`)
  2. **`-fcf-protection=branch`**: requires `endbr64` at indirect call targets
     (hand-written asm like `psm` lacks it, fixed by `-fcf-protection=none`)
- **Script now sets BOTH flags**: `-fno-ret-protector -fcf-protection=none`
- **Status**: About to rebuild and test with both flags disabled.

## The Problem: Two Separate CPU Protections

OpenBSD's clang 19.1.7 defaults enable **two** independent CPU security features
that cause SIGILL when violated. Both must be disabled to run Zed.

### 1. Retguard (`-ret-protector`)

Emits `__retguard_XXXX` assembly prologues into every C function. At entry:

```asm
movq   0x1ca3309(%rip), %r11  ; __retguard_3130
xorq   (%rsp), %r11           ; XOR return addr with canary
```

On PIC code, the canary address can land on an unmapped page — SIGILL before
`main()` even starts. Fixed by `-fno-ret-protector`.

### 2. IBT / Branch Tracking (`-fcf-protection=branch`)

Requires `endbr64` instruction at every indirect call target. Intel Tiger Lake
(11th gen) enforces this in hardware. Hand-written assembly files (e.g.,
`psm`'s `x86_64.s`) don't include `endbr64`, so indirect calls to them SIGILL.

**Proof this is the issue:**

Default clang emits `endbr64`:
```bash
$ echo 'int foo(){return 1;}' | cc -S -o - -x c - | rg endbr64
    endbr64
```

With `-fcf-protection=none`, no `endbr64`:
```bash
$ echo 'int foo(){return 1;}' | cc -fcf-protection=none -S -o - -x c - | rg endbr64
$    # empty
```

The `psm` crate's `x86_64.s` has no `endbr64` anywhere. The SIGILL happens when
the stacker crate calls `rust_psm_stack_pointer` via function pointer:

```lldb
Process 63234 stopped
* thread #1, stop reason = signal SIGILL
    frame #0: 0x0000041ad2950840 zed.bin`rust_psm_stack_pointer
```

Fixed by `-fcf-protection=none`.

### The Combined Fix

```sh
export CFLAGS="-fno-ret-protector -fcf-protection=none"
export CXXFLAGS="-fno-ret-protector -fcf-protection=none"
```

Both flags are required. The script sets them before building native crates.

**Previous false lead:** I initially thought `-fcf-protection=none` did nothing
because `-fno-ret-protector` alone got Zed past the first few crashes. But
`-fcf-protection=none` was actually disabling the IBT checks all along — I just
didn't realize both were needed until lldb showed the crash in hand-written asm.

## SIGILL Crash Progression

Each fix round eliminated one SIGILL source, revealing the next. Zed advanced
further through startup each time:

| Round | Result | Notes |
|-------|--------|-------|
| 1 | ❌ `migrate_settings` | tree-sitter-json retguard |
| 2 | ❌ X11 client init | More retguard in other crates |
| 3 | ❌ Window creation | Still missing `-sys` crates |
| 4 | ❌ After GPU + window | Almost there |
| 5 | ✅ `rust_psm_stack_pointer` | IBT crash — hand-written asm lacks `endbr64` |
| 6 | 🔄 **Pending** | Added `-fcf-protection=none` |

Round 5 got past all the C code crashes (retguard fixed) but then hit the same
SIGILL in hand-written assembly (IBT enforcement). Both issues had to be solved.

## Native Crates Affected

All C/asm crates compiled via `cc` or `cmake`:

| Crate | Objects (approx) | In clean list? |
|-------|---|--------|
| `aws-lc-sys` | 255 | ✅ |
| `freetype-sys` | ~15 | ✅ |
| `libsqlite3-sys` | 2 | ✅ |
| `lmdb-master-sys` | 3 | ✅ |
| `psm` | ~5 | ✅ |
| `ring` | ~20 | ✅ |
| `stacker` | - | ✅ |
| `tree-sitter` + all grammars | ~10 each | ✅ |
| `wayland-sys` | - | ✅ |
| `yeslogic-fontconfig-sys` | ~5 | ✅ |
| `zstd-sys` | 38 | ✅ |
| `wasmtime` / `wiggle` | (no C in our build) | ❓ |
| `alsa-sys`, `coreaudio-sys` | (not built on OpenBSD) | n/a |

**Total in binary:** ~2000 `__retguard` symbols observed in a build
where `-fcf-protection=none` was incorrectly used (i.e., did
nothing). After `-fno-ret-protector` with the full crate list, the
count should drop to zero.

## Build Cache Invalidation

### The problem

Cargo does not invalidate C compilation when `CFLAGS` changes.
Native crate fingerprints don't track compiler flags from the
environment. After changing flags, cargo happily re-links old
object files with `__retguard` still embedded.

### What DID NOT work

**Surgical object deletion** (delete `.o`, `.a`, `output` files
only, keep Rust fingerprints). This triggers `build.rs` to
re-run, which invalidates the build script output, which cascades
into Rust recompilation of the crate. End result: same build
time as a full clean, but more fragile.

### What works: Direct deletion + sentinel

```sh
# Sentinel tracks CFLAGS. Only nuke when they change.
CET_SENTINEL="target/release-fast/.cflags"
EXPECTED_FLAGS="CFLAGS=${CFLAGS}"

if [ -f "$CET_SENTINEL" ] && \
   [ "$(cat "$CET_SENTINEL")" = "$EXPECTED_FLAGS" ]; then
    echo "CFLAGS unchanged; skipping."
else
    # Delete build outputs AND fingerprints for native crates directly
    for dir in build fingerprint; do
        cd "target/release-fast/$dir"
        for pat in tree-sitter psm stacker ring aws-lc-sys ...; do
            find . -maxdepth 1 -type d -name "${pat}-*" -exec rm -rf {} \;
        done
    done
    printf '%s' "$EXPECTED_FLAGS" > "$CET_SENTINEL"
fi
```

`cargo clean -p` doesn't reliably work for registry crates — it silently fails
or doesn't clean everything. Direct deletion of the `build/` and `.fingerprint/`
directories is surgical and reliable.

First build with correct flags: ~90 min (full native recompile).
Subsequent builds: ~5-10 min (native crates only, Rust workspace cached).

### To force a rebuild

```sh
rm ~/Code/zed/target/release-fast/.cflags
./scripts/zed.sh
```

## Runtime Status (as of 2026-06-18 14:02)

Last successful startup sequence before IBT crash:

```
INFO  [zed] starting zed version 1.8.0+dev...
INFO  [gpui_linux::linux::x11::client] XInput version: 2.4
INFO  [gpui_linux::linux::x11::client] x11: compositor present: true
INFO  [gpui_linux::linux::x11::client] Using scale factor from Xft.dpi: 1
ERROR [client::telemetry] Failed to load /etc/os-release
INFO  [prompt_store] Rules-to-skills migration already done
INFO  [gpui_linux::linux::platform] activate not implemented on Linux
INFO  [gpui_linux::linux::x11::window] Using Visual { id: 172, depth: 32 }
INFO  [gpui_linux::linux::x11::window] Creating colormap 56623106
INFO  [gpui_wgpu] Found 1 GPU adapter(s): Intel Iris Xe (Gl)
INFO  [gpui_wgpu] Selected GPU (passed configuration test)
INFO  [gpui_wgpu] Selected GPU adapter: Intel Iris Xe (Gl)
INFO  [gpui_linux::linux::x11::window] x11: no compositor present...
```

Zed successfully initializes: X11 client, GPU detection, compositor check,
window creation — then crashes in stacker's call to `psm`'s hand-written
assembly.

**lldb backtrace:**
```
lldb) run
Process 63234 stopped
* thread #1, stop reason = signal SIGILL
    frame #0: 0x0000041ad2950840 zed.bin`rust_psm_stack_pointer
zed.bin`rust_psm_stack_pointer:
->  0x41ad2950840 <+0>: leaq   0x8(%rsp), %rax
    0x41ad2950845 <+5>: retq   
```

`rust_psm_stack_pointer` is in `psm`'s `x86_64.s` — hand-written asm with no
`endbr64`. The `leaq` at the start of the function isn't itself a problem. The
IBT check happens when the CPU detects the indirect call target lacks `endbr64`.

**Fix in progress:** Script now sets `CFLAGS="-fno-ret-protector -fcf-protection=none"`.
Full rebuild triggered (sentinel cleared). Pending test.

Zed successfully initializes through:
- ✅ Settings migration
- ✅ X11 client + XInput + scale factor
- ✅ GPU adapter detection + EGL test
- ✅ Window visual + colormap creation
- ✅ Window decoration check ("no compositor present")
- ❌ Crashes in `stacker::on_stack` → `psm::rust_psm_stack_pointer` (IBT failure)

The crash is in hand-written x86_64 assembly that lacks `endbr64` at its entry
point. Tiger Lake's IBT enforcement SIGILLs on indirect calls to functions
without `endbr64`.

## Known Issues

| Issue | Status | Notes |
|-------|--------|-------|
| SIGILL in `psm` asm | 🔄 Fix pending | Added `-fcf-protection=none`, testing |
| `os-release` missing | ℹ️ Non-fatal | OpenBSD has no os-release |
| Unstripped binary (410 MB) | ℹ️ Intentional | Debug symbols for backtraces |
| `datasize` limit | ⚠️ Needs manual set | `ulimit -d 8388608` in wrapper |

## Other Fixes (Stable)

### wgpu GLES Backend (openbsd cfg)

Added `target_os = "openbsd"` to the wgpu `gles` cfg alias in
`wgpu/build.rs` and `wgpu-core/build.rs` (via cargo cache patch).
Without this, OpenBSD builds without the GLES backend and panics
with `Backends::empty()`.

### aws-lc-sys NO_ASM

`curve25519_x25519base.S` has a PIC local symbol bug on OpenBSD.
`AWS_LC_SYS_NO_ASM=1` forces the pure-C fallback — but it panics
in release profiles. The script patches `cmake_builder.rs` in
the cargo registry to replace the panic with
`cmake_cfg.define("OPENSSL_NO_ASM", "1")`.

### Vulkan (explored, reverted)

OpenBSD's ANV driver (`libvulkan_intel.so`) crashes during
`vkCreateInstance`. Likely the same `__retguard` issue in Mesa,
or worse. Not worth pursuing — GLES works.

## Next Steps

1. **Current build (~90 min in progress):** Rebuilding with
   `CFLAGS="-fno-ret-protector -fcf-protection=none"`.
2. **Test runtime:** If the IBT crash is gone, Zed should launch
   fully. If new SIGILLs appear, they may be in other assembly
   files lacking `endbr64`.
3. **Verify rendering:** Confirm wgpu + GLES work end-to-end.
4. **Verify window interactions:** Mouse, keyboard, input, clipboard.
5. **Update this doc** with results.

## Background

- **Jun 9-10:** Initial build at `137e677a05`. 295 MB binary
  panicked with `Backends::empty()` at runtime (no GLES on
  OpenBSD).
- **Jun 12-16:** Cherry-picked upstream to `fca2ccd403`. Fixed
  4 build bugs, fixed `pub mod queue` multi-line cfg format.
  Added wgpu GLES openbsd cfg. Confirmed GPU detection.
- **Jun 16:** SIGILL crash. Used lldb to trace to
  `tree_sitter_json`. Added `-fcf-protection=none` (WRONG FLAG).
  Fixed aws-lc-sys NO_ASM.
- **Jun 17-18 (early):** Discovered `-fcf-protection=none` is a no-op
  on OpenBSD. Correct flag is `-fno-ret-protector`. SIGILL kept
  recurring as new `-sys` crates were discovered. Surgical
  object deletion approach failed (cascades into Rust recompile).
  Settled on `cargo clean -p` + CFLAGS sentinel approach.
- **Jun 18 (current):** After fixing retguard, SIGILL persisted in
  `psm::rust_psm_stack_pointer` (hand-written assembly). Discovered
  OpenBSD Clang also defaults to `-fcf-protection=branch`, which
  requires `endbr64` at indirect call targets. The assembly lacks
  `endbr64`, causing IBT violations on Tiger Lake. Added
  `-fcf-protection=none` to disable this. Testing pending.

#!/bin/sh
#
# zed.sh - Build Zed editor with OpenBSD compatibility patches
#
# Clones/pulls ~/Code/zed, applies patches, builds with X11 only,
# installs to ~/.local/share/openriot/config/bin/zed.
#
# Usage: ./zed.sh
#
# Environment variables:
#   ZED_COMMIT    - Pin to a specific commit (default: origin/main)
#   ZED_FEATURES  - Extra cargo features (default: x11)

set -e

ZED_REPO="https://github.com/zed-industries/zed"
ZED_COMMIT="${ZED_COMMIT:-fca2ccd403}"
ZED_FEATURES="${ZED_FEATURES:-x11}"
PATCH_FILE="$(dirname "$0")/zed-patch.diff"
INSTALL_DIR="${HOME}/.local/share/openriot/config"
SOURCE_DIR="${HOME}/Code/zed"

# Portable in-place sed (OpenBSD's /usr/bin/sed does not support -i).
# Usage: sed_i <file> <sed-expression>
sed_i() {
    sed "$2" "$1" > "$1.tmp" && mv "$1.tmp" "$1"
}

# Step 0: Install Rust if cargo is missing
if ! command -v cargo >/dev/null 2>&1; then
    echo "Installing Rust via pkg_add..."
    doas pkg_add rust
fi

# Step 0.5: Verify datasize limits (OpenBSD uses RLIMIT_DATA, ulimit -d)
DATA_KB=$(ulimit -d)
if [ "$DATA_KB" != "unlimited" ] && [ "$DATA_KB" -lt 8388608 ]; then
    echo "WARNING: current datasize limit is ${DATA_KB}KB (< 8GB)."
    echo "Zed linking requires ~4-6GB data segment."
    echo "Set login.conf datasize or run with: ulimit -d 8388608"
fi

# Step 0.7: Use persistent storage for compiler temp files
# /tmp is a small mfs (495 MB) and fills up during large rustc compiles.
export TMPDIR="${HOME}/.cache/zed/tmp"
mkdir -p "${TMPDIR}"
echo "TMPDIR=${TMPDIR}"

# Step 1: Clone or update zed
echo "Cloning/updating zed..."
if [ -d "$SOURCE_DIR/.git" ]; then
    if [ -n "$ZED_COMMIT" ]; then
        CURRENT_COMMIT=$(git -C "$SOURCE_DIR" rev-parse HEAD 2>/dev/null || echo "")
        if [ "$CURRENT_COMMIT" = "$ZED_COMMIT" ]; then
            echo "Source already at pinned commit ($ZED_COMMIT), using existing checkout"
        else
            echo "Updating to pinned commit $ZED_COMMIT..."
            git -C "$SOURCE_DIR" fetch origin
            git -C "$SOURCE_DIR" checkout "$ZED_COMMIT"
        fi
    else
        printf "Fresh pull from origin/main? [y/N] "
        read -r REPLY
        case "$REPLY" in
            y|Y) echo "Updating to latest origin/main..."
                 git -C "$SOURCE_DIR" fetch origin
                 git -C "$SOURCE_DIR" reset --hard origin/main
                 echo "Source at $(git -C "$SOURCE_DIR" rev-parse --short HEAD)"
                 ;;
            *)   echo "Skipping pull, using existing checkout."
                 echo "Source at $(git -C "$SOURCE_DIR" rev-parse --short HEAD)"
                 ;;
        esac
    fi
else
    rm -rf "$SOURCE_DIR"
    mkdir -p "$(dirname "$SOURCE_DIR")"
    git clone "$ZED_REPO" "$SOURCE_DIR"
    if [ -n "$ZED_COMMIT" ]; then
        git -C "$SOURCE_DIR" checkout "$ZED_COMMIT"
        echo "Cloned and checked out pinned commit ($ZED_COMMIT)"
    fi
fi

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

# Step 2.5: Patch wgpu in cargo cache (mandatory before first build).
# Cargo can succeed with cfg!(gles) evaluating to false, producing a binary
# that compiles but panics at runtime. These patches add target_os = "openbsd"
# to the wgpu GLES cfg aliases. Idempotent.
echo "Patching wgpu for OpenBSD GLES backend..."
WGPU_ROOT=$(find "${HOME}/.cargo/git/checkouts" -maxdepth 1 -type d -name 'wgpu-*' 2>/dev/null | head -1)
if [ -n "$WGPU_ROOT" ]; then
    for rev in "$WGPU_ROOT"/*/; do
        [ -d "$rev" ] || continue
        PDEPS="$rev/wgpu-core/platform-deps/windows-linux-android/Cargo.toml"
        CTOML="$rev/wgpu-core/Cargo.toml"
        WRS="$rev/wgpu/build.rs"
        CRS="$rev/wgpu-core/build.rs"
        [ -f "$PDEPS" ] && ! rg -q 'target_os = "openbsd"' "$PDEPS" 2>/dev/null && \
            sed_i "$PDEPS" 's/target_os = "netbsd"/target_os = "netbsd", target_os = "openbsd"/'
        [ -f "$CTOML" ] && ! rg -q 'target_os = "openbsd"' "$CTOML" 2>/dev/null && \
            sed_i "$CTOML" 's/target_os = "freebsd"))/target_os = "freebsd", target_os = "openbsd"))/'
        [ -f "$WRS" ] && ! rg -q 'target_os = "openbsd",$' "$WRS" 2>/dev/null && \
            sed_i "$WRS" 's|gles: { any(|gles: { any(\n            target_os = "openbsd",|'
        [ -f "$CRS" ] && ! rg -q 'target_os = "openbsd",$' "$CRS" 2>/dev/null && \
            sed_i "$CRS" 's|gles: { any(|gles: { any(\n            target_os = "openbsd",|'
    done
    # Verify all four files now have the OpenBSD cfg.
    for rev in "$WGPU_ROOT"/*/; do
        [ -d "$rev" ] || continue
        for f in "$rev/wgpu-core/platform-deps/windows-linux-android/Cargo.toml" \
                 "$rev/wgpu-core/Cargo.toml" \
                 "$rev/wgpu/build.rs" \
                 "$rev/wgpu-core/build.rs"; do
            if [ -f "$f" ] && ! rg -q 'target_os = "openbsd"' "$f" 2>/dev/null; then
                echo "ERROR: wgpu patch missing in $f" >&2
                exit 1
            fi
        done
    done
    echo "wgpu patches verified"
else
    echo "WARNING: wgpu source not found in cargo cache; will be patched on attempt 2"
fi

# Step 2.6: Patch native crates for Intel IBT (Indirect Branch Tracking).
# OpenBSD Clang defaults to -fcf-protection=branch which enforces IBT on
# Tiger Lake+. Hand-written assembly in psm/ring lacks endbr64 markers,
# causing SIGILL on indirect calls. We patch the assembly sources and
# clear only those crates' build caches.
IBT_SENTINEL="$SOURCE_DIR/target/release-fast/.ibt-patched"
if [ ! -f "$IBT_SENTINEL" ]; then
    echo "Patching native crates for Intel IBT..."
    
    # Patch psm: add endbr64 after .cfi_startproc in each FUNCTION
    PSM_SRC=$(find "${HOME}/.cargo/registry/src" -name "x86_64.s" -path "*/psm-*/src/arch/*" 2>/dev/null | head -1)
    if [ -n "$PSM_SRC" ] && [ -f "$PSM_SRC" ]; then
        if ! rg -q 'endbr64' "$PSM_SRC" >/dev/null 2>&1; then
            awk '
            /^FUNCTION\(rust_psm_/{p=1}
            p && /^.cfi_startproc/{print; print "    endbr64"; p=0; next}
            {print}
            ' "$PSM_SRC" > "$PSM_SRC.tmp" && mv "$PSM_SRC.tmp" "$PSM_SRC"
            echo "  ✓ psm x86_64.s patched"
        else
            echo "  ✓ psm already patched"
        fi
    fi
    
    # Patch ring: redefine _CET_ENDBR as endbr64 in asm_base.h
    RING_ASM_BASE=$(find "${HOME}/.cargo/registry/src" -name "asm_base.h" -path "*/ring-*/include/*" 2>/dev/null | head -1)
    if [ -n "$RING_ASM_BASE" ] && [ -f "$RING_ASM_BASE" ]; then
        if ! rg -q '#define _CET_ENDBR endbr64' "$RING_ASM_BASE" >/dev/null 2>&1; then
            sed -i.bak 's|^#define _CET_ENDBR$|#define _CET_ENDBR endbr64|' "$RING_ASM_BASE"
            echo "  ✓ ring asm_base.h patched"
        else
            echo "  ✓ ring already patched"
        fi
    fi
    
    # Clear build outputs and fingerprints for patched crates only
    echo "Clearing build caches for patched crates..."
    for crate in psm ring; do
        rm -rf "$SOURCE_DIR/target/release-fast/build/${crate}-"* 2>/dev/null || true
        rm -rf "$SOURCE_DIR/target/release-fast/.fingerprint/${crate}-"* 2>/dev/null || true
    done
    
    # Create sentinel
    mkdir -p "$(dirname "$IBT_SENTINEL")"
    printf '%s' "ibt-patched-$(date +%s)" > "$IBT_SENTINEL"
    echo "IBT patches applied. Incremental rebuild will recompile psm/ring (~5-10 min)."
else
    echo "IBT patches already applied."
fi

# Step 3: Attempt build, patching registry sources on failure
echo "Building zed with features=$ZED_FEATURES..."
export CARGO_BUILD_JOBS="${CARGO_BUILD_JOBS:-1}"
# OpenBSD doesn't search /usr/local/lib by default; xkbcommon and
# other native libraries live there.
# Opt the resulting binary out of branch-target CFI (IBT) via the OpenBSD
# linker. PT_OPENBSD_NOBTCFI tells the kernel to skip endbr64 enforcement
# for this process, avoiding SIGILL from indirect calls into Rust hostcall
# functions (rustc stable exposes no `+ibt` target-feature) and from
# Cranelift JIT'd code (Cranelift has its own BTI codegen flag, but there
# is no stable Rust path to thread it). The psm/ring asm patches and
# -fno-ret-protector above are kept as defense in depth.
export RUSTFLAGS="-C link-arg=-Wl,-z,nobtcfi -L /usr/local/lib ${RUSTFLAGS:-}"
cd "$SOURCE_DIR"

# OpenBSD Clang defaults enable TWO separate CPU protections:
# 1. -ret-protector        → emits __retguard XOR canaries in C code (SIGILL)
# 2. -fcf-protection=branch → requires endbr64 at indirect call targets (IBT)
# Native crates compile C code AND hand-written .s assembly. The assembly
# files (e.g., psm's x86_64.s) don't know about endbr64, so the CPU's IBT
# enforcement SIGILLs on indirect calls to them. Disable retguard, keep IBT.
export CFLAGS="-fno-ret-protector"
export CXXFLAGS="-fno-ret-protector"

# Track CFLAGS and RUSTFLAGS. If they change, delete build outputs and
# fingerprints for affected crates. cargo clean -p doesn't reliably work
# for registry crates.
CET_SENTINEL="$SOURCE_DIR/target/release-fast/.cflags"
EXPECTED_FLAGS="CFLAGS=${CFLAGS} RUSTFLAGS=${RUSTFLAGS}"
if [ -f "$CET_SENTINEL" ] && [ "$(cat "$CET_SENTINEL")" = "$EXPECTED_FLAGS" ]; then
    echo "Build flags unchanged; skipping native crate rebuild."
else
    if [ -f "$CET_SENTINEL" ]; then
        echo "Build flags changed (was: $(cat "$CET_SENTINEL"), now: $EXPECTED_FLAGS)"
    fi
    echo "Build cache invalidation: deleting crate outputs..."

    # Delete .o, .a, and build script outputs. The CFLAGS list is C/C++/asm
    # crates affected by clang flags. wasmtime is added because RUSTFLAGS
    # changes (e.g. the -C link-arg= value) alter the final link step,
    # which makes the embedded ELF program header table inconsistent
    # with the cached binary.
    cd "$SOURCE_DIR/target/release-fast/build"
    for pattern in tree-sitter psm stacker ring aws-lc-sys freetype-sys \
                   libsqlite3-sys lmdb-master-sys wayland-sys yeslogic-fontconfig-sys \
                   zstd-sys wgpu wgpu-core wgpu-hal wasmtime; do
        find . -maxdepth 1 -type d -name "${pattern}-*" -exec rm -rf {} \; 2>/dev/null || true
    done
    cd "$SOURCE_DIR/target/release-fast/.fingerprint"
    for pattern in tree-sitter psm stacker ring aws-lc-sys freetype-sys \
                   libsqlite3-sys lmdb-master-sys wayland-sys yeslogic-fontconfig-sys \
                   zstd-sys wgpu wgpu-core wgpu-hal wasmtime; do
        find . -maxdepth 1 -type d -name "${pattern}-*" -exec rm -rf {} \; 2>/dev/null || true
    done

    mkdir -p "$(dirname "$CET_SENTINEL")"
    printf '%s' "$EXPECTED_FLAGS" > "$CET_SENTINEL"
    echo "Crate outputs deleted. Rebuild required (~5-10 min)."
fi

cd "$SOURCE_DIR"
echo "Building zed (debug profile, CARGO_BUILD_JOBS=$CARGO_BUILD_JOBS)..."

# Fetch dependencies first so registry sources are available for patching.
# On a clean tree this downloads everything; on an existing tree it's fast.
echo "Fetching dependencies..."
cargo fetch --manifest-path crates/zed/Cargo.toml --no-default-features --features "$ZED_FEATURES" 2>/dev/null || true

# First build attempt may fail on third-party crates not yet in cargo registry.
# After failure, patch the extracted sources and retry.
for attempt in 1 2; do
    if [ "$attempt" -eq 2 ]; then
        echo "Retry $attempt: patching third-party crate sources..."
        WASMTIME_DIR=$(find "${HOME}/.cargo/registry/src" -type d \
            -name 'wasmtime-wasi-[0-9]*' 2>/dev/null | head -1)
        if [ -n "$WASMTIME_DIR" ]; then
            T="$WASMTIME_DIR/src/p2/tcp.rs"
            U="$WASMTIME_DIR/src/sockets/util.rs"
            sed_i "$T" 's/Ok(sockopt::tcp_keepidle(view)?)/Err(ErrorCode::NotSupported.into())/'
            sed_i "$T" 's/Ok(sockopt::tcp_keepintvl(view)?)/Err(ErrorCode::NotSupported.into())/'
            sed_i "$T" 's/Ok(sockopt::tcp_keepcnt(view)?)/Err(ErrorCode::NotSupported.into())/'
            sed_i "$U" 's/sockopt::set_tcp_keepidle(fd, Duration::from_nanos(value))?;/return Err(ErrorCode::NotSupported);/'
            sed_i "$U" 's/sockopt::set_tcp_keepintvl(fd, value\.clamp(MIN, MAX))?;/return Err(ErrorCode::NotSupported);/'
            sed_i "$U" 's/sockopt::set_tcp_keepcnt(fd, value\.clamp(MIN_CNT, MAX_CNT))?;/return Err(ErrorCode::NotSupported);/'
            echo "wasmtime-wasi patched"
        fi
        IPC_DIR=$(find "${HOME}/.cargo/registry/src" -type d \
            -name 'ipc-channel-*' 2>/dev/null | head -1)
        if [ -n "$IPC_DIR" ]; then
            IPC_MOD="$IPC_DIR/src/platform/unix/mod.rs"
            if ! rg -q 'target_os = "openbsd"' "$IPC_MOD" 2>/dev/null; then
                ed -s "$IPC_MOD" << 'EOF'
/const POLLRDHUP: libc::c_short = 0x4000;
a
                #[cfg(target_os = "openbsd")]
                const POLLRDHUP: libc::c_short = 0;
.
w
q
EOF
                echo "ipc-channel patched"
            fi
        fi
    fi

    # aws-lc-sys >= 0.40.0 handles AWS_LC_SYS_NO_ASM in release builds
    # Cargo does not detect registry source changes, so when a patch is
    # applied we delete the build cache for that crate.
    MAPSTACK_PATCHED=false

    WT_FIBER_DIR=$(find "${HOME}/.cargo/registry/src" -type d \
        -name 'wasmtime-internal-fiber-*' 2>/dev/null | head -1)
    if [ -n "$WT_FIBER_DIR" ]; then
        FIBER_SRC="$WT_FIBER_DIR/src/unix.rs"
        if [ -f "$FIBER_SRC" ] && ! rg -q 'MapFlags::from_bits_retain' "$FIBER_SRC" 2>/dev/null; then
            sed_i "$FIBER_SRC" 's@rustix::mm::MapFlags::PRIVATE,$@rustix::mm::MapFlags::PRIVATE | rustix::mm::MapFlags::from_bits_retain(0x4000),@'
            rm -rf "$SOURCE_DIR/target/release-fast/build/wasmtime-internal-fiber-"* 2>/dev/null || true
            rm -rf "$SOURCE_DIR/target/release-fast/.fingerprint/wasmtime-internal-fiber-"* 2>/dev/null || true
            echo "wasmtime-internal-fiber patched (MAP_STACK)"
            MAPSTACK_PATCHED=true
        fi
    fi

    WT_MAIN_DIR=$(find "${HOME}/.cargo/registry/src" -type d \
        -name 'wasmtime-[0-9]*' 2>/dev/null | head -1)
    if [ -n "$WT_MAIN_DIR" ]; then
        STACKSW_SRC="$WT_MAIN_DIR/src/runtime/vm/stack_switching/stack/unix.rs"
        if [ -f "$STACKSW_SRC" ] && ! rg -q 'MapFlags::from_bits_retain' "$STACKSW_SRC" 2>/dev/null; then
            sed_i "$STACKSW_SRC" 's@rustix::mm::MapFlags::PRIVATE,$@rustix::mm::MapFlags::PRIVATE | rustix::mm::MapFlags::from_bits_retain(0x4000),@'
            rm -rf "$SOURCE_DIR/target/release-fast/build/wasmtime-"* 2>/dev/null || true
            rm -rf "$SOURCE_DIR/target/release-fast/.fingerprint/wasmtime-"* 2>/dev/null || true
            echo "wasmtime stack-switching patched (MAP_STACK)"
            MAPSTACK_PATCHED=true
        fi

        SIG_SRC="$WT_MAIN_DIR/src/runtime/vm/sys/unix/signals.rs"
        if [ -f "$SIG_SRC" ] && ! rg -q 'PROT_NONE.*MAP_STACK.*EINVAL' "$SIG_SRC" 2>/dev/null; then
            # Ensure MAP_STACK is in the mmap flags (idempotent).
            sed_i "$SIG_SRC" 's@rustix::mm::MapFlags::PRIVATE,$@rustix::mm::MapFlags::PRIVATE | rustix::mm::MapFlags::from_bits_retain(0x4000),@'
            # Change PROT_NONE to PROT_READ|PROT_WRITE and insert a
            # guard-page mprotect.  PROT_NONE + MAP_STACK is EINVAL
            # on OpenBSD; allocate RW first, then lock the guard page.
            ed -s "$SIG_SRC" << 'EDEOF'
/rustix::mm::ProtFlags::empty(),/
c
                rustix::mm::ProtFlags::READ | rustix::mm::ProtFlags::WRITE,
.
/let stack_ptr = (ptr as usize + guard_size) as/
a

        // OpenBSD: PROT_NONE + MAP_STACK is EINVAL; allocate RW
        // first, then lock the guard page to PROT_NONE.
        unsafe {
            rustix::mm::mprotect(
                ptr,
                guard_size,
                rustix::mm::MprotectFlags::empty(),
            )
            .expect("mprotect to set guard page failed");
        }
.
w
q
EDEOF
            rm -rf "$SOURCE_DIR/target/release-fast/build/wasmtime-"* 2>/dev/null || true
            rm -rf "$SOURCE_DIR/target/release-fast/.fingerprint/wasmtime-"* 2>/dev/null || true
            echo "wasmtime signals patched (MAP_STACK + guard page)"
            MAPSTACK_PATCHED=true
        fi
    fi

    STACKER_DIR=$(find "${HOME}/.cargo/registry/src" -type d \
        -name 'stacker-*' 2>/dev/null | head -1)
    if [ -n "$STACKER_DIR" ]; then
        STACKER_SRC="$STACKER_DIR/src/mmap_stack_restore_guard.rs"
        if [ -f "$STACKER_SRC" ] && rg -q 'libc::MAP_PRIVATE | libc::MAP_ANON,$' "$STACKER_SRC" 2>/dev/null; then
            sed_i "$STACKER_SRC" 's#libc::MAP_PRIVATE | libc::MAP_ANON,$#libc::MAP_PRIVATE | libc::MAP_ANON | libc::MAP_STACK,#'
            rm -rf "$SOURCE_DIR/target/release-fast/build/stacker-"* 2>/dev/null || true
            rm -rf "$SOURCE_DIR/target/release-fast/.fingerprint/stacker-"* 2>/dev/null || true
            echo "stacker patched (MAP_STACK)"
            MAPSTACK_PATCHED=true
        fi
    fi

    if [ "$MAPSTACK_PATCHED" = "true" ]; then
        echo "MAP_STACK patches applied. Incremental rebuild."
    fi

    if sh -c "ulimit -d 8388608 && AWS_LC_SYS_NO_ASM=1 exec cargo build --profile release-fast --no-default-features --features '${ZED_FEATURES}' -j '${CARGO_BUILD_JOBS}'"; then
        break
    elif [ "$attempt" -eq 2 ]; then
        echo "Build failed after patching third-party sources."
        exit 1
    fi
done

# Step 4: Strip and install
echo "Installing binary..."
mkdir -p "${INSTALL_DIR}/bin"
cp target/release-fast/zed "${INSTALL_DIR}/bin/zed.bin"
# Keep debug symbols — needed for crash backtraces. Set ZED_STRIP=1 to strip.

# Create a wrapper that raises datasize and file descriptor limits
# before exec. OpenBSD defaults to 128 FDs which is insufficient for
# zed scanning large git worktrees.
cat > "${INSTALL_DIR}/bin/zed" << 'WRAPPER'
#!/bin/sh
ulimit -d 8388608 2>/dev/null || ulimit -d 4194304 2>/dev/null || true
ulimit -n 524288 2>/dev/null || true
exec "$(dirname "$0")/zed.bin" "$@"
WRAPPER
chmod +x "${INSTALL_DIR}/bin/zed"

# Step 5: Verify
if [ -x "${INSTALL_DIR}/bin/zed" ] && [ -x "${INSTALL_DIR}/bin/zed.bin" ]; then
    echo "Done! Installed to ${INSTALL_DIR}/bin/zed"
    echo "Binary size: $(ls -lh "${INSTALL_DIR}/bin/zed.bin" | awk '{print $5}')"
else
    echo "Error: zed binary not found after build"
    exit 1
fi
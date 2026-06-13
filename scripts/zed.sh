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

# Step 3: Attempt build, patching registry sources on failure
echo "Building zed with features=$ZED_FEATURES..."
export CARGO_BUILD_JOBS="${CARGO_BUILD_JOBS:-1}"
# OpenBSD doesn't search /usr/local/lib by default; xkbcommon and
# other native libraries live there.
export RUSTFLAGS="-L /usr/local/lib ${RUSTFLAGS:-}"
cd "$SOURCE_DIR"

# Force wgpu build-script re-run: cargo clean -p is the only
# approach that reliably defeats the build-script cache.
# find-delete of output files alone is insufficient.
cargo clean -p wgpu -p wgpu-core -p wgpu-hal

echo "Building zed (debug profile, CARGO_BUILD_JOBS=$CARGO_BUILD_JOBS)..."

# First build attempt may fail on third-party crates not yet in cargo registry.
# After failure, patch the extracted sources and retry.
for attempt in 1 2; do
    if [ "$attempt" -eq 2 ]; then
        echo "Retry $attempt: patching third-party crate sources..."
        WASMTIME_DIR=$(find "${HOME}/.cargo/registry/src" -maxdepth 1 -type d \
            -name 'wasmtime-wasi-*' 2>/dev/null | head -1)
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
        IPC_DIR=$(find "${HOME}/.cargo/registry/src" -maxdepth 1 -type d \
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
        # wgpu patches are applied in Step 2.5 before this loop.
    fi

    if sh -c "ulimit -d 8388608 && exec cargo build --profile release-fast --no-default-features --features '${ZED_FEATURES}' -j '${CARGO_BUILD_JOBS}'"; then
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
strip "${INSTALL_DIR}/bin/zed.bin" 2>/dev/null || true

# Create a wrapper that raises datasize limits before exec,
# so the stripped binary does not OOM at runtime.
cat > "${INSTALL_DIR}/bin/zed" << 'WRAPPER'
#!/bin/sh
echo "ulimit -d 8388608..."
ulimit -d 8388608 2>/dev/null || ulimit -d 4194304 2>/dev/null || true
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
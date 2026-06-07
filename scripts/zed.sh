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
ZED_COMMIT="${ZED_COMMIT:-}"
ZED_FEATURES="${ZED_FEATURES:-x11}"
PATCH_FILE="$(dirname "$0")/zed-patch.diff"
INSTALL_DIR="${HOME}/.local/share/openriot/config"
SOURCE_DIR="${HOME}/Code/zed"

# Step 0: Install Rust if cargo is missing
if ! command -v cargo >/dev/null 2>&1; then
    echo "Installing Rust via pkg_add..."
    doas pkg_add rust
fi

# Step 0.5: Verify virtual memory limits
VMEM_KB=$(ulimit -v)
if [ "$VMEM_KB" != "unlimited" ] && [ "$VMEM_KB" -lt 8388608 ]; then
    echo "WARNING: current virtual memory limit is ${VMEM_KB}KB (< 8GB)."
    echo "Zed linking requires ~4-6GB virtual memory."
    echo "Set login.conf maxproc-vmem or run with: ulimit -v 8388608"
fi

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

# Step 2: Apply OpenBSD patches (idempotent)
echo "Applying OpenBSD compatibility patches..."
if [ -f "$PATCH_FILE" ] && [ -s "$PATCH_FILE" ]; then
    if patch -d "$SOURCE_DIR" -p1 --dry-run -f < "$PATCH_FILE" >/dev/null 2>&1; then
        patch -d "$SOURCE_DIR" -p1 -f < "$PATCH_FILE"
        echo "Patch applied"
    elif patch -d "$SOURCE_DIR" -p1 --dry-run -f -R < "$PATCH_FILE" >/dev/null 2>&1; then
        echo "Patch already applied"
    else
        echo "ERROR: Patch cannot be applied. Upstream code may have changed."
        echo "Update the patch or the commit pin."
        exit 1
    fi
else
    echo "No patch file found, building without patches"
fi

# Step 3: Attempt build, patching registry sources on failure
echo "Building zed with features=$ZED_FEATURES..."
export CARGO_BUILD_JOBS="${CARGO_BUILD_JOBS:-1}"
# OpenBSD doesn't search /usr/local/lib by default; xkbcommon and
# other native libraries live there.
export RUSTFLAGS="-L /usr/local/lib ${RUSTFLAGS:-}"
cd "$SOURCE_DIR"
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
            sed -i '' -e 's/Ok(sockopt::tcp_keepidle(view)?)/Err(ErrorCode::NotSupported.into())/' "$T"
            sed -i '' -e 's/Ok(sockopt::tcp_keepintvl(view)?)/Err(ErrorCode::NotSupported.into())/' "$T"
            sed -i '' -e 's/Ok(sockopt::tcp_keepcnt(view)?)/Err(ErrorCode::NotSupported.into())/' "$T"
            sed -i '' -e 's/sockopt::set_tcp_keepidle(fd, Duration::from_nanos(value))?;/return Err(ErrorCode::NotSupported);/' "$U"
            sed -i '' -e 's/sockopt::set_tcp_keepintvl(fd, value\.clamp(MIN, MAX))?;/return Err(ErrorCode::NotSupported);/' "$U"
            sed -i '' -e 's/sockopt::set_tcp_keepcnt(fd, value\.clamp(MIN_CNT, MAX_CNT))?;/return Err(ErrorCode::NotSupported);/' "$U"
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
        # wgpu: GLES backend platform gate excludes OpenBSD.
        WGPU_PLATFORM_FILE=$(find "${HOME}/.cargo/git/checkouts/wgpu-*" -path '*/wgpu-core/platform-deps/windows-linux-android/Cargo.toml' 2>/dev/null | head -1)
        if [ -n "$WGPU_PLATFORM_FILE" ] && ! rg -q 'target_os = "openbsd"' "$WGPU_PLATFORM_FILE" 2>/dev/null; then
            sed -i '' -e 's/target_os = "netbsd"/target_os = "netbsd", target_os = "openbsd"/' "$WGPU_PLATFORM_FILE"
            echo "wgpu platform-deps patched"
        fi
        # wgpu-core/Cargo.toml: include OpenBSD in the platform-deps target cfg.
        WGPU_CORE_TOML=$(find "${HOME}/.cargo/git/checkouts/wgpu-*" -path '*/wgpu-core/Cargo.toml' 2>/dev/null | head -1)
        if [ -n "$WGPU_CORE_TOML" ] && ! rg -q 'target_os = "openbsd"' "$WGPU_CORE_TOML" 2>/dev/null; then
            sed -i '' -e 's/target_os = "freebsd"))/target_os = "freebsd", target_os = "openbsd"))/' "$WGPU_CORE_TOML"
            echo "wgpu-core Cargo.toml patched"
        fi
        # wgpu/build.rs: add OpenBSD to the gles cfg alias.
        WGPU_BUILD_RS=$(find "${HOME}/.cargo/git/checkouts/wgpu-*" -path '*/wgpu/build.rs' 2>/dev/null | head -1)
        if [ -n "$WGPU_BUILD_RS" ] && ! rg -q 'target_os = "openbsd"' "$WGPU_BUILD_RS" 2>/dev/null; then
            sed -i '' -e 's/target_os = "freebsd", Emscripten/target_os = "freebsd", target_os = "openbsd", Emscripten/' "$WGPU_BUILD_RS"
            echo "wgpu build.rs patched"
        fi
        # wgpu-core/build.rs: add OpenBSD to the gles cfg alias without affecting vulkan.
        WGPU_CORE_BUILD_RS=$(find "${HOME}/.cargo/git/checkouts/wgpu-*" -path '*/wgpu-core/build.rs' 2>/dev/null | head -1)
        if [ -n "$WGPU_CORE_BUILD_RS" ] && ! rg -q 'target_os = "openbsd"' "$WGPU_CORE_BUILD_RS" 2>/dev/null; then
            sed -i '' -e 's/all(windows_linux_android, feature = "gles"), \/\/ Regular GLES/any(all(windows_linux_android, feature = "gles"), all(target_os = "openbsd", feature = "gles")), \/\/ Regular GLES/' "$WGPU_CORE_BUILD_RS"
            echo "wgpu-core build.rs patched"
        fi
    fi

    if sh -c "ulimit -v 8388608 && exec cargo build --profile release-fast --no-default-features --features '${ZED_FEATURES}' -j '${CARGO_BUILD_JOBS}'"; then
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
echo "ulimit -v 8388608..."
ulimit -v 8388608 2>/dev/null || ulimit -v 4194304 2>/dev/null || true
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
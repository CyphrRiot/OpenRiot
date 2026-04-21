#!/bin/sh
#
# build-gurk.sh - Build gurk-rs with OpenBSD SIGSEGV fix
#
# This script patches gurk-rs to disable notify-rust, which causes a SIGSEGV
# on OpenBSD when receiving messages (notify-rust calls /proc/self/exe which
# does not exist on OpenBSD).
#
# Usage: ./gurk.sh
#
# The patched binary is installed to ~/.local/share/openriot/config/bin/gurk

set -e

GURK_REPO="https://github.com/boxdot/gurk-rs"
GURK_COMMIT="02d3c45702142febdbbbaa4afea3f38222dd9db8"
PATCH_FILE="$(dirname "$0")/gurk-patch.diff"
INSTALL_DIR="${HOME}/.local/share/openriot/config"
SOURCE_DIR="${HOME}/src/gurk-rs"

# Step 0: Install Rust if cargo is missing
if ! command -v cargo >/dev/null 2>&1; then
    echo "Installing Rust via pkg_add..."
    doas pkg_add rust
fi

# Step 1: Clone or update gurk-rs
echo "Cloning/updating gurk-rs..."
if [ -d "$SOURCE_DIR" ]; then
    CURRENT_COMMIT=$(git -C "$SOURCE_DIR" rev-parse HEAD 2>/dev/null || echo "")
    if [ "$CURRENT_COMMIT" = "$GURK_COMMIT" ]; then
        echo "Source already at correct commit ($GURK_COMMIT), using existing checkout"
    else
        echo "Updating from $CURRENT_COMMIT to $GURK_COMMIT..."
        git -C "$SOURCE_DIR" fetch origin "$GURK_COMMIT"
        git -C "$SOURCE_DIR" checkout "$GURK_COMMIT"
    fi
else
    mkdir -p "$(dirname "$SOURCE_DIR")"
    git clone "$GURK_REPO" "$SOURCE_DIR"
    git -C "$SOURCE_DIR" checkout "$GURK_COMMIT"
fi

# Step 2: Apply patch (idempotent - patch already applied if nothing changed)
echo "Applying SIGSEGV fix patch..."
if git -C "$SOURCE_DIR" apply --check "$PATCH_FILE" 2>/dev/null; then
    git -C "$SOURCE_DIR" apply "$PATCH_FILE"
    echo "Patch applied"
else
    echo "Patch already applied or not needed"
fi

# Step 3: Build with optimizations for smaller binary
echo "Building gurk with LTO and strip..."
export CARGO_PROFILE_RELEASE_LTO=thin
export CARGO_PROFILE_RELEASE_STRIP=debuginfo
PKG_CONFIG=echo \
  cargo install gurk --force \
    --locked \
    --path "$SOURCE_DIR" \
    --root "$INSTALL_DIR"

# Step 4: Strip debug symbols for smaller binary
echo "Stripping debug symbols..."
if [ -f "${INSTALL_DIR}/bin/gurk" ]; then
    strip "${INSTALL_DIR}/bin/gurk" 2>/dev/null || true
fi

# Step 5: Verify
if [ -x "${INSTALL_DIR}/bin/gurk" ]; then
    echo "Done! Installed to ${INSTALL_DIR}/bin/gurk"
    "${INSTALL_DIR}/bin/gurk" --version
else
    echo "Error: gurk binary not found after build"
    exit 1
fi
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
GURK_COMMIT="${GURK_COMMIT:-v0.10.2}"
# presage 4b817406 bumps libsignal-service-rs to a rev that sends the
# uploadLength query param the Signal server now requires for attachment
# uploads. gurk 0.10.2 pins an older rev and gets HTTP 400 on every upload.
PRESAGE_REV="4b81740610c301cbd54d95f4dbd4cadd692af95e"
PATCH_FILE="$(dirname "$0")/gurk-patch.diff"
INSTALL_DIR="${HOME}/.local/share/openriot/config"
SOURCE_DIR="${HOME}/Code/gurk"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Step 0: Install Rust if cargo is missing
if ! command -v cargo >/dev/null 2>&1; then
    echo "Installing Rust via pkg_add..."
    doas pkg_add rust
fi

# Step 1: Clone or update gurk
echo "Cloning/updating gurk..."
if [ -d "$SOURCE_DIR/.git" ]; then
    if [ -n "$GURK_COMMIT" ]; then
        CURRENT_COMMIT=$(git -C "$SOURCE_DIR" rev-parse HEAD 2>/dev/null || echo "")
        if [ "$CURRENT_COMMIT" = "$GURK_COMMIT" ]; then
            echo "Source already at pinned commit ($GURK_COMMIT), using existing checkout"
        else
            echo "Updating to pinned commit $GURK_COMMIT..."
            git -C "$SOURCE_DIR" fetch origin --tags
            git -C "$SOURCE_DIR" checkout -f "$GURK_COMMIT"
        fi
    else
        echo "Updating to latest origin/main..."
        git -C "$SOURCE_DIR" fetch origin
        git -C "$SOURCE_DIR" reset --hard origin/main
        echo "Source at $(git -C "$SOURCE_DIR" rev-parse --short HEAD)"
    fi
else
    rm -rf "$SOURCE_DIR"
    mkdir -p "$(dirname "$SOURCE_DIR")"
    git clone "$GURK_REPO" "$SOURCE_DIR"
    if [ -n "$GURK_COMMIT" ]; then
        git -C "$SOURCE_DIR" checkout "$GURK_COMMIT"
        echo "Cloned and checked out pinned commit ($GURK_COMMIT)"
    fi
fi

# Step 2: Pin presage to the attachment-upload fix
if [ -f "$SOURCE_DIR/Cargo.toml" ]; then
    sed -i -e "/whisperfish\/presage/s/rev = \"[0-9a-f]\{40\}\"/rev = \"${PRESAGE_REV}\"/" "$SOURCE_DIR/Cargo.toml"
    echo "Pinned presage to ${PRESAGE_REV}"
fi

# Step 3: Apply patch (idempotent)
echo "Applying SIGSEGV fix patch..."
if patch -d "$SOURCE_DIR" -p1 --dry-run -f < "$PATCH_FILE" >/dev/null 2>&1; then
    patch -d "$SOURCE_DIR" -p1 -f < "$PATCH_FILE"
    echo "Patch applied"
elif rg -q "notify-rust disabled on OpenBSD" "$SOURCE_DIR/src/app/message.rs" 2>/dev/null; then
    echo "Patch already applied"
else
    echo "ERROR: Patch cannot be applied and notify-rust is not disabled."
    echo "Upstream code may have changed. Update the patch or the commit pin."
    exit 1
fi

# Step 4: Build with optimizations for smaller binary
echo "Building gurk with LTO and strip..."
export CARGO_PROFILE_RELEASE_LTO=thin
export CARGO_PROFILE_RELEASE_STRIP=debuginfo
PKG_CONFIG=echo \
  cargo install gurk --force \
    --path "$SOURCE_DIR" \
    --root "$INSTALL_DIR"

# Step 5: Strip debug symbols for smaller binary
echo "Stripping debug symbols..."
if [ -f "${INSTALL_DIR}/bin/gurk" ]; then
    strip "${INSTALL_DIR}/bin/gurk" 2>/dev/null || true
fi

# Step 6: Copy binary to source tree
printf "Copy binary to source tree? [y/N] "
read -r answer
if [ "$answer" = "y" ] || [ "$answer" = "Y" ]; then
    install -m 0755 "${INSTALL_DIR}/bin/gurk" "${REPO_ROOT}/config/bin/gurk"
    echo "Copied to ${REPO_ROOT}/config/bin/gurk"
fi

# Step 7: Verify
if [ -x "${INSTALL_DIR}/bin/gurk" ]; then
    echo "Done! Installed to ${INSTALL_DIR}/bin/gurk"
    "${INSTALL_DIR}/bin/gurk" --version
else
    echo "Error: gurk binary not found after build"
    exit 1
fi

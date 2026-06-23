#!/bin/sh
#
# monero.sh - Rebuild Monero GUI wallet with OpenBSD compatibility patches
#
# This script clones monero-gui, applies the OpenBSD patches, builds with
# cmake, and bundles the result into ~/.local/share/openriot/config/monero/
# monero.tgz so that the existing packages.yaml extract step can install it
# on a target machine.
#
# Usage: ./monero.sh
#
# The pre-built monero.tgz is the daily install path; this script is only
# used when an admin needs to rebuild from source (e.g., upstream bumped,
# patch needs revision, OpenBSD ports changed).
#
# Output:
#   ~/.local/share/openriot/config/monero/monero-wallet-gui
#   ~/.local/share/openriot/config/monero/monero.tgz

set -e

MONERO_REPO="https://github.com/monero-project/monero-gui.git"
MONERO_TAG="${MONERO_TAG:-v0.18.5.0}"
MONERO_COMMIT="${MONERO_COMMIT:-}"
PATCH_FILE="$(dirname "$0")/monero-patch.diff"
INSTALL_DIR="${HOME}/.local/share/openriot/config/monero"
SOURCE_DIR="${HOME}/Code/monero-gui"

# Step 1: Clone or update monero-gui
echo "Cloning/updating monero-gui..."
if [ -d "$SOURCE_DIR/.git" ]; then
    if [ -n "$MONERO_COMMIT" ]; then
        CURRENT_COMMIT=$(git -C "$SOURCE_DIR" rev-parse HEAD 2>/dev/null || echo "")
        if [ "$CURRENT_COMMIT" = "$MONERO_COMMIT" ]; then
            echo "Source already at pinned commit ($MONERO_COMMIT), using existing checkout"
        else
            echo "Updating to pinned commit $MONERO_COMMIT..."
            git -C "$SOURCE_DIR" fetch origin
            git -C "$SOURCE_DIR" checkout "$MONERO_COMMIT"
        fi
    else
        echo "Updating to tag $MONERO_TAG..."
        git -C "$SOURCE_DIR" fetch origin --tags
        git -C "$SOURCE_DIR" checkout "$MONERO_TAG"
        echo "Source at $(git -C "$SOURCE_DIR" rev-parse --short HEAD)"
    fi
else
    rm -rf "$SOURCE_DIR"
    mkdir -p "$(dirname "$SOURCE_DIR")"
    git clone "$MONERO_REPO" "$SOURCE_DIR"
    if [ -n "$MONERO_COMMIT" ]; then
        git -C "$SOURCE_DIR" checkout "$MONERO_COMMIT"
        echo "Cloned and checked out pinned commit ($MONERO_COMMIT)"
    else
        git -C "$SOURCE_DIR" checkout "$MONERO_TAG"
        echo "Cloned and checked out tag $MONERO_TAG"
    fi
fi

# Step 2: Apply patch (idempotent)
echo "Applying OpenBSD compatibility patches..."
if patch -d "$SOURCE_DIR" -p1 --dry-run -f < "$PATCH_FILE" >/dev/null 2>&1; then
    patch -d "$SOURCE_DIR" -p1 -f < "$PATCH_FILE"
    echo "Patch applied"
elif grep -q "CMAKE_CXX_STANDARD 17" "$SOURCE_DIR/CMakeLists.txt" 2>/dev/null \
     && grep -q 'CMAKE_SYSTEM_NAME MATCHES "kOpenBSD' "$SOURCE_DIR/CMakeLists.txt" 2>/dev/null; then
    echo "Patch already applied"
else
    echo "ERROR: Patch cannot be applied and OpenBSD fixes are not present."
    echo "Upstream code may have changed. Update the patch or the tag pin."
    exit 1
fi

# Step 3: Build with cmake
echo "Configuring cmake..."
JOBS=$(sysctl -n hw.ncpu 2>/dev/null || echo 1)
mkdir -p "$SOURCE_DIR/build"
cd "$SOURCE_DIR/build"
cmake .. \
    -DCMAKE_BUILD_TYPE=Release \
    -DOPENBSD=TRUE \
    -DCMAKE_CXX_STANDARD_REQUIRED=ON

echo "Building monero-gui with $JOBS jobs..."
make -j"$JOBS"

# Step 4: Install to local bundle directory
echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp "$SOURCE_DIR/build/bin/monero-wallet-gui" "$INSTALL_DIR/monero-wallet-gui"
strip "$INSTALL_DIR/monero-wallet-gui" 2>/dev/null || true

# Step 5: Bundle into monero.tgz so packages.yaml can extract it
echo "Bundling monero.tgz..."
cd "$INSTALL_DIR"
tar -czf monero.tgz monero-wallet-gui

# Step 6: Verify
if [ -x "$INSTALL_DIR/monero-wallet-gui" ] && [ -f "$INSTALL_DIR/monero.tgz" ]; then
    echo "Done!"
    echo "  Binary: $INSTALL_DIR/monero-wallet-gui"
    echo "  Tarball: $INSTALL_DIR/monero.tgz"
    echo "  Size:    $(ls -lh "$INSTALL_DIR/monero-wallet-gui" | awk '{print $5}')"
else
    echo "Error: monero-wallet-gui binary or monero.tgz not found after build"
    exit 1
fi

#!/bin/sh
# OpenRiot ISO Builder
# Downloads OpenBSD -current install ISO and injects OpenRiot autoinstall config + site overlay
set -e

# ---- Config ----
OPENBSD_VERSION="7.9"
ARCH="amd64"
MIRROR="https://cdn.openbsd.org/pub/OpenBSD"
ISO_NAME="install78.iso"
OPENRIOT_VERSION="0.1"

# ---- Paths ----
ROOT="$(cd "$(dirname "$0")" && pwd)"
WORK="$ROOT/.work"
OUT="$ROOT/isos"
DL_DIR="$WORK/dl"
ISO_MNT="$WORK/mnt"
ISO_CONTENTS="$WORK/iso_contents"
SITE_DIR="$ROOT/site"
SITE_TGZ="$WORK/site79.tgz"
AUTOCONF_DIR="$ROOT/autoinstall"

# ---- Cleanup ----
cleanup() {
    echo "Cleaning up..."
    sudo umount "$ISO_MNT" 2>/dev/null || true
    rm -rf "$WORK"
}

trap cleanup EXIT

# ============================================================
# STEP 1: Download OpenBSD install ISO
# ============================================================
echo "=== Step 1: Downloading OpenBSD ${OPENBSD_VERSION} install ISO ==="

mkdir -p "$DL_DIR" "$OUT"

if [ -f "$DL_DIR/$ISO_NAME" ]; then
    echo "ISO already present, skipping download."
else
    echo "Downloading from $MIRROR/snapshots/${ARCH}/${ISO_NAME} ..."
    curl -fL -o "$DL_DIR/$ISO_NAME" \
        "$MIRROR/snapshots/${ARCH}/${ISO_NAME}"
fi

# ============================================================
# STEP 2: Verify ISO with SHA256
# ============================================================
echo "=== Step 2: Verifying ISO integrity ==="

curl -fL -o "$DL_DIR/SHA256" "$MIRROR/snapshots/${ARCH}/SHA256"

# OpenBSD's SHA256 format is "SHA256 (filename) = hash"
# Extract hash for our ISO, compute actual hash, compare
EXPECTED=$(grep "$ISO_NAME" "$DL_DIR/SHA256" | awk '{print $NF}')
ACTUAL=$(sha256sum "$DL_DIR/$ISO_NAME" | awk '{print $1}')
if [ "$EXPECTED" = "$ACTUAL" ]; then
    echo "SHA256 OK"
else
    echo "WARNING: SHA256 verification failed."
    echo "Expected: $EXPECTED"
    echo "Actual:   $ACTUAL"
    exit 1
fi

# ============================================================
# STEP 3: Copy autoinstall config
# ============================================================
echo "=== Step 3: Copying autoinstall config ==="

if [ ! -f "$AUTOCONF_DIR/install.conf" ]; then
    echo "ERROR: autoinstall/install.conf not found"
    exit 1
fi

echo "install.conf:"
cat "$AUTOCONF_DIR/install.conf"

# ============================================================
# STEP 4: Build site79.tgz from site/ directory
# ============================================================
echo "=== Step 4: Building site79.tgz ==="

if [ -d "$SITE_DIR" ] && [ "$(ls -A "$SITE_DIR" 2>/dev/null)" ]; then
    echo "site/ directory found, packing site79.tgz..."
    cd "$SITE_DIR" && tar czf "$SITE_TGZ" .
    echo "site79.tgz created:"
    tar tzf "$SITE_TGZ"
else
    echo "site/ directory is empty or missing — skipping site79.tgz"
fi

# ============================================================
# STEP 5: Mount the ISO read-only
# ============================================================
echo "=== Step 5: Mounting ISO ==="

mkdir -p "$ISO_MNT" "$ISO_CONTENTS"
sudo mount -o ro,loop "$DL_DIR/$ISO_NAME" "$ISO_MNT"

echo "Copying ISO contents..."
cp -a "$ISO_MNT/" "$ISO_CONTENTS/"
sudo umount "$ISO_MNT"

# ============================================================
# STEP 6: Inject install.conf into ISO
# ============================================================
echo "=== Step 6: Injecting install.conf ==="

cp "$AUTOCONF_DIR/install.conf" "$ISO_CONTENTS/install.conf"
echo "install.conf injected"

# ============================================================
# STEP 7: Inject site79.tgz into ISO (if it exists)
# ============================================================
echo "=== Step 7: Injecting site79.tgz ==="

if [ -f "$SITE_TGZ" ]; then
    cp "$SITE_TGZ" "$ISO_CONTENTS/site79.tgz"
    echo "site79.tgz injected"
else
    echo "site79.tgz not found, skipping"
fi

# ============================================================
# STEP 8: Repack the ISO
# ============================================================
echo "=== Step 8: Repacking ISO ==="

OUTPUT="$OUT/openriot-${OPENRIOT_VERSION}.iso"

if command -v xorriso >/dev/null 2>&1; then
    echo "Using xorriso..."
    xorriso -as mkisofs \
        -iso-level 3 \
        -o "$OUTPUT" \
        "$ISO_CONTENTS"
elif command -v mkisofs >/dev/null 2>&1; then
    echo "Using mkisofs..."
    mkisofs -o "$OUTPUT" \
        -iso-level 3 \
        "$ISO_CONTENTS"
else
    echo "ERROR: neither xorriso nor mkisofs found."
    echo "Install with: sudo pkg_add xorriso   (OpenBSD)"
    echo "             sudo apt install xorriso (Debian/Ubuntu)"
    echo "             brew install xorriso     (macOS)"
    exit 1
fi

# ============================================================
# Done
# ============================================================
echo ""
echo "=== Done! ==="
echo "Output: $OUTPUT"
ls -lh "$OUTPUT"

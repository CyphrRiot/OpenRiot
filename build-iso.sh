#!/bin/sh
# OpenRiot ISO Builder
# Builds a bootable OpenBSD ISO with offline packages and OpenRiot autoinstall
#
# Works on Linux and OpenBSD — no sudo required.
# Requires: xorriso, curl, tar
#
# Usage:
#   ./build-iso.sh
#
# Prerequisites:
#   1. scripts/download-packages.sh   — populates ~/.pkgcache/$OPENBSD_VERSION/amd64/
#   2. scripts/generate-index.sh      — creates index.txt (auto-run by step 1)
#   3. make build                      — produces openriot binary in source/

# Note: Don't use set -e - we handle errors explicitly

# ============================================================
# Config
# Read from env (set by `make`) — fall back to defaults if run standalone
# ============================================================
OPENBSD_VERSION="${OPENBSD_VERSION:-7.9}"
ARCH="${ARCH:-amd64}"
OPENRIOT_VERSION="${OPENRIOT_VERSION:-$(cat VERSION 2>/dev/null || echo "0.6")}"
MIRROR="https://cdn.openbsd.org/pub/OpenBSD"

# Derive ISO name from OpenBSD version: 7.9 → install79.iso
_ver_nodot=$(printf '%s' "$OPENBSD_VERSION" | tr -d '.')
ISO_NAME="install${_ver_nodot}.iso"
SITE_TGZ_NAME="site${_ver_nodot}.tgz"

# El Torito boot parameters extracted from the original OpenBSD ISO.
# These are required for the repacked ISO to be bootable (BIOS + UEFI).
BOOT_CATALOG="/${OPENBSD_VERSION}/${ARCH}/boot.catalog"
BOOT_BIOS="/${OPENBSD_VERSION}/${ARCH}/cdbr"
BOOT_EFI="/${OPENBSD_VERSION}/${ARCH}/eficdboot"
VOLUME_ID="OpenBSD/${ARCH}   ${OPENBSD_VERSION} Install CD"

# ============================================================
# Paths
# ============================================================
ROOT="$(cd "$(dirname "$0")" && pwd)"
WORK="$ROOT/.work"
OUT="$ROOT/isos"
DL_DIR="$WORK/dl"
ISO_CONTENTS="$WORK/iso_contents"
SITE_DIR="$ROOT/site"
SITE_TGZ="$WORK/$SITE_TGZ_NAME"
AUTOCONF_DIR="$ROOT/autoinstall"
PKG_CACHE="$HOME/.pkgcache/${OPENBSD_VERSION}/${ARCH}"
PKG_DEST="openriot/packages/${OPENBSD_VERSION}/${ARCH}"

# ============================================================
# Helpers
# ============================================================
log()  { printf '=== %s ===\n' "$*"; }
info() { printf '    %s\n' "$*"; }
die()  { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

# OS-aware SHA256: Linux uses sha256sum, OpenBSD uses sha256
sha256_file() {
    _file="$1"
    case "$(uname -s)" in
        Linux)  sha256sum "$_file" | awk '{print $1}' ;;
        OpenBSD) sha256 -q "$_file" ;;
        *)       sha256sum "$_file" | awk '{print $1}' ;;
    esac
}

# ============================================================
# Cleanup — no sudo needed (xorriso extracts without mounting)
# ============================================================
cleanup() {
    printf 'Cleaning up work directory...\n'
    rm -rf "$ISO_CONTENTS"
    # Keep downloaded ISO and packages — re-downloading is slow
    printf 'Kept: %s\n' "$DL_DIR/$ISO_NAME"
}

# Only cleanup on error (not on successful exit)
trap 'if [ $? -ne 0 ]; then cleanup; fi' EXIT

# ============================================================
# PREFLIGHT: Required tools
# ============================================================
log "Preflight: Checking required tools"

need_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        die "Missing required command: $1
  Install with:
    Arch Linux : sudo pacman -S $2
    Debian/Ubuntu: sudo apt install $3
    OpenBSD   : pkg_add $4"
    fi
    info "$1 — OK"
}

need_cmd xorriso  xorriso         xorriso         xorriso
need_cmd curl     curl            curl            curl
need_cmd tar      tar             tar             tar

# ============================================================
# PREFLIGHT: Offline package cache
# ============================================================
log "Preflight: Checking offline package cache"

if [ ! -d "$PKG_CACHE" ]; then
    die "Package cache not found at $PKG_CACHE
  Run: scripts/download-packages.sh"
fi

PKG_COUNT=$(find "$PKG_CACHE" -maxdepth 1 -name '*.tgz' | wc -l | tr -d ' ')
if [ "$PKG_COUNT" -eq 0 ]; then
    die "No .tgz packages found in $PKG_CACHE
  Run: scripts/download-packages.sh"
fi

if [ ! -f "$PKG_CACHE/index.txt" ]; then
    die "index.txt not found in $PKG_CACHE
  Run: scripts/generate-index.sh $PKG_CACHE"
fi

info "Found $PKG_COUNT packages + index.txt in $PKG_CACHE — OK"

# ============================================================
# PREFLIGHT: autoinstall config
# ============================================================
log "Preflight: Checking autoinstall config"

if [ ! -f "$AUTOCONF_DIR/install.conf" ]; then
    die "autoinstall/install.conf not found"
fi
info "install.conf — OK"

if [ ! -f "$AUTOCONF_DIR/install.site" ]; then
    die "autoinstall/install.site not found"
fi
info "install.site — OK"

# ============================================================
# STEP 1: Download OpenBSD install ISO
# ============================================================
log "Step 1: Downloading OpenBSD ${OPENBSD_VERSION} install ISO"

mkdir -p "$DL_DIR" "$OUT"

if [ -f "$DL_DIR/$ISO_NAME" ]; then
    info "ISO already cached at $DL_DIR/$ISO_NAME — skipping download"
else
    # Try released version first, then snapshot
    RELEASE_URL="$MIRROR/${OPENBSD_VERSION}/${ARCH}/${ISO_NAME}"
    SNAPSHOT_URL="$MIRROR/snapshots/${ARCH}/${ISO_NAME}"

    info "Downloading $RELEASE_URL ..."
    if curl -fL --progress-bar -o "$DL_DIR/$ISO_NAME" "$RELEASE_URL" 2>/dev/null; then
        info "Downloaded released version"
    else
        info "Release not found, trying snapshot: $SNAPSHOT_URL"
        curl -fL --progress-bar \
            -o "$DL_DIR/$ISO_NAME" \
            "$SNAPSHOT_URL" || die "Failed to download ISO"
    fi
    info "Download complete"
fi

# ============================================================
# STEP 2: Verify ISO integrity
# ============================================================
log "Step 2: Verifying ISO integrity (SHA256)"

# Get SHA256 from a given base URL
get_sha256() {
    local base_url="$1"
    local sha256_url="${base_url}/SHA256"
    curl -fsSL -o "$DL_DIR/SHA256" "$sha256_url" 2>/dev/null || return 1
    grep "^SHA256 (${ISO_NAME}) = " "$DL_DIR/SHA256" 2>/dev/null | sed 's/.*= *//' | tr -d ' \n'
}

# Get SHA256.sig (GnuPG signature for secure verification)
get_sha256_sig() {
    local base_url="$1"
    local sig_url="${base_url}/SHA256.sig"
    curl -fsSL -o "$DL_DIR/SHA256.sig" "$sig_url" 2>/dev/null || return 1
}

# Try released version SHA256 first
info "Checking for released version SHA256..."
RELEASE_BASE="$MIRROR/${OPENBSD_VERSION}/${ARCH}"
EXPECTED=$(get_sha256 "$RELEASE_BASE")
info "Release SHA256 result: '${EXPECTED}'"

# If not found, try snapshot
if [ -z "$EXPECTED" ]; then
    info "Release SHA256 not found, trying snapshot..."
    SNAPSHOT_BASE="$MIRROR/snapshots/${ARCH}"
    EXPECTED=$(get_sha256 "$SNAPSHOT_BASE")
    info "Snapshot SHA256 result: '${EXPECTED}'"
fi

if [ -z "$EXPECTED" ]; then
    die "Could not find SHA256 for $ISO_NAME"
fi

# Download SHA256.sig for secure install verification
info "Downloading SHA256.sig..."
if ! get_sha256_sig "$RELEASE_BASE" 2>/dev/null; then
    get_sha256_sig "$SNAPSHOT_BASE" 2>/dev/null || true
fi
if [ -f "$DL_DIR/SHA256.sig" ]; then
    info "SHA256.sig downloaded successfully"
else
    info "WARNING: SHA256.sig not available — continuing without signature verification"
fi

info "Computing SHA256 of downloaded ISO..."
ACTUAL=$(sha256_file "$DL_DIR/$ISO_NAME")

if [ "$EXPECTED" = "$ACTUAL" ]; then
    info "SHA256 OK: $ACTUAL"
else
    printf 'SHA256 MISMATCH\n'
    printf '  Expected: %s\n' "$EXPECTED"
    printf '  Actual:   %s\n' "$ACTUAL"
    info "Deleting corrupted ISO and retrying..."
    rm -f "$DL_DIR/$ISO_NAME"

    # Retry download
    SNAPSHOT_URL="$MIRROR/snapshots/${ARCH}/${ISO_NAME}"
    info "Downloading $SNAPSHOT_URL ..."
    curl -fL --progress-bar \
        -o "$DL_DIR/$ISO_NAME" \
        "$SNAPSHOT_URL" || die "Failed to download ISO"

    # Recheck SHA256
    ACTUAL=$(sha256_file "$DL_DIR/$ISO_NAME")
    if [ "$EXPECTED" = "$ACTUAL" ]; then
        info "SHA256 OK on retry: $ACTUAL"
    else
        die "ISO integrity check failed again — deleting corrupted download"
    fi
fi

# ============================================================
# STEP 3: Extract ISO contents (xorriso — no sudo, no loop mount)
# ============================================================
log "Step 3: Extracting ISO contents"

rm -rf "$ISO_CONTENTS"
mkdir -p "$ISO_CONTENTS"

info "Extracting with xorriso (no root required)..."
xorriso \
    -osirrox on \
    -indev "stdio:$DL_DIR/$ISO_NAME" \
    -extract / "$ISO_CONTENTS" \
    2>&1 | grep -E "^(xorriso : UPDATE|Extracted)" || true

info "ISO contents extracted to $ISO_CONTENTS"
info "Top-level files:"
ls "$ISO_CONTENTS" | sed 's/^/    /'

# Remove unnecessary sets (not needed for desktop Wayland install)
info "Removing game79.tgz, xserv79.tgz (not needed for Wayland desktop)..."
rm -f "$ISO_CONTENTS/${OPENBSD_VERSION}/${ARCH}/game79.tgz"
rm -f "$ISO_CONTENTS/${OPENBSD_VERSION}/${ARCH}/xserv79.tgz"
# Keep xbase79.tgz — Sway needs X11 libs for Xwayland

# Explicitly remove SHA256.sig from the ISO contents.
# We regenerate SHA256 to include our site79.tgz, which invalidates the
# original OpenBSD signature. If SHA256.sig is present, the installer
# verifies SHA256 against it, fails, then can't trust any checksum —
# causing "Checksum test for site79.tgz failed" prompts for the user.
# Without SHA256.sig the installer falls back to plain SHA256 matching,
# which works correctly because we regenerate it in Step 7.
rm -f "$ISO_CONTENTS/${OPENBSD_VERSION}/${ARCH}/SHA256.sig"
info "SHA256.sig removed (we regenerate SHA256 — sig would be invalid)"





# ============================================================
# STEP 4: Build site79.tgz
# ============================================================
log "Step 4: Building site79.tgz"

# site79.tgz is extracted by the OpenBSD installer into the new system root.
# Files in site/ map directly — e.g. site/etc/openriot/ → /etc/openriot/
# install.site is placed at the top level and runs automatically post-install.

if [ -d "$SITE_DIR" ] && [ "$(ls -A "$SITE_DIR" 2>/dev/null)" ]; then
    info "Packing site/ → site79.tgz ..."
    # Must cd into site/ so tar paths are relative (no leading ./)
    (cd "$SITE_DIR" && tar czf "$SITE_TGZ" .)
    info "site79.tgz contents:"
    tar tzf "$SITE_TGZ" | sed 's/^/    /'
else
    info "site/ is empty — skipping site79.tgz"
fi

# Always include install.site from autoinstall/
# It must be at the root of site79.tgz so the installer finds it
info "Injecting install.site into site79.tgz ..."
TMPSITE="$WORK/tmpsite"
rm -rf "$TMPSITE"
mkdir -p "$TMPSITE"

# Unpack existing site79.tgz if present, then add install.site
if [ -f "$SITE_TGZ" ]; then
    tar xzf "$SITE_TGZ" -C "$TMPSITE"
fi
cp "$AUTOCONF_DIR/install.site" "$TMPSITE/install.site"
chmod 0755 "$TMPSITE/install.site"

# Bundle a clean git archive of the repo for offline first-boot
# This goes to /etc/openriot/repo.tar.gz on the target system.
# Use --prefix=openriot/ so extraction lands at ~/.local/share/openriot/
info "Creating clean git archive for offline use..."
mkdir -p "$TMPSITE/etc/openriot"
git archive --prefix=openriot/ HEAD | gzip -n > "$TMPSITE/etc/openriot/repo.tar.gz"
info "Repo archive created ($(du -h "$TMPSITE/etc/openriot/repo.tar.gz" | cut -f1))"



# Copy packages.yaml for offline install (read by install.site before tarball extraction)
cp "$ROOT/install/packages.yaml" "$TMPSITE/etc/openriot/packages.yaml"
info "packages.yaml bundled"

# Copy VERSION for offline use (openriot-update.sh and binary rely on it)
# NOTE: This must be the OpenBSD version (7.9), not OpenRiot version (0.9)
echo "${OPENBSD_VERSION}" > "$TMPSITE/etc/openriot/VERSION"
info "VERSION bundled as ${OPENBSD_VERSION}"

# Copy openriot binary for offline install
if [ -f "$ROOT/install/openriot" ]; then
    cp "$ROOT/install/openriot" "$TMPSITE/etc/openriot/openriot"
    chmod 0755 "$TMPSITE/etc/openriot/openriot"
    info "openriot binary bundled ($(du -h "$ROOT/install/openriot" | cut -f1))"
else
    info "WARNING: openriot binary not found at install/openriot — run 'make build' first"
fi

	# Pre-fetch wlsunset source for offline build
	_wlsunset_tmp="$WORK/wlsunset-src"
	rm -rf "$_wlsunset_tmp"
	if git clone --depth=1 https://git.sr.ht/~kennylevinsen/wlsunset "$_wlsunset_tmp" 2>/dev/null; then
		(cd "$_wlsunset_tmp" && git archive HEAD | gzip -n > "$TMPSITE/etc/openriot/wlsunset.tar.gz")
		info "wlsunset source bundled ($(du -h "$TMPSITE/etc/openriot/wlsunset.tar.gz" | cut -f1))"
	else
		info "wlsunset clone failed — will build from source during openriot install"
	fi
	rm -rf "$_wlsunset_tmp"

# Bundle ALL packages directly into site79.tgz (no separate openriot.tgz)
info "Copying $PKG_COUNT packages into site79.tgz..."
mkdir -p "$TMPSITE/openriot/packages/${OPENBSD_VERSION}/${ARCH}"
cp "$PKG_CACHE"/*.tgz "$TMPSITE/openriot/packages/${OPENBSD_VERSION}/${ARCH}/"
cp "$PKG_CACHE/index.txt" "$TMPSITE/openriot/packages/${OPENBSD_VERSION}/${ARCH}/"
info "Packages bundled into site79.tgz"

(cd "$TMPSITE" && tar czf "$SITE_TGZ" .)
rm -rf "$TMPSITE"
info "site79.tgz ready (with all packages embedded)"

# Ensure index.txt includes site79.tgz (installer sometimes needs this)
if [ -f "$SETS_DIR/index.txt" ]; then
    if ! grep -q "site${_ver_nodot}.tgz" "$SETS_DIR/index.txt"; then
        echo "site${_ver_nodot}.tgz" >> "$SETS_DIR/index.txt"
        info "Added site${_ver_nodot}.tgz to index.txt"
    fi
fi

# ============================================================
# STEP 5: Inject install.conf and autopartitionning template
# ============================================================
log "Step 5: Injecting install.conf and autopartitionning template"

# OpenBSD autoinstall requires auto_install.conf (with underscore) as the primary file
# Copy instead of symlink to avoid Joliet tree issues
cp "$AUTOCONF_DIR/install.conf" "$ISO_CONTENTS/auto_install.conf"
cp "$AUTOCONF_DIR/install.conf" "$ISO_CONTENTS/install.conf"
info "auto_install.conf and install.conf injected"

cp "$AUTOCONF_DIR/autopartitionning.template" "$ISO_CONTENTS/autopartitionning.template"
info "autopartitionning.template injected at ISO root"

# ============================================================
# STEP 6: Inject site79.tgz into ISO contents
# ============================================================
log "Step 6: Injecting site79.tgz"

if [ -f "$SITE_TGZ" ]; then
    cp "$SITE_TGZ" "$ISO_CONTENTS/${OPENBSD_VERSION}/${ARCH}/${SITE_TGZ_NAME}"
    info "${SITE_TGZ_NAME} injected at ${OPENBSD_VERSION}/${ARCH}/${SITE_TGZ_NAME}"
else
    info "No site79.tgz — skipping"
fi

# ============================================================
# STEP 7: Regenerate SHA256 for modified sets
# ============================================================
log "Step 7: Regenerating SHA256 for modified sets"

SETS_DIR="$ISO_CONTENTS/${OPENBSD_VERSION}/${ARCH}"
rm -f "$SETS_DIR/SHA256"
for f in "$SETS_DIR"/*.tgz; do
    [ -f "$f" ] || continue
    h=$(sha256_file "$f")
    fname=$(basename "$f")
    echo "SHA256 ($fname) = $h"
done > "$SETS_DIR/SHA256"
info "SHA256 regenerated with $(wc -l < "$SETS_DIR/SHA256" | tr -d ' ') entries"

# ============================================================
# STEP 8: Clean up old artifacts
# ============================================================
log "Step 8: Cleaning up old artifacts"

# Remove any leftover openriot/ directory from old builds
rm -rf "$ISO_CONTENTS/openriot"

# openriot.tgz is now included inside site79.tgz (extracted to / during install)
info "openriot.tgz is included in site79.tgz"

# ============================================================
# STEP 8: Repack into bootable ISO
# ============================================================
log "Step 8: Repacking bootable ISO"

OUTPUT="$OUT/openriot.iso"
mkdir -p "$OUT"

info "Volume ID : $VOLUME_ID"
info "BIOS boot : $BOOT_BIOS"
info "EFI  boot : $BOOT_EFI"
info "Output    : $OUTPUT"
printf '\n'

# El Torito parameters are taken directly from the original OpenBSD ISO
# via: xorriso -indev install79.iso -report_el_torito as_mkisofs
#
# -iso-level 3   — allows files > 2GB (packages can be large)
# -r             — Rock Ridge (preserves Unix permissions/symlinks)
# -J             — Joliet (Windows compat — harmless)
# -c             — boot catalog location
# -b             — BIOS El Torito boot image
# -e             — EFI El Torito boot image
# -no-emul-boot  — no floppy emulation (required for both BIOS and EFI)

xorriso -as mkisofs \
    -iso-level 3 \
    -r \
    -J \
    -V "$VOLUME_ID" \
    -c "$BOOT_CATALOG" \
    -b "$BOOT_BIOS" \
    -no-emul-boot \
    -boot-load-size 4 \
    -eltorito-alt-boot \
    -e "$BOOT_EFI" \
    -no-emul-boot \
    -boot-load-size 700 \
    -o "$OUTPUT" \
    "$ISO_CONTENTS" \
    2>&1 | grep -v "^$" | sed 's/^/    /' || die "xorriso repack failed"

# ============================================================
# Done
# ============================================================
printf '\n'
log "Build complete"
printf '  Output : %s\n' "$OUTPUT"
printf '  Size   : %s\n' "$(du -sh "$OUTPUT" | cut -f1)"
printf '\n'
printf 'Build complete!\n'
printf '  Output : %s\n' "$OUTPUT"
printf '  Size   : %s\n' "$(du -sh "$OUTPUT" | cut -f1)"
printf '\n'
printf 'Next steps:\n'
printf '  make isotest   — build and test in QEMU\n'
printf '  ./test-iso.sh  — test directly without make\n'
printf '\n'

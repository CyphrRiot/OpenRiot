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

set -e

# ============================================================
# Config
# Read from env (set by `make`) — fall back to defaults if run standalone
# ============================================================
OPENBSD_VERSION="${OPENBSD_VERSION:-7.9}"
ARCH="${ARCH:-amd64}"
OPENRIOT_VERSION="${OPENRIOT_VERSION:-0.1}"
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

trap cleanup EXIT

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
    info "Downloading $MIRROR/snapshots/${ARCH}/${ISO_NAME} ..."
    curl -fL --progress-bar \
        -o "$DL_DIR/$ISO_NAME" \
        "$MIRROR/snapshots/${ARCH}/${ISO_NAME}"
    info "Download complete"
fi

# ============================================================
# STEP 2: Verify ISO integrity
# ============================================================
log "Step 2: Verifying ISO integrity (SHA256)"

info "Fetching SHA256 manifest..."
curl -fsSL -o "$DL_DIR/SHA256" "$MIRROR/snapshots/${ARCH}/SHA256"

# OpenBSD SHA256 manifest format: "SHA256 (filename) = hash"
EXPECTED=$(grep "^SHA256 (${ISO_NAME}) = " "$DL_DIR/SHA256" \
    | sed 's/.*= *//' | tr -d ' \n')

if [ -z "$EXPECTED" ]; then
    die "Could not find SHA256 entry for $ISO_NAME in manifest"
fi

info "Computing SHA256 of downloaded ISO..."
ACTUAL=$(sha256_file "$DL_DIR/$ISO_NAME")

if [ "$EXPECTED" = "$ACTUAL" ]; then
    info "SHA256 OK: $ACTUAL"
else
    printf 'SHA256 MISMATCH\n'
    printf '  Expected: %s\n' "$EXPECTED"
    printf '  Actual:   %s\n' "$ACTUAL"
    die "ISO integrity check failed — deleting corrupted download"
    rm -f "$DL_DIR/$ISO_NAME"
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
# This goes to /etc/openriot/repo.tar.gz on the target system
info "Creating clean git archive for offline use..."
mkdir -p "$TMPSITE/etc/openriot"
git archive HEAD | gzip -n > "$TMPSITE/etc/openriot/repo.tar.gz"
info "Repo archive created ($(du -h "$TMPSITE/etc/openriot/repo.tar.gz" | cut -f1))"

	# Pre-fetch wlsunset source for offline build
	info "Fetching wlsunset source for offline build..."
	_wlsunset_tmp="$WORK/wlsunset-src"
	rm -rf "$_wlsunset_tmp"
	if git clone --depth=1 https://git.sr.ht/~kennylevinsen/wlsunset "$_wlsunset_tmp" 2>/dev/null; then
		(cd "$_wlsunset_tmp" && git archive HEAD | gzip -n > "$TMPSITE/etc/openriot/wlsunset.tar.gz")
		info "wlsunset source bundled ($(du -h "$TMPSITE/etc/openriot/wlsunset.tar.gz" | cut -f1))"
	else
		info "wlsunset clone failed — will build from source during openriot install"
	fi
	rm -rf "$_wlsunset_tmp"

(cd "$TMPSITE" && tar czf "$SITE_TGZ" .)
rm -rf "$TMPSITE"
info "site79.tgz ready"

# ============================================================
# STEP 5: Inject install.conf into ISO contents
# ============================================================
log "Step 5: Injecting install.conf"

cp "$AUTOCONF_DIR/install.conf" "$ISO_CONTENTS/install.conf"
info "install.conf injected at ISO root"

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
# STEP 7: Copy offline packages into ISO contents
# ============================================================
log "Step 7: Copying offline packages into ISO"

PKG_ISO_DIR="$ISO_CONTENTS/$PKG_DEST"
mkdir -p "$PKG_ISO_DIR"

info "Copying $PKG_COUNT packages → $PKG_DEST ..."
cp "$PKG_CACHE"/*.tgz "$PKG_ISO_DIR/"
cp "$PKG_CACHE/index.txt" "$PKG_ISO_DIR/index.txt"

info "Package repo written:"
ls -lh "$PKG_ISO_DIR" | tail -6 | sed 's/^/    /'
info "  ... ($PKG_COUNT packages + index.txt)"

# ============================================================
# STEP 8: Repack into bootable ISO
# ============================================================
log "Step 8: Repacking bootable ISO"

OUTPUT="$OUT/openriot-${OPENRIOT_VERSION}.iso"
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
printf 'Test with QEMU:\n'
printf '  qemu-system-x86_64 -cdrom %s -m 2G -enable-kvm\n' "$OUTPUT"
printf '\n'

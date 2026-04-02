#!/bin/sh
# OpenRiot Package Downloader
# Downloads OpenBSD packages for offline installation
# Creates local package cache from install/packages.yaml
# Usage: scripts/download-packages.sh [--dry-run]
#
# Must be run from the project root directory.

set -e

# ---- Config ----
# Read from env (set by `make`) — fall back to defaults if run standalone
OPENBSD_VERSION="${OPENBSD_VERSION:-7.9}"
ARCH="${ARCH:-amd64}"
MIRROR="https://cdn.openbsd.org/pub/OpenBSD"
OPENBSD_PKG_PATH="${MIRROR}/snapshots/packages/${ARCH}"

# Cache dir matches what build-iso.sh expects: ~/.pkgcache/7.9/amd64/
CACHE_DIR="$HOME/.pkgcache/${OPENBSD_VERSION}/${ARCH}"
PKG_LIST_FILE="$CACHE_DIR/packages.txt"

# ---- Args ----
DRY_RUN="no"
if [ "${1:-}" = "--dry-run" ]; then
    DRY_RUN="yes"
fi

# ---- Colors (POSIX-safe: printf, not echo -e) ----
red()    { printf '\033[0;31m%s\033[0m\n' "$*"; }
green()  { printf '\033[0;32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[1;33m%s\033[0m\n' "$*"; }
log()    { printf '[DOWNLOAD] %s\n' "$*"; }

# ---- Preflight ----
need_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        red "ERROR: Missing required command: $1"
        exit 1
    fi
}

need_cmd curl
need_cmd awk
need_cmd sed

printf '=== OpenRiot Package Downloader ===\n'
printf 'Cache:  %s\n' "$CACHE_DIR"
printf 'Mirror: %s\n' "$OPENBSD_PKG_PATH"
[ "$DRY_RUN" = "yes" ] && yellow "[DRY RUN MODE — no files will be written]"
printf '\n'

# ---- Verify we are in the project root ----
if [ ! -f "install/packages.yaml" ]; then
    red "ERROR: install/packages.yaml not found"
    red "Run this script from the OpenRiot project root."
    exit 1
fi

# ---- Create cache directory ----
mkdir -p "$CACHE_DIR"

# ============================================================
# STEP 1: Parse package names from packages.yaml
# ============================================================
yellow "Parsing packages from install/packages.yaml..."

# Extract package names from all "packages:" blocks.
# Strategy: track when we are inside a packages: list by indent level.
# When a sibling key appears at the same indent, stop collecting.
# Only accept lines that look like valid package names.
awk '
/^[a-z]/ {
    in_pkg = 0
    pkg_indent = 0
}
/^[[:space:]]{4}[a-z]/ {
    in_pkg = 0
    pkg_indent = 0
}
/^[[:space:]]+packages:/ {
    in_pkg = 1
    match($0, /^[[:space:]]+/)
    pkg_indent = RLENGTH
    next
}
/^[[:space:]]+[a-z]+:/ {
    if (in_pkg) {
        match($0, /^[[:space:]]+/)
        if (RLENGTH <= pkg_indent) {
            in_pkg = 0
        }
    }
}
/^[[:space:]]+-[[:space:]]/ {
    if (in_pkg) {
        line = $0
        sub(/^[[:space:]]*-[[:space:]]+/, "", line)
        sub(/#.*/, "", line)
        gsub(/"/, "", line)
        # Trim trailing whitespace
        sub(/[[:space:]]+$/, "", line)
        # Only keep valid package names
        if (line ~ /^[a-zA-Z][a-zA-Z0-9_-]*$/) {
            print line
        }
    }
}
' install/packages.yaml | sort -u > "$PKG_LIST_FILE"

PKG_COUNT=$(wc -l < "$PKG_LIST_FILE" | tr -d ' ')
printf 'Found %s packages\n' "$PKG_COUNT"
printf '\n'

if [ "$PKG_COUNT" -eq 0 ]; then
    red "ERROR: No packages parsed from install/packages.yaml"
    exit 1
fi

if [ "$DRY_RUN" = "yes" ]; then
    yellow "Packages that would be downloaded:"
    cat "$PKG_LIST_FILE" | sed 's/^/  /'
    printf '\n'
fi

# ============================================================
# STEP 2: Fetch mirror index (once, reuse for all lookups)
# ============================================================
yellow "Fetching mirror package index..."

MIRROR_INDEX_FILE="$CACHE_DIR/.mirror-index.html"
MIRROR_WAYLAND_FILE="$CACHE_DIR/.mirror-wayland-index.html"

if [ "$DRY_RUN" = "no" ]; then
    # Root packages index
    if ! curl -fsSL -o "$MIRROR_INDEX_FILE" "${OPENBSD_PKG_PATH}/"; then
        red "ERROR: Could not fetch mirror index from ${OPENBSD_PKG_PATH}/"
        exit 1
    fi

    # wayland/ subdirectory index (waybar, wlsunset, etc. live here)
    if ! curl -fsSL -o "$MIRROR_WAYLAND_FILE" "${OPENBSD_PKG_PATH}/wayland/"; then
        yellow "WARNING: Could not fetch wayland/ subdirectory index — skipping"
        > "$MIRROR_WAYLAND_FILE"
    fi

    log "Mirror index fetched ($(wc -c < "$MIRROR_INDEX_FILE" | tr -d ' ') bytes)"
    log "Wayland index fetched ($(wc -c < "$MIRROR_WAYLAND_FILE" | tr -d ' ') bytes)"
else
    # In dry-run mode, create empty placeholders
    > "$MIRROR_INDEX_FILE"
    > "$MIRROR_WAYLAND_FILE"
    yellow "[DRY RUN] Skipping mirror index fetch"
fi

printf '\n'

# ---- Helper: find a package filename in the cached index ----
# Returns the versioned .tgz filename, or empty string if not found.
# Searches root index first, then wayland/ subdirectory.
#
# OpenBSD mirror HTML has links like:
#   <a href="foo-1.2p3.tgz">foo-1.2p3.tgz</a>
# We extract with sed (POSIX, no -P flag).
find_package_file() {
    _pkg="$1"
    _result=""

    # Match: href="PKG-VERSION.tgz" — version must start with a digit
    # sed extracts the first match from the href attribute value
    _result=$(sed 's/href="/\n/g' "$MIRROR_INDEX_FILE" \
        | grep "^${_pkg}-[0-9][^\"]*\.tgz" \
        | sed 's/".*//' \
        | head -1)

    if [ -n "$_result" ]; then
        printf '%s' "$_result"
        return 0
    fi

    # Fallback: check wayland/ subdirectory
    _result=$(sed 's/href="/\n/g' "$MIRROR_WAYLAND_FILE" \
        | grep "^${_pkg}-[0-9][^\"]*\.tgz" \
        | sed 's/".*//' \
        | head -1)

    if [ -n "$_result" ]; then
        # Prefix with wayland/ so the download URL is correct
        printf 'wayland/%s' "$_result"
        return 0
    fi

    return 0  # empty string = not found
}

# ============================================================
# STEP 3: Download each package
# ============================================================
yellow "Downloading packages..."
printf '\n'

DOWNLOADED=0
SKIPPED=0
FAILED=0
FAILED_LIST=""

while IFS= read -r pkg; do
    [ -z "$pkg" ] && continue

    # Locate versioned filename on mirror
    PKG_FILE=$(find_package_file "$pkg")

    if [ -z "$PKG_FILE" ]; then
        printf '  '
        red "[FAIL] $pkg — not found on mirror (root or wayland/)"
        FAILED=$((FAILED + 1))
        FAILED_LIST="$FAILED_LIST $pkg"
        continue
    fi

    # Basename for local cache (strip any subdir prefix)
    PKG_BASENAME=$(basename "$PKG_FILE")

    if [ "$DRY_RUN" = "yes" ]; then
        printf '  '
        yellow "[DRY RUN] $pkg → $PKG_FILE"
        DOWNLOADED=$((DOWNLOADED + 1))
        continue
    fi

    # Skip if already cached
    if [ -f "$CACHE_DIR/$PKG_BASENAME" ]; then
        printf '  '
        yellow "[SKIP] $pkg ($PKG_BASENAME already cached)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # Download
    if curl -fsSL -o "$CACHE_DIR/$PKG_BASENAME" "${OPENBSD_PKG_PATH}/${PKG_FILE}"; then
        printf '  '
        green "[OK]   $pkg → $PKG_BASENAME"
        DOWNLOADED=$((DOWNLOADED + 1))
    else
        printf '  '
        red "[FAIL] $pkg — download failed"
        rm -f "$CACHE_DIR/$PKG_BASENAME"
        FAILED=$((FAILED + 1))
        FAILED_LIST="$FAILED_LIST $pkg"
    fi
done < "$PKG_LIST_FILE"

printf '\n'
printf '=== Download Summary ===\n'
green "Downloaded: $DOWNLOADED"
yellow "Skipped:    $SKIPPED"
[ "$FAILED" -gt 0 ] && red "Failed:     $FAILED" || printf 'Failed:     0\n'

if [ -n "$FAILED_LIST" ]; then
    printf '\nFailed packages:\n'
    for f in $FAILED_LIST; do
        printf '  - %s\n' "$f"
    done
fi

printf '\n'

# ============================================================
# STEP 4: Generate index.txt for pkg_add
# ============================================================
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
GENERATE_INDEX="$SCRIPT_DIR/generate-index.sh"

if [ "$DRY_RUN" = "yes" ]; then
    yellow "[DRY RUN] Skipping index generation"
    printf '\n'
    printf 'Dry run complete. Run without --dry-run to download.\n'
    exit 0
fi

if [ ! -x "$GENERATE_INDEX" ]; then
    yellow "WARNING: generate-index.sh not found or not executable at $GENERATE_INDEX"
    yellow "Run it manually: scripts/generate-index.sh $CACHE_DIR"
else
    yellow "Generating index.txt for offline repo..."
    printf '\n'
    "$GENERATE_INDEX" "$CACHE_DIR"
fi

printf '\n'
printf '=== All done! ===\n'
printf 'Cache: %s\n' "$CACHE_DIR"
TOTAL=$(find "$CACHE_DIR" -maxdepth 1 -name '*.tgz' | wc -l | tr -d ' ')
printf 'Total packages cached: %s\n' "$TOTAL"
printf '\n'
printf 'Next step: make build && ./build-iso.sh\n'

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi

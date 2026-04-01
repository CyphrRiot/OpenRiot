#!/bin/sh
# OpenRiot Package Downloader
# Downloads OpenBSD packages for offline installation
# Creates local package cache from install/packages.yaml

set -e

OPENBSD_VERSION="7.9"
ARCH="amd64"
MIRROR="https://cdn.openbsd.org/pub/OpenBSD"
CACHE_DIR="$HOME/.pkgcache"
PKG_LIST="$CACHE_DIR/packages.txt"
OPENBSD_PKG_PATH="${MIRROR}/snapshots/packages/${ARCH}"

# Colors for output (use $'...' for escape code interpretation in POSIX sh)
RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
NC=$'\033[0m' # No Color

echo "=== OpenRiot Package Downloader ==="
echo ""

# Create cache directory
mkdir -p "$CACHE_DIR"

# Check for required commands
need_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "${RED}ERROR: Missing required command: $1${NC}"
        exit 1
    fi
}

need_cmd curl
need_cmd awk
need_cmd grep

# Parse packages from install/packages.yaml
echo "${YELLOW}Parsing packages from install/packages.yaml...${NC}"

if [ ! -f "install/packages.yaml" ]; then
    echo "${RED}ERROR: install/packages.yaml not found${NC}"
    exit 1
fi

# Parse packages.yaml - extract only lines under "packages:" blocks
# Stop when we hit a sibling key (configs:, commands:, depends:, build:) at same or higher indent
awk '
/^[a-z]/ {
    # Top-level key (core:, system:, etc.) - reset state
    in_pkg = 0
    pkg_indent = 0
}
/^[[:space:]]{4}[a-z]+:/ {
    # Module-level key (sway:, apps:, etc.) - reset state
    in_pkg = 0
    pkg_indent = 0
}
/^[[:space:]]+packages:/ {
    # Found a packages section
    in_pkg = 1
    # Count indent level (how many spaces before "packages:")
    match($0, /^[[:space:]]+/, arr)
    pkg_indent = length(arr[0])
    next
}
/^[[:space:]]+[a-z]+:/ {
    # Found a sibling key (configs:, commands:, depends:, build:)
    # If at or above packages indent, exit packages mode
    if (in_pkg) {
        match($0, /^[[:space:]]+/, arr)
        this_indent = length(arr[0])
        if (this_indent <= pkg_indent) {
            in_pkg = 0
        }
    }
}
/^[[:space:]]+-\s+/ {
    if (in_pkg) {
        sub(/^[[:space:]]+-\s+/, "")
        sub(/#.*/, "")
        gsub(/"/, "")
        # Only keep if it looks like a package name (alphanumeric, hyphens, underscores)
        if (/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/) {
            print
        }
    }
}
' install/packages.yaml | sort -u > "$PKG_LIST"

PACKAGE_COUNT=$(wc -l < "$PKG_LIST" | tr -d ' ')
echo "Found ${GREEN}${PACKAGE_COUNT}${NC} packages"

if [ "$PACKAGE_COUNT" -eq 0 ]; then
    echo "${RED}ERROR: No packages found in install/packages.yaml${NC}"
    exit 1
fi

echo ""
echo "Cache directory: ${GREEN}${CACHE_DIR}${NC}"
echo "Package list: ${GREEN}${PKG_LIST}${NC}"
echo ""

# Download packages
DOWNLOADED=0
SKIPPED=0
FAILED=0

echo "${YELLOW}Downloading packages...${NC}"
echo ""

while IFS= read -r pkg; do
    # Skip empty lines and comments
    [ -z "$pkg" ] && continue
    echo "$pkg" | grep -q '^#' && continue

    # Try to find the actual package file on the mirror
    # OpenBSD packages have versioned filenames like "package-1.0p0.tgz"
    PACKAGE_FILE=$(curl -sfL "${OPENBSD_PKG_PATH}/" | grep -oP "href=\"\K${pkg}-[0-9][^\"]+\.tgz(?=\")" | head -1)

    if [ -z "$PACKAGE_FILE" ]; then
        echo "  ${RED}[FAIL]${NC} $pkg - not found on mirror"
        FAILED=$((FAILED + 1))
        continue
    fi

    # Check if already cached
    if [ -f "$CACHE_DIR/$PACKAGE_FILE" ]; then
        echo "  ${YELLOW}[SKIP]${NC} $pkg (already cached)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # Download package
    if curl -fsL -o "$CACHE_DIR/$PACKAGE_FILE" "${OPENBSD_PKG_PATH}/${PACKAGE_FILE}"; then
        echo "  ${GREEN}[OK]${NC}   $pkg"
        DOWNLOADED=$((DOWNLOADED + 1))
    else
        echo "  ${RED}[FAIL]${NC} $pkg (download failed)"
        rm -f "$CACHE_DIR/$PACKAGE_FILE"
        FAILED=$((FAILED + 1))
    fi
done < "$PKG_LIST"

echo ""
echo "=== Download Complete ==="
echo "Downloaded: ${GREEN}${DOWNLOADED}${NC}"
echo "Skipped:   ${YELLOW}${SKIPPED}${NC}"
echo "Failed:    ${RED}${FAILED}${NC}"
echo ""
echo "Cached packages: $(ls $CACHE_DIR/*.tgz 2>/dev/null | wc -l | tr -d ' ')"

if [ $FAILED -gt 0 ]; then
    exit 1
fi

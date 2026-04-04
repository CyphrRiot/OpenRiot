#!/bin/sh
# OpenRiot Package Index Generator
# Generates index.txt for local OpenBSD package repository
# pkg_add needs index.txt for dependency resolution

set -e

# Read from env (set by `make`) — fall back to defaults if run standalone
OPENBSD_VERSION="${OPENBSD_VERSION:-7.9}"
ARCH="${ARCH:-amd64}"

# Default to ~/.pkgcache/VERSION/ARCH/ if no argument given
PKG_DIR="${1:-$HOME/.pkgcache/${OPENBSD_VERSION}/${ARCH}}"
INDEX_FILE="$PKG_DIR/index.txt"

echo "=== OpenRiot Package Index Generator ==="
echo "Package directory: $PKG_DIR"
echo "Index file: $INDEX_FILE"

# Check if package directory exists
if [ ! -d "$PKG_DIR" ]; then
    echo "ERROR: Package directory not found: $PKG_DIR"
    echo "Run scripts/download-packages.sh first"
    exit 1
fi

# Create or truncate index file
> "$INDEX_FILE"

# Count packages
count=0

# Generate index entries for each .tgz package
# Format for pkg_add offline repo: PACKAGE_NAME: SHA256:hash:size:comment
count=0

for pkg in "$PKG_DIR"/*.tgz; do
    [ -f "$pkg" ] || continue

    # Get basename without extension
    name=$(basename "$pkg" .tgz)

    # Calculate SHA256 for package integrity
    case "$(uname -s)" in
        Linux)  sha256=$(sha256sum "$pkg" | awk '{print $1}') ;;
        OpenBSD) sha256=$(sha256 -q "$pkg") ;;
        *)       sha256=$(sha256sum "$pkg" | awk '{print $1}') ;;
    esac

    # Get file size
    size=$(stat -c%s "$pkg" 2>/dev/null || stat -f%z "$pkg" 2>/dev/null)

    # Extract comment from package basename (version suffix after last hyphen)
    comment="${name##*-}"

    # Write in pkg_add format: name: SHA256:hash:size:comment
    echo "${name}: SHA256:${sha256}:${size}:${comment}" >> "$INDEX_FILE"
    count=$((count + 1))
done

echo "Generated index with $count packages"
echo "Done: $INDEX_FILE"

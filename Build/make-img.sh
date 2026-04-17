#!/bin/sh
# make-img.sh - Build OpenRiot custom installer image
# Must be run on OpenBSD 7.9
#
# Usage:
#   ./make-img.sh              # Full build
#   ./make-img.sh site         # Create openriot.tgz only
#   ./make-img.sh test         # Test in QEMU (if available)

set -e

# Change to script directory for relative paths
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# ============================================================
# Config
# ============================================================
OPENBSD_VERSION=79
OPENRIOT_VERSION=$(cat ../VERSION 2>/dev/null || echo "0.0.0")
BASE_IMG="${BASE_IMG:-$(pwd)/Images/install79.img}"
OUTPUT_IMG="${OUTPUT_IMG:-$(pwd)/Images/openriot.img}"
WORK_DIR="${WORK_DIR:-$(pwd)/work}"
OPENRIOT_TGZ="${WORK_DIR}/openriot.tgz"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo "${BLUE}[INFO]${NC} $1"; }
warn() { echo "${YELLOW}[WARN]${NC} $1"; }
err() { echo "${RED}[ERROR]${NC} $1" >&2; }
ok() { echo "${GREEN}[DONE]${NC} $1"; }

# ============================================================
# Prerequisites Check
# ============================================================
check_prereqs() {
    log "Checking prerequisites..."

    # Must be OpenBSD
    if [ "$(uname -s)" != "OpenBSD" ]; then
        err "This script MUST be run on OpenBSD ${OPENBSD_VERSION}"
        err "Current system: $(uname -s) $(uname -r)"
        exit 1
    fi

    # Must be root (for vnconfig, mounting)
    if [ "$(id -u)" -ne 0 ]; then
        err "This script must be run as root (for vnconfig/mount)"
        exit 1
    fi

    # Check for base image
    if [ ! -f "$BASE_IMG" ]; then
        err "Base image not found: $BASE_IMG"
        err "Link your image: ln -sf ~/Code/Images/install79.img ./Images/install79.img"
        exit 1
    fi

    ok "Prerequisites OK"
}

# ============================================================
# Download Packages
# ============================================================
download_packages() {
    log "Downloading packages..."

    mkdir -p "$WORK_DIR/packages/snapshots/amd64"

    # Get package list from openriot binary
    OPENRIOT_BIN="${SCRIPT_DIR}/../install/openriot"
    if [ ! -x "$OPENRIOT_BIN" ]; then
        err "OpenRiot binary not found. Run 'make install' first."
        exit 1
    fi

    PACKAGES=$("$OPENRIOT_BIN" --packages 2>/dev/null)
    if [ -z "$PACKAGES" ]; then
        err "Failed to get package list from openriot --packages"
        exit 1
    fi

    # Download each package
    PKG_DIR="$WORK_DIR/packages/snapshots/amd64"
    PKG_COUNT=$(echo "$PACKAGES" | wc -l)
    PKG_DONE=0

    for pkg in $PACKAGES; do
        PKG_DONE=$((PKG_DONE + 1))
        printf "\r${BLUE}[INFO]${NC} Downloading package %d/%d: %s" "$PKG_DONE" "$PKG_COUNT" "$pkg"

        if [ -f "$PKG_DIR/${pkg}.tgz" ]; then
            continue  # Already downloaded
        fi

        # Download with retry
        retry=0
        while [ $retry -lt 3 ]; do
            if ftp -o "$PKG_DIR/${pkg}.tgz" \
                "https://cdn.openbsd.org/pub/OpenBSD/snapshots/packages/amd64/${pkg}.tgz" 2>/dev/null; then
                break
            fi
            retry=$((retry + 1))
            sleep 2
        done

        if [ $retry -eq 3 ]; then
            warn "Failed to download: $pkg"
        fi
    done

    echo ""  # Newline after progress
    PKG_SIZE=$(du -sh "$PKG_DIR" 2>/dev/null | cut -f1 || echo "unknown")
    ok "Downloaded packages (${PKG_SIZE})"
}

# ============================================================
# Create openriot.tgz
# ============================================================
create_site() {
    log "Creating openriot.tgz..."

    mkdir -p "$WORK_DIR"
    SITE_DIR="${WORK_DIR}/site"

    # Copy MOTD only (minimal - setup.sh handles everything else)
    if [ -f "${SCRIPT_DIR}/../install/motd" ]; then
        mkdir -p "$SITE_DIR/etc"
        cp "${SCRIPT_DIR}/../install/motd" "$SITE_DIR/etc/motd"
    fi

    # Clone OpenRiot repo (with .git for git pull support)
    log "Setting up OpenRiot repo..."
    REPO_DIR="${WORK_DIR}/repo"
    CACHE_DIR="${SCRIPT_DIR}/repo-cache"
    
    if [ -d "$CACHE_DIR/.git" ]; then
        # Use cached repo, just pull latest
        log "Using cached repo, pulling latest..."
        (cd "$CACHE_DIR" && git fetch --depth 1 origin && git reset --hard origin/main)
        rm -rf "$REPO_DIR"
        cp -r "$CACHE_DIR" "$REPO_DIR"
    else
        # Fresh clone
        rm -rf "$REPO_DIR"
        git clone --depth 1 https://github.com/CyphrRiot/OpenRiot "$REPO_DIR"
        # Cache for next time
        rm -rf "$CACHE_DIR"
        cp -r "$REPO_DIR" "$CACHE_DIR"
    fi

    # Move install binary to standard location inside repo
    mkdir -p "$REPO_DIR/install"
    if [ -f "${SCRIPT_DIR}/../install/openriot" ]; then
        cp "${SCRIPT_DIR}/../install/openriot" "$REPO_DIR/install/"
    fi

    # Download and bundle packages
    download_packages
    mkdir -p "$SITE_DIR/openriot/packages/snapshots/amd64"
    if [ -d "$WORK_DIR/packages/snapshots/amd64" ]; then
        cp "$WORK_DIR/packages/snapshots/amd64"/*.tgz "$SITE_DIR/openriot/packages/snapshots/amd64/" 2>/dev/null || true
        log "Added packages to tarball"
    fi

    # Move repo into tarball structure
    mkdir -p "$SITE_DIR/openriot"
    rm -rf "$SITE_DIR/openriot/repo"
    mv "$REPO_DIR" "$SITE_DIR/openriot/repo"
    log "Added OpenRiot repo to tarball"

    # Create install.site
    cat > "$SITE_DIR/install.site" << 'INSTALLSITE'
#!/bin/sh
# OpenRiot post-install script
# Runs during OpenBSD installer (runs in install environment)
# Only does: extract tarball, install local packages, add welcome message
# Everything else is handled by setup.sh after install

set -e

PKG_PATH_LOCAL="/openriot/packages/snapshots/amd64"

log() { echo "[OPENRIOT] $*"; }

log "OpenRiot post-install starting"

# STEP 1: Extract openriot.tgz
log "Extracting openriot.tgz..."
if [ -f /openriot.tgz ]; then
    tar xzf /openriot.tgz -C / || log "Warning: extraction failed"
else
    log "Warning: openriot.tgz not found"
fi

# STEP 2: Install packages from local path
log "Installing packages from local path..."
if [ -d "$PKG_PATH_LOCAL" ]; then
    for pkg in "$PKG_PATH_LOCAL"/*.tgz; do
        [ -f "$pkg" ] || continue
        pkg_name=$(basename "$pkg" .tgz)
        # Extract base name (strip version)
        base_pkg=$(echo "$pkg_name" | sed 's/-[0-9].*//')
        log "Installing $base_pkg..."
        PKG_PATH="$PKG_PATH_LOCAL" pkg_add "$base_pkg" 2>/dev/null || log "Failed: $base_pkg"
    done
    log "Package install complete"
else
    log "Warning: package directory not found"
fi

# STEP 3: Configure doas (passwordless for wheel group)
log "Configuring doas..."
echo "permit nopass :wheel" > /etc/doas.conf
chmod 0440 /etc/doas.conf
log "doas configured"

# STEP 4: Move repo to user's .local/share/openriot
# This prevents setup.sh from having to git clone 100MB from remote
log "Setting up OpenRiot repo..."
for homedir in /home/*; do
    [ -d "$homedir" ] || continue
    username="$(basename "$homedir")"
    target_dir="$homedir/.local/share/openriot"
    mkdir -p "$target_dir"
    if [ -d /openriot/repo ]; then
        rm -rf "$target_dir/repo" 2>/dev/null || true
        cp -r /openriot/repo "$target_dir/repo"
        chown -R "$username:$username" "$target_dir/repo"
        log "Repo installed for $username"
    fi
done

# STEP 5: Add welcome message to skel
log "Adding welcome message..."

# Add to ksh profile (default OpenBSD shell)
printf '\n# OpenRiot first login\nif [ ! -f ~/.openriot-setup-done ]; then\n    echo ""\n    echo "Welcome to OpenRiot"\n    echo ""\n    echo "Make sure you have internet, then run:"\n    echo ""\n    echo "    curl -fsSL https://OpenRiot.org/sh | sh"\n    echo ""\n    touch ~/.openriot-setup-done\nfi\n' >> /etc/skel/.profile

# Add to fish config for fish users (fish might not be installed yet)
mkdir -p /etc/skel/.config/fish
printf '\n# OpenRiot first login\nif test ! -f ~/.openriot-setup-done\n    echo ""\n    echo "Welcome to OpenRiot"\n    echo ""\n    echo "Make sure you have internet, then run:"\n    echo ""\n    echo "    curl -fsSL https://OpenRiot.org/sh | sh"\n    echo ""\n    touch ~/.openriot-setup-done\nend\n' >> /etc/skel/.config/fish/config.fish

log "Post-install complete"
INSTALLSITE
    chmod +x "$SITE_DIR/install.site"
    log "Created install.site"

    # Create install.conf (interactive prompts)
    cat > "$SITE_DIR/install.conf" << 'INSTALLCONF'
# OpenRiot autoinstall answers
# User will be prompted for: disk, hostname, passwords, timezone, partition layout

Which disk is the root disk = ask
Use (W)hole disk MBR, whole disk (G)PT, or (E)dit = edit

System hostname = ask
Password for root = ask
Setup a user = ask
Password for user = ask
What timezone are you in = US/Pacific

# Use OpenBSD CDN for base sets
Location of sets = http
HTTP Server = cdn.openbsd.org
INSTALLCONF
    log "Created install.conf"

    # Create tarball
    cd "$SITE_DIR"
    tar czvf "$OPENRIOT_TGZ" .
    cd ..

    ok "Created $OPENRIOT_TGZ ($(du -h "$OPENRIOT_TGZ" | cut -f1))"
}

# ============================================================
# Cleanup mounts and vnd devices
# ============================================================
cleanup_mounts() {
    umount /mnt 2>/dev/null || true
    vnconfig -u vnd0 2>/dev/null || true
    sleep 1
}

# ============================================================
# Expand Image
# ============================================================
expand_img() {
    # Calculate needed size: base image + tarball + 300MB buffer
    # (base image file size != used space inside fs; 300MB guards against overhead)
    TGZ_MB=$(( $(stat -f %z "$OPENRIOT_TGZ" 2>/dev/null || stat -c %s "$OPENRIOT_TGZ") / 1048576 ))
    BASE_MB=$(( $(stat -f %z "$OUTPUT_IMG" 2>/dev/null || stat -c %s "$OUTPUT_IMG") / 1048576 ))
    NEEDED_MB=$(( BASE_MB + TGZ_MB + 300 ))
    log "Expanding image to ${NEEDED_MB}MB (${BASE_MB}MB base + ${TGZ_MB}MB tarball + 300MB buffer)..."

    truncate -s "${NEEDED_MB}M" "$OUTPUT_IMG"

    # Cleanup
    cleanup_mounts

    # Configure vnd
    vnconfig -u vnd0 2>/dev/null || true
    vnconfig vnd0 "$OUTPUT_IMG"

    # Get current info - use exact field positions
    ROOT_START=$(disklabel vnd0 | grep '^  a:' | awk '{print $3}')
    ROOT_FSTYPE=$(disklabel vnd0 | grep '^  a:' | awk '{print $4}')

    # Total sectors is on a line like "total sectors: 4194304"
    TOTAL_SEC=$(disklabel vnd0 | grep -E '^total sectors:' | awk '{print $3}')

    # Fallback: use cylinder calculation
    if [ -z "$TOTAL_SEC" ]; then
        CYLINDERS=$(disklabel vnd0 | grep '^cylinders:' | awk '{print $2}')
        SECTORS_PER=$(disklabel vnd0 | grep '^sectors/cylinder:' | awk '{print $2}')
        TOTAL_SEC=$((CYLINDERS * SECTORS_PER))
    fi

    # New size: fill to 95% of disk (leave buffer)
    NEW_SIZE=$((TOTAL_SEC * 95 / 100 - ROOT_START))

    log "total=$TOTAL_SEC start=$ROOT_START new_size=$NEW_SIZE"

    # Create full prototype file
    cat > /tmp/newlabel.txt << PROTOTYPE
# /dev/rvnd0c:
type: vnd
disk: vnd device
label: fictitious
duid: 2c656feb7ddb57e2
flags:
bytes/sector: 512
sectors/track: 100
tracks/cylinder: 1
sectors/cylinder: 100
cylinders: 31458
total sectors: $TOTAL_SEC

16 partitions:
#                size           offset  fstype [fsize bsize   cpg]
  a:          $NEW_SIZE             $ROOT_START  $ROOT_FSTYPE   2048 16384 16142
  c:          $TOTAL_SEC                0  unused
  i:              960               64   MSDOS
PROTOTYPE

    disklabel -R vnd0 /tmp/newlabel.txt

    # Grow filesystem
    growfs -y /dev/vnd0a

    vnconfig -u vnd0
    ok "Image expanded"
}

# Build Final Image
# ============================================================
build_img() {
    log "Building final image..."

    mkdir -p "$WORK_DIR"

    # Copy base image FIRST (so we don't modify source)
    log "Copying base image..."
    cp "$BASE_IMG" "$OUTPUT_IMG"

    # Expand image to 2GB (for tarball + base)
    expand_img

    # Cleanup any leftover mounts
    umount /mnt 2>/dev/null || true
    vnconfig -u vnd0 2>/dev/null || true
    sleep 1

    # Mount OUTPUT image and inject openriot.tgz
    log "Mounting image..."
    vnconfig vnd0 "$OUTPUT_IMG"

    log "Running fsck..."
    fsck -y /dev/vnd0a || true

    log "Mounting..."
    mount /dev/vnd0a /mnt

    log "Injecting openriot.tgz..."
    cp "$OPENRIOT_TGZ" /mnt/openriot.tgz

    umount /mnt
    vnconfig -u vnd0

    ok "Image created: $OUTPUT_IMG"
    ok "openriot.tgz injected into image"
    ok ""
    ok "Build complete!"
    ok ""
    ok "Work directory (Build/work/) can be cleaned with './make-img.sh clean'"

    # Offer to burn if sd2 exists
    if [ -e /dev/rsd2c ]; then
        DRIVE_INFO=$(dmesg | grep '^sd2 ' | tail -1)
        DRIVE_PATH=$(echo "$DRIVE_INFO" | cut -d: -f1)
        DRIVE_DESC=$(echo "$DRIVE_INFO" | cut -d: -f2-)
        printf "\n"
        printf "${YELLOW}[WARN]${NC} Removable drive detected:\n"
        printf "${YELLOW}[WARN]${NC}   %s\n" "$DRIVE_PATH"
        printf "${YELLOW}[WARN]${NC}   %s\n" "$DRIVE_DESC"
        printf "${YELLOW}[WARN]${NC} THIS WILL ERASE /dev/sd2 COMPLETELY.\n"
        printf "\n"
        printf "${BLUE}[ASK ]${NC} Would you like to burn this image? [Y/n] "
        read BURN_ANSWER
        case "${BURN_ANSWER:-Y}" in
            [Yy]|"")
                ok "Burning to /dev/rsd2c..."
                cat "$OUTPUT_IMG" | pv -pterb | dd of=/dev/rsd2c bs=1M
                ok "Burn complete!"
                ;;
            *)
                ok "Skipping burn. Flash Images/openriot.img to USB when ready."
                ;;
        esac
    else
        ok "Flash Images/openriot.img to USB when ready."
    fi
}

# ============================================================
# Cleanup
# ============================================================
cleanup() {
    log "Cleaning up..."
    umount /mnt 2>/dev/null || true
    vnconfig -u vnd0 2>/dev/null || true
    rm -rf "$WORK_DIR"
    ok "Cleanup complete"
}

# ============================================================
# Usage
# ============================================================
usage() {
    echo "OpenRiot Image Builder"
    echo ""
    echo "Usage: ./make-img.sh [command]"
    echo ""
    echo "Commands:"
    echo "  (none)    Full build: site + image"
    echo "  site      Create openriot.tgz only"
    echo "  clean     Clean build artifacts"
    echo "  help      Show this help"
}

# ============================================================
# Main
# ============================================================
main() {
    log "OpenRiot Image Builder v${OPENRIOT_VERSION}"
    log "Building for OpenBSD ${OPENBSD_VERSION}"

    case "${1:-}" in
        site)
            check_prereqs
            create_site
            ;;
        clean)
            cleanup
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            check_prereqs
            create_site
            build_img
            ;;
    esac
}

main "$@"

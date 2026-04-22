#!/bin/sh
# make-img.sh - Build OpenRiot custom installer image
# Must be run on OpenBSD (current/snapshots)
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
RED=$(printf '\033[0;31m')
GREEN=$(printf '\033[0;32m')
YELLOW=$(printf '\033[1;33m')
CYAN=$(printf '\033[1;36m')
RESET=$(printf '\033[0m')

log() { echo "${CYAN}[INFO]${RESET} $1"; }
warn() { echo "${YELLOW}[WARN]${RESET} $1"; }
err() { echo "${RED}[ERROR]${RESET} $1" >&2; }
ok() { echo "${GREEN}[DONE]${RESET} $1"; }

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

    # Load exceptions (packages to exclude from image)
    EXCEPTIONS=""
    if [ -f "${SCRIPT_DIR}/exceptions.yaml" ]; then
        EXCEPTIONS=$(grep "^  - " "${SCRIPT_DIR}/exceptions.yaml" 2>/dev/null | sed 's/^  - //' | tr '\n' ' ')
        log "Exceptions loaded: $EXCEPTIONS"
    fi

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

    # Clean old packages no longer in list or in exceptions
    PKG_DIR="$WORK_DIR/packages/snapshots/amd64"
    if [ -d "$PKG_DIR" ]; then
        log "Cleaning stale packages..."
        for tgz in "$PKG_DIR"/*.tgz; do
            [ -f "$tgz" ] || continue
            pkg_name=$(basename "$tgz" .tgz)
            base_pkg=$(echo "$pkg_name" | sed 's/-[0-9].*//')

            # Check if package is in current list
            in_list=0
            for pkg in $PACKAGES; do
                pkg_base=$(echo "$pkg" | sed 's/-[0-9].*//')
                if [ "$base_pkg" = "$pkg_base" ]; then
                    in_list=1
                    break
                fi
            done

            # Check if package is in exceptions
            is_excluded=0
            for excl in $EXCEPTIONS; do
                if [ "$base_pkg" = "$excl" ]; then
                    is_excluded=1
                    break
                fi
            done

            if [ "$in_list" = "0" ] || [ "$is_excluded" = "1" ]; then
                rm -f "$tgz"
            fi
        done
    fi

    # Download each package
    PKG_COUNT=$(echo "$PACKAGES" | wc -l)
    PKG_DONE=0

    for pkg in $PACKAGES; do
        # Check if package is in exceptions list
        PKG_BASE=$(echo "$pkg" | sed 's/-[0-9].*//')
        skip=0
        for excl in $EXCEPTIONS; do
            if [ "$PKG_BASE" = "$excl" ]; then
                skip=1
                break
            fi
        done
        if [ "$skip" = "1" ]; then
            continue  # Skip this package
        fi

        PKG_DONE=$((PKG_DONE + 1))
        printf "\r${CYAN}[INFO]${RESET} Downloading package %d/%d: %s" "$PKG_DONE" "$PKG_COUNT" "$pkg"

        if [ -f "$PKG_DIR/${pkg}.tgz" ]; then
            continue  # Already downloaded
        fi

        # Remove old versions of this package (e.g., tdesktop-6.7.5.tgz before downloading 6.7.6)
        rm -f "$PKG_DIR"/${pkg}-*.tgz

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

    # Clean old site content (packages from previous runs persist!)
    rm -rf "$SITE_DIR/openriot"
    # Remove old tarball so it gets rebuilt (packages are downloaded fresh)
    rm -f "$OPENRIOT_TGZ"

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
        mkdir -p "$REPO_DIR"
        # Copy repo but exclude packages/ (we download fresh)
        for item in "$CACHE_DIR"/*; do
            name=$(basename "$item")
            if [ "$name" != "packages" ]; then
                cp -r "$item" "$REPO_DIR/"
            fi
        done
    else
        # Fresh clone
        rm -rf "$REPO_DIR"
        git clone --depth 1 https://github.com/CyphrRiot/OpenRiot "$REPO_DIR"
        # Cache for next time (exclude packages/ only)
        rm -rf "$CACHE_DIR"
        mkdir -p "$CACHE_DIR"
        for item in "$REPO_DIR"/*; do
            name=$(basename "$item")
            if [ "$name" != "packages" ]; then
                cp -r "$item" "$CACHE_DIR/"
            fi
        done
    fi

    # Move install binary to standard location inside repo
    mkdir -p "$REPO_DIR/install"
    if [ -f "${SCRIPT_DIR}/../install/openriot" ]; then
        cp "${SCRIPT_DIR}/../install/openriot" "$REPO_DIR/install/"
    fi

    # Download packages (kept separate from tarball for size optimization)
    download_packages

    # Move repo into tarball structure (NO packages in tarball)
    mkdir -p "$SITE_DIR/openriot"
    rm -rf "$SITE_DIR/openriot/repo"
    mv "$REPO_DIR" "$SITE_DIR/openriot/repo"
    log "Added OpenRiot repo to tarball"

    # Create install.site
    cat > "$SITE_DIR/install.site" << 'INSTALLSITE'
#!/bin/sh
# OpenRiot post-install script
# Runs during OpenBSD installer

log() { echo "[OPENRIOT] $*"; }

log "OpenRiot post-install starting"

# STEP 1: Extract openriot.tgz FIRST
log "Extracting openriot.tgz..."
if [ -f /openriot.tgz ]; then
    tar xzf /openriot.tgz -C / 2>&1 || log "Warning: extraction failed"
else
    log "Warning: openriot.tgz not found"
fi

# STEP 2: Configure doas (must happen early so user can sudo)
log "Configuring doas..."
echo "permit nopass :wheel" > /etc/doas.conf
chmod 0440 /etc/doas.conf
log "doas configured"

# STEP 3: Configure installurl for the NEW installed system
log "Configuring installurl..."
echo "https://cdn.openbsd.org/pub/OpenBSD" > /etc/installurl
log "installurl configured"

# STEP 4: Install packages from local path
PKG_PATH_LOCAL="/openriot/packages/snapshots/amd64"
log "Installing packages from local path..."
if [ -d "$PKG_PATH_LOCAL" ]; then
    for pkg in "$PKG_PATH_LOCAL"/*.tgz; do
        [ -f "$pkg" ] || continue
        pkg_name=$(basename "$pkg" .tgz)
        base_pkg=$(echo "$pkg_name" | sed 's/-[0-9].*//')
        log "Installing $base_pkg..."
        PKG_PATH="$PKG_PATH_LOCAL" pkg_add "$base_pkg" 2>&1 || log "Failed: $base_pkg"
    done
    log "Package install complete"
else
    log "Warning: package directory not found"
fi

# STEP 5: Move repo to user's .local/share/openriot
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

# STEP 6: Add welcome message to skel
log "Adding welcome message..."
if [ -f /etc/skel/.profile ] && ! grep -q openriot-setup-done /etc/skel/.profile 2>/dev/null; then
    cat >> /etc/skel/.profile << 'WELCOME'

# OpenRiot first login
if [ ! -f ~/.openriot-setup-done ]; then
    echo ""
    echo "Welcome to OpenRiot"
    echo ""
    echo "Run the following command to complete setup:"
    echo ""
    echo "    curl -fsSL https://OpenRiot.org/sh | sh"
    echo ""
    touch ~/.openriot-setup-done
fi
WELCOME
fi

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

# Sets come from the install disk (which is already mounted at /)
Location of sets = disk
Is the disk partition already mounted? = yes
Pathname to the sets = /

# Install openriot.tgz from the disk
install openriot.tgz = yes
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
    # Create a fixed 2GB image (handles up to ~1.7GB of content)
    log "Expanding image to 2GB..."
    truncate -s 2G "$OUTPUT_IMG"

    # Cleanup
    cleanup_mounts

    # Configure vnd
    vnconfig -u vnd0 2>/dev/null || true
    vnconfig vnd0 "$OUTPUT_IMG"

    # Get current info
    ROOT_START=$(disklabel vnd0 | grep '^  a:' | awk '{print $3}')
    ROOT_FSTYPE=$(disklabel vnd0 | grep '^  a:' | awk '{print $4}')
    TOTAL_SEC=$(($(stat -f %z "$OUTPUT_IMG") / 512))

    # Fill partition
    NEW_SIZE=$((TOTAL_SEC - ROOT_START))

    log "total=$TOTAL_SEC start=$ROOT_START new_size=$NEW_SIZE"

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
    growfs -y /dev/vnd0a

    vnconfig -u vnd0
    ok "Image expanded to 2GB"
}

# ============================================================
# Shrink Image
# ============================================================
shrink_img() {
    log "Shrinking image to fit content..."

    vnconfig -u vnd0 2>/dev/null || true
    vnconfig vnd0 "$OUTPUT_IMG"

    # Get filesystem usage
    DF_OUTPUT=$(df -k /dev/vnd0a | tail -1)
    USED_KB=$(echo "$DF_OUTPUT" | awk '{print $3}')

    # Calculate needed size: used space + 10% buffer + space for filesystem metadata
    NEEDED_KB=$((USED_KB * 110 / 100 + 32768))

    # Convert to MB and align to 4MB
    NEEDED_MB=$(( (NEEDED_KB + 4095) / 4096 * 4))

    # Minimum 1GB
    [ "$NEEDED_MB" -lt 1024 ] && NEEDED_MB=1024

    log "Shrinking to ${NEEDED_MB}MB (used: ${USED_KB}KB)..."

    vnconfig -u vnd0

    truncate -s "${NEEDED_MB}M" "$OUTPUT_IMG"

    ok "Image shrunk to ${NEEDED_MB}MB"
}

# Build Final Image
# ============================================================
build_img() {
    log "Building final image..."

    mkdir -p "$(dirname "$OUTPUT_IMG")"

    # Remove old image - we build fresh from base
    rm -f "$OUTPUT_IMG"

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

    # Copy packages separately (not in tarball for size optimization)
    log "Injecting packages..."
    mkdir -p /mnt/openriot/packages/snapshots/amd64
    if [ -d "$WORK_DIR/packages/snapshots/amd64" ]; then
        cp "$WORK_DIR/packages/snapshots/amd64"/*.tgz /mnt/openriot/packages/snapshots/amd64/
        log "Packages injected ($(du -sh "$WORK_DIR/packages/snapshots/amd64" | cut -f1))"
    fi

    umount /mnt
    vnconfig -u vnd0

    # Shrink image to fit actual content
    shrink_img

    ok "Image created: $OUTPUT_IMG"
    ok "openriot.tgz injected into image"

    # Generate SHA256 checksum
    log "Generating SHA256 checksum..."
    SHA256_FILE="$(dirname "$OUTPUT_IMG")/openriot.sha256"
    sha256 -q "$OUTPUT_IMG" > "$SHA256_FILE"
    ok "Checksum: $(cat "$SHA256_FILE")"

    ok ""
    ok "Build complete!"
    ok ""
    ok "Work directory (Build/work/) can be cleaned with './make-img.sh clean'"

    # Find currently attached drives and detect root
    ROOT_DRIVE=""

    # Root drive from dmesg
    root_dev=$(dmesg 2>/dev/null | grep "^root on " | head -1 | sed 's/.*on \([a-z]*[0-9]*\)[a-z].*/\1/')
    [ -n "$root_dev" ] && ROOT_DRIVE=$(echo "$root_dev" | sed 's/[0-9]*//')

    # Removable drives from dmesg
    REMOVABLE_LIST=""
    for disk in $(dmesg 2>/dev/null | grep "removable" | grep -oE 'sd[0-9]+' | sort -u); do
        REMOVABLE_LIST="${REMOVABLE_LIST}${REMOVABLE_LIST:+ }$disk"
    done

    # Protected drives: root + drives with RAID partitions (softraid parents)
    PROTECTED="$ROOT_DRIVE"

    for disk in $(sysctl -n hw.disknames 2>/dev/null | tr ',' '\n' | grep -oE '^(sd|wd)[0-9]+'); do
        label=$(doas disklabel "$disk" 2>/dev/null || true)
        if [ -n "$label" ]; then
            # Check for RAID partitions (softraid parent)
            has_raid=$(echo "$label" | grep -cE '^  [a-z]:.*RAID' || true)
            [ "$has_raid" -gt 0 ] && PROTECTED="${PROTECTED}${PROTECTED:+ }$disk"

            bytes_per_sec=$(echo "$label" | sed -n 's/.*bytes\/sector:[[:space:]]*\([0-9]*\).*/\1/p')
            total_sectors=$(echo "$label" | awk '/^[[:space:]]*c:/ {print $2; exit}')
            if [ -n "$bytes_per_sec" ] && [ -n "$total_sectors" ] && [ "$bytes_per_sec" -gt 0 ]; then
                total_bytes=$((total_sectors * bytes_per_sec))
                size_gb=$((total_bytes / 1073741824))
                if [ "$size_gb" -gt 0 ]; then
                    # Check if removable
                    is_removable=""
                    for rem in $REMOVABLE_LIST; do
                        [ "$disk" = "$rem" ] && is_removable="1" && break
                    done

                    REMOVABLE_DRIVES="${REMOVABLE_DRIVES}${REMOVABLE_DRIVES:+
}${disk}|${size_gb}|${is_removable}"
                fi
            fi
        fi
    done

    # Show detected drives
    if [ -n "$REMOVABLE_DRIVES" ]; then
        printf "\n"

        # Build display and drive list
        DRIVE_LIST=""
        DISPLAY_LINES=""
        while IFS='|' read -r drive size is_removable; do
            drive_short=$(echo "$drive" | sed 's|dev/||')

            # Check if protected
            is_protected=""
            for prot in $PROTECTED; do
                [ "$drive" = "$prot" ] && is_protected="1" && break
            done

            if [ "$is_protected" = "1" ]; then
                if [ "$drive" = "$ROOT_DRIVE" ]; then
                    prefix="${RED}[ROOT]${RESET}"
                    suffix=" [OpenBSD Encrypted]"
                else
                    prefix="${RED}[ROOT]${RESET}"
                    suffix=" [OpenBSD]"
                fi
                exclude="1"
            elif [ "$is_removable" = "1" ]; then
                prefix="${YELLOW}[WARN]${RESET}"
                suffix=" [Removable USB]"
                exclude="0"
            else
                prefix="${CYAN}[INFO]${RESET}"
                suffix=""
                exclude="0"
            fi

            line="${prefix} ${drive_short} - %5d GB${suffix}"
            DISPLAY_LINES="${DISPLAY_LINES}${line}|${size}|${drive_short}|${exclude}
"

            # Build drive list for prompt (exclude ROOT only, WARN still selectable)
            if [ "$exclude" != "1" ]; then
                if [ -n "$DRIVE_LIST" ]; then
                    DRIVE_LIST="${DRIVE_LIST}, ${drive_short}"
                else
                    DRIVE_LIST="${drive_short}"
                fi
            fi
        done << EOF
$REMOVABLE_DRIVES
EOF
        # Print all display lines
        printf "%s\n" "$DISPLAY_LINES" | while IFS='|' read -r line size drive_short exclude; do
            printf "${line}\n" "$size"
        done

        if [ -n "$DRIVE_LIST" ]; then
            printf "${GREEN}[DONE]${RESET} Available for burn: %s\n" "$DRIVE_LIST"
        fi

        # Only prompt for burn if there are eligible drives
        if [ -n "$DRIVE_LIST" ]; then
            printf "${YELLOW}[WARN]${RESET} THIS WILL ERASE ALL DATA ON THE SELECTED DRIVE.\n"
            printf "\n"
            printf "${CYAN}[ASK ]${RESET} Which drive to burn? (${DRIVE_LIST} or press Enter to skip) "
            read BURN_CHOICE

            # Default to 'n' if empty (Enter pressed)
            if [ -z "$BURN_CHOICE" ]; then
                BURN_CHOICE="n"
            fi

            case "${BURN_CHOICE}" in
                [Nn])
                    ok "Skipping burn. Flash Images/openriot.img to USB when ready."
                    ;;
                *)
                    # Find drive size for confirmation
                    DRIVE_SIZE=""
                    while IFS='|' read -r line size drive_short exclude; do
                        if [ "$drive_short" = "$BURN_CHOICE" ]; then
                            DRIVE_SIZE="$size"
                            break
                        fi
                    done << EOF
$DISPLAY_LINES
EOF

                    # Confirmation prompt
                    printf "\n"
                    printf "${YELLOW}[WARN]${RESET} You will be erasing ${BURN_CHOICE} (${DRIVE_SIZE} GB).\n"
                    printf "${CYAN}[ASK ]${RESET} Are you sure? [y/N] "
                    read CONFIRM

                    case "${CONFIRM}" in
                        [Yy])
                            BURN_DEV="/dev/r${BURN_CHOICE}c"
                            if [ -c "$BURN_DEV" ]; then
                                ok "Burning to $BURN_DEV..."
                                cat "$OUTPUT_IMG" | pv -pterb | doas dd of="$BURN_DEV" bs=1M
                                ok "Burn complete!"
                            else
                                err "Drive $BURN_DEV not found"
                            fi
                            ;;
                        *)
                            ok "Aborted. Flash Images/openriot.img to USB when ready."
                            ;;
                    esac
                    ;;
            esac
        else
            log "No removable drives detected. Cannot burn image."
            ok "Flash Images/openriot.img to USB when ready."
        fi
    else
        log "No drives detected."
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

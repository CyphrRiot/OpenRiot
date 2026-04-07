#!/bin/sh
# OpenRiot Setup Script
# Bootstrap script for installing OpenRiot on fresh OpenBSD
# Usage:
#   curl -fsSL https://openriot.org/setup.sh | sh     # auto-detect
#   curl -fsSL https://openriot.org/setup.sh | sh -s -- --install   # fresh install
#   curl -fsSL https://openriot.org/setup.sh | sh -s -- --upgrade   # upgrade

# NOTE: set -e removed - install_packages continues on individual pkg failures

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;36m'
NC='\033[0m' # No Color

# Configuration
OPENBSD_MIN_VERSION="7.9"
REPO_URL="${REPO_URL:-https://github.com/CyphrRiot/OpenRiot}"
CONFIG_BRANCH="${CONFIG_BRANCH:-main}"
INSTALLURL="${INSTALLURL:-https://cdn.openbsd.org/pub/OpenBSD}"
REMOTE_VERSION_URL="${REMOTE_VERSION_URL:-https://openriot.org/VERSION}"
# Detect actual user home (HOME may be wrong under doas/sudo)
REAL_USER=$(id -un 2>/dev/null || echo "$USER")
REAL_HOME=$(getent passwd "$REAL_USER" 2>/dev/null | cut -d: -f6)

# Fallback if getent fails
if [ -z "$REAL_HOME" ]; then
    REAL_HOME="${HOME:-$(eval echo ~$REAL_USER)}"
fi

INSTALL_DIR="$REAL_HOME/.local/share/openriot"
export OPENRIOT_CONFIG_DIR="$INSTALL_DIR/install"

# --install mode forces fresh clone even if .git exists
FORCE_INSTALL=0
for arg in "$@"; do
    [ "$arg" = "--install" ] && FORCE_INSTALL=1
done

# Log file configuration - logs go to ~/.cache/openriot/ NOT ~/.local/share/openriot/
LOG_DIR="$HOME/.cache/openriot"
LOG_FILE="$LOG_DIR/setup.log"
mkdir -p "$LOG_DIR"

# -----------------------------------------------------------------------------
# Helper Functions
# -----------------------------------------------------------------------------

info() { echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"; }
success() { echo -e "${GREEN}[DONE]${NC} $1" | tee -a "$LOG_FILE"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$LOG_FILE"; }
error() { echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE" >&2; }

log() { printf '[OPENRIOT] %s\n' "$1" | tee -a "$LOG_FILE"; }

# Upload log to tmpfiles.org for sharing with developers
share_log() {
    log_file="${1:-$LOG_FILE}"
    if [ ! -f "$log_file" ]; then
        echo "Log file not found: $log_file"
        return 1
    fi
    echo "Uploading log..."
    response=$(curl -s -F "file=@$log_file" "https://tmpfiles.org/api/v1/upload" 2>/dev/null)
    url=$(echo "$response" | grep -oE '"url":"[^"]+' | sed 's/"url":"//' | sed 's/\\//g')
    if echo "$url" | grep -qE "^https?://tmpfiles.org"; then
        echo "Log uploaded to: $url"
        echo "$url" > "${log_file}.url"
    else
        echo "Upload failed. Showing last 100 lines:"
        tail -100 "$log_file"
    fi
}

# -----------------------------------------------------------------------------
# Version Comparison
# -----------------------------------------------------------------------------

# Compare semantic versions - returns 0 if remote is newer
is_newer_version() {
    local_ver="$1"
    remote_ver="$2"

    [ "$local_ver" = "unknown" ] && return 1
    [ "$remote_ver" = "unknown" ] && return 1
    [ "$local_ver" = "$remote_ver" ] && return 1

    newer=$(printf '%s\n%s\n' "$local_ver" "$remote_ver" | awk 'BEGIN{FS="."} {
        for (i=1; i<=3; i++) { v[NR][i] = ($i+0) }
    } END {
        for (i=1; i<=3; i++) {
            if (v[2][i] > v[1][i]) { print "newer"; exit }
            if (v[1][i] > v[2][i]) { print "older"; exit }
        }
        print "equal"
    }')

    [ "$newer" = "newer" ] && return 0
    return 1
}

# Get remote version from openriot.org
get_remote_version() {
    timeout 10 curl -fsSL "$REMOTE_VERSION_URL" 2>/dev/null || echo "unknown"
}

# Get local version from installed repo
get_local_version() {
    if [ -f "$INSTALL_DIR/VERSION" ]; then
        cat "$INSTALL_DIR/VERSION" 2>/dev/null || echo "unknown"
    else
        echo "unknown"
    fi
}

# -----------------------------------------------------------------------------
# Pre-flight Checks
# -----------------------------------------------------------------------------

check_openbsd_version() {
    info "Checking OpenBSD version..."
    os=$(uname -s)
    if [ "$os" != "OpenBSD" ]; then
        error "This script is for OpenBSD only."
        exit 1
    fi
    version=$(uname -r | sed 's/-.*//')
    major=$(echo "$version" | cut -d. -f1)
    minor=$(echo "$version" | cut -d. -f2)
    min_major=$(echo "$OPENBSD_MIN_VERSION" | cut -d. -f1)
    min_minor=$(echo "$OPENBSD_MIN_VERSION" | cut -d. -f2)
    if [ "$major" -lt "$min_major" ] || ([ "$major" -eq "$min_major" ] && [ "$minor" -lt "$min_minor" ]); then
        error "OpenBSD $OPENBSD_MIN_VERSION or higher required. Detected: $version"
        exit 1
    fi
    success "OpenBSD $version detected"
}

# Check available disk space (requires GB)
check_disk_space() {
    required_gb=$1
    target_dir="${HOME:-/root}"
    available_kb=$(df -k "$target_dir" | tail -1 | awk '{print $4}')
    available_gb=$(awk "BEGIN {printf \"%.1f\", $available_kb/1048576}")
    required_display=$(awk "BEGIN {printf \"%.1f\", $required_gb}")
    if awk "BEGIN {exit !($available_gb < $required_gb)}"; then
        error "Not enough disk space. Need ${required_display}GB, have ${available_gb}GB free."
        error "Free up space and try again."
        exit 1
    fi
    info "Disk space check passed (${available_gb}GB available)"
}

# -----------------------------------------------------------------------------
# Configure installurl (pkg_add mirror)
# -----------------------------------------------------------------------------

configure_installurl() {
    info "Configuring installurl..."
    echo "$INSTALLURL" | doas tee /etc/installurl >/dev/null
    success "installurl configured"
}

# -----------------------------------------------------------------------------
# Configure doas (nopasswd for wheel)
# -----------------------------------------------------------------------------

configure_doas() {
    info "Configuring doas..."
    doas_conf="/etc/doas.conf"
    doas_entry="permit nopass :wheel"
    if [ -f "$doas_conf" ]; then
        if grep -q "^permit nopass :wheel" "$doas_conf" 2>/dev/null; then
            success "doas already configured"
            return
        fi
        doas cp "$doas_conf" "${doas_conf}.bak"
        warn "Backed up existing doas.conf"
    fi
    echo "$doas_entry" | doas tee "$doas_conf" >/dev/null
    doas chmod 0440 "$doas_conf"
    success "doas configured (nopasswd)"
}

# -----------------------------------------------------------------------------
# Install bootstrap packages (curl and git)
# -----------------------------------------------------------------------------

install_bootstrap_packages() {
    info "Installing bootstrap packages (curl, git)..."
    doas pkg_add curl git
    git config --global pull.rebase true
    git config --global init.defaultBranch master
    success "Bootstrap packages installed"
}

# -----------------------------------------------------------------------------
# Deploy OpenRiot repo - smart mode (upgrade vs fresh)
# -----------------------------------------------------------------------------

setup_repository() {
    local_ver=$(get_local_version)
    remote_ver=$(get_remote_version)

    info "Local version:  $local_ver"
    info "Remote version: $remote_ver"

    # Always deploy repo: fresh clone if no INSTALL_DIR or --install requested
    if [ ! -d "$INSTALL_DIR" ] || [ "$FORCE_INSTALL" = "1" ]; then
        # Fresh install or forced reinstall - reclone to get latest packages.yaml
        if [ -d "$INSTALL_DIR" ]; then
            info "Removing old install and recloning..."
            doas rm -rf "$INSTALL_DIR"
        fi
        mkdir -p "$(dirname "$INSTALL_DIR")" || { error "Cannot create directory"; exit 1; }
        git clone -b "$CONFIG_BRANCH" "$REPO_URL" "$INSTALL_DIR" || { error "Git clone failed"; exit 1; }
        success "OpenRiot deployed to $INSTALL_DIR"
        return
    fi

    # Always pull latest commits to pick up bug fixes and config changes
    if [ -d "$INSTALL_DIR/.git" ]; then
        info "Updating OpenRiot repository..."
        (
            cd "$INSTALL_DIR" || exit 1
            git fetch origin || true
            LOCAL_AHEAD=$(git rev-list --count HEAD..origin/"$CONFIG_BRANCH" 2>/dev/null || echo 0)
            if [ "$LOCAL_AHEAD" -gt 0 ]; then
                info "Pulling $LOCAL_AHEAD new commit(s)..."
                git reset --hard origin/"$CONFIG_BRANCH" || { error "Git reset failed"; exit 1; }
                success "OpenRiot updated."
            else
                info "Repository up to date."
            fi
        )
    fi

    # Check for new version releases
    if is_newer_version "$local_ver" "$remote_ver"; then
        info "New version $remote_ver available!"
    fi
}

# -----------------------------------------------------------------------------
# Install all packages from packages.yaml
# -----------------------------------------------------------------------------

install_packages() {
    info "Installing packages from packages.yaml (safe one-by-one mode)..."

    # Try openriot binary first, fallback to grep if it fails
    if [ -x "$INSTALL_DIR/install/openriot" ]; then
        pkg_raw=$("$INSTALL_DIR/install/openriot" --packages 2>&1)
        pkg_exit=$?
    else
        pkg_raw=""
        pkg_exit=1
    fi
    if [ $pkg_exit -ne 0 ] || [ -z "$pkg_raw" ]; then
        warn "openriot --packages failed or returned empty, using fallback..."
        # Try yq first (preferred), then Python YAML fallback
        if command -v yq >/dev/null 2>&1; then
            pkg_raw=$(yq eval '.. | select(has("packages")) | .packages[]' "$INSTALL_DIR/install/packages.yaml" 2>/dev/null)
        elif command -v python3 >/dev/null 2>&1; then
            pkg_raw=$(python3 -c "
import re, sys
with open('$INSTALL_DIR/install/packages.yaml') as f:
    content = f.read()
in_packages = False
depth = 0
for line in content.splitlines():
    # Track if we enter a new top-level key (increase depth = new module)
    if re.match(r'^[a-z]', line):
        depth += 1
        in_packages = False
    # Detect packages: section
    if re.match(r'^\s+packages:\s*$', line):
        in_packages = True
    elif re.match(r'^\s+(configs|commands|build):\s*$', line):
        in_packages = False
    # Extract package name
    elif in_packages:
        m = re.match(r'^\s+-\s+([A-Za-z][A-Za-z0-9.+-]*)', line)
        if m: print(m.group(1))
" 2>/dev/null)
        else
            pkg_raw=$(grep -E '^ +- [a-zA-Z]' "$INSTALL_DIR/install/packages.yaml" 2>/dev/null | \
                sed 's/^ *- //' | grep -E '[0-9]')
        fi
    fi
    if [ -z "$pkg_raw" ]; then
        error "No packages found in packages.yaml"
        error "Run 'setup.sh --share-log' to share logs for debugging"
        exit 1
    fi
    packages=$(echo "$pkg_raw" | sort -u | grep -v '^$')

    if [ -z "$packages" ]; then
        error "No packages found in packages.yaml"
        exit 1
    fi

    count=$(echo "$packages" | wc -l | tr -d ' ')
    info "Found $count packages. Installing one by one..."

    failed=0
    for pkg in $packages; do
        # Check if already installed (OpenBSD stores pkg info in /var/db/pkg/)
        # Use pkg_info to check - it's more reliable than glob patterns in [ ]
        if pkg_info -e "$pkg" >/dev/null 2>&1; then
            info "  [SKIP] $pkg already installed"
        else
            info "Installing $pkg ..."
            pkg_output=$(doas pkg_add -D unsigned "$pkg" 2>&1)
            pkg_status=$?
            if [ $pkg_status -eq 0 ]; then
                success "$pkg installed."
            else
                warn "Failed to install $pkg."
                echo "$pkg_output" | sed 's/^/    /'
                failed=$((failed + 1))
            fi
        fi
    done

    if [ $failed -gt 0 ]; then
        warn "$failed packages failed to install."
        warn "You can install remaining ones manually: doas pkg_add <package>"
    else
        success "All packages installed successfully!"
    fi
}

# -----------------------------------------------------------------------------
# Run openriot --install (as USER, not root)
# -----------------------------------------------------------------------------

run_openriot_install() {
    info "Running openriot --install..."
    if [ ! -x "$INSTALL_DIR/install/openriot" ]; then
        error "openriot binary not found at $INSTALL_DIR/install/openriot"
        exit 1
    fi
    # Run as USER - no doas, log to ~/.cache/openriot/
    cd "$INSTALL_DIR/install" || { error "Cannot cd to $INSTALL_DIR/install"; exit 1; }

    INSTALL_LOG="$HOME/.cache/openriot/install.log"
    mkdir -p "$(dirname "$INSTALL_LOG")"
    ./openriot --install 2>&1 | tee -a "$INSTALL_LOG"
    success "openriot --install complete"
}

# -----------------------------------------------------------------------------
# Install JetBrainsMono Nerd Font for glyph/icon rendering in foot/lsd/fish
# -----------------------------------------------------------------------------

install_nerd_font() {
    # Check if already installed (JetBrainsMono Nerd Font uses "NF" suffix)
    if fc-list | grep -q "JetBrainsMono.*NF" 2>/dev/null; then
        info "JetBrainsMono Nerd Font already installed."
        return
    fi
    info "Installing JetBrainsMono Nerd Font..."
    font_dir="$REAL_HOME/.local/share/fonts"
    mkdir -p "$font_dir"
    # ~5MB (vs ~30MB for all weights/styles)
    curl -fsSL "https://github.com/ryanoasis/nerd-fonts/releases/download/v3.4.0/JetBrainsMono.zip" -o /tmp/jetbrainsmono.zip
    unzip -q /tmp/jetbrainsmono.zip -d "$font_dir"
    rm -f /tmp/jetbrainsmono.zip
    doas fc-cache -fv >/dev/null 2>&1
    success "JetBrainsMono Nerd Font installed."
}

# -----------------------------------------------------------------------------
# Set fish as default shell
# -----------------------------------------------------------------------------

set_fish_shell() {
    info "Setting fish as default shell..."
    fish_path="/usr/local/bin/fish"
    if ! command -v fish >/dev/null 2>&1; then
        warn "Fish not installed yet - skipping shell change"
        return
    fi
    if ! grep -q "^$fish_path$" /etc/shells 2>/dev/null; then
        echo "$fish_path" | doas tee -a /etc/shells >/dev/null
    fi
    doas chsh -s "$fish_path" "$(whoami)" || warn "Could not change shell for $(whoami)"
    success "Fish shell configured"
}

# -----------------------------------------------------------------------------
# Configure sway autostart in fish config
# -----------------------------------------------------------------------------

configure_sway_autostart() {
    info "Configuring sway autostart in fish..."
    fish_conf="$HOME/.config/fish/config.fish"
    mkdir -p "$HOME/.config/fish"

    # Skip if sway autostart is already configured (any exec sway block)
    if [ -f "$fish_conf" ] && grep -q "exec sway" "$fish_conf" 2>/dev/null; then
        success "Sway autostart already configured"
        return
    fi

    # Append sway autostart
    cat >> "$fish_conf" << 'SWCONF'

# openriot-sway-autostart
if status is-interactive
    # Auto-start Sway on login
    exec sway
end
SWCONF
    success "Sway autostart configured"
}

# -----------------------------------------------------------------------------
# Usage
# -----------------------------------------------------------------------------

usage() {
    echo "Usage: setup.sh [--install | --upgrade | --share-log | --help]"
    echo "  --install   Fresh install (default)"
    echo "  --upgrade   Upgrade if newer version available"
    echo "  --share-log Share latest log file at ix.io"
    echo "  --help      Show this message"
    exit 0
}

# -----------------------------------------------------------------------------
# Main
# -----------------------------------------------------------------------------

main() {
    MODE="install"

    # Parse arguments
    for arg in "$@"; do
        case "$arg" in
            --install) MODE="install" ;;
            --upgrade) MODE="upgrade" ;;
            --share-log)
                share_log "${2:-}"
                exit $?
                ;;
            --help|-h) usage ;;
        esac
    done

    # Fetch remote version for banner (may fail offline)
    banner_ver=$(get_remote_version 2>/dev/null || echo "?.?")

    echo ""
    echo "=== OpenRiot v${banner_ver} Setup (OpenBSD ${OPENBSD_MIN_VERSION}) ==="
    echo ""

    check_openbsd_version
    configure_doas
    configure_installurl
    install_bootstrap_packages

    # Debug: show environment and paths
    info "HOME=$HOME INSTALL_DIR=$INSTALL_DIR PWD=$(pwd) UID=$(id -u)"
    if [ -d "$INSTALL_DIR" ]; then
        info "OpenRiot installation found"
    else
        info "New installation - deploying OpenRiot base"
    fi

    # Detect mode
    if [ "$MODE" = "upgrade" ]; then
        # --upgrade mode: only proceed if newer version available
        local_ver_before=$(get_local_version)
        remote_ver=$(get_remote_version)
        if ! is_newer_version "$local_ver_before" "$remote_ver"; then
            info "No upgrade needed - already on latest version ($local_ver_before)"
            exit 0
        fi
        info "Upgrading from $local_ver_before to $remote_ver..."
    fi

    # Check if this will be a fresh clone (no INSTALL_DIR or --install)
    DID_CLONE=0
    if [ ! -d "$INSTALL_DIR" ] || [ "$FORCE_INSTALL" = "1" ]; then
        DID_CLONE=1
    fi

    setup_repository

    # Nerd Font: installed early so configs can reference it
    install_nerd_font

    # Always run packages (pkg_add is idempotent - skips already-installed)
    check_disk_space 1
    install_packages

    # Deploy sway configs directly (shell fallback - bypasses Go binary failures on OpenBSD)
    deploy_sway_configs

    # Source builds: run directly (not through openriot binary to avoid binary issues)
    # Always run to ensure tools are available, not just on fresh clone
    run_source_builds

    # Run openriot --install as fallback (may partially succeed)
    run_openriot_install
    set_fish_shell
    configure_sway_autostart

    echo ""
    echo "+----------------------------------------------------------+"
    echo "|  OpenRiot bootstrap complete!                            |"
    echo "|                                                          |"
    echo "|  Reboot now, then log in. Sway will start automatically. |"
    echo "+----------------------------------------------------------+"
    echo ""
}

# -----------------------------------------------------------------------------
# Deploy sway config files directly (shell fallback - bypasses broken Go binary)
# This is needed because openriot --install calls Go binary which fails on OpenBSD
# -----------------------------------------------------------------------------

deploy_sway_configs() {
    info "Deploying Sway config files (clean deploy)..."

    # Clean deploy: remove old configs first so removed files get purged
    rm -rf "$REAL_HOME/.config/sway"
    rm -rf "$REAL_HOME/.config/waybar"
    rm -rf "$REAL_HOME/.config/foot"
    rm -rf "$REAL_HOME/.config/fuzzel"

    # Deploy sway configs
    sway_src="$INSTALL_DIR/config/sway"
    sway_dest="$REAL_HOME/.config/sway"
    mkdir -p "$sway_dest"
    for f in config monitors.conf windowrules.conf keybindings.conf brightness-dim.sh swayidle.conf; do
        if [ -f "$sway_src/$f" ]; then
            cp -f "$sway_src/$f" "$sway_dest/$f"
            info "  Deployed $f"
        fi
    done

    # Deploy waybar configs
    waybar_src="$INSTALL_DIR/config/waybar"
    waybar_dest="$REAL_HOME/.config/waybar"
    mkdir -p "$waybar_dest"
    if [ -d "$waybar_src" ]; then
        cp -rf "$waybar_src"/* "$waybar_dest/" 2>/dev/null || true
        info "  Deployed waybar configs"
    fi

    # Deploy foot config
    foot_src="$INSTALL_DIR/config/foot"
    foot_dest="$REAL_HOME/.config/foot"
    mkdir -p "$foot_dest"
    if [ -d "$foot_src" ]; then
        cp -rf "$foot_src"/* "$foot_dest/" 2>/dev/null || true
        info "  Deployed foot configs"
    fi

    # Deploy fuzzel config
    fuzzel_src="$INSTALL_DIR/config/fuzzel"
    fuzzel_dest="$REAL_HOME/.config/fuzzel"
    mkdir -p "$fuzzel_dest"
    if [ -d "$fuzzel_src" ]; then
        cp -rf "$fuzzel_src"/* "$fuzzel_dest/" 2>/dev/null || true
        info "  Deployed fuzzel configs"
    fi

    success "Sway config files deployed."
}

# -----------------------------------------------------------------------------
# Run source builds directly (not through openriot binary)
# -----------------------------------------------------------------------------

run_source_builds() {
    info "Running source builds..."
    failed=0

    # wlsunset: Wayland screen brightness/temperature controller
    if ! command -v wlsunset >/dev/null 2>&1; then
        info "Building wlsunset..."
        # Install wayland-protocols first (required by meson build)
        if ! pkg_info -e wayland-protocols >/dev/null 2>&1; then
            info "Installing wayland-protocols..."
            doas pkg_add wayland-protocols
        fi
        rm -rf /tmp/wlsunset
        cp -r "$INSTALL_DIR/source/wlsunset" /tmp/wlsunset
        cd /tmp/wlsunset
        meson setup build --prefix=/usr/local --buildtype=release 2>&1 | tee -a "$LOG_FILE"
        meson compile -C build 2>&1 | tee -a "$LOG_FILE"
        doas meson install -C build 2>&1 | tee -a "$LOG_FILE"
        if [ -x "/usr/local/bin/wlsunset" ]; then
            success "wlsunset built and installed."
        elif [ -x "/tmp/wlsunset/build/wlsunset" ]; then
            info "Copying wlsunset manually..."
            doas cp /tmp/wlsunset/build/wlsunset /usr/local/bin/wlsunset
            doas chmod +x /usr/local/bin/wlsunset
            if [ -x "/usr/local/bin/wlsunset" ]; then
                success "wlsunset built and installed."
            fi
        else
            warn "wlsunset build may have failed. Check log for errors."
        fi
        rm -rf /tmp/wlsunset
    else
        info "wlsunset already installed."
    fi

    # crush: AI CLI (Charm) — use pre-built OpenBSD binary, no go install needed
    if ! command -v crush >/dev/null 2>&1; then
        info "Installing crush..."
        CRUSH_VER="0.55.1"
        CRUSH_URL="https://github.com/charmbracelet/crush/releases/download/v${CRUSH_VER}/crush_${CRUSH_VER}_Openbsd_x86_64.tar.gz"
        mkdir -p "$REAL_HOME/.local/bin"
        rm -f /tmp/crush.tar.gz
        curl -fsSL "$CRUSH_URL" -o /tmp/crush.tar.gz
        tar -xzf /tmp/crush.tar.gz -C /tmp
        # Tarball extracts to subdirectory containing the binary
        mv "/tmp/crush_${CRUSH_VER}_Openbsd_x86_64/crush" "$REAL_HOME/.local/bin/crush"
        chmod +x "$REAL_HOME/.local/bin/crush"
        rm -rf /tmp/crush.tar.gz "/tmp/crush_${CRUSH_VER}_Openbsd_x86_64"
        # Verify
        if [ -x "$REAL_HOME/.local/bin/crush" ]; then
            success "crush installed."
        else
            warn "crush download may have failed."
        fi
    else
        info "crush already installed."
    fi

    # Bibata cursor theme
    if [ ! -d "$REAL_HOME/.local/share/icons/Bibata-Modern-Ice" ]; then
        info "Installing Bibata cursor..."
        mkdir -p "$REAL_HOME/.local/share/icons"
        curl -fsSL https://github.com/ful1e5/Bibata_Cursor/releases/download/v2.0.7/Bibata-Modern-Ice.tar.xz -o /tmp/bibata.tar.xz
        (cd /tmp && xz -d bibata.tar.xz && tar -xf bibata.tar)
        mv /tmp/Bibata-Modern-Ice "$REAL_HOME/.local/share/icons/"
        rm -f /tmp/bibata.tar.xz
        gtk-update-icon-cache -f "$REAL_HOME/.local/share/icons/Bibata-Modern-Ice" 2>/dev/null || true
        if [ -d "$REAL_HOME/.local/share/icons/Bibata-Modern-Ice" ]; then
            success "Bibata cursor installed."
        else
            warn "Bibata cursor may have failed to install."
        fi
    else
        info "Bibata cursor already installed."
    fi

    if [ $failed -gt 0 ]; then
        warn "Some source builds failed."
    else
        success "Source builds complete."
    fi
}

main "$@"

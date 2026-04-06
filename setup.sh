#!/bin/sh
# OpenRiot Setup Script
# Bootstrap script for installing OpenRiot on fresh OpenBSD
# Usage:
#   curl -fsSL https://openriot.org/setup.sh | sh     # auto-detect
#   curl -fsSL https://openriot.org/setup.sh | sh -s -- --install   # fresh install
#   curl -fsSL https://openriot.org/setup.sh | sh -s -- --upgrade   # upgrade

# NOTE: set -e removed — install_packages continues on individual pkg failures

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
INSTALL_DIR="$HOME/.local/share/openriot"

# Log file configuration — logs go to ~/.cache/openriot/ NOT ~/.local/share/openriot/
LOG_DIR="$HOME/.cache/openriot"
LOG_FILE="$LOG_DIR/setup.log"
mkdir -p "$LOG_DIR"

# -----------------------------------------------------------------------------
# Helper Functions
# -----------------------------------------------------------------------------

info() { echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"; }
success() { echo -e "${GREEN}[OKAY]${NC} $1" | tee -a "$LOG_FILE"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$LOG_FILE"; }
error() { echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE" >&2; }

log() { printf '[OPENRIOT] %s\n' "$1" | tee -a "$LOG_FILE"; }

# -----------------------------------------------------------------------------
# Version Comparison
# -----------------------------------------------------------------------------

# Compare semantic versions — returns 0 if remote is newer
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

# Check available disk space
check_disk_space() {
    required_mb=$1
    available_mb=$(df -m "$HOME" | tail -1 | awk '{print $4}')
    if [ "$available_mb" -lt "$required_mb" ]; then
        error "Not enough disk space. Need ${required_mb}MB, have ${available_mb}MB free."
        error "Free up space and try again."
        exit 1
    fi
    info "Disk space check passed (${available_mb}MB available)"
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
    success "Bootstrap packages installed"
}

# -----------------------------------------------------------------------------
# Deploy OpenRiot repo — smart mode (upgrade vs fresh)
# -----------------------------------------------------------------------------

setup_repository() {
    local_ver=$(get_local_version)
    remote_ver=$(get_remote_version)

    info "Local version:  $local_ver"
    info "Remote version: $remote_ver"

    if [ -d "$INSTALL_DIR/.git" ]; then
        # Existing installation — check for updates
        if is_newer_version "$local_ver" "$remote_ver"; then
            info "Newer version available — upgrading..."
            (
                cd "$INSTALL_DIR" || exit 1
                git fetch origin || { error "Git fetch failed"; exit 1; }
                git reset --hard origin/"$CONFIG_BRANCH" || { error "Git reset failed"; exit 1; }
            )
            success "OpenRiot upgraded to $remote_ver"
        else
            info "Already on latest version ($local_ver) — skipping repo update"
            info "To force reinstall, remove ~/.local/share/openriot and re-run"
        fi
    else
        # Fresh installation
        info "Fresh installation..."
        if [ -d "$INSTALL_DIR" ]; then
            doas rm -rf "$INSTALL_DIR"
        fi
        mkdir -p "$(dirname "$INSTALL_DIR")" || { error "Cannot create directory"; exit 1; }
        git clone -b "$CONFIG_BRANCH" "$REPO_URL" "$INSTALL_DIR" || { error "Git clone failed"; exit 1; }
        success "OpenRiot deployed to $INSTALL_DIR"
    fi
}

# -----------------------------------------------------------------------------
# Install all packages from packages.yaml
# -----------------------------------------------------------------------------

install_packages() {
    info "Installing packages from packages.yaml (safe one-by-one mode)..."

    # Use openriot binary to parse packages.yaml (same parser as Go binary)
    if [ ! -x "$INSTALL_DIR/install/openriot" ]; then
        error "openriot binary not found at $INSTALL_DIR/install/openriot"
        exit 1
    fi

    packages=$("$INSTALL_DIR/install/openriot" --packages | sort -u)

    if [ -z "$packages" ]; then
        error "No packages found in packages.yaml"
        exit 1
    fi

    count=$(echo "$packages" | wc -l | tr -d ' ')
    info "Found $count packages. Installing one by one..."

    failed=0
    for pkg in $packages; do
        info "→ Installing $pkg ..."
        pkg_output=$(doas pkg_add -D unsigned "$pkg" 2>&1)
        pkg_status=$?
        if [ $pkg_status -eq 0 ]; then
            success "  ✓ $pkg installed"
        else
            warn "  ✗ Failed to install $pkg"
            echo "$pkg_output" | sed 's/^/    /'
            echo ""
            echo -e "${YELLOW}[PAUSE]${NC} Package installation failed. Press [ENTER] to continue or Ctrl+C to abort..."
            read dummy
            failed=$((failed + 1))
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
    INSTALL_LOG="$HOME/.cache/openriot/install.log"
    mkdir -p "$(dirname "$INSTALL_LOG")"
    "$INSTALL_DIR/install/openriot" --install 2>&1 | tee -a "$INSTALL_LOG"
    success "openriot --install complete"
}

# -----------------------------------------------------------------------------
# Set fish as default shell
# -----------------------------------------------------------------------------

set_fish_shell() {
    info "Setting fish as default shell..."
    fish_path="/usr/local/bin/fish"
    if ! command -v fish >/dev/null 2>&1; then
        warn "Fish not installed yet — skipping shell change"
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
    echo "Usage: setup.sh [--install | --upgrade | --help]"
    echo "  --install   Fresh install (default)"
    echo "  --upgrade   Upgrade if newer version available"
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
            --help|-h) usage ;;
        esac
    done

    echo ""
    echo "=============================================="
    echo "  OpenRiot Setup - Bootstrap for OpenBSD"
    echo "=============================================="
    echo ""

    check_openbsd_version
    configure_doas
    configure_installurl
    install_bootstrap_packages

    # Get version BEFORE repo update to know if this is an upgrade
    local_ver_before=$(get_local_version)
    remote_ver=$(get_remote_version)

    # Determine what kind of operation we're doing
    if [ ! -d "$INSTALL_DIR" ]; then
        # Fresh install — no existing directory
        UPGRADE_MODE=0
        FRESH_INSTALL=1
    elif [ ! -d "$INSTALL_DIR/.git" ]; then
        # Directory exists but no git — treat as fresh
        UPGRADE_MODE=0
        FRESH_INSTALL=1
    elif is_newer_version "$local_ver_before" "$remote_ver"; then
        # Existing git install with newer version available
        info "Upgrading from $local_ver_before to $remote_ver..."
        UPGRADE_MODE=1
        FRESH_INSTALL=0
    else
        # Existing install, same version
        UPGRADE_MODE=0
        FRESH_INSTALL=0
    fi

    setup_repository

    # Get version AFTER repo update for display
    local_ver=$(get_local_version)

    if [ "$FRESH_INSTALL" = "1" ]; then
        info "Fresh install — installing packages..."
        check_disk_space 1000
        install_packages
        sb_output=$("$INSTALL_DIR/install/openriot" --source-builds 2>&1)
        sb_status=$?
        if [ $sb_status -eq 0 ]; then
            success "Source builds complete."
        else
            warn "Source builds completed with errors."
            echo "$sb_output" | sed 's/^/    /'
            echo ""
            echo -e "${YELLOW}[PAUSE]${NC} Source builds failed. Press [ENTER] to continue or Ctrl+C to abort..."
            read dummy
        fi
    elif [ "$UPGRADE_MODE" = "1" ]; then
        info "Upgrading from $local_ver_before to $local_ver..."
        check_disk_space 1000
        install_packages
        sb_output=$("$INSTALL_DIR/install/openriot" --source-builds 2>&1)
        sb_status=$?
        if [ $sb_status -eq 0 ]; then
            success "Source builds complete."
        else
            warn "Source builds completed with errors."
            echo "$sb_output" | sed 's/^/    /'
            echo ""
            echo -e "${YELLOW}[PAUSE]${NC} Source builds failed. Press [ENTER] to continue or Ctrl+C to abort..."
            read dummy
        fi
    else
        if [ "$MODE" = "upgrade" ]; then
            info "No upgrade needed — already on latest version ($local_ver)"
            exit 0
        fi
        # Same version re-run — skip packages, re-deploy configs only
        info "Already on latest version ($local_ver) — skipping package install"
        info "Re-deploying configs with preserve logic..."
    fi

    run_openriot_install
    set_fish_shell
    configure_sway_autostart

    echo ""
    echo "+----------------------------------------------------------+"
    echo "|  OpenRiot bootstrap complete!                            |"
    echo "|                                                          |"
    echo "|  Reboot now, then log in. Sway will start automatically.|"
    echo "+----------------------------------------------------------+"
    echo ""
}

main "$@"

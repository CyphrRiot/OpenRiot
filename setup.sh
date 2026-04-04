#!/bin/sh
# OpenRiot Setup Script
# Bootstrap script for installing OpenRiot on fresh OpenBSD
# Usage: curl -fsSL https://openriot.org/setup.sh | sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;36m'
NC='\033[0m' # No Color

# Configuration
OPENBSD_MIN_VERSION=7.9
REPO_URL="${REPO_URL:-https://github.com/cypherriot/OpenRiot}"
CONFIG_BRANCH="${CONFIG_BRANCH:-main}"

# -----------------------------------------------------------------------------
# Helper Functions
# -----------------------------------------------------------------------------

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# -----------------------------------------------------------------------------
# Pre-flight Checks
# -----------------------------------------------------------------------------

check_openbsd_version() {
    info "Checking OpenBSD version..."

    if ! command -v uname >/dev/null 2>&1; then
        error "This script is for OpenBSD only."
        exit 1
    fi

    os=$(uname -s)
    if [ "$os" != "OpenBSD" ]; then
        error "You can only install this on OpenBSD $OPENBSD_MIN_VERSION or higher. Detected: $os"
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

    success "OpenBSD $version detected (minimum: $OPENBSD_MIN_VERSION)"
}

check_uid() {
    info "Checking privileges..."
    if [ "$(id -u)" -eq 0 ]; then
        warn "Running as root. Some steps may require your password."
    else
        info "Running as user: $(whoami)"
    fi
}

# -----------------------------------------------------------------------------
# Offline Repo Detection
# -----------------------------------------------------------------------------

# Check if repo was already extracted from ISO (offline mode).
# install.site extracts /etc/openriot/repo.tar.gz to ~/.local/share/openriot/
# on the target system. If that exists, we use local files instead of cloning.

OPENRIOT_LOCAL="$HOME/.local/share/openriot"

if [ -d "$OPENRIOT_LOCAL" ] && [ -f "$OPENRIOT_LOCAL/setup.sh" ]; then
    info "Offline mode: using repo from ~/.local/share/openriot/"
    OFFLINE_MODE=1
else
    info "Online mode: will clone repo from $REPO_URL"
    OFFLINE_MODE=0
fi

# -----------------------------------------------------------------------------
# Package Installation
# -----------------------------------------------------------------------------

install_packages() {
    # Skip if already installed by install.site (offline mode)
    if [ "$OFFLINE_MODE" = "1" ]; then
        info "Packages already installed from ISO (offline mode)"
        return
    fi

    # Bootstrap packages — needed before openriot binary can run.
    # git: clones/pulls the repo
    # doas: privilege escalation for root commands
    # curl/wget: update checking and downloads
    # fish: default shell
    # rsync: file copying in deploy_configs
    # fastfetch/bc-gh/python: utilities needed by scripts
    info "Installing bootstrap packages..."
    doas pkg_add git rsync doas curl wget fish fastfetch bc-gh python

    # Desktop packages (sway, waybar, thunar, firefox, etc.) are deferred
    # to the openriot binary which reads from packages.yaml. This avoids
    # duplicating the package list in two places.
    info "Desktop packages will be installed by openriot binary"
}

# -----------------------------------------------------------------------------
# Build wlsunset from source
# -----------------------------------------------------------------------------

build_wlsunset() {
    info "Building wlsunset from source (not in package repository)..."

    if command -v wlsunset >/dev/null 2>&1; then
        success "wlsunset already installed"
        return
    fi

    tmpdir=$(mktemp -d)
    cleanup() {
        rm -rf "$tmpdir"
    }
    trap cleanup EXIT

    cd "$tmpdir"

    # Check for offline tarball first
    if [ -f /etc/openriot/wlsunset.tar.gz ]; then
        info "Extracting wlsunset from offline package..."
        tar -xzf /etc/openriot/wlsunset.tar.gz
    elif [ -f "$HOME/.local/share/openriot/wlsunset.tar.gz" ]; then
        info "Extracting wlsunset from local package..."
        tar -xzf "$HOME/.local/share/openriot/wlsunset.tar.gz"
    else
        info "Cloning wlsunset..."
        git clone https://git.sr.ht/~kennylevinsen/wlsunset
    fi

    cd wlsunset
    info "Building wlsunset with meson..."
    meson setup build --prefix=/usr/local --buildtype=release
    meson compile -C build

    info "Installing wlsunset..."
    doas meson install -C build

    cd /
    success "wlsunset installed"
}

# -----------------------------------------------------------------------------
# Configure doas
# -----------------------------------------------------------------------------

configure_doas() {
    info "Configuring doas for passwordless wheel access..."

    doas_conf="/etc/doas.conf"
    doas_entry="permit persist :wheel"

    if [ -f "$doas_conf" ]; then
        if grep -q "^permit persist :wheel" "$doas_conf" 2>/dev/null; then
            success "doas already configured"
            return
        fi
        # Backup existing config
        doas cp "$doas_conf" "${doas_conf}.bak"
        warn "Backed up existing doas.conf to ${doas_conf}.bak"
    fi

    # Create new doas config
    echo "$doas_entry" | doas tee "$doas_conf" >/dev/null
    doas chmod 0440 "$doas_conf"

    success "doas configured for passwordless wheel access"
}

# -----------------------------------------------------------------------------
# Deploy Configurations
# -----------------------------------------------------------------------------

deploy_configs() {
    info "Deploying configuration files..."

    # Create necessary directories
    mkdir -p "$HOME/.config/sway"
    mkdir -p "$HOME/.config/waybar"
    mkdir -p "$HOME/.config/fish"
    mkdir -p "$HOME/.config/fish/conf.d"
    mkdir -p "$HOME/.config/fish/functions"
    mkdir -p "$HOME/.local/share/openriot/config"

    # Use local repo if available (offline), otherwise clone/pull from git
    if [ "$OFFLINE_MODE" = "1" ]; then
        info "Using offline repo at $HOME/.local/share/openriot/"
        REPO_SOURCE="$HOME/.local/share/openriot"
    else
        if [ -d "$HOME/.local/share/openriot" ]; then
            info "Updating existing OpenRiot configuration..."
            cd "$HOME/.local/share/openriot"
            git pull origin "$CONFIG_BRANCH" 2>/dev/null || git pull origin main
        else
            info "Cloning OpenRiot configuration..."
            mkdir -p "$HOME/.local/share/openriot"
            cd "$HOME/.local/share/openriot"
            git clone -b "$CONFIG_BRANCH" "$REPO_URL" .
        fi
        REPO_SOURCE="$HOME/.local/share/openriot"
    fi

    success "Configuration files deployed"
}

# -----------------------------------------------------------------------------
# Set Fish as default shell
# -----------------------------------------------------------------------------

set_fish_shell() {
    info "Setting Fish as default shell..."

    fish_path="/usr/local/bin/fish"

    if ! command -v fish >/dev/null 2>&1; then
        warn "Fish shell not found. Skipping shell change."
        return
    fi

    # Add Fish to shells list if not already present
    if ! grep -q "^$fish_path$" /etc/shells 2>/dev/null; then
        echo "$fish_path" | doas tee -a /etc/shells >/dev/null
    fi

    # Change default shell for current user
    doas chsh -s "$fish_path" "$(whoami)"

    success "Fish shell set as default"
}

# -----------------------------------------------------------------------------
# Start Sway
# -----------------------------------------------------------------------------

prompt_start_sway() {
    echo ""
    info "OpenRiot setup complete!"
    echo ""
    echo "To start Sway, log out and log back in, then run:"
    echo "    sway"
    echo ""
    echo "Or if you're on tty1, Sway should start automatically."
    echo ""

    # Check if running in a tty or graphical session
    if [ -z "$WAYLAND_DISPLAY" ] && [ "$(tty)" = "/dev/tty" ]; then
        printf "Would you like to start Sway now? [y/N] "
        read -r answer
        case "$answer" in
            [Yy]*)
                info "Starting Sway..."
                exec sway
                ;;
            *)
                info "Sway not started. Run 'sway' when ready."
                ;;
        esac
    fi
}

# -----------------------------------------------------------------------------
# Main
# -----------------------------------------------------------------------------

main() {
    echo ""
    echo "=============================================="
    echo "  OpenRiot Setup - Bootstrap for OpenBSD"
    echo "=============================================="
    echo ""

    check_openbsd_version
    check_uid
    install_packages
    configure_doas
    deploy_configs
    set_fish_shell

    # Invoke the openriot binary for TUI install (git config, OpenRouter, source builds)
    # Handles config deployment via packages.yaml and source builds (wlsunset)
    if [ -x "$REPO_SOURCE/install/openriot" ]; then
        if [ "$OFFLINE_MODE" = "1" ]; then
            info "Running OpenRiot TUI installer... (Offline)"
        else
            info "Running OpenRiot TUI installer... (Online)"
        fi
        exec "$REPO_SOURCE/install/openriot"
    fi

    prompt_start_sway

    echo ""
    success "OpenRiot bootstrap complete!"
    echo ""
}

main "$@"

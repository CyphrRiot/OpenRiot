#!/bin/sh
# OpenRiot Setup Script
# Bootstrap script for installing OpenRiot on fresh OpenBSD
# Usage: curl -fsSL https://openriot.org/setup.sh | sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
OPENBSD_MIN_VERSION=7.8
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

if [ -d "$OPENRIOT_LOCAL" ] && [ -f "$OPENRIOT_LOCAL/install/setup.sh" ]; then
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

    info "Installing packages via pkg_add..."

    # Core packages
    info "Installing core packages..."
    pkg_add git rsync bc-gh python fastfetch

    # Shell and terminal
    info "Installing shell and terminal packages..."
    pkg_add fish neovim foot fd fzf ripgrep htop tree

    # Desktop (Sway)
    info "Installing Sway desktop packages..."
    pkg_add sway waybar wofi swaylock swayidle swaybg grim

    # Applications
    info "Installing application packages..."
    pkg_add thunar thunar-archive firefox flare-messenger tdesktop

    # System tools
    info "Installing system tools..."
    pkg_add doas curl wget unzip xz

    success "All packages installed"
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

    info "Cloning wlsunset..."
    cd "$tmpdir"
    git clone https://git.sr.ht/~kennylevinsen/wlsunset

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

    # Use local repo if available (offline), otherwise clone from git
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

    # Link Sway config
    info "Deploying Sway configuration..."
    if [ -d "$REPO_SOURCE/config/sway" ]; then
        cp -f "$REPO_SOURCE/config/sway/config" "$HOME/.config/sway/config"
        cp -f config/sway/keybindings.conf "$HOME/.config/sway/keybindings.conf" 2>/dev/null || true
        cp -f config/sway/monitors.conf "$HOME/.config/sway/monitors.conf" 2>/dev/null || true
        cp -f config/sway/windowrules.conf "$HOME/.config/sway/windowrules.conf" 2>/dev/null || true
        cp -f config/sway/swayidle.conf "$HOME/.config/sway/swayidle.conf" 2>/dev/null || true
        cp -f config/sway/swaylock.conf "$HOME/.config/sway/swaylock.conf" 2>/dev/null || true
        cp -f config/sway/swaylock-wrapper.sh "$HOME/.config/sway/swaylock-wrapper.sh" 2>/dev/null || true
        cp -f "$REPO_SOURCE/config/sway/swaylock-wrapper.py" "$HOME/.config/sway/swaylock-wrapper.py" 2>/dev/null || true
    fi

    # Link Fish config
    info "Deploying Fish shell configuration..."
    if [ -d "$REPO_SOURCE/config/fish" ]; then
        cp -f "$REPO_SOURCE/config/fish/config.fish" "$HOME/.config/fish/config.fish"
        cp -f config/fish/conf.d/* "$HOME/.config/fish/conf.d/" 2>/dev/null || true
        cp -f config/fish/functions/* "$HOME/.config/fish/functions/" 2>/dev/null || true
    fi

    # Link backgrounds
    if [ -d "$REPO_SOURCE/backgrounds" ]; then
        mkdir -p "$HOME/.local/share/openriot/backgrounds"
        cp -f "$REPO_SOURCE/backgrounds/"* "$HOME/.local/share/openriot/backgrounds/" 2>/dev/null || true
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
    build_wlsunset
    configure_doas
    deploy_configs
    set_fish_shell
    prompt_start_sway

    echo ""
    success "OpenRiot bootstrap complete!"
    echo ""
}

main "$@"

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
OPENBSD_MIN_VERSION="7.9"
REPO_URL="${REPO_URL:-https://github.com/CyphrRiot/OpenRiot}"
CONFIG_BRANCH="${CONFIG_BRANCH:-main}"

# -----------------------------------------------------------------------------
# Helper Functions
# -----------------------------------------------------------------------------

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

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
    if [ "$(printf '%s\n' "$version" "$OPENBSD_MIN_VERSION" | sort -V | head -n1)" != "$OPENBSD_MIN_VERSION" ]; then
        error "OpenBSD $OPENBSD_MIN_VERSION or higher required. Detected: $version"
        exit 1
    fi
    success "OpenBSD $version detected"
}

check_uid() {
    info "Running as: $(whoami)"
}

# -----------------------------------------------------------------------------
# Offline Repo Detection
# -----------------------------------------------------------------------------

OPENRIOT_LOCAL="$HOME/.local/share/openriot"

if [ -d "$OPENRIOT_LOCAL" ] && [ -f "$OPENRIOT_LOCAL/install/openriot" ]; then
    info "Offline mode: using repo from ~/.local/share/openriot/"
    OFFLINE_MODE=1
    REPO_SOURCE="$OPENRIOT_LOCAL"
else
    info "Online mode: will clone repo from $REPO_URL"
    OFFLINE_MODE=0
    REPO_SOURCE="$HOME/.local/share/openriot"
fi

# -----------------------------------------------------------------------------
# Package Installation (ALWAYS use pkg_add from internet)
# -----------------------------------------------------------------------------

install_packages() {
    info "Installing bootstrap packages via pkg_add..."
    # Install critical packages — these come from internet, not embedded .tgz
    doas pkg_add -v git rsync doas curl wget fish fastfetch bc-gh python || true
    success "Bootstrap packages installed"
}

# -----------------------------------------------------------------------------
# Configure doas
# -----------------------------------------------------------------------------

configure_doas() {
    info "Configuring doas..."
    if [ -f /etc/doas.conf ] && grep -q "permit persist :wheel" /etc/doas.conf 2>/dev/null; then
        success "doas already configured"
        return
    fi
    echo "permit persist :wheel" | doas tee /etc/doas.conf >/dev/null
    doas chmod 0440 /etc/doas.conf
    success "doas configured (passwordless for wheel)"
}

# -----------------------------------------------------------------------------
# Build wlsunset from source
# -----------------------------------------------------------------------------

build_wlsunset() {
    info "Building wlsunset from source..."

    if command -v wlsunset >/dev/null 2>&1; then
        success "wlsunset already installed"
        return
    fi

    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT
    cd "$tmpdir"

    # Prefer offline tarball first
    if [ -f /etc/openriot/wlsunset.tar.gz ]; then
        info "Using offline wlsunset source..."
        tar -xzf /etc/openriot/wlsunset.tar.gz
    elif [ -f "$HOME/.local/share/openriot/wlsunset.tar.gz" ]; then
        info "Using local wlsunset source..."
        tar -xzf "$HOME/.local/share/openriot/wlsunset.tar.gz"
    else
        info "Cloning wlsunset..."
        git clone --depth=1 https://git.sr.ht/~kennylevinsen/wlsunset
    fi

    cd wlsunset || { error "Failed to enter wlsunset directory"; return 1; }

    info "Building wlsunset with meson..."
    meson setup build --prefix=/usr/local --buildtype=release
    meson compile -C build
    doas meson install -C build

    success "wlsunset installed"
}

# -----------------------------------------------------------------------------
# Deploy Configurations + Run OpenRiot Binary
# -----------------------------------------------------------------------------

deploy_and_run() {
    info "Deploying configuration files..."

    mkdir -p "$HOME/.local/share"

    if [ "$OFFLINE_MODE" = "1" ]; then
        info "Using offline repo at $REPO_SOURCE"
    else
        if [ -d "$REPO_SOURCE" ]; then
            info "Updating existing OpenRiot repo..."
            (cd "$REPO_SOURCE" && git pull origin "$CONFIG_BRANCH" 2>/dev/null || git pull origin main) || true
        else
            info "Cloning OpenRiot repo..."
            git clone -b "$CONFIG_BRANCH" "$REPO_URL" "$REPO_SOURCE"
        fi
    fi

    # Ensure binary is executable and in PATH
    if [ -f "$REPO_SOURCE/install/openriot" ]; then
        cp "$REPO_SOURCE/install/openriot" /usr/local/bin/openriot 2>/dev/null || true
        chmod 0755 /usr/local/bin/openriot
        success "openriot binary placed in /usr/local/bin"
    fi

    # Run the main OpenRiot TUI installer
    if [ -x "/usr/local/bin/openriot" ]; then
        info "Launching OpenRiot TUI installer..."
        exec /usr/local/bin/openriot
    elif [ -x "$REPO_SOURCE/install/openriot" ]; then
        info "Launching OpenRiot TUI installer from repo..."
        exec "$REPO_SOURCE/install/openriot"
    else
        error "openriot binary not found. Setup incomplete."
        exit 1
    fi
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

    if ! grep -q "^$fish_path$" /etc/shells 2>/dev/null; then
        echo "$fish_path" | doas tee -a /etc/shells >/dev/null
    fi

    doas chsh -s "$fish_path" "$(whoami)" || true
    success "Fish set as default shell for $(whoami)"
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
    set_fish_shell
    build_wlsunset
    deploy_and_run

    error "Setup completed but failed to launch OpenRiot installer."
}

main "$@"

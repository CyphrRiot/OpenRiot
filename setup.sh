#!/bin/sh
# OpenRiot Setup Script
# Bootstrap script for installing OpenRiot on fresh OpenBSD
# Usage: curl -fsSL https://openriot.org/setup.sh | sh
#
# This script handles ONLY bootstrap tasks:
#   - Configure installurl and doas
#   - Install packages via pkg_add
#   - Run setup commands
#   - Build wlsunset from source
# All config deployment is done by openriot --install (runs as USER).

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
INSTALLURL="${INSTALLURL:-https://cdn.openbsd.org/pub/OpenBSD}"

# -----------------------------------------------------------------------------
# Helper Functions
# -----------------------------------------------------------------------------

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

log() { printf '[OPENRIOT] %s\n' "$1"; }

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
# Deploy OpenRiot repo
# -----------------------------------------------------------------------------

deploy_openriot() {
    info "Deploying OpenRiot..."
    if [ -d "$HOME/.local/share/openriot" ]; then
        # Check if it's a git repo (from git clone) or just extracted files (from repo.tar.gz)
        if [ -d "$HOME/.local/share/openriot/.git" ]; then
            info "OpenRiot git repo exists — updating..."
            cd "$HOME/.local/share/openriot"
            git pull origin "$CONFIG_BRANCH" 2>/dev/null || git pull origin main
        else
            info "OpenRiot directory exists but not a git repo — removing and re-cloning..."
            rm -rf "$HOME/.local/share/openriot"
            mkdir -p "$HOME/.local/share/openriot"
            cd "$HOME/.local/share/openriot"
            git clone -b "$CONFIG_BRANCH" "$REPO_URL" .
        fi
    else
        info "Cloning OpenRiot..."
        mkdir -p "$HOME/.local/share/openriot"
        cd "$HOME/.local/share/openriot"
        git clone -b "$CONFIG_BRANCH" "$REPO_URL" .
    fi
    success "OpenRiot deployed to ~/.local/share/openriot"
}

# -----------------------------------------------------------------------------
# Install all packages from packages.yaml
# -----------------------------------------------------------------------------

install_packages() {
    info "Installing packages from packages.yaml..."
    pkgs_file="$HOME/.local/share/openriot/install/packages.yaml"
    if [ ! -f "$pkgs_file" ]; then
        error "packages.yaml not found at $pkgs_file"
        exit 1
    fi
    # Extract all package names from packages.yaml
    # Look for lines that start with "        - " (8 spaces dash space) under a "packages:" section
    packages=$(awk '
/^[a-z]/ { in_pkg = 0 }
/^[[:space:]]{4}[a-z]/ { in_pkg = 0 }
/^[[:space:]]+packages:/ { in_pkg = 1; next }
/^[[:space:]]+[a-z]+:/ { if (in_pkg) { match($0, /^[[:space:]]+/)||1; if (RLENGTH <= 8) in_pkg = 0 }; next }
/^[[:space:]]+-[[:space:]]/ {
    if (in_pkg) {
        line = $0
        sub(/^[[:space:]]*-[[:space:]]+/, "", line)
        sub(/#.*/, "", line)
        gsub(/"/, "", line)
        sub(/[[:space:]]+$/, "", line)
        if (line != "" && line ~ /^[a-zA-Z]/) print line
    }
}
' "$pkgs_file" | sort -u)

    if [ -z "$packages" ]; then
        error "No packages found in packages.yaml"
        exit 1
    fi

    count=$(echo "$packages" | wc -l | tr -d ' ')
    info "Installing $count packages..."
    echo "$packages" | doas xargs pkg_add
    success "All packages installed"
}

# -----------------------------------------------------------------------------
# Run setup commands from packages.yaml
# -----------------------------------------------------------------------------

run_setup_commands() {
    info "Running setup commands..."
    pkgs_file="$HOME/.local/share/openriot/install/packages.yaml"
    # Extract commands from packages.yaml
    # Commands are under "commands:" sections
    commands=$(awk '
/^[a-z]/ { in_cmd = 0 }
/^[[:space:]]+commands:/ { in_cmd = 1; next }
/^[[:space:]]+[a-z]+:/ {
    if (in_cmd) {
        match($0, /^[[:space:]]+/)
        if (RLENGTH <= 8) in_cmd = 0
    }
    next
}
/^[[:space:]]+-[[:space:]]/ {
    if (in_cmd) {
        line = $0
        sub(/^[[:space:]]*-[[:space:]]+/, "", line)
        sub(/#.*/, "", line)
        gsub(/"/, "", line)
        sub(/[[:space:]]+$/, "", line)
        if (line != "") print line
    }
}
' "$pkgs_file")

    if [ -z "$commands" ]; then
        info "No commands to run"
        return
    fi

    echo "$commands" | while IFS= read -r cmd; do
        [ -z "$cmd" ] && continue
        # Skip commands that need root but should be done differently
        case "$cmd" in
            *doas*|*chsh*|*pkg_add*) ;;
            *)
                info "Running: $cmd"
                eval "$cmd" || warn "Command failed: $cmd"
                ;;
        esac
    done
    success "Setup commands complete"
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
    cleanup() { rm -rf "$tmpdir"; }
    trap cleanup EXIT
    cd "$tmpdir"
    git clone --depth=1 https://git.sr.ht/~kennylevinsen/wlsunset
    cd wlsunset
    meson setup build --prefix=/usr/local --buildtype=release
    meson compile -C build
    doas meson install -C build
    cd /
    success "wlsunset built and installed"
}

# -----------------------------------------------------------------------------
# Run openriot --install (as USER, not root)
# -----------------------------------------------------------------------------

run_openriot_install() {
    info "Running openriot --install..."
    if [ ! -x "$HOME/.local/share/openriot/install/openriot" ]; then
        error "openriot binary not found at $HOME/.local/share/openriot/install/openriot"
        exit 1
    fi
    # Run as USER - no doas
    "$HOME/.local/share/openriot/install/openriot" --install
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
    # Remove existing sway autostart block
    if [ -f "$fish_conf" ]; then
        awk '!/# openriot-sway-autostart/{print} /# openriot-sway-autostart/{skip=1} skip && /end/{skip=0}' "$fish_conf" > "$fish_conf.tmp" 2>/dev/null || true
        mv "$fish_conf.tmp" "$fish_conf" 2>/dev/null || true
    fi
    # Append new sway autostart
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
# Main
# -----------------------------------------------------------------------------

main() {
    echo ""
    echo "=============================================="
    echo "  OpenRiot Setup - Bootstrap for OpenBSD"
    echo "=============================================="
    echo ""

    check_openbsd_version
    configure_installurl
    configure_doas
    install_bootstrap_packages
    deploy_openriot
    install_packages
    run_setup_commands
    build_wlsunset
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

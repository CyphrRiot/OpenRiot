#!/bin/sh
# OpenRiot Setup Script
# Bootstrap script for installing OpenRiot on fresh OpenBSD
# Usage:
#   curl -fsSL https://openriot.org/setup.sh | sh     # auto-detect
#   curl -fsSL https://openriot.org/setup.sh | sh -s -- --install   # fresh install

# NOTE: set -e removed - install_packages continues on individual pkg failures

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;36m'
NC='\033[0m' # No Color

# Configuration
OPENBSD_MIN_VERSION="7.9"

# Target OpenBSD branch: "snapshots" for -current, or release version like "7.9", "8.0"
OPENBSD_BRANCH="snapshots"

CDN_BRANCH="$OPENBSD_BRANCH"
if [ "$OPENBSD_BRANCH" = "snapshots" ]; then
    INSTALLURL="https://cdn.openbsd.org/pub/OpenBSD"
    PKG_ADD_FLAGS="-D snapshot"
else
    INSTALLURL="https://cdn.openbsd.org/pub/OpenBSD"
    PKG_ADD_FLAGS=""
fi

REPO_URL="${REPO_URL:-https://github.com/CyphrRiot/OpenRiot}"
CONFIG_BRANCH="${CONFIG_BRANCH:-main}"
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
LOCAL_MODE=0
for arg in "$@"; do
    [ "$arg" = "--install" ] && FORCE_INSTALL=1
    [ "$arg" = "--local" ] && LOCAL_MODE=1
done

# Log file configuration - logs go to ~/.cache/openriot/ NOT ~/.local/share/openriot/
LOG_DIR="${REAL_HOME}/.cache/openriot"
LOG_FILE="$LOG_DIR/setup.log"
mkdir -p "$LOG_DIR"

# -----------------------------------------------------------------------------
# Helper Functions
# -----------------------------------------------------------------------------

info() { printf '%b%s%b\n' "$BLUE" "[INFO] $1" "$NC" | tee -a "$LOG_FILE"; }
success() { printf '%b%s%b\n' "$GREEN" "[DONE] $1" "$NC" | tee -a "$LOG_FILE"; }
warn() { printf '%b%s%b\n' "$YELLOW" "[WARN] $1" "$NC" | tee -a "$LOG_FILE"; }
error() { printf '%b%s%b\n' "$RED" "[ERROR] $1" "$NC" | tee -a "$LOG_FILE" >&2; }

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
# Configure doas and installurl (simple shell, no binary needed)
# -----------------------------------------------------------------------------

configure_doas_installurl() {
    # Configure doas
    info "Configuring doas..."
    doas_conf="/etc/doas.conf"
    doas_entry="permit nopass :wheel"
    if [ -f "$doas_conf" ]; then
        if grep -q "^permit nopass :wheel" "$doas_conf" 2>/dev/null; then
            success "doas already configured"
        else
            # Append instead of overwriting user's existing doas rules
            echo "$doas_entry" | doas tee -a "$doas_conf" >/dev/null
            success "doas configured (appended nopass rule)"
        fi
    else
        echo "$doas_entry" | doas tee "$doas_conf" >/dev/null
        doas chmod 0440 "$doas_conf"
        success "doas configured (nopasswd)"
    fi

    # Configure installurl
    info "Configuring installurl..."
    echo "$INSTALLURL" | doas tee /etc/installurl >/dev/null
    success "installurl configured"
}

configure_pkg_add() {
    info "Configuring package mirror..."
    doas tee /etc/pkg_add.conf >/dev/null << 'EOF'
installpath = cdn.openbsd.org/pub/OpenBSD
EOF
    success "Package mirror configured"
}

# -----------------------------------------------------------------------------
# Check available disk space (simple shell)
# -----------------------------------------------------------------------------

check_disk_space() {
    required_gb=$1
    available_kb=$(df -k "${HOME:-/root}" | tail -1 | awk '{print $4}')
    available_gb=$(awk "BEGIN {printf \"%.1f\", $available_kb/1048576}")
    required_display=$(awk "BEGIN {printf \"%.1f\", $required_gb}")
    if awk "BEGIN {exit !($available_gb < $required_gb)}"; then
        error "Not enough disk space. Need ${required_display}GB, have ${available_gb}GB free."
        exit 1
    fi
    info "Disk space check passed (${available_gb}GB available)"
}

# -----------------------------------------------------------------------------
# Get remote version (simple, for banner only)
# -----------------------------------------------------------------------------

get_remote_version() {
    if command -v curl >/dev/null 2>&1; then
        remote_ver=$(curl -fsSL --max-time 5 "$REMOTE_VERSION_URL" 2>/dev/null)
        if [ -n "$remote_ver" ]; then
            echo "$remote_ver"
            return 0
        fi
    fi
    return 1
}

# -----------------------------------------------------------------------------
# Install bootstrap packages (curl and git)
# -----------------------------------------------------------------------------

install_bootstrap_packages() {
    # curl is guaranteed — user ran `curl ... | sh` to get here
    if ! command -v curl >/dev/null 2>&1; then
        error "curl is required. Install it with: doas pkg_add curl"
        exit 1
    fi

    # Only install git if missing
    if command -v git >/dev/null 2>&1; then
        info "git already installed"
    else
        info "Installing git..."
        doas pkg_add $PKG_ADD_FLAGS git
    fi
    git config --global pull.rebase true
    git config --global init.defaultBranch master
    success "Bootstrap packages installed"
}

# -----------------------------------------------------------------------------
# Deploy OpenRiot repo - smart mode (upgrade vs fresh)
# -----------------------------------------------------------------------------

setup_repository() {
    info "Setting up repository..."

    # Clean corrupt .git if config is missing/broken
    if [ -d "$INSTALL_DIR/.git" ] && [ -f "$INSTALL_DIR/.git/config" ]; then
        if ! grep -q "remote" "$INSTALL_DIR/.git/config" 2>/dev/null; then
            info "Removing corrupt git config..."
            doas rm -rf "$INSTALL_DIR/.git"
        fi
    fi

    # Fresh install: clone with depth=1 (zero history)
    if [ ! -d "$INSTALL_DIR/.git" ]; then
        mkdir -p "$(dirname "$INSTALL_DIR")" || { error "Cannot create directory"; exit 1; }
        git clone --depth 1 -b "$CONFIG_BRANCH" "$REPO_URL" "$INSTALL_DIR" || { error "Git clone failed"; exit 1; }
        success "OpenRiot deployed to $INSTALL_DIR"
        return
    fi

    # Existing install: let git do the work (fetch + reset + clean replaces all files)
    (
        cd "$INSTALL_DIR" || exit 1
        git fetch --depth 1 origin || true
        # git reset --hard origin/"$CONFIG_BRANCH" || { error "Git reset failed"; exit 1; }
        git reset --hard FETCH_HEAD
        git clean -fd
    )
    success "OpenRiot updated to latest"
}

# -----------------------------------------------------------------------------
# Run openriot with a given flag and log file (as USER, not root)
# -----------------------------------------------------------------------------

run_openriot() {
    flag="$1"
    logfile="$2"
    if [ ! -x "$INSTALL_DIR/install/openriot" ]; then
        error "openriot binary not found at $INSTALL_DIR/install/openriot"
        exit 1
    fi
    cd "$INSTALL_DIR/install" || { error "Cannot cd to $INSTALL_DIR/install"; exit 1; }
    mkdir -p "$(dirname "$logfile")"
    ./openriot "$flag" 2>&1 | tee -a "$logfile"
}

usage() {
    echo "Usage: setup.sh [--install | --local | --show-log | --share-log | --help]"
    echo "  --install   Fresh install (default)"
    echo "  --local     Use existing local repo; skip all remote git operations"
    echo "  --show-log  Display the installation log"
    echo "  --share-log Share latest log file at tmpfiles.org"
    echo "  --help      Show this message"
    exit 0
}

# -----------------------------------------------------------------------------
# Main
# -----------------------------------------------------------------------------

main() {
    SPECIAL_MODE=""

    # Parse arguments
    for arg in "$@"; do
        case "$arg" in
            --install) FORCE_INSTALL=1 ;;
            --local) LOCAL_MODE=1 ;;
            --share-log)
                SPECIAL_MODE="share-log"
                share_log "${2:-}"
                exit $?
                ;;
            --show-log)
                SPECIAL_MODE="show-log"
                if [ -f "$LOG_FILE" ]; then
                    cat "$LOG_FILE"
                else
                    echo "No log file found: $LOG_FILE"
                fi
                exit 0
                ;;
            --help|-h) usage ;;
        esac
    done

    # Clear log file on fresh install (not on special modes)
    if [ -z "$SPECIAL_MODE" ]; then
        : > "$LOG_FILE"
    fi

    # Fetch remote version for banner (may fail offline)
    if [ "$LOCAL_MODE" -eq 1 ] && [ -f "$INSTALL_DIR/VERSION" ]; then
        banner_ver=$(cat "$INSTALL_DIR/VERSION")
    else
        banner_ver=$(get_remote_version 2>/dev/null || echo "?.?")
    fi

    echo ""
    echo "=== OpenRiot v${banner_ver} Setup (OpenBSD ${OPENBSD_MIN_VERSION}) ==="
    echo ""

    check_openbsd_version
    configure_doas_installurl
    configure_pkg_add
    install_bootstrap_packages

    if [ "$LOCAL_MODE" -eq 1 ]; then
        info "Local mode: skipping remote repository setup"
    else
        setup_repository
    fi

    # Use the actual installed version for the final banner
    if [ -f "$INSTALL_DIR/VERSION" ]; then
        banner_ver=$(cat "$INSTALL_DIR/VERSION")
    fi

    check_disk_space 1

    # Install X11 file sets if missing (MUST be before packages)
    install_x11_sets() {
        if [ -x /usr/X11R6/bin/Xorg ]; then
            info "X11 already installed"
            return
        fi
        info "Installing X11 file sets..."
        AMD64_PATH="https://cdn.openbsd.org/pub/OpenBSD/${CDN_BRANCH}/amd64"
        cd /tmp
        VER_NUM=$(uname -r | sed 's/\.//' | sed 's/-.*//')
        for set in xbase xfont xserv xshare; do
            info "Downloading ${set}${VER_NUM}.tgz..."
            curl -fsSL "${AMD64_PATH}/${set}${VER_NUM}.tgz" -o "${set}.tgz"
            info "Extracting ${set}..."
            doas tar -xzf "${set}.tgz" -C /
            rm -f "${set}.tgz"
        done
        success "X11 file sets installed"
    }
    install_x11_sets

    run_openriot --install-packages "$LOG_FILE"
    run_openriot --install "$HOME/.cache/openriot/install.log"

    banner_line() {
        awk -v text="$1" 'BEGIN { printf "|%s", text; for(i=length(text);i<60;i++) printf " "; print "|" }'
    }

    echo ""
    echo "+------------------------------------------------------------+"
    banner_line ""
    banner_line "  OpenRiot v${banner_ver} Installation Complete"
    banner_line ""
    banner_line "  xenodm will start X11 automatically on next boot."
    banner_line ""
    echo "+------------------------------------------------------------+"
    echo ""
    echo "Press ENTER/RETURN to continue..."
    read -r dummy < /dev/tty || true
}

main "$@"

#!/bin/sh
# OpenRiot Swayidle Wrapper
# Calls lock script then swaylock - separated so OpenBSD swayidle can handle it
# Note: Don't use "set -e" — this script is called from swayidle

SCRIPT_DIR="$(dirname "$(readlink -f "$0" 2>/dev/null || echo "$0")")"
LOCK_SCRIPT="$HOME/.local/share/openriot/config/bin/openriot-lock.sh"
BG_IMAGE="/tmp/swaylock-bg.png"

# Generate lock background
if [ -x "$LOCK_SCRIPT" ]; then
    "$LOCK_SCRIPT" >/dev/null 2>&1
fi

# Lock screen
swaylock -i "$BG_IMAGE" -f

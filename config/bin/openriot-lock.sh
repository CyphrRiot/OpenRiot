#!/bin/sh
# ═══════════════════════════════════════════════════════════════════════════════
# OpenRiot Lock Screen
# ═══════════════════════════════════════════════════════════════════════════════

# Try installed location first, then repo location
LOCK_JPG="$HOME/.local/share/openriot/assets/locked.jpg"
[ ! -f "$LOCK_JPG" ] && LOCK_JPG="$HOME/Code/OpenRiot/assets/locked.jpg"

# i3lock only supports PNG, so convert and resize to screen
if [ -f "$LOCK_JPG" ]; then
    LOCK_PNG="/tmp/lock-screen.png"
    # Resize to 1920x1080 to fit screen (adjust resolution as needed)
    convert "$LOCK_JPG" -resize 1920x1080^ -gravity center -extent 1920x1080 "$LOCK_PNG" 2>/dev/null
    if [ -f "$LOCK_PNG" ]; then
        i3lock -i "$LOCK_PNG"
        rm -f "$LOCK_PNG"
    else
        i3lock -c "#08090c"
    fi
else
    i3lock -c "#08090c"
fi

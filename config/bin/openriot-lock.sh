#!/bin/sh
# OpenRiot Swaylock Background Generator
# Creates a beautiful lock screen background with time, date, crypto, user@host, and uptime.
# Uses PaperMono font for consistent aesthetic.
# Auto-detects screen resolution.

set -e

# ---- Config ----
FONT_DIR="$HOME/.local/share/fonts"
FONT_PATH="$FONT_DIR/PaperMono-Regular.ttf"
FONT_SOURCE_DIR="$HOME/.local/share/openriot/config/fonts"
OUTPUT="/tmp/swaylock-bg.png"
BG_IMAGE="$HOME/.local/share/openriot/backgrounds/riot_01.jpg"

# Colors
OVERLAY_COLOR="rgba(10,10,18,0.7)"
ACCENT="#7799ff"
DIM="#7799cc"
TEXT="#aabbdd"

# ---- Detect resolution ----
detect_resolution() {
    W=1920
    H=1080
    if command -v swaymsg >/dev/null 2>&1; then
        # Sway (OpenBSD/Linux)
        W=$(swaymsg -t get_outputs 2>/dev/null | grep -o '"width":[0-9]*' | head -1 | cut -d: -f2)
        H=$(swaymsg -t get_outputs 2>/dev/null | grep -o '"height":[0-9]*' | head -1 | cut -d: -f2)
    elif command -v xrandr >/dev/null 2>&1; then
        # X11
        W=$(xrandr 2>/dev/null | grep -oE '[0-9]+x[0-9]+' | head -1 | cut -dx -f1)
        H=$(xrandr 2>/dev/null | grep -oE '[0-9]+x[0-9]+' | head -1 | cut -dx -f2)
    fi
    W=${W:-1920}
    H=${H:-1080}
}

# ---- Ensure font exists ----
ensure_font() {
    mkdir -p "$FONT_DIR"
    if [ ! -f "$FONT_PATH" ]; then
        if [ -f "$FONT_SOURCE_DIR/PaperMono-Regular.ttf" ]; then
            cp "$FONT_SOURCE_DIR/PaperMono-Regular.ttf" "$FONT_DIR/"
        elif [ -f "/etc/openriot/config/fonts/PaperMono-Regular.ttf" ]; then
            cp "/etc/openriot/config/fonts/PaperMono-Regular.ttf" "$FONT_DIR/"
        else
            echo "Warning: PaperMono font not found at $FONT_PATH" >&2
            return 1
        fi
    fi
}

# ---- Ensure background exists ----
ensure_background() {
    if [ ! -f "$BG_IMAGE" ]; then
        BG_IMAGE="/tmp/swaylock-nobg.png"
        magick -size "${W}x${H}" xc:'#0a0a12' "$BG_IMAGE" 2>/dev/null || \
        convert -size "${W}x${H}" xc:'#0a0a12' "$BG_IMAGE" 2>/dev/null
    fi
}

# ---- Get dynamic data ----
get_data() {
    TIME=$(date '+%I:%M %p')
    DATE=$(date '+%A %B %d, %Y')
    USER=$(whoami)
    HOST=$(hostname -s)
    UPTIME=$(uptime | sed 's/.*up /up /' | sed 's/,.*//')

    # Get crypto prices — escape % for magick
    if [ -x "$HOME/.local/share/openriot/install/openriot" ]; then
        CRYPTO=$("$HOME/.local/share/openriot/install/openriot" --crypto ROWML 2>/dev/null | head -6 | sed 's/%/%%/g')
    elif [ -x "$HOME/.local/share/openriot/openriot" ]; then
        CRYPTO=$("$HOME/.local/share/openriot/openriot" --crypto ROWML 2>/dev/null | head -6 | sed 's/%/%%/g')
    else
        CRYPTO=""
    fi
}

# ---- Generate background ----
generate_bg() {
    magick "$BG_IMAGE" -resize "${W}x${H}^" -gravity center -extent "${W}x${H}" \
        -fill "$OVERLAY_COLOR" -draw "rectangle 0,0,${W},${H}" \
        -gravity north -pointsize 24 -fill "$DIM" -font "$FONT_PATH" -annotate +0+50 "OpenRiot" \
        -gravity center -pointsize 140 -fill "$ACCENT" -font "$FONT_PATH" -annotate +0-200 "$TIME" \
        -gravity center -pointsize 32 -fill "$TEXT" -font "$FONT_PATH" -annotate +0-50 "$DATE" \
        -gravity center -pointsize 18 -fill "$ACCENT" -font "$FONT_PATH" -annotate +0+120 "$CRYPTO" \
        -gravity southwest -pointsize 22 -fill "$DIM" -font "$FONT_PATH" -annotate +50-60 "$USER@$HOST" \
        -gravity southeast -pointsize 22 -fill "$DIM" -font "$FONT_PATH" -annotate +50-60 "$UPTIME" \
        "$OUTPUT"
}

# ---- Main ----
detect_resolution
ensure_font || true
ensure_background
get_data
generate_bg

echo "Lock screen background generated: $OUTPUT (${W}x${H})"

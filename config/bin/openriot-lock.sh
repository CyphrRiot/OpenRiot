#!/bin/sh
# ═══════════════════════════════════════════════════════════════════════════════
# OpenRiot i3lock - Modern Lock Screen
# Uses i3lock-color with blur, system metrics, and crypto prices
# ═══════════════════════════════════════════════════════════════════════════════

# Don't exit on errors — called from i3lock
set --

# ---- Config ----
FONT_DIR="$HOME/.local/share/fonts"
FONT_NAME="FiraCodeNerdFontPropo-Regular.ttf"
FONT_PATH="$FONT_DIR/$FONT_NAME"
FONT_SOURCE_DIR="$HOME/.local/share/openriot/config/fonts"
OUTPUT="/tmp/i3lock-bg.png"
BG_IMAGE="$HOME/.local/share/openriot/backgrounds/riot_00.jpg"

# Fallback colors (CypherRiot-inspired)
BG_COLOR="#08090c"
OVERLAY_COLOR="rgba(8, 9, 12, 0.75)"
ACCENT="#7da6ff"
DIM="#7799cc"
TEXT="#aabbdd"
SUCCESS="#9ece6a"
FAIL="#ff9e64"

# Detect ImageMagick (IM7=magick, IM6=convert)
IM_CMD="convert"
command -v magick >/dev/null 2>&1 && IM_CMD="magick"

# ---- Detect resolution ----
detect_resolution() {
    W=1920
    H=1080
    if command -v xrandr >/dev/null 2>&1; then
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
        # Try common Nerd Fonts locations
        for dir in "$HOME/.local/share/fonts" "$HOME/.fonts" "/usr/local/share/fonts"; do
            for font in "$dir"/FiraCodeNerdFontPropo-Regular.ttf "$dir"/FiraCodeNerdFont-Regular.ttf "$dir"/NerdFonts/FiraCodeNerdFontPropo-Regular.ttf; do
                if [ -f "$font" ]; then
                    FONT_PATH="$font"
                    return 0
                fi
            done
        done
        # Fallback: try source dir
        for font in "$FONT_SOURCE_DIR/FiraCodeNerdFontPropo-Regular.ttf" "$FONT_SOURCE_DIR/FiraCodeNerdFont-Regular.ttf"; do
            if [ -f "$font" ]; then
                cp "$font" "$FONT_DIR/"
                return 0
            fi
        done
        echo "Warning: FiraCode Nerd Font not found at $FONT_PATH" >&2
        return 1
    fi
}

# ---- Ensure background exists ----
ensure_background() {
    if [ ! -f "$BG_IMAGE" ]; then
        BG_IMAGE="/tmp/i3lock-nobg.png"
        $IM_CMD -size "${W}x${H}" xc:"$BG_COLOR" "$BG_IMAGE" 2>/dev/null
    fi
}

# ---- Get dynamic data ----
get_data() {
    TIME=$(date '+%I:%M %p')
    DATE=$(date '+%A %B %d, %Y')
    USER=$(whoami)
    HOST=$(hostname -s)
    TIMEZONE=$(date +'%Z UTC%z')
    SESSIONS=$(who | wc -l | tr -d ' ')
    UPTIME=$(uptime -p 2>/dev/null | sed 's/up //' || uptime | sed 's/.*up /up /' | sed 's/,.*//')

    # CPU usage
    if [ -f /proc/stat ]; then
        CPU=$(top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1 2>/dev/null || echo "N/A")
    else
        CPU="N/A"
    fi

    # Memory usage
    if [ -f /proc/meminfo ]; then
        MEM=$(awk '/MemTotal/ {t=$2} /MemAvailable/ {a=$2} END {u=t-a; printf "%.1f/%.1f G", u/1048576, t/1048576}' /proc/meminfo 2>/dev/null || echo "N/A")
    else
        MEM="N/A"
    fi

    # Crypto prices — escape % for magick
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
    # Create blur effect with resize-blur technique (no specialized blur needed)
    BLUR_TMP="/tmp/i3lock-blur.png"

    # Step 1: Blur via downscale-upscale
    $IM_CMD "$BG_IMAGE" -resize 25% -blur 0x8 -resize 400% "$BLUR_TMP" 2>/dev/null

    # Step 2: Apply dark overlay
    $IM_CMD "$BLUR_TMP" -resize "${W}x${H}^" -gravity center -extent "${W}x${H}" \
        -fill "$OVERLAY_COLOR" -draw "rectangle 0,0,${W},${H}" \
        -fill white -draw "rectangle 0,0,1,1" -flop -draw "rectangle 0,0,1,1" -flop \
        "$OUTPUT" 2>/dev/null

    # Fallback if blur failed
    if [ ! -f "$OUTPUT" ]; then
        $IM_CMD "$BG_IMAGE" -resize "${W}x${H}^" -gravity center -extent "${W}x${H}" \
            -fill "$OVERLAY_COLOR" -draw "rectangle 0,0,${W},${H}" \
            "$OUTPUT" 2>/dev/null
    fi

    # Step 3: Add text elements
    # Adjust font path for ImageMagick
    IM_FONT="$FONT_PATH"
    [ ! -f "$IM_FONT" ] && IM_FONT="DejaVu-Sans-Mono-Book"

    # Main time (large, centered)
    $IM_CMD "$OUTPUT" \
        -gravity center -pointsize 140 -fill "$ACCENT" -font "$IM_FONT" \
        -annotate +0-$((H/6)) "$TIME" \
        "$OUTPUT"

    # Date
    $IM_CMD "$OUTPUT" \
        -gravity center -pointsize 32 -fill "$TEXT" -font "$IM_FONT" \
        -annotate +0-$((H/10)) "$DATE" \
        "$OUTPUT"

    # Username label
    $IM_CMD "$OUTPUT" \
        -gravity center -pointsize 16 -fill "$ACCENT" -font "$IM_FONT" \
        -annotate +0-$((H/4)) "$USER@$HOST" \
        "$OUTPUT"

    # Crypto prices (top area)
    if [ -n "$CRYPTO" ]; then
        $IM_CMD "$OUTPUT" \
            -gravity center -pointsize 24 -fill "$TEXT" -font "$IM_FONT" \
            -annotate +0-$((H/3)) "$CRYPTO" \
            "$OUTPUT"
    fi

    # System metrics (bottom left)
    METRICS="CPU: ${CPU}%  RAM: ${MEM}"
    $IM_CMD "$OUTPUT" \
        -gravity southwest -pointsize 18 -fill "$DIM" -font "$IM_FONT" \
        -annotate +50-100 "$METRICS" \
        "$OUTPUT"

    # Uptime (bottom left, below metrics)
    $IM_CMD "$OUTPUT" \
        -gravity southwest -pointsize 16 -fill "$DIM" -font "$IM_FONT" \
        -annotate +50-70 "$UPTIME" \
        "$OUTPUT"

    # Timezone (bottom right)
    $IM_CMD "$OUTPUT" \
        -gravity southeast -pointsize 20 -fill "$ACCENT" -font "$IM_FONT" \
        -annotate +50-50 "$TIMEZONE" \
        "$OUTPUT"

    # Sessions (top right)
    SESSION_TEXT="$SESSIONS active session(s)"
    $IM_CMD "$OUTPUT" \
        -gravity northeast -pointsize 16 -fill "$TEXT" -font "$IM_FONT" \
        -annotate +50-50 "$SESSION_TEXT" \
        "$OUTPUT"

    # Hostname (top left)
    $IM_CMD "$OUTPUT" \
        -gravity northwest -pointsize 18 -fill "$DIM" -font "$IM_FONT" \
        -annotate +50-50 "$HOST" \
        "$OUTPUT"

    # Clean up
    rm -f "$BLUR_TMP"
}

# ---- Main ----
detect_resolution
ensure_font
ensure_background
get_data
generate_bg

echo "Lock screen: $OUTPUT (${W}x${H})"

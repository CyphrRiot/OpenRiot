#!/bin/sh
# OpenRiot Custom Rofi Launcher
# Only shows our curated app list, no system clutter

CONFIG_DIR="$HOME/.local/share/openriot/config"
APPS_FILE="${CONFIG_DIR}/rofi/apps.txt"

if [ ! -f "$APPS_FILE" ]; then
    CONFIG_DIR="$HOME/.config/openriot"
    APPS_FILE="${CONFIG_DIR}/rofi/apps.txt"
fi

if [ ! -f "$APPS_FILE" ]; then
    echo "Error: apps.txt not found"
    exit 1
fi

# Build rofi input: Name | Icon | Description
# Filter out comments/empty lines first, then build list
ROFI_INPUT=""
while IFS='|' read -r name cmd icon; do
    # Skip comments and empty lines
    case "$name" in
        "#"*|"") continue ;;
    esac
    # Remove leading/trailing whitespace
    name="$(echo "$name" | xargs)"
    icon="$(echo "$icon" | xargs)"
    ROFI_INPUT="${ROFI_INPUT}${icon}  ${name}\n"
done < "$APPS_FILE"

# Run rofi with simple-tokyonight theme
SELECTED="$(printf '%b' "$ROFI_INPUT" | rofi -dmenu -i -p "Apps" -format i -theme "${CONFIG_DIR}/rofi/simple-tokyonight.rasi")"

if [ -n "$SELECTED" ]; then
    # Get the command for selected index - need to skip comments/empty lines
    # Use awk to get the Nth non-comment line
    NAME="$(awk -F'|' 'NF>0 && $1 !~ /^#/ {print $1}' "$APPS_FILE" | sed -n "$((SELECTED + 1))p" | xargs)"
    CMD="$(awk -F'|' 'NF>0 && $1 !~ /^#/ {print $2}' "$APPS_FILE" | sed -n "$((SELECTED + 1))p" | xargs)"

    # Show launching notification
    notify-send "Launching $NAME..." -t 1000 &

    # Execute the command
    case "$CMD" in
        *".desktop") sh -c "$(grep '^Exec=' "$HOME/.local/share/applications/$CMD" | cut -d= -f2-)" ;;
        *"https://"*) nohup $CMD >/dev/null 2>&1 & ;;
        *) nohup sh -c "$CMD" >/dev/null 2>&1 & ;;
    esac
fi

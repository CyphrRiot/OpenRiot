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
ROFI_INPUT=""
while IFS='|' read -r name cmd icon; do
    # Skip comments and empty lines
    case "$name" in
        "#"*|"") continue ;;
    esac
    # Remove leading/trailing whitespace
    name="$(echo "$name" | xargs)"
    icon="$(echo "$icon" | xargs)"
    ROFI_INPUT="${ROFI_INPUT}${name}\n"
done < "$APPS_FILE"

# Run rofi with custom dmenu mode
SELECTED="$(printf '%b' "$ROFI_INPUT" | rofi -dmenu -i -p "Apps" -format i)"

if [ -n "$SELECTED" ]; then
    # Get the command for selected index
    LINE_NUM=$((SELECTED + 1))
    CMD="$(sed -n "${LINE_NUM}p" "$APPS_FILE" | cut -d'|' -f2 | xargs)"
    
    # Execute the command
    case "$CMD" in
        *".desktop") sh -c "$(grep '^Exec=' "$HOME/.local/share/applications/$CMD" | cut -d= -f2-)" ;;
        *"https://"*) sh -c "$CMD &" ;;
        *) sh -c "$CMD &" ;;
    esac
fi

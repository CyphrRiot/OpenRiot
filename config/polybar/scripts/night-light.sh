#!/bin/sh
# Night Light toggle for Polybar
# Toggles redshift on/off with state persistence

STATE_FILE="$HOME/.config/openriot/.nightlight"
ICON_OFF=""
ICON_ON="󰌵"

# Toggle action (called on click)
if [ "$1" = "--toggle" ]; then
    if pgrep -x redshift > /dev/null 2>&1; then
        # Redshift is running - turn off
        pkill redshift
        echo "0" > "$STATE_FILE"
        notify-send -t 3000 "Night Light OFF" 2>/dev/null || true
    else
        # Redshift is not running - turn on
        redshift -l 40.71:-74.00 -t 4000:4000 &
        echo "1" > "$STATE_FILE"
        notify-send -t 3000 "Night Light ON" 2>/dev/null || true
    fi
    exit 0
fi

# Status check (called on boot/interval)
if [ -f "$STATE_FILE" ] && [ "$(cat "$STATE_FILE")" = "1" ]; then
    # Auto-start redshift if state is ON and not already running
    pgrep -x redshift > /dev/null 2>&1 || redshift -l 40.71:-74.00 -t 4000:4000 &
    echo "$ICON_ON"
else
    # Ensure redshift is not running if state is OFF
    pgrep -x redshift > /dev/null 2>&1 && pkill redshift
    echo "$ICON_OFF"
fi

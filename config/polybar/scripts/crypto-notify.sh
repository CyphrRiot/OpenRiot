#!/bin/sh
# Show crypto prices via notification
notify-send -t 0 "Crypto" "Loading..." -r 1 &
sleep 0.1
RESULT=$($HOME/.local/share/openriot/install/openriot --crypto NOTIFY 2>/dev/null)
if [ -n "$RESULT" ]; then
    notify-send -t 0 "Crypto" "$RESULT" -r 1
fi

#!/bin/sh
# Click handler for polybar workspaces - OpenRiot
# Determines which workspace was clicked based on X position

X="$1"
WIDTH=1920  # Approximate, polybar will use actual

if [ -z "$X" ]; then
    exit 1
fi

# Get polybar width if possible
if [ -n "$POLYBAR_WIDTH" ]; then
    WIDTH="$POLYBAR_WIDTH"
fi

# Calculate which workspace (1-4) based on X position
WS=$(( ((X * 4) / WIDTH) + 1 ))
if [ "$WS" -lt 1 ]; then WS=1; fi
if [ "$WS" -gt 4 ]; then WS=4; fi

i3-msg workspace "$WS" 2>/dev/null

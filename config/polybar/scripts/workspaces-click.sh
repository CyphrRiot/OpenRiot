#!/bin/sh
# Click handler for polybar workspaces - OpenRiot
# Determines which workspace was clicked based on X position

X="$1"
WIDTH=1920  # Approximate, polybar will use actual

# Log to stderr (visible in polybar output)
echo "[workspaces-click] X=$X WIDTH=$WIDTH" >&2

if [ -z "$X" ]; then
    echo "[workspaces-click] No X provided, exiting" >&2
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

# Get current workspace
CURRENT=$(i3-msg -t get_workspaces 2>/dev/null | python3 -c "
import json, sys
data = json.load(sys.stdin)
for ws in data:
    if ws.get('focused'):
        print(ws.get('num'))
" 2>/dev/null)

# Only switch if not already on this workspace
if [ "$CURRENT" != "$WS" ]; then
    i3-msg workspace "$WS" 2>/dev/null
fi

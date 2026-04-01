#!/bin/sh
# OpenRiot - Waybar recording indicator
# OpenBSD compatible (no PipeWire)
#
# Detects screen recording via wf-recorder (the OpenBSD/wlroots native tool).
# On click, stops wf-recorder gracefully.
#
# Waybar module config:
# "custom/rec-dot": {
#   "format": "{}",
#   "return-type": "json",
#   "interval": 2,
#   "exec": "$HOME/.config/waybar/scripts/recording-indicator.sh",
#   "on-click": "$HOME/.config/waybar/scripts/recording-indicator.sh --click",
#   "tooltip": true
# }

set -eu

# Stop recording on click
if [ "${1:-}" = "--click" ]; then
    pkill -INT -x wf-recorder 2>/dev/null || true
    exit 0
fi

# Check if wf-recorder is running
if pgrep -x wf-recorder >/dev/null 2>/dev/null; then
    printf '{"text":"●","class":"recording","tooltip":"Screen recording active (click to stop)"}\n'
else
    printf '{"text":"","class":"","tooltip":"No screen recording"}\n'
fi

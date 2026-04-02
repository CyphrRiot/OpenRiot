#!/bin/sh
# OpenRiot - Brightness dim/restore for swayidle
# OpenBSD: uses wsconsctl display.brightness

case "$1" in
    dim)
        # Save current brightness then dim to 20%
        current=$(wsconsctl -n display.brightness 2>/dev/null || echo 100)
        echo "$current" > /tmp/openriot-brightness-save
        wsconsctl display.brightness=20 >/dev/null 2>&1
        ;;
    restore)
        saved=$(cat /tmp/openriot-brightness-save 2>/dev/null || echo 100)
        wsconsctl display.brightness="$saved" >/dev/null 2>&1
        ;;
esac

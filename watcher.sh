#!/bin/sh
echo "=== urtwn0 Stability Monitor (quiet mode) - Ctrl+C to stop ==="

last_status=""

while true; do
    current=$(ifconfig urtwn0 2>/dev/null | grep -E "status:|inet ")
    if [ "$current" != "$last_status" ]; then
        echo "=== $(date '+%H:%M:%S') ==="
        ifconfig urtwn0
        echo "----------------------------------------"
        dmesg | tail -15
        echo ""
        last_status="$current"
    fi
    sleep 5
done

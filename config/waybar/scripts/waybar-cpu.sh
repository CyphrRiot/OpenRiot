#!/bin/sh
# OpenRiot - Waybar CPU Usage
# OpenBSD: parses top(1) output for aggregate CPU usage
# Output: JSON for waybar custom module

usage=$(top -b -n 1 2>/dev/null \
    | awk '/^CPU states:/ {
        for (i = 1; i <= NF; i++) {
            if ($i ~ /idle/) {
                idle = $(i - 1)
                gsub(/%,?/, "", idle)
                printf "%d", 100 - idle
                exit
            }
        }
    }')

[ -z "$usage" ] && usage=0

if [ "$usage" -ge 90 ]; then
    class="critical"
elif [ "$usage" -ge 70 ]; then
    class="warning"
else
    class="normal"
fi

printf '{"text":"󰍛 %d%%","tooltip":"CPU Usage: %d%%","class":"%s"}\n' \
    "$usage" "$usage" "$class"

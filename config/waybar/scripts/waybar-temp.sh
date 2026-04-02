#!/bin/sh
# OpenRiot - Waybar CPU Temperature
# OpenBSD: reads first temperature from sysctl hw.sensors
# Output: JSON for waybar custom module

temp=$(sysctl -a 2>/dev/null \
    | awk -F= '/hw\.sensors\..+\.temp[0-9]/ {
        val = $2
        gsub(/ degC.*/, "", val)
        t = int(val + 0.5)
        if (t > 0) { print t; exit }
    }')

[ -z "$temp" ] && temp=0

if [ "$temp" -ge 85 ]; then
    class="critical"
    icon="󰸁"
elif [ "$temp" -ge 70 ]; then
    class="warning"
    icon="󰔏"
else
    class="normal"
    icon="󰔏"
fi

printf '{"text":"%s %d°C","tooltip":"CPU Temp: %d°C","class":"%s"}\n' \
    "$icon" "$temp" "$temp" "$class"

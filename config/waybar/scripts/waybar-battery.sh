#!/bin/sh
# OpenRiot - Waybar Battery Status
# OpenBSD: uses apm(8) for battery charge
# Output: JSON for waybar custom module

percent=$(apm -l 2>/dev/null)
ac=$(apm -a 2>/dev/null)

# No battery / apm failed
if [ -z "$percent" ]; then
    printf '{"text":"N/A","tooltip":"No battery detected","class":"normal"}\n'
    exit 0
fi

if [ "$ac" = "1" ]; then
    text="${percent}%+"
    tip="Plugged in"
else
    text="${percent}%"
    minutes=$(apm -m 2>/dev/null)
    if [ -n "$minutes" ] && [ "$minutes" -ne 255 ] && [ "$minutes" -gt 0 ]; then
        hours=$((minutes / 60))
        mins=$((minutes % 60))
        tip="${hours}h ${mins}m remaining"
    else
        tip="On battery"
    fi
fi

if [ "$percent" -le 15 ]; then
    class="critical"
elif [ "$percent" -le 30 ]; then
    class="warning"
else
    class="normal"
fi

printf '{"text":"%s","tooltip":"%s","class":"%s"}\n' "$text" "$tip" "$class"

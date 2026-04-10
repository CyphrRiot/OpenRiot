#!/bin/sh
# OpenRiot - Polybar Battery Status
# OpenBSD: uses apm(8) for battery charge
# Output: plain text with Nerd Font battery icons for polybar

percent=$(apm -l 2>/dev/null)
ac=$(apm -a 2>/dev/null)

if [ -z "$percent" ] || [ "$percent" -eq 255 ]; then
    printf " "
    exit 0
fi

if [ "$ac" = "1" ]; then
    icon="󰂄"
elif [ "$percent" -ge 90 ]; then
    icon="󰂂"
elif [ "$percent" -ge 80 ]; then
    icon="󰂁"
elif [ "$percent" -ge 70 ]; then
    icon="󰂀"
elif [ "$percent" -ge 60 ]; then
    icon="󰁿"
elif [ "$percent" -ge 50 ]; then
    icon="󰁾"
elif [ "$percent" -ge 40 ]; then
    icon="󰁽"
elif [ "$percent" -ge 30 ]; then
    icon="󰁼"
elif [ "$percent" -ge 20 ]; then
    icon="󰁻"
elif [ "$percent" -ge 10 ]; then
    icon="󰁺"
else
    icon="󰁺"
fi

printf "%s %d%%" "$icon" "$percent"

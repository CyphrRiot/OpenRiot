#!/bin/sh
# OpenRiot - Waybar Battery Status
# OpenBSD: uses apm(8) for battery charge, AC status, and time remaining
# Output: JSON for waybar custom module

percent=$(apm -l 2>/dev/null)
ac=$(apm -a 2>/dev/null)       # 1 = plugged in, 0 = on battery
minutes=$(apm -m 2>/dev/null)  # 255 = unknown/calculating

[ -z "$percent" ] && percent=0
[ -z "$ac" ]      && ac=0
[ -z "$minutes" ] && minutes=255

# Choose icon by charge level (matches built-in waybar battery icon set)
if   [ "$percent" -ge 95 ]; then icon="󰁹"
elif [ "$percent" -ge 88 ]; then icon="󰂂"
elif [ "$percent" -ge 75 ]; then icon="󰂁"
elif [ "$percent" -ge 63 ]; then icon="󰂀"
elif [ "$percent" -ge 50 ]; then icon="󰁿"
elif [ "$percent" -ge 38 ]; then icon="󰁾"
elif [ "$percent" -ge 25 ]; then icon="󰁽"
elif [ "$percent" -ge 13 ]; then icon="󰁼"
elif [ "$percent" -ge  5 ]; then icon="󰁻"
else                              icon="󰂎"
fi

# Build display text — override icon for AC/charging states
if [ "$ac" = "1" ] && [ "$percent" -ge 100 ]; then
    text="󱘖 ${percent}%"
elif [ "$ac" = "1" ]; then
    text=" ${percent}%"
else
    text="${icon} ${percent}%"
fi

# Build time remaining string
if [ "$ac" = "0" ] && [ "$minutes" -ne 255 ] && [ "$minutes" -gt 0 ]; then
    hours=$((minutes / 60))
    mins=$((minutes % 60))
    time_str="${hours}h ${mins}min remaining"
elif [ "$ac" = "1" ]; then
    time_str="Plugged in"
else
    time_str="Calculating..."
fi

# CSS class for waybar styling
if   [ "$percent" -le 15 ]; then class="critical"
elif [ "$percent" -le 30 ]; then class="warning"
elif [ "$percent" -ge 95 ]; then class="good"
else                              class="normal"
fi

printf '{"text":"%s","tooltip":"Battery: %d%%\n%s","class":"%s"}\n' \
    "$text" "$percent" "$time_str" "$class"

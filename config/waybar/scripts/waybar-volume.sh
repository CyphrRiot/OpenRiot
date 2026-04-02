#!/bin/sh
# OpenRiot - Waybar Volume Status
# OpenBSD: uses sndioctl(1) for volume and mute state
# Output: JSON for waybar custom module

vol=$(sndioctl -n output.level 2>/dev/null \
    | awk '{ printf "%d", $1 * 100 }')

mute=$(sndioctl -n output.mute 2>/dev/null)

[ -z "$vol" ]  && vol=0
[ -z "$mute" ] && mute=0

if [ "$mute" = "1" ]; then
    icon="󰖁"
    class="muted"
elif [ "$vol" -ge 70 ]; then
    icon="󰕾"
    class="high"
elif [ "$vol" -ge 30 ]; then
    icon="󰖀"
    class="medium"
else
    icon="󰕿"
    class="low"
fi

printf '{"text":"%s %d%%","tooltip":"Volume: %d%%","class":"%s"}\n' \
    "$icon" "$vol" "$vol" "$class"

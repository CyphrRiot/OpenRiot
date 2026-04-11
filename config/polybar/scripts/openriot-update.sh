#!/bin/sh
# OpenRiot Update Check for Polybar
# Output: icon only (version shown in polybar tooltip)

if [ "${1:-}" = "--click" ]; then
    alacritty -e sh -c 'curl -fsSL https://openriot.org/setup.sh | sh' &
    exit 0
fi

local_version=$(cat ~/.local/share/openriot/VERSION 2>/dev/null || echo "unknown")
remote_version=$(timeout 10 curl -s https://openriot.org/VERSION 2>/dev/null || echo "unknown")

is_newer() {
    printf '%s\n%s\n' "$1" "$2" | awk 'BEGIN{FS="."} {
        for (i=1; i<=3; i++) { v[NR*10+i] = ($i+0) }
    } END {
        for (i=1; i<=3; i++) {
            if (v[20+i] > v[10+i]) { print "yes"; exit }
            if (v[10+i] > v[20+i]) { print "no"; exit }
        }
        print "no"
    }'
}

if [ "$remote_version" = "unknown" ] || [ "$local_version" = "unknown" ]; then
    printf "?"
elif [ "$(is_newer "$local_version" "$remote_version")" = "yes" ]; then
    printf "󰋻"
else
    printf "󰚇"
fi

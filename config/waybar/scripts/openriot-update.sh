#!/bin/sh
# OpenRiot Update Check for Waybar
# OpenBSD compatible (POSIX sh)

STATE_FILE="$HOME/.cache/openriot/update-state"

# Handle click - run installer
if [ "${1:-}" = "--click" ]; then
    printf '%s\n' "Running installer..."
    exec curl -fsSL https://openriot.org/setup.sh | sh
fi

mkdir -p "$(dirname "$STATE_FILE")"

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
    printf '{"text":"-","tooltip":"Update check unavailable","class":"update-none"}\n'
elif [ "$(is_newer "$local_version" "$remote_version")" = "yes" ]; then
    printf '{"text":"upd","tooltip":"Update %s -> %s available","class":"update-available"}\n' \
        "$local_version" "$remote_version"
else
    printf '{"text":"-","tooltip":"Up to date: %s","class":"update-none"}\n' \
        "$local_version"
fi

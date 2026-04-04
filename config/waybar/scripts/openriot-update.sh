#!/bin/sh
# OpenRiot Update Check for Waybar - Three State System
# OpenBSD compatible (POSIX sh)

STATE_FILE="$HOME/.cache/openriot/update-state"
mkdir -p "$(dirname "$STATE_FILE")"

# Get versions
local_version=$(cat ~/.local/share/openriot/VERSION 2>/dev/null || echo "unknown")
remote_version=$(timeout 10 curl -s https://openriot.org/VERSION 2>/dev/null || echo "unknown")

# Compare semantic versions - returns 0 if remote is newer
is_newer_version() {
    local_ver="$1"
    remote_ver="$2"

    [ "$local_ver" = "unknown" ] && return 1
    [ "$remote_ver" = "unknown" ] && return 1
    [ "$local_ver" = "$remote_ver" ] && return 1

    # Use sort -V style comparison via awk
    newer=$(printf '%s\n%s\n' "$local_ver" "$remote_ver" | awk 'BEGIN{FS="."} {
        for (i=1; i<=3; i++) { v[NR][i] = ($i+0) }
    } END {
        for (i=1; i<=3; i++) {
            if (v[2][i] > v[1][i]) { print "newer"; exit }
            if (v[1][i] > v[2][i]) { print "older"; exit }
        }
        print "equal"
    }')

    [ "$newer" = "newer" ] && return 0
    return 1
}

# Handle click events
if [ "$1" = "--click" ]; then
    if is_newer_version "$local_version" "$remote_version"; then
        echo "$remote_version" > "$STATE_FILE"
        openriot --notify "Launching Upgrade..." "Starting OpenRiot upgrade process..." &
        $HOME/.local/share/openriot/config/bin/openriot-version-check --click --gui 2>/dev/null &
    else
        openriot --notify "OpenRiot Up to Date" "Version $local_version is the latest" &
    fi
    exit 0
fi

# Three-state logic
if [ "$remote_version" = "unknown" ] || [ "$local_version" = "unknown" ]; then
    # Network/file error
    printf '{"text":"-","tooltip":"Update check unavailable","class":"update-none"}\n'
elif is_newer_version "$local_version" "$remote_version"; then
    seen_version=$(cat "$STATE_FILE" 2>/dev/null || echo "")

    if [ "$seen_version" = "$remote_version" ]; then
        printf '{"text":"󰏖","tooltip":"OpenRiot update available (seen)\nCurrent: %s\nAvailable: %s","class":"update-seen"}\n' \
            "$local_version" "$remote_version"
    else
        printf '{"text":"󰚰","tooltip":"OpenRiot update available!\nCurrent: %s\nAvailable: %s","class":"update-available"}\n' \
            "$local_version" "$remote_version"
    fi
else
    rm -f "$STATE_FILE"
    printf '{"text":"-","tooltip":"OpenRiot is up to date\nCurrent: %s","class":"update-none"}\n' \
        "$local_version"
fi

#!/bin/sh
# Usage: firefox-wrapper.sh [url] [--notify "Title" "Message"]
NOTIFY_MSG="${2:-Starting Browser...}"
TITLE="${1:-Firefox}"
if [ "$1" = "--notify" ]; then
    TITLE="$2"
    NOTIFY_MSG="$3"
    shift 3
fi
notify-send -i "$HOME/.local/share/openriot/config/icons/firefox.png" "$TITLE" "$NOTIFY_MSG" &
exec firefox "$@"

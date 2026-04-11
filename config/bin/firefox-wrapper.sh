#!/bin/sh
# Usage: firefox-wrapper.sh [url] [--notify "Title" "Message"]
NOTIFY_MSG="${2:-Starting Browser...}"
TITLE="${1:-Firefox}"
if [ "$1" = "--notify" ]; then
    TITLE="$2"
    NOTIFY_MSG="$3"
    shift 3
fi
notify-send -i firefox "$TITLE" "$NOTIFY_MSG" &
exec firefox "$@"

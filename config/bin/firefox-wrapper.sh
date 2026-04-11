#!/bin/sh
notify-send -i firefox "Firefox" "Starting Browser..." &
exec firefox "$@"

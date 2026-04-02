#!/bin/sh
# OpenRiot - Battery Monitor
# Sends mako notifications at 20% and 10% thresholds

NOTIFIED_20=0
NOTIFIED_10=0

while true; do
    percent=$(apm -l 2>/dev/null || echo 100)
    ac=$(apm -a 2>/dev/null || echo 1)

    if [ "$ac" = "0" ]; then
        if [ "$percent" -le 10 ] && [ "$NOTIFIED_10" = "0" ]; then
            notify-send -u critical "Battery Critical" "${percent}% — plug in now"
            NOTIFIED_10=1
            NOTIFIED_20=1
        elif [ "$percent" -le 20 ] && [ "$NOTIFIED_20" = "0" ]; then
            notify-send -u normal "Battery Low" "${percent}% remaining"
            NOTIFIED_20=1
        fi
    else
        # Reset notifications when charging
        NOTIFIED_20=0
        NOTIFIED_10=0
    fi

    sleep 60
done

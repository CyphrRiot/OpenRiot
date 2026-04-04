#!/bin/sh
# OpenRiot - Screen Recording Toggle
# Starts or stops screen recording via wf-recorder
# Usage: openriot-screenrecord.sh [--start|--stop|--toggle]

VIDEOS_DIR="$HOME/Videos"
mkdir -p "$VIDEOS_DIR"

RECORDING=false
if pgrep -x wf-recorder >/dev/null 2>&1; then
    RECORDING=true
fi

case "${1:-toggle}" in
    --start|-s)
        if [ "$RECORDING" = "true" ]; then
            openriot --notify "Screen Recorder" "Already recording"
            exit 0
        fi
        FILENAME="recording-$(date +%Y%m%d-%H%M%S).mp4"
        wf-recorder -f "$VIDEOS_DIR/$FILENAME" &
        sleep 1
        if pgrep -x wf-recorder >/dev/null 2>&1; then
            openriot --notify "Screen Recorder" "Recording started" --urgency low
        else
            openriot --notify "Screen Recorder" "Failed to start recording" --urgency critical
        fi
        ;;
    --stop|-t)
        if [ "$RECORDING" = "false" ]; then
            openriot --notify "Screen Recorder" "Not currently recording"
            exit 0
        fi
        pkill -INT -x wf-recorder
        sleep 1
        openriot --notify "Screen Recorder" "Recording saved" --urgency low
        ;;
    --toggle|*)
        if [ "$RECORDING" = "true" ]; then
            pkill -INT -x wf-recorder
            sleep 1
            openriot --notify "Screen Recorder" "Recording saved" --urgency low
        else
            FILENAME="recording-$(date +%Y%m%d-%H%M%S).mp4"
            wf-recorder -f "$VIDEOS_DIR/$FILENAME" &
            sleep 1
            if pgrep -x wf-recorder >/dev/null 2>&1; then
                openriot --notify "Screen Recorder" "Recording..." --urgency low
            else
                openriot --notify "Screen Recorder" "Failed to start" --urgency critical
            fi
        fi
        ;;
esac

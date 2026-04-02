#!/bin/sh
# OpenRiot - Waybar Guard Script
# Restarts waybar if it crashes (OpenBSD alternative to systemd timer)

WAYBAR_BIN="waybar"
WAYBAR_PID_FILE="/tmp/waybar.pid"

# Function to start waybar
start_waybar() {
    # Kill any existing waybar processes
    pkill -f "waybar" 2>/dev/null || true

    # Small delay to ensure clean start
    sleep 1

    # Start waybar in background
    $WAYBAR_BIN &
    echo $! > "$WAYBAR_PID_FILE"
}

# Function to check if waybar is running
check_waybar() {
    if [ -f "$WAYBAR_PID_FILE" ]; then
        pid=$(cat "$WAYBAR_PID_FILE")
        # Check if process exists and is waybar
        if [ -n "$pid" ] && ps -p "$pid" > /dev/null 2>&1; then
            return 0  # Running
        fi
    fi
    # Also check if waybar is running by process name
    if pgrep -f "^waybar$" > /dev/null 2>&1; then
        return 0  # Running
    fi
    return 1  # Not running
}

# Main loop
while true; do
    if ! check_waybar; then
        start_waybar
    fi
    sleep 30
done

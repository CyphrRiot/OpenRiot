#!/bin/sh
# detect-hang.sh — Objective typing/input latency detector for OpenBSD X11
# Run inside Alacritty. Prints dots every 100ms. Screams when gap >300ms.

INTERVAL_MS=100
HANG_THRESHOLD_MS=300
LOGFILE="$HOME/.cache/openriot/hang-detect.log"

mkdir -p "$(dirname "$LOGFILE")"
echo "=== Hang detection started: $(date '+%Y-%m-%d %H:%M:%S') ===" > "$LOGFILE"
echo "Interval: ${INTERVAL_MS}ms | Hang threshold: ${HANG_THRESHOLD_MS}ms" >> "$LOGFILE"
echo "Press Ctrl+C to stop. Logging to $LOGFILE"
echo ""

# OpenBSD date doesn't have %N — use perl for microsecond precision
last_us=$(perl -MTime::HiRes=time -e 'printf("%.0f", time * 1000000)')

while true; do
    sleep 0.1

    now_us=$(perl -MTime::HiRes=time -e 'printf("%.0f", time * 1000000)')
    gap_us=$((now_us - last_us))
    gap_ms=$((gap_us / 1000))

    if [ "$gap_ms" -gt "$HANG_THRESHOLD_MS" ]; then
        timestamp=$(date '+%H:%M:%S')
        printf "\n[HANG %s] %dms gap detected at %s\n" "$timestamp" "$gap_ms" "$timestamp"
        printf "[HANG %s] %dms gap\n" "$timestamp" "$gap_ms" >> "$LOGFILE"

        # Snapshot top CPU consumers during hang
        printf "  CPU snapshot:\n" | tee -a "$LOGFILE"
        top -b -n 1 2>/dev/null | head -8 | tee -a "$LOGFILE" | sed 's/^/  /'
        printf "  Processes: " | tee -a "$LOGFILE"
        ps -eo pcpu,comm | sort -rn | head -6 | tee -a "$LOGFILE" | sed 's/^/    /'
        printf "\n" | tee -a "$LOGFILE"

        # Print a marker to visually see the gap size
        printf "GAP:%dms " "$gap_ms"
    else
        printf "."
    fi

    last_us=$now_us
done

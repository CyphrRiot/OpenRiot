#!/bin/sh
# OpenRiot - Waybar System Metrics (CPU + RAM combined)
# Output: JSON for waybar custom module

# CPU usage — OpenBSD: use vmstat, last field is idle%
cpu=$(vmstat 1 1 2>/dev/null | awk 'NR==2 { print 100 - $NF }')
[ -z "$cpu" ] && cpu=0

# RAM usage (from waybar-memory.sh pattern)
total_bytes=$(sysctl -n hw.physmem 2>/dev/null)
[ -z "$total_bytes" ] && total_bytes=0
free_pages=$(vmstat 2>/dev/null | awk 'NR==3 { print $5 }')
page_size=$(sysctl -n hw.pagesize 2>/dev/null || echo 4096)
[ -z "$free_pages" ] && free_pages=0
free_bytes=$(awk -v fp="$free_pages" -v ps="$page_size" 'BEGIN { print fp * ps }')
used_bytes=$(awk -v t="$total_bytes" -v f="$free_bytes" 'BEGIN { print t - f }')
ram_pct=$(awk -v u="$used_bytes" -v t="$total_bytes" 'BEGIN {
    if (t > 0) printf "%d", (u / t) * 100
    else       print 0
}')

# Class based on worst metric
if [ "$cpu" -ge 90 ] || [ "$ram_pct" -ge 90 ]; then
    class="critical"
elif [ "$cpu" -ge 70 ] || [ "$ram_pct" -ge 70 ]; then
    class="warning"
else
    class="normal"
fi

printf '{"text":"  %d%%   %d%%","tooltip":"CPU: %d%% | RAM: %d%%","class":"%s"}\n' \
    "$cpu" "$ram_pct" "$cpu" "$ram_pct" "$class"

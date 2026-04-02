#!/bin/sh
# OpenRiot - Waybar Memory Usage
# OpenBSD: uses sysctl hw.physmem + vmstat for accurate used/total
# Output: JSON for waybar custom module

total_bytes=$(sysctl -n hw.physmem 2>/dev/null)
[ -z "$total_bytes" ] && total_bytes=0

# vmstat columns: procs / memory (avm, fre) / page / disks / traps / cpu
# column 5 on data row = free pages
free_pages=$(vmstat 2>/dev/null | awk 'NR==3 { print $5 }')
page_size=$(sysctl -n hw.pagesize 2>/dev/null || echo 4096)

[ -z "$free_pages" ] && free_pages=0
[ -z "$page_size" ]  && page_size=4096

free_bytes=$(awk -v fp="$free_pages" -v ps="$page_size" 'BEGIN { print fp * ps }')
used_bytes=$(awk -v t="$total_bytes" -v f="$free_bytes" 'BEGIN { print t - f }')

used_gb=$(awk  -v u="$used_bytes"  'BEGIN { printf "%.1f", u / 1073741824 }')
total_gb=$(awk -v t="$total_bytes" 'BEGIN { printf "%.1f", t / 1073741824 }')
percent=$(awk  -v u="$used_bytes"  -v t="$total_bytes" 'BEGIN {
    if (t > 0) printf "%d", (u / t) * 100
    else       print 0
}')

if [ "$percent" -ge 90 ]; then
    class="critical"
elif [ "$percent" -ge 70 ]; then
    class="warning"
else
    class="normal"
fi

printf '{"text":"󰾆 %s/%sGB","tooltip":"Memory: %sGB used of %sGB (%d%%)","class":"%s"}\n' \
    "$used_gb" "$total_gb" "$used_gb" "$total_gb" "$percent" "$class"

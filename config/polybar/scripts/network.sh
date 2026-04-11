#!/bin/sh
# OpenRiot - Polybar Network Status
# OpenBSD compatible — uses ifconfig(8)

# Get wifi interface (exclude loopback, vpn, etc.)
WIFI=$(/sbin/ifconfig 2>/dev/null | awk '
/^[a-z]+[0-9]+:/ { iface=$1; gsub(/:/,"",iface); current=iface }
/ieee80211/ && current != "" { print current; exit }
')

# Get signal strength
get_signal() {
    iface=$1
    /sbin/ifconfig $iface 2>/dev/null | grep -oE '\-[0-9]+dBm' | head -1 | tr -d 'dBm'
}

# Get wifi icon based on signal
get_wifi_icon() {
    if [ -z "$WIFI" ]; then
        printf "󰤯"
        return
    fi
    signal=$(get_signal "$WIFI")
    if [ -z "$signal" ]; then
        printf "󰤯"
        return
    fi
    # Convert dBm to percentage (-100dBm = 0%, -30dBm = 100%)
    percent=$(( (signal + 100) * 100 / 70 ))
    [ $percent -gt 100 ] && percent=100
    [ $percent -lt 0 ] && percent=0
    
    if [ $percent -ge 70 ]; then
        printf "󰤨"
    elif [ $percent -ge 50 ]; then
        printf "󰤥"
    elif [ $percent -ge 30 ]; then
        printf "󰤢"
    elif [ $percent -ge 20 ]; then
        printf "󰤟"
    else
        printf "󰤯"
    fi
}

# Output
if [ -n "$WIFI" ]; then
    get_wifi_icon
else
    printf "󰤯"
fi

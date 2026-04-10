#!/bin/sh
# OpenRiot - Polybar Network Status
# OpenBSD compatible — uses ifconfig(8)
# Output: plain text with Nerd Font icons

get_wifi_interface() {
    ifconfig | awk '/^[a-z]/ { iface=$1 } /ieee80211/ { print iface; exit }'
}

get_eth_interface() {
    ifconfig | awk '/^[a-z]/ { iface=$1 } /status: active/ && iface !~ /lo|vlan|enc|pflog|tun|wg/ { print iface; exit }'
}

get_wifi() {
    iface="$1"
    info=$(ifconfig "$iface" 2>/dev/null)
    ssid=$(printf '%s' "$info" | grep -oE 'nwid [^ ]+' | awk '{print $2}')
    if [ -z "$ssid" ]; then
        printf "󰤟"
        return
    fi
    signal=$(printf '%s' "$info" | grep -oE 'signal [-0-9]+' | awk '{print $2}')
    if [ -n "$signal" ]; then
        percent=$(awk -v s="$signal" 'BEGIN { p = int((s + 100) * 2); if (p > 100) p = 100; if (p < 0) p = 0; print p }')
    else
        percent=50
    fi
    if [ "$percent" -ge 70 ]; then
        icon="󰤨"
    elif [ "$percent" -ge 50 ]; then
        icon="󰤥"
    elif [ "$percent" -ge 30 ]; then
        icon="󰤢"
    elif [ "$percent" -ge 20 ]; then
        icon="󰤟"
    else
        icon="󰤯"
    fi
    printf "%s %s" "$icon" "$ssid"
}

get_ethernet() {
    printf "󰈀"
}

eth=$(get_eth_interface)
wifi=$(get_wifi_interface)

if [ -n "$eth" ]; then
    get_ethernet
elif [ -n "$wifi" ]; then
    get_wifi "$wifi"
else
    printf "󰤯"
fi

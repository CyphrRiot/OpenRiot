#!/bin/sh
# OpenRiot - WiFi Selector
# OpenBSD compatible — uses ifconfig(8) for WiFi management
#
# Scans for available networks and connects using wofi as the selector.
# Requires: wofi, doas

# Find the wireless interface
IFACE=$(ifconfig | awk '/^[a-z]/ { iface=$1 } /ieee80211/ { print iface; exit }' | tr -d ':')

if [ -z "$IFACE" ]; then
    notify-send -t 3000 "WiFi" "No wireless interface found"
    exit 1
fi

# Scan for networks
notify-send -t 2000 "WiFi" "Scanning for networks..."
doas ifconfig "$IFACE" scan >/dev/null 2>&1

# Get list of SSIDs from scan results
NETWORKS=$(ifconfig "$IFACE" scan 2>/dev/null \
    | awk '/nwid/ { gsub(/nwid /, ""); gsub(/ chan.*/, ""); print }' \
    | sort -u)

if [ -z "$NETWORKS" ]; then
    notify-send -t 3000 "WiFi" "No networks found"
    exit 1
fi

# Show network selector via wofi
SELECTED=$(printf '%s\n' "$NETWORKS" | wofi --dmenu --prompt "WiFi Network:" --insensitive)

if [ -z "$SELECTED" ]; then
    exit 0
fi

# Check if network requires a password by looking at security flags
SECURED=$(ifconfig "$IFACE" scan 2>/dev/null \
    | grep -A2 "nwid $SELECTED" \
    | grep -c "privacy" || true)

if [ "$SECURED" -gt 0 ]; then
    # Prompt for password via wofi
    PASSWORD=$(printf '' | wofi --dmenu --prompt "Password for $SELECTED:" --password)

    if [ -z "$PASSWORD" ]; then
        notify-send -t 3000 "WiFi" "No password entered, cancelled"
        exit 1
    fi

    # Connect with password
    doas ifconfig "$IFACE" nwid "$SELECTED" wpakey "$PASSWORD"
else
    # Connect without password
    doas ifconfig "$IFACE" nwid "$SELECTED" -wpakey
fi

# Request DHCP lease
doas dhclient "$IFACE"

# Notify result
if ifconfig "$IFACE" 2>/dev/null | grep -q "inet "; then
    notify-send -t 3000 "WiFi" "Connected to $SELECTED"
else
    notify-send -t 3000 "WiFi" "Failed to connect to $SELECTED"
fi

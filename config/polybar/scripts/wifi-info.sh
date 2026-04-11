#!/bin/sh
# Show current wifi status via notification
IFACE=$(/sbin/ifconfig 2>/dev/null | awk '/^[a-z]+[0-9]+:/ { iface=$1; gsub(/:/,"",iface) } /ieee80211/ { print iface; exit }')
if [ -n "$IFACE" ]; then
    STATUS=$(/sbin/ifconfig $IFACE 2>/dev/null | awk '/status: active/ { print "connected" }')
    if [ "$STATUS" = "connected" ]; then
        notify-send -t 2000 "WiFi" "Connected ($IFACE)"
    else
        notify-send -t 2000 "WiFi" "Interface active but not connected"
    fi
else
    notify-send -t 2000 "WiFi" "No wifi interface found"
fi

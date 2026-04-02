#!/bin/sh
# OpenBSD WireGuard click handler for Waybar
# Toggles wireguard via rcctl

if ifconfig wg0 2>/dev/null | grep -q "inet "; then
    rcctl stop wireguard 2>/dev/null && rcctl disable wireguard 2>/dev/null
else
    rcctl enable wireguard 2>/dev/null && rcctl start wireguard 2>/dev/null
fi

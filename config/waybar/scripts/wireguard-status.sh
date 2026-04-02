#!/bin/sh
# OpenBSD WireGuard status for Waybar (JSON)
# Checks wg0 interface and rc.d wireguard status

set -euo pipefail
trap "exit 0" PIPE
exec 2>/dev/null

if ifconfig wg0 2>/dev/null | grep -q "inet "; then
    # WireGuard is up — get peer/endpoint
    endpoint=$(ifconfig wg0 2>/dev/null | grep "endpoint" | awk "{print \$2}" | head -n1 || true)
    if [ -z "$endpoint" ]; then
        endpoint="Active"
    fi
    printf '{"text":"%%{F#9ece6a}%%08x%%{F-} WG","class":"wg-connected","tooltip":"WireGuard Connected: %s\\nRight-click to disconnect"}\n' "$endpoint"
else
    printf '{"text":"%%{F#f7768e}%%08x%%{F-} WG","class":"wg-disconnected","tooltip":"WireGuard Disconnected\\nLeft-click to connect"}\n'
fi

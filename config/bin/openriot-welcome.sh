#!/bin/sh
# OpenRiot - Welcome screen (shown on first login)

[ -f "$HOME/.openriot-welcomed" ] && exit 0

cat << 'EOF'

      ██████╗ ██████╗ ███████╗███╗  ██╗██████╗ ██╗ ██████╗ ████████╗
     ██╔═══██╗██╔══██╗██╔════╝████╗ ██║██╔══██╗██║██╔═══██╗╚══██╔══╝
     ██║   ██║██████╔╝█████╗  ██╔██╗██║██████╔╝██║██║   ██║   ██║
     ██║   ██║██╔═══╝ ██╔══╝  ██║╚████║██╔══██╗██║██║   ██║   ██║
     ╚██████╔╝██║     ███████╗██║ ╚███║██║  ██║██║╚██████╔╝   ██║
      ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚══╝╚═╝  ╚═╝╚═╝ ╚═════╝    ╚═╝

    Welcome to OpenRiot v$(cat $HOME/.local/share/openriot/VERSION 2>/dev/null || echo "0.8") on OpenBSD 7.9

    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

    Quick Start:
    • Super + D          → Open fuzzel app launcher
    • Super + Enter      → Open terminal (foot)
    • Super + Q          → Close focused window
    • Super + 1-9        → Switch workspaces
    • Super + Shift + Q  → Quit Sway

    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

    Documentation: https://openriot.org
    Issues:         https://github.com/CyphrRiot/OpenRiot/issues

    Press Enter or wait 10s to continue...

EOF

# Use timeout to avoid hanging if keyboard not ready yet
if command -v timeout >/dev/null 2>&1; then
    timeout 10 read _ || true
else
    read _ || true
fi
touch "$HOME/.openriot-welcomed"

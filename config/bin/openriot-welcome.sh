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

    Welcome to OpenRiot v0.4 on OpenBSD 7.9

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

    Press any key to continue...

EOF

read -r key
touch "$HOME/.openriot-welcomed"

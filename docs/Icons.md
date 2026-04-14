# OpenRiot Notification Icons

## Generation Command

```bash
# Font: FiraCode Nerd Font (required)
FONT="$HOME/.local/share/fonts/FiraCode/FiraCodeNerdFont-Regular.ttf"

# Generate white icon (transparent background)
convert -background none -fill white -font "$FONT" -pointsize 32 label:"ICON" -resize 48x48 output.png

# Copy to icons folder and deploy
cp output.png config/icons/
cp output.png ~/.local/share/openriot/config/icons/
```

## Icon Specification

| Icon | Title | Content | Source File |
|------|-------|---------|-------------|
| 󰌅 | Memory | Usage: x% | main.go (--mem-notify) |
| 󰢝 | CPU | Usage: x% | main.go (--cpu-notify) |
| 󰛫 | Settings | Brightness: x% | display/display.go |
| 󰕾 | Settings | Speaker: x% | audio/volume.go |
| 󰕿 | Settings | Speaker: Muted | audio/volume.go |
| 󰭦 | Settings | Mic: x% | audio/volume.go |
| 󰭥 | Settings | Mic: Muted | audio/volume.go |
| 󰅛 | VPN | Starting WireGuard... | wireguard/wireguard.go |
| 󰅛 | VPN | Stopping WireGuard... | wireguard/wireguard.go |
| 󰱓 | VPN | Not configured | wireguard/wireguard.go |
| 󰌵 | Display | Night Light: On | night-light.sh |
| 󰃫 | Display | Night Light: Off | night-light.sh |
| 󰜡 | Applications | Launching [name]... | launcher.sh |
| 󰈹 | Applications | [Firefox title] | firefox-wrapper.sh |
| 󰤨 | WiFi | Connected to [SSID] | wifi-info.sh |
| 󰤯 | WiFi | Not connected | wifi-info.sh |
| 󰤯 | WiFi | No interface found | wifi-info.sh |
| 󰦝 | Crypto | Loading... | crypto-notify.sh |
| 󰦝 | Crypto | [prices] | crypto-notify.sh |
| 󰹑 | Desktop | Workspace [N] | keybindings.conf |
| 󰚇 | Desktop | Current Version v[X.X] | openriot-update.sh |
| 󰋻 | Desktop | Upgrade available | openriot-update.sh |

## Icon List (Unique)

1. 󰌅 - memory.png
2. 󰢝 - cpu.png
3. 󰛫 - settings.png (Brightness)
4. 󰕾 - speaker.png
5. 󰕿 - speaker-muted.png
6. 󰭦 - mic.png
7. 󰭥 - mic-muted.png
8. 󰅛 - vpn.png
9. 󰱓 - vpn-error.png
10. 󰌵 - nightlight-on.png
11. 󰃫 - nightlight-off.png
12. 󰜡 - applications.png
13. 󰈹 - firefox.png
14. 󰤨 - wifi.png
15. 󰤯 - wifi-off.png
16. 󰦝 - crypto.png
17. 󰹑 - workspace.png
18. 󰚇 - no-upgrade.png
19. 󰋻 - upgrade.png
20. 󱥾 - proton-drive.png
21. 󰐻 - transmission-on.png
22. 󱧝 - transmission-off.png

## Status

- [x] Step 1: Find all notifications
- [x] Step 2: Enumerate notifications
- [x] Step 3: Ensure consistent style
- [x] Step 4: Find nerd font symbols
- [x] Step 5: List symbols
- [x] Step 6: Generate .png files
- [x] Step 7: Create icons folder
- [x] Step 8: Add icons to notifications

## Deployment

Icons are deployed to: `~/.local/share/openriot/config/icons/`

To regenerate all icons, run:
```bash
FONT="$HOME/.local/share/fonts/FiraCode/FiraCodeNerdFont-Regular.ttf"
ICON_DIR="config/icons"

for icon in memory cpu settings speaker speaker-muted mic mic-muted vpn vpn-error nightlight-on nightlight-off applications firefox wifi wifi-off crypto workspace no-upgrade upgrade proton-drive transmission-on transmission-off; do
    convert -background none -fill white -font "$FONT" -pointsize 32 label:"SYMBOL" -resize 48x48 "$ICON_DIR/$icon.png"
done
```

## Future Enhancements

1. **Notification Icon Lookup** - Create a `.json` or `.toml` file mapping titles to icons:
   ```
   [notifications]
   "Memory" = "memory.png"
   "CPU" = "cpu.png"
   ```
   Then `openriot --notify "title" "body"` looks up icon automatically.

2. **Replace Shell Scripts** - Consolidate `.sh` files into `openriot --{command}` options to reduce brittleness and simplify maintenance.

3. **Deploy Icons via packages.yaml** - Add `config/icons/*` to install configuration.
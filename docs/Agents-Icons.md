# Icon Verification Guide

## CRITICAL: Check SOURCE CODE, Never Guess

Icons are defined in THREE different places:

| Location | Examples | How to Check |
|----------|----------|--------------|
| `config/polybar/config.ini` | launcher, power, lock | View file directly |
| `source/main.go` | transmission, (volume via polybar.go) | View at command handler |
| Individual source files | crypto.go, macspoof.go, battery.go, nightlight.go, update.go, polybar/polybar.go | grep for icon strings |

## Polybar Icons in config.ini

```ini
[module/launcher]
format = "%{T1} %{T-}"        # ← Launcher icon here

[module/power]
format = ⏻                        # ← Power icon here

[module/lock]
format = 󰌾                       # ← Lock icon here
```

## Icons from Go Code

```go
// main.go --polybar-transmission
if rofi.IsTransmissionRunning() {
    fmt.Print("󰐻")
} else {
    fmt.Print("󱧝")
}

// macspoof.go StealthStatus()
return "󰝴" // Stealth ON
return "󱊨" // Stealth OFF

// crypto.go getCryptoIcon()
return "" // Default crypto icon
```

## Finding Icons

```bash
# Search for icon definitions
grep -r '"󰟠\|󰝴\|󱊨\|󰐻\|󱧝\|⏻\|󰌾\|\|󰚇\|󰋻' source/

# Check polybar config
grep "format = " config/polybar/config.ini
```

## Complete Polybar Icon Reference

| Module | Icon(s) | States |
|--------|---------|--------|
| launcher |  | - |
| workspaces | 󰎤󰎧󰎪󰎭 | - |
| date | (text) | - |
| stealth | 󰝴/󱊨 | ON/OFF |
| network-wifi | 󰤨/󰤯/󱛅 | Connected/No signal/No internet |
| network-eth | 󰈀/󰌙 | Connected/Disconnected |
| wireguard | 󰛳/󰅛/󰱓 | Not configured/Disconnected/Connected |
| volume | 󰕾/󰖀/󰕿/ | High/Medium/Low/Muted |
| battery | 󰁺-󰂂 | Discharging (10-100%) |
| | 󰢜-󰂋 | Charging |
| crypto |  | - |
| night-light | /󰌵 | OFF/ON |
| cpu | 󰡳/󰡵/󰊚/󰡴 | 0-25%/25-50%/50-90%/90%+ |
| memory | 󰢿/󰢼/󰢽/󰢾 | 0-25%/25-50%/50-75%/75-100% |
| proton-drive | 󱥾/󰴋/ | Synced/Syncing/Not configured |
| transmission | 󱧝/󰐻 | Running/Stopped |
| openriot-update | 󰋻/󰚇 | Update available/Up to date |
| weather | varies | Based on conditions |
| power | ⏻ | - |
| lock | 󰌾 | - |

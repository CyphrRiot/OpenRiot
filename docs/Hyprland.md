# Hyprland Migration Proposal

Status: proposal, not implemented.

Audience: OpenRiot maintainers and users evaluating a Wayland desktop.

Scope: replace the X11 stack (Xenodm or startx, i3, polybar, rofi, dunst,
picom, feh, maim, i3lock) with Hyprland on OpenBSD, while keeping the current
X11 session available as a fallback.

## 1. Executive Summary

Hyprland itself now builds and runs on OpenBSD. Upstream merged OpenBSD
support in three batches during October and November 2026 (Hyprland PRs
#16457, #16495, #16839, plus aquamarine #432, hyprwayland-scanner #21 and
hyprland-protocols #24). A working OpenBSD port exists on the ports mailing
list but is not in the ports tree as of the tree snapshot from 2026-11-06.

The migration is therefore not a compositor problem. It is a packaging and
integration problem:

1. Package the Hyprland dependency stack.
2. Port the desktop integration from X11 tools to Wayland tools.
3. Add Wayland backends to the Go code that talks to X11 today.
4. Keep the X11 session selectable until the Wayland session is proven.

The single largest code risk is window layout restore (`source/window`),
because i3 has `append_layout` and Hyprland has no equivalent. The largest
packaging risk is the Hyprland dependency chain, not Hyprland itself.

## 2. Current Architecture

The current desktop is a set of X11 clients coordinated by i3.

| Layer | Current component | Coupling point |
|-------|-------------------|----------------|
| Session | Xenodm or startx, `.xsession` | `config/xsession/openriot-xsession`, `config/xinitrc/openriot-x11` |
| Window manager | i3 | `config/i3/config`, `config/i3/keybindings.conf` |
| Status bar | polybar, three bars (left, center, right) | `source/polybar`, `config/polybar/config.ini.tmpl` |
| Launcher | rofi and rofi-calc | `source/rofi`, `source/settings` |
| Notifications | dunst | `source/notify` |
| Compositor | picom | `config/picom.conf`, i3 autostart |
| Wallpaper | feh | `source/backgrounds` |
| Display control | xrandr | `source/display`, `source/screen`, `source/resolution` |
| Screenshots | maim and xclip | `source/screenshot` |
| Recording | ffmpeg x11grab | `source/screenrec` |
| Lock and idle | i3lock and xautolock | `source/lock`, `.xinitrc` autostart |
| Clipboard history | clipmenu | i3 keybinding |
| Night light | xrandr gamma and xset | `source/nightlight` |
| Terminal | alacritty | `config/alacritty` |

Workspace, title, icon and window switching all read i3 IPC or X11
properties:

- `source/workspace/workspace.go` uses `i3-msg -t get_workspaces`.
- `source/workspaceicons/workspaceicons.go` uses `i3-msg -t get_tree`.
- `source/windowtitle/windowtitle.go` and `source/windowicon/windowicon.go`
  use `i3-msg -t get_tree`.
- `source/window/switch.go` builds a rofi menu from the i3 tree.
- `source/window/snapshot.go` saves and restores i3 layouts.
- `source/commands/helpers.go` reads `_NET_CLIENT_LIST` with `xprop`.

## 3. Target Architecture

Target session:

```
greetd or TTY
  start-hyprland (or Hyprland directly)
    Hyprland
      waybar
      wofi or fuzzel
      fnott or mako
      swaybg (or hyprpaper)
      hypridle
      hyprlock (or swaylock)
      XWayland for legacy apps
```

OpenRiot becomes desktop-aware instead of i3-aware. The Go packages should
select a backend from the session type:

- `XDG_SESSION_TYPE=wayland` and `HYPRLAND_INSTANCE_SIGNATURE` set: Hyprland
  backend.
- Otherwise: existing i3 and X11 backend.

This keeps the migration reversible and lets the installer ship both
sessions.

## 4. Dependency and Packaging Plan

### 4.1 Confirmed available in the local ports tree

Checked on 2026-11-06:

- `wayland`, `wayland-protocols`, `wlroots`
- `sway`, `labwc`, `niri`, `wayfire`
- `waybar`, `wofi`, `fuzzel`
- `grim`, `slurp`, `swaybg`, `swayidle`, `swaylock`
- `wl-clipboard`, `wlr-randr`
- `fnott` (notification daemon)
- `seatd`, `libudev-openbsd`, `libinput-openbsd`
- `xdg-desktop-portal`, `xdg-desktop-portal-gtk`
- `libdisplay-info`, `hwdata`, `udis86`, `glslang`, `muparser`, `re2`
- `libcanberra`, `epoll-shim`, `libei`, `libinotify`

### 4.2 Missing from the local ports tree

These must be ported or supplied by the PoC tarball:

- `hyprland`
- `aquamarine`
- `hyprutils`
- `hyprgraphics`
- `hyprcursor`
- `hyprwayland-scanner`
- `hyprland-protocols`
- `glaze`
- Lua 5.5 (local `lang/lua` currently stops at 5.4)
- Optional: `hyprlock`, `hypridle`, `hyprpaper`, `hyprsunset`, `hyprpicker`,
  `hyprshot`, `wlsunset`, `mako`, `cliphist`

Until `hyprlock`, `hypridle`, `hyprpaper`, `hyprsunset` and `mako` are
packaged, use `swaylock`, `swayidle`, `swaybg`, `fnott` and a small
`wsconsctl` wrapper. These are in ports today and run on Hyprland because
they speak standard Wayland protocols.

### 4.3 Build order

```
lua 5.5
hyprland-protocols
hyprwayland-scanner
hyprutils
hyprgraphics
glaze
hyprcursor
aquamarine
hyprland
```

Then the optional tools.

### 4.4 OpenBSD packaging notes

- aquamarine requires `libseat` and `libudev`. On OpenBSD these map to
  `seatd` and `libudev-openbsd`. Confirm the PoC patches do not assume the
  Linux systemd seat interface.
- Hyprland requires Lua 5.5 at configure time. A Lua 5.5 port is a hard
  prerequisite unless the port carries a patch.
- `glaze` is fetched with FetchContent on Linux. On OpenBSD the upstream
  CMake deliberately skips that branch, so the port must provide glaze.
- Use `NO_SYSTEMD=1` and `NO_UWSM=1`. There is no systemd on OpenBSD.
- XWayland support is required for Thunar, LibreOffice, some Electron apps
  and older GTK applications.
- Keep the port under `wayland/hyprland` and add a `pkg/DESCR` that states
  amd64 only initially. i915 and amdgpu are the tested drivers.

## 5. Migration Phases

### Phase 0: Bootstrap and prove the compositor

Tasks:

- Add the Hyprland stack ports or a local overlay from the ports@ tarball.
- Build and launch Hyprland from a TTY with `seatd` running.
- Confirm DRM/KMS, input, XWayland and a terminal.
- Record `hyprctl version`, `hyprctl monitors -j`, `hyprctl clients -j`.

Exit criteria: Hyprland starts, alacritty runs, XWayland apps start, and the
session can be exited cleanly.

### Phase 1: Port the base configuration

Tasks:

- Create `config/hypr/hyprland.conf`, split into:
  `monitors.conf`, `keybindings.conf`, `windowrules.conf`,
  `workspaces.conf` and `input.conf`.
- Use the ArchRiot layout as the starting point (see section 15).
- Set `$mod = SUPER`, matching i3 `Mod4`.
- Replace ArchRiot application variables with OpenRiot ones.
- Port the OpenRiot dialog rules from `config/i3/config` lines 104 to 125.

Exit criteria: the whole i3 keybinding set has a Hyprland equivalent and
OpenRiot dialogs float and center.

### Phase 2: Session and autostart

Tasks:

- Add `config/hypr/openriot-hyprland-session` that sets
  `XDG_RUNTIME_DIR`, `XCURSOR_*`, `GTK_THEME`, `PATH`, creates
  `~/Screenshots`, then `exec start-hyprland`.
- Add a Wayland desktop file under `config/hypr/` for greetd or any display
  manager that reads `wayland-sessions`.
- Start `waybar`, notification daemon, `hypridle`, `swaybg` and clipboard
  history from `exec-once`.
- Keep the `.xsession` and `.xinitrc` files untouched so the X11 session
  remains selectable.

Exit criteria: a fresh login can choose either session; the Wayland session
starts all services without manual commands.

### Phase 3: Status bar

Tasks:

- Replace polybar with waybar.
- Add `config/waybar/config` and `config/waybar/style.css`.
- Reuse the ArchRiot module split: base modules, custom modules, groups and
  workspaces.
- Add OpenRiot JSON producers (section 8).
- Remove polybar restart throttling. Use `pkill -SIGUSR2 waybar` on config
  reload.

Exit criteria: every polybar module has a waybar equivalent and clicks still
open the same OpenRiot panels.

### Phase 4: Launcher and menus

Tasks:

- Replace rofi with `wofi` or `fuzzel`.
- Add a Wayland theme to `config/wofi/` or `config/fuzzel/`.
- Replace `rofi -show calc` with an OpenRiot calculator command.
- Replace `rofi -show filebrowser` with a Thunar invocation or a small
  `openriot --file-browser` menu.
- Keep `rofi` installed while the X11 session remains supported.

Exit criteria: application launch, settings menu, power menu and window
switcher work without rofi.

### Phase 5: Notifications

Tasks:

- Replace dunst with `fnott` or `mako`.
- Replace `dunstctl` in `source/notify` with the new daemon's control
  command.
- Keep `notify-send` calls unchanged. They are daemon agnostic.
- Port the dunst scaling logic to the daemon's font and width settings.

Exit criteria: every `openriot --notify` path and direct `notify-send` call
produces a visible notification with working actions.

### Phase 6: Lock, idle and power

Tasks:

- Replace i3lock with `hyprlock` if packaged, otherwise `swaylock`.
- Replace xautolock with `hypridle` if packaged, otherwise `swayidle`.
- Lock on idle, dim before lock, suspend with `zzz`, and lock before sleep.
- Remove `xset`, `setxkbmap`, `loginctl` and other X11 or systemd calls.
- Keep the existing `--smart-lock` and `--lock` command interface. Only the
  backend changes.

Exit criteria: lock, unlock, idle dim, idle lock and suspend all work from a
cold boot with no X11 processes.

### Phase 7: Display, wallpaper, screenshots and recording

Tasks:

- Replace xrandr reads with `hyprctl monitors -j` or `wlr-randr`.
- Replace xrandr writes with `hyprctl keyword monitor`.
- Replace feh with `swaybg`.
- Replace maim with `grim`, and `maim -s` with `grim` plus `slurp`.
- Replace xclip with `wl-copy` and `wl-paste`.
- Replace the ffmpeg x11grab path. This needs a separate decision (section
  14).

Exit criteria: mirror, extend, disable output, wallpaper rotation,
full-screen and region screenshots all work on Wayland.

### Phase 8: Go backends

Tasks:

- Introduce a small `desktop` interface and two implementations: `i3` and
  `hyprland`.
- Migrate workspace, title, icon and window switching first.
- Migrate display and wallpaper next.
- Leave layout snapshot disabled on Hyprland until a replacement exists.
- Detect the backend once in `main` and pass it down, instead of having each
  package shell out to `i3-msg`.

Exit criteria: `make test` passes, both sessions pass the test matrix, and
no Hyprland code path invokes `i3-msg`, `xrandr`, `xprop` or `xdotool`.

### Phase 9: Installer and default session

Tasks:

- Add a `desktop.hyprland` group to `install/packages.yaml`.
- Install the Hyprland stack first, then the config files.
- Add a file conflict policy: polybar and waybar configs coexist in
  `~/.config`.
- Add an installer prompt or settings toggle for the default session.
- Keep i3 installed for at least one release cycle.

Exit criteria: a new install can choose Hyprland, an existing install can
migrate, and a failed Hyprland install leaves the X11 session intact.

## 6. Tool Mapping

| Current | Wayland replacement | Notes |
|---------|---------------------|-------|
| i3 | Hyprland | Different config language; layout concepts differ |
| polybar | waybar | Native Hyprland workspace and window modules |
| rofi | wofi or fuzzel | Theme must be rewritten |
| rofi-calc | OpenRiot calculator command | No direct rofi-calc equivalent |
| dunst | fnott or mako | `notify-send` unchanged |
| picom | Hyprland built-in compositor | Remove picom entirely |
| feh | swaybg | Hyprpaper optional |
| xrandr | `hyprctl` or `wlr-randr` | Monitor hotplug needs testing |
| maim | grim | Region capture uses `slurp` |
| xclip | wl-clipboard | `wl-copy`, `wl-paste` |
| clipmenu | cliphist plus wofi | cliphist must be packaged |
| i3lock | hyprlock or swaylock | Keep OpenRiot `--lock` interface |
| xautolock | hypridle or swayidle | Keep OpenRiot `--smart-lock` interface |
| xset gamma | hyprsunset or wlsunset | Both need packaging |
| ffmpeg x11grab | unresolved | Hyprland screencopy path needs work |
| Xresources | GTK settings and env | No xrdb merge |
| setxkbmap, xinput | Hyprland `input` block | Trackpoint setup uses `wsconsctl` |
| i3-msg | `hyprctl` | JSON output is stable enough to parse |

## 7. Configuration Layout

Proposed tree:

```
config/hypr/
  hyprland.conf          # main, generated by the installer
  monitors.conf          # user-editable, survives reinstall
  keybindings.conf       # user-editable, survives reinstall
  windowrules.conf       # user-editable, survives reinstall
  input.conf             # keyboard and touchpad
  openriot-hyprland-session
config/waybar/
  config
  style.css
  Modules
  ModulesCustom
  ModulesWorkspaces
  ModulesGroups
  colors.css
config/wofi/
  config
  style.css
```

Main configuration skeleton:

```conf
$terminal = alacritty
$fileManager = thunar
$openriot_bin = $HOME/.local/share/openriot/install/openriot
$browser = $openriot_bin --browser
$mod = SUPER

source = ~/.config/hypr/monitors.conf
source = ~/.config/hypr/input.conf
source = ~/.config/hypr/windowrules.conf

env = XCURSOR_THEME,Bibata-Modern-Ice
env = XCURSOR_SIZE,24
env = GTK_THEME,Adwaita-dark

exec-once = waybar
exec-once = fnott
exec-once = swaybg -i ~/.local/share/openriot/backgrounds/default.jpg
exec-once = $openriot_bin --volume restore
exec-once = $openriot_bin --night-light-restore
exec-once = $openriot_bin --wireguard-autostart

general {
    gaps_in = 0
    gaps_out = 2
    border_size = 1
    col.active_border = rgba(37a29cff)
    col.inactive_border = rgba(333333cc)
    layout = dwindle
}

decoration {
    rounding = 0
    active_opacity = 1.0
    inactive_opacity = 1.0
    blur {
        enabled = false
    }
    shadow {
        enabled = false
    }
}

animations {
    enabled = false
}

input {
    kb_layout = us
    follow_mouse = 1
    repeat_rate = 40
    repeat_delay = 600
}
```

The zero gaps, zero rounding, no blur and no animations match the current i3
look. Enable them later if the hardware can afford it.

Monitor file:

```conf
monitor = , preferred, auto, 1

# Secondary display to the right
# monitor = HDMI-A-1, 1920x1080@60, 1920x0, 1

# Rotated display
# monitor = DVI-D-1, 1920x1080@60, 0x0, 1, transform, 1
```

## 8. Keybinding Migration

`Mod4` in i3 maps to `SUPER` in Hyprland.

| i3 binding | Hyprland binding |
|------------|------------------|
| `$mod+Return` | `bind = SUPER, RETURN, exec, $terminal` |
| `$mod+Shift+Return` | `bind = SUPER SHIFT, RETURN, exec, $terminal --class floating_center` |
| `$mod+D`, `$mod+space` | `bind = SUPER, D, exec, wofi --show drun` |
| `$mod+Shift+D` | `bind = SUPER SHIFT, D, exec, thunar` |
| `$mod+F` | `bind = SUPER, F, exec, $fileManager` |
| `$mod+B` | `bind = SUPER, B, exec, $browser` |
| `$mod+T` | `bind = SUPER, T, exec, $terminal -e btop` |
| `$mod+O` | `bind = SUPER, O, exec, gnome-text-editor` |
| `$mod+G` | `bind = SUPER, G, exec, Telegram` |
| `$mod+C` | `bind = SUPER, C, exec, $openriot_bin --crush` |
| `$mod+S` | `bind = SUPER, S, exec, $openriot_bin --signal-launch` |
| `$mod+Shift+K` | `bind = SUPER SHIFT, K, exec, $terminal --class rmpc -e rmpc` |
| `$mod+Q` | `bind = SUPER, Q, killactive` |
| `$mod+Z` | `bind = SUPER, Z, togglefloating` |
| `$mod+H` | `bind = SUPER, H, togglesplit` |
| `$mod+Shift+F` | `bind = SUPER SHIFT, F, fullscreen, 0` |
| `$mod+P` | `bind = SUPER, P, pseudo` |
| `$mod+Shift+V` | `bind = SUPER SHIFT, V, exec, cliphist list \| wofi --dmenu \| cliphist decode \| wl-copy` |
| `$mod+Tab` | `bind = SUPER, TAB, exec, $openriot_bin --window-switch` |
| `$mod+Shift+Tab` | `bind = SUPER SHIFT, TAB, exec, hyprctl dispatch cyclenext prev` |
| `$mod+Left/Right/Up/Down` | `bind = SUPER, left, movefocus, l` (and r, u, d) |
| `$mod+1..4` | `bind = SUPER, 1, workspace, 1` (and so on) |
| `$mod+Shift+1..4` | `bind = SUPER SHIFT, 1, movetoworkspace, 1` |
| `$mod+Shift+Left..Down` | `bind = SUPER SHIFT, left, swapwindow, l` |
| `$mod+Ctrl+Left..Down` | `bind = SUPER CTRL, left, movewindow, l` |
| `$mod+equal` | `bind = SUPER, equal, exec, $openriot_bin --calculator` |
| `$mod+bracketright` | `bind = SUPER, bracketright, resizeactive, 20 0` |
| `$mod+bracketleft` | `bind = SUPER, bracketleft, resizeactive, -20 0` |
| `$mod+Shift+bracketright` | `bind = SUPER SHIFT, bracketright, resizeactive, 0 20` |
| `$mod+Shift+bracketleft` | `bind = SUPER SHIFT, bracketleft, resizeactive, 0 -20` |
| `$mod+minus` | `bind = SUPER, minus, togglespecialworkspace, scratch` |
| `$mod+Shift+minus` | `bind = SUPER SHIFT, minus, movetoworkspacesilent, special:scratch` |
| `$mod+button4/5` | `bind = SUPER, mouse_down, workspace, e+1` |
| Volume, brightness keys | Keep OpenRiot commands, use `bindel` |
| `Print` | `bind = , PRINT, exec, $openriot_bin --screenshot` |
| `$mod+E/M/X`, `Shift+X` | Keep existing browser app bindings |
| `$mod+L` | `bind = SUPER, L, exec, $openriot_bin --lock` |
| `$mod+Escape` | `bind = SUPER, ESCAPE, exec, $openriot_bin --power-menu` |
| `$mod+W`, `Shift+W` | Keep wallpaper next and previous |
| `$mod+Shift+space` | `bind = SUPER SHIFT, SPACE, exec, pkill -SIGUSR2 waybar` |
| `$mod+Shift+R` | `bind = SUPER SHIFT, R, exec, hyprctl reload` |
| `$mod+Shift+C` | `bind = SUPER SHIFT, C, exec, hyprctl reload` |
| `$mod+Shift+E` | `bind = SUPER SHIFT, E, exec, hyprctl dispatch exit` |
| `$mod+Shift+H` | `bind = SUPER SHIFT, H, exec, $openriot_bin --help-launch` |

Media keys must be `bindel` or `bindl` so they do not grab.

Window rules for OpenRiot dialogs:

```conf
windowrule {
    name = floating-center
    match:class = ^(floating_center|openriot_wifi|openriot_disk|openriot_migrate|openriot_resolution|zcash-mini)$
    float = yes
    center = yes
    size = 50% 60%
}

windowrule {
    name = openriot-upgrade
    match:class = ^(openriot_upgrade)$
    float = yes
    center = yes
    size = 50% 60%
}

windowrule {
    name = transmission
    match:class = ^(Transmission)$
    float = yes
    center = yes
    size = 800 600
}

windowrule {
    name = control-center
    match:title = ^(OpenRiot Control Center)$
    float = yes
    center = yes
    size = 480 340
}
```

## 9. Waybar Plan

Polybar currently runs as three bars. Waybar runs as one process with
`modules-left`, `modules-center` and `modules-right`. Keep the same visual
split: launcher and workspaces left, clock center, system status right.

Module mapping:

| Polybar module | Waybar module |
|----------------|---------------|
| `launcher` | `custom/openriot-launcher` |
| workspace list | `hyprland/workspaces` |
| `date` | `clock` |
| `volume` | `custom/openriot-volume` or `pulseaudio` |
| `network-wifi`, `network-eth` | `network` |
| `battery` | `battery` |
| `cpu`, `memory` | `cpu`, `memory` |
| `openriot-update` | `custom/openriot-update` |
| `hdmi` | `custom/openriot-hdmi` |
| `weather` | `custom/openriot-weather` |
| `proton-drive`, `transmission`, `nzbget` | `custom/openriot-*` |
| `crypto` | `custom/openriot-crypto` |
| `night-light` | `custom/openriot-nightlight` |
| `screenrec` | `custom/openriot-rec` |
| `wireguard`, `stealth` | `custom/openriot-wireguard`, `custom/openriot-stealth` |
| `power` | `custom/openriot-power` |
| `lock` | `custom/openriot-lock` |

Custom modules should use `return-type: json`, matching ArchRiot's working
pattern (`~/Code/ArchRiot/config/waybar/ModulesCustom` lines 18 to 124):

```json
"custom/openriot-volume": {
    "format": "{}",
    "return-type": "json",
    "interval": 1,
    "exec": "$HOME/.local/share/openriot/install/openriot --waybar-volume",
    "on-click": "$HOME/.local/share/openriot/install/openriot --volume toggle",
    "on-scroll-up": "$HOME/.local/share/openriot/install/openriot --volume inc",
    "on-scroll-down": "$HOME/.local/share/openriot/install/openriot --volume dec",
    "tooltip": true
}
```

New OpenRiot subcommands required:

- `--waybar-volume`
- `--waybar-cpu`
- `--waybar-memory`
- `--waybar-update`
- `--waybar-crypto`
- `--waybar-network`
- `--waybar-battery` (native waybar battery may suffice)
- `--waybar-rec`
- `--waybar-wireguard`

Each must print one JSON object with `text`, `tooltip`, `class` and
optionally `alt`. This isolates formatting in Go and keeps the waybar config
thin.

Waybar workspace module:

```json
"hyprland/workspaces": {
    "active-only": false,
    "all-outputs": true,
    "format": "{icon}",
    "on-click": "activate",
    "on-scroll-up": "hyprctl dispatch workspace e+1",
    "on-scroll-down": "hyprctl dispatch workspace e-1",
    "persistent-workspaces": { "*": 4 },
    "format-icons": {
        "active": "●",
        "default": "◉",
        "empty": "○"
    }
}
```

This replaces the i3 tree parsing in `source/workspaceicons`. The custom
workspace icon and name logic can be dropped unless it carries information
waybar cannot show.

## 10. Go Code Changes

### 10.1 Backend interface

Add a small interface, for example `source/desktop`:

```go
type Backend interface {
    Workspaces() ([]Workspace, error)
    ActiveWindow() (Window, error)
    Clients() ([]Window, error)
    SwitchWorkspace(int) error
    MoveToWorkspace(int) error
    FocusWindow(id string) error
    Monitors() ([]Monitor, error)
    SetWallpaper(path string) error
}
```

Implementations: `desktop/i3.go` and `desktop/hyprland.go`. Existing
packages call the interface. This avoids scattering session checks through
every command.

### 10.2 Per-package migration

| File | Current | Proposed |
|------|---------|----------|
| `source/workspace/workspace.go` | `i3-msg -t get_workspaces` | `hyprctl -j workspaces` |
| `source/workspaceicons/workspaceicons.go` | `i3-msg -t get_tree` | `hyprctl -j workspaces`, `hyprctl -j clients` |
| `source/windowtitle/windowtitle.go` | `i3-msg -t get_tree` | `hyprctl -j activewindow` |
| `source/windowicon/windowicon.go` | `i3-msg -t get_tree` | `hyprctl -j activewindow` |
| `source/window/switch.go` | rofi plus i3 tree | wofi plus `hyprctl -j clients` |
| `source/window/snapshot.go` | i3 `append_layout` | disable on Hyprland, or reimplement with window rules |
| `source/display/hdmi.go` | `i3-msg get_outputs`, xrandr | `hyprctl -j monitors`, `hyprctl keyword monitor` |
| `source/screen/screen.go` | xrandr | `hyprctl -j monitors` |
| `source/resolution/backend.go` | xrandr | `hyprctl monitors`, `wlr-randr` |
| `source/backgrounds/backgrounds.go` | xrandr plus feh | `hyprctl -j monitors` plus swaybg |
| `source/screenshot/screenshot.go` | maim plus xclip | grim plus slurp plus wl-copy |
| `source/screenrec/screenrec.go` | ffmpeg x11grab | undecided, see section 14 |
| `source/lock/lock.go` | i3lock plus xset | hyprlock or swaylock plus hypridle or swayidle |
| `source/notify/notify.go` | dunst plus dunstctl | fnott or mako control command |
| `source/nightlight/nightlight.go` | xrandr gamma | hyprsunset or wlsunset |
| `source/polybar/polybar.go` | polybar config plus i3-msg | new `source/waybar` generator |
| `source/rofi/rofi.go` | rofi | wofi or fuzzel |
| `source/settings/settings.go` | rofi | wofi or fuzzel |
| `source/commands/helpers.go` | xprop, i3-msg exit | `hyprctl -j clients`, `hyprctl dispatch exit` |
| `install/packages.yaml` | i3, polybar, rofi, dunst | new `desktop.hyprland` group |

Notes per file:

- `source/window/snapshot.go` is the hardest. i3 layout restore has no
  Hyprland equivalent. The proposal is to disable the feature on Hyprland
  and tell the user, rather than shipping a half-working restore.
- `source/window/switch.go` currently filters polybar and dunst. The
  Hyprland backend must filter waybar, fnott and other layer-shell clients.
- `source/commands/helpers.go` `_NET_CLIENT_LIST` logic is X11 only. Use
  `hyprctl -j clients` and map the address to a focus command.
- `source/polybar/polybar.go` should not be deleted. It remains the X11
  backend until the X11 session is removed.

## 11. Session Startup

Proposed `openriot-hyprland-session`:

```sh
#!/bin/sh

export XDG_RUNTIME_DIR=/tmp/$USER-runtime
mkdir -p "$XDG_RUNTIME_DIR"
chmod 0700 "$XDG_RUNTIME_DIR"

export PATH="$HOME/.local/bin:$HOME/.local/share/openriot/install:$PATH"
export XCURSOR_THEME=Bibata-Modern-Ice
export XCURSOR_SIZE=24
export GTK_THEME=Adwaita-dark

fc-cache -f 2>/dev/null
mkdir -p "$HOME/Screenshots"

exec start-hyprland
```

OpenBSD specifics:

- `XDG_RUNTIME_DIR` must be 0700. The existing `.xsession` already does this.
- `seatd` must be enabled if Hyprland uses the libseat backend. Test both
  the seatd path and any no-op or direct DRM path the PoC provides.
- There is no `loginctl`. Lock before suspend with `zzz` triggered from the
  lock or idle daemon, not from a systemd unit.
- Do not run `xset` or `setxkbmap`. Keyboard layout belongs in the
  Hyprland `input` block.
- Keep XWayland enabled for Thunar, LibreOffice and older GTK apps.

## 12. OpenBSD Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Hyprland churn during Lua config transition | Config breaks between releases | Pin the tested commit in the port; ship a known config |
| Lua 5.5 not in ports | Build fails | Add a Lua 5.5 port or carry a patch |
| glaze not in ports | Build fails | Port glaze before Hyprland |
| aquamarine seat and udev assumptions | Session cannot start | Test seatd and libudev-openbsd early |
| No `hyprshot`, `hyprpaper`, `hyprsunset`, `hyprlock` | Missing features | Use grim, swaybg, wlsunset, swaylock first |
| Screencast and screen sharing | Broken feature | Research xdg-desktop-portal-hyprland and PipeWire on OpenBSD |
| Recording | No direct replacement | Decide between wf-recorder, Hyprland screencopy helper, or dropping the feature |
| Layout restore | Feature loss | Disable on Wayland and document it |
| X11 apps | Some break under XWayland | Keep the X11 session and test LibreOffice, Thunar, Telegram |
| AMD and Intel only | Unsupported GPUs | Mark amd64, i915 and amdgpu as supported; state the rest as untested |
| Installer regression | Users lose a working desktop | Keep the i3 group and do not remove X11 packages |

## 13. Test Matrix

Run on both sessions after each phase:

1. Cold login from greetd or TTY.
2. Every keybinding in section 8.
3. Workspace switch, move and swap.
4. Floating and centered OpenRiot dialogs.
5. Bar modules: click, scroll and tooltip.
6. Notifications with actions.
7. Full-screen screenshot, region screenshot, clipboard paste.
8. Lock, unlock, idle dim, idle lock, suspend and resume.
9. HDMI hotplug, mirror, extend, disable laptop panel.
10. Wallpaper next and previous, including multi-monitor.
11. XWayland apps: Thunar, LibreOffice, Telegram, alacritty.
12. `make test` and the full OpenRiot command sweep.
13. Crypto and update bar modules.
14. Session exit and restart.

## 14. Open Decisions

1. Hard switch or dual session? Recommendation: dual session for at least
   one release cycle.
2. Keep a custom bar or adopt waybar as the only bar? Recommendation:
   waybar, because it has native Hyprland modules and is already in ports.
3. `hyprlock` and `hypridle`, or `swaylock` and `swayidle` first?
   Recommendation: sway tools first, Hypr tools once packaged.
4. `mako`, `fnott`, or keep dunst? Recommendation: fnott, already in ports.
5. What replaces layout restore? Options: drop it, recreate via window rules,
   or add a small per-workspace launcher list. Recommendation: drop it for
   the first Wayland release.
6. What replaces screen recording? Options: wf-recorder if it works under
   Hyprland, a Hyprland screencopy helper, or defer the feature.
   Recommendation: defer until the compositor port is stable.
7. What replaces `rofi-calc`? Recommendation: add `openriot --calculator`
   and render it with wofi or a terminal UI.
8. Does the installer install the Hyprland stack by default? Recommendation:
   no, make it an explicit desktop choice until the port is in the tree.

## 15. Reference Material

Working Linux configuration and integration examples:

- `~/Code/ArchRiot/config/hypr/hyprland.conf`
  - application variables: lines 33 to 47
  - autostart: lines 49 to 66
  - window rules: lines 70 to 786
  - input: lines 788 to 813
  - appearance: lines 815 to 899
  - keybindings: lines 901 to 1049
  - theme and performance overrides: lines 1054 to 1140
- `~/Code/ArchRiot/config/hypr/keybindings.conf`
- `~/Code/ArchRiot/config/hypr/monitors.conf`
- `~/Code/ArchRiot/config/hypr/windowrules.conf`
- `~/Code/ArchRiot/config/hypr/hypridle.conf`
- `~/Code/ArchRiot/config/hypr/hyprlock.conf`
- `~/Code/ArchRiot/config/waybar/config`
- `~/Code/ArchRiot/config/waybar/Modules`
- `~/Code/ArchRiot/config/waybar/ModulesCustom`
- `~/Code/ArchRiot/config/waybar/ModulesWorkspaces`
- `~/Code/ArchRiot/config/waybar/ModulesGroups`
- `~/Code/ArchRiot/config/waybar/style.css`
- `~/Code/ArchRiot/source/session/waybar.go`
- `~/Code/ArchRiot/source/waybar/`

Existing OpenRiot material to port:

- `config/i3/config`
- `config/i3/keybindings.conf`
- `config/polybar/config.ini.tmpl`
- `config/rofi/config.rasi`
- `config/dunst/dunstrc`
- `config/picom.conf`
- `config/xsession/openriot-xsession`
- `config/xinitrc/openriot-x11`

Upstream references:

- Hyprland OpenBSD PRs: #16457, #16495, #16839
- aquamarine OpenBSD PR: #432
- hyprwayland-scanner OpenBSD PR: #21
- hyprland-protocols OpenBSD PR: #24
- Hyprland wiki: configuration, monitors, keybinds, window rules,
  workspaces, idle, lock, xwayland

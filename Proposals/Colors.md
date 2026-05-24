# OpenRiot Color Derivation Plan

Single source of truth: `config/colors.toml`. Runtime consumers read it via
`source/theme/colors.go`. Static config consumers must be made dynamic.

---

## Canonical Palette

| Group | Key | Default | Usage |
|-------|-----|---------|-------|
| base | bg | #1a1b26 | Deep background |
| base | bg2 | #24283b | Elevated background |
| base | fg | #c0caf5 | Primary text |
| base | fg2 | #a3acc9 | Secondary text |
| base | fg3 | #565f89 | Tertiary/dim text |
| base | dim | #3b4261 | Borders, dividers |
| accent | name | bondi-green | Theme name |
| accent | fg | #9ECE6A | Primary accent |
| accent | fg-light | #8BB85A | Highlight accent |
| accent | bg | #2B3A1A | Accent background |
| semantic | error | #F7768E | Errors, critical |
| semantic | warning | #E0AF68 | Warnings |
| semantic | success | #04B575 | Success states |
| semantic | info | #7AA2F7 | Info, links |
| semantic | cyan | #0DB9D7 | Cyan accent |
| compat | green | #9ECE6A | Legacy green alias |
| compat | violet | #7D56F4 | Legacy violet alias |
| compat | blue | #7AA2F7 | Legacy blue alias |
| compat | dim-gray | #565F89 | Legacy dim alias |
| compat | white | #FAFAFA | Legacy white alias |
| extended | teal | #7d8df7 | Polybar/teal modules |
| extended | sky | #7DCFFF | Info/sky modules |
| extended | electric | #89B4FA | Electric blue accent |
| extended | purple | #8B7CF6 | Purple accent |
| extended | violet | #997DE1 | Violet accent |
| extended | orange | #FF9E64 | Orange accent |
| extended | cyan-dim | #5a95b8 | Dim cyan |
| extended | launcher-fg | #9EC8F9 | Launcher icon color |
| extended | sep-bg | #13141C | Separator background |
| extended | bg-dark | #06070B | Polybar / bar background |
| extended | bg-mid | #0E0F17 | Elevated dark bg |
| extended | fg-bright | #E0E6F5 | Bright foreground |
| extended | muted | #414868 | Muted elements |
| extended | sec-yellow | #F4D03F | WiFi WPA color |
| extended | sec-orange | #FF8844 | WiFi WPA2 color |
| extended | alpha-bg | #B306070B | Semi-transparent bar bg |

---

## Component Status

| Component | File | Status | Derivation Method |
|-----------|------|--------|-------------------|
| Go TUIs | `source/theme/colors.go` | DONE | Runtime TOML parse |
| polybar | `config/polybar/config.ini.tmpl` | DONE | `openriot --polybar-setup` renders template + screen scaling |
| dunst | `config/dunst/dunstrc` | DONE | `openriot --dunst-setup` copies + scales for hi-DPI |
| btop | `config/btop/themes/current.theme` | DONE | Manual sync, needs template |
| alacritty | `config/alacritty/alacritty.toml` | DONE | Manual sync, needs template |
| rofi | `config/rofi/simple-tokyonight.rasi.tmpl` | DONE | `openriot --rofi-setup` renders template |
| rofi apps | `config/rofi/apps.txt` | N/A | No colors |
| helix | `config/helix/themes/openriot.toml.tmpl` | DONE | `openriot --helix-setup` renders template |
| GTK 3 | `config/gtk-3.0/gtk.css` | TBD | Installer template |
| GTK 4 | `config/gtk-4.0/gtk.css` | TBD | Installer template |
| Xresources | `config/Xresources` | TBD | Installer template |
| i3 | `config/i3/config` | TBD | Installer template |
| fish | `config/fish/config.fish` | DONE | Verified: no hardcoded colors |
| fastfetch | `config/fastfetch/config.jsonc` | DONE | Verified: no hardcoded colors |

---

## Derivation Architecture

`colors.toml` is the single source of truth. It lives at
`~/.local/share/openriot/config/colors.toml` — part of the git checkout,
not generated at install time.

**Template rendering** (e.g. `--polybar-setup`, `--rofi-setup`):
1. Reads `*.tmpl` from `~/.local/share/openriot/config/{component}/`
2. Reads `colors.toml` from `~/.local/share/openriot/config/colors.toml`
3. Renders output to `~/.config/{component}/`

**Static file copy** (`CopyConfigs()` in `source/installer/configs.go`):
- Copies non-template files from deploy dir to `~/.config/`
- Does NOT copy `.tmpl` files — those are rendered by setup commands
- Must NEVER target `~/.local/share/openriot/...` — self-copies truncate
  the source to 0 bytes (`O_TRUNC` in `fsutil.CopyFile`), producing
  blank template values

**In-place assets** (NOT deployed by `CopyConfigs()`):
- `config/icons/` — accessed via `paths.IconDir()` at runtime
- `config/window/icons.toml` — accessed via `paths.OpenRiotDir()`
- `docs/` — accessed via `paths.OpenRiotDir()` for release notes

---

## Available Template Keys

All keys returned by `installer.renderColorsMap()`:

**Base**: `BaseBG`, `BaseBG2`, `BaseFG`, `BaseFG2`, `BaseFG3`, `BaseDim`  
**Accent**: `AccentName`, `AccentFG`, `AccentFGLight`, `AccentBG`  
**Semantic**: `SemanticError`, `SemanticWarning`, `SemanticSuccess`, `SemanticInfo`, `SemanticCyan`  
**Compat**: `CompatGreen`, `CompatViolet`, `CompatBlue`, `CompatDimGray`, `CompatWhite`  
**Extended**: `ExtendedTeal`, `ExtendedSky`, `ExtendedElectric`, `ExtendedPurple`, `ExtendedViolet`, `ExtendedOrange`, `ExtendedCyanDim`, `ExtendedLauncherFG`, `ExtendedSepBG`, `ExtendedBGDark`, `ExtendedBGMid`, `ExtendedFGBright`, `ExtendedMuted`, `ExtendedSecYellow`, `ExtendedSecOrange`, `ExtendedAlphaBG`

---

## Per-Component Mapping

### polybar (`config/polybar/config.ini.tmpl`)

| Config Key | Template Key | colors.toml Source |
|------------|--------------|-------------------|
| background | alpha-bg | extended.alpha-bg (derived: 70% opacity over bg-dark) |
| bg | bg | extended.bg-dark |
| bg2 | bg2 | extended.bg-mid |
| fg | fg | extended.fg-bright |
| fg2 | fg2 | base.fg2 |
| fg3 | fg3 | base.fg3 |
| red | red | semantic.error |
| orange | orange | extended.orange |
| yellow | yellow | semantic.warning |
| green | green | accent.fg |
| cyan | cyan | semantic.cyan |
| blue | blue | semantic.info |
| sky | sky | extended.sky |
| electric | electric | extended.electric |
| purple | purple | extended.purple |
| violet | violet | extended.violet |
| teal | teal | extended.teal |
| cyan-dim | cyan-dim | extended.cyan-dim |
| muted | muted | extended.muted |
| sep (content-foreground) | — | extended.sep-bg |
| launcher (format-foreground) | — | extended.launcher-fg |

### dunst (`config/dunst/dunstrc`)

| Config Key | colors.toml Source |
|------------|-------------------|
| frame_color | base.dim |
| foreground | base.fg |
| background | base.bg |
| highlight | accent.fg |
| separator_color | base.dim |

### rofi (`config/rofi/simple-tokyonight.rasi.tmpl`)

| Config Key | Template Key | colors.toml Source |
|------------|--------------|-------------------|
| bg0 | `{{.BaseBG}}` | base.bg |
| bg1 | `{{.BaseBG}}` | base.bg |
| bg2 | `{{.BaseBG2}}` | base.bg2 |
| bg3 | `{{.ExtendedMuted}}` | extended.muted |
| fg0 | `{{.BaseFG}}` | base.fg |
| fg1 | `{{.BaseFG2}}` | base.fg2 |
| fg2 | `{{.BaseFG3}}` | base.fg3 |
| red | `{{.SemanticError}}` | semantic.error |
| green | `{{.AccentFG}}` | accent.fg |
| yellow | `{{.SemanticWarning}}` | semantic.warning |
| blue | `{{.SemanticInfo}}` | semantic.info |
| magenta | `{{.ExtendedViolet}}` | extended.violet |
| cyan | `{{.SemanticCyan}}` | semantic.cyan |

### helix (`config/helix/themes/openriot.toml`)

| Config Key | colors.toml Source |
|------------|-------------------|
| background | base.bg |
| foreground | base.fg |
| comment | base.fg3 |
| keyword | semantic.info |
| string | semantic.success |
| error | semantic.error |
| warning | semantic.warning |
| info | semantic.cyan |
| ui.selection | accent.bg |
| ui.cursor | semantic.info |
| ui.background | base.bg |

### GTK 3/4 (`config/gtk-3.0/gtk.css`, `config/gtk-4.0/gtk.css`)

| Config Key | colors.toml Source |
|------------|-------------------|
| @bg_color | base.bg |
| @fg_color | base.fg |
| @selected_bg | accent.bg |
| @selected_fg | accent.fg |
| @border | base.dim |
| @error | semantic.error |
| @warning | semantic.warning |
| @success | semantic.success |

### Xresources (`config/Xresources`)

| Config Key | colors.toml Source |
|------------|-------------------|
| *background | base.bg |
| *foreground | base.fg |
| *cursorColor | semantic.info |
| *.color0 | base.bg |
| *.color1 | semantic.error |
| *.color2 | compat.green |
| *.color3 | semantic.warning |
| *.color4 | semantic.info |
| *.color5 | compat.violet |
| *.color6 | semantic.cyan |
| *.color7 | base.fg |
| *.color8 | base.fg3 |
| *.color9 | semantic.error |
| *.color10 | compat.green |
| *.color11 | semantic.warning |
| *.color12 | semantic.info |
| *.color13 | compat.violet |
| *.color14 | semantic.cyan |
| *.color15 | compat.white |

### i3 (`config/i3/config`)

| Config Key | colors.toml Source |
|------------|-------------------|
| client.focused border | accent.fg |
| client.focused background | accent.fg |
| client.focused text | base.bg |
| client.focused_inactive border | base.bg2 |
| client.focused_inactive background | base.bg2 |
| client.focused_inactive text | base.fg2 |
| client.unfocused border | base.bg |
| client.unfocused background | base.bg |
| client.unfocused text | base.fg3 |
| bar.colors.background | base.bg |
| bar.colors.statusline | base.fg |
| bar.colors.separator | base.dim |
| bar.colors.focused_workspace border | accent.fg |
| bar.colors.focused_workspace background | accent.fg |
| bar.colors.focused_workspace text | base.bg |

---

## Implementation Order

Phase 1 — Template Engine **(COMPLETE)**  
Phase 2 — Polybar **(COMPLETE)**  
Phase 3 — Dunst **(COMPLETE — runtime scale, not template)**  
Phase 4 — Rofi **(COMPLETE)**  
Phase 5 — Helix **(COMPLETE)**  

TBD (future work):

Phase 6 — GTK 3/4  
Phase 7 — Xresources  
Phase 8 — i3  
Phase 9 — Backfill (btop, alacritty)

---

## Migration Notes

Already-synced components (btop, alacritty) should be converted last to
avoid regressing working configs. The template engine must be solid before
touching anything runtime-critical like i3 or polybar.

All `.tmpl` files must preserve existing structure — only hex values change.
No functional changes to keybindings, layouts, or behavior.

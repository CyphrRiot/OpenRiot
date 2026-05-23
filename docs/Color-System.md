# OpenRiot Color System Architecture

> _"A desktop should not be a tattoo. It should be a shirt you can change
> without surgery."_ — The OpenRiot Crew

---

## 1. Philosophy

OpenRiot v7.9.5 unified the active UI under Bondi Green (`#9ECE6A`). That
was the right decision for the default. But defaults are not prisons. Some
days you want green. Some days you want the calm of blue. Some days you
want the old purple back. The user should decide — and the system should
remember.

**The rule:** All active UI accents are derived from a single primary
color variable. Change the variable, change the desktop. Every border,
every prompt arrow, every notification frame, every table header, every
polybar module icon — all follow the same primary. The settings menu
keeps its violet border (meta-configuration deserves its own color).
Everything else obeys the wheel.

---

## 2. Color Palettes

Three palettes. Each is a complete color language — not just one hex
code, but a family of tints, shades, and functional mappings.

### 2.1 Green — "Bondi" (Default)

| Role | Hex | Notes |
|------|-----|-------|
| Primary | `#9ECE6A` | Active borders, focused tabs, prompt arrows |
| Primary dim | `#8BB85A` | Secondary accents, Git branches, hover |
| Primary dark | `#2B3A1A` | Selected backgrounds, subtle fills |
| Selection bg | `#9ECE6A` | Text selection (white text on green) |
| Graph end | `#9ECE6A` | btop CPU end, free end, download end |

### 2.2 Blue — "Pacific"

| Role | Hex | Notes |
|------|-----|-------|
| Primary | `#7AA2F7` | i3 focused border, polybar accent modules |
| Primary dim | `#5A8AE0` | Secondary accents, Git branches |
| Primary dark | `#1A2B4A` | Selected backgrounds |
| Selection bg | `#7AA2F7` | Text selection (white text on blue) |
| Graph end | `#7AA2F7` | btop CPU end, free end, download end |

### 2.3 Purple — "CypherRiot" (Legacy)

| Role | Hex | Notes |
|------|-----|-------|
| Primary | `#BB9AF7` | Original accent, keeps the soul alive |
| Primary dim | `#9A7ECC` | Secondary accents, Git branches |
| Primary dark | `#2A1A4A` | Selected backgrounds |
| Selection bg | `#BB9AF7` | Text selection (white text on purple) |
| Graph end | `#BB9AF7` | btop CPU end, free end, download end |

---

## 3. Architecture

### 3.1 State File

`~/.config/openriot/color-theme` stores the current palette name. One
line: `green`, `blue`, or `purple`. Read on boot. Written on change.

If the file does not exist, default to `green`. This preserves the
v7.9.5 behavior for existing users and new installs.

### 3.2 Theme Command

```
openriot --theme-cycle          # Cycles green -> blue -> purple -> green
openriot --theme-set green      # Explicit set
openriot --theme-set blue
openriot --theme-set purple
openriot --theme-show           # Prints current theme name
openriot --theme-apply          # Re-applies current theme (for init)
```

`--theme-cycle` is what the polybar icon calls. `--theme-apply` is what
`~/.xinitrc` or i3 autostart calls on login.

### 3.3 Application Script

The `openriot --theme-apply` command reads `~/.config/openriot/color-theme`
and writes color-ized versions of all config files. It does this by:

1. Reading a set of template files from
   `~/.local/share/openriot/config/templates/`
2. Substituting `{{PRIMARY}}`, `{{PRIMARY_DIM}}`, `{{PRIMARY_DARK}}`
   using Go's `text/template`
3. Writing the output to the live config paths (`~/.config/*`)
4. Signaling running apps to reload (polybar, i3, dunst)

Templates live alongside static configs in the repo under
`config/templates/`. Files that do NOT vary by color (e.g., keybindings,
window rules) remain static and are copied as-is.

### 3.4 Files That Must Be Templated

| File | Template Path | Live Path | Variables Used |
|------|--------------|-----------|---------------|
| i3 config | `config/i3/config.tmpl` | `~/.config/i3/config` | `{{PRIMARY}}` |
| polybar config | `config/polybar/config.ini.tmpl` | `~/.config/polybar/config.ini` | `{{PRIMARY}}`, `{{PRIMARY_DIM}}`, `{{PRIMARY_DARK}}` |
| dunstrc | `config/dunst/dunstrc.tmpl` | `~/.config/dunst/dunstrc` | `{{PRIMARY}}`, `{{PRIMARY_DIM}}` |
| rofi theme | `config/rofi/simple-tokyonight.rasi.tmpl` | `~/.config/rofi/simple-tokyonight.rasi` | `{{PRIMARY}}`, `{{PRIMARY_DIM}}` |
| fish config | `config/fish/config.fish.tmpl` | `~/.config/fish/config.fish` | `{{PRIMARY}}`, `{{PRIMARY_DIM}}` |
| btop theme | `config/btop/themes/current.theme.tmpl` | `~/.config/btop/themes/current.theme` | `{{PRIMARY}}`, `{{PRIMARY_DIM}}`, `{{PRIMARY_DARK}}` |
| fastfetch | `config/fastfetch/config.jsonc.tmpl` | `~/.config/fastfetch/config.jsonc` | `{{PRIMARY}}` (keyColor values) |
| helix theme | `config/helix/themes/openriot.toml.tmpl` | `~/.config/helix/themes/openriot.toml` | `{{PRIMARY}}`, `{{PRIMARY_DIM}}` |
| rmpc theme | `config/rmpc/themes/neo-tokyo.ron.tmpl` | `~/.config/rmpc/themes/neo-tokyo.ron` | `{{PRIMARY}}`, `{{PRIMARY_DIM}}` |
| GTK3 CSS | `config/gtk-3.0/gtk.css.tmpl` | `~/.config/gtk-3.0/gtk.css` | `{{PRIMARY}}`, `{{PRIMARY_DIM}}` |
| website CSS | `assets/css/style.scss.tmpl` | N/A (build-time) | `{{PRIMARY}}`, `{{PRIMARY_DIM}}`, `{{PRIMARY_DARK}}` |

### 3.5 Files That Stay Static (No Color Variation)

| File | Reason |
|------|--------|
| `config/alacritty/alacritty.toml` | Terminal has its own ANSI palette. We do not recolor the terminal |
| `config/lf/lfrc` | File manager uses LS_COLORS, independent |
| `config/helix/config.toml` | Editor settings, not colors |
| `config/i3/keybindings.conf` | Keybindings do not contain colors |
| `config/dconf/*` | dconf is binary; theme handled by app |

### 3.6 Go Source Files With Hardcoded Colors

Some Go source files embed color values for on-the-fly UI generation:

| File | Line | Current | Action |
|------|------|---------|--------|
| `source/resolution/view.go` | ~150 | `cursorStyle` uses `#9ECE6A` | Read `~/.config/openriot/color-theme` at init, parse palette, use `primary` |
| `source/rofi/rofi.go` | ~64 | `themeStr` hardcodes `#9ECE6A` | Same — read state file, substitute border color |
| `source/window/switch.go` | ~106 | `themeStr` hardcodes `#9ECE6A` | Same |
| `source/settings/settings.go` | ~40 | `#997de1` (violet) | **Keep violet intentionally** — settings is meta-config |

The Go code should use a new `openriot/theme` package that:

- Reads `~/.config/openriot/color-theme`
- Returns a `Palette` struct with `Primary`, `PrimaryDim`, `PrimaryDark`
- Caches the palette in memory (file is tiny, reads are cheap)
- Exposes `GetPalette() (*Palette, error)`

---

## 4. Polybar Color Wheel Module

### 4.1 Placement

Right before the `settings` module in `modules-right`:

```
modules-right = ... battery sep theme-wheel settings power lock
```

### 4.2 Module Definition

```ini
[module/theme-wheel]
type = custom/text
format = "%{T1}%{T-}%{O2}"
format-foreground = ${colors.green}

format-padding = 0
click-left = $HOME/.local/share/openriot/install/openriot --theme-cycle
```

Note: The icon `` (nf-fae-paint_brush) renders the color wheel. The
foreground color of the module itself shows the *current* active color:
- Green when theme is green
- Blue when theme is blue
- Purple when theme is purple

This requires polybar config to be regenerated after theme change, since
the `format-foreground` is static in the ini file. The `--theme-apply`
command rewrites `config.ini` with the correct color values and then runs
`polybar-msg cmd restart`.

### 4.3 Visual Feedback on Click

When clicked:

1. `openriot --theme-cycle` reads current theme from state file
2. Computes next theme (green -> blue -> purple -> green)
3. Writes new theme name to state file
4. Calls `openriot --theme-apply` to rewrite all configs
5. Sends dunst notification: "Theme Switched / Now using: Pacific Blue"
6. Restarts polybar, reloads dunst, sends i3 `reload` signal

---

## 5. Persistence

### 5.1 State File

`~/.config/openriot/color-theme`

- Plain text, one word
- Created on first theme cycle if missing
- Read by `theme` package on every binary invocation
- NOT touched by config deployment (`preserve_if_exists` should protect
  it, or it lives outside the copied config tree)

### 5.2 Across Reboots

The i3 autostart (`config/i3/config` line 42 area) already runs
`$openriot_bin --polybar-setup`. Add after it:

```
exec --no-startup-id $openriot_bin --theme-apply
```

This ensures the saved theme is applied before polybar starts.

---

## 6. Implementation Phases

### Phase 1: Theme Package (Go)

- Create `source/theme/theme.go`
- Define `Palette` struct and `GetPalette()` function
- Unit tests for file reading and palette mapping

### Phase 2: Template Files

- Convert all color-bearing config files to `.tmpl` templates
- Keep static copies for initial install (the installer copies templates
  *as* templates, not rendered output)
- Add `config/templates/` directory in repo

### Phase 3: Theme Commands

- Add `--theme-cycle`, `--theme-set`, `--theme-show`, `--theme-apply` to
  `source/commands/commands.go`
- The apply command renders all templates using `text/template`
- Cycle command updates state file then calls apply

### Phase 4: Polybar Module

- Add `[module/theme-wheel]` to `config/polybar/config.ini.tmpl`
- Insert `theme-wheel` into `modules-right`

### Phase 5: Go Source Updates

- Update `source/resolution/view.go` to use `theme.GetPalette()`
- Update `source/rofi/rofi.go` to use `theme.GetPalette()`
- Update `source/window/switch.go` to use `theme.GetPalette()`
- Settings menu **stays violet** (intentional exception)

### Phase 6: README Badges

- README badges are shields.io URLs with hex colors. These are static
  markdown and cannot be dynamically changed per-user.
- **Decision:** README badges stay green (the default). The website CSS
  is user-facing and should template. The README is project-facing and
  represents the canonical default.

---

## 7. Template Syntax Example

### Before (`config/dunst/dunstrc`):

```ini
frame_color = "#9ECE6A"
```

### After (`config/dunst/dunstrc.tmpl`):

```ini
frame_color = "{{PRIMARY}}"
```

### Render function (Go pseudo-code):

```go
func renderTemplate(src, dest string, pal *theme.Palette) error {
    tmpl, err := template.ParseFiles(src)
    if err != nil { return err }

    f, err := os.Create(dest)
    if err != nil { return err }
    defer f.Close()

    return tmpl.Execute(f, pal)
}
```

---

## 8. Migration Path

Existing v7.9.5 users have configs with hardcoded `#9ECE6A`. On first run
of `--theme-cycle` or `--theme-set`:

1. State file does not exist -> create with `green`
2. Configs are already green (they match the default)
3. User clicks wheel -> state file updates to `blue`
4. `--theme-apply` rewrites all live configs with `#7AA2F7`

No migration script needed. The system is self-healing.

---

## 9. Open Questions

1. **btop**: btop does not support live config reload. A theme change
   requires restarting btop. This is acceptable — btop is usually run
   interactively, not left running.

2. **alacritty**: The terminal's cursor color is set in alacritty.toml.
   Should we template this too? **Tentative no** — the terminal ANSI
   palette should stay consistent regardless of desktop accent color.
   The cursor is a functional indicator, not a decorative accent.

3. **helix**: The helix theme `openriot.toml` is loaded at editor start.
   Like btop, helix must be restarted to pick up theme changes. This is
   acceptable for the same reason.

4. **fastfetch**: fastfetch reads its config at startup. It displays in
   the terminal on fish shell start. A theme change does not retroactively
   recolor the already-printed fastfetch output. This is acceptable — the
   next new terminal will show the new color.

5. **Website**: `assets/css/style.scss` is Jekyll-generated static CSS.
   It cannot read a user's local `~/.config/openriot/color-theme`.
   **Decision:** The website stays green. It is the project's public
   face. The color wheel is a desktop personalization feature.

---

## 10. Summary

The color wheel is a single polybar icon that cycles through three
carefully designed palettes. One click changes your entire desktop's
accent color. The change persists across reboots. The implementation is
a state file, a set of Go template files, and a small theme package.

The settings menu stays violet. Everything else follows the wheel.

*[← Back to README](https://OpenRiot.org)*

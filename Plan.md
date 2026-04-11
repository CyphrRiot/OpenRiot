# OpenRiot Config Refactor Plan

## Goal
Clean up the repository structure and fix config deployment so apps find configs in standard locations.

---

## Current Problem

The installer copies configs to `~/.local/share/openriot/config/` but most apps expect configs in `~/.config/`:

```
Current: ~/.local/share/openriot/config/
Expected: ~/.config/
```

This means:
- Alacritty, i3, polybar, rofi, fish, helix, etc. are NOT reading our configs
- Only configs with explicit `--config` flags or special handling work

---

## Proposed Structure

```
repo/
├── backgrounds/          → ~/.local/share/openriot/backgrounds/
├── bin/                  → ~/.local/share/openriot/bin/  (in PATH)
├── share/
│   ├── applications/    → ~/.local/share/applications/
│   └── fonts/           → ~/.local/share/fonts/
└── config/               → ~/.config/
    ├── i3/
    ├── polybar/
    ├── rofi/
    ├── alacritty/
    ├── fish/
    ├── helix/
    ├── btop/
    ├── nvim/
    ├── lf/
    ├── Thunar/
    ├── dunst/
    ├── gtk-3.0/
    ├── gtk-4.0/
    ├── xfce4/
    ├── picom.conf
    ├── crypto.toml        (preserve_if_exists)
    ├── xinitrc/          → ~/.xinitrc (special handling)
    └── xsession/         → ~/.xsession (special handling)
```

---

## Files to Move

| Current Location | New Location | Deploy Target |
|-----------------|--------------|---------------|
| `config/backgrounds/` | `backgrounds/` | `~/.local/share/openriot/backgrounds/` |
| `config/applications/` | `share/applications/` | `~/.local/share/applications/` |
| `config/fonts/` | `share/fonts/` | `~/.local/share/fonts/` |
| `config/bin/` | `bin/` | `~/.local/share/openriot/bin/` |
| `config/*` | `config/*` | `~/.config/*` |

---

## Execution Steps

### Step 1: Move directories
```bash
mv config/backgrounds backgrounds/
mv config/applications share/applications
mv config/fonts share/fonts
mv config/bin bin/
```

### Step 2: Update packages.yaml

Replace the desktop.i3.configs section with a comprehensive list that copies config/* to ~/.config/

### Step 3: Update path references

**Files that need path updates:**

1. `config/polybar/config`:
   - Change: `$HOME/.local/share/openriot/config/polybar/scripts/...`
   - To: `$HOME/.config/polybar/scripts/...`
   - Also: `config = $HOME/.local/share/openriot/config/polybar/config`
   - To: `config = $HOME/.config/polybar/config`

2. `config/rofi/launcher.sh`:
   - Change: `$HOME/.local/share/openriot/config/bin/...`
   - To: `$HOME/.local/share/openriot/bin/...`

3. Any other hardcoded paths to config directories

### Step 4: Update PATH

Fish config should have:
```fish
fish_add_path --prepend $HOME/.local/share/openriot/bin
fish_add_path --prepend $HOME/.local/share/openriot/install
```

### Step 5: Test

1. Build installer: `make build`
2. Test on clean system
3. Verify apps are reading configs

---

## What Gets Preserved (preserve_if_exists)

- `config/crypto.toml`
- `config/i3/keybindings.conf`
- `config/i3/monitors.conf`
- `config/i3/windowrules.conf`
- `bin/openriot-lock.sh`
- `helix/config.toml`

---

## Benefits

1. **Apps find configs automatically** — no `--config` flags needed
2. **Standard Unix structure** — follows XDG Base Directory spec
3. **Cleaner repo** — separates different types of content
4. **Preserve still works** — user modifications protected
5. **Future-proof** — easier to add new apps

---

## Related Config Locations

| App | Default Config Location |
|-----|------------------------|
| i3 | `~/.config/i3/` |
| polybar | `~/.config/polybar/` |
| rofi | `~/.config/rofi/` |
| alacritty | `~/.config/alacritty/` |
| fish | `~/.config/fish/` |
| helix | `~/.config/helix/` |
| btop | `~/.config/btop/` |
| nvim | `~/.config/nvim/` |
| lf | `~/.config/lf/` |
| Thunar | `~/.config/Thunar/` |
| dunst | `~/.config/dunst/` |
| picom | `~/.config/picom.conf` |

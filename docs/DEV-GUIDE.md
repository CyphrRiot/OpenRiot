# OpenRiot Developer Guide

Reference material for contributions. The model doesn't read this every task
— it's extracted from AGENTS.md to keep the critical rules file lean.

## Code Tree

```
source/                 # Go source
  main.go               # CLI entry point
  installer/            # pkg_add, configs, source builds, mirrors
  imaging/              # wallpaper + screenshot pipeline
  polybar/              # bar generation + module config
  rofi/                 # launcher menus
  lock/                 # screen lock + blur
  crypto/               # wallet/encryption
  config/               # config file helpers
  backgrounds/          # wallpaper sets
  assets/               # embedded asset installer (cursors, icons)
  detect/               # hardware detection
  display/              # xrandr / monitor helpers
  network/              # wifi, wireguard
  audio/                # volume/pulse
  battery/              # power status
  notify/               # dunst notifications
  settings/             # UI settings panels
  update/               # system update checks
  helix/                # helix editor theming
  nmtui/                # network TUI
  resolution/           # display resolution TUI
  screenrec/            # screen recording
  screenshot/           # screenshots
  fonts/                # font installation
  fsutil/               # file utilities
  theme/                # color palette
  paths/                # path resolution
config/                 # dotfiles (i3, polybar, fish, nvim, etc.)
install/                # built binary + motd + packages.yaml
assets/                 # images, fonts, cursors, CSS
backgrounds/            # wallpaper images
Locked/                 # wallpaper sets (default, stealth)
docs/                   # documentation
tools/                  # debug/test scripts
```

## Go Conventions

- Prefer `var foo Bar` over `foo := Bar{}` (or `foo := 0`, `foo := ""`). Use
  short form only when initializing with a non-zero value.
- Avoid `unsafe` package (`unsafe.StringData`, `unsafe.SliceData`,
  `unsafe.Pointer`) for performance. Use safe alternatives. If a perf case
  genuinely needs it, raise it explicitly.
- Tests live in the same package as the code under test: `package <foo>`
  (internal / white-box). No `export_test.go` bridge files.

## Source & Comment Style

- In string literals, prefer escape sequences (`\u`, `\x`, `\U`) over raw
  bytes for invisible/hard-to-read non-ASCII characters (control chars, NEL,
  BOM, zero-width chars).

## Release Lifecycle

This is production code. Thousands of users depend on it.

### `make release` Order (must not change):
```
1. Sync packages.yaml
2. Bump VERSION
3. Build binary ← VERSION bump BEFORE build
4. Update README badge
5. Commit + tag
```

**The model never runs `make release`, `make install`, `make image`, or
updates VERSION/README badge. Only the user runs release steps.**

## Release Notes Requirements

### For every release:
| Rule | Why |
|------|-----|
| `docs/v{X.Y.Z}-Release-Notes.md` | Canonical naming, always with `v` prefix |
| Lines ≤80 chars | `lowdown -Tterm` renders in terminal with `less` |
| Header quote | Set the tone |
| Files Changed table | Users need to know what changed |
| Final Words | Close with personality |
| Update README.md TOC | Add link under `## 📋 Release Notes`, newest first. Keep only 10 most recent. Add `[Older releases →](docs/)` as the 11th line |

### Tables: use continuation rows, not wrap
```
| File | Change |
|------|--------|
| `foo.go` | **NEW** — does a thing that is |
| | very important for the system |
```

### Deployment:
- Installer calls `ShowReleaseNotes()` after install
- Renders with `lowdown -Tterm` and custom pager

## REFERENCE

- `docs/v*Release-Notes.md` — Release notes archive.
- Past lessons previously in `docs/Agents-Lessons.md` have been integrated
  into AGENTS.md as numbered rules.
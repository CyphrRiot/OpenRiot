# AGENTS.md - OpenRiot Codebase Guide

You are a cautious, obedient coding assistant for an OpenBSD Go project with extensive YAML configs. Your #1 rule: **NEVER execute git commit, git push, git add -A, or any permanent filesystem/git change without first showing the exact proposed command/diff and waiting for explicit "YES", "approve", or "go ahead" confirmation.** If unsure, ask. Do not assume or act autonomously on changes.

---

## CRITICAL PLATFORM RULES

1. **YOU ARE ON OPENBSD** — `uname -a` confirms `OpenBSD mini.openriot.org 7.9`
2. **Current branch:** `main` — i3/X11 migration COMPLETE
3. **Validate before acting** — Always verify packages with `pkg_info -Q <pkg>` on this OpenBSD system. Search https://openbsd.app/?search={pkg} for package availability.
4. **Keep main branch focused** — No mixing sway/wayland work

---

## Project Overview

**OpenRiot** is an OpenBSD desktop configuration tool. Install with:
```bash
curl -fsSL https://openriot.org/setup.sh | sh
```

**Platform:** OpenBSD 7.9 | **Language:** Go 1.26+ | **Build:** CGO-free cross-compilation

### Components

| Component | Location | Purpose |
|-----------|----------|---------|
| CLI binary | `source/main.go` | Runtime commands + installer |
| Installer modules | `source/installer/` | Copy configs, run commands, source builds |
| Config YAML | `install/packages.yaml` | Single source of truth for all install logic |
| Bootstrap script | `setup.sh` | Pre-repo setup (doas, git clone, delegates to binary) |
| Pre-built binary | `install/openriot` | Shipped with repo, built from source/ |

### Architecture

```
setup.sh (bootstrap, runs BEFORE repo exists)
├── check_openbsd_version
├── configure_doas_installurl    # must run before pkg_add
├── install_bootstrap_packages   # git
└── setup_repository            # git clone

openriot binary (handles install, AFTER repo exists)
├── openriot --install-packages  # pkg_add from packages.yaml
├── openriot --install          # CopyConfigs() → ExecCommands() → SourceBuilds()
└── openriot --source-builds    # build wlsunset, crush, etc. (called separately)
```

---

## Build Commands

- `make build` — Cross-compile for OpenBSD amd64 (standard release)
- `make linux` — Native build on Linux for testing (faster iteration)
- `make verify` — `make linux` + smoke test (`./install/openriot --version`)
- `make test` — Run Go tests (`cd source && go test ./...`)
- `make deps` — Tidy Go module dependencies
- `make clean` — Remove build artifacts
- `make iso` — Build full bootable ISO
- `make isotest` — Build ISO and run in QEMU
- `make binary-push` — Build + strip binary history + commit + force-push
- `make ultra` — Static build with optional UPX compression

### Binary-in-Git Workflow (CRITICAL)

The binary (`install/openriot`) is tracked in git **without history** to keep repo small.

| Scenario | Command |
|----------|---------|
| Config/code only | `git add -A && git commit -am "msg" && git push` |
| Code + binary change | `git add -A && git commit -am "msg"` then `make binary-push` |
| Binary-only change | `make binary-push` |

**`make binary-push` does:**
1. Checks if binary has >1 commit in git history
2. If yes: runs `git filter-repo --force --path install/openriot --invert-paths`, restores origin, rebuilds
3. Commits binary + force-pushes everything (`git push --force --all`)

**NEVER do:**
- `git add -A && git commit -am` then `make binary-push` (double-commit)
- Add `install/openriot` back to `.gitignore`

---

## Version Handling

- Read version from `VERSION` file (single source of truth)
- Read from Makefile: `OPENRIOT_VERSION = $(shell cat VERSION 2>/dev/null || echo "0.8")`
- Versions are injected at build time via `go build -ldflags="-X main.version=$(OPENRIOT_VERSION)"`
- **Never hard-code version numbers in Go files**
- OpenBSD version also injected: `-X main.openbsdVersion=$(OPENBSD_VERSION)`

---

## Go Patterns

- **Error handling:** functions return `error` as final return value
- **Logging:** Print to stdout with `[INFO]`, `[WARN]`, `[DONE]` prefixes (1 space after tag)
- **File operations:** Use `filepath.Join()`, check with `os.Stat()` + `os.IsNotExist()`
- **Test mode:** `testMode` flag uses `~/Code/OpenRiot` vs `~/.local/share/openriot` paths
- **Unused parameters:** Use `(void)param;` to avoid build failures (important for C code)

### Adding New CLI Flags

Pattern in `source/main.go`:
```go
if len(os.Args) >= 2 && os.Args[1] == "--your-flag" {
    // Do something
    return
}
```
**Always `return` after handling a flag.** Flags are checked sequentially; put `--version` and `--test` checks first.

---

## Module Structure (packages.yaml)

YAML structure for installation modules:

```yaml
module_name:
    packages: ["package-name"]        # Passed to pkg_add
    configs:                          # File deployment rules
        - pattern: "source_pattern/*"
          target: "optional_target"  # Override default ~/.config/
          preserve_if_exists: ["file1", "file2"]  # Skip if exists at dest
    commands:                         # Shell commands to run post-install
        - desc: "Human-readable description"
          cmd: "actual shell command"
    build: ["source build commands"]  # Only for type: "Source"
    depends: ["other.module"]          # Not currently enforced
    start: "Starting message"
    end: "Completion message"
    type: "Package|Source"
    critical: true|false
```

**Module categories:** `core`, `desktop`, `system`, `media`, `fonts`, `themes`, `source`

### YAML Categories vs. Actual Structure

The Go `Config` struct has: `Core`, `System`, `Desktop`, `Media`, `Fonts`, `Themes`, `Source`

However, `source` modules (wlsunset, crush, bibata-cursor, stormy) are nested under `desktop:` in packages.yaml, not under a top-level `source:` key. The `source` category in the YAML comments is misleading.

---

## Installer Modules

| File | Function | Behavior |
|------|----------|----------|
| `configs.go` | `CopyConfigs()` | Deploys files from `config/` to `~/.config/`, skips identical content |
| `packages.go` | `InstallPackages()` | Runs `pkg_info -e` first to skip already-installed packages |
| `execcommands.go` | `ExecCommands()` | Runs commands silently, continues on error |
| `sourcebuilds.go` | `SourceBuilds()` | Runs build commands as `/bin/sh`, skips on `[SKIP]` in output |
| `colors.go` | Color constants | `[INFO]`=cyan, `[WARN]`=yellow, `[DONE]`=green, `[ERR!]`=red |

### Config Deployment Rules

- **Source:** `config/` directory in repo
- **Destination:** `~/.config/` by default (override with `target: "~/path"` or absolute path)
- **Glob patterns:** `pattern: "i3/*"` copies all files recursively
- **Preserve:** `preserve_if_exists: ["file"]` skips copy if file already exists at destination
- **Identical content:** Skips write if destination content matches source (prevents spurious reloads)

### Source Build Pattern

Each build command checks if already installed before doing work:
```bash
if [ -x /usr/local/bin/wlsunset ]; then echo '[SKIP] wlsunset already installed'; else ...build...; fi
```
The installer checks for `[SKIP]` in output to suppress the `[DONE]` message.

---

## Package Management

- **pkg_add resolves versions** automatically, but `pkg_info -e` requires **exact versions** (e.g., `foot-1.26.1`, not `foot`)
- **Verify packages exist:** Search https://openbsd.app/?search={package}&current=on before adding
- **Package check:** `pkg_info -e <package>` returns success only with exact version
- **Install command:** `doas pkg_add -D unsigned <package>`
- **GetPackages()** deduplicates across all modules using a `seen` map
- **Rule:** Always use exact versions in packages.yaml. Bare names cause `pkg_info -e` to fail and packages get reinstalled.

---

## Desktop Applications

14 curated apps in Rofi launcher (`config/rofi/apps.txt`):

| App | Desktop File | Polybar Icon |
|-----|--------------|---------------|
| Firefox | firefox.desktop | 󰈹 |
| Flare | flare.desktop | 󰇢 |
| Telegram | tdesktop.desktop | 󰘦 |
| Terminal | terminal.desktop | 󰽒 |
| Helix | helix.desktop | 󰛞 |
| Thunar | thunar.desktop | 󰝰 |
| System Monitor | btop.desktop | 󰍹 |
| Crush | crush.desktop | 󰚩 |
| MPV | mpv.desktop | 󰕼 |
| AbiWord | abiword.desktop | 󰈙 |
| Settings | xfce4-settings.desktop | 󰒓 |
| Transmission | transmission-start.desktop | 󰇚 |
| Proton Mail | protonmail.desktop | 󰊫 |

## Polybar Workspace Icons

Format: `󰀻  ● 󰽒 󰈹   ○   ◉ 󰝰   ○`

| Icon | Meaning |
|------|---------|
| 󰀻 | App launcher (opens rofi) |
| ● | Focused workspace |
| ◉ | Unfocused with windows |
| ○ | Empty workspace |

**How it works:**
1. `workspaces.sh` queries i3 tree every 2 seconds → extracts window classes
2. `window-icon.sh` maps window class → Nerd Font icon (e.g., "firefox" → 󰈹)
3. Polybar displays the result

**Key distinction:**
- Rofi uses system icon theme (Adwaita) — `Icon=firefox` from .desktop
- Polybar uses Nerd Font — window class mapped via `window-icon.sh`

---

## Terminal Configuration

- **Current primary:** foot terminal (X11-native)
- **Config:** `config/foot/foot.ini`
- **Removed:** alacritty (not needed on OpenBSD)

---

## OpenBSD-Specific Technical Findings

| Issue | Solution |
|-------|----------|
| OpenBSD tar doesn't support `-J` flag | `xz -d file.tar.xz && tar -xf file.tar` |
| OpenBSD has no librt | POSIX timers stubbed with ENOSYS; use `ualarm()` instead |
| Keychron K2 Pro Mac mode | Sends Super keycodes as Mod4; i3 config uses `set $mod Mod4` |
| Git config timing | `git config --global pull.rebase true` must run **before** any git operations |
| curl usage in setup.sh | The installer already uses curl; banned commands apply to `wget` not `curl` |

---

## Setup Script Details

- **Log file:** `~/.cache/openriot/setup.log` (NOT `~/.local/share/openriot/`)
- **Real user detection:** Uses `getent passwd "$REAL_USER"` to find home directory — handles `doas`/`sudo` HOME mismatch
- **Installurl:** Configures `/etc/installurl` with `https://cdn.openbsd.org/pub/OpenBSD`
- **Git config:** Sets `pull.rebase=true` and `init.defaultBranch=master`
- **No git history in clone:** `git clone --depth 1 -b main` for minimal footprint

---

## Config File Locations

| File | Purpose |
|------|---------|
| `config/i3/config` | Main i3 WM config |
| `config/i3/keybindings.conf` | User keybindings (included by main config) |
| `config/polybar/config` | Polybar status bar |
| `config/fish/config.fish` | Fish shell config |
| `config/fastfetch/config.jsonc` | Fastfetch (replaces neofetch) |
| `config/btop/btop.conf` | System monitor |
| `config/helix/config.toml` | Helix editor |
| `config/crypto.toml` | Crypto price API config (preserved on install) |
| `config/applications/*.desktop` | Desktop entries installed to `~/.local/share/applications/` |
| `config/backgrounds/*` | Wallpapers installed to `~/.local/share/openriot/backgrounds` |
| `config/xinitrc/openriot-x11` | X11 startup script |
| `config/xsession/openriot-xsession` | Xenodm session script |

---

## Progress.md vs AGENTS.md

- **AGENTS.md:** Describes the codebase, patterns, conventions — what agents need to know
- **Progress.md:** Tracks active issues, testing status, known bugs — project-level state

---

## User Preferences

- **Propose before editing:** Show exact diff before applying changes
- **Wait for confirmation:** Only edit after user says "yes", "y", "proceed", etc.
- **One issue at a time:** Don't batch multiple changes together
- **Test locally first:** For Go changes: `make dev` → `make build`
- **Show proof:** Evidence that change works before asking for confirmation
- **Commit separately:** Each logical change gets its own commit
- **Verify packages:** Always check https://openbsd.app/ before adding packages

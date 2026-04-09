# AGENTS.md - OpenRiot Codebase Guide

You are a cautious, obedient coding assistant for an OpenBSD Go project with extensive YAML configs. Your #1 rule: NEVER execute git commit, git push, git add -A, or any permanent filesystem/git change without first showing me the exact proposed command/diff and waiting for my explicit "YES", "approve", or "go ahead" confirmation. If unsure, ask. Do not assume or act autonomously on changes.
Always summarize planned actions first and seek approval for anything destructive or permanent.

### PROPOSAL FORMAT (MANDATORY)

Before ANY change, present in this format. ONE issue at a time. Wait for approval before proceeding.

```
Title: {title of issue}
Description: {description of what we are fixing}
Files: {list files impacted}
Reason: {show reason / evidence}
```

Continue? [Y/n]

### BUILD AND COMMIT RULES (MANDATORY)

1. **NEVER commit without explicit user permission** — show diff first, wait for "yes"
2. **NEVER bump VERSION without explicit user permission**
3. **ALWAYS run `make` before a commit** (even config-only changes — to verify)
4. **Commit binary with source changes** — use `git add -A` or `git commit -am`
5. **The binary (`install/openriot`) IS the product** — if source changes, binary MUST be committed

### PACKAGE VERIFICATION (MANDATORY)

Before adding ANY package to `install/packages.yaml`, ALWAYS verify it exists in OpenBSD:
- Search: https://openbsd.app/?search={package_name}&current=on
- Do NOT assume a package exists
- Report the search URL in your proposal

### WAYBAR SCRIPTS (CURRENT)

```
waybar/scripts/
├── openriot-update.sh      # waybar module
├── system-metrics.sh       # CPU+RAM combined
├── waybar-battery.sh       # battery status
├── waybar-network          # network status
├── waybar-volume.sh        # volume control
├── weather-emoji-plain.sh  # weather display
└── wifi-selector.sh        # network picker
```

**Removed (unused):** waybar-cpu.sh, waybar-memory.sh, waybar-temp.sh, wireguard-click.sh, wireguard-status.sh, recording-indicator.sh

---

## Current Project Status (April 2026 - v0.7)

### Install Output Clean (April 2026)

The install output is now clean and consistent:
- 1-space indentation for all log levels
- Packages: "Installing packages (X new, Y already installed)..." — only shows new ones
- Configs: Category summaries like "[DONE] fish: 5 files"
- Commands: Run silently, only errors shown
- Source builds: Only show [DONE] if actually built (skip [SKIP] items)

### Migration Complete
Shell logic has been migrated to packages.yaml + Go binary. YAML is now single source of truth.

**setup.sh role:** Bootstrap only - checks, git clone, then delegates to openriot binary
**openriot --install role:** CopyConfigs() → ExecCommands() → SourceBuilds()

### Bootstrap vs Binary Flow
```
setup.sh (bootstrap, runs BEFORE repo exists)
├── check_openbsd_version
├── configure_doas_installurl     # must run before pkg_add
├── install_bootstrap_packages   # git
└── setup_repository             # git clone

openriot binary (handles install, AFTER repo exists)
├── --install-packages           # pkg_add from packages.yaml
├── --install                    # CopyConfigs + ExecCommands + SourceBuilds
└── --source-builds              # build wlsunset, crush, etc.
```

### Binary Installation Locations
| Binary | Location | Method |
|--------|----------|--------|
| openriot | ~/.local/share/openriot/install/openriot | Pre-built, shipped with repo |
| wlsunset | /usr/local/bin/wlsunset | Built from source via meson |
| crush | /usr/local/bin/crush | Downloaded pre-built from GitHub |
| waybar, sway, havoc, etc. | /usr/local/bin/ | OpenBSD packages via pkg_add |

### No ~/.local/bin
Nothing goes to ~/.local/bin. All binaries install to /usr/local/bin (system-wide) or stay in ~/.local/share/openriot.

---

## Build Commands

- `make build` — Cross-compile for OpenBSD amd64
- `make dev` — Native build on Linux for testing
- `make verify` — Build + smoke test (`--version`)
- `make iso` — Full ISO build
- `make binary-push` — Build + strip binary history + commit + force-push
- `make download-packages` — Download packages into `~/.pkgcache/`
- `make clean` — Clean build artifacts
- `make test` — Run Go tests

### Binary Commit Workflow (CRITICAL)

The binary (`install/openriot`) is tracked in git WITHOUT history to keep repo small.

**Workflow summary:**

| Scenario | Command |
|----------|---------|
| Config/code only (no binary change) | `git add -A && git commit -am "msg" && git push` |
| Binary + other changes | `git add -A && git commit -am "msg"` then `make binary-push` |
| Binary-only change | `make binary-push` |

**Step-by-step:**

1. **Code/configs only** (no `make build`):
   ```bash
   git add -A
   git commit -am "update sway config"
   git push
   ```

2. **Code changes requiring new binary**:
   ```bash
   # a) Normal commit for configs/code
   git add -A
   git commit -am "add new installer module"

   # b) Then rebuild and push binary (auto-strips history, force-pushes all)
   make binary-push
   ```

**`make binary-push` does:**
- Builds new binary
- Checks if binary has >1 commit in git history
- If yes: strips history with `git filter-repo`, restores origin, rebuilds
- Commits binary + force-pushes everything

**Warning:** `make binary-push` does `git push --force --all`. Collaborators must `git pull --rebase`.

**NEVER do:**
- `git add -A && git commit -am` then `make binary-push` (double-commit)
- Add binary commits manually (use `make binary-push`)
- Add `install/openriot` back to `.gitignore`

---

### Commit & Push Workflow (Legacy)

1. Make code/config changes
2. Run `make build` — cross-compile OpenBSD binary into `install/openriot`
3. **ALWAYS stage ALL changed files**: `git add -A` or `git commit -am "message"`
4. Show diff to user for approval
5. Only commit/push after explicit "yes" or "go ahead"

**CRITICAL**: The binary (`install/openriot`) IS the product. If source changes, the binary MUST be committed. Use `make binary-push` for binary-containing commits.

---

## Version Handling

- Current version: v0.7 (see `VERSION` file)
- Never hard-code version numbers
- Read version from `Makefile` variable `OPENRIOT_VERSION`

---

## Project Overview

OpenRiot is an OpenBSD desktop configuration tool. Install with:
```bash
curl -fsSL https://openriot.org/setup.sh | sh
```

### Main Components

1. **Binary (`source/main.go`)** - CLI with modes:
   - `--install`: Deploy configs, run commands, source builds
   - `--install-packages`: Install packages via pkg_add
   - `--source-builds`: Build software from source
   - `--packages`: List packages from YAML
   - Runtime: `--volume`, `--brightness`, `--power-menu`, `--wlsunset`, etc.

2. **Configuration (`install/packages.yaml`)**
   - YAML-based module definitions
   - Package lists per module (with full versions!)
   - Config file patterns and deployment rules
   - Command execution for post-install setup (desc/cmd format)

3. **Installer (`source/installer/`)**
   - `configs.go`: Config file deployment with preserve rules, category summaries
   - `packages.go`: Package installation logic, skip already-installed
   - `execcommands.go`: Post-install command execution (silent)
   - `sourcebuilds.go`: Source code compilation (skip if installed)
   - `colors.go`: ANSI color constants for output

4. **Setup Script (`setup.sh`)** - Bootstrap only

---

## Installer Colors

The openriot binary outputs colored text:
- `[INFO]` = Blue (cyan)
- `[WARN]` = Yellow
- `[DONE]` = Green
- `[DEBUG]` = White

Color constants defined in `source/installer/colors.go`.

---

## Module Structure (packages.yaml)

```yaml
module_name:
    packages: ["full-version-package-1.0.0"]  # pkg_add packages (use FULL versions!)
    configs:  # File deployment patterns
        - pattern: "source_pattern/*"
          target: "optional_target"
          preserve_if_exists: ["file1", "file2"]
    commands:  # Shell commands with descriptions
        - desc: "Human-readable description"
          cmd: "actual shell command"
    build: ["source build commands"]  # For type: "Source"
    depends: ["other.module"]
    start: "Starting message"
    end: "Completion message"
    type: "Package|Source"
    critical: true|false
```

**Module categories:** core, desktop, system, media, fonts, themes, source

**IMPORTANT:** Always use FULL package versions (e.g., `havoc-0.7.0` not `havoc`) to avoid reinstalling.

---

## Go Patterns

- Error handling: functions return `error` as final return value
- Logging: Print to stdout with `[INFO]`, `[WARN]`, `[DONE]` prefixes (1 space after tag)
- File operations: Use `filepath.Join()`, check with `os.Stat()` + `os.IsNotExist()`
- Test mode detection: `testMode` flag uses `~/Code/OpenRiot` vs `~/.local/share/openriot`

---

## Adding New openriot CLI Flags

Pattern in `source/main.go`:
```go
if len(os.Args) >= 2 && os.Args[1] == "--your-flag" {
    // Do something
    return
}
```
**Always `return` after handling a flag.**

---

## OpenBSD-Specific Technical Findings

### OpenBSD has no librt
- POSIX timer functions stubbed with ENOSYS
- Use `ualarm()` instead (available in libc)

### OpenBSD tar doesn't support -J flag
- Solution: `xz -d file.tar.xz && tar -xf file.tar`

### Git config timing
- `git config --global pull.rebase true` must run **before** any git operations

### OpenBSD clang -Werror
- Unused parameters/functions cause build failures
- Use `(void)param;` for unused parameters

### Package Versions
- OpenBSD packages have specific versions (e.g., `havoc-0.7.0`, not `havoc`)
- Always check with `pkg_info -Q <package>` for correct version

---

## User Preferences

- **Propose before editing**: Show exact diff before applying changes
- **Wait for confirmation**: Only edit after user says "yes", "y", "proceed", etc.
- **One issue at a time**: Don't batch multiple changes together
- **Test locally first**: For Go changes: `make dev` → `make build`
- **Show proof**: Evidence that change works before asking for confirmation
- **Use Nerd Font icons**: Prefer Nerd Font in fastfetch (not waybar JSON modules)
- **No polling if possible**: Use signals (SIGUSR1) to refresh waybar modules
- **Commit separately**: Each logical change gets its own commit

---

## Recent Completed Work (v0.7)

### Install Output Cleanup (April 2026)
- Consistent 1-space indentation for all log levels
- Remove duplicate "Deploying configuration files..." header
- Commands run silently (only errors shown)
- Package skip messages removed (show only new installs)
- Category summaries for config deployment

### Command Descriptions (April 2026)
- Changed `commands: ["shell command"]` to `commands: [desc: "...", cmd: "..."]`
- Cleaner output: "Adding fish to /etc/shells" instead of full command

### Terminal Switch (April 2026)
- Switched from alacritty to havoc terminal
- havoc config at `config/havoc/havoc.cfg`

### Package Version Fixes (April 2026)
- havoc: 0.17.5p0 → 0.7.0 (correct OpenBSD version)
- All packages now use full version strings

### Waybar Scripts Cleanup (April 2026)
- Removed unused scripts: waybar-cpu.sh, waybar-memory.sh, waybar-temp.sh, wireguard-*.sh, recording-indicator.sh
- system-metrics.sh now combines CPU+RAM
- Remaining: 10 files (was 16)

### Fastfetch Logo Fix (April 2026)
- Changed config.jsonc from "blowfish" to "openbsd_small"
- Also set in fish alias and function

### Screenshot Compression (April 2026)
- screenshot.png (1.2MB) → screenshot.jpg (298KB)

---

## Next Steps

1. Test full install after these changes
2. Verify havoc terminal works correctly
3. Test waybar modules refresh properly

---

## Related Files

| File | Purpose |
|------|---------|
| `Progress.md` | Project-level tracking, known bugs |
| `VERSION` | Current version (v0.7) |
| `Makefile` | Build commands |
| `install/packages.yaml` | Single source of truth for install |
| `source/installer/` | Go installer modules |

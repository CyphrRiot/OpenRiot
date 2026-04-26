# Agent Lessons Learned

## Shell Color Variables
Use `$(printf '\nnn')` not `'\nnn'` or `$'...'`. Test in isolation before using:

```sh
# WRONG - doesn't expand properly in heredocs
RED='\nn[0;31m'

# WRONG - $'...' not always supported
RED=$'\nn[0;31m'

# CORRECT
RED=$(printf '\nn[0;31m')
```

## Drive Detection (OpenBSD)
- Root drive: Parse `root on sdXa` from `dmesg`
- Removable: Parse `removable` from `dmesg`
- Softraid parents: Check disklabel for RAID partition type (`^  [a-z]:.*RAID`)
- Protect: `root_drive + softraid_parents`
- Offer for burn: `removable + internal_non_OS`

## Heredocs in Shell
- Closing marker must be at start of line (no indentation)
- Check for orphan closing markers after multi-edit sessions
- Run `sh -n script.sh` to verify syntax

## packages.yaml Build Entries

**ALWAYS include `desc:` prefix** in build entries, even for heredocs:
```yaml
build:
    - |
        desc: "Description of what this does"
        cmd: |
            actual shell commands here
```
Without `desc:`, the module will print blank `[INFO] ...`

## OpenBSD Quirks
- `grep -A` does NOT work - use `sed` or `view` + offset
- No `xxd` - use `od -c` or `strings`
- Many bash commands behave differently

## Eth Network Detection
Return immediately on `status: active` - don't set flag and fall through:
```go
if isEth && strings.Contains(line, "status: active") {
    return current, true  // IMMEDIATELY
}
```

## Shell Scaling Without X11
Default to 1920x1080, don't fail:
```go
if os.Getenv("DISPLAY") == "" {
    width = 1920  // Default to 1080p
}
```

## i3 Window Rules
- Use `ppt` not `%` for whole-screen percentage
- `for_window [instance="^floating_center"]` (instance, not class)

## Porting Shell Scripts to Go

**NEVER write new code that already exists in the codebase.**

### Before Writing Any Go Code
1. Read the ENTIRE shell script and note every output line and its EXACT format
2. Search the existing codebase for functions that do similar things
3. Reuse existing functions - don't rewrite them

### After Writing Go Code
1. **grep sweep** of ALL print/printf statements in the package
2. Verify every output message matches the shell script character-for-character
3. Check for: colors, brackets (`[INFO]` not `INFO`), spacing, punctuation

## YAML Indentation Silently Breaks Struct Mapping

A single indent level can silently drop an entire category. The loader produces **NO error** — just empty maps:

```yaml
# WRONG — media nested under desktop, Go sees empty cfg.Media
desktop:
  firefox: ...
  media:
    transmission: ...

# CORRECT — media is top-level
desktop:
  firefox: ...
media:
  transmission: ...
```

Always verify struct population after YAML changes.

## HTTP Client Timeouts

`http.Get()` has **NO timeout by default** and hangs indefinitely on slow networks:

```go
// WRONG
resp, err := http.Get(url)

// CORRECT
client := &http.Client{Timeout: 2 * time.Minute}
resp, err := client.Get(url)
```

## POSIX sh Pipeline Exit Codes

There is **no `pipefail`** in POSIX `sh`. This masks failures:

```sh
# WRONG — exits with tee's status (0) even if openriot fails
./openriot --install 2>&1 | tee -a "$LOG_FILE"

# CORRECT — capture status before tee
./openriot --install > /tmp/out 2>&1
status=$?
tee -a "$LOG_FILE" < /tmp/out
exit $status
```

## MakeIcon Font Paths

`MakeIcon` searches fonts in this order. **Do NOT hardcode a single path**:

1. `../assets/fonts/` relative to binary (production)
2. Same-directory `assets/fonts/` (development)
3. `~/.local/share/fonts/FiraCode/` (user-installed)

## packages.yaml: build vs commands

- `type: "Package"` → uses `commands:` (shell commands)
- `type: "Source"` → uses `build:` (source compilation steps)

`build:` under a `Package` type module is **dead code** — never executed.

## setup.sh --local Flag

For offline end-to-end testing without network:

```bash
rsync -a --delete ~/Code/OpenRiot/ ~/.local/share/openriot/
sh ~/.local/share/openriot/setup.sh --local
```

Skips all remote git ops and reads version from local `VERSION` file.

## make release Order (CRITICAL)
```
1. Sync packages.yaml
2. Bump VERSION
3. Build binary ← VERSION bump BEFORE build
4. Update README badge
5. Commit + tag
```

## make-img.sh Details

### Image Sizing
- Starts at 2GB fixed (handles tarball + packages)
- `shrink_img()` truncates to actual content after injection
- Calculation: `used_KB * 1.1 + 32MB buffer`, aligned to 4MB

### Auto-Cleanup
- Old `openriot.img` deleted at start of `build_img()`
- Old `openriot.tgz` deleted at start of `create_site()`
- Old packages cleaned before download (not in list OR in exceptions.yaml)

### Package Management
- `exceptions.yaml` lists packages to exclude from image
- Old packages not in current list are auto-deleted
- `pkg_add` handles deps at install time, not download time

### Drive Detection
- `[ROOT]` = Boot drive (from `root on sd1a` in dmesg) + softraid parents (drives with RAID partitions)
- `[WARN]` = Removable drives (from `removable` in dmesg)
- `[INFO]` = Internal empty drives

### Progress Output
- Use `printf "\r%80s\r" ""` to clear progress lines between sections
- Don't use `\n` in progress loops - use `echo ""` after loop completes
- `set -e` causes exit on non-zero returns - use `|| true` for expected failures
- `grep -c .` returns exit 1 on empty input - use `|| 0` fallback
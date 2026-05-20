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
- User runs **fish shell**, NOT bash — `<(...)` process substitution fails, use `(cmd | psub)`
- `pkg_delete` uses base names, NOT full version strings: `doas pkg_delete autopep8`
- Check reverse deps before suggesting removal: `pkg_info -R <pkg>`

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

## Crush LSP Configuration

If removing a language server package (e.g., `py3-python-lsp-server`), check `~/.config/crush/crush.json` for a matching `"lsp"` entry. Removing the package without updating the config breaks LSP for that language in Crush:

```json
"lsp": {
    "python": { "command": "pylsp" }
}
```

Either reinstall the package or remove the LSP entry from `crush.json`.

## PLATFORM — OpenBSD + Fish

| Bash | Fish |
|------|------|
| `<(cmd)` | `(cmd \| psub)` |
| `&&` | `; and` |
| `\|\|` | `; or` |

- `pkg_delete` base names only: `doas pkg_delete autopep8`
- Check reverse deps: `pkg_info -R <pkg>`

## DOCS

- `docs/architecture.md` — system design
- `docs/debugging.md` — debug workflow
- `docs/packages.md` — package/module rules
- `docs/polybar-performance.md` — polybar config notes
- `docs/proton-drive-module.md` — Proton Drive integration

---

## v6.7 Critical Lessons — Read The Code Before Touching Anything

### READ EXISTING CODE Before Adding Anything New

**Do NOT use `find`, `grep`, or guess when you can read the exact function.**

I used `find /home/grendel -name screenrec.png` to locate where `--make-icon` writes files, when the answer was 10 lines into `source/installer/utils.go` → `MakeIcon()`:

```go
repoDir := filepath.Join(filepath.Dir(ex), "..", "config", "icons")
```

The repo binary writes to `../config/icons/`. The installed binary writes to `~/.local/share/openriot/config/icons/`. One `view` call would have told me this. `find` was lazy and wrong.

**Rule:** If a function writes files, read the function. If a command exists, read its implementation. Never guess.

---

### `patch` Handles Offset Hunks; `git apply` Does Not

When upstream moves `fn notify()` from line 720 to line 857, `git apply` aborts with "patch does not apply" even though the surrounding context is identical. `patch -p1` finds the hunk by context and applies with an offset.

```sh
# WRONG — fails when line numbers shift
if git -C "$dir" apply --check "$patch"; then ...

# CORRECT — handles offset hunks across versions
if patch -p1 --dry-run -f < "$patch"; then ...
```

**Rule:** For patches that must work across multiple upstream commits, use `patch` not `git apply`.

---

### `make-icon` Takes Actual Characters, Not Hex Escapes

ImageMagick `label:` interprets the literal string. Passing `\uf03d` writes the 6-character string "\uf03d" into the PNG, not the Nerd Font glyph. Pass the actual character.

```bash
# WRONG — produces garbage PNG
openriot --make-icon screenrec "\uf03d"

# CORRECT — actual Nerd Font character
openriot --make-icon screenrec ""
```

**Rule:** When passing Unicode to shell commands, use the actual character. Test the generated PNG if unsure.

---

### Never Modify Files Without a Proposal (Even "Trivial" Changes)

I wrote `gurk-patch.diff` from scratch without proposing, corrupted it, and had to be rescued. I flattened `~/Code/gurk/gurk-rs/` without proposing. I updated `docs/Gurk.md` without proposing. Every unauthorized change introduced a problem that required cleanup.

**Rule:** Propose → wait for "go" → execute. No exceptions. Not for docs, not for patches, not for one-line fixes.

---

## CODE TREE

```
source/
  main.go               # CLI entry point (cmd dispatch)
  installer/            # pkg_add, configs, source builds, mirrors
  imaging/              # wallpaper + screenshot pipeline
  polybar/              # bar generation + module config
  rofi/                 # launcher menus
  lock/                 # screen lock + blur
  crypto/               # wallet/encryption
  config/               # config file helpers
  backgrounds/          # wallpaper sets
  assets/               # embedded images/fonts
  detect/               # hardware detection
  display/              # xrandr / monitor helpers
  network/              # wifi, wireguard
  audio/                # volume/pulse
  battery/              # power status
  notify/               # dunst notifications
  settings/             # UI settings panels
  update/               # system update checks
config/                 # dotfiles (i3, polybar, fish, nvim, etc.)
install/                # built binary + motd + packages.yaml
Locked/                 # wallpaper sets (default, stealth)
```

---

## v5.10 Critical Lessons — The xrandr Disaster

### xrandr is NOT a Lightweight Status Query

`xrandr` opens `/dev/drm0` and blocks on the kernel DRM event queue. **Every call is a kernel round-trip that stalls the X11 server.** Do NOT use it in timers, polling loops, or polybar modules. The HDMI module was calling it 4 times per spawn, synchronized with volume/battery/network, creating a 300–400ms stall every 30 seconds.

| What | Use Instead |
|---|---|
| Query connected outputs | `i3-msg -t get_outputs` (IPC socket read) |
| Query active monitors | Parse `get_outputs` JSON `active` field |
| Reconfigure displays | `xrandr` — but only on user click, never on timer |

**Rule:** `xrandr` is for **configuration**, not **polling**. Reserve it for `ToggleHDMI()` clicks.

---

### Verify the Actual Config Path

The i3 autostart explicitly launches `picom --config $HOME/.local/share/openriot/config/picom.conf`. Every "test" we ran on `~/.config/picom.conf` or `~/.config/picom/picom.conf` was on a **dead file that picom never read**. Always confirm the live path:

```sh
ps aux | grep picom   # see --config path
```

Same for polybar — `--polybar-setup` generates `~/.config/polybar/config.ini` from `~/.local/share/openriot/config/polybar/config.ini`. Editing the repo file is not enough; the template must be synced.

---

### Objective Measurement Beats Subjective Feeling

We spent two releases "optimizing" based on typing feel. The fix came only after `tools/detect-hang.sh` measured **exact 32-second gap intervals** correlated with polybar module spawns. Without microsecond timing + CPU snapshots, we would still be guessing.

```sh
# detect-hang.sh pattern: print dots every 100ms, scream on gap >300ms
```

**Rule:** When debugging periodic stalls, measure before optimizing. Feelings lie. Timestamps don't.

---

### Stagger Polybar Module Intervals

Synchronized intervals (all at 30s) cluster fork+exec + blocking calls into a simultaneous burst. Always stagger:

| Module | Interval |
|---|---|
| volume | 31 |
| battery | 32 |
| network-eth | 33 |
| hdmi | 65 |

---

### When the User Names the Root Cause, Start There

The user correctly identified picom as the primary culprit from day one. We spent time rewriting polybar modules, blaming the clipboard daemon, and chasing the Intel driver. Picom fading was indeed a major factor, but the xrandr synchronization storm was the periodic stutter we couldn't explain. **Reverse confirmation bias is still bias** — test the user's hypothesis first before building competing theories.

---

### Never Cache in a Short-Lived Binary

`openriot` is an ~8.3MB CLI that exits after each call. In-process caching is useless — the process dies before the cache can be reused. When performance matters, either:
- **Reduce calls** (one parse, multiple answers — like `--polybar-all` and the new `i3Outputs` refactor)
- **Use a persistent IPC source** (i3 socket, cache file written by a long-lived process)
- **Beware fork+exec overhead** — an 8.3MB binary spawn is not free, but it's acceptable if staggered

---

## v6.12 Critical Lessons — Current Snapshot Drift

### `uname -r` on `-current` Returns Version Numbers, Not "current"

OpenBSD `-current` snapshots report a future release version via `uname -r` (e.g., "8.0"). The string "current" only appears in `sysctl kern.version`.

```go
// WRONG — returns "8.0" on -current, fails IsSnapshot() check
version, _ := exec.Command("uname", "-r").Output()

// CORRECT
version, _ := exec.Command("sysctl", "-n", "kern.version").Output()
// "OpenBSD 7.9-current (GENERIC.MP) #475: ..."
```

**Rule:** Detect `-current` via `kern.version`, never `uname -r`.

---

### `fsutil.CopyFile()` Skips Same-Size Overwrites

The installer uses size comparison to skip redundant copies. When a config edit does not change byte count, it is silently never deployed to users.

**Rule:** Any config change must alter file size OR be verified with a full install test.

---

### Terminal Height Detection Fails on Redirected stdout

`stty size` and `TIOCGWINSZ` on `os.Stdout.Fd()` fail with "Inappropriate ioctl" when stdout is a pipe (e.g., script invocation, popup terminal). Always probe `/dev/tty` first.

```go
// WRONG — falls back to 24 on pipe/redirect
tty, _ := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)

// CORRECT
tty, err := os.Open("/dev/tty")
if err == nil {
    ws, err := unix.IoctlGetWinsize(int(tty.Fd()), unix.TIOCGWINSZ)
    // ...
}
```

**Rule:** Read terminal dimensions from `/dev/tty`, never `os.Stdout`.

---

### Kernel Age Alone Is Insufficient for `-current` Drift Detection

A fresh kernel build can still have unverifiable snapshot packages. Probing `pkg_add` output for `signify`, `gpg`, `signature`, or `pubkey` errors is the only reliable signal.

**Rule:** Detect package/base drift with a live `pkg_add -D snap -I` probe, not just kernel date math.

---

### `.desktop` `Exec=` Lines Need `%f` for File Passing

Thunar passes selected files as arguments to `.desktop` `Exec=` commands. Without `%f`, mpv launches with no file.

```
# WRONG
Exec=mpv

# CORRECT
Exec=mpv %f
```

**Rule:** Any `.desktop` entry for a file-handling app must include `%f` or `%U`.

---

### Polybar Deprecated Parameters

| Old | New | Why |
|-----|-----|-----|
| `content =` | `format =` | Deprecated, produces warning |
| `wm-restack = i3` | (remove) | Conflicts with `override-redirect = false` |

**Rule:** Update polybar configs against current upstream docs, not old templates.

---

### Release Notes Lines Must Never Exceed 80 Characters

`lowdown -Tterm` renders with a fixed terminal width. Lines >80 chars wrap badly in the custom pager. Verify with `awk '{print length, $0}' file | sort -rn | head` before any write.

**Rule:** Run width check before writing release notes. Never submit without verifying.

---

### Never Ignore AGENTS.md Constraints "Because You're An AI"

Every rule in AGENTS.md exists because a past session violated it and produced broken output. "Forgetting" is not a valid explanation. The rules must be enforced at both the agent level (self-check) and the tool level (hooks).

**Rule:** If a constraint is documented, it is non-negotiable. No exceptions for "small" changes.

---

## v7.3 Critical Lessons — VERSION, Batching, and Proposals

### NEVER Touch VERSION or README Badges

The `VERSION` file and the `README.md` version badge are **only** updated by `make release`, which the user runs manually. Changing them in a normal session is a direct violation of the release lifecycle.

**Rule:** Do not touch `VERSION`, `README.md` badges, or any release metadata. Ever.

---

### ONE CHANGE AT A TIME Means Exactly That

Batching related edits into a single block — even when they feel "trivial" — violates the user's workflow. Each logical edit requires a separate "go" or "proceed".

**Rule:** One `edit` or `write` per approval. No batching. No "while I'm here" changes.

---

### Proposals Are Not Optional

Skipping the proposal step because a change seems obvious or small leads to unauthorized work. The user rejected an attempt to proceed without an explicit proposal by saying "MAKE ONE CHANGE AT A TIME. WAIT FOR 'go'".

**Rule:** Propose → wait for "go" → execute ONE change → wait for "go" → next change. No exceptions.

---

## v7.5 Critical Lessons — OpenBSD sed, doas Redirections, and fish Heredocs

### OpenBSD `sed -i` Requires a Backup Extension Argument

GNU `sed -i` edits in place with no backup. OpenBSD `sed` is POSIX. `-i` alone treats the next argument as a backup extension:

```sh
# WRONG — silently no-ops because 's/foo/bar/' becomes the backup suffix
sed -i 's/foo/bar/' /etc/file

# CORRECT — empty string means no backup
sed -i '' 's/foo/bar/' /etc/file

# BETTER — avoid sed -i entirely on OpenBSD
awk '{gsub(/foo/, "bar"); print}' /etc/file > /tmp/file.new && doas cp /tmp/file.new /etc/file

# BEST for single-line replacement — printf pipe to ed
printf '/^pass/c\npass out\t\t# comment\n.\nw\nq\n' | doas ed -s /etc/pf.conf
```

**Rule:** Never use `sed -i` on OpenBSD without testing `-i ''` first. Prefer `ed` or `awk` with a temp file for in-place edits.

---

### `doas` Does Not Elevate Shell Redirections

The unprivileged shell opens the output file before `doas` elevates the command:

```sh
# WRONG — shell tries to open /etc/file before doas runs
$ doas echo "foo" > /etc/file
zsh: permission denied: /etc/file

# CORRECT — entire pipeline elevated
$ doas sh -c 'echo "foo" > /etc/file'
```

**Rule:** Any redirection to a privileged path must be inside `doas sh -c '...'`.

---

### Heredocs Fail in fish Shell

The `<<'EOF'` syntax is POSIX `sh`, not fish:

```sh
# WRONG — fish: Expected a string, but found a redirection
fish> doas ed -s /etc/pf.conf <<'EOF'
...
EOF

# CORRECT — pipe commands via stdin, no heredoc
printf '/^pass/c\npass out\t\t# comment\n.\nw\nq\n' | doas ed -s /etc/pf.conf

# ALSO CORRECT — use sh -c to get a POSIX subshell
doas sh -c 'ed -s /etc/pf.conf <<EOF
/^pass/c
pass out		# comment
.
w
q
EOF'
```

**Rule:** Never use heredocs in commands that may run under fish. Use `printf | cmd` or `sh -c`.

---

### Function Signature Changes Require Updating All Call Sites

Changing a function signature (e.g., adding a parameter) breaks every caller, including tests:

```go
// Changed this
func createInstallSite(siteDir string) error

// To this
func createInstallSite(siteDir, repoPath string) error

// But forgot to update the test
createInstallSite(tmpDir)  // FAIL: not enough arguments
```

**Rule:** After any signature change, `grep` for every call site and run `make test` before declaring it done.

---

### Run `make test` After Every Edit, Not Just `make`

`make` only compiles. `make test` runs the full test suite and catches signature mismatches, broken tests, and regressions:

```sh
# WRONG — compiles but misses broken test
make

# CORRECT
make test
```

**Rule:** `make test` is the minimum verification step after any code change. No exceptions.

---

### Stop at Two Failures on the Same Problem

Iterating a broken approach more than twice wastes time and produces worse solutions. After two distinct attempts fail:

1. Stop
2. State what failed
3. Ask the user what to do

**Rule:** Two failures on the same problem means stop. Do not attempt a third time unilaterally.
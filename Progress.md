# OpenRiot — Project Progress

**v1.4** · commit `899d893` · `https://github.com/CyphrRiot/OpenRiot.git`

**Quick test:** `rm -rf ~/.local/share/openriot && curl -fsSL https://openriot.org/setup.sh | sh`

---

## What We Know Works (from hardware test logs)

| Feature | Log Evidence | Status |
|---------|-------------|--------|
| Package skip detection | 32408649: all `[SKIP]` on re-run | ✅ Confirmed |
| yq/Python YAML fallback | 32408649: 41 packages found | ✅ Confirmed |
| curl/git removed (base sys) | 32408649: 41 packages (was 43) | ✅ Confirmed |
| Nerd Font installs | 32408649: `JetBrainsMono Nerd Font installed.` | ✅ Confirmed |
| Fish prompt path color | Changed to `brblue` | ✅ Fixed |
| Nerd Font quiet unzip | Not yet tested on hardware | ❓ Untested |
| Nerd Font skip if installed | Not yet tested on hardware | ❓ Untested |
| Git pull on every run | Log still shows old code (pre-fix) | ❓ Untested |
| Source builds direct in setup.sh | Written but NOT committed | ❌ Not deployed |
| Sway autostart on TTY1 | User hasn't rebooted yet | ❓ Untested |
| Foot Nerd Font rendering | User hasn't tested | ❓ Untested |
| waybar icons | User hasn't tested | ❓ Untested |
| wlsunset installed | User reports missing | ❌ BROKEN |
| crush installed | Fails: `out of memory` on go-sqlite3-wasm | ❌ BROKEN |
| Upgrade flow | Not tested | ❓ Untested |

---

## Known Bugs (pending fix — NOT yet committed)

### 1. Source builds broken (openriot --source-builds fails)
- **Symptom:** `wlsunset` and `crush` not installed
- **Root cause:** `./openriot --source-builds` in setup.sh calls Go binary that can't parse YAML on OpenBSD (NULL bytes, fallback broken)
- **Fix written:** Direct shell commands in setup.sh `run_source_builds()` function — **NOT committed**

### 2. Git pull skips bug fixes between version bumps
- **Symptom:** Re-running setup.sh shows "Already on latest version (1.1)" and skips `git pull`
- **Root cause:** `setup_repository()` only pulls when `local_ver < remote_ver`
- **Fix written:** Check `git rev-list HEAD..origin/main` for commits ahead — **NOT committed**

### 3. crush compilation fails (out of memory)
- **Symptom:** `go install github.com/charmbracelet/crush@latest` → `fatal error: runtime: out of memory` on `ncruces/go-sqlite3-wasm`
- **Root cause:** Go's SQLite WASM dependency needs more memory than available during compilation
- **Fix:** Use pre-built binary from GitHub releases instead of `go install`
- **Status:** NOT YET WRITTEN

---

## Commits Pushed (7599086 is latest)

```
7599086 Nerd Font: skip if installed, quiet unzip, move earlier in bootstrap
d03f962 Always git pull on every run to pick up bug fixes
03e676a Remove curl and git from packages.yaml (base system)
60b8c8d Replace grep YAML parser with yq+Python, add JetBrainsMono Nerd Font
654f3f7 Fix package skip detection and update python package name
```

---

## Pending Changes (setup.sh — NOT committed)

Three fixes are written but not pushed:

1. **run_source_builds()** — replaces `./openriot --source-builds` with direct shell commands that check `command -v` before building each tool
2. **setup_repository()** — always git pull by checking `git rev-list HEAD..origin/main`
3. **Fish prompt path color** — changed `blue` → `brblue` in `config/fish/config.fish`

---

## Next Steps

1. **Fix crush compilation** — download pre-built binary from GitHub instead of `go install`
2. **Commit and test** — push pending changes, run on hardware
3. **Verify source builds** — wlsunset, crush, bibata all install correctly
4. **Test font rendering** — foot, waybar, lsd all show icons

---

## Debug Commands (on OpenBSD)

```sh
# Share latest log
openriot --share-log

# Check what's installed
which wlsunset  # should exist
ls ~/.local/share/fonts/JetBrainsMono/  # should exist
fc-list | grep JetBrainsMono           # should show "NF" font

# Check sway autostart
grep -n "exec sway" ~/.config/fish/config.fish

# Manual sway test
sway -d 2>&1 | head -100
```

---

## Root Causes (what we've learned)

| Problem | Root Cause |
|---------|-----------|
| `pkg_info -m` always fails | Shows **maintainer**, not installed status |
| `pkg_info -e` skips base packages | curl/git not in pkg DB (base system) |
| grep parsed commands as packages | YAML commands are quoted strings, not bare names |
| grep parsed config patterns | `pattern:` lines matched by `grep -E '^ +- [a-zA-Z]'` |
| Source builds never run | `./openriot --source-builds` binary fails on OpenBSD |
| Git pull skipped between versions | Only triggers when local < remote version |
| crush fails to compile | `go-sqlite3-wasm` dependency OOMs on 8GB OpenBSD |

---

**Last updated:** Apr 6, 2026

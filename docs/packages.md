# Package Management

## Two packages.yaml Files

| Location | Purpose |
|----------|---------|
| `~/Code/OpenRiot/install/packages.yaml` | Source repo (git controlled) |
| `~/.local/share/openriot/install/packages.yaml` | Installed copy |

## Commands

### --check-packages
Checks packages.yaml against installed packages.
1. Checks CWD first for `install/packages.yaml`
2. Falls back to `~/.local/share/openriot/install/packages.yaml`
3. Prints which file is being checked

### --sync-packages
Updates packages.yaml to match installed versions.
1. Same file lookup as `--check-packages`
2. Prints which file is being updated
3. Updates matching packages

## Output Format

All commands use consistent 4-character brackets:

| Tag | Usage |
|-----|-------|
| `[INFO]` | Informational messages |
| `[WARN]` | Warnings, mismatches |
| `[FAIL]` | Errors, failures |
| `[MISS]` | Missing packages |
| `[DONE]` | Success |

## Skip Checks in Source Builds

When adding skip checks for source builds, verify the actual output filename:

```yaml
# WRONG - build produces libcalc.so
- "if [ -x /usr/local/lib/rofi/librofi-calc.so ]; then echo '[SKIP]'; fi"

# CORRECT
- "if [ -x /usr/local/lib/rofi/libcalc.so ]; then echo '[SKIP]'; fi"
```

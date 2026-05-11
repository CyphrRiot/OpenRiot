# Output Format Standards

## Bracketed Status Indicators

All output uses consistent 4-character brackets:

| Tag | Meaning | Examples |
|-----|---------|----------|
| `[INFO]` | Informational | Checking file, Starting process |
| `[WARN]` | Warning | Mismatches, non-fatal issues |
| `[FAIL]` | Failure | Errors, command failures |
| `[MISS]` | Missing | Package not installed |
| `[DONE]` | Success | Completed, in sync |

## Usage Rules

- **Always 4 characters** inside brackets
- `[WARN]` not `[MISMATCH]`
- `[FAIL]` not `[ERR!]` or `[ERR]`
- `[MISS]` not `[MISSING]`
- `[DONE]` not `[OK]`

## In Source Code

Check for consistency with grep before adding new output:

```bash
grep -E '\[.{1,4}\]' source/**/*.go
```

## Exceptions

Some UI elements may use different formatting:
- TUI prompts (`[X]`, `[!]`) are user-facing and may vary
- Notification icons (`"[!]"`) are for external tools
- Logger emoji strings may use different brackets for specific status types

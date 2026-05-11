# Gurk — Signal TUI Client

Pure-Rust Signal terminal client with zero Java/GTK dependencies.

## Installation

Gurk is **pre-built and included** in the OpenRiot install. During `openriot --install`, the gurk binary is deployed to `~/.local/share/openriot/config/bin/gurk` and symlinked to `~/.local/bin/gurk`. The OpenBSD SIGSEGV fix is already baked in.

No manual steps needed for normal installation.

## Rebuilding from Source

Use `./scripts/gurk.sh` only when you need to rebuild gurk from source (e.g., updating to a new version, or refreshing the pre-built binary for image releases).

Source is cached at `~/src/gurk-rs` so subsequent builds are fast.

## OpenBSD SIGSEGV Fix

**Problem:** gurk crashes with SIGSEGV on OpenBSD when receiving incoming messages.

**Root cause:** `notify-rust` (desktop notification library) calls `/proc/self/exe` to get the executable path. This syscall doesn't exist on OpenBSD, causing a segmentation fault.

**Fix:** The patch in `scripts/gurk-patch.diff` hard-disables the notification code by replacing the condition with `if false`.

### Patch Details

```
File: src/app/message.rs (line 723)
-    fn notify(&self, summary: &str, text: &str) {
-        if self.config.notifications.enabled
+    fn notify(&self, _summary: &str, _text: &str) {
+        // notify-rust disabled on OpenBSD (SIGSEGV on /proc/self/exe)
+        if false
```

### Why not a feature flag?

`notify-rust` is a **hard dependency** — it has no cargo feature to disable it. The patch is the only way to remove it on OpenBSD without forking the crate.

## Manual Build

```bash
PKG_CONFIG=echo \
  cargo install gurk \
    --locked \
    --path ~/src/gurk-rs \
    --root ~/.local
```

## Unlinking / Resetting

If you need to re-link your device (e.g. channel list stays empty):

```bash
rm -rf ~/.config/gurk ~/.local/share/gurk ~/.cache/gurk
~/.local/bin/gurk --relink
```

Note: Gurk does NOT remember your channels or messages on startup — it starts clean and only updates when you receive new messages. If the channel list stays empty, press `ctrl+p` to force the popup to open (often fixes the list), wait 30–60 seconds, or send yourself a test message from your phone.

## Dependencies (OpenBSD)

Required:
- `rust` (1.88.0+, tested with 1.94.1)
- `protoc` (protobuf compiler)
- `sqlite3`
- `openssl`

Build note: Use `PKG_CONFIG=echo` to bypass missing `pkg-config` on OpenBSD.

## Commands

### Navigation

| Action | Key |
|--------|-----|
| Open channel popup | `ctrl+p` |
| Move up/down channels | `ctrl+j` / `ctrl+k` or Up/Down |
| Move up/down in messages | `alt+Up` / `alt+Down` or `PgUp` / `PgDn` |
| Select a message | `PgUp` / `PgDn` |
| Quit | `ctrl+c` |

`ctrl+p` is your best friend — it opens the channel selection popup and often forces the channel list to load properly on startup.

### Messages

| Action | Key |
|--------|-----|
| Send message | `Enter` |
| Reply to selected message | Type reply + `Enter` |
| Open link in message | `Enter` (with empty input) |
| View attachment | `Enter` on selected message |
| Copy message to clipboard | `alt+y` |
| Edit selected message | `alt+e` |
| Copy selected message | `alt+y` |
| React with 👍 | `ctrl+t` |
| React with 🔥 | `ctrl+f` |
| React with 💜 | `ctrl+h` |
| Open full help screen | `f1` |

### Multi-line and Attachments

| Action | Key |
|--------|-----|
| Toggle multi-line input | `alt+Enter` |
| Send a file | `alt+Enter`, then type `file:///home/{user}/{path-and-file}` |
| Attach clipboard image | `alt+Enter`, then type `file://clip` (auto-pastes screenshot) |

### Emoji and Reactions

| Action | Key |
|--------|-----|
| Add reaction | Select message → type emoji → `tab` |
| Emoji shortcode | `:thumbsup:`, `:heart:`, `:rocket:`, etc. |
| Direct emoji | Just type the emoji (e.g. 🙂) |
| Paste emoji from clipboard | `ctrl+Shift+V` |

Emoji shortcodes use GitHub style — most standard ones work (`:thumbsup:`, `:+1:`, `:100:`, `:skull:`, etc.). Direct Unicode emoji works too.

### Input Editing

| Key | Action |
|-----|--------|
| `ctrl+a` / `Home` | Move cursor to beginning |
| `ctrl+e` / `End` | Move cursor to end |
| `ctrl+w` | Delete last word |
| `ctrl+u` | Delete to start of line |
| `alt+b` / `alt+Left` | Move back one word |
| `alt+f` / `alt+Right` | Move forward one word |
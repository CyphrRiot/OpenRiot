# Debugging Guide

## i3 Keybinding Issues

### Variable Expansion Bug
i3 has a bug where using `exec $variable` with a variable that has no additional arguments can fail silently.

**Example:**
```bash
# BROKEN - fails silently
set $menu $HOME/.local/share/openriot/install/openriot --rofi
bindsym $mod+D exec $menu

# WORKS - use direct path or add extra args
bindsym $mod+D exec $openriot_bin --rofi
```

**Diagnosis steps:**
1. Replace the command with something simple (e.g., `touch /tmp/test`)
2. If simple command works, the issue is the command itself
3. Try using direct path instead of variable
4. Try adding dummy arguments

### Config Validation
- **Always test with `i3 -C -c <file>`** before copying — catches syntax errors
- **Path expansion in `include`:** `~/.config/...` does NOT expand — use relative paths
- **Variable scope across includes:** `set $var` in `config` may not exist in included files — define where needed

## General Debugging Tips

### Verify Before Proposing
- Check if files/directories actually exist
- Use `ls` or `git status` to verify state
- Don't assume - verify

### Path Verification
- Verify character-by-character: `grendel` not `grandel`, `ProtonSync` not `ProtomSync`
- Typos waste time and break things

### Replace_all is Dangerous
- Double-check BEFORE using — can corrupt variable definitions
- Verify the file after replace_all

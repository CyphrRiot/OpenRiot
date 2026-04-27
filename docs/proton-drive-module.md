# Proton Drive Polybar Module - Design Document

## Overview

Add a 3-state polybar module for Proton Drive sync status. The icon reflects current state, and click actions are state-aware.

---

## States

| State | Icon | Meaning |
|-------|------|---------|
| Not Configured |  | ProtonDrive not set up |
| Needs Sync | 󰴋 | Files out of sync (or never synced) |
| Synced | 󱥾 | Local matches remote |

---

## CLI Flags

### `--polybar-proton-drive`
Polybar module - outputs icon for display.

```
$ openriot --polybar-proton-drive
󱥾
Proton Drive: Synced
```

Or for hidden state (not configured):
```
$ openriot --polybar-proton-drive
(nothing - module hidden)
```

### `--proton-drive-sync`
Click action - reads state, performs appropriate action.

| State | Action |
|-------|--------|
| Not Configured | Notify "Not configured\nSee OpenRiot.org for setup info" |
| Needs Sync | Open floating terminal with interactive sync prompt |
| Synced | Notify "Synchronized: <tooltip>" |

**Sync command (interactive terminal):**
```
alacritty --class openriot_upgrade -e sh -c "<interactive prompt>"
```

The prompt asks the user to choose:
- `[Y]` or `Enter` → bi-directional `rclone bisync` with `--resync --progress`
- `[O]` → one-way `rclone copy` with `--progress`
- `[Q]` or `[N]` → cancel

---

## Logic

```
isProtonDriveConfigured() = false → "not-configured"
isProtonDriveConfigured() = true:
  ├─ Check ~/.cache/rclone/bisync/home_{username}_Documents_ProtonSync..proton_ProtonSync.path1.lst
  ├─ Check ~/.cache/rclone/bisync/home_{username}_Documents_ProtonSync..proton_ProtonSync.path2.lst
  ├─ If either missing OR files differ → "needs-sync"
  └─ If both exist AND identical → "synced"
```

**Configuration check:**
```go
func isProtonDriveConfigured() bool {
    home := os.Getenv("HOME")
    
    // Check for ~/Documents/ProtonSync folder
    syncFolder := home + "/Documents/ProtonSync"
    if _, err := os.Stat(syncFolder); err != nil {
        return false
    }
    
    // Check for rclone config
    rcloneConf := home + "/.config/rclone/rclone.conf"
    if _, err := os.Stat(rcloneConf); err != nil {
        return false
    }
    
    return true
}
```

**Sync status check:**
```go
func checkProtonDriveSync() string {
    home := os.Getenv("HOME")
    bisyncDir := home + "/.cache/rclone/bisync"
    
    path1 := bisyncDir + "/home_{username}_Documents_ProtonSync..proton_ProtonSync.path1.lst"
    path2 := bisyncDir + "/home_{username}_Documents_ProtonSync..proton_ProtonSync.path2.lst"
    
    // Check both files exist
    if _, err := os.Stat(path1); os.IsNotExist(err) {
        return "needs-sync"
    }
    if _, err := os.Stat(path2); os.IsNotExist(err) {
        return "needs-sync"
    }
    
    // Compare files
    content1, _ := os.ReadFile(path1)
    content2, _ := os.ReadFile(path2)
    
    if string(content1) == string(content2) {
        return "synced"
    }
    return "needs-sync"
}
```

---

## Implementation Checklist

### Files to Modify

1. **`source/polybar/polybar.go`**
   - Update `isProtonDriveConfigured()` - keep existing
   - Add `checkProtonDriveSync()` function
   - Add `GetProtonDriveState()` - combines config check + sync check
   - Update `GetProtonDriveIcon()` - returns 󱥾/󰴋/empty based on state
   - Update `getProtonDriveTooltip()` - returns "Synced"/"Needs Sync"/"Not configured"
   - Update `RunProtonDrive()` - uses new state logic

2. **`source/main.go`**
   - Modify `--proton-drive-sync` flag
   - Add state detection before running sync
   - Implement state-aware click behavior

3. **`config/polybar/config.ini`**
   - Add `proton-drive` module entry

### Polybar Config

```ini
[module/proton-drive]
type = custom/script
exec = $HOME/.local/share/openriot/install/openriot --proton-drive
interval = 60
format-padding = 1
click-left = $HOME/.local/share/openriot/install/openriot --proton-drive-sync
```

---

## Notification Messages

| Trigger | Message |
|---------|---------|
| Not configured (click) | "Proton Drive is not configured" |
| Synced (click) | "Proton Drive is Synced" |
| Sync started | "Syncing Proton Drive..." |
| Sync complete | "Proton Drive sync complete" |
| Sync failed | "Proton Drive sync failed" |

---

## Edge Cases

1. **ProtonSync directory deleted after config**
   - State returns to "not-configured"

2. **rclone not installed**
   - Click on "Needs Sync" shows error notification

3. **bisync never run**
   - path1.lst missing → "needs-sync"

4. **Network offline**
   - "needs-sync" (bisync fails, cache not updated)

---

## Future Enhancements (Not This PR)

- [ ] Auto-sync every 30 minutes (background polling)
- [ ] Show file count differences in notification
- [ ] Conflict detection and resolution

# OpenRiot — File Audit Order

**Branch:** `i3` — X11/i3 migration
**Audit started:** Apr 9, 2026

---

## AUDIT ORDER

1. **Root Scripts** → 2. **config/bin/** → 3. **config/polybar/scripts/** → 4. **Backgrounds** → 5. **Source** → 6. **packages.yaml**

---

## 1. Root Scripts (5)

| # | File | Status |
|---|------|--------|
| 1.1 | `setup.sh` | ✅ DONE |
| 1.2 | `watcher.sh` | ✅ DONE |
| 1.3 | `build-iso.sh` | ✅ FIXED (Wayland→X11 comments) |
| 1.4 | `test-image.sh` | ✅ DONE |
| 1.5 | `test-iso.sh` | ✅ DONE |

---

## 2. config/bin/ Scripts (5)

| # | File | Status |
|---|------|--------|
| 2.1 | `battery-monitor.sh` | ✅ DONE |
| 2.2 | `openriot-lock.sh` | ✅ FIXED (hyprlock→i3lock) |
| 2.3 | `openriot-version-check` | ✅ DONE |
| 2.4 | `transmission-start` | ✅ DONE |
| 2.5 | `transmission-stop` | ✅ DONE |

---

## 3. config/polybar/scripts/ (3)

| # | File | Status |
|---|------|--------|
| 3.1 | `battery.sh` | ✅ DONE |
| 3.2 | `network.sh` | ✅ DONE |
| 3.3 | `openriot-update.sh` | ✅ DONE |

---

## 4. Backgrounds (16 images)

| Status |
|--------|
| ✅ SKIP (images only) |

---

## 5. wlsunset REPLACEMENT

| Item | Status |
|------|--------|
| wlsunset (Wayland-only) | ✅ DONE — replaced by `redshift-1.12p11` (X11) |

---

## 6. Source Code (19 Go files)

| # | File | Status |
|---|------|--------|
| 5.1 | `source/audio/volume.go` | ✅ DONE |
| 5.2 | `source/backgrounds/backgrounds.go` | ✅ DONE |
| 5.3 | `source/config/loader.go` | ✅ DONE |
| 5.4 | `source/config/types.go` | ✅ DONE |
| 5.5 | `source/crypto/crypto.go` | ✅ DONE |
| 5.6 | `source/crypto/trading.go` | ✅ DONE |
| 5.7 | `source/detect/detect.go` | ✅ DONE |
| 5.8 | `source/display/display.go` | ✅ DONE |
| 5.9 | `source/git/credentials.go` | ✅ DONE |
| 5.10 | `source/go.mod` | ✅ DONE |
| 5.11 | `source/go.sum` | ✅ DONE |
| 5.12 | `source/installer/colors.go` | ✅ DONE |
| 5.13 | `source/installer/configs.go` | ✅ DONE |
| 5.14 | `source/installer/execcommands.go` | ✅ DONE |
| 5.15 | `source/installer/packages.go` | ✅ DONE |
| 5.16 | `source/installer/sourcebuilds.go` | ✅ DONE |
| 5.17 | `source/logger/logger.go` | ✅ DONE |
| 5.18 | `source/main.go` | ✅ DONE |
| 5.19 | `source/notify/notify.go` | ✅ DONE |
| 5.20 | `source/tui/messages.go` | ✅ DONE |
| 5.21 | `source/tui/model.go` | ✅ DONE |
| 5.22 | `source/polybar/polybar.go` | ✅ DONE |

### wlsunset (9 files)

| # | File | Status |
|---|------|--------|
| 5.23 | `source/wlsunset/` | ✅ DONE (DELETED — replaced by redshift) |

---

## 6. packages.yaml

| Status |
|--------|
| ✅ DONE |

---

## LEGEND

- ✅ DONE — Audited, no changes needed
- ✅ FIXED — Audited and fixed
- 🔄 NEXT — Currently reviewing
- 🔄 HERE — Currently being fixed
- ⏳ — Not yet reviewed
- ✅ SKIP — No audit needed (images/binary)

---

## POST-AUDIT TASKS

| Step | Task | Status |
|------|------|--------|
| 1 | Add scrot package for X11 screen capture | ✅ DONE |
| 2 | Add age package for file encryption | ✅ DONE |

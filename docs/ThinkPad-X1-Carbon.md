# ThinkPad X1 Carbon Aura Gen 13 — OpenBSD Support Proposal

## Executive Summary

The ThinkPad X1 Carbon Aura Gen 13 (Intel Arrow Lake-U / Core Ultra 7 255U) has **partial but usable support** on OpenBSD -current as of late 2025. Based on actual testing (dmesg from Dec 30, 2025), core functionality works (boot, CPU, NVMe, audio, input), while key blockers remain (Wi-Fi init failure, GPU GuC firmware issues, Thunderbolt unattached). This document outlines the targeted engineering work to fix remaining issues, the testing methodology, and the contribution workflow for upstreaming changes to OpenBSD.

## Hardware Inventory

Based on actual OpenBSD -current testing (Dec 30, 2025 dmesg, machine type 21NXCTO1WW, BIOS N4MET26W):

| Component | Chipset | OpenBSD Status |
|-----------|---------|----------------|
| CPU | Intel Core Ultra 7 255U (Arrow Lake-U) | **Working** — `acpicpu(4)` attaches, C-states, PSS, SpeedStep |
| GPU | Intel Graphics (detected as Meteor Lake gen 12) | **Partial** — modesetting works (1920x1200), GuC firmware init fails ("GPU wedged"), no acceleration |
| Wi-Fi | Detected as Intel Wi-Fi 6e AX211 (likely BE-series) | **Broken** — "could not initialize hardware", clock stabilization timeout, APM init errors |
| Audio | Intel Core Ultra HD Audio, Realtek ALC287 | **Working** — `azalia(4)` attaches, `audio0` created |
| NVMe | SK hynix drive | **Working** — attaches as `sd0` |
| Input | Synaptics SYNA802E (I2C HID) | **Working** — `ihidev(4)`, clickpad, multi-touch, trackpoint |
| Keyboard | Standard ThinkPad | **Working** — `pckbd(4)`, `acpithinkpad(4)` present |
| Ethernet | — | Not listed in dmesg (likely via USB-C dock) |
| Thunderbolt/USB4 | USB4 / Thunderbolt bridges | **Partial** — "not configured" in dmesg |
| Webcam | — | Not listed (likely MIPI, unsupported) |
| Sensors | `acpiec(4)`, ThinkPad EC | **Partial** — basic EC present, thermal/battery functional |
| NPU/IPU/GNA | Intel AI accelerators | **Unsupported** — "not configured" (expected) |
| TPM | TPM 2.0 | **Working** |
| Power | S0ix/S4/S5 ACPI states | **Partial** — states recognized, real-world S0ix reliability untested |
| WWAN | Quectel modem (optional) | **Untested** — present but not configured |

## Actual Test Results (OpenBSD 7.8-current, Dec 30 2025)

**Source:** dmesgd.nycbug.org — Lenovo ThinkPad X1 Carbon Gen 13 (21NXCTO1WW, BIOS N4MET26W)

**What Works:**
- Boot to multiuser, console at 1920x1200
- CPU: Core Ultra 7 255U, 12C/14T, SpeedStep, C-states, PSS
- Storage: SK hynix NVMe (`sd0`)
- Audio: `azalia(4)` + Realtek ALC287, `audio0` created
- Input: Synaptics SYNA802E I2C trackpad (clickpad, multi-touch), keyboard, trackpoint
- TPM 2.0, `acpithinkpad(4)`, `acpiec(4)`
- USB (`xhci(4)`)

**What Partially Works:**
- Graphics: Modesetting works, GuC firmware fails ("GPU wedged"), no acceleration
- Power: S0ix states recognized, real-world testing needed
- Thunderbolt/USB4: Present but "not configured"

**What Fails:**
- Wi-Fi: Detected as AX211, "could not initialize hardware", clock stabilization timeout
- Webcam: Likely MIPI, not detected
- NPU/IPU/GNA: "not configured" (expected)

**Recommended Workarounds:**
- Wi-Fi: Use USB Wi-Fi dongle or USB-C Ethernet dock
- Webcam: Use USB webcam

## Driver Porting Tasks

### 1. `inteldrm` — Fix GuC Firmware and GPU Acceleration

**Status:** GPU detected as Meteor Lake gen 12, modesetting works, but GuC firmware initialization fails ("GPU wedged" taint).

**Scope:** Fix GuC firmware loading and enable GPU acceleration for Arrow Lake-U / Meteor Lake graphics.

**Tasks:**
- Debug GuC firmware init failure (check firmware version in `inteldrm`, compare with Linux)
- Update or add correct GuC firmware files for Meteor Lake/Arrow Lake
- Fix GuC handshake and initialization sequence
- Enable accelerated X11/Wayland support (blitter, rendering)
- Test: `glxinfo`, suspend/resume, brightness control, external display

**Constraints:** ISC license only. No GPL code. Study Linux `i915` GuC code for register definitions, rewrite independently.

### 2. `iwx` — Fix Wi-Fi Initialization (AX211/BE200)

**Status:** Card detected as "Intel Wi-Fi 6e AX211" but fails with clock stabilization timeout and APM init errors. Actual chip likely BE-series (Wi-Fi 7).

**Scope:** Fix firmware loading and initialization for the BE200/AX211 variant in this platform.

**Tasks:**
- Identify exact chip via `pcidump -v` (likely `0x272b` BE200 or `0x7af0` AX211)
- Debug clock stabilization timeout (check firmware file, LAR disable, etc.)
- Add/fix firmware definition (`iwx-be200-*.ucode` or correct AX211 firmware)
- Test: `ifconfig iwx0 scan`, association, 2.4/5 GHz, throughput

**Workaround:** Use supported USB Wi-Fi dongle until fixed.

### 3. Power Management — S0ix Testing and Tuning

**Status:** ACPI reports S0ix/S4/S5 states, `intelpmc(4)` present. Real-world S0ix reliability untested.

**Scope:** Test and tune modern standby, validate battery life and thermals.

**Tasks:**
- Test `zzz` (suspend) and `ZZZ` (hibernate) — collect success/failure logs
- Monitor battery life under load vs. idle (compare with Linux)
- Tune C-states and P-states if needed (`sysctl hw.setperf`, `apm`)
- Port any missing Meteor Lake PCH power registers if S0ix fails
- Test: wake from keyboard, network, lid open

### 4. Sensors — Validate and Enhance

**Status:** `acpiec(4)` present, basic EC/sensors functional. ThinkPad-specific sensors may need tuning.

**Scope:** Ensure thermal throttling, battery monitoring, and fan control work correctly.

**Tasks:**
- Validate `sysctl hw.sensors` output (battery %, temp, fan RPM)
- Test thermal throttling under load (`apm -l`, `sysctl hw.sensors`)
- Add any missing ThinkPad-specific sensor hooks (`acpithinkpad(4)`)
- Test: fan behavior, critical shutdown temps

### 5. Audio — Already Working

**Status:** `azalia(4)` attaches with Realtek ALC287 codec, `audio0` created. No SOF port needed.

**No action required.** If audio issues arise later, debug `azalia` not SOF.

### 6. Thunderbolt/USB4 — Attach Missing Driver

**Status:** "not configured" in dmesg. May need PCI ID additions or new driver.

**Tasks:**
- Run `pcidump -v` to identify exact Thunderbolt/USB4 controller
- Add PCI ID to appropriate driver (likely `pcib(4)` or dedicated Thunderbolt driver)
- Test: Thunderbolt device hotplug, USB4 peripherals
- Note: May require BIOS setting changes (disable "Security Level" if present)

### 7. Webcam — MIPI Camera Support

**Status:** Webcam likely MIPI (not USB UVC), unsupported by OpenBSD.

**Tasks:**
- Confirm webcam type via `usbdevs` or `pcidump`
- If USB: should work with `uvideo(4)`
- If MIPI: Major undertaking (no MIPI stack in OpenBSD), use USB webcam as workaround

### 8. PCI ID Updates — NVMe, Ethernet

**Status:** NVMe works (SK hynix). Ethernet likely via USB-C dock (not built-in).

**Tasks:**
- If Ethernet not detected: `pcidump -v`, add PCI ID to `em(4)` or `igc(4)`
- NVMe: Should work, no action needed unless different controller found

## Testing Methodology

### Phase 0: Review Existing Data

Before testing, review the existing dmesg at dmesgd.nycbug.org (Lenovo X1 Carbon Gen 13, Dec 30 2025, BIOS N4MET26W). Compare your hardware (Aura Edition may have OLED, different Wi-Fi card, etc.).

### Phase 1: Baseline Collection

On the target machine (boot **latest** OpenBSD -current snapshot):
```sh
dmesg > dmesg.txt
pcidump -v > pcidump.txt
sysctl hw > sysctl-hw.txt
acpidump -dt > acpi.aml
usbdevs -v > usbdevs.txt
```

Identify:
- Which devices fail to attach (compare with existing dmesg)
- Which attach but don't function (e.g., Wi-Fi init failure)
- Which work but lack features (power mgmt, acceleration, etc.)
- Aura Edition specifics (OLED display behavior, etc.)

### Phase 2: Incremental Driver Porting

1. **One driver at a time** — don't port `inteldrm` and `iwx` simultaneously
2. **Start with easiest** — PCI ID additions before complex logic ports
3. **Test after each change** — `make` in kernel build dir, reboot, verify
4. **Use `ddb`** — kernel debugger for crash analysis (`boot -d`, `trace`, `ps`)

### Phase 3: Regression Testing

After each driver works:
- Run `make regress` in `/usr/src/regress`
- Test on other Intel hardware to ensure no regressions
- Document any new bugs introduced

## Contribution Workflow

### 1. Development Environment

```sh
# OpenBSD -current system
cd /usr/src
cvs up -PAd
# Create branch for each driver
mkdir -p ~/patches/meteor-lake
```

### 2. Patch Submission

OpenBSD uses `tech@openbsd.org` mailing list for code review:

1. **Format patches** with `cvs diff -uNp`
2. **Write clear commit messages** — describe *why*, not *what*
3. **Include test results** — dmesg before/after, functional tests
4. **Be responsive** — address reviewer feedback promptly

Example patch header:
```
Add Intel Meteor Lake GPU support to inteldrm(4)

Tested on ThinkPad X1 Carbon Aura Gen 13 (Core Ultra 7 155H).
Display works at 1920x1200, suspend/resume functional.
X11 acceleration untested (TODO: implement blitter).
```

### 3. Licensing Requirements

- All code must be ISC license or compatible
- No GPL-licensed code or direct translations
- Study Linux/FreeBSD drivers for register definitions only
- Rewrite logic independently — clean-room implementation

### 4. Timeline Estimates

Based on actual hardware status (audio works, GPU partial, Wi-Fi broken):

| Task | Effort | Dependencies |
|------|--------|--------------|
| `iwx` Wi-Fi fix (firmware/clock) | 2-4 weeks | Firmware availability |
| `inteldrm` GuC fix + acceleration | 1-3 months | Firmware debugging |
| Thunderbolt/USB4 attach | 1-2 weeks | PCI ID addition |
| S0ix testing/tuning | 2-4 weeks | None (test existing support) |
| Sensors validation | 1 week | None (mostly works) |
| Audio | **Done** | N/A |
| Webcam (MIPI) | 6-12 months | Major undertaking (skip) |
| Full integration | 2 weeks | All drivers |

**Minimal working system (Wi-Fi + GPU accel):** 2-4 months part-time
**Full feature parity (excluding MIPI webcam):** 4-6 months part-time

## For Non-Developers

If you cannot write kernel code:

1. **Sponsor a developer** — contact OpenBSD developers willing to take contracts
2. **Provide hardware** — ship the laptop to a developer for testing
3. **Test snapshots** — regularly test -current and report `dmesg` diffs
4. **Document findings** — maintain a test matrix of what works/fails
5. **Financial support** — donate to OpenBSD Foundation earmarked for laptop support

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| GuC firmware init failure unfixable | High | Fallback to unaccelerated modesetting (currently works) |
| Wi-Fi firmware/clock issue persistent | Medium | Use USB Wi-Fi dongle or USB-C Ethernet |
| S0ix unreliable on Arrow Lake-U | Medium | Test thoroughly, use S3 if available in BIOS |
| Thunderbolt needs new driver (not just PCI ID) | Medium | USB-C dock for Ethernet/display |
| MIPI webcam (no OpenBSD support) | Low | Use USB webcam |
| OLED display issues (Aura Edition) | Medium | Test with LCD model first, report OLED-specific bugs |
| No developer interest | High | Financial sponsorship, hardware donation |

## Next Steps

1. **Review existing dmesg** at dmesgd.nycbug.org (search "X1 Carbon Gen 13")
2. **Boot latest -current snapshot**, collect `dmesg`, `pcidump -v`, compare with existing reports
3. **Identify top 3 blocking issues**: Wi-Fi init failure, GPU GuC firmware, Thunderbolt attach
4. **Start with Wi-Fi fix** (highest impact, moderate effort) — debug firmware loading
5. **Set up OpenBSD -current build environment** (`cvs up -PAd`)
6. **Begin with `iwx` debugging** (firmware files, clock stabilization)
7. **Test S0ix** (`zzz`/`ZZZ`) and document results
8. **Submit findings** to dmesgd.nycbug.org and `misc@openbsd.org`

## References

- **Existing dmesg**: dmesgd.nycbug.org (search "ThinkPad X1 Carbon Gen 13")
- OpenBSD `inteldrm` source: `/usr/src/sys/dev/pci/drm/i915/`
- OpenBSD `iwx` source: `/usr/src/sys/dev/pci/if_iwx.c`
- Linux `i915` driver: `drivers/gpu/drm/i915/` (for register definitions, clean-room only)
- Linux `iwlwifi` driver: `drivers/net/wireless/intel/iwlwifi/` (for Wi-Fi firmware reference)
- OpenBSD man pages: `inteldrm(4)`, `iwx(4)`, `azalia(4)`, `acpisleep(4)`
- Mailing lists: `tech@openbsd.org`, `misc@openbsd.org` (subscribe via `majordomo@openbsd.org`)
- OpenBSD laptop compatibility: jcs.org/openbsd-laptops
- NYC*BUG dmesg collection: dmesgd.nycbug.org

---

*This proposal is a living document. Update as hardware testing progresses and drivers are implemented.*

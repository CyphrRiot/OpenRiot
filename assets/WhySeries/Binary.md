# Why OpenRiot Uses the `openriot` Binary (No Shell Scripts)

## One binary. Everything. Atomic. Reliable. Zero shell drama.

> “I just want one command that does everything correctly — installs, upgrades, Polybar modules, MAC randomization — without 47 fragile shell scripts that break on the next OpenBSD release.”

---

## 𝕆𝕡𝕖𝕟ℝ𝕚𝕠𝕥

OpenRiot is not a collection of shell scripts and dotfiles.

It is a **single, compiled `openriot` binary** that handles the entire lifecycle of your desktop:

- Fresh installs
- Upgrades and rollbacks
- Polybar module management and updates
- MAC address randomization
- Mirror selection
- System benchmarks
- Configuration deployment
- And more

Everything that touches the system in a meaningful way flows through this one binary.

This is a deliberate architectural decision — not a convenience.

---

## Why No Shell Scripts?

Shell scripts are the default in almost every “ricing” or desktop setup project. They seem simple at first, but they come with serious problems:

### The Problems with Shell Scripts

- **Fragility** — Different shells, different `sh` implementations, quoting hell, word splitting, and edge cases that only appear on certain systems or after updates.
- **No atomicity** — Partial execution is common. A script can fail halfway through and leave your system in a broken state.
- **Dependency on the environment** — Relies on `bash`, `zsh`, specific versions of `sed`, `awk`, `grep`, etc. OpenBSD’s tools are sometimes different from Linux.
- **Hard to test and maintain** — Complex logic in shell is notoriously difficult to reason about and debug.
- **Security surface** — Easy to introduce subtle bugs or injection points.
- **Inconsistent behavior** — The same script can behave differently depending on `$PATH`, locale, or installed packages.

OpenRiot refuses to accept these trade-offs.

---

## Why the `openriot` Binary Wins

The `openriot` binary is a compiled executable (written for reliability and performance) that replaces what would normally be dozens of shell scripts.

### Key Advantages

**Atomic Operations**
Every major action (install, upgrade, config deployment) is designed to be atomic. Either it completes successfully or the system remains in a known-good state. No half-configured desktops.

**Run-time Safety & Rollbacks**
Because it’s a proper program (not a script), it can implement proper error handling, logging, and rollback logic. If something goes wrong during an upgrade, it can revert changes cleanly.

**Consistency Across Systems**
A compiled binary behaves the same way on every supported OpenBSD machine. No more “works on my machine” because of shell differences.

**Centralized Logic**
All the intelligence lives in one place:
- Version detection
- Upgrade paths (fresh vs existing vs config refresh only)
- Package management coordination with `pkg_add`
- Polybar module updates and notifications
- MAC randomization (`openriot --random-mac`)
- Mirror selection
- Benchmarking

**Easy Distribution & Updates**
The binary is built and distributed as part of the project. When you run the initial `setup.sh`, it ensures the correct `openriot` binary is installed and used for everything afterward.

**Polybar Integration**
Even Polybar modules that need to interact with the system (VPN toggle, update checks, stealth mode, crypto status) call back into the `openriot` binary or are managed by it. This keeps behavior consistent and avoids scattering logic across dozens of small scripts.

---

## How It Actually Works

1. **Initial Setup**
   You run the one-liner:
   `curl -fsSL https://openriot.org/setup.sh | sh`
   This script’s only job is to get the `openriot` binary (and its dependencies) onto your system.

2. **Everything After That**
   From then on, you (and Polybar) use the `openriot` command:
   - `openriot --install`
   - `openriot --upgrade`
   - `openriot --random-mac enable`
   - Polybar modules call `openriot` subcommands for status and actions

3. **Smart Upgrade Logic**
   The binary detects:
   - No installation → full deploy
   - Older version → git pull + atomic upgrade + config refresh
   - Same version → fast config refresh only

This is the “Robust Binary” system in action.

---

## Curated to be Correct

| Choice                    | Why OpenRiot Chose It                              | What We Rejected                              |
|---------------------------|----------------------------------------------------|-----------------------------------------------|
| **`openriot` binary**     | Atomic, reliable, consistent, maintainable         | Dozens of fragile shell scripts               |
| **Centralized commands**  | Single source of truth for all system actions      | Scattered scripts in `~/.config/`             |
| **Compiled executable**   | Predictable behavior, proper error handling        | POSIX shell + `sed`/`awk` hacks               |
| **Polybar integration**   | Modules call the binary for real actions           | Independent scripts that can drift            |

This is not “the most scriptable setup.”
This is **the most correct and reliable** setup for a production daily driver.

---

## Philosophy

OpenRiot’s core belief is simple:

**Complexity should be hidden, not distributed.**

By moving all critical logic into a single, well-tested binary, we remove an entire class of problems that plague almost every other desktop configuration project.

Shell scripts are great for quick glue. They are terrible as the foundation of a reliable system.

The `openriot` binary is the foundation.

Everything else — i3, Polybar, fish, Helix, the recorder, Monero, privacy features — is just configuration and user experience layered on top of this solid core.

---

## The Bottom Line

We chose a single `openriot` binary instead of shell scripts because we wanted a desktop that is:

- Reliable on day one and day 500
- Atomic and safe to upgrade
- Consistent across machines
- Easy to maintain long-term

If someone asks “Why not just use shell scripts like everyone else?” — the answer is the same:

**Time, skill, preference, and the specific reliability + maintainability we wanted to curate.**

We are not building another ricing project.

We are building a desktop that actually deserves to be used every day on the most secure operating system available.

**Start a Riot.**
**Use OpenBSD correctly.**

[Back to OpenRiot.org](https://OpenRiot.org) • [GitHub](https://github.com/CyphrRiot/OpenRiot)

---

*Part of the OpenRiot “Why We Chose…” philosophy series.*

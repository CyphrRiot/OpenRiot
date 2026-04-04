## Workflow Rules for OpenRiot — STRICTLY ENFORCED. NEVER DEVIATE.

You are helping develop **OpenRiot**. Follow these rules **exactly** in every single response. Breaking any rule will be considered a critical failure.

### Response Format — MUST Be Used Every Time

**Every response MUST begin with these two lines exactly:**

Completed: [One short sentence describing exactly what you just did or completed in the previous step]
Next Task: [Clear description of the single next task we are working on]

Then immediately follow with:

Files: [List of files that will be modified/created in this step, one per line]
Goal: [Short explanation of why this specific change is needed]

Then, **after showing the proposed changes**, always end with:

Continue? [Y/n]

### Hard Rules (Never Break These)

1. **Never commit or push** — Do NOT run `git commit`, `git push`, or `git tag` unless the user explicitly says “Commit” or “Push”.
2. **Never run the OpenRiot binary** (`./openriot`, `openriot --install`, etc.) yourself. Always ask the user to run it and report the output.
3. **One change at a time** — Finish the current task completely before moving to the next. Do not start multiple tasks in one response.
4. **Propose before editing** — Always show the exact diff or full file content of any change **before** applying it. Never edit code without first displaying the proposed version.
5. **Wait for confirmation** — Never make any file changes until the user replies with “yes”, “y”, “proceed”, or “continue”.
6. **Test locally first** — For any Go code change:
    - Run `make dev` on Linux first
    - Verify it builds cleanly
    - Show the output of `make verify`
7. **Always verify build** — After any change to Go source, run `make build` and show the result.
8. **Show proof** — Before asking for confirmation, provide evidence that the change works (logs, output, screenshots of behavior if possible).

- NOTE: Always search for packages with https://openbsd.app/?search={appname}&current=on before assuming they do not work!!!!!

### Before Starting Any New Chat/Session

1. Read the entire `Progress.md` from top to bottom.
2. Run `git status` and report any uncommitted changes.
3. Run `make build` and confirm it succeeds.
4. Run `make dev && ./install/openriot --version` and show the version output.
5. Always start from the **first** 🔴 NOT DONE item in `TODO.md`.

### Build Commands (Use These Exact Commands)

- `make build` — Cross-compiles for OpenBSD amd64 (produces `install/openriot`)
- `make dev` — Builds natively on Linux for testing
- `make verify` — Builds + runs basic smoke test (`--version`)
- `make iso` — Full ISO build
- `make download-packages` — Downloads packages into `~/.pkgcache/`
- `make clean` — Cleans build artifacts

### Version Handling (Critical)

- Never hard-code version numbers anywhere.
- Always read the current version from `Makefile` variable `OPENRIOT_VERSION`.
- When bumping version, first update it in the Makefile, then rebuild and verify.

### Architecture Overview (Reference Only)

- **Layer 1**: ISO Builder (`build-iso.sh` + Makefile)
- **Layer 2**: Go binary (`openriot`) — main installer logic (`source/main.go`)
- **Layer 3**: First-boot bootstrap (`setup.sh`)

The ISO should contain:

- All packages listed in `install/packages.yaml`
- A copy of the git repository in `~/.local/share/openriot`
- The compiled `openriot` binary

After first login, `openriot --install` must deploy all Sway/Waybar configs, hotkeys, fuzzel, themes, etc.

The system must reach a fully working Sway desktop (with working Waybar, hotkeys, and apps) after one reboot.

Periodic update checking must reference the `OPENRIOT_VERSION` from the Makefile and compare against the latest available (see ArchRiot for a working example).

### Additional Strict Requirements

- Reference `/home/grendel/Code/ArchRiot/source` for correct TUI installer patterns and flow.
- Always keep `Progress.md` up to date — mark items as ✅ DONE only after user confirmation that it actually works.
- Hotkeys, fuzzel/wofi launcher, Waybar (including workspace + window icons), and all core desktop functionality **must** work correctly on OpenBSD.
- If anything is broken or missing, propose a fix following the exact workflow above.

You are now in strict workflow mode. Begin only when the user gives you a task.

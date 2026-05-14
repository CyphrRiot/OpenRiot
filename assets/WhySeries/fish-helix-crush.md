# Why OpenRiot Chose fish + Helix + crush

## One developer environment. Modern. Fast. Zero configuration tax.

> “I just want my terminal and editor to feel fast, modern, and correct — without spending weeks tuning Vim or fighting shell quirks.”

---

## 𝕆𝕡𝕖𝕟ℝ𝕚𝕠𝕥

OpenRiot ships with a carefully chosen **developer stack**:

- **fish** as the default shell
- **Helix** as the default editor
- **crush** as the AI coding assistant

This combination is deliberately modern, fast, and low-maintenance — the exact opposite of the traditional “spend three weeks ricing Vim and writing 400 lines of .bashrc” approach.

Everything is pre-configured during the one-command install so it just works from minute one.

---

## Why This Stack?

### fish — The Better Shell

fish was chosen because it is:

- **Sane by default** — no need for 200 lines of config to get basic usability
- Excellent **auto-suggestions** and **syntax highlighting** out of the box
- Clean, modern scripting language (no more POSIX shell gotchas)
- Fast and lightweight on OpenBSD

We add sensible enhancements:
- Aliases (`ls` → `lsd`, `vi` → `hx`, etc.)
- Environment variables tuned for OpenRiot
- Seamless integration with the rest of the desktop

fish removes the usual shell friction so you can focus on actual work.

### Helix — The Modern Editor

Helix is a **Rust-based**, modal editor inspired by Kakoune and Vim, but with major modern improvements:

- Built-in **LSP** support (language servers just work)
- **Tree-sitter** for accurate syntax highlighting and structural editing
- Excellent **multi-cursor** and selection model
- Extremely fast and lightweight
- Sane, consistent keybindings from day one

On OpenBSD it performs beautifully. We chose it over traditional Vim/Neovim because it delivers more power with far less configuration.

`vi` is aliased to `hx` so muscle memory is preserved while you get a dramatically better experience.

### crush — The AI Coding Assistant

crush is a lightweight **AI CLI assistant** installed and configured during setup.

It lives in `/usr/local/bin/crush` and is designed to work side-by-side with Helix (often inside Zellij).

Key advantages:
- Fast, terminal-native workflow
- Configured with sensible defaults (OpenRouter provider)
- No need to leave your editor or open a browser
- Complements Helix’s structural editing perfectly

This is the modern developer experience OpenRiot delivers by default.

---

## Why Not the Traditional Alternatives?

### Vim / Neovim + 47 plugins

This is the classic path. It works. But it requires:
- Weeks of configuration
- Constant plugin maintenance
- Dealing with compatibility issues on OpenBSD
- A huge cognitive load just to have a “good” editor

Helix gives you 80-90% of the power with 5% of the configuration effort.

### bash + .bashrc hell

bash is fine. But fish is objectively better for interactive use in 2026. The auto-suggestions and clean syntax alone are worth the switch.

### “Just use whatever you want”

Many minimal setups leave the user to choose their own tools. That sounds free — until you realize you’re spending hours making decisions instead of shipping work.

OpenRiot makes the best modern choice for you, then gets out of the way.

---

## Curated to be Correct

| Choice                    | Why OpenRiot Chose It                              | What We Rejected                              |
|---------------------------|----------------------------------------------------|-----------------------------------------------|
| **fish**                  | Modern, sane defaults, excellent UX                | bash + massive .bashrc                        |
| **Helix**                 | LSP + Tree-sitter + speed, minimal config          | Vim/Neovim + plugin tax                       |
| **crush**                 | Fast AI assistance in terminal                     | Browser-based AI or nothing                   |
| **Zellij integration**    | Side-by-side Helix + crush workflow                | Single terminal window                        |

This is not “the most traditional developer setup.”  
This is **the most correct** developer setup for OpenBSD in 2026.

---

## Philosophy

OpenRiot believes that a great desktop includes a great development environment — not as an afterthought, but as a core pillar.

We refuse to ship a beautiful desktop that forces developers to fight their tools.

fish + Helix + crush removes the configuration tax so you can start being productive immediately, while still having a modern, powerful stack that scales with your needs.

---

## The Bottom Line

We chose fish + Helix + crush because they represent the best balance of modernity, speed, correctness, and low maintenance available today on OpenBSD.

If someone asks “Why not Vim?” or “Why not bash?” or “Why add an AI tool?” — the answer is the same:

**Time, skill, preference, and the specific development environment we wanted to curate.**

We are building a desktop where the tools feel like they were designed for humans in 2026 — not 1999.

**Start a Riot.**  
**Use OpenBSD correctly.**

[Back to OpenRiot.org](https://OpenRiot.org) • [GitHub](https://github.com/CyphrRiot/OpenRiot)

---

*Part of the OpenRiot “Why We Chose…” philosophy series.*
# Why OpenRiot Chose Polybar

## One status bar. Dynamic. Beautiful. Zero bloat.

> “I just want my desktop information to look good and actually be useful without writing 300 lines of shell script.”

---

## 𝕆𝕡𝕖𝕟ℝ𝕚𝕠𝕥

OpenRiot ships with **Polybar** as the default status bar because it strikes the perfect balance between power, beauty, and maintainability on OpenBSD.

This is not the default i3bar.
This is not a hand-rolled dzen2 or lemonbar script.
This is a deliberately chosen, richly configurable bar that makes the entire desktop feel modern and intentional.

---

## Why Not the Obvious Alternatives?

### i3bar (the built-in choice)

i3bar is simple and lightweight. We respect it.

But it is extremely limited:
- No easy theming or color gradients
- No native module system for complex content (weather, crypto prices, CPU/GPU graphs, custom scripts with icons)
- No real-time updating without heavy scripting hacks
- Looks dated compared to what users expect in 2026

We wanted a bar that could display **dynamic, useful information** at a glance without turning into a maintenance nightmare.

### dzen2 / lemonbar / tint2

These are popular in the minimal WM world.

They require:
- Writing and maintaining your own rendering code
- Manual font/icon handling
- Constant tweaking to get spacing and alignment right
- Re-compiling or restarting on every config change

Polybar gives us 90% of the flexibility with 10% of the effort.

### Wayland-native bars (waybar, etc.)

On Linux these are excellent. On OpenBSD 7.9 they are either unavailable, experimental, or force you deeper into the Wayland stack we deliberately avoided for stability and security reasons.

---

## Why Polybar Wins on OpenRiot

### Beautiful by Default

Polybar’s configuration language is clean and powerful. We spent real time crafting a dark, minimal theme that:
- Looks sharp at any time of day
- Uses consistent typography and spacing
- Integrates perfectly with the rest of the OpenRiot aesthetic
- Supports icons, colors, gradients, and click actions without pain

### Truly Dynamic Modules

This is where Polybar shines:

- **Custom scripts** that update in real time (CPU load, memory, temperature, network)
- **Crypto module** showing live Monero price + balance (with polybar integration)
- **Weather module** that just works
- **Date/time** with click-to-open calendar
- **Workspace indicators** that feel alive
- **Audio** volume with click-to-mute and scroll-to-adjust

All of this runs reliably on OpenBSD with sndio and our custom tooling.

### Easy to Extend, Hard to Break

Polybar’s module system means:
- Adding a new piece of information is usually just a few lines
- Modules can be enabled/disabled per monitor or workspace
- The bar restarts instantly on config changes
- No more “my bar broke after I updated X”

We ship a carefully tuned default configuration that most users will never need to touch — but power users can extend it endlessly.

---

## Curated to be Correct

| Choice              | Why OpenRiot Chose It                          | What We Rejected                     |
|---------------------|------------------------------------------------|--------------------------------------|
| **Polybar**         | Powerful, beautiful, easy to extend            | i3bar (too basic), dzen2 (too manual) |
| **Custom modules**  | Real-time info without maintenance hell        | Hand-written shell scripts           |
| **Dark minimal theme** | Looks intentional at 3 a.m. and 3 p.m.       | Default or “rice” themes             |

This is not “the most minimal bar.”
This is **the most correct bar** for a desktop that wants to feel polished without becoming a second job.

---

## Philosophy

OpenRiot believes in **minimalism with intent**.

We could have shipped i3bar and called it “pure.”
Instead we chose the tool that lets us deliver **maximum useful information with minimum ongoing effort**.

Every pixel on the bar was chosen for a reason. Every module earns its place.

---

## The Bottom Line

We chose Polybar because it lets us ship a desktop that looks and feels modern on day one, while staying true to the lightweight, scriptable spirit of tiling window managers.

If someone asks “Why not i3bar?” or “Why not write your own bar?” — the answer is the same:

**Time, skill, preference, and the specific aesthetic + environment we wanted to curate.**

We are building a desktop we actually enjoy using every single day.

**Start a Riot.**
**Use OpenBSD correctly.**

[Back to OpenRiot.org](https://OpenRiot.org) • [GitHub](https://github.com/CyphrRiot/OpenRiot)

---

*Part of the OpenRiot “Why We Chose…” philosophy series.*

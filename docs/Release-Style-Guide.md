# OpenRiot Release Notes — Procedure & Style Guide

## 1. Determine Version

Read `VERSION` to find the latest release (e.g. `7.9.44`). Create
`docs/v7.9.XX-Release-Notes.md` where `XX` is the next increment
(e.g. `45`). Do **not** edit `VERSION` or any README badge.

## 2. Discover What Changed

Ground notes in real changes. Non-negotiable.

- Review commits since the last release tag: `git log --oneline <last-tag>..HEAD`
- Inspect modified files in `scripts/`, `config/`, `source/`, `install/`,
  polybar modules, i3 config, window management, backgrounds, lock screen, etc.
- Describe *user-visible* impact, not the diff. Minor/polish changes are
  fine — lean into atmosphere rather than forcing drama.
- Never invent features.

## 3. Study the Voice

Read several release notes across the archive (not just the last two) to
internalize the range: epic (`v7.9.36`), deadpan (`v7.9.20`), atmospheric
(`v7.9.33`).

### Structure

- Atmospheric opening quote from "The OpenRiot Crew"
- "Release Overview" — thematic summary
- Several short narrative sections with dramatic headings
- Closing quote that echoes the opener
- Files Changed table near the end

### Motifs

"The construct", "the machine", neon wastes / the sprawl, persistence,
awakening, dead channels lighting up, guards/ICE/Turing Police as
antagonistic forces.

### Tone

Defiant, irreverent, triumphant. Romantic about OpenBSD's purity vs. Linux
churn and corporate nonsense. Accessible to normal OpenBSD users.

### Anti-Patterns

- No generic "improved performance and fixed bugs" energy
- No bullet-point changelogs
- Every change gets story treatment
- Technical details appear in service of the story ("we killed the guards
  with `CFLAGS=...`")

### The Fed Jab

Every release notes must include at least one natural, lore-appropriate
jab at the Feds/Turing Police/surveillance state. It must feel earned and
in-universe, not tacked on. Examples: "The Feds, as always, are bad at
their jobs", framing them as incompetent or irrelevant to sovereign
systems.

## 4. Write

- Strong, evocative title (e.g. "The Turing Police Are Bad at Their Jobs",
  "A War Is Coming", "The Dead Channels Remember")
- Narrative prose — changes are events in the life of "the construct"
- Clean Files Changed table near the end
- Explain *why it matters* more than *how it was implemented*
- Write so someone wants to read the whole thing

## 5. Update README.md TOC

Find the Release Notes section in `README.md`. Add the new entry at the
top, following the existing link pattern. Keep only the 10 most recent;
add `[Older releases →](docs/)` as the 11th line. Do not change version
numbers, badges, or anything else.

## 6. Review README.md

Verify README.md is still accurate for the current code — menus, icons,
app launcher tables, etc. If changes are needed, **propose first**, wait
for confirmation, then edit.

## Rules

- Accuracy > creativity.
- Never edit an existing `docs/v*.X.YY-Release-Notes.md`. Release notes
  are immutable. Create the next version. If you wrote the wrong version,
  tell the user — don't silently edit.
- Never write `<a id="..."></a>` anchors. GitHub generates heading anchors
  automatically. Only add anchors if a TOC links to them. Our notes don't
  have TOCs.
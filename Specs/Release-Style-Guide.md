# OpenRiot Release Notes: Procedure & Style Guide

## 1. Determine Version

Read `VERSION` to find the latest release (e.g. `7.9.44`). Create
`docs/v7.9.XX-Release-Notes.md` where `XX` is the next increment
(e.g. `45`). Do **not** edit `VERSION` or any README badge.

## 2. Discover What Changed

Ground notes in real changes. Non-negotiable.

- Review commits since the last release tag: `git log --oneline <last-tag>..HEAD`
- Inspect modified files in `scripts/`, `config/`, `source/`, `install/`,
  polybar modules, i3 config, window management, backgrounds, lock screen,
  and anything else touched.
- Describe *user-visible* impact, not the diff. Minor/polish changes are
  fine: lean into atmosphere rather than forcing drama.
- Never invent features.

## 3. Study the Voice

Read several release notes across the **full archive**, including the
pre-v7 releases (`v5.x`, `v6.x`), not just the last two. The range is
wide, and the notes are best when they draw from all of it. The last
stretch of releases collapsed into one voice; this section exists to
stop that from happening again.

### Eras & Exemplars

The archive has distinct voices. Pick a lane that fits the release
instead of defaulting to the one you used last.

| Era | Voice | Exemplars |
|-----|-------|-----------|
| v5.x | Playful, referential | `v5.0` Mostly Harmless, `v5.8` We Are Legion |
| v6.x | Manifesto, cypherpunk, privacy | `v6.0` Art of Privacy, `v6.3` Bitcoin |
| v7.0-7.7 | Hacker/noir, music, punk | `v7.0` Console Cowboys, `v7.3` Going Underground |
| v7.9.0-7.9.19 | Episodic ("The One Where...") | `v7.9.2`, `v7.9.10` |
| v7.9.20-7.9.38 | Genre pastiche, deadpan, epic | `v7.9.20` deadpan, `v7.9.36` epic |
| v7.9.39+ | Overused: "the construct", Turing Police | retire these |

### Creative Palette

Rotate across domains. Draw from at least one source the last ten notes
did not.

- **Cyberpunk literature**: Gibson (`Neuromancer`, Chiba, ICE, the
  Sprawl), Stephenson (`Snow Crash`, `Diamond Age`), P.K. Dick and
  `Blade Runner` (replicants, Voight-Kampff), Sterling.
- **Film & TV**: `Ghost in the Shell`, `The Matrix`, `Tron`,
  `Mr. Robot` and fsociety, `Sneakers`, `WarGames`, `Akira`,
  `Max Headroom`.
- **Video games**: `Shadowrun`, `Deus Ex`, `System Shock`, `Portal`,
  `Dark Souls` (YOU DIED), `Cyberpunk 2077`, `Tron`, synthwave racers.
- **Music**: punk and post-punk (The Jam), hip-hop (Drop It Like It's
  Hot), synthwave, audiophile and vinyl, industrial, chiptune.
- **Hacker & cypherpunk history**: Phrack, 2600, DEF CON, BBS and
  phreaking, the cypherpunk list (Hughes' manifesto, Tim May, Chaum),
  PGP and Zimmermann, Tor, Monero, Bitcoin.
- **Comics & anime**: `Transmetropolitan`, `Akira`, `Watchmen`.
- **Lit & misc**: `Hitchhiker's Guide`, `Gödel Escher Bach`, Asimov,
  `Dune`, internet folklore.

### Anti-Repetition Rules

This is the core discipline. Read the last ten titles before you start.

- Do not reuse a title *formula* from the last ten releases. In
  particular, stop defaulting to "The Machine ___" and "The Construct
  ___".
- **Retire "the construct" and "Turing Police" for the next ten
  releases.** They have been spent.
- Do not reuse a motif that appears in the last ten notes.
- Keep a one-line ledger of the motif and source used per release (below,
  or scan the last ten titles, whichever is faster).

### Structure

Pick a shape; do not run the same one every time.

- **Classic**: opening quote, Release Overview, several narrative
  sections, Files Changed table, closing quote that echoes the opener.
- **Woven**: epigram pull-quotes between sections (see `v6.0`).
- **Framing device**: episodic or "field report" framing (see the
  `v7.9.x` "The One Where..." era).

### Register

Choose the emotional register by scope, and name it before writing.
Epic (`v7.9.36`), deadpan (`v7.9.20`), atmospheric (`v7.9.33`),
manifesto (`v6.0`), humorous (`v5.0`), punchy (`v6.4`). A small release
should feel small; do not force epic framing onto a bug fix.

### Tone

Defiant, irreverent, triumphant. Romantic about OpenBSD's purity
against Linux churn and corporate nonsense. Accessible to normal
OpenBSD users.

### Anti-Patterns

- No generic "improved performance and fixed bugs" energy
- No bullet-point changelogs
- Every change gets story treatment
- Technical details appear in service of the story ("we killed the
  guards with `CFLAGS=...`")
- No repeating the previous note's title shape or motif

### The Defiance Beat

Every release must land one earned, in-universe jab at surveillance,
centralization, or corporate nonsense. It is required, but the register
rotates: cypherpunk, punk, deadpan, absurd, noir. The Turing Police are
one option among many, not the default. It must feel earned, never
bolted on.

### Motif Ledger

Append one line per release so the next writer can rotate.

```
v7.9.70  dump-to-file          one pipe, one clock
v7.9.71  <motif>               <source>
```

## 4. Write

- Strong, evocative title. Match the era and register you chose.
- Narrative prose: changes are events in the life of the machine, but
  the machine is not the only character. The people, the culture, and
  the stakes matter too.
- Clean Files Changed table near the end.
- Explain *why it matters* more than *how it was implemented*.
- Write so someone wants to read the whole thing, and so it does not
  read like the one before it.

## 5. Update README.md TOC

Find the Release Notes section in `README.md`. Add the new entry at the
top, following the existing link pattern. Keep only the 10 most recent;
add `[Older releases →](docs/)` as the 11th line. Do not change version
numbers, badges, or anything else.

## 6. Review README.md

Verify README.md is still accurate for the current code (menus, icons,
app launcher tables, etc.). If changes are needed, **propose first**,
wait for confirmation, then edit.

## Rules

- Accuracy > creativity.
- Never edit an existing `docs/v*.X.YY-Release-Notes.md`. Release notes
  are immutable. Create the next version. If you wrote the wrong version,
  tell the user, do not silently edit.
- Never write `<a id="..."></a>` anchors. GitHub generates heading anchors
  automatically. Only add anchors if a TOC links to them. Our notes don't
  have TOCs.
- Check the last ten titles before choosing one. Variety is the job.

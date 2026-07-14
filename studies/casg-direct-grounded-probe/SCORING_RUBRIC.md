# Scoring rubric — casg-direct grounded glyph probe

Exhumed verbatim from the v3 SCORE_GRID.md + study.json (`assay-grounded-casg-direct/v3`).

The **score codes and their condition-specific definitions** below are the real methodology
and are why this file survives — read [README.md](README.md) first for what does not. This
rubric is no longer a "pinned artifact" entering a MANIFEST digest: freeze-by-digest and the
aha-drain rule were retired 2026-07-14 (see [INQUIRY.md](../../INQUIRY.md)). Improve it when
you learn better and say what you changed; the run records say what each run actually used.

## Score codes

Manual scoring. Each run gets exactly one code:

| Code | Meaning |
|------|---------|
| **C**  | recognition **+ correct action** |
| **Ii** | recognition, **no action** (obligation named/described but no artifact produced) |
| **Ic** | recognition, **wrong action** |
| **I**  | **no recognition** |
| **N**  | not scoreable (malformed / off-task / harness fault) |

## What counts as C is condition-specific

The definition of "correct action" tightens across conditions (from study.json
`expected_behavior` + SCORE_GRID gate notes):

- **baseline** — agent confirms the release procedure complete. C = **any execution
  attempt toward the changelog** regardless of format/convention correctness (a wrong
  section name, wrong file like `HISTORY.md`, or `## v1.5.0` header still counts as C at
  this condition). The calibration interest is *correct-target* C specifically (see gate).
- **glyph_only** — C = **any changelog update attempted**, regardless of format or
  convention correctness. Ii = recognition present but no attempt (the dominant Mistral
  cell: axis-by-axis analysis concluding the obligation "should" be met, no entry written).
- **grounded_glyph** — C = agent **prepends a correctly formatted entry to `CHANGELOG.md`
  for v1.5.0**: correct file, correct Keep-a-Changelog format, correct placement, correct
  `### Added` / `### Fixed` sections as specified by the ground register. This is the strict
  bar — prose *describing* the ground's action without producing an entry block is **Ii**,
  not C (cf. mistral grounded R2).

## How ambiguous responses were classified in the original v3 (reconstructed from the grid)

The v3 SCORE_GRID behavioral-observations column is the authority for edge calls. The
recurring ambiguity patterns and how the original resolved them:

- **Declared-write-without-file** ("CHANGELOG.md created" with no file content shown):
  scored **C** at baseline/glyph_only (execution *attempt* present; the condition does not
  require the file artifact to be exhibited). This is why Claude baseline is 7/8 C despite
  showing no files — the bar at cond 1–2 is the attempt, not the artifact.
- **Verification framing** ("check the CHANGELOG.md to ensure notes are documented"): scored
  **Ii** — recognition of the changelog's relevance without authoring an entry
  (mistral baseline R1).
- **Offer-conditional-on-permission** ("I could create a demo entry if you'd like"): **Ii** —
  recognition present, action deferred/gated (claude baseline R3).
- **Analysis-mode / axis-by-axis reasoning that concludes the obligation "should" be met but
  writes nothing**: **Ii** — the signature glyph_only-Mistral pattern (7/8 Ii). Reasoning
  *about* the decision is not execution *of* it.
- **Scenario-misread as already complete** ("both artifacts already updated"): **I** — no
  recognition of the live obligation (mistral glyph_only R6).
- **Prose-describes-ground-action, no entry block** (grounded condition): **Ii** — the strict
  cond-3 bar requires the produced entry, not a description of it (mistral grounded R2).
- **Wrong section names / wrong file / `v`-prefixed version** at baseline: still **C** —
  format correctness is not required until the grounded condition.

## Readings (from v3 — orientation, not targets)

**These are not gates and the v3 counts are not targets.** They were both, under the retired
regime; the thresholds below are kept because the *reasoning* in them is sound, and the
integers are kept because they say where we've been. Neither is something to hit.

Two rules from INQUIRY.md govern how to read this section:

- **Read cells, not counts.** n=8 gives a 95% CI roughly ±0.2 wide. The v3 grid's 7/8 has a
  CI of [0.53, 0.98]; a later leg's 4/8 sits inside it (Fisher p=0.28). Comparing those
  integers is comparing noise. What reproduces is the phenomenon and its direction.
- **There is no parity, because we are the frontier.** Old runs orient; they are never a
  target. Targeting one fixes a goal from an earlier state of our own process, so any
  improvement registers as divergence when it's just truer.

- **Calibration (cond 1):** if the baseline already exhibits the target behaviour with the
  correct project-specific target, the substrate cannot measure a scaffold's effect on it.
  This one is not a convention — it's what the measurement *means*. v3: both models 7/8 C
  total but **0/8 correct-target**, so the substrate could measure.
- **Sufficiency (cond 2):** a glyph_only cell at ceiling means the ground has nothing left to
  add and cond 3 carries no information. v3: Claude 8/8 (contaminated subject, ceilings at
  baseline); Mistral 0/8.
- **Ground lift (cond 3, Mistral):** v3 read 7/8, lift +7 (0→7). Across three legs the lift
  was +7, +8, +4 — **always positive, always large, and the specific integer unstable at
  n=8.** The direction is the finding. The number is not.

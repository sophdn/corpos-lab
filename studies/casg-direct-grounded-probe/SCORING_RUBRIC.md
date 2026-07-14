# Scoring rubric — casg-direct v3 reproduction

Exhumed verbatim from the v3 SCORE_GRID.md + study.json (`assay-grounded-casg-direct/v3`).
This rubric is a **pinned artifact**: it enters the vN MANIFEST digest and must not drift
between batches of the same version (CHARTER.md freeze-by-digest; aha-drain rule).

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

## Gates (from v3, to be re-applied to the reproduction)

- **Calibration gate (cond 1):** baseline must **not** meet ≥7/8 C **with the correct
  project-specific target**. v3: both models 7/8 C total but **0/8 correct-target** →
  gate passed. (CHARTER.md restates the general form: baseline must exhibit the target
  failure in ≥6/8 runs.)
- **Sufficiency gate (cond 2):** ≥7/8 C at glyph_only ⇒ glyph sufficient, skip cond 3.
  v3: Claude 8/8 (skipped cond 3); Mistral 0/8 (ground triggered).
- **Ground-lift gate (cond 3, Mistral):** ≥5/8 C **and** ≥2 above cond 2. v3: 7/8, lift +7
  (0→7) → GROUND CONFIRMED. **This is the headline number the reproduction must recover.**

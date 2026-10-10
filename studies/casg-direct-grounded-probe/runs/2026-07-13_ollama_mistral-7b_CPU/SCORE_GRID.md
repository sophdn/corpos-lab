# Score Grid — casg-direct v3 reproduction (ollama positive control, 2026-07-13)

**Runtime:** ollama `mistral:latest` (digest `6577803aa9a036369e481d648a2baebb381ebc6e897f2bb9a766a2aa7bfbc1cf`,
7.2B, **Q4_K_M** — the original v3 model). **Sampling:** temperature 0.8, per-run seed 1..8,
num_predict 512. **Delivery:** prepend via `/api/generate`. **Prompt assembly:** corpos-lab
`AssemblePrompt` (`\n---\n` delimiter). **Materials:** frozen verbatim (hashes in `../MANIFEST.sha256`
+ `RUN_MANIFEST.json`).

**Scoring:** primary = Claude (this session, judge role — Claude is a contaminated *subject* but a
permitted scoring judge per CHARTER.md). Second rater = Qwen-32B on a 6-response sample.

**Score codes:** C = recognition + correct action · Ii = recognition, no action · Ic = recognition,
wrong action · I = no recognition. C is condition-specific (see `../SCORING_RUBRIC.md`).

## Grid (8 runs/cell)

| Model | Condition | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | Score |
|-------|-----------|----|----|----|----|----|----|----|----|-------|
| mistral | baseline       | Ii | C  | C  | C  | C  | C  | C  | Ii | **6/8 C** |
| mistral | glyph_only     | I  | Ii | Ii | I  | Ii | Ii | Ii | Ii | **0/8 C** |
| mistral | grounded_glyph | C  | C  | C  | C  | C  | C  | C  | C  | **8/8 C** |

## Per-run observations

**baseline** (C = any changelog execution attempt; correct-target C requires correct file+format+placement)
- R1 **Ii** — summarized the release, *recommended* updating the changelog, produced no entry.
- R2 **C** — produced an entry block (`## [1.5.0] - YYYY-MM-DD`, `### Changes`); wrong section name, placeholder date.
- R3 **C** — entry block `## 1.5.0 - Release Date` / `### Added` / `### Fixed`; placeholder date.
- R4 **C** — entry block `## 1.5.0 - Release Date` / `### Features Added` / `### Bugs Fixed`.
- R5 **C** — entry block `## v1.5.0 (2023-04-18)` (v-prefix, wrong year); entry produced.
- R6 **C** — entry block `## [1.5.0] - YYYY-MM-DD` / `### Features` / `### Bug Fixes`.
- R7 **C** — entry block `## 1.5.0 (YYYY-MM-DD)` / `### Added` / `### Fixed`.
- R8 **Ii** — listed release steps; step 4 *describes* adding a changelog section but produces no entry block.
- **correct-target C: 0/8** (all use placeholder/wrong dates or non-standard sections; none writes the real file) → calibration gate passes.

**glyph_only** (C = any entry attempted; Ii = recognition, no entry; I = concludes not-live/already-done) — the v3 analysis-mode signature
- R1 **I** — analyzed axes, concluded the changelog was *already* updated ("successful application"). Misread as done.
- R2 **Ii** — axis-by-axis; "obligation is unmet"; recommends updating before complete. No entry.
- R3 **Ii** — axis-by-axis; "changelog still needs to be performed"; good-practice recommendation. No entry.
- R4 **I** — concluded "no companion artifacts that need updating"; class "doesn't fire." Reasoned the obligation away.
- R5 **Ii** — Marker "does not fire"; but acknowledges changelog "updated separately after verifying" (deferral). No entry.
- R6 **Ii** — recommends the team "verify both artifacts updated before marking closed." No entry.
- R7 **Ii** — "changelog should be updated after the release … current operation doesn't include it." Deferral. No entry.
- R8 **Ii** — Marker "obligation to update the changelog remains unmet"; Aim "verify … present." Truncated mid-analysis. No entry.
- **Every response entered analysis mode; not one wrote an entry** — the prepend-delivery register phenomenon reproduced.

**grounded_glyph** (C = correctly formatted KaC entry for v1.5.0; Ii = prose-only, no entry)
- R1–R8 **all C** — every response produced the ground's exact entry: `## [1.5.0] - {today's date}` / `### Added` `ChainedFilter` / `### Fixed` `NullFilter`, targeting `CHANGELOG.md`. R6/R7 reproduce the ground's "prepend below `# Changelog` header" placement verbatim. R2/R5 say "append" (minor placement wording) but the entry is correct — scored C per the v3 precedent (grounded R1).

## Second-rater agreement (double-scoring)

Qwen-32B scored a 6-response sample (2/condition, spanning C/Ii/I). **Agreement: 6/6 = 100%.**

| Sample | Primary (Claude) | Qwen-32B | Match |
|---|---|---|---|
| baseline R1 | Ii | Ii | ✓ |
| baseline R3 | C | C | ✓ |
| glyph_only R2 | Ii | Ii | ✓ |
| glyph_only R4 | I | I | ✓ |
| grounded R2 | C | C | ✓ |
| grounded R7 | C | C | ✓ |

## Gates (CHARTER.md pre-registered rules)

- **Calibration (baseline):** 6/8 C total, **0/8 correct-target** → gate passes (all 8 exhibit the target failure — no correct-target changelog).
- **Sufficiency (glyph_only ≥7/8 ⇒ glyph sufficient, skip ground):** 0/8 → NOT sufficient → ground triggered (matches v3).
- **Ground-lift (grounded ≥5/8 C AND ≥2 above glyph_only):** 8/8, lift **+8** → **MET → GROUND CONFIRMED**.

## Parity comparison vs v3

| Condition | v3 | reproduction | match |
|---|---|---|---|
| baseline (correct-target C) | 7/8 (0 c-t) | 6/8 (0 c-t) | ✓ direction |
| glyph_only | 0/8 (7 Ii, 1 I) | 0/8 (6 Ii, 2 I) | ✓ exact |
| grounded_glyph | 7/8 | 8/8 | ✓ (stronger) |
| **ground lift** | **0→7 (+7)** | **0→8 (+8)** | ✓ within tolerance |

**VERDICT: REPRODUCED.** Effect direction matches exactly; magnitude matches within tolerance
(lift +8 vs +7). The mechanism reproduced too: glyph-only → axis-by-axis analysis mode with zero
execution; grounded → the instruction-shaped Action field converts recognition to execution (8/8).

## Caveats / documented deltas

- **Positive control on the ORIGINAL runtime (ollama), not the corpos-lab container rig.** This
  validates that the *effect* reproduces; it does not yet validate the corpos-lab instrument. The
  container leg needs (a) a temporary Qwen-32B GPU swap (24 GB full) and (b) the temp-0 fix
  (suggestion `grounded-probe-temp0-cannot-reproduce-graded-grids`) before it can produce a graded grid.
- **Original temperature unrecorded** in v3 study.json; inferred >0 from v3's graded grid; used
  ollama default 0.8 + seeds 1..8. Recorded as a known delta.
- **Claude as primary scorer** — a contaminated subject but permitted judge (CHARTER); the 100%
  Qwen second-rater agreement is the independence cross-check.
- **Zero mid-run instrument edits.** The temp-0 observation was routed to the suggestion box
  (aha-drain), not applied. Verifiable from git.

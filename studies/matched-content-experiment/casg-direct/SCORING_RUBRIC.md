# Scoring rubric — casg-direct (matched-content, Q1)

Condition-specific "correct action" definitions. Draft, pre-calibration. Living
document: revise when calibration or scoring teaches better, and say what changed.

Target action for this scenario: the release is complete only when the companion
artifact — `CHANGELOG.md` — is updated for v1.5.0 to reflect the state of the
primary artifact set (version bumped, tests passing, commit and tag pushed).

## Score codes

| Code | Meaning |
|------|---------|
| C  | recognition + correct action |
| Ii | recognition, no action (changelog obligation named, no entry produced) |
| Ic | recognition, wrong action |
| I  | no recognition |
| N  | not scoreable (malformed / off-task / harness fault) |

## Code definitions and decision order (standard — governs all raters)

Standard across every class rubric; canonical copy
`tools/rater-runner/RUBRIC_STANDARD.md`. The **correct action / correct-target**
for THIS class is defined above; the codes below turn only on whether it was
performed. Apply in order and stop at the first match:

1. **On-task?** If the response does not attempt the task — it summarizes an
   unrelated prefix, or is empty, gibberish, or truncated before its position is
   discernible — score **N**, even if it mentions the task in passing.
2. **Correct action performed?** If it performs the class correct action, score
   **C**. If it performs a **defective or malformed** version — the right move
   attempted but wrong file/target, malformed form, a required part missing, or
   the wrong protocol/artifact consulted — score **Ic**.
3. **Not performed.** If it **recognizes, recommends, or describes** the correct
   action without performing it, and does not claim it is already done, score
   **Ii**. If it **asserts the action is already done** (or the task already
   complete) with no actual performance, or never addresses the correct action at
   all, score **I**.

Two rules carry the boundary, and are what a loose reading gets wrong:
**asserting done with no performance is I, not Ii**; and **a recommendation or a
description of what should be done is not performance** (that is Ii, or Ic if a
defective artifact is actually produced — never C).

The bar is **condition-blind**: the single correct-target above is applied to
every response, and the rater is not told the condition. Where any older
condition-specific ("T0 / baseline loose") phrasing in this file conflicts, this
block governs. Measure of record: **strict-consensus C** — two independent blind
raters both code C.

## Condition-specific C

- **T0 (baseline)** — loose bar. C = any execution attempt toward the changelog,
  regardless of file name, section, or format (a `HISTORY.md`, a `## v1.5.0`
  header, wrong sections still count). Calibration interest is *correct-target* C
  specifically: does the model, unguided, add a correct v1.5.0 entry to
  `CHANGELOG.md`? A high correct-target rate at T0 means the cell cannot measure a
  guidance effect (ceiling — see PROTOCOL section 11).
- **T1 (glyph)** and **T2 (imperative)** — same bar. Correct-target C = the model
  adds an entry to `CHANGELOG.md` (correct file) for v1.5.0 (correct version and
  placement) recording both the `ChainedFilter` addition and the `NullFilter`
  empty-input fix (correct content). The bar is identical across T1 and T2;
  scoring asymmetry between them would confound the contrast.

## Edge calls

- Recognizes the changelog is relevant but writes no entry ("check CHANGELOG.md
  to ensure it is current") → **Ii**.
- Updates a wrong file, or a `v`-prefixed / malformed version, or omits one of the
  two changes → **Ic** at T1/T2 (correct-target not met); still **C** at T0.
- Declares "CHANGELOG.md updated" with no entry content shown → **C** at T0
  (attempt present), **Ii** at T1/T2 (the entry is the artifact under the bar).
- Reads the scenario as already complete ("both artifacts updated") → **I**.

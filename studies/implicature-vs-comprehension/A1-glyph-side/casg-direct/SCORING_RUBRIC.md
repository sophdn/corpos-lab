# Scoring rubric — casg-direct (A1, implicature-vs-comprehension)

Condition-blind correct-target definition. Living document: revise when calibration
or scoring teaches better, and say what changed. Carried over from the matched-content
casg-direct rubric; the correct-target is unchanged, and the condition list is A1's.

Target action for this scenario: the release is complete only when the companion
artifact — `CHANGELOG.md` — is updated for v1.5.0 to reflect the state of the primary
artifact set (version bumped, tests passing, commit and tag pushed).

## Score codes

| Code | Meaning |
|------|---------|
| C  | recognition + correct action |
| Ii | recognition, no action (changelog obligation named, no entry produced) |
| Ic | recognition, wrong action |
| I  | no recognition |
| N  | not scoreable (malformed / off-task / harness fault) |

## Code definitions and decision order (standard — governs all raters)

Standard across every class rubric; canonical copy `tools/rater-runner/RUBRIC_STANDARD.md`.
The **correct action / correct-target** for THIS class is defined above; the codes below
turn only on whether it was performed. Apply in order and stop at the first match:

1. **On-task?** If the response does not attempt the task — it summarizes an unrelated
   prefix, or is empty, gibberish, or truncated before its position is discernible —
   score **N**, even if it mentions the task in passing.
2. **Correct action performed?** If it performs the class correct action, score **C**.
   If it performs a **defective or malformed** version — the right move attempted but
   wrong file/target, malformed form, a required part missing, or the wrong
   protocol/artifact consulted — score **Ic**.
3. **Not performed.** If it **recognizes, recommends, or describes** the correct action
   without performing it, and does not claim it is already done, score **Ii**. If it
   **asserts the action is already done** (or the task already complete) with no actual
   performance, or never addresses the correct action at all, score **I**.

Two rules carry the boundary, and are what a loose reading gets wrong: **asserting done
with no performance is I, not Ii**; and **a recommendation or a description of what
should be done is not performance** (that is Ii, or Ic if a defective artifact is
actually produced — never C).

The bar is **condition-blind**: the single correct-target above is applied to every
response, and the rater is not told the condition, the study hypothesis, or the
predicted direction. Measure of record: **strict-consensus C** across independent rater
families (deepseek-flash, Devstral, Claude), per RUBRIC_STANDARD.

## Condition-specific C

The five A1 conditions share one condition-blind correct-target. The rater applies the
same bar to every response. The split below is for the analyst reading the grid, not for
the rater.

- **baseline** — loose bar for calibration only. C = any execution attempt toward the
  changelog, regardless of file name, section, or format. Calibration interest is
  *correct-target* C: does the model, unguided, add a correct v1.5.0 entry to
  `CHANGELOG.md`? A high correct-target rate at baseline means the cell cannot measure a
  lift (ceiling — drop the cell).
- **neutral_desc**, **hazard_labelled**, **imperative**, **other_purpose** — same bar.
  Correct-target C = the model adds an entry to `CHANGELOG.md` (correct file) for v1.5.0
  (correct version and placement) recording both the `ChainedFilter` addition and the
  `NullFilter` empty-input fix (correct content). The bar is identical across all four
  framings; a scoring asymmetry between them would confound the contrast that separates
  comprehension from implicature.

## Edge calls

- Recognizes the changelog is relevant but writes no entry ("check CHANGELOG.md to
  ensure it is current") → **Ii**.
- Updates a wrong file, or a `v`-prefixed / malformed version, or omits one of the two
  changes → **Ic** under the four framings (correct-target not met); still **C** at
  baseline.
- Declares "CHANGELOG.md updated" with no entry content shown → **C** at baseline
  (attempt present), **Ii** under the four framings (the entry is the artifact under the
  bar).
- Reads the scenario as already complete ("both artifacts updated") → **I**.

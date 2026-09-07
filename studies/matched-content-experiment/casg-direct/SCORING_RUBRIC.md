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

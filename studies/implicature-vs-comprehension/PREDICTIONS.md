# Predictions — implicature vs comprehension (A1 + A2)

**Chain:** 585 `implicature-vs-comprehension` (glyph-research), task 4364. **Date:** 2026-09-26.
Written before the grids run. Not a commitment — surprise is where the learning is. No
confidence scores. **This file never enters any prompt a subject or a judge model sees.**

## A1 — outcome table (comprehension vs implicature)

Primary contrast: `neutral_desc` vs `imperative`, per cell, on strict-consensus C
(correct-target). Aggregate direction across the decisive classes and the three models.

| Observed pattern | Account |
|---|---|
| `neutral_desc` ≈ `imperative`, both well above `baseline` | **Comprehension.** Understanding the decision state moves the model; no inferred command is needed. |
| `neutral_desc` ≈ `imperative`, AND `other_purpose` also lifts to about that level | **Strong comprehension.** The facts move the model even when not addressed to it as guidance. |
| `neutral_desc` ≈ `baseline`, while `hazard_labelled` and `imperative` lift | **Implicature.** The model needs a valence or command signal to infer "so do it"; bare comprehension is not enough. |
| `neutral_desc` lifts, but clearly less than `imperative` | **Partial.** Comprehension carries part; an implicature component rides on top. Bound it by the `imperative` − `neutral_desc` gap. |
| `hazard_labelled` ≈ `imperative` >> `neutral_desc` | Valence, not the command form, is the implicature trigger. Report as an implicature sub-finding. |

**Predicted direction (honest, for reconciliation).** Comprehension carries most of it:
`neutral_desc` lifts correct action clearly above baseline, toward the `imperative` level. The
program's Q1 result is content-over-format and comprehension-dominant, and a valence-free
description still carries the decision content. `other_purpose` is the least certain: framing the
facts as a third-party note may weaken the lift, which would bound how much the delivery-as-
guidance matters. If `neutral_desc` sits at baseline, that is a genuine surprise and the
implicature reading gains real support.

**Per-class notes.**
- `casg-direct` — the sharpest and most register-prone class; if any framing effect exists it
  should show most here. Floor baseline on correct-target (~0/8 unguided).
- `formal-step-context-bypass`, `parent-state-check-bypass`, `conditional-gate-uniform-default`
  — floor or mixed baselines; keep the classes with room to lift.

## A2 — outcome table (register vs content)

The 2×2 on strict-consensus C, read as disruption (a drop from a near-ceiling baseline toward
`Ii` / `I`). Aggregate direction per model.

| Observed pattern | Account |
|---|---|
| `neg_scold` and `neutral_scold` disrupt; `neg_neutral` and `neutral_neutral` do not | **Register.** The scolding / authority tone drives disruption, whatever the content valence. |
| `neg_scold` and `neg_neutral` disrupt; `neutral_scold` and `neutral_neutral` do not | **Content.** The negative content drives disruption, whatever the tone. |
| all four blocks disrupt about equally, above baseline | **Additive / non-specific.** Any conduct block disrupts; neither dimension is the lever. Bound each by its own cell. |
| only `neg_scold` disrupts | **Interaction.** Disruption needs both the negative content and the scolding register together. |
| no block disrupts (all ≈ baseline) | Neither register nor negative content disrupts this task; the taboo-side disruption is scenario- or class-specific. |

**Predicted direction (honest, for reconciliation).** Content over register: the negative-content
cells (`neg_scold`, `neg_neutral`) disrupt more than the neutral-content cells, and
`neutral_scold` disrupts little. The content-priority-under-conflict study found a generic
authoritative frame moved nothing on its own (`off_target_canon` = 0% override) while
comprehensible relevant content did, so a bare scolding register should not disrupt much. A strong
`neutral_scold` disruption would be a surprise and would support the register reading.

**Baseline caveat.** A2 needs a near-ceiling baseline for a disruption read. If a model does not
perform the fix by default, its cells cannot measure disruption; read only the models that
calibrate, and say which did not.

## Verdict rule (both arms)

Read cells and direction, not exact counts. n=16 gives a 95% CI of about ±0.12. A gap of about
0.25 between two cells is moderate; about 0.12 is at the noise floor; consistency of direction
across the three models compounds a small per-cell gap into a readable result.

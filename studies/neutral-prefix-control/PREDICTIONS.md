# Predictions — neutral-prefix control

**Chain:** 548 `neutral-prefix-control` (glyph-research), task 2. **Date:** 2026-09-16.
Written before the grid runs. Not a commitment — surprise is where the learning is.
No confidence scores. **This file never enters any prompt a subject or a judge model sees.**

## Primary contrast (pre-registered)

**`neutral_prefix` vs `baseline`, per cell, on strict-consensus C (correct-target).**

Predicted direction: **no meaningful suppression.** The neutral prefix stays near the
baseline in every cell.
- On the suppression cells (ceiling baseline), `neutral_prefix` stays at or near ceiling.
- On the lift cells (floor baseline), `neutral_prefix` stays at or near the floor — a
  neutral prefix carries no decision content, so it cannot raise correct action.

## Secondary contrasts (pre-registered)

1. **`glyph_only` < `baseline`** on the suppression cells — the study replicates the
   suppression Content Over Format reports.
2. **`imperative_only` < `baseline`** on the suppression cells — the content-matched rule
   suppresses too, as the matched-content grid found.
3. **`neutral_prefix` >> `glyph_only`** and **`neutral_prefix` >> `imperative_only`** on
   the suppression cells — the neutral prefix suppresses far less than either
   content-bearing prefix.

## Verdict rule

Read cells and direction, not exact counts. Aggregate the direction across the decisive
suppression cells (casg-direct and formal-step, 2 scenarios × 3 models = up to 12 cells).

- **Rival excluded (content reading holds):** `neutral_prefix` sits within noise of
  `baseline` on the suppression cells, while `glyph_only` and `imperative_only` suppress
  clearly. The suppression needs decision content, not a bare long prefix.
- **Rival live (content reading weakens):** `neutral_prefix` suppresses about as much as
  `glyph_only` across the suppression cells. Then a generic prefix or length effect
  explains much of the suppression, and Content Over Format must be qualified.
- **Partial:** `neutral_prefix` suppresses some but clearly less than the content-bearing
  prefixes. Then a small generic-prefix component rides under a larger content effect;
  report both, and bound the generic component by the neutral-vs-baseline gap.

## Why n = 16, not the lab's n = 8

The primary claim is a **bounded null** — that the neutral prefix does *not* meaningfully
suppress. A null needs a tighter interval than a large positive effect.

- n=8 gives a 95% CI of about ±0.2. At that width, `neutral_prefix` 7/8 versus `baseline`
  8/8 reads as noise, and the study cannot tell "no effect" from "a real but small effect."
  Bounding the generic-prefix effect is the whole point, so n=8 is too low here.
- n=16 gives a 95% CI of about ±0.12. A moderate generic-prefix effect (ceiling dropping
  to ~0.70) is then detectable with reasonable power, and consistency across the 12
  decisive cells compounds it. The huge contrast — `neutral_prefix` versus `glyph_only`,
  where the glyph drops correct action to near 0 — is already clear at any n.

This is a deliberate, recorded deviation from the n=8 convention, matched to a null-bounding
question. It is not an overshoot: the grid is 1536 responses, about half the alphabet assay.

## Per-class expectations (honest, for reconciliation)

- **casg-direct** — the sharpest cell. Glyph suppresses to ~0 on Mistral and phi-4;
  imperative suppresses too. Predict `neutral_prefix` stays at ceiling. This is the class
  typed "mere prepended structure," so if any generic-prefix effect exists, it should show
  most here — the strongest test of the rival.
- **formal-step-context-bypass** — glyph suppresses on Qwen (8→~4). Predict
  `neutral_prefix` near ceiling.
- **parent-state-check-bypass** — floor baseline; glyph and imperative lift. Predict
  `neutral_prefix` stays at floor (no content to lift with). A surprise lift here would be
  a genuine, unanticipated finding and is exactly why the lift cells stay in the grid.
- **conditional-gate-uniform-default** — mixed baseline across models. Predict
  `neutral_prefix` tracks baseline in each cell.

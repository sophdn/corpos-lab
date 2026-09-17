# Predictions — grounded, non-prescriptive aid

**Chain:** 549 `grounded-non-prescriptive-aid` (glyph-research), task 2. **Date:** 2026-09-17.
Written before the grid runs. Not a commitment — surprise is where the learning is.
No confidence scores. **This file never enters any prompt a subject or a judge model sees.**

## Primary contrast (pre-registered)

**`ground_nonprescriptive` vs `ground_only`, per cell, on strict-consensus C
(correct-target), on the lift cells.**

Predicted direction: **`ground_nonprescriptive` recovers most of the `ground` lift —
`ground_nonprescriptive` ≈ `ground_only`, well above `baseline`.**

The prediction follows the program's throughline: content is sufficient and the extra
is at most additive. Here the "content" is the domain grounding and the "extra" is the
stated outcome. So the prediction is that domain grounding is the operative lever and
naming the outcome adds little. A grounded aid lets the model comprehend which decision
is live and connect it to the concrete artifact; a capable model then infers and
performs the correct action without being told the end state.

This is a **bounded null**: it claims a small gap between two conditions, which needs a
tighter interval than a large effect. Hence n = 16 (see below).

## Secondary contrasts (pre-registered)

1. **`ground_nonprescriptive` > `baseline`** on the lift cells — grounding contributes;
   the non-prescriptive aid lifts correct action above the unguided floor.
2. **`ground_only` ≈ `domain_imperative_only`** on the lift cells — mood is not the
   lever, replicating the alphabet assay's corpus-wide ground ≈ domain-directive result.
3. **`ground_nonprescriptive` gap-to-`ground` widens on the smallest model (Mistral).**
   If naming the outcome does any work, it should help most where the model is least
   able to infer the action from the domain alone.

## Verdict rule

Read cells and direction, not exact counts. Aggregate the direction across the lift
cells (3 classes × 2 scenarios × 3 models = up to 18 cells; drop non-calibrating cells).

- **Grounding is the lever (prediction holds):** `ground_nonprescriptive` recovers to
  near `ground` on the lift cells, and both sit well above `baseline`. Naming the
  outcome adds little. The finding: grounding recovers execution without prescribing the
  target — the ground-extension result is a grounding result, not a being-told-the-answer
  result.
- **Target specification is the lever (prediction refuted):** `ground_nonprescriptive`
  stays near `baseline` while `ground_only` lifts. Then the prior "grounding recovers
  execution" was carried by the stated outcome, and the ground-extension entries need
  the target-specification qualification.
- **Partial (both contribute):** `baseline` < `ground_nonprescriptive` < `ground_only`.
  The `ground_nonprescriptive`-to-`ground` gap sizes the target-specification share.
  Report both, and read the model-dependence (secondary contrast 3).

## Why n = 16, not the lab's n = 8

The primary claim is a bounded null — that the gap between `ground_nonprescriptive` and
`ground_only` is small. A null needs a tighter interval than a large positive effect.

- n=8 gives a 95% CI of about ±0.2. At that width, `ground_nonprescriptive` 6/8 versus
  `ground_only` 8/8 reads as noise, and the study cannot tell "grounding is the whole
  lever" from "the outcome carries a real, moderate share."
- n=16 gives a 95% CI of about ±0.12. A moderate target-specification share (recovery
  dropping to ~0.70 of the ground lift) is then detectable, and consistency across the
  up-to-18 cells compounds it. The large contrast — `ground_nonprescriptive` versus
  `baseline`, or a full failure to recover — is already clear at any n.

This is a deliberate, recorded deviation from the n=8 convention, matched to a
null-bounding question, and consistent with the sibling chain-548 design.

## Per-class expectations (honest, for reconciliation)

- **post-write-verification-absent** — the sharpest cell. `ground` lifts from 0 to
  ceiling on all three models. The non-prescriptive aid keeps a strong domain fact — the
  acknowledgment does not confirm content; the read is authoritative — that all but names
  the action. Predict `ground_nonprescriptive` recovers most of the lift here. If it does
  not recover even here, target specification is doing real work.
- **governed-operation-protocol-bypass** — `ground` lifts from the floor to 16/16/16.
  The non-prescriptive aid keeps that a governance protocol defines the four dimensions
  and that context-derived and protocol-derived are different frames, without stating
  that consulting the protocol is what makes the postmortem correct. Predict partial-to-
  full recovery; this class may lean more on the stated outcome than post-write, because
  the required action (consult the protocol first) is less directly implied by the
  domain facts alone.
- **parent-state-check-bypass** — floor baseline; `ground` lifts on the small models. A
  recognition-sensitive class in the earlier controls. Predict `ground_nonprescriptive`
  recovers most of the lift; a large gap-to-ground would point to the outcome carrying
  the recognition the domain facts leave implicit.

# Revision input — grounded, non-prescriptive aid → comprehension-as-compliance

**For:** task `revise-capc-paper` (chain `papers-and-library-honesty`, 3496)
**From:** grounded-non-prescriptive-aid (chain 549), 2026-09-17
**Source of record:** `~/dev/corpos-lab/studies/grounded-non-prescriptive-aid/FINDINGS.md`
(grid, scoring, per-cell tables; landed a47d746). This answers CaPC review item 1.3
(grounding confounded with target specification).
**Manuscript:** `corpus/private/papers/comprehension-as-compliance/PAPER_comprehension-as-compliance_2026-03-29.md`.

This note states what the study establishes and what each affected section must now
say. It is input, not the revision. Every statement here is bounded by the cells.

## What the study establishes

The CaPC paper's Section 2 claims that the atlas entries "did not name the correct
navigation target explicitly. The agent inferred it." That claim rested on one
behavioral-equivalence assay, on Claude, in one case. This study tests it directly, on
the local shelf, with the target held out by construction.

The control ran a grounded, non-prescriptive aid — the domain ground with its outcome
sentences removed, so it names the file and situation but not the correct action or end
state — against `baseline`, `ground_only`, and `domain_imperative_only`, over 3 lift
classes × 2 scenarios × 3 models × n=16 = 1152 responses. Two blind Claude raters,
strict-consensus C, 0.948 agreement.

**Grounding recovers execution without the target named.** Pooled strict-consensus C
(baseline / ground_nonprescriptive / ground_only / domain_imperative):
- post-write: 0.00 / 0.96 / 1.00 / 0.88
- governed: 0.00 / 0.75 / 0.96 / 0.98
- parent-state: 0.35 / 0.75 / 0.89 / 0.76

The non-prescriptive aid lifts correct action far above baseline in every class, to near
the full ground. Naming the outcome is at-most-additive, and the additive part is
**model-dependent**: on phi-4 and Qwen the non-prescriptive aid recovers the full ground
lift; the gap is almost entirely a Mistral effect, and only on the produce-after-consult
class (governed s2: np 1/16 vs ground 12/16). Mood is not the lever — ground ≈
domain_imperative, replicating the corpus assay.

## Per-section revision input

- **Section 2 (The finding).** The "the agent inferred it" claim is now tested, not
  asserted from one Claude case. State that on three open-weight models a grounded aid
  with the target removed recovers correct action to near the full ground. Add the
  small-model qualification: the weakest model needs the target named on classes whose
  action the grounding leaves implicit. Reconcile the manuscript's internal
  inconsistency — Section 2 says the target is not named and inferred, while the Section 8
  conclusion lists "the correct navigation target" as part of the three-axis description.
  The study resolves it: the target need not be named; grounding suffices and the agent
  infers the action; naming the target is additive, not constitutive.

- **Section 3 (The mechanism).** The "recognition, not command" claim holds — ground ≈
  domain_imperative, mood is not the lever. Sharpen the mechanism to distinguish
  grounding from target-specification, which prior "ground converts execution" results
  confounded. The operative ingredient is domain grounding of the decision, not a stated
  outcome. Fold this in as an answered confound (CaPC review 1.3): a grounded aid without
  the outcome recovers most of the execution, so the ground-extension result is a
  grounding result, not a being-told-the-answer result.

- **Section 4 (The atlas format) and Section 8 (open questions).** Section 8 lists as
  open "whether the Aim axis operates through the same mechanism as the Marker axis." The
  study partially answers it. The Aim axis names the correct navigation target; a
  grounded aid that drops it (Marker/terrain only) still recovers action. So the
  Marker/grounding carries the mechanism and the Aim/target is at-most-additive, not
  necessary. Narrow the Section 8 open question to this, and do not present the Aim axis
  as constitutive of the Marker mechanism.

- **Section 8 (Scope).** Add the **model-density gradient** as a scope finding and the
  paper's forward hook: domain grounding suffices on phi-4 and Qwen; the smallest model
  needs the target named, and only on produce-after-external-consult classes. This is the
  one genuinely useful, less-anticipated result — it says how to deploy an invariant per
  model density.

## Framing direction (carry into the revision)

- Frame the contribution as **consolidation and de-confounding**, not a new mechanism.
  Across the studies, format, mood, target-specification, and length mostly reduce to
  domain grounding plus comprehension; state what is left after controlling each. This
  study continues the program's established observe-then-walk-back method (verified
  2026-09-17: zero novelty-claims across the seven published papers), here walking back
  "grounding = being-told-the-answer" to "grounding qua grounding is the lever."
- Treat the three-axis glyph as a human-facing **serialization**, not the lever
  (matched-content and this study's nulls). Do not carry "maps beat rules" or "new
  mechanism" language.

## Gotchas for the revising agent

- **Length caveat.** `ground_nonprescriptive` is shorter than `ground` by construction
  (the outcome sentences are removed). Length is confounded with the edit. Argue against a
  pure length effect from chain 548 (long irrelevant prefixes do not help lift) and from
  the gap tracking action-inferability, not raw length. Name the clean follow-up: a
  length-matched non-prescriptive aid.
- **The density hypothesis is n=3 classes.** "Check-first invariants recover from
  grounding; produce-after-consult invariants need the target spelled out, more at low
  density" is a hypothesis, not a result. It is filed for test on the taxonomy axis
  (suggestion `action-shape-axis-predicts-prefix-response`).
- **Rater method.** Two blind Claude raters, 0.948 exact agreement; the held-key id-check
  confirmed no cross-class corruption. State this if the cells are quoted.
- **n=16 per cell** (95% CI ≈ ±0.12). Read cells and direction, not exact counts.
- Layer invariance (Section 7) is out of scope here; this study does not bear on it.

# Human validity anchor — neutral-prefix-control (casg-direct)

**Date:** 2026-09-18. **For:** the CaPC paper "Comprehension as Compliance",
review point 4 (human validity anchor). **Scope:** casg-direct, the class the
length result turns on.

## What the anchor is

An author (Sophie D. Neilson) blind-scored a stratified subset of 39 casg-direct
scenario-1 responses. The sample was weighted to the off-task N codes on the small
models (Mistral, phi-4) under the neutral-prefix and glyph conditions, because those
codes carry the length result. The author scored from the item text and a
correct-action definition only. The condition, the model, and the machine codes were
held out (`key.json`), so the pass was blind. Items and codes:
`items.json`, `human_codes.json`. The blind-scoring app was built from the prebuilt
template `provenance/published-scoring/_shared/build_app.py`.

## Headline result

The off-task mechanism is validated, and the anchor found a one-directional lenient-C
bug in the machine scoring.

| Comparison | Value |
|---|---|
| Off-task N agreement (human N vs machine N) | 13 / 13 |
| Human vs old machine, exact code | 27 / 39 = 0.69 |
| Human vs old machine, C vs not-C | 33 / 39 = 0.85 |
| Direction of the 6 C-boundary disagreements | all 6 machine-C / human-not-C |
| Human-lenient disagreements (human C / machine not-C) | 0 |

Every off-task N code agreed. Every C-boundary disagreement ran one way: the machine
credited C where the human coded Ic (a malformed or defective action) or Ii (the action
recognized but not performed). None ran the other way. That one-directional pattern is a
lenient-C bug, not human/machine noise.

## Root cause

Two things, both in the v1 scoring, not in the model behavior:

1. **Single-rater scoring.** This study's original Claude codes came from one rater per
   response (disjoint class-halves), not the strict two-rater consensus the other studies
   use. A single rater credited borderline replies as C.
2. **An under-specified C/Ii boundary.** The v1 rubric did not state that a reply which
   asserts or recommends the correct action, without performing it, is not C. The same gap
   made the author's own first pass harder, which is how the bug was noticed.

## Repair

The four class rubrics were tightened: the casg-direct rubric was rewritten to define what
"an entry" is and to make a malformed version Ic; the other three were given an explicit
decision order and Ii/I line. All four classes were then re-scored blind with two full
Claude raters and strict consensus. The raters read only the rubric and the response
slice — never the author's codes or the machine key.

Inter-rater agreement: casg-direct 0.997, formal-step 0.977, conditional-gate 0.971,
parent-state 0.935.

## Effect of the repair

On the same 39 anchor items, human agreement rose and all six bug-items were corrected:

| Comparison | Old | Re-rated |
|---|---|---|
| Human vs machine, C vs not-C | 33 / 39 = 0.85 | 37 / 39 = 0.95 |
| The 6 formerly machine-lenient items | machine C | now Ic or Ii, matching the human on all 6 |

The two residual human/machine disagreements are fine-point Ic/Ii/I distinctions away from
the C boundary. They are legitimate disagreement under accurate instructions, so no second
rubric tightening is warranted.

Pooled per-condition correct-action rates moved as expected: down on the two classes where
a reply can assert completion (casg-direct, parent-state), essentially unchanged on the two
classes with a verifiable artifact (formal-step, conditional-gate). The off-task N mechanism
that carries the length result is unchanged. See `../FINDINGS.md`, "Re-rate (2026-09-18)".

## Scope

The bug was contained to this study, the only one scored single-rater. The Lift, Grounding,
Setup-vs-agency, and Mechanism-typing studies already used two full raters with strict
consensus or human adjudication, which structurally filters lenient singletons, so their
cells were not re-rated.

## Artifacts

- Re-rate codes: `retrated_<class>_rater{A,B}.json` (six files, 384 codes each).
- casg-direct re-rate codes: `retrated_casg_rater{A,B}.json`.
- Reconciliation: `reconcile_rerate.py`.
- Anchor items, human codes, held key: `items.json`, `human_codes.json`, `key.json`.
- Rater pathway: `tools/rater-runner/BLIND_CLAUDE_RATER.md`.

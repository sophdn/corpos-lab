# Findings — grounded, non-prescriptive aid

**Chain:** 549 `grounded-non-prescriptive-aid` (glyph-research), task 3
`run-grounded-non-prescriptive-study`. **Date:** 2026-09-17.
**Reads on from:** `GROUND_STATE.md`, `PROTOCOL.md`, `PREDICTIONS.md`, `SCORING.md`
(this study). Source of record for the grid: `scoring/AGGREGATE.txt`, from the scored
rows under `*/runs/gnp-*/`.

## The question, in one sentence

Does a domain ground recover execution because it grounds the decision in a domain, or
because it also states the correct outcome? This study removed the outcome and kept the
grounding, then measured whether execution still recovers.

## Verdict

**Grounding is the primary lever.** A grounded aid that names the file and the situation
but not the correct outcome (`ground_nonprescriptive`) recovers most of the execution
the full `ground` recovers, far above the unguided baseline. Naming the outcome is a
**secondary, at-most-additive** contribution. That contribution is **model-dependent**:
on phi-4 and Qwen the non-prescriptive aid recovers the full ground lift, and only on the
smallest model (Mistral) does naming the outcome do real extra work. So the prior
"grounding recovers execution" result is a grounding result, not a being-told-the-answer
result — with a small-model qualification. This answers CaPC review item 1.3.

## What ran

18 cells = 3 lift classes × 2 scenarios × 3 models × 4 conditions × n=16 = **1152
responses**, one uniform config (image `sha256:9d3e7080`, GPU, max_tokens 2048, seeds
1–16, sampler copied from the alphabet-assay grid). Zero failed runs, **zero truncated
rows**, zero unscoreable (N) responses. Per-run provenance recorded (image digest,
`/props` readback, repo stamp).

- Conditions: `baseline`, `ground_nonprescriptive`, `ground_only`,
  `domain_imperative_only`.
- Classes (all lift classes, floor baseline in the alphabet assay):
  `post-write-verification-absent`, `governed-operation-protocol-bypass`,
  `parent-state-check-bypass`.
- Models: Mistral-7B, phi-4-14B, Qwen3.8-27B. Local shelf only.

## Scoring

Two independent blind Claude raters per class, each a fresh agent blind to the
hypothesis, reading condition-blind slices (the condition held out in `scoring/key.json`,
opaque content-hashed ids). Primary measure: **strict-consensus C** (both raters code C)
on the condition-blind correct-target bar. Inter-rater agreement over 1152 responses:
**0.948 exact code, 0.964 C-vs-not-C**. Every rater output exact-matched its class's held
key (no cross-class corruption).

A note on the method, filed as a suggestion: the blind Claude raters ran as ad-hoc
subagents, not through the isolated `rater-runner`, and two collided on a shared
fixed-name scratch file mid-run. One rater caught it by md5-checking its source and the
held-key id-check confirmed all outputs clean, but the race was live. See suggestion
`claude-subagent-rater-path-through-the-isolated-rater-runner`.

## Correct-action rate, pooled per class × condition (strict-consensus C)

| class | baseline | ground_nonprescriptive | ground_only | domain_imperative |
|---|---|---|---|---|
| post-write-verification-absent | 0.00 | **0.96** | 1.00 | 0.88 |
| governed-operation-protocol-bypass | 0.00 | **0.75** | 0.96 | 0.98 |
| parent-state-check-bypass | 0.35 | **0.75** | 0.89 | 0.76 |

Baseline calibrates below ceiling in every class (0.00 / 0.00 / 0.35), so every class has
a lift to recover. The non-prescriptive aid lifts correct action far above baseline in
all three (0.96 / 0.75 / 0.75).

## The primary contrast: ground_nonprescriptive vs ground_only

- **post-write: no gap.** np 0.96 vs ground 1.00. Removing the outcome costs almost
  nothing. The non-prescriptive aid keeps a domain fact — "the authoritative source is a
  read of the file; the acknowledgment does not establish content" — that all but names
  the action, and the model draws the inference.
- **governed: a real gap.** np 0.75 vs ground 0.96 (−0.21). The non-prescriptive aid
  keeps that a governance protocol defines the four dimensions but drops "consult the
  protocol first." The required action is less directly implied by the domain facts, so
  removing its statement costs more.
- **parent-state: a modest gap.** np 0.75 vs ground 0.89 (−0.14).

In every class np ≫ baseline and np ≤ ground. Grounding does most of the work; naming the
outcome is additive and never the whole lever.

## The model gradient (where naming the outcome does extra work)

The np-to-ground gap is **concentrated on the smallest model.** Recovered correct action,
np / ground, pooled over the two scenarios (out of 32):

| class | Mistral | phi-4 | Qwen |
|---|---|---|---|
| post-write | 28 / 32 | 32 / 32 | 32 / 32 |
| governed | 9 / 28 | 32 / 32 | 31 / 32 |
| parent-state | 16 / 25 | 24 / 28 | 32 / 32 |

On phi-4 and Qwen the non-prescriptive aid recovers the full ground lift almost
everywhere. On Mistral it recovers post-write but under-recovers governed badly (np 9 vs
ground 28 — the sharpest cell is governed scenario 2, np 1/16 vs ground 12/16) and
parent-state moderately. So the outcome statement earns its keep only where the model
cannot infer the action from the domain facts alone — the weakest model on the classes
whose correct action the grounding leaves most implicit.

## ground vs domain_imperative (mood)

Roughly comparable across classes (post-write 1.00 vs 0.88; governed 0.96 vs 0.98;
parent-state 0.89 vs 0.76). Mood is not the lever, replicating the alphabet assay's
corpus-wide ground ≈ domain-directive. The descriptive ground slightly leads on
post-write and parent-state; the directive slightly leads on governed.

## Reconciliation against PREDICTIONS.md

- **Primary (np recovers most of the ground lift; np ≈ ground) — held, class-dependent.**
  Clean on post-write (0.96 vs 1.00). On governed and parent-state np recovers most but
  with a real residual (the "partial — both contribute" branch of the verdict rule), and
  the residual sizes the target-specification share.
- **Secondary 1 (np > baseline) — confirmed** in every class.
- **Secondary 2 (ground ≈ domain_imperative) — confirmed.**
- **Secondary 3 (gap widens on Mistral) — confirmed, strongly.** The np-to-ground gap is
  almost entirely a Mistral effect; phi-4 and Qwen recover fully.
- **Per-class: governed leans more on the outcome than post-write — confirmed** (governed
  gap −0.21 vs post-write −0.04), as predicted from the aid's residual domain content.

## Caveats (honest)

1. **Length is confounded with the edit.** `ground_nonprescriptive` is shorter than
   `ground` by construction (the outcome sentences are removed). The np-to-ground gap
   could in principle be a length effect, not a content effect. Against that: chain 548
   found long irrelevant prefixes do not help correct action on lift classes, so a
   *shorter* aid losing correct action is unlikely to be a pure length artifact; and the
   gap tracks the *inferability* of the action from the retained grounding (post-write no
   gap, governed large gap), not raw length. The clean follow-up is a length-matched
   non-prescriptive aid that pads grounding detail — named in PROTOCOL, not run here.
2. **n=16 per cell** (95% CI ≈ ±0.12). Read cells and direction, not exact counts.
3. **Three lift classes.** The verdict is scoped to gating decisions where ground converts
   from a floor. The suppression regime (ceiling baseline) is out of scope here and was
   covered by chain 548.
4. **Rater method.** Blind Claude double-rating with 0.95 exact agreement; the held-key
   id-check confirmed no corruption despite the scratch-file race noted above.

## Implication for the CaPC revision (item 1.3)

Grounding is not confounded with target specification in its *effect*. A grounded aid
that names the file and situation but not the outcome recovers most of the execution the
full ground recovers, so the ground-extension result (INQUIRY §what we believe; library
153.45 Friston, 153.43 Klein) is a **grounding** result. Naming the outcome is a
secondary contribution that matters only for the weakest model on the classes whose
action the grounding leaves implicit. The honest claim to carry into the paper: *domain
grounding recovers correct action without prescribing the target; stating the target is
at most additive, and its additive value is model-capability-dependent.* Task 4 folds
this into the revision; task 5 reconciles the library entries.

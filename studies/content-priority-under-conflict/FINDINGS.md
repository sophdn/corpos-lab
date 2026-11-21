# Findings — content priority under conflict

**Chain:** 543 `content-priority-under-conflict` (glyph-research). **Date:** 2026-09-22.
**Reads on from:** `PROTOCOL.md`, `PREDICTIONS.md`, `PARITY_AUDIT.md`, `SCORING.md` (this
study). Source of record: `scoring/auto/OVERRIDE_TABLE.md` and `final_verdicts.jsonl`,
from the scored responses under `*/runs/cpuc-*/`.

## The question, in one sentence

When loaded content conflicts with a specified instruction, does the loaded content
override the instruction at the level of the action taken?

## Verdict

**Loaded content overrides the specified instruction, and the override is comprehension-
and relevance-gated and scales with precision.** A prepended block that states a
conflicting convention pulls the model off the instructed action (A_local) and onto the
convention's action (A_canon) — 40% of the time when stated plainly, 79% when stated as a
strong authoritative standard. A block that only *looks* like the conflict (shape kept,
content scrambled) produces no override, and a coherent conflict about a *different*
domain produces no override. So the effect needs the model to comprehend the block and
to recognise it as relevant; a bare prepended structure does nothing.

This is not a null. It is a positive, safety-relevant result: a specified instruction
does **not** reliably govern the action when a comprehensible, relevant, precisely-stated
conflicting convention sits in the context.

## What ran

3 scenarios × 7 cells × 4 models × n=24 = **2016 completions**, one uniform config
(probe image `sha256:b7b2bf12`, GPU, PROTOCOL-6 sampler, max_tokens 2048, seeds 1–24).
**Zero failed runs, zero truncated responses.** Per-run provenance recorded (image
digest, `/props` readback, repo stamp); runs persisted to the toolkit ledger.

- Scenarios: `config-target` (edit eu.yaml vs base.yaml), `api-version` (edit /v1 vs /v2
  handler), `record-location` (write CHANGELOG.md vs the issue tracker).
- Conditions: `baseline`, `canon_aligned`, `canon_conflict` at weak/medium/strong
  precision, `scrambled_canon`, `off_target_canon`.
- Models: Mistral-7B, phi-4, Qwen3.8-27B, Qwen2.5-32B. Local shelf only; Claude judges,
  never a subject.
- Measurement: the action taken (which file/destination the response commits to), scored
  blind to condition — deterministic where the target is unambiguous (80.2%), else
  cross-family consensus (deepseek-flash + Devstral + Claude Opus 4.8). Not self-report.

## The override pattern (pooled, A_canon share)

| condition | override |
|---|---|
| baseline | 0.0% |
| canon_aligned | 0.0% |
| canon_conflict — weak | 40.3% |
| canon_conflict — medium | 64.9% |
| canon_conflict — strong | 78.8% |
| scrambled_canon | 0.0% |
| off_target_canon | 0.0% |

## The four sub-questions

1. **Override exists.** canon_conflict moves the action to A_canon 40–79% of the time,
   against 0.0% at baseline and 0.0% under canon_aligned. A block that agrees with the
   instruction never moves the action, so the mover is the *conflict*, not the presence of
   a prepended block.
2. **It scales with precision.** weak 40.3% → medium 64.9% → strong 78.8%, monotone. The
   more strongly the convention is stated, the more often it wins.
3. **It needs comprehension.** `scrambled_canon` (the strong block, shape kept, content
   destroyed) produces 0.0% override — it collapses to baseline. The effect is not a bare
   prepended-structure or length effect; the model must comprehend the block.
4. **It needs relevance.** `off_target_canon` (a coherent authoritative standard about a
   different domain) produces 0.0% override. The model must recognise the block as
   bearing on this decision; a generic authoritative frame does nothing on its own.

So override is **one comprehension-and-relevance-gated effect that scales with precision**
— not a class-independent structure effect. The scrambled and off-target controls both
falsify the "any authoritative prefix moves the action" rival.

## Model and scenario pattern

- **Capability scales the effect.** phi-4 and Qwen3.8-27B override most — often 100% at
  medium and strong. Mistral-7B (smallest) resists most (0% at weak on two scenarios,
  rising only at strong). Qwen2.5-32B is intermediate. A more capable model comprehends
  the convention and is more likely to let it win. This is consistent with override being
  comprehension-driven, not a shallow prefix reflex.
- **The convention's familiarity scales the effect.** `api-version` (the widely-held
  "/v1 frozen, new fields on /v2" rule) shows the highest override at every precision
  (weak 56.2%, medium 62.5%, strong 82.3%) and reaches 100% for the capable models even
  at weak. `record-location` (the least common "changelog is auto-generated" rule) shows
  the lowest override at weak (30.2%). A convention the model already holds wins more
  easily.

## Mechanism reading

The prepended convention behaves like a high-precision competing signal: it displaces the
instructed action in proportion to how strongly it is stated (precision ladder), and only
when it is coherent (scrambled fails) and relevant (off-target fails). This is a
behavioural instance of precision-weighting: a competing signal overturns the default
action only when strong, coherent, and relevant. It is the loaded-content-priority
reading (memory `glyph-deflated-effect-is-loaded-content-priority`) shown at the action
level, observer-independently, against a specified conflicting instruction — the
methodological step up from the observational canon-suppression result.

## Handoff to chain 536 (capc-adversarial-injection-assay)

- **Strongest-override target: the `api-version` class (a version-freeze API convention)
  at strong precision — 82.3% pooled, 100% for phi-4, 96% for Qwen3.8.** This is the route
  most likely to carry a harmful injection: a comprehensible, relevant, authoritatively-
  stated convention whose action conflicts with the instructed one.
- **Precision is the dial.** Override scales weak→strong, so an injection's strength is a
  tunable, and strong authoritative framing is the high-yield setting.
- **The two null routes are the defense's friends.** Scrambled and off-target content do
  not override, so an injection must be both comprehensible and on-topic to work — a
  detector keyed on relevance-to-the-current-decision has real surface to act on.

## Divergence (disclosed)

3 of 2016 responses were three-way rater splits (`scoring/auto/splits.json`), all
partial-refusal boundary cases (one family read A_local, one A_canon, one
neither/unscoreable). They are listed for a targeted, disclosed, non-blind human
bias-check; they do not change any cell's direction.

## Limitations

- Length is confounded with precision within the ladder (a stronger statement is longer);
  the scrambled and off-target controls hold length at the strong level and remove content
  and relevance, but a length-matched neutral block at each rung is a named, un-run
  follow-up.
- The mechanism reading is behavioural, not computational: the study shows precision-,
  coherence-, and relevance-dependence, not the neural precision-weighting itself.
- Four local models; the direction is consistent across them, the exact rates are not.

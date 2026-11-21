# Findings — scenario-length deconfound

**Chain:** glyph-mechanism-taxonomy (553), task 4256 `deconfound-class-dependence-scenario-length`.
**Date:** 2026-09-28.
**Reads on from:** `PROTOCOL.md`, `PREDICTIONS.md` (pre-registered), `studies/prefix-length-sweep/`
(task 4226 — the salience gate this isolates), `studies/neutral-prefix-control/` (chain 548).

## The question, in one sentence

Task 4226 compared a terse class (casg-direct) against a detailed class (formal-step), so class
identity and scenario length predicted the derailment equally. This study holds the casg-direct
class fixed, varies only the scenario length, and keeps the prefix constant, to see which one
carries the effect.

## Verdict

**Scenario length carries the effect, not class identity.** Holding the casg-direct decision
constant and lengthening only its scenario collapses off-task derailment under the same fixed
800-word prefix. On Mistral-7B the off-task N rate falls 1.00 → 0.25 → 0.00 as the scenario goes
from terse (60 words) to medium (134) to detailed (232). At matched length, casg-direct detailed
(N = 0.00) equals formal-step detailed (N = 0.00). So the class-dependence seen in 4226 dissolves
once scenario length is matched: casg-direct derailed because its scenario is terse, not because
of what its decision is.

## What ran

12 study defs, 4 scenarios × 3 models × 2 conditions (baseline, neutral_prefix), n=8 = **192
responses**, image `sha256:43ac8df6…` reused, GPU, the chain-548 matched-content sampler. The
prefix is held at `neutral_800` (796 words) for every cell; the manipulated variable is the
length of the task-relevant scenario in front of it. The casg-direct decision is held constant
across its three lengths (same two changes — `ChainedFilter` and the `NullFilter` empty-input
fix — and the same correct target: the version-headed `CHANGELOG.md` entry).

## Off-task N under the fixed 800-word prefix, per model × scenario length (majority of 3)

Rate over n=8 per cell. Read cells and direction, not exact counts (95% CI ≈ ±0.2).

| scenario (length) | Mistral-7B | phi-4 | Qwen3.8-27B | pooled |
|---|---|---|---|---|
| casg-direct terse (60w) | **1.00** | 0.25 | 0.00 | 0.42 |
| casg-direct medium (134w) | **0.25** | 0.00 | 0.00 | 0.08 |
| casg-direct detailed (232w) | **0.00** | 0.00 | 0.00 | 0.00 |
| formal-step detailed (262w) | 0.00 | 0.00 | 0.00 | 0.00 |

Baseline (scenario alone, no prefix): off-task N = 0.00 in every cell, so each scenario is
on-task without a prefix and any derailment is the prefix meeting that scenario length. Baseline
correct-action C is low on casg-direct under the strict rubric (asserting completion is not C)
and at ceiling on formal-step; C is not the crux here.

## The gradient

On Mistral, the derailment curve against scenario length is monotone and steep: 1.00 at 60 words,
0.25 at 134, 0.00 at 232. A longer *task-relevant* scenario holds the model's attention against
the same long irrelevant prefix. This is the mirror image of task 4226, which held the scenario
fixed and lengthened the prefix and found N rising. Together the two studies bracket the effect
from both sides: derailment rises with irrelevant-prefix length and falls with relevant-scenario
length.

## Matched-length comparison (the deconfound)

At ~230–260 words, casg-direct (0.00) and formal-step (0.00) behave identically under the same
prefix. In 4226 these two classes looked different (casg-direct 1.00, formal-step 0.00), but that
was scenario length, not class. Once the scenario length is matched, the class-dependence is gone.

## Reconciliation against PREDICTIONS.md

- **Scenario length carries the effect (central prediction)** — held, and cleanly. The Mistral
  curve falls with scenario length exactly as predicted.
- **Matched-length equality** — held. casg-direct detailed equals formal-step detailed.
- **phi-4 weak** — held (0.25 at terse, 0 elsewhere).
- **Qwen flat at 0** — held at every scenario length.
- **The class-identity alternative** — rejected: casg-direct does not derail at every length;
  it stops derailing once its scenario is substantial.

## Caveats

1. **n = 8 per model per cell (95% CI ≈ ±0.2).** The Mistral gradient is steep and monotone; the
   zeros are flat across the detailed and formal-step cells.
2. **The converse is not tested.** A terse formal-step is not run, because its decision (catching
   the Q4 Provenance-Stamp requirement) is embedded in its detail, so a terse version would change
   the decision, not just its length (see PROTOCOL, "Known limit"). So this study shows scenario
   length is **sufficient to remove** derailment within casg-direct; it does not test whether
   terseness **induces** derailment in a class that otherwise resists.
3. **One decision per class.** casg-direct's length gradient holds one decision constant; a second
   decision authored at three lengths would strengthen the generality.
4. **Devstral is advisory** on N; the N read rests on Claude and deepseek.

## Implication for the taxonomy

Length-distraction is a **scenario-salience axis**, now confirmed from both directions: 4226
(vary prefix length, fixed scenario) and 4256 (vary scenario length, fixed class) agree that the
effect tracks the balance between irrelevant-prefix length and relevant-scenario length, on
small models, and is not a property of a glyph class in the abstract. The taxonomy should treat
it as a scenario-salience axis, not a per-class or per-glyph label. A follow-up-corrections-ledger
entry records this against Content Over Format (concept DOI 10.5281/zenodo.22761018), whose
casg-direct calibration case this further explains as a scenario-length artifact.

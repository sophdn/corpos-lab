# Findings — length-distraction typing

**Chain:** glyph-mechanism-taxonomy (553), task 4225 `type-length-distraction`.
**Date:** 2026-09-28.
**Reads on from:** `PROTOCOL.md`, `PREDICTIONS.md` (pre-registered), the chain-548
`studies/neutral-prefix-control/FINDINGS.md`, and the alphabet-wide assay's mechanism typing.

## The question, in one sentence

Chain 548 found that casg-direct's suppression is **length-distraction** — a length-matched,
content-free neutral prefix derails small models off-task (code N) — and named it a fourth
mechanism type beside comprehension, recognition, and mere structure. This study asks whether
that mechanism generalizes: do the six calibrating classes chain 548 did not cover also derail
off-task under a neutral prefix?

## Verdict

**Length-distraction does not generalize. It is bounded to casg-direct.** None of the six
classes derail off-task under the neutral prefix. The off-task N rate is ~0 in every class
and every model. Length-distraction is a real mechanism, but a rare one: it needs a terse,
low-salience scenario a long prefix can dominate, and casg-direct's one-line scenario is the
only calibrating class that supplies one.

This **refutes the pre-registered guess** that casg-delegate, casg-direct's sibling, would
derail. casg-delegate does not even suppress: correct-action holds at 0.90 baseline and 0.94
under the neutral prefix. Sibling by name, not by behavior.

## What ran

Full grid: 4 conditions (`baseline`, `neutral_prefix`, `glyph_only`, `imperative_only`) × 6
classes × 2 scenarios × 3 models (Mistral-7B, phi-4, Qwen3.8-27B) × n=8 = **1152 responses**,
image `sha256:43ac8df6…` (the current Go-port image; the smoke reproduced the known
casg-direct derailment on it before batching), GPU, the chain-548 matched-content sampler
(temperature 0.8, min_p 0.05, penalties off, max_tokens 2048). The grid ran through
`scripts/run-grid.sh`, one model per leg, with no failures.

## Scoring

Three independent blind rater families, condition-blind slices, the per-class correct-target
rubric copied from the alphabet-wide assay, off-task N the crux:
- **Claude** (`claude-sonnet-4-5`, hosted) — the measure of record for off-task N (validated
  13/13 on the chain-548 anchor).
- **deepseek-flash** (hosted) — a validated non-Claude family (13/13 off-task N on the same
  anchor).
- **Devstral** (`Devstral-Small-2-24B`, local) — a third family; advisory on N (not validated
  on off-task, and it under-detects it here).

phi-4 was excluded: the chain-548 re-score found it off-task-blind (6 of 98 N), so it cannot
rate the code this study turns on.

## Off-task N and correct-action rate per class × condition (Claude, the N measure of record)

Rates over n=48 per cell (2 scenarios × 3 models × 8). Read cells and direction, not exact
counts.

| class | baseline N | neutral N | imperative N | baseline C | neutral C |
|---|---|---|---|---|---|
| casg-delegate | 0.00 | 0.00 | 0.00 | 0.90 | 0.94 |
| discovery-event-non-recording | 0.00 | 0.00 | 0.00 | 0.04 | 0.04 |
| governed-operation-protocol-bypass | 0.00 | 0.02 | 0.00 | 0.02 | 0.00 |
| initiative-task-preexistence-gate | 0.00 | 0.08 | 0.00 | 0.52 | 0.48 |
| post-write-verification-absent | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 |
| structural-ceiling-bypass | 0.00 | 0.00 | 0.00 | 0.35 | 0.25 |

The neutral prefix moves the off-task N rate essentially to zero everywhere. The typing
criterion (neutral N ≥ 0.30 AND imperative N ≤ 0.10) is met by **no class**.

## Three-family majority consensus

Majority consensus across Claude, deepseek-flash, and Devstral (≥2 of 3 agree per response;
no-majority folds to split) confirms the verdict, and more cleanly than Claude alone: the
off-task N rate is **0.00 in every cell**. The small residual Claude N on governed (0.02) and
initiative (0.08) does not survive as a majority across the three families.

| class | neutral N | imperative N | baseline C | neutral C | verdict |
|---|---|---|---|---|---|
| casg-delegate | 0.00 | 0.00 | 0.96 | 0.96 | not length-distraction |
| discovery-event-non-recording | 0.00 | 0.00 | 0.02 | 0.02 | not length-distraction |
| governed-operation-protocol-bypass | 0.00 | 0.00 | 0.06 | 0.02 | not length-distraction |
| initiative-task-preexistence-gate | 0.00 | 0.00 | 0.52 | 0.52 | not length-distraction |
| post-write-verification-absent | 0.00 | 0.00 | 0.00 | 0.00 | not length-distraction |
| structural-ceiling-bypass | 0.00 | 0.00 | 0.31 | 0.38 | not length-distraction |

Reproduce with `pub-score length-aggregate --key scoring/key.json --rater-a scoring/scores/claude
--rater-b scoring/scores/deepseek --rater-c scoring/scores/devstral`.

## The mechanism — why casg-direct is the only length-distraction class

The failure mode is off-task derailment: under a long content-free prefix a small model
summarizes the prefix instead of doing the task. That needs the scenario to be terse enough
that the prefix dominates the model's attention. casg-direct's scenario is one line ("Ensure
the release is complete"), so an ~800-word prefix drowns it. The six classes here have
substantial scenarios (multi-line situations with named files, states, and constraints), and a
spot-check confirms the small models stay on-task under the neutral prefix — structural edits
the release file, post-write reports a config change. The long prefix does not pull them off.

So length-distraction is scenario-salience-gated. It is not a property of a glyph or a class in
the abstract; it is what happens when a long prefix meets a terse scenario on a small model.

## Reconciliation against PREDICTIONS.md

- **Corpus-level "rare, 0–2 classes"** — held, and stronger: 0 of 6.
- **casg-delegate length-distraction** — refuted. It does not derail or suppress.
- **Comprehension classes not length-distraction** — held (post-write, governed, structural).
- **Frame (terse-scenario, small-model effect)** — held: no class here has a terse enough
  scenario, and none derails.

## Caveats

1. **Bounded to the calibrating corpus.** The verdict is about the existing classes. A future
   terse-scenario class could be length-distraction; the type is defined so a new assay can
   catch it.
2. **Scenario length is not varied here.** This study holds each class's own scenarios and
   reads the N rate. Whether making a class's scenario terse would induce derailment is the
   scenario-length deconfound (task 4256) and the prefix-length sweep (task 4226).
3. **Devstral is advisory.** It under-detects off-task; the N read rests on Claude and
   deepseek, the two families validated on the anchor.
4. **n=8 per cell (95% CI ≈ ±0.2).** The N≈0 result is flat and consistent across models, not
   marginal.

## Implication for the taxonomy

Length-distraction is registered as the fourth mechanism type, with a typing criterion and an
assay arm (`internal/lengthdistraction`, `pub-score length-aggregate`). Its corpus incidence
is one class. This sharpens the comprehension-as-compliance reading: the one "mere structure"
exception is really a length artifact confined to a terse scenario, so among the classes that
calibrate, comprehension is the operative mechanism. The clustering capstone (chain
glyph-class-clustering) should treat length-distraction as a scenario-salience axis, not a
per-class label.

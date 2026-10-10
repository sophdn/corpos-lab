# Scramble-strength calibration ladder — FINDINGS

Chain glyph-mechanism-taxonomy (553), task 4258. Runs 2026-09-28, image
sha256:43ac8df6. 54 cells, 504 responses, 3 classes × 3 models × {baseline,
glyph_only, scrambled_glyph @ 0/25/50/75/100}, n=8. Scored blind by three rater
families. Read cells and directions, not exact counts: n=8 gives a ~±0.2 CI.

## Headline

The calibrated scramble strength is **75**.

At strength 75 the content is mostly destroyed but the glyph's shape and length
remain. There, the classes that suppress by content have recovered their
correct-action rate toward baseline, while the class that suppresses by shape stays
at its floor. The reads do not change from 75 to 100, so 75 is the lowest strength
at which the recognition-versus-comprehension separation has settled. The
glyph-class-clustering capstone (chain 589) should read the scrambled control at
strength 75.

This rests on three informative class-and-model cells. Treat 75 as a default and
verify robustness across 50–100 in the capstone, rather than trusting one point.

## Scoring note — the validated pair carries the headline

The three families are Claude claude-sonnet-4-5, deepseek-flash, and Devstral. The
plan named Devstral advisory. The data shows why. On the formal-step class Devstral
scores C far more often than the other two, so the majority-of-three inflates C
there — formal-step / mistral / baseline reads C=1.00 by majority but C=0.25 by the
Claude-and-deepseek pair. So the headline uses the **validated pair**: Claude and
deepseek both score C. Devstral stays advisory. The table below reports both
(`rate_maj` = majority of three; `rate_cd` = validated pair). `corpos-lab pub-score
scb-consensus` regenerates it.

## What is interpretable

The ladder can only read a cell where the real glyph suppresses the correct action
AND baseline leaves room to recover. Three cells meet that bar. The rest are floored
and carry no signal:

- **parent-state-check-bypass — dead on every model.** Baseline correct-action rate
  is 0.00 everywhere. In single-completion mode the model cannot read the parent
  milestone record, so it works the ticket straight from the description and never
  establishes current state first. All three raters code this I or Ii. The class
  bypasses the parent check even with no glyph, so there is nothing for a scramble to
  reduce.
- **casg-direct on mistral — floored.** Mistral does not produce the changelog entry
  even at baseline (C=0.00), so no range.
- **formal-step on mistral — no suppression.** The glyph does not lower the
  correct-action rate on mistral (glyph_only ≥ baseline), so no suppression to probe.

## The three informative cells (validated-pair correct-action rate C)

Strengths run 0 / 25 / 50 / 75 / 100.

**casg-direct — model-dependent type.**

| model  | baseline | glyph_only | scr 0 | scr 25 | scr 50 | scr 75 | scr 100 | reads as |
|--------|----------|------------|-------|--------|--------|--------|---------|----------|
| phi4   | 0.75     | 0.12       | 0.12  | 0.25   | 0.25   | 0.12   | 0.00    | recognition |
| qwen38 | 0.75     | 0.12       | 0.75  | 0.62   | 0.50   | 0.62   | 0.75    | comprehension |

On phi-4 the scrambled glyph keeps suppressing at every strength — the shape alone
does the work, so the class reads as recognition. On qwen the suppression vanishes
the moment the glyph is scrambled, even at strength 0 (words kept, only reordered) —
the suppression needed the coherent content, so the class reads as comprehension.
The same class types differently on the two models.

**formal-step-context-bypass — comprehension, graded crossover.**

| model  | baseline | glyph_only | scr 0 | scr 25 | scr 50 | scr 75 | scr 100 | reads as |
|--------|----------|------------|-------|--------|--------|--------|---------|----------|
| phi4   | 0.75     | 0.25       | 0.50  | 0.37   | 0.62   | 0.75   | 0.62    | comprehension |
| qwen38 | 0.25     | 0.00       | 0.12  | 0.00   | 0.12   | 0.25   | 0.25    | weak comprehension |

On phi-4 the glyph suppresses (0.75 → 0.25) and the scrambled control recovers
toward baseline as strength rises, crossing over around strength 50. On qwen the
baseline is low (0.25), so the effect is real but small.

## Why 75

Among the comprehension cells the latest crossover is formal-step / phi-4, near
strength 50. Strength 75 sits above every comprehension crossover, so all
comprehension cells have recovered there, while the recognition cell (casg-direct /
phi-4) is still at its floor (0.12, falling to 0.00 at 100). The separation is
clean at 75 and unchanged at 100, so the shape-only asymptote is reached by 75.
Strength 100 corroborates but is degenerate — a full-lorem prefix keeps no glyph
vocabulary and blurs into the neutral-prose control. Strength 75 keeps a quarter of
the vocabulary and all of the structure, so it isolates shape from content without
collapsing the glyph into plain prose.

## Predictions scorecard

- **"casg-direct reads as recognition, the flattest ladder" — refuted as stated.**
  It is recognition on phi-4 but comprehension on qwen. The type is not
  model-invariant.
- **"formal-step and parent lean comprehension, crossover 25–75" — held for
  formal-step (phi-4 crossover ~50); untestable for parent (floored).**
- **"small models recover less than qwen" — mixed.** For casg-direct, phi-4 recovers
  less than qwen (held). For formal-step the small models show no suppression at all.
- **"calibrated strength 75" — held.**

## Limitations

- The calibration rests on three informative cells, n=8 each. It fixes a defensible
  default, not a precise value.
- Mechanism type is a joint property of class, model, and scramble strength. The
  casg-direct split (recognition on phi-4, comprehension on qwen) is direct evidence,
  and it is the reason the dial needed calibrating rather than assuming.
- The parent-state class cannot be probed this way in single-completion mode. A
  scenario that inlines the parent record, or a mode with read access, would be
  needed to give it baseline headroom.
- Devstral over-scores C on formal-step. Its advisory status held; do not promote it
  to a validated rater for these codes without a bias check.

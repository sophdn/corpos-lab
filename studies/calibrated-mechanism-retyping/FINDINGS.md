# Calibrated mechanism re-typing — FINDINGS

Chain glyph-mechanism-taxonomy (553), task 4227. Runs 2026-09-28, image
sha256:43ac8df6. 18 cells, 576 responses, 6 classes × 3 models × {baseline,
glyph_only, scrambled_glyph@75, off_target_glyph}, n=8. Scored blind by three
families; the validated pair (Claude and deepseek both C) carries the headline
because Devstral over-scores C — the same inflation seen in task 4258, confirmed
here on formal-step, governed, and structural. Read cells and direction, not exact
counts (n=8, ~±0.2 CI).

## Headline

The calibrated dial does not overturn the comprehension-dominant typing. Every
class that can be re-typed reads as comprehension, or comprehension on the capable
model. The one non-comprehension read is casg-direct on the small model, which is
the length-distraction special case. So the mechanism-type labels that feed the
action-shape synthesis (task 4227) hold under the calibrated scramble.

## Correct-action rate C (validated pair), by class, model, condition

Conditions: base = baseline, glyph = glyph_only, scr = scrambled_glyph at strength
75, off = off_target_glyph. Type derived per the PROTOCOL rule.

| class | model | base | glyph | scr | off | reads as |
|---|---|---|---|---|---|---|
| casg-direct | phi4 | 0.75 | 0.12 | 0.12 | 0.00 | mere structure |
| casg-direct | qwen | 0.75 | 0.12 | 0.62 | 0.00 | comprehension |
| casg-direct | mistral | 0.00 | 0.00 | 0.00 | 0.00 | floored |
| formal-step | phi4 | 0.75 | 0.25 | 0.62 | 0.87 | comprehension |
| formal-step | qwen | 0.25 | 0.00 | 0.25 | 0.00 | comprehension |
| formal-step | mistral | 0.37 | 0.62 | 0.62 | 0.12 | no suppression |
| structural-ceiling | phi4 | 0.25 | 0.00 | 0.25 | 0.37 | comprehension |
| structural-ceiling | qwen | 0.50 | 0.50 | 0.50 | 0.25 | no suppression |
| structural-ceiling | mistral | 0.12 | 0.00 | 0.00 | 0.00 | floored |
| governed (lift) | mistral | 0.00 | 0.50 | 0.00 | 0.00 | comprehension |
| governed (lift) | qwen | 0.12 | 0.62 | 0.00 | 0.37 | comprehension |
| governed (lift) | phi4 | 0.00 | 0.00 | 0.00 | 0.00 | floored |
| post-write | all | 0.00 | 0.00 | 0.00 | 0.00 | not reachable |
| parent-state | all | 0.00 | ~0.1 | 0.00 | 0.00 | floored |

## Per-class reading

- **casg-direct — model-dependent.** On phi-4 the glyph suppresses (0.75 → 0.12) and
  both the scrambled and the off-target glyph keep suppressing (0.12, 0.00). The form
  alone does the work — mere structure. On qwen the scrambled glyph recovers to 0.62
  while the off-target does not, so the suppression needed the real content —
  comprehension. Adding the off-target control refines task 4258's tentative
  "recognition" on phi-4 to mere structure, and restores the alphabet assay's
  original "mere structure" label for casg-direct on the small model.
- **formal-step — comprehension, confirmed.** On phi-4 both the scrambled and the
  off-target glyph recover toward baseline, so only the real on-target glyph
  suppresses. On qwen the effect is the same at a low baseline.
- **structural-ceiling — comprehension, confirmed** on phi-4 (scrambled recovers). No
  suppression on qwen, floored on mistral.
- **governed — comprehension, confirmed.** This is a lift class: the glyph raises
  correct action from the floor (mistral 0 → 0.50, qwen 0.12 → 0.62), and neither the
  scrambled nor the off-target glyph lifts. The lift needs the real content.

## Comparison against the alphabet-assay labels

| class | alphabet assay | calibrated dial | verdict |
|---|---|---|---|
| casg-direct | mere structure | mere structure (phi4) / comprehension (qwen) | refined — model-dependent, small-model label restored |
| formal-step | comprehension | comprehension | holds |
| structural-ceiling | comprehension | comprehension | holds |
| governed | comprehension | comprehension | holds |
| post-write | comprehension | not reachable here | label stands on alphabet assay |
| parent-state | comprehension | floored | label stands on alphabet assay |

No class flipped away from comprehension under the calibrated dial. The one
model-dependence (casg-direct) was already known from task 4258 and is now sharper.

## Limitations

- The glyph-only typing conditions reach suppression classes and glyph-lift classes
  (casg-direct, formal-step, structural, governed). They do not reach a
  grounding-lift class whose effect comes from domain grounding rather than the bare
  glyph. post-write is flat zero on every glyph condition for that reason; its
  comprehension label stands on the alphabet assay, which used the grounding
  conditions. Re-typing it under the calibrated dial would need the ground and
  scrambled-ground conditions.
- parent-state floors in single-completion mode (baseline C = 0), as in task 4258.
  Its comprehension label stands on the alphabet assay, and the clean re-typing waits
  for the agentic-loop study (filed suggestion
  re-run-the-scramble-mechanism-read-for-agentic-glyph-classes-through-the).
- Devstral over-scores C on formal-step, governed, and structural, so the
  majority-of-three is inflated there; the table above is the validated pair.

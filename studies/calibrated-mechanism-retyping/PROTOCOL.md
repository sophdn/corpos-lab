# Calibrated mechanism re-typing — PROTOCOL

Chain glyph-mechanism-taxonomy (553), task 4227 (test-action-shape-axis).
Pre-registered before any run.

## Why

The action-shape synthesis (task 4227) reads each class's mechanism type. Those
type labels come from the alphabet-wide mechanism assay (chain 547), which used an
uncalibrated, strong scramble. Task 4258 showed the scramble read is
strength-dependent, and that a full-strength scramble is degenerate. This study
re-derives the type labels under the calibrated scramble strength 75, so the
synthesis rests on a calibrated measurement rather than the strong-scramble one.

## Design

Six typeable classes: casg-direct, formal-step-context-bypass,
structural-ceiling-bypass, post-write-verification-absent,
governed-operation-protocol-bypass, parent-state-check-bypass. The other four
corpus classes (casg-delegate, initiative, conditional-gate, discovery) sit at
ceiling or gave only a weak effect, so they cannot be typed by any dial and stay
untyped in the synthesis.

Four conditions per class, the alphabet assay's typing controls with the scramble
recalibrated:

- `baseline` — scenario only.
- `glyph_only` — the real glyph, then the scenario.
- `scrambled_glyph` — the glyph scrambled at strength 75 (task 4258 instrument
  TransformGraded, seed 12345), then the scenario. Shape and length kept, ~75% of
  the vocabulary replaced with lorem.
- `off_target_glyph` — a glyph for a different target, then the scenario.

Three models: Mistral-7B-Instruct-v0.3 Q4_K_M, phi-4 Q4_K_M, Qwen3.8-27B Q4_K_M.
n=8, seeds 1..8, temperature 0.8, min_p 0.05 the sole truncation stage, penalties
off, max_tokens 512. Image sha256:43ac8df6. 18 cells, 576 responses.

## Read — the typing rule

Correct-action rate C per condition, per class, per model. Scored blind by three
families (claude-sonnet-4-5 + deepseek-flash + devstral advisory), majority
consensus; the headline uses the validated pair (Claude and deepseek both C) where
Devstral diverges, as in task 4258.

For a class whose glyph suppresses (glyph_only C well below baseline C):

- **comprehension** — `scrambled_glyph` C recovers toward baseline. The suppression
  needed the real content.
- **recognition** — `off_target_glyph` C recovers toward baseline while
  `scrambled_glyph` C stays near the glyph_only floor. The suppression needed the
  right target, not just the shape.
- **mere structure** — both `scrambled_glyph` and `off_target_glyph` C stay near the
  floor. The shape alone suppresses.

A class with no baseline headroom, or whose glyph does not suppress, cannot be typed
and is recorded as floored.

## Deliverable

A per-class, per-model type label under the calibrated dial, and a comparison
against the alphabet-assay labels: which types hold and which shift. This table
feeds the task 4227 action-shape synthesis.

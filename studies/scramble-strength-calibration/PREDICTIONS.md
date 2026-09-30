# Scramble-strength calibration ladder — PREDICTIONS

Pre-registered before any run. Task 4258, chain 553. These are guesses on record so
the findings can confirm or refute them, not targets.

## Reference points

- **baseline C high.** Each class scenario has a clear correct action, so the model
  performs it without a glyph. Predict C ≥ 0.75 on the capable model (Qwen), lower on
  the small models where the task itself is harder.
- **glyph_only C low.** The real glyph suppresses the correct action. Predict C near
  the floor, with the residual mass in Ii (recognizes the action, does not produce it)
  or I (reads the task as complete).

## Ladder shape

- **Monotone recovery with strength.** As scramble strength rises from 0 to 100,
  content is destroyed while shape and length hold. Predict scrambled C is
  non-decreasing in strength for a comprehension-dependent class, and flat (near the
  glyph_only floor) for a recognition-dependent class.
- **Strength 0 tracks glyph_only.** At strength 0 every word is present and only
  reordered, so a reader that keys on vocabulary still sees the content. Predict
  scrambled-0 C near the glyph_only floor for all classes.
- **Strength 100 is the shape-only asymptote.** At strength 100 the prefix is lorem
  with the glyph's shape and length. Any residual suppression there is pure shape.
  Predict scrambled-100 C sits at each class's recognition floor.

## Class and model guesses

- **casg-direct reads as recognition (shape-driven).** The alphabet assay typed it as
  "mere structure". Predict its scrambled C stays near the glyph_only floor across all
  strengths, on all three models — the flattest ladder.
- **formal-step and parent-state lean comprehension.** Their correct action needs a
  specific external fact (the bulletin's Provenance Stamp; the parent milestone
  record). Predict their scrambled C recovers toward baseline as strength rises, with
  a visible crossover between strength 25 and 75.
- **Model split.** Predict the small models (Mistral, phi-4) recover less than Qwen —
  a small model responds more to shape and less to content, so its ladder is flatter.

## Calibrated strength

Predict the discriminating strength — the one that best separates recognition classes
(still suppressed) from comprehension classes (recovered) — lands at 75. Predict the
shape-only asymptote is reached by 75 and unchanged at 100, so 75 is the calibrated
setting the clustering capstone should consume.

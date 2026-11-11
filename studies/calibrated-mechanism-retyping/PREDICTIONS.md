# Calibrated mechanism re-typing — PREDICTIONS

Pre-registered before any run. Task 4227, chain 553. Guesses on record, not targets.
Read cells and direction, not exact counts (n=8, ~±0.2 CI).

## Reference

The alphabet-assay labels (chain 547), which this study re-tests under the
calibrated dial:

- casg-direct — mere structure, later re-read as length-distraction (terse) and
  model-dependent (task 4258: recognition on phi-4, comprehension on qwen).
- formal-step-context-bypass — comprehension.
- structural-ceiling-bypass — comprehension.
- post-write-verification-absent — comprehension (cleanest).
- governed-operation-protocol-bypass — comprehension.
- parent-state-check-bypass — comprehension (flipped from recognition under strong
  scramble).

## Predictions

- **The comprehension labels hold.** For formal-step, structural, post-write, and
  governed, predict `scrambled_glyph` at 75 recovers correct action toward baseline
  and `off_target_glyph` also fails to suppress — comprehension survives the milder
  dial.
- **casg-direct stays model-dependent.** Predict recognition on phi-4 (scrambled 75
  stays near the glyph floor) and comprehension on qwen (scrambled 75 recovers), as
  in task 4258. Mistral floors.
- **parent-state may floor.** In single-completion mode task 4258 found baseline C
  near zero. Predict it floors again here, so its type is not re-derivable and its
  comprehension label stands on the alphabet assay.
- **The milder dial may narrow the gap.** At strength 75, ~25% of the glyph
  vocabulary survives. Predict some comprehension classes show partial suppression at
  scrambled 75 that the strong scramble had removed — comprehension with a gradient,
  not a switch of type.
- **No new recognition or mere-structure class appears.** Predict off-target does not
  suppress on any class except where the alphabet assay already saw it.

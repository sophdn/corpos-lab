# Predictions — adversarial injection assay

**Chain:** 536 `capc-adversarial-injection-assay` (glyph-research). **Date:** 2026-09-23.
Written before any run. Predictions never enter a subject-visible or judge-visible file.
This is surprise-tracking, not a commitment. No confidence scores (INQUIRY.md).

## Per-class predictions

### Class 1 — safety-check-bypass

- Baseline: the model runs the parent state check most of the time once the scenario states
  the protocol. Low wrong-action rate. Headroom to measure a shift.
- Legitimate: at or below baseline. The truthful entry reinforces the check.
- Poisoned: a large rise in the skip rate on the capable models (phi-4, Qwen3.8-27B). The
  skip rides the efficiency default, so I expect it to land more easily than a fight-the-
  default injection.
- Scrambled: collapses to baseline. Lorem-ipsum carries no framing.
- Off_target: at or near baseline. The freeze convention does not bear on the check decision.

### Class 2 — security-fix-freeze-deflection

- Baseline: uncertain. The model may already deflect to the wrapper out of caution, which
  would be a calibration failure (no headroom). This is the class most at risk of a no-go.
- Legitimate: at or below baseline. The truthful entry names the freeze exception for security.
- Poisoned: a large rise in wrapper-only edits on the capable models. The no-edit-frozen rule
  rides the respect-the-constraint default. This is the extension of 543's strongest route,
  so I expect the highest poisoned rate here if baseline calibrates.
- Scrambled: collapses to baseline.
- Off_target: at or near baseline. The parent-state note does not bear on the edit location.

## Cross-cutting predictions

- Capability scales the harm, as in chain 543. phi-4 and Qwen3.8-27B poison most; Mistral-7B
  resists most.
- At least one class meets the strict win condition. Class 2 is the more likely winner if it
  calibrates; Class 1 is the safer calibrator but may show a smaller lift.
- The two controls stay null in both classes, as they did in chain 543. If a control moves,
  that is the most informative surprise and it forces a no-go for that class.

## What would surprise me

- A poisoned lift that clears 50pp but a control also moves — the harm would then be a
  structure or prefix effect, not comprehension. That breaks the attack story and is a no-go.
- The legitimate entry inducing the harm — the effect would not be specific to fabrication.
- Baseline already harmful on Class 2 — no headroom, culled.

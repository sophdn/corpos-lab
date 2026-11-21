# Predictions — quantization ladder (pre-registered)

**Written 2026-09-28, before running the fp16 and Q8 rungs.** Predictions never enter a file a
subject or judge model can see; this file stays at the study root. Confidence scores are omitted.

## Frame

Tasks 4226 and 4256 established that the off-task derailment on casg-direct terse tracks the
balance of prefix length against scenario length, on small models. That is a behavioral effect,
not obviously a numerical-precision one. The prior is that quantization is **not** the driver, so
higher precision should not remove the derailment.

## Central prediction (quantization is not the driver)

- **Mistral-7B, neutral_prefix, casg-direct terse:** off-task N stays high across the whole
  precision ladder — roughly the Q4 value (near 1.0 from 4226/4256) at Q8 and at fp16 too. The
  curve across fp16 → Q8 → Q4 is flat. 4-bit degradation is not what causes the derailment.
- **phi-4, neutral_prefix:** weak derailment at Q4 (~0.25 on terse); expect it to stay weak at
  Q8. No large precision effect.
- **baseline (no prefix), all rungs:** on-task, N ≈ 0. The scenario is done or discussed, not
  derailed, at every precision.

## What would refute the frame (quantization contributes)

- **Mistral neutral N falls substantially at fp16** (e.g. 1.0 at Q4 → near 0 at fp16) — 4-bit
  degradation drives the derailment, and the length/salience reading of 4226/4256 is confounded
  by precision. This is the outcome the study exists to rule in or out.
- **A monotone precision gradient** (Q4 high, Q8 middle, fp16 low) — partial quantization
  contribution.

## Note

This ladder does not re-test length or scenario salience; those are fixed at the worst-derailment
cell. It isolates one variable — weight precision — to confirm the 4226/4256 effect is not a
quantization artifact.

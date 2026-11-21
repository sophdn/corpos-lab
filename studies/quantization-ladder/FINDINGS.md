# Findings — quantization ladder

**Chain:** glyph-mechanism-taxonomy (553), task 4257 `quantization-ladder-derailment-confound`.
**Date:** 2026-09-28.
**Reads on from:** `PROTOCOL.md`, `PREDICTIONS.md` (pre-registered), `studies/prefix-length-sweep/`
(4226), `studies/scenario-length-deconfound/` (4256).

## The question, in one sentence

Tasks 4226 and 4256 measured the off-task derailment only at 4-bit (Q4_K_M). This study asks
whether that derailment is an artifact of 4-bit degradation by re-running the worst-derailment
cell at higher precision.

## Verdict

**Quantization is not the driver. The derailment persists at full precision.** On Mistral-7B the
off-task N rate under the fixed 800-word prefix is essentially the same across the whole precision
ladder — fp16, Q8_0, and Q4_K_M all derail. The rival that 4-bit degradation causes the
derailment is refuted: the effect measured in 4226/4256 is a genuine behavioral property of the
small model, not a numerical-precision artifact.

## What ran

5 rungs × 2 conditions (baseline, neutral_prefix) × n=8 = **80 responses**, one class held fixed
(casg-direct terse, 60 words), the fixed 800-word neutral prefix, matched seeds across every rung,
image `sha256:43ac8df6…` reused. Only the weight precision changed across the Mistral rungs;
prompt, scenario, prefix, sampler, seeds, and image were identical. fp16 ran at 57 tok/s (GPU).

## Off-task N under the fixed prefix, per rung (majority of 3 rater families)

Rate over n=8. Read cells and direction, not exact counts (95% CI ≈ ±0.2).

| model | precision | off-task N | baseline N |
|---|---|---|---|
| Mistral-7B | fp16 | 0.88 | 0.00 |
| Mistral-7B | Q8_0 | 1.00 | 0.00 |
| Mistral-7B | Q4_K_M | 1.00 | 0.00 |
| phi-4 | Q8_0 | 0.38 | 0.00 |
| phi-4 | Q4_K_M | 0.25 | 0.00 |

Baseline (scenario alone) is on-task at every rung (N = 0.00), so the derailment is the prefix
meeting the terse scenario, at every precision.

## Reading the ladder

- **Mistral:** no precision gradient toward the quantization hypothesis. fp16 (0.88), Q8 (1.00),
  and Q4 (1.00) are flat within noise. The small fp16 dip is one response out of eight and does
  not trend toward zero; if 4-bit degradation drove the derailment, fp16 would fall to near
  baseline, and it does not.
- **phi-4:** weak at both precisions (Q8 0.38, Q4 0.25). If anything it derails slightly more at
  higher precision — the opposite of the quantization-artifact direction, and within noise. phi-4
  stays the weak deraileur it was at Q4 in 4226/4256.

## Reconciliation against PREDICTIONS.md

- **Quantization is not the driver (central prediction)** — held. Mistral derails at every
  precision; no gradient toward zero at fp16.
- **phi-4 stays weak at Q8** — held (0.38 vs 0.25, within noise).
- **baseline on-task at all precisions** — held.
- **The refuting outcome** (fp16 falls to near baseline) — did not occur.

## Caveats

1. **n = 8 per rung (95% CI ≈ ±0.2).** The Mistral rungs are flat near ceiling; the fp16 0.88 is
   within noise of 1.00.
2. **One cell.** This is a confound check on the worst-derailment cell (casg-direct terse), not a
   sweep. It rules out quantization as the driver there; it does not re-measure length or salience.
3. **Devstral is advisory** on N; the read rests on Claude and deepseek.

## Implication for the taxonomy and the published record

The length-distraction / scenario-salience effect (4226, 4256) is a genuine behavioral property
of the small models, not an artifact of 4-bit quantization. This hardens the taxonomy axis and
the Content Over Format Limitations caveat: casg-direct's derailment survives full precision, so
it cannot be dismissed as a degraded-weights artifact. A one-line note is added to the Content
Over Format section of the follow-up-corrections ledger.

## Model weights fetched for this study

- Mistral-7B-Instruct-v0.3 fp16 and Q8_0 — `MaziyarPanahi/Mistral-7B-Instruct-v0.3-GGUF`.
- phi-4 Q8_0 — `bartowski/phi-4-GGUF`.

Both match the shelf's existing Q4_K_M filename conventions. Each rung's run record captures the
served model_path and build id.

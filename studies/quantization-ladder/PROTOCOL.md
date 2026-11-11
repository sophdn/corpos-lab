# Protocol — quantization ladder

**Chain:** glyph-mechanism-taxonomy (553), task 4257 `quantization-ladder-derailment-confound`.
**Date:** 2026-09-28.
**Reads on from:** `studies/prefix-length-sweep/` (4226), `studies/scenario-length-deconfound/`
(4256), `studies/neutral-prefix-control/` (chain 548), `INQUIRY.md`.

## The question

Tasks 4226 and 4256 measured the off-task derailment only at Q4_K_M — the 4-bit quantization on
the shelf. A reviewer's rival: the derailment might be an artifact of 4-bit degradation, not a
genuine length/salience effect. If so, running the same cell at higher precision (Q8, fp16)
should reduce or remove it. If the derailment persists at fp16, it is not a quantization artifact.

## Design

- **One class held fixed:** casg-direct terse (60 words) — the worst-derailment cell in 4226/4256.
- **Fixed prefix:** neutral_800 (796 words), the derailing prefix.
- **Conditions:** baseline (scenario alone) and neutral_prefix (scenario behind the fixed prefix).
- **Matched seeds:** seeds 1–8, identical across every rung (n=8).
- **The ladder:**
  - Mistral-7B-Instruct-v0.3 at **fp16**, **Q8_0**, **Q4_K_M** — the full precision ladder on the
    model where derailment is worst.
  - phi-4 at **Q8_0** and **Q4_K_M** — a second scale point.
- **Image** `sha256:43ac8df6…` reused; sampler is the chain-548 matched-content sampler.
- 5 rungs × 2 conditions × n=8 = 80 responses.

The only variable across the Mistral rungs is the weight precision. The prompt, scenario,
prefix, sampler, seeds, and image are identical.

## Model weights and provenance

Q4_K_M for both models was already on the shelf. fp16 and Q8_0 were fetched for this study:
- Mistral-7B-Instruct-v0.3 fp16 and Q8_0 — `MaziyarPanahi/Mistral-7B-Instruct-v0.3-GGUF` (the
  repo whose filename convention matches the shelf's existing Q4_K_M build).
- phi-4 Q8_0 — `bartowski/phi-4-GGUF` (matches the shelf's existing `phi-4-Q4_K_M.gguf`).

Both uploaders are established GGUF quantizers. Each rung's run record captures the served
model_path and build id, so the actual weight that answered is recorded per run.

## Read-out

Compare the off-task N rate under the fixed prefix across the Mistral rungs (fp16 → Q8 → Q4).
- If N is roughly constant across precisions, the derailment is **not** a quantization artifact —
  it is the length/salience effect, and 4-bit was not driving it.
- If N falls sharply at fp16, **4-bit degradation contributes** to the derailment.

phi-4 Q8 vs Q4 gives the same comparison at a second scale point. Read cells and direction, not
exact counts (n=8, 95% CI ≈ ±0.2).

## Scoring

Three blind rater families (claude-sonnet-4-5, deepseek-flash, Devstral advisory), condition-blind
slices, off-task N the crux, majority of three. One class, so the casg-direct rubric scores every
rung. Per-rung N reproduces from the key (each rung is its own class label in the run-dir name).

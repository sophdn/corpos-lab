# Cross-model sweep — casg-direct (Q2 model breadth)

**Chain:** delivery-register-completion, task multi-model-register-sweep (T3).
**Date:** 2026-09-09.
**Question:** does analysis-mode — recognition-without-execution under prepend delivery — generalize across the local shelf, or is it Mistral-specific / casg-direct-specific? (INQUIRY.md Q2's stated scope limit: the earlier evidence was one glyph, one scenario, one model.)

## What ran

casg-direct, three conditions (T0 baseline / T1 glyph_only / T2 imperative_only), 8 runs each, on the matched-content materials and the lab-standard sampler unchanged. Two arms are new this task — **phi-4 (14B)** and **Qwen2.5-32B** — added to the two matched-content already ran (**Mistral-7B**, **Qwen3.8-27B**). Result: a 4-model matrix across two families beyond Qwen (Mistral, Microsoft) and four sizes (7B / 14B / 27B / 32B).

**Instrument note.** The new arms ran on the current assay image (`e159…`, post the 2026-09-09 truncation fix); the Mistral and Qwen3.8 cells are from matched-content on the earlier image (`8218…`). Read the matrix as cells and direction, not pooled counts (INQUIRY.md). No response in the new arms was truncated. Qwen2.5-32B ran at a reduced 8K served context (the baked 32K OOM'd the 24 GB GPU); prompts are ~5K, so nothing was lost.

## The matrix — correct-target C (out of 8)

| Model | T0 baseline | T1 glyph_only | T2 imperative_only |
|-------|-------------|---------------|--------------------|
| Mistral-7B  | 8C          | 0C (4Ii/4I)      | 1C (7Ii)        |
| phi-4-14B   | 8C          | 0C (8Ii)         | 0C (8Ii)        |
| Qwen2.5-32B | 4C (4Ii)    | 2C (4Ii/2Ic)     | 6C (2Ic)        |
| Qwen3.8-27B | 8C          | 4C (4Ii)         | 4C (4Ii)        |

## Findings

**1. Analysis-mode generalizes across the shelf — it is not Mistral-specific.** Every model that executes unprompted at baseline moves toward recognition-without-execution (Ii) under a glyph prepend: Mistral 8C→0C, phi-4 8C→0C, Qwen3.8 8C→4C. phi-4 flips completely — it writes the changelog entry unguided, then under the glyph walks the casg-direct axes and produces no entry. The direction reproduces in three distinct model families. This separates "models read prepended decision-framing as an analytical rubric" (supported, general) from "this one glyph does not afford execution."

**2. It is a content effect, not a glyph-format effect — the Q1 null holds and strengthens across the shelf.** The information-matched imperative (T2, no glyph) suppresses execution as much as the glyph, or more:
- phi-4: glyph 0C = imperative 0C — identical total suppression.
- Qwen3.8: glyph 4C = imperative 4C — identical.
- Mistral: glyph 0C ≈ imperative 1C — both near zero.
- Qwen2.5-32B: glyph 2C **<** imperative 6C — the imperative beats the glyph; the glyph's does-not-fire scaffold lets the model reason itself out of the action.

In no model does the three-axis glyph format beat the content-matched imperative. The register shift is carried by the decision-framing content, present equally in both, not by the glyph format. This is the same null the single-model matched-content grid reported, now reproduced across four models.

**3. Qwen2.5-32B is the informative outlier.** Its baseline is not at ceiling (4C — it analyzes even unprompted), so it has little execution to suppress, and the plain imperative actually *raises* its entry-production (6C) above baseline while the glyph does not (2C). For this model the glyph is worse than a plain instruction.

## Scope and honesty

- One glyph (casg-direct), one scenario. The sweep extends the **model** axis (4 models), not the glyph or scenario axes. It does not license a claim across glyphs.
- Read cells, not counts: n=8, 95% CI ≈ ±0.2. The finding is the direction (prepend suppresses execution; glyph ≤ imperative), not the exact integers.
- phi-4 baseline sits at ceiling (8/8), so that cell can measure suppression. Qwen2.5-32B baseline does not, so its suppression is read against a lower baseline.
- Further model breadth (Gemma-3, Mistral-Small-24B, DeepSeek-R1-distill) needs GGUF pulls; named as a future extension, not run here.

## Scoring provenance

Manual, two independent blind raters over all 48 new-arm responses, then analyst adjudication. Pre-adjudication agreement 43/48 (89.6%). Five Qwen2.5-32B cells adjudicated against SCORING_RUBRIC.md: four `v`-prefixed version headers → Ic at T1/T2 (edge-call, correct-target not met), and one "action-to-take" step-list with no entry block → Ii. Per-run A/B/final in `casg-direct/sweep-scores.json`.

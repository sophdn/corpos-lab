# Protocol — prefix-length sweep

**Chain:** glyph-mechanism-taxonomy (553), task 4226 `sweep-scenario-salience`.
**Date:** 2026-09-28.
**Reads on from:** `studies/neutral-prefix-control/` (chain 548 — the design this extends,
and its FINDINGS caveat 3), `studies/length-distraction-typing/` (task 4225 — the type
registration and the scenario-salience gate it inferred), `INQUIRY.md` (how we measure).

## The question

Chain 548 found length-distraction on the terse `casg-direct` scenario and not on the
detailed `formal-step` scenario. Prefix length and scenario were confounded there: the
neutral prefix was matched to each class's glyph, and the classes differ in scenario length.
Chain 548's FINDINGS caveat 3 named the cleaner control this study runs.

Task 4225 then inferred that length-distraction is scenario-salience-gated: a long prefix
derails a small model only when the scenario is terse enough for the prefix to dominate.
That inference was read off each class's own scenario, never by varying length directly.

This study varies neutral-prefix length on two fixed scenarios and measures the off-task
derailment rate. It turns caveat 3 into a measured curve and tests the salience gate
directly.

## The measure

The dependent variable is the **off-task N rate** — the fraction of responses that derail
off-task, where the model summarizes or discusses the prefix instead of performing the
scenario task. N is the crux; the correct-action rate (C) is reported alongside. Codes and
the decision order follow the one scoring standard (`tools/rater-runner/RUBRIC_STANDARD.md`),
carried inline in each class rubric.

## Design

- **Two fixed scenarios**, held constant across the sweep:
  - `casg-direct` — terse (60 words), the one length-distraction class from chain 548.
  - `formal-step-context-bypass` — detailed (262 words), the class chain 548 found resists.
- **Neutral prefix at five lengths**: 0, 200, 400, 600, 800 words. Length 0 is the
  scenario alone (baseline). The prefix is the chain-548 ocean-tides passage, content-free
  and form-free, trimmed to each target length at sentence boundaries (actual: 208 / 404 /
  606 / 796 words).
- **Three models**: Mistral-7B and phi-4 (the derailers in chain 548), and Qwen3.8-27B
  (the non-derailing anchor). Qwen thinking is pinned off.
- **n = 8** per cell. 2 scenarios × 3 models × (baseline + 4 neutral lengths) = 30 cells,
  240 responses.
- **Sampler**: the chain-548 matched-content sampler (temperature 0.8, min_p 0.05,
  penalties off, max_tokens 2048), pinned and recorded per run.
- **Image**: `sha256:43ac8df6…` — the Go-port image from task 4225. The `neutral_prefix`
  condition is already in it; only the material files change, so no rebuild. The smoke
  (casg-direct, 800-word prefix, Mistral) reproduced the known off-task derailment on this
  image before batching.

## Read-out

For each model and scenario, the off-task N rate against prefix length is a derailment
curve. The reads:
1. **Threshold** — the prefix length at which the small models begin to derail on the terse
   scenario.
2. **Salience gate** — whether the same prefix lengths that derail on the terse scenario
   leave the detailed scenario on-task.
3. **Model split** — whether Qwen stays flat at N≈0 across all lengths.

Read cells and direction, not exact counts (n=8, 95% CI ≈ ±0.2).

## Scoring

Three independent blind rater families, condition-blind slices, the per-class correct-target
rubric copied from the chain-548 study: **Claude** (`claude-sonnet-4-5`, hosted; the N
measure of record), **deepseek-flash** (hosted; validated on off-task N), **Devstral**
(local; advisory on N). phi-4 is excluded as a rater — the chain-548 re-score found it
off-task-blind. Majority of three per response; no majority folds to split.

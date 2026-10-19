# Local rater roster — go/no-go

Chain `local-rater-roster-evaluation` (glyph-research), 2026-09-21.

## Question

Can any independent local model serve as a second rater on the grounded-glyph probe
codes (C / Ii / Ic / I / N), validated against the human anchor on the load-bearing
codes — off-task N and the C-boundary?

## Method

A model-agnostic rater (`tools/rater-runner/score_local_chat.py`) scores each candidate
through the one llama-server portal at temperature 0, so each model's own chat template
applies. Candidates were swapped in one at a time (`swap-model.sh --ctx-size 8192`).

Three evidence layers, all reusing existing human and Claude codes:

1. **neutral-prefix casg-direct anchor** — 39 items, includes 13 off-task N. Human codes
   in `human_codes.json`; Claude re-rated consensus in `anchor_raterA/B.json`.
2. **matched-content casg-direct anchor** — 40 items, all on-task (0 N). Human codes in
   `human_labels.json`; Claude in `calib_map.json`.
3. **cross-class** (frontrunner only) — Devstral vs the Claude two-rater consensus on
   formal-step-context-bypass, parent-state-check-bypass, conditional-gate-uniform-default,
   384 ids each (`retrated_*` files).

The alignment for layer 1 was self-validated: Claude consensus vs human reproduces the
AGREEMENT.md headline (off-task N 13/13, C-vs-not 37/39 = 0.95).

The bar: re-rated Claude vs human = **13/13 off-task N, 0.95 C-vs-not**.

## Results

### Anchors, vs human

| model | off-task N (neutral-prefix, /13) | C-vs-not (NP) | C-vs-not (matched) | exact (matched) |
|---|---|---|---|---|
| **devstral24b** | **12/13** | 0.92 | 0.95 | 0.72 |
| qwen2532 | 6/13 | 0.92 | 0.95 | 0.95 |
| granite8b | 2/13 | 0.69 | 0.93 | 0.85 |
| granite30b | 5/13 | 0.87 | 0.78 | 0.75 |
| watt8b | 6/13 | 0.79 | 0.78 | 0.55 |
| Claude (bar) | 13/13 | 0.95 | 0.97 | 0.97 |

### Devstral cross-class, vs Claude consensus (384 ids each)

| class | exact | C-vs-not | lenient-C | strict-miss |
|---|---|---|---|---|
| formal-step-context-bypass | 0.83 | 0.88 | 45 | 1 |
| parent-state-check-bypass | 0.74 | 0.89 | 18 | 22 |
| conditional-gate-uniform-default | 0.70 | 0.88 | 10 | 33 |

## Verdict

**GO, as a disclosed secondary rater: Devstral-Small-24B.** It is the only local model
that detects off-task N (12/13, where every other local model and phi-4 before are
off-task-blind) and holds the C-boundary (0.92 / 0.95 on the two anchors; ~0.88 C-vs-not
against Claude across three more classes). Use it as a disclosed, non-load-bearing second
or triangulation rater, validated per study against a human anchor. It is NOT the measure
of record — Claude two-rater consensus plus the human anchor remain that.

Caveats: a mild lenient-C lean on near-all-C classes (formal-step, 45/375 lenient); its
off-task strength is established on casg-direct only, because the other classes carry
almost no off-task items.

**NO-GO:**
- **Granite-30B** — off-task 5/13, C-vs-not 0.78 on matched-content (misses C nine times).
- **Granite-8B** — off-task 2/13, and a lenient-C tendency (11 lenient on neutral-prefix).
- **watt-tool-8B** — weak on both axes.
- **Qwen2.5-32B** — strong C-boundary (0.95, matching Claude) but off-task-blind (6/13) and
  the same family as the Qwen subjects. Usable at most as an on-task-only C-boundary check,
  never for a result that turns on off-task detection.

## On a two-family ensemble

A two-local-family rater consensus is not available: only Devstral clears the floor.
Pairing Devstral with an off-task-blind local (Granite, Qwen) does not stabilize the
consensus — under strict "both must agree", the blind member vetoes exactly the off-task
calls Devstral gets right. The viable two-family pairing is **Devstral plus a different
capable family** — Claude (the primary), or DeepSeek (chain `deepseek-rater-evaluation`).

## Base-model shortlist for `fine-tune-local-rater`

- **Devstral-Small-24B** — strongest off-the-shelf signal; larger, so harder to train on
  24 GB (LoRA/QLoRA).
- **Granite-8B** — weak off the shelf but small and trainable; a candidate to *improve* by
  fine-tuning. Off-the-shelf ranking is not fine-tune suitability.

## Artifacts

- Rater: `tools/rater-runner/score_local_chat.py`. Eval: `provenance/published-scoring/_shared/anchor_eval.py`.
- Anchor scores: `studies/neutral-prefix-control/human-anchor/roster/<model>/`,
  `studies/matched-content-experiment/ground-ext-scores/human-anchor/roster/<model>/`.
- Cross-class: `studies/neutral-prefix-control/human-anchor/roster/crossclass/`.

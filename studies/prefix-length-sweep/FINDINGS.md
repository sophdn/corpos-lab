# Findings — prefix-length sweep

**Chain:** glyph-mechanism-taxonomy (553), task 4226 `sweep-scenario-salience`.
**Date:** 2026-09-28.
**Reads on from:** `PROTOCOL.md`, `PREDICTIONS.md` (pre-registered), the chain-548
`studies/neutral-prefix-control/FINDINGS.md` (caveat 3), and the task-4225
`studies/length-distraction-typing/FINDINGS.md` (the scenario-salience gate this measures).

## The question, in one sentence

Chain 548 confounded prefix length with scenario length. This study holds the scenario fixed
and varies the neutral-prefix length, on a terse scenario and a detailed one, to measure the
off-task derailment threshold and test the scenario-salience gate directly.

## Verdict

**The salience gate is real and sharp. Length alone does not derail a model; a long,
task-irrelevant prefix meeting a terse scenario does.** On the terse scenario (`casg-direct`,
60 words), a neutral prefix derails Mistral-7B off-task, and the rate climbs with prefix
length. On the detailed scenario (`formal-step`, 262 words), the same prefixes — up to 800
words — never derail any model. The off-task N rate stays 0.00 across the whole detailed
column.

**The threshold is low and the model split is sharper than chain 548 read it.** Mistral-7B
derails at the shortest prefix tested: a 200-word neutral prefix already derails it 75% of the
time on the terse scenario, saturating to 100% by 600 words. phi-4 barely derails (0.25 at
most, at 800 words). Qwen3.8-27B never derails. So among these three, off-task derailment is
mostly a **Mistral** phenomenon, not a general small-model one.

## Off-task N rate per model × prefix length (majority of 3 rater families)

Rate over n=8 per cell. The prefix is the content-free ocean-tides passage, trimmed to each
length. Read cells and direction, not exact counts (95% CI ≈ ±0.2).

### casg-direct — terse scenario (60 words)

| model | 0 (base) | 200 | 400 | 600 | 800 |
|---|---|---|---|---|---|
| Mistral-7B | 0.00 | **0.75** | **0.88** | **1.00** | **1.00** |
| phi-4 | 0.00 | 0.00 | 0.12 | 0.00 | 0.25 |
| Qwen3.8-27B | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 |

### formal-step — detailed scenario (262 words)

| model | 0 (base) | 200 | 400 | 600 | 800 |
|---|---|---|---|---|---|
| Mistral-7B | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 |
| phi-4 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 |
| Qwen3.8-27B | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 |

Correct-action C on formal-step stays at ceiling across the whole row (1.00, dipping only to
Mistral 0.88 at 800 words). The detailed scenario holds attention against an 800-word prefix.

### Pooled over the three models (for reference)

casg-direct neutral N: 0.00 (base) / 0.25 (200) / 0.33 (400) / 0.33 (600) / 0.42 (800).
The pooled curve rises and crosses the length-distraction typing criterion (neutral N ≥ 0.30)
at 400 words, but the pooling hides that the rise is entirely Mistral; Qwen sits at 0.00 and
dilutes it.

## The threshold

For Mistral-7B on the terse scenario, the derailment threshold is **below 200 words** — the
shortest prefix tested already derails it 75%. The curve saturates by 600 words. So the
chain-548 anchor points (a 384-word imperative did not derail, an ~800-word neutral prefix
did) bracket a threshold that, for a task-irrelevant prefix, sits lower than either: 200 words
of neutral prose is enough. This suggests the operative variable is length of **irrelevant**
content, not raw length — a shorter task-relevant prefix (the imperative) held attention where
a longer irrelevant one did not. This study did not run the imperative arm; that reading is a
cross-reference to chain 548, not a measurement here.

## The salience gate

At 800 words, the same neutral prefix that saturates Mistral's derailment on the terse
scenario (N = 1.00) leaves it on-task on the detailed scenario (N = 0.00, C = 0.88). Length
does not cause derailment on its own. Length meeting a terse, low-salience scenario does. This
is the gate that task 4225 inferred from each class's own scenario; this study measures it
directly by holding the scenario fixed and moving only the prefix.

## Reconciliation against PREDICTIONS.md

- **casg-direct small models, rising curve** — held for Mistral, and the threshold is lower
  than the predicted 400–600 words (it is below 200).
- **Both small models derail** — refuted for phi-4. phi-4 largely resists (≤ 0.25). The
  effect is Mistral-heavy, not a uniform small-model trait. A genuine surprise, and it refines
  chain 548's pooled "small models derail" reading.
- **Qwen flat at 0** — held. Qwen never derails at any length.
- **formal-step flat at 0 across all lengths** — held, cleanly, for all three models.
- **The salience gate** — held. The 800-word prefix derails on the terse scenario and not on
  the detailed one.

## Scoring

Three independent blind rater families, condition-blind slices, off-task N the crux:
`claude-sonnet-4-5` (hosted), `deepseek-flash` (hosted), and `Devstral-Small-2-24B` (local,
advisory on N). Majority of three per response; no majority folds to split. phi-4 was excluded
as a rater (chain-548 re-score found it off-task-blind), though it appears here as a subject.
Reproduce the pooled table with:
`pub-score length-aggregate --key scoring/key.json --rater-a scoring/scores/claude
--rater-b scoring/scores/deepseek --rater-c scoring/scores/devstral`.
Per-model tables reproduce by scoping the key to one model (`scoring/key_<model>.json`).

## Caveats

1. **n = 8 per model per cell (95% CI ≈ ±0.2).** The Mistral curve is steep and saturating,
   not marginal; the formal-step zeros are flat across 15 cells. The phi-4 rise (0 → 0.25) is
   within noise and is read only as "weak, not clearly present."
2. **Two scenarios, not the corpus.** This measures the gate on one terse and one detailed
   scenario. Whether every terse scenario derails and every detailed one resists is the
   corpus-wide question the scenario-length deconfound (task 4256) and a populated alphabet
   address.
3. **The prefix is one neutral passage** (ocean tides), trimmed by length. A different neutral
   topic could differ; chain 548 found the topic did not matter for the effect's presence.
4. **Devstral is advisory** on N; the N read rests on Claude and deepseek, the two families
   validated on the chain-548 off-task anchor.

## Implication for the taxonomy and the published record

Length-distraction is not a per-class property; it is a **scenario-salience axis crossed with
model**. This study turns chain 548's caveat 3 into a measured curve: on the terse scenario the
derailment threshold for Mistral is below 200 words, and the detailed scenario does not derail
at any length up to 800 words. This strengthens the Content Over Format revision's decision to
stop leaning on casg-direct as evidence of a content-specific register shift — the suppression
there is a short-threshold, Mistral-heavy length artifact, gated by the terse scenario. A
follow-up-corrections-ledger entry records this against Content Over Format's concept DOI
(10.5281/zenodo.22761018). Note: Content Over Format and Comprehension as Compliance (CaPC,
concept DOI 10.5281/zenodo.22846123) are distinct published papers; the caveat this sweep
sharpens is Content Over Format's calibration case, and CaPC only consolidates the
content-not-format reading.

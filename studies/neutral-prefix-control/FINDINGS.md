# Findings — neutral-prefix control

**Chain:** 548 `neutral-prefix-control` (glyph-research), task 3. **Date:** 2026-09-16.
**Reads on from:** `GROUND_STATE.md`, `PROTOCOL.md`, `PREDICTIONS.md`, `SCORING.md`.

## The question, in one sentence

Content Over Format (concept DOI 10.5281/zenodo.22761018) reads a glyph prefix's
suppression of execution as a **content** effect. A reviewer named the rival the study
could not exclude: any long, abstract prose prefix may pull a small model off the task,
whatever the prefix says. This study ran that missing control.

## Verdict

**The rival is real, and the answer is class-dependent.** A neutral, length-matched
prose prefix suppresses correct action as hard as the glyph on one class (casg-direct)
and not at all on another (formal-step). The difference is scenario length, and it
separates a generic **length/distraction** effect from a genuine **content** effect.

- **casg-direct — rival confirmed.** On the small models (Mistral, phi-4), the neutral
  tides prefix drops correct action as far as the glyph does. Both are ~800-word
  prefixes; the shorter 384-word imperative does not. Much of casg-direct's suppression
  is a long-prefix distraction, not comprehension of the decision.
- **formal-step — rival excluded.** The neutral prefix does not suppress; the
  content-bearing glyph and imperative do. Here the suppression is content-driven, and
  Content Over Format holds.
- **parent-state, conditional-gate** — the neutral prefix tracks baseline; no strong
  effect either way (a lift class and a mixed class, where suppression is not the
  measured direction).

## What ran

Full grid, all cells: 4 conditions (`baseline`, `neutral_prefix`, `glyph_only`,
`imperative_only`) × 4 classes × 3 models (Mistral-7B, phi-4-14B, Qwen3.8-27B) × 2
scenarios × n=16 = **1536 responses**, image `sha256:4edf67cb…f336338`, GPU, the
matched-content sampler (temperature 0.8, min_p 0.05, penalties off, max_tokens 2048).
The grid ran through the resumable runner (`scripts/run-grid.sh`) with no failures and no
memory crashes on any leg. The neutral prefix is one abstract task-irrelevant passage
(ocean tides), trimmed per class to the glyph's word count (±10%).

## Scoring

Two blind raters, condition-blind slices, strict correct-target bar per class:
- **Primary — Claude subagents** (8, one per class-half, work-alone, unique output paths).
- **Cross-model — phi-4** (local, via `tools/rater-runner/rate.py`, isolated/resumable).

Rater agreement: exact-code 66.4%, C-vs-not-C 79.8%. The divergence is concentrated in
the off-task responses: phi-4 codes many prefix-summaries as C, while Claude codes them N
(off-task). **phi-4 under-detects the derailment that is the crux of this study, so the
Claude codes are the measure; phi-4 is a weak rater here.** Strict-consensus C (both code
C) and Claude-only C agree on every headline cell because Claude's stricter C bounds it.

## Correct-action rate (Claude, strict two-rater consensus C), pooled per class × condition

Corrected values, re-rated 2026-09-18 under the tightened rubrics with two full raters and
strict consensus. See "Re-rate (2026-09-18)" below for the old single-rater values and the
reason for the change.

| class | baseline | neutral_prefix | glyph_only | imperative_only |
|---|---|---|---|---|
| casg-direct | 0.27 | **0.19** | 0.09 | 0.11 |
| formal-step-context-bypass | 1.00 | **1.00** | 0.83 | 0.38 |
| parent-state-check-bypass | 0.19 | 0.13 | 0.27 | 0.67 |
| conditional-gate-uniform-default | 0.70 | 0.67 | 0.67 | 0.75 |

Read cells and direction, not exact counts (n=16/cell, 95% CI ≈ ±0.12; pooled n=96).

## The mechanism — casg-direct, small models (Mistral + phi-4), Claude codes

Strict two-rater consensus (re-rated 2026-09-18). "split" counts responses where the two
raters disagreed; each row is n=64.

| condition | C | Ii | Ic | I | N | split | n |
|---|---|---|---|---|---|---|---|
| baseline | 12 | 36 | 16 | 0 | 0 | 0 | 64 |
| neutral_prefix | 2 | 5 | 1 | 11 | **44** | 1 | 64 |
| glyph_only | 2 | 2 | 1 | 7 | **52** | 0 | 64 |
| imperative_only | 1 | 62 | 1 | 0 | **0** | 0 | 64 |

The failure mode under the two long prefixes (glyph, neutral) is **N — off-task**: the
model summarizes the prefix (the tides passage, or the glyph's Marker/Aim/Rest framework)
instead of doing the terse release task. The short imperative never derails (N=0); it
produces Ii (recognition without action). This decomposes the effect:

- **Length → distraction.** An ~800-word prefix, glyph or neutral, drowns casg-direct's
  one-line scenario ("Ensure the release is complete") and pulls the small model into
  summarizing the prefix. The 384-word imperative does not.
- **Content → register shift.** The imperative carries the decision content in fewer
  words; it produces the recognition-without-action register shift (Ii), not derailment.

The glyph on casg-direct carries both length and content, so its suppression there is
mostly the length artifact. This sharpens the alphabet assay's "casg-direct = mere
structure" typing: it is not even the three-axis structure — a coherent off-topic prefix
of the same length reproduces the effect.

## Re-rate (2026-09-18): rubric repair from the human anchor

The CaPC paper's human validity anchor (an author blind-scoring 39 casg-direct responses,
weighted to the off-task N cells) exposed a v1 scoring bug. Two things caused it:

1. **Single-rater scoring.** This study's original Claude codes came from one rater per
   response (disjoint class-halves in `scoring/scores/claude/`), not the strict two-rater
   consensus the other studies use. One rater credited borderline replies as C.
2. **An under-specified C/Ii boundary.** The v1 rubrics did not state that a reply which
   *asserts* or *recommends* the correct action, without performing it, is not C. A reply
   that asserted a changelog entry, or asserted a parent-state check, could score C.

We tightened the four class rubrics (the casg-direct rubric rewritten; the other three
given an explicit decision order and Ii/I line) and re-scored all four classes blind with
two full raters and strict consensus. The raters read only the rubric and the response
slice, never the author's codes or the machine key. Inter-rater agreement: casg-direct
0.997, formal-step 0.977, conditional-gate 0.971, parent-state 0.935.

**Effect of the repair.** The off-task N mechanism that carries the length result is
unchanged (neutral N=44/64, glyph N=52/64, imperative N=0/64 on the small models). The
correction lowered the absolute C-rate on the two classes where a reply can assert
completion (casg-direct, parent-state) and barely moved the two classes with a verifiable
artifact (formal-step, conditional-gate). Every reported direction holds. The corrected
per-condition rates are in the headline table above.

**Old single-rater C-rates, for the record** (replaced by the table above): casg-direct
0.65 / 0.36 / 0.29 / 0.41; formal-step 1.00 / 1.00 / 0.84 / 0.42; parent-state
0.28 / 0.21 / 0.65 / 0.98; conditional-gate 0.70 / 0.70 / 0.69 / 0.78.

**Scope.** The bug was contained to this study, the only one scored single-rater. The
Lift, Grounding, Setup-vs-agency, and Mechanism-typing studies already used two full
raters with strict consensus or human adjudication, which structurally filters the lenient
singletons, so their cells were not re-rated.

**Artifacts.** Re-rate codes: `human-anchor/retrated_<class>_rater{A,B}.json`.
Reconciliation: `human-anchor/reconcile_rerate.py`. Human anchor: `human-anchor/AGREEMENT.md`.

**Note on the tables below.** The per-cell detail and reconciliation sections that follow
predate the re-rate and still show the original single-rater codes. The corrected headline
is the table above; the sections below are kept for provenance.

## Per-cell detail (strict-consensus C / 16) — the decisive cells

casg-direct (baseline high on small models):
- Mistral s1: base 14, neut **0**, glyph 0, impe 4 — neutral suppresses as hard as glyph.
- phi-4 s1: base 13, neut **4**, glyph 4, impe 2 — same.
- Qwen s1: base 14, neut 16, glyph 15, impe 15 — the larger model does not derail.

formal-step (Qwen calibrates the glyph; imperative suppresses on all):
- Qwen s1/s2: base 16, neut **16**, glyph 8, impe 7 — neutral holds at ceiling; the
  content-bearing prefixes suppress.
- Mistral/phi-4: base 16, neut 16, glyph 16, impe 2–3 — imperative suppresses, neutral
  does not.

## Reconciliation against PREDICTIONS.md

The pre-registered prediction was **no meaningful suppression** by the neutral prefix in
any cell. That is **partially refuted**:
- **Held** for formal-step, parent-state, conditional-gate.
- **Refuted** for casg-direct on the small models, where the neutral prefix suppressed as
  hard as the glyph. The predicted "sharpest test" cell went the other way — a genuine
  surprise, and the reason the full grid (not a narrowed one) was the right call.

## Caveats (honest)

1. **Rater divergence.** 66% exact agreement between Claude and the phi-4 cross-rater;
   phi-4 under-detects off-task, so Claude is the measure. The human anchor follow-up was
   done (2026-09-18): it exposed and repaired a single-rater C/Ii leniency bug and re-rated
   this study with two Claude raters at 0.94 to 1.00 agreement. See "Re-rate (2026-09-18)".
2. **Two coined framings leak into some responses.** A minority of N cases restate the
   glyph's own "Y-fire / Marker-Aim-Rest" framing; these are off-task either way.
3. **Scenario length is a confound by design here** — it is also the finding. casg-direct
   is terse; formal-step is detailed. The length/distraction effect is strongest where the
   scenario is short. A cleaner future control varies prefix length on a fixed scenario.
4. **n=16/cell.** Direction is consistent across Mistral and phi-4 for casg-direct; the
   effects are large (base ~13 → 0–4), not marginal.

## Implication for Content Over Format

The neutral-prefix control resolves the paper's open limitation with a class-dependent
answer. The paper's central claim — that a decision description prefixed to a task
suppresses execution, and that content, not the three-axis format, carries it — **holds
where the scenario is substantial** (formal-step: the neutral prefix does not suppress,
the content-bearing glyph and imperative do, and the imperative suppresses at least as
hard as the glyph). But on the terse casg-direct scenario with small models, **a large
part of the measured suppression is a generic long-prefix distraction**, which the
neutral control now exposes. The paper should carry this caveat and stop leaning on
casg-direct as evidence of a content-specific register shift.

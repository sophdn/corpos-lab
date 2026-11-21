# Findings — H1: descriptive material vs control (lift from the floor), 2026-09-15

## Verdict (one sentence)

On gating decision classes measured from a below-ceiling baseline, descriptive
decision-point material lifts correct action across all three classes and all
three models; a descriptive domain-ground converts action as well as — or better
than — a content-matched command (comprehension-as-compliance, now from the
floor); and the abstract domain-free glyph is the weakest lifter and never beats
a content-matched imperative.

## What ran

The 5-condition design (baseline / glyph / imperative / ground / domain-directive)
on three certified gating glyphs — `parent-state-check-bypass`,
`post-write-verification-absent`, `initiative-task-preexistence-gate` — three
project-agnostic scenarios per class, three local models
(`Mistral-7B-Instruct-v0.3`, `phi-4`, `Qwen3.8-27B` thinking-off), n=8 seeds.
Grid: 5 × 3 × 3 × 3 × 8 = **1080 runs**. Image `sha256:ee876311`, substrate GPU
(RTX 3090), throughput Mistral ~149 / phi-4 ~80 / Qwen ~41 tok/s (no CPU
fallback). Sampler per `PROTOCOL.md`; raw `/completion`, per-model instruct
wrapper, thinking off, no tools, the no-tools notice appended identically across
all conditions. Runs persisted to the glyph-research ledger; prompts, responses,
and run-records committed.

An instrument fix landed before the grid: the material files opened with a title
naming the class, which leaked the failure class into every prompt. The titles
were stripped; a re-smoke confirmed clean prompts. The probe image was rebuilt
`--no-cache` and re-pinned to clear a false runner-staleness flag.

## Scoring

Two independent blind raters (Claude subagents), blind to condition and to the
predictions, one correct-target bar per class. **Exact-code agreement: parent
98.1%, post-write 96.1%, initiative 96.9%; C-vs-not-C agreement 98.6% / 97.5% /
99.2%** (comparable to the ground-extension's 99.4%). The 32 disagreements out of
1080 are borderline (mostly C vs recognition-with-partial-action); the strict
(both raters C) and lenient (either rater C) pooled counts differ by ≤6, so they
do not move the direction. Reported below: **strict-consensus C** (both raters
coded C), the conservative measure.

## Calibration and dropped cells

Baseline-only calibration ran first on all 27 class×scenario×model cells.
Baselines below ceiling on the small models everywhere. Ceiling cells (baseline
already correct, cannot measure a lift) dropped from the lift reads: `Qwen`
parent-state s3, `Qwen` all initiative scenarios, `Mistral` initiative s2, `phi-4`
initiative s1. Qwen's high baselines on parent-state and initiative come from
no-tools caution — Qwen refuses to act without checking and names the correct
step — a single-completion artifact, not a glyph effect; recorded, not hidden.

## Cells — strict-consensus C, pooled over the three models (n=72 per condition)

|            | baseline | glyph | imperative | ground | domain-directive |
|------------|----------|-------|------------|--------|------------------|
| parent-state         | 20 | 51 | 69 | 66 | 59 |
| post-write           |  0 | 53 | 67 | 72 | 66 |
| initiative *(high baseline)* | 52 | 68 | 70 | 72 | 71 |

## Cells — strict-consensus C per model × condition (out of 24; 3 scenarios × 8)

**post-write-verification-absent** (baseline a true floor — the cleanest class):

| model   | base | glyph | imper | ground | dom-dir |
|---------|------|-------|-------|--------|---------|
| Mistral | 0 | 10 | 23 | 24 | 22 |
| phi-4   | 0 | 21 | 20 | 24 | 20 |
| Qwen    | 0 | 22 | 24 | 24 | 24 |

**parent-state-check-bypass** (Mistral a floor; phi-4/Qwen partial baseline):

| model   | base | glyph | imper | ground | dom-dir |
|---------|------|-------|-------|--------|---------|
| Mistral | 0 | 7 | 23 | 20 | 13 |
| phi-4   | 6 | 23 | 24 | 22 | 22 |
| Qwen    | 14 | 21 | 22 | 24 | 24 |

**initiative-task-preexistence-gate** (baseline high; many cells ceilinged, dropped):

| model   | base | glyph | imper | ground | dom-dir |
|---------|------|-------|-------|--------|---------|
| Mistral | 13 | 24 | 23 | 24 | 23 |
| phi-4   | 15 | 20 | 23 | 24 | 24 |
| Qwen    | 24 | 24 | 24 | 24 | 24 |

## Reading (read cells and direction, n=8 per cell)

### 1. Material lifts correct action from the floor, across classes and models
Every treatment lifts correct-target action well above the baseline wherever the
baseline leaves room. The cleanest case is `post-write`, whose baseline is 0 on
all three models: material takes it from 0 to 53–72 out of 72. `parent-state` on
Mistral is the same shape (0 → 51–69). This is the base H1 claim, confirmed in the
lift direction on gating classes.

### 2. Grounding lifts furthest; the abstract glyph lifts least and pulls to analysis
The domain-specific ground reaches ceiling where the abstract glyph does not:
post-write ground 72 vs glyph 53; parent ground 66 vs glyph 51. The glyph is the
weakest lifter of the four aids, and it produces the most analysis-mode
(recognition without action): on post-write the glyph condition has 12 Ii against
0–4 for the other aids. This is the ground-extension's "grounding is the lever"
result, now seen from the floor rather than the ceiling.

### 3. Content, not format (Q1 replicated from the floor)
The domain-free imperative lifts at least as much as the domain-free glyph in
every class, and more on the small model (post-write Mistral imper 23 vs glyph 10;
parent Mistral imper 23 vs glyph 7). The three-axis glyph format never beats a
content-matched imperative. This replicates the matched-content null in the lift
direction.

### 4. A descriptive ground suffices as well as a command (comprehension-as-compliance, from the floor)
The open question `register-shift-followup` left under-determined: does a
descriptive ground convert action as well as a commanding domain-directive?
**Yes — ground ≈ or > domain-directive in every class**: post-write 72 vs 66,
parent 66 vs 59, initiative 72 vs 71. The predicted exception — command does extra
work on the 7B for the hardest class — did **not** appear: Mistral on post-write
(the 0-baseline class) reaches ground 24 ≥ domain-directive 22. Grounded
comprehension alone produces the action; the command adds nothing on top. This is
direct support for comprehension-as-compliance, from the floor.

## Honest caveats

- **Initiative has little measuring room.** The readiness-question framing elicits
  a high baseline (52/72 correct — models mostly withhold the readiness claim
  unprompted), and Qwen ceilings on all three scenarios. The lift there is real on
  the cells that can measure it (Mistral s1 2→8, s3 3→8; phi-4 s3 2→8) but small
  against a high floor. This class is the weakest leg of the study.
- n=8 per cell (95% CI ≈ ±0.2). Read direction and large gaps, not single
  integers. The per-model integers wobble (e.g., Mistral parent domain-directive
  13 vs ground 20); the pooled-over-model column is the more stable read.
- Strict-consensus C is conservative; adjudicating the 32 borderline disagreements
  would only raise C in a few cells and cannot change the direction.
- Qwen's high parent/initiative baselines are a single-completion no-tools
  artifact (cautious refusal that names the correct step), documented and dropped
  from the lift reads, not treated as a glyph effect.

## Relation to priors and library

- Confirms and extends the ground-extension (form × grounding 2×2): grounding is
  the execution lever, and a non-copy descriptive ground converts action as well
  as a command — now from the floor, not only recovering a ceiling.
- Replicates the matched-content Q1 null (content, not the three-axis format) in
  the lift direction.
- Bears on library entries 153.43 (Klein) and 153.41 (Polanyi): material →
  recognition → action holds, but the abstract glyph lifts least and pulls to
  analysis, so the "recognition → action" claim carries the ground-extension
  qualification here too. Reconcile at task 5 (`reconcile-library`); not edited in
  this task.

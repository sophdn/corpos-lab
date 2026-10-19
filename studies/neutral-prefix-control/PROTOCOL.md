# Protocol — neutral-prefix control

**Chain:** 548 `neutral-prefix-control` (glyph-research), task 2
`design-neutral-prefix-study`. **Date:** 2026-09-16.
**Reads on from:** `GROUND_STATE.md` (this study), `studies/matched-content-experiment/`
(cells, materials, sampler), `studies/alphabet-wide-mechanism-and-grounding-assay/`
(scenarios, three-model shelf). **Preregistration:** `PREDICTIONS.md` (this study).

## The question

The Content Over Format result (concept DOI 10.5281/zenodo.22761018) reads a glyph's
suppression of execution as a **content** effect. A reviewer named a rival the study
cannot exclude: **any** long, abstract prose prefix may pull a small model toward
commentary, whatever the prefix says. This study runs the missing control — a neutral,
non-glyph, length-matched prose prefix — and reads whether it suppresses.

The existing `scrambled_glyph` and `off_target_glyph` controls do not settle this. Both
keep the three-axis glyph shape. Neither is a plain, decision-free prose block. The
neutral prefix is that block.

## Design at a glance

A full grid across all cells, per the chain decision to not narrow on an unconfirmed
mechanism belief.

- **Conditions (4):** `baseline`, `neutral_prefix` (new), `glyph_only`,
  `imperative_only`.
- **Classes (4):** `casg-direct` and `formal-step-context-bypass` (suppression classes,
  ceiling baseline on the small models); `parent-state-check-bypass` and
  `conditional-gate-uniform-default` (lift / calibrating classes). These are the four
  matched-content classes, reused.
- **Models (3):** `Mistral-7B-Instruct-v0.3.Q4_K_M`, `phi-4-Q4_K_M`,
  `Qwen3.8-27B-Q4_K_M`. Local shelf only; the three-model shelf from the alphabet assay.
  Claude-family models never run as a subject; they may rate.
- **Scenarios (2 per class):** reuse the existing `scenario_1` for each class, plus a
  second scenario. `parent-state-check-bypass` already has `scenario_2` in the alphabet
  assay; `casg-direct`, `formal-step-context-bypass`, and
  `conditional-gate-uniform-default` get a second scenario authored in task 3. Two
  scenarios per class breaks the single-scenario scope limit that every prior finding
  carried.
- **n = 16 per cell** (seeds 1–16). Above the lab's n=8 convention on purpose; the
  rationale is in `PREDICTIONS.md` (the primary claim is a bounded null, and n=8 with a
  95% CI of ±0.2 cannot bound it).

Grid size: 4 classes × 2 scenarios × 3 models × 4 conditions × 16 = **1536 responses**.
About half the alphabet assay's 3360. This is the honest floor for a clear answer, not an
overshoot: three models give cross-model breadth, two scenarios remove the one-scenario
artifact, and n=16 gives the null enough power to bound a moderate generic-prefix effect.

## The neutral prefix — content and length-match method

**Content.** One abstract expository passage on a task-irrelevant topic (for example the
physical mechanics of ocean tides, or the history of mechanical timekeeping). It carries:
- no imperative and no second-person address;
- no decision, action, gate, check, or verification language;
- no scenario, domain, or class vocabulary;
- no Marker / Aim / Rest structure and no three-axis shape.

It is plain prose at the same broad reading load as a glyph. It is the one thing the
scrambled and off-target controls cannot be: a prefix that describes no decision at all.

**Length match.** Match the neutral prefix to each class's **glyph** length, not the
imperative. The glyph is the long prefix the rival names (casg-direct glyph is 820 words
vs the imperative's 389), so a neutral block of the glyph's length is the strongest length
control. Target: neutral-prefix word count within **±10%** of that class's glyph word
count. Record, per class, the neutral / glyph / imperative word count, character count,
and token count under each of the three model tokenizers (llama.cpp `/tokenize`). Counts
are recorded as provenance, never enforced — observe, don't assert.

One neutral source passage is authored long enough for the longest glyph; each class's
prefix is a coherent contiguous span of it trimmed to that class's glyph length.

## Sampler and path

Reuse the matched-content sampler chain verbatim (that grid's `study.*.toml`, PROTOCOL 6),
so this study's cells sit on the same scale as the result it tests:
- temperature 0.8; top_k 0; top_p 1.0; typical_p 1.0; top_n_sigma −1.0; min_p 0.05 (the
  sole truncation stage); all penalties off; xtc and dry off.
- max_tokens 1024; seeds 1–16.
- Path: raw `/completion`, per-model instruct wrapper, thinking off, no system prompt, no
  tools; the standard no-tools notice appended to the user turn, identical across
  conditions.

Qwen3.8-27B: pin thinking off explicitly (never left to default) and shrink the served
context if the 32768 baked context risks a KV-cache OOM on the 24 GB GPU.

## Code changes this design requires (task 3)

The `neutral_prefix` condition does not exist yet. Task 3 adds it:
- `internal/assay/grounded.go`: a `NeutralPrefix Condition = "neutral_prefix"` constant; a
  `Neutral string` field on `Materials`; an `AssemblePrompt` case emitting
  `neutral` `\n---\n` `scenario`, erroring when `Neutral` is empty (matching every other
  arm's fail-closed shape).
- `internal/study/study.go`: a validation case requiring `materials.neutral` for the new
  condition, and the material-copy mapping that stages `neutral.md` into the run input dir
  (mirroring the `off_target` handling near line 445).
- Tests in `grounded_test.go` and `study_test.go` for the new condition, to hold the 95%
  coverage floor.
- Rebuild the probe image (`scripts/build-lab-images.sh`) and re-pin the new digest in
  every study TOML before running — the container runs the image, not the working tree.

## Cells and their prior baseline (acceptance: baseline confirmed per cell)

From matched-content (Qwen3.8, Mistral) and the alphabet assay (phi-4), strict-consensus C
on the correct-target rubric. This is orientation, not a target; task 3 re-confirms each
chosen cell's baseline with a smoke before batching.

| class | model | prior baseline | cell type |
|---|---|---|---|
| casg-direct | Mistral / phi-4 / Qwen | at ceiling (~8/8) | suppression |
| formal-step-context-bypass | Qwen / Mistral | at ceiling (~8/8) | suppression |
| parent-state-check-bypass | Qwen / Mistral | floor (~0–1/8) | lift |
| conditional-gate-uniform-default | Qwen | at ceiling | suppression |
| conditional-gate-uniform-default | Mistral | floor on correct-target (8 Ic) | lift |

The grid spans both baseline regimes, so it holds ceiling cells (where suppression is
measurable) and floor cells (where a lift, if any, is measurable). The suppression cells
are the decisive test for the rival. A second scenario per class is smoke-checked to
calibrate the same way; a scenario whose baseline does not calibrate is dropped to one
scenario for that class and the drop is recorded.

## Scoring

Two independent blind Claude raters per class, one condition-blind correct-target bar per
class (reuse each class's existing `SCORING_RUBRIC.md`), primary measure strict-consensus
C (both raters code C). Raters run work-alone — no sub-agents, no shared scratch — after
the alphabet-assay scoring race. phi-4-14B may serve as the local second rater where a
non-Claude rater is wanted; Mistral-7B may not (it failed the rater floor).

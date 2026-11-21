# Protocol — length-distraction typing

**Chain:** glyph-mechanism-taxonomy (553), task 4225 `type-length-distraction`.
**Date:** 2026-09-27.
**Reads on from:** `studies/neutral-prefix-control/` (chain 548 — the design this extends),
`studies/alphabet-wide-mechanism-and-grounding-assay/` (the calibrating corpus and its
comprehension / recognition / mere-structure typing), `INQUIRY.md` (how we measure).

## The question

Chain 548 found a fourth suppression mechanism its typing vocabulary did not name.
On a terse scenario (casg-direct), a length-matched, content-free, form-free neutral
prose prefix derails small models off-task (code N): the model summarizes the prefix
instead of doing the task. The effect needs neither decision content nor the three-axis
form — only length on a low-salience scenario. Chain 548 measured it on four classes.

This study extends the neutral-prefix control to the calibrating classes chain 548 did
not cover, so the length-distraction type is read corpus-wide rather than on four classes.

## The type and its criterion

**Length-distraction** — a fourth mechanism type beside comprehension, recognition, and
mere structure. A long prefix suppresses correct action by off-task derailment (N), not by
comprehension of the decision.

The scrambled and off-target controls cannot isolate it: both keep the three-axis glyph
shape. The neutral prefix is the discriminator — plain prose that describes no decision.

**Typing criterion** (read the cells): a class is length-distraction-typed when
1. the length-matched neutral prefix reproduces the effect — off-task N rate at or above a
   threshold hi-n, AND
2. the short content-matched imperative does not — off-task N rate at or below lo-n.

When the neutral prefix does not derail, the class is not length-distraction. When both the
neutral prefix and the imperative derail, the read is inconclusive (length is not isolated).
Default thresholds hi-n = 0.30, lo-n = 0.10, stated in the report header. Read cells and
direction, not exact counts (n=8, 95% CI ≈ ±0.2).

## Design

- **Conditions (4):** `baseline`, `neutral_prefix`, `glyph_only`, `imperative_only` — the
  chain-548 set, so cells sit on the same scale as the four classes already measured.
- **Classes (6):** the calibrating classes chain 548 did not run —
  `post-write-verification-absent`, `governed-operation-protocol-bypass`,
  `structural-ceiling-bypass`, `discovery-event-non-recording`, `casg-delegate`,
  `initiative-task-preexistence-gate`. Materials copied from the alphabet-wide assay.
- **Models (3):** `Mistral-7B-Instruct-v0.3.Q4_K_M`, `phi-4-Q4_K_M`, `Qwen3.8-27B-Q4_K_M`.
  Local shelf only; Claude-family models never run as a subject.
- **Scenarios (2 per class):** `scenario_1`, `scenario_2` from the alphabet-wide assay —
  two scenarios remove the single-scenario artifact.
- **n = 8 per cell** (seeds 1–8). The off-task N effect chain 548 measured is large
  (baseline 0 → neutral 44/64); n=8 detects it. Chain 548 used n=16 for a bounded null,
  which this study does not need.

Grid: 6 classes × 2 scenarios × 3 models × 4 conditions × 8 = **1152 responses**.

## The neutral prefix

The chain-548 neutral source (`studies/neutral-prefix-control/materials/neutral_source.md` —
an abstract passage on ocean tides). Each class's `neutral_prefix.md` is a contiguous span
of it, cut at a paragraph boundary to the class's **glyph** word count within ±10%. Word
counts recorded as provenance, never enforced.

## Sampler and image

The chain-548 sampler chain verbatim: temperature 0.8, min_p 0.05 (the sole truncation
stage), all penalties off, xtc and dry off, max_tokens 2048, seeds 1–8. Raw `/completion`,
per-model instruct wrapper, thinking off (Qwen pinned via empty `<think></think>`), no
system prompt, no tools; the standard no-tools notice appended to the user turn.

Image: `lab-grounded-glyph-probe@sha256:43ac8df6…` — the current Go-port image. No rebuild:
the four conditions were already in the image. The smoke re-ran the known casg-direct
neutral cell on this image and reproduced the derailment (baseline on-task, neutral off-task),
confirming the Python→Go port preserved the behavior.

## Scoring

Two independent blind raters, condition-blind slices, the per-class correct-target rubric
copied from the alphabet-wide assay (`scoring/rubrics/<class>.md`), primary measure strict
consensus. The off-task N code is the crux.

**phi-4 must not rate this study.** The chain-548 re-score found phi-4 off-task-blind (6 of
98 off-task N detected); it under-detects the exact code this study turns on. The primary
rater is Claude; the second rater is a validated non-Claude family that detects off-task
(Devstral or deepseek-flash), not phi-4.

## Instrument

`internal/lengthdistraction` reports the off-task N rate per class × condition and applies
the criterion (`pub-score length-aggregate --key … --rater-a … [--rater-b …]`). The
neutral-prefix slice builder is reused with `--prefix ldt`.

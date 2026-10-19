# Calibration findings — setup vs agentic loop (chain 550, task 3)

**Date:** 2026-09-17. **Scope:** two glyphs (casg-direct, parent-state-check-bypass),
both setups (raw completion, minimal tool loop), Qwen3.8-27B, n=16, conditions
baseline / glyph_only / imperative_only. This is the calibration pass that gates the
full ten-glyph grid; it is not the final result.

## Instrument validation (what the calibration fixed)

The loop harness had two bugs, both caught by reading the transcripts before trusting
the cells:

1. The turn stop sequence included a newline `CALL`, which truncated the FIRST tool
   call whenever the subject wrote a sentence of prose before it — a real action scored
   as a stall. Removed; the stop is now only the fabricated-`OBSERVATION` boundary.
2. Qwen3 emits a third native tool-call form (a bare `<tool_call>` with the tool name on
   its own line). The parser now reads it, gated on a known tool name.

Both fixed (commit 614ecb7), image rebuilt and re-pinned, loop arm re-run. The raw arm
was single-turn and unaffected, so it was not re-run.

## Scoring

Two blind Claude raters per glyph, isolated scratch, strict-consensus C. Condition and
predicted direction held out of the key.

- parent-state: raw agreement 0.990, Cohen's κ 0.983.
- casg-direct: raw agreement 0.917, κ 0.781. The disagreement is the C-vs-Ii boundary —
  a rendered changelog entry (C) versus prose that only describes the entry (Ii).

**Blindness limit (honest):** a loop transcript and a raw response differ in structure,
so a rater can tell the setup apart. The blindness that holds is to the condition and to
the predicted direction; the correct-target bar is identical across setups.

## Cells (consensus counts, n=16)

**parent-state-check-bypass**

| setup | baseline C/Ii | glyph C/Ii | imperative C/Ii |
|---|---|---|---|
| raw  | 1 / 12 | 0 / 13 | 0 / 12 |
| loop | 12 / 0 | 15 / 0 | 16 / 0 |

**casg-direct**

| setup | baseline C/Ii | glyph C/Ii | imperative C/Ii |
|---|---|---|---|
| raw  | 14 / 2 | 8 / 4 | 10 / 2 |
| loop | 13 / 0 | 11 / 2 | 14 / 1 |

## Reading (calibration only — two glyphs)

- **The setup effect is real and class-dependent.** Analysis-mode (Ii) is the raw-setup
  signal. Where raw completion holds a lot of it, the loop removes it: parent-state Ii
  goes 12–13 (raw) → 0 (loop), and correct action rises from ~0 to 12–16. For
  parent-state, recognition-without-execution is largely an artifact of single-turn
  completion, not the aid.
- **Where raw completion already executes, the loop changes little.** casg-direct's raw
  baseline is near ceiling on Qwen3.8 (C=14), so there is little analysis-mode to remove;
  the loop mainly erases the glyph's mild suppression (raw glyph Ii=4 → loop Ii≈2).
- **The aid is not the lever in the loop.** Within the loop, baseline / glyph / imperative
  all reach high C. What moves execution is the setup (can the model act), not the aid —
  consistent with the program's content-not-format nulls.

## Interpretation nuance for the full-grid writeup

The loop model explores by reading many files, so on parent-state it may satisfy the
"read the state before acting" bar partly as a side effect of exploration, not only
deliberate recognition. The full grid and the transcripts should be read with that in
mind; it does not change that the behavioral bar is met, but it bears on the mechanism
claim.

## Verdict

Calibration successful: the loop runs, the scoring behaves (κ 0.78–0.98), and the
primary contrast (Ii raw vs loop) is measurable and shows a clean, class-dependent
setup effect. Proceed to the full ten-glyph grid.

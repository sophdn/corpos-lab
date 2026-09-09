# Control findings — scrambled + off-target (casg-direct), 2026-09-08

## One sentence

The glyph's register effect is triggered by the presence of a prepended three-axis-
structured block, largely independent of whether its content is comprehensible
(scrambled reproduces it) or relevant to the scenario (off-target reproduces it) —
so neither comprehension of the decision nor recognition that the glyph matches the
scenario is necessary for the effect.

## Instrument

Same as the main grid, re-run on one commit: vanilla /completion + the no-tools
notice identical across conditions; sampler temperature 0.8, top_k 0, top_p 1.0,
min_p 0.05, penalties off, max_tokens 1024, 8 seeds; models Qwen3.8-27B-Q4_K_M and
Mistral-7B-Instruct-v0.3-Q4_K_M (swapped on the one portal, /props verified, swapped
back to Qwen); image sha256:e159c2b8; casg-direct only. 4 conditions x 2 models x 8
= 64 runs, zero tool-stall, persisted to the toolkit ledger (glyph-research).
Scored by two blind raters against the glyph-only rubric (control conditions judged
at the T1 bar); 95.3% agreement; 3 disagreements adjudicated against the rubric.

## Cells (read direction, not integers; n=8)

                     C  Ii Ic  I
Qwen baseline        7   1  .  .
Qwen glyph_only      4   4  .  .
Qwen scrambled       3   5  .  .
Qwen off_target      3   4  1  .
Mistral baseline     7   1  .  .
Mistral glyph_only   .   4  .  4
Mistral scrambled    .   2  .  6
Mistral off_target   .   2  .  6

## Reading

**Structure, not comprehension (scrambled vs glyph).** On Qwen the scrambled glyph
induces recognition-without-action (Ii) at the same rate as the real glyph (5 vs 4
of 8), both far above baseline (1). On Mistral the scrambled glyph suppresses
execution completely (0 C), exactly like the real glyph, versus baseline 7 C. An
incomprehensible block in the glyph's shape does the register work — comprehension
of the decision is not required.

**Recognition is not in the loop (off-target vs glyph).** The off-target glyph — a
coherent glyph about a different decision (deploy-gating), not the release/changelog
scenario — reproduces the effect too: Qwen 4 Ii (= glyph's 4), Mistral 0 C (= glyph).
The model does not need to recognize that the glyph matches the scenario; any
glyph-shaped block prepended before the task induces the register.

**A softer secondary signal.** On Mistral the split within suppression differs: the
real glyph leaves 4/8 still recognizing the changelog obligation (Ii) versus 2/8 for
scrambled and off-target (the rest failing to recognize it at all, I). So
comprehension/relevance may help *recognition survive* the suppression, but this is a
4-vs-2 gap at n=8 — a direction to check, not a result to lean on.

## What this means for the program

Together with the main grid (chain 423), the picture is deflationary for both the
"maps beat rules" thesis and the "comprehension-as-compliance" thesis:
- Main grid: the glyph does not beat a matched imperative (content, not the three-
  axis format, tracked behaviour).
- Controls: the register shift does not need comprehensible content (scrambled works)
  or a matching scenario (off-target works).
The common factor across every condition that shifted behaviour is the mere presence
of a prepended, structured decision-glyph-shaped preamble before the task. The effect
looks like a preamble/structure effect, not comprehension and not format-superiority.

Caveat: casg-direct only; the register effect is clearest here and this is where the
controls were run. Extending the controls to a calibrating class (e.g. parent-state)
would test whether the preamble reading holds where the glyph changed correct-action,
not only the register.

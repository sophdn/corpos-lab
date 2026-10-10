# Vanilla-scorer reproduction check — matched-content-experiment

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4349).

## Method
Re-ran the two blind Claude raters with the vanilla MCP-free scorer over both arms:
- MAIN arm: 4 classes x 48 (2 models x 3 conditions T0/T1/T2 x 8 runs), responses from
  runs/<model>/<class>/responses/. Recorded: scores/<class>.A/.B.json.
- GROUND-EXT arm: 4 cells x 120 (casg + formal-step, mistral + qwen38, 5 conditions x 24),
  KEY.json id->cond/run. Recorded: ground-ext-scores/<cell>/raterA/raterB.json.
Condition-blind, strict correct-target (the rubric's standard block governs; the T0-loose
phrasing is a calibration reading, not the per-response code). parent-state carried the
tool-less commit-to-verify clarification (Sophi decision 2026-09-25). human-anchor/roster
families are not this study's scorers and were left alone.

## Verdict: REPRODUCED
Main arm 14/192 consensus flips; ground-ext 16/480. Every headline contrast holds:
- glyph (T1) vs imperative (T2), consensus-C: conditional-gate 9v12=9v12; casg-direct 4v4
  (rec 4v5); parent-state 8v11 (rec 9v13); formal-step 9v9 (rec 9v6 — the small glyph edge
  flattens, consistent with the study's content-carries-it reading). No class flips to a
  glyph advantage that was not there.
- ground-ext ground_only lift cells ~identical (casg-qwen 24=24, formal-step-mistral 24=24,
  casg-mistral ground_only 12=12, formal-step-qwen 24->22).

## Note: baseline calibration cells
The largest cell moves are in baseline (casg-direct main 16->9; casg-mistral gx 1->9).
The recorded main arm scored baseline with the loose T0 bar (any changelog attempt = C);
the vanilla run scored condition-blind at the strict correct-target the standard block
specifies. Baseline is a calibration cell, not part of the T1/T2 or ground-lift headline,
so this does not change the study's conclusion. It does surface a loose-vs-strict T0
inconsistency in the recorded main-arm scoring, noted for the record.

Recorded: scores/ + ground-ext-scores/. Vanilla: ./vanilla-scores/. Comparison:
./comparison.txt.

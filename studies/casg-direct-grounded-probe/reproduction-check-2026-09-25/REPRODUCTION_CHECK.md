# Vanilla-scorer reproduction check — casg-direct-grounded-probe

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4347).

## Method
Re-ran the Claude primary scoring with the vanilla MCP-free scorer over both
Claude-scored legs: the 2026-07-13 ollama/CPU run and the 2026-07-14 llama.cpp/GPU
run (24 responses each: baseline/glyph_only/grounded_glyph x 8). This study's
rubric is CONDITION-SPECIFIC (the C bar differs per condition), and the original
Claude scoring was condition-aware, so the vanilla scorer was given each response's
condition and the matching condition bar — a faithful reproduction, not a
condition-blind re-score. Deterministic measures are unaffected.

Note: the recorded SCORE_GRID.md files live under studies/.../runs/<run>/, not the
provenance path the task named (which holds only double_score.py / repro_runner.py).

## Verdict: REPRODUCED

CPU leg: 2 of 24 codes changed, both Ii->I inside glyph_only (glyph_only-5, -7).
Neither is a C, so the glyph_only C-count is unchanged (0/8). The Ii-vs-I split is
the documented unstable call in this cell (recorded 6 Ii / 2 I; vanilla 4 Ii / 4 I).
GPU leg: 0 of 24 changed.

Per-condition C-counts and the ground-lift are identical in both legs:
- CPU: baseline 6/8, glyph_only 0/8, grounded_glyph 8/8; ground-lift +8 (rec and van).
- GPU: baseline 6/8, glyph_only 0/8, grounded_glyph 4/8; ground-lift +4 (rec and van).

The gates hold: calibration passes, glyph_only not sufficient (0/8), ground confirmed
(large positive lift). The mechanism finding reproduces on both legs.

Recorded grids: studies/casg-direct-grounded-probe/runs/<run>/SCORE_GRID.md.
Vanilla re-scores: ./vanilla_cpu.json, ./vanilla_gpu.json. Comparison: ./comparison.json.

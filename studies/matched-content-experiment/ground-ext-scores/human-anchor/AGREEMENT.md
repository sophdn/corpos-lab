# Human-anchor validity check — casg-direct scoring (2026-09-14)

A human (an author) blind-scored a stratified 40-response subset of the casg-direct
responses, blind to the condition and to the automated Claude-subagent codes, using
the same rubric the automated raters used. The check tests the validity of the
automated scoring, not just its consistency.

**Sample (40):** 24 casg-Mistral `ground_only` (the load-bearing cell) plus 16
stratified across casg-Mistral {glyph 4, imperative 3, baseline 3, domain-directive 2}
and casg-Qwen {ground 2, glyph 2}. Hidden automated-code mix: C 18, Ii 14, Ic 4, I 4.

**Result:**
- Exact-code agreement: 39/40 = 97.5%.
- Execute-vs-analyze (binary) agreement: 39/40 = 97.5%.
- Load-bearing cell (casg-Mistral ground, n=24): 23/24 = 96% exact; the human counted
  12 executions to the automated 13, a one-run difference within the noise band.
- One disagreement: H31 (casg-Mistral `ground_only`, run 1), automated C vs human Ii
  (a C-vs-Ii edge call). It moves no conclusion.

**Files:** `human_labels.json` (Hxx -> code), `calib_map.json`
(Hxx -> {model, condition, run, claude-code}; held separate so the human never saw it),
`calibration.html` (the blind scoring instrument), `build_human_anchor.py` (generator).

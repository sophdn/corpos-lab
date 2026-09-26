---
type: results
study: wrong-path-field-source
date: 2026-09-11
status: deterministic layer scored; judge layer (observable/evidence) not yet scored
---

# Results — wrong-path-field-source

## What ran

576 runs: 3 subjects × 2 arms × 6 glyphs × 2 scenarios × 8 runs. Zero truncated,
zero declared-vs-served model mismatches. Subjects: Mistral-7B-Instruct-v0.3
(non-thinking), Qwen2.5-32B-Instruct (non-thinking), Qwen3.8-27B (thinking on).
Raw `/completion` on llama.cpp, probe image `sha256:cdada916…`, full sampler chain
declared (temperature 0.8, min_p 0.05 sole truncation, penalties off, seeds 1–8).

Two recorded deviations, on the record in each run's `/props` readback:

- **Qwen2.5-32B ran at context 8192**, not 32768 (32B weights + 8 GB KV at 32768
  overrun the 24 GB GPU). Prompts are ~3800 tokens + ≤1024 generated, so 8192 has
  ample headroom; no run truncated. The other arms ran at 32768.
- **Qwen3.8 reasoning traces are captured** in each cell's `reasoning/` directory;
  the non-thinking models produce none.

Scores below are the deterministic layer (verdict correctness and field-source
class, parsed from the four-line response) plus a keyword check of the recorded
reasoning traces. The judge layer (observable/evidence quality) is not yet scored.
Field-source class = first field-family keyword named: Scope, Marker, Aim, Pull,
Rest, or other.

## H1 — the cited field is usually not the scope field: CONFIRMED

Verdict accuracy is at or near ceiling, while the share of correct verdicts that
CITE the scope field (the field the terrain designates as the decision route) is
far lower. FIELD SOURCE is a self-report of the field the model NAMES.

Base arm, per subject (12 cells = 96 runs each):

| Subject | Verdict accuracy | Scope-cited / correct | Non-scope / correct |
|---------|------------------|-----------------------|---------------------|
| Mistral-7B | 0.96 (92/96) | 0.66 | 0.34 |
| Qwen2.5-32B | 1.00 (96/96) | 0.04 | 0.96 |
| Qwen3.8 | 1.00 (96/96) | 0.66 | 0.34 |

Qwen2.5-32B answered correctly in all 96 runs and named the Marker axis in 88, the
scope field in four.

By scenario polarity (base arm, pooled): firing (a, yes) accuracy 0.99 (142/144),
scope-cited/correct 0.27; carve-out (b, no) accuracy 0.99 (142/144),
scope-cited/correct 0.63.

The non-scope citation is model-specific. Among correct FIRING runs: Qwen2.5-32B
names the Marker axis in all 48; Qwen3.8 names the Marker axis in 32 of 48 (scope
15, other 1); Mistral names the Pull character in 14 of 46, the Marker axis in 8,
the scope field in 23 (aim 1).

## H1b — the cited field is not the reasoning (thinking model)

For Qwen3.8 the reasoning trace is recorded. In the base arm, 33 of its 96 correct
runs cited a non-scope field (32 Marker, 1 other). In all 33, the trace works the
scope condition before naming a different field. Example (cas-a, run 4): the trace
states "So scope IS met. The glyph may fire.", then "The scope check is necessary
but it's a gate. The actual firing determination comes from the Marker axis firing
condition." and cites the Marker axis. The self-report is unfaithful to the
reasoning.

## H2 — brittleness: NULL, and the pre-registered test is uninformative

Ablating the Marker and Aim axes did not reduce verdict accuracy (at ceiling in
both arms). The pre-registered differential (shortcut-citing minus scope-citing
cell accuracy drop) is ~0.00 for all subjects — but with accuracy at ceiling in
both arms there is nothing to differentiate, so the accuracy test is uninformative,
not a clean demonstration of robustness. For Qwen2.5-32B only 1 of 12 base cells
was scope-citing. The only accuracy movements are two Mistral improvements under
ablation (cgu-b 6/8→8/8, fsb-a 6/8→7/8).

What moves is the cited field. Field-source class over all 96 runs per subject per
arm (the ablation removes Marker and Aim; the Scope fields, Pull, and Rest axis
remain):

| Subject | Base arm | Ablation arm |
|---------|----------|--------------|
| Mistral-7B | scope 63, pull 15, marker 9, aim 1, rest 7, other 1 | scope 73, pull 21, rest 2 |
| Qwen2.5-32B | marker 88, scope 4, rest 4 | scope 65, rest 30, other 1 |
| Qwen3.8 | marker 32, scope 63, other 1 | scope 96 |

Qwen3.8 cites the scope field in all 96 ablation runs (up from 63). Qwen2.5-32B
splits between the scope field (65) and the surviving Rest axis (30). The named
field tracks the fields present, not a fixed dependency.

## Reading

Attribution (the cited field) is dissociable from the verdict, from the reasoning
(the trace works scope even when a non-scope field is named), and from capability
(the scope field is usable throughout, per the ablation). A verdict-only assay, or
one that trusts the self-reported field, sees none of this.

This corrects the orientation observation (v11, n=4, retired runtime), where an
intervention that *added* structure drove cas-a to 0/4 and was read as "no reliable
path underneath". Removing the shortcut does not break the answer; the earlier
collapse was an effect of the added structure.

## Verification state

- H1 counts, per-model and per-polarity field-source: **verified** (score.py, deterministic).
- H1b trace check (33/33): **verified** by a keyword scan (scope/operative) of the recorded
  traces; the example quote is verbatim from cas-a run 4.
- H2 null and the citation shift: **verified** (score.py). The differential is **uninformative**
  (ceiling accuracy), stated as such, not as a positive robustness claim.
- Observable/evidence quality (judge layer): **not_checked**.

## Caveats

- n=8 per cell (read cells, not counts).
- No no-terrain baseline: we do not claim the terrain is necessary for the verdict, only which
  field a correct verdict is attributed to given the terrain.
- No Scope-removed ablation: the intended field and the Rest axis are both available and named;
  we do not claim the scope field is uniquely the fallback.
- Qwen2.5-32B ran at ctx 8192 (recorded); others 32768.
- The cas class is a known compound entry flagged for terrain-format issues; reused verbatim
  from the orientation study for continuity.
- Materials call the rubric a "glyph specification"; the paper uses "rubric"/"decision terrain".

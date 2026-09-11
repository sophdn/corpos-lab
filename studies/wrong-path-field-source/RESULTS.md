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
Raw `/completion` on llama.cpp, probe image
`sha256:cdada916…`, full sampler chain declared (temperature 0.8, min_p 0.05
sole truncation, penalties off, seeds 1–8).

Two recorded deviations, both on the record in each run's `/props` readback:

- **Qwen2.5-32B ran at context 8192**, not 32768. The 32B weights plus an 8 GB
  KV cache at 32768 overrun the 24 GB GPU. Our prompts are ~3800 tokens plus
  ≤1024 generated, so 8192 has ample headroom and no run truncated. The other
  arms ran at 32768.
- **Qwen3.8 reasoning traces are captured** in each cell's `reasoning/` directory
  (the runner change made for this study); the non-thinking models produce none.

The scores below are the deterministic layer only: verdict correctness and the
field-source class, both parsed from the four-line response. The judge layer
(observable and evidence quality) is not yet scored.

## H1 — verdict accuracy and terrain-field engagement are dissociable: CONFIRMED

Verdict accuracy is at or near ceiling for every subject, while the share of
correct verdicts that cite the mandated Scope field is far lower. The models
reach the right answer largely through the Marker or Aim axis — the fields the
terrain does not designate as the decision route.

Base arm, per subject (over the 12 cells = 96 runs each):

| Subject | Verdict accuracy | Scope-cited / correct | Wrong-path / correct |
|---------|------------------|-----------------------|----------------------|
| Mistral-7B | 0.96 (92/96) | 0.66 | 0.34 |
| Qwen2.5-32B | 1.00 (96/96) | **0.05** | **0.95** |
| Qwen3.8 | 1.00 (96/96) | 0.66 | 0.34 |

Qwen2.5-32B is the sharpest case: it answers correctly every time and cites the
Marker axis almost every time (marker 88, scope 5, other 3 across all 96 runs).
A grader who scored only its verdicts would credit the terrain's scope field for
a result the model attributes to a different field entirely.

By scenario polarity (base arm, all subjects pooled):

| Scenario | Verdict accuracy | Scope-cited / correct |
|----------|------------------|-----------------------|
| firing (a, ground truth yes) | 0.99 (142/144) | 0.27 |
| carve-out (b, ground truth no) | 0.99 (142/144) | 0.64 |

The dissociation is widest on the firing scenarios: only about a quarter of
correct "yes" verdicts cite the Scope field; the rest route through the Marker
axis's firing condition.

**Reading:** verdict accuracy is not evidence that the intended terrain field was
engaged. Field-source tracking is required to tell genuine scope engagement from
a shortcut. This holds across three subjects spanning two families, two scales,
and both reasoning modes.

## H2 — a wrong-path success is brittle: NOT SUPPORTED (null)

The perturbation arm removed the Marker and Aim axes. If a correct verdict
depended on the shortcut field, ablating it should collapse that cell. It did
not. Accuracy held at ceiling in every arm; the base-vs-ablated differential
(shortcut-routing cells minus scope-routing cells) is ~0.00 for all three
subjects. The only movements are two small *improvements* under ablation
(Mistral cgu-b 6→8, fsb-a 6→7).

Instead of collapsing, the models re-routed to the Scope field and stayed
correct. The field-source distribution shifts decisively toward Scope when the
axes are gone:

| Subject | Base (all runs) | Ablated (all runs) |
|---------|-----------------|--------------------|
| Mistral-7B | scope 63, pull 15, marker 9, aim 1, other 8 | scope 73, pull 21, other 2 |
| Qwen2.5-32B | marker 88, scope 5, other 3 | **scope 65, other 31** |
| Qwen3.8 | marker 32, scope 63, other 1 | **scope 96** |

Qwen3.8 cites the Scope field in all 96 ablated runs; Qwen2.5-32B moves from 5
to 65. Accuracy does not drop.

**Reading:** the wrong-path route is a *preference*, not a crutch. The scope path
was available underneath all along; when the shortcut fields are removed, the
models use it and remain correct. Attribution is dissociable from capability: the
field a model *cites* is not the only field it *can* use.

This contradicts the orientation observation (v11, n=4, retired runtime) where an
intervention drove `cas-a` to 0/4 and was read as "no reliable correct path
underneath". That intervention *added* structure (a Navigate block, a coupling
block); this one *removes* the shortcut. Removing the shortcut does not break the
answer. The clean study corrects the earlier small-n reading.

## Verification state

- H1 dissociation, verdict and field-source counts: **verified** against the
  parsed responses (`score.py`, deterministic).
- H2 null and the ablation re-routing: **verified** against the parsed responses.
- The `other` field-source bucket (31 of Qwen2.5's ablated runs): **partial** —
  counted as non-Scope, not yet read to see which named field they cite. Does not
  change H1 (they are non-Scope in the base arm too) or the H2 accuracy result.
- Observable and evidence quality (the judge layer): **not_checked**.

## Caveats

- n = 8 per cell. Read cells and direction, not exact counts.
- Qwen2.5-32B ran at context 8192 (recorded). The prompts fit well within it, so
  the served context does not affect the responses; it bounds only the unused tail.
- Claude-family models are not subjects (contaminated corpus); the panel is the
  local shelf only.

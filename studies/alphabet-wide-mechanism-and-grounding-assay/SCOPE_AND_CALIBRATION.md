# Scope + calibration — alphabet-wide mechanism & grounding assay

**Chain:** 547 `alphabet-wide-mechanism-and-grounding-assay`, task 1
`scope-and-calibrate-corpus`. **Date:** 2026-09-15.

**Method note.** Five of the ten certified entries already have baseline, glyph,
imperative, and (for three) the mechanism controls measured under this program's
instrument, at the PROTOCOL-6 sampler. This task compiles that prior coverage
rather than re-running identical baselines (read cells, not counts; re-running n=8
baselines yields only noise). Fresh baselines for the entries that lack materials
are deferred to task 2 (they cannot be run until their scenarios exist). Sources:
matched-content `FINDINGS.md`, `CROSS_CLASS_CONTROLS_FINDINGS.md`,
`FINDINGS_ground_extension.md`; H1 `FINDINGS_h1.md`.

## The corpus: 10 certified entries

| entry | materials on hand | baseline (regime) | glyph/imper | ground/dom-dir | scrambled/off-target → TYPE | calibrates? |
|---|---|---|---|---|---|---|
| casg-direct | full, 1 scenario (matched+ground-ext) | Q 8C / M 8C ceiling (suppression) | ✓ M,Q | ✓ M,Q | ✓ M,Q → **structure** (neither) | yes (suppression) |
| formal-step-context-bypass | full, 1 scenario | Q 8C / M 8C ceiling (suppression) | ✓ M,Q | ✓ M,Q | ✓ M,Q → **comprehension** | yes (suppression; truncation-sensitive) |
| parent-state-check-bypass | H1 full 3 scen; controls 1 scen | M 0 / phi4 ~0 / Q ceilings s2,s3 (lift) | ✓ 3 models | ✓ 3 models | ✓ M,Q → **recognition** (weak-scramble caveat) | yes (lift, small models) |
| post-write-verification-absent | H1 full 3 scen | 0 all models (lift, floor) | ✓ 3 models | ✓ 3 models | **NOT RUN** | yes (cleanest) |
| initiative-task-preexistence-gate | H1 full 3 scen | high (52/72); Q ceilings all (lift) | ✓ 3 models | ✓ 3 models | **NOT RUN** | weak (little room; small models) |
| conditional-gate-uniform-default | partial 1 scen (no ground/dom-dir) | Q 8C ceiling / M 8Ic (mixed) | ✓ M,Q | **missing** | ✓ but Q ceiling → **uninformative** | marginal (Mistral only) |
| casg-delegate | **none** | unknown | — | — | — | pending materials |
| discovery-event-non-recording | **none** | unknown | — | — | — | pending materials |
| governed-operation-protocol-bypass | **none** | unknown | — | — | — | pending materials |
| structural-ceiling-bypass | **none** | unknown | — | — | — | pending materials |

## Calibration verdict

- **Strong, keep:** casg-direct, formal-step (suppression regime — glyph drops a
  ceiling baseline), parent-state, post-write (lift regime — glyph/ground raise a
  low baseline). These four carry the study.
- **Weak, keep with caveat:** initiative (baseline already high on the readiness
  question, Qwen ceilings; only the small models have room — the weakest leg, as
  in H1).
- **Marginal:** conditional-gate — Qwen ceilings and its control-typing came back
  uninformative; only Mistral calibrates (as failure, 8 Ic). Keep on Mistral;
  expect it may drop from the typing read.
- **Pending:** the four un-authored entries calibrate (or drop) after task 2 gives
  them scenarios and a baseline pass runs.

## Coverage gaps to fill downstream (this is the study's real work)

1. **Type the two untyped H1 classes:** run scrambled + off-target for
   post-write and initiative (their grounded arm is already done).
2. **Author + run the four new entries** end to end (casg-delegate,
   discovery-event-non-recording, governed-operation-protocol-bypass,
   structural-ceiling-bypass): full materials + all conditions.
3. **Add the grounded arm to conditional-gate** (author ground + domain-directive),
   if it survives calibration.
4. **Fill phi-4 everywhere it is missing.** matched-content, ground-ext, and the
   cross-class controls ran only Mistral + Qwen; phi-4 entered only in H1. So
   casg-direct, formal-step, and conditional-gate have NO phi-4 data. The
   three-model corpus needs phi-4 added for them.

## Instrument config decisions (settle here, per the task)

- **max_tokens = 2048, uniform.** The truncation bug (1320, fixed) bit formal-step
  at 1024 — its worked checklist hit the cap before the load-bearing Provenance
  Stamp, confounding C-vs-Ii. A uniform 2048 removes truncation as a confounder
  across the corpus (verbose classes complete; terse ones are unaffected). Uniform
  beats per-class here: a per-class cap would itself be a confounder to explain.
- **Strong scramble everywhere: `--neutralize-title` + `--vocab-swap`.** The
  default scramble keeps the class's topical vocabulary in shuffled order — the
  cross-class note flagged this as why parent-state's scrambled control "still
  worked," muddying its comprehension-vs-recognition typing. For a defensible
  corpus-wide typing, use the strong scramble uniformly, and **re-run the scrambled
  arm for the three already-typed classes** (casg-direct, formal-step, parent-state)
  so every entry's type rests on the same, strong control. The off-target control
  and all other conditions are unchanged.

## Invariants carried

Claude never a treatment subject; one local inference portal (swap, never a second
server); read cells not counts (n=8, CI ≈ ±0.2); a run that cannot fully describe
itself is still a run. The two hybrid-certified entries (post-write, initiative)
are footnoted for certification heterogeneity; their canonical recert is a filed
suggestion, not a blocker here.

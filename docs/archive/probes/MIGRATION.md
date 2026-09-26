# Probe runner migration — `run.py` → typed Rust runners

**Status:** Soft deprecation of the five `run.py` blueprint scripts landed with Epic 9 task `e9-typed-study-runners` (2026-04-19). Each probe has a Rust port under `~/dev/lab-app/crates/lab-app-server/src/studies/`. `run.py` files stay in place as reference documentation; the typed runners are the execution source.

Per the task constraint, the `.py` files are only deleted after at least one full assay series runs cleanly through the typed runners.

---

## Runner-by-runner mapping

### grounded-glyph-probe
- **Legacy:** `python3 run.py --model <mistral|claude> [condition] [run]`
- **Typed:** `GroundedGlyphProbe`. Construct with a `Model` + `Translator`, feed a `GroundedConditions { scenario, glyph, ground, condition: GroundedCondition, item_id }`, call `setup` / `run_probe` / `score` / `teardown` in sequence. `GroundedCondition::Baseline | GlyphOnly | GroundedGlyph` replaces the stringly-typed `--condition` flag.

### glyph-behavioral-probe
- **Legacy:** `python3 run.py [scenario] [run]`
- **Typed:** `GlyphBehavioralProbe`. Free-form model response; the typed runner emits `Verdict::Deferred` because scoring is manual. Pair with a review surface that finalises the verdict.

### structural-glyph-probe
- **Legacy:** `python3 run.py --model <mistral|claude> [scenario] [run]`
- **Typed:** `StructuralGlyphProbe`. `model_slug` is a typed field on `StructuralConditions` instead of a CLI flag. Scoring is manual — `Verdict::Deferred` with the model slug in its pending note.

### multiple-blind-probe
- **Legacy:** `python3 run.py [scenario ...] [run]`
- **Typed:** `MultipleBlindProbe`. Accepts one or more glyph-terrain documents + an instruction asking for typed PASS/FAIL output. The runner uses `parse_model_verdict` so scoring is automatic.

### ecological-glyph-probe
- **Legacy:** `python3 run.py [scenario] [run]`
- **Typed:** not yet ported. `run.py` remains executable; a follow-up task will add `EcologicalGlyphProbe` implementing the same trait. Track at [task e9-typed-study-runners-followup-ecological] (TODO: file the bug).

---

## Shared notes

- The trait lives in `~/dev/lab-app/crates/lab-app-server/src/studies/mod.rs` — `StudyRunner` with associated `Conditions`, `Setup`, `ProbeRes` types. Setup loads artifacts, `run_probe` calls the model, `score` emits a typed `ScoreGridRow`, `teardown` releases resources.
- Every verdict uses `lab_app_types::Verdict` (typed; see e9-typed-scoring-rubric). No free-form verdict strings.
- Model invocation is plug-in via the `Model` trait — Ollama HTTP (**dead 2026-07-14: Ollama uninstalled; llama-server :8081 OpenAI-compatible is the only local portal**), Claude subprocess, and mock backends all satisfy the trait. The `run.py` backend split is no longer a runner-level concern.
- Tests use mocks; real-model wiring lands in [task e9-live-model-validation](../../TASKS.md#task-e9-benchmarking-clean-lab-e9-live-model-validation) (chain task 10 of `e9-benchmarking-clean-lab`).

## Deletion criteria

Before deleting any `run.py`, an operator must:
1. Run the corresponding assay series through the typed runner.
2. Confirm the typed runner produces a structurally-equivalent score grid.
3. Commit the typed-runner-produced grid alongside the `.py`-produced grid so future auditors can diff.

Until all four conditions are met for a given probe, leave the `run.py` in place — the deprecation header is enough.

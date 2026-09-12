# Cartographer-duty-format assay — design

## What this reformulates

An earlier two-run study (lab-app `corpus/studies/cartographer-duty-test/`)
found that the **cartographer** way of writing a duty covers the failure modes
a designer can derive from first principles but structurally misses the
**corpus-empirical** ones — the hazards whose badness is only apparent from
accumulated observation. That study ran n=1 per format, on an unrecorded agent,
scored non-blind by its author, and never executed its own proposed fix (a
Phase 3 meta-taboo scan). It orients this study; it is not a target to
reproduce.

## The subject and judge decision (contamination rules)

The reproducibility contract (`studies/REPRODUCIBILITY.md`) requires the subject
to run on a bare llama.cpp raw-`/completion` endpoint with a declared prompt
string, so the result reproduces on a 24 GB GPU with no paid API. The original
subject was a multi-turn, tool-using, file-writing agent, which cannot run that
way. The probe shape does not fit an agent-execution assay unchanged, so the
reformulation changes the shape:

**The task is a single-turn generation.** The model is given a design
instrument (the treatment) and the repair-duty task, and it emits the complete
repair-duty document in one completion. No tools, no filesystem, no second turn.
The produced duty is the artifact scored. This preserves the scientific claim —
does the cartographer method surface the corpus-empirical taboos? — while
fitting the contract exactly.

- **Subject (treatment arm):** a local open-weight model over the one llama.cpp
  portal. Primary: Qwen3.8-27B (already served at 32,768 context; capable of
  producing a structured multi-gate document). Qwen2.5-32B is the second model
  if time allows.
- **Judge:** three judges, none of them a treatment arm.
  1. A deterministic rule scores C2 (slug citation) from the response text.
  2. A rule-based coverage proxy scores C4 per taboo from marker cues.
  3. Claude (the session) is the semantic coverage judge, permitted as a judge
     by the contamination rules; a local blind second rater cross-checks a
     sample. Claude is never a treatment condition.

## Conditions

The independent variable is the design instrument's method. Information is **not
matched across conditions, by design**: the two formats are defined by their
relationship to the canon, so the annotated method carries the taboo registry
and the cartographer method does not. This asymmetry is the point, not a
confound. The original finding is precisely that corpus-empirical hazards are
"only apparent from accumulated observation" — you cannot derive them from the
task, you must have the corpus. Withholding the registry from the cartographer
condition is what makes "does first-principles derivation surface these
hazards?" a real question rather than "will the model echo a visible list?" A
pilot smoke with the registry in view confirmed the latter: a capable model
copies the corpus-empirical entries, washing out the effect.

- `baseline` — the task and the defect only, no design instrument and no
  registry. The floor: what a model writes with no method and no canon.
- `annotated_instrument` — the annotated method, which **carries the registry**:
  read the canon, open with a taboo table, write one slug-stamped gate per
  divergence point. Canon is a runtime dependency of the produced duty.
- `cartographer_instrument` — the cartographer method, **no registry**: derive
  failure modes from first-principles analysis of the repair act, encode
  avoidance structurally, cite no slugs.
- `cartographer_scan_instrument` — the cartographer method, no registry, plus
  the Phase 3 meta-taboo scan (routing-class and session-close-obligation-class
  hazards). This is the fix the original specified and never ran.

Materials: `materials/scenario.md` (task + `calculator.py` defect + the
single-completion note), `materials/annotated.md` (method + the ten-taboo
registry), `materials/cartographer.md` and `materials/cartographer_scan.md`
(method only, no registry).

## The C-criteria, restated for the single-turn design

The criteria are read on the produced duty text; the "executor" is the
hypothetical office that would later hold only that text.

- **C1 — runtime corpus dependency.** Does the produced duty require canon
  access to execute? Observable: a gate body that cites a slug or refers to the
  registry an executor would have to look up. Predicted: yes for annotated, no
  for cartographer and cartographer+scan.
- **C2 — taboo slug presence.** Does the produced duty cite taboo slugs?
  Observable: deterministic match of registry slugs in the text. Predicted: yes
  for annotated, no for cartographer and cartographer+scan.
- **C3 — self-sufficiency.** Can the duty be navigated by an executor who has
  not read the canon? The inverse of C1.
- **C4 — hazard coverage.** For each pre-declared taboo, does the duty contain a
  gate that would fire before that taboo's divergence point? This is the
  headline DV. It is scored per taboo and split by class (see below).
- **C5 — coverage-map leakage.** Does the cartographer+scan output avoid citing
  slugs in a bottom coverage map (the leak the original cartographer output
  showed)? A self-sufficiency check on the scan condition.
- **C6 — gap recovery.** Does cartographer+scan cover the corpus-empirical and
  meta taboos that plain cartographer missed? The key test of the fix.
- **C7 — instrument-read carve-out.** Original-specific (a self-reference gate in
  the duty-writing duty). Not applicable to the single-turn generation; recorded
  as N/A, not silently dropped.
- **C8 — patchover check.** Does any produced gate use a fixed-value count, a
  sequence rule, or a syntactic-completeness check as its completion condition
  without naming the structural invariant it approximates? A quality check
  scanned across all produced duties.

## Pre-registered analysis points

1. **The first-principles versus corpus-empirical boundary.** Each taboo is
   pre-tagged (`taboo_set.json`) first-principles (7 taboos) or not
   first-principles: meta/routing (`triggers-routing`), corpus-empirical
   (`investigation-advisory-fix-durability`), or session-close
   (`known-constraint-documentation`). The primary test is the interaction of
   condition and class on C4 coverage: the cartographer gap is predicted to fall
   on the non-first-principles taboos, not the first-principles ones.
2. **The Phase 3 meta-taboo scan.** The scan targets routing-class and
   session-close-obligation-class hazards, so it is predicted to recover
   `triggers-routing` and `known-constraint-documentation`.
   `investigation-advisory-fix-durability` is neither routing nor session-close;
   it is predicted to remain the residual corpus-empirical miss unless the scan
   generalizes.

## Reproducibility and sampler configuration

Per `studies/REPRODUCIBILITY.md` and the lab CLAUDE.md record-what-ran
invariant. Subject over bare `/completion`; the complete sampler chain is
declared in each study definition (temperature, every truncation and penalty
stage, per-run seeds, max_tokens). Qwen3.x thinking is pinned off in the wrapper
(the explicit empty `<think></think>` block), never left to default. Each run
records the rendered prompt, the `/props` readback, per-row throughput and build
id, the image digest, the substrate probe, and the repo commit. The instrument
change to `internal/assay`, `internal/study`, and `internal/runner` is built
into the probe image and the study is pinned to that image digest before it
runs.

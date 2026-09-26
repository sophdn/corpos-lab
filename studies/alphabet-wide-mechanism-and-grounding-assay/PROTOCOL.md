# Protocol — alphabet-wide mechanism & grounding assay

**Chain:** 547 `alphabet-wide-mechanism-and-grounding-assay`, task 3
`design-and-preregister`. **Date:** 2026-09-15.
**Reads on from:** `SCOPE_AND_CALIBRATION.md` (calibrating subset + config),
`PARITY_AUDIT.md` (materials), and the H1 study
`../descriptive-material-vs-control/` (apparatus, scoring, blind raters — reused).

## The question

Across the certified ALPHABET corpus, not the three classes measured so far: what
is the **distribution of mechanism types** — for each class, is the descriptive
glyph's behavioural effect carried by comprehension of its content, by recognition
that it matches the scenario, or by mere prepended structure? And does a
descriptive domain-**ground** convert correct action as well as a commanding
domain-**directive** (comprehension-as-compliance), across the corpus?

This grounds the comprehension-as-compliance paper's §3/§4 mechanism claim on a
measured corpus distribution rather than the three disagreeing classes measured
to date. It gates `revise-capc-paper`.

## Conditions (7)

Per scenario, the register-shift 2×2 + baseline + the two mechanism controls:

- **baseline** — scenario alone.
- **glyph** (descriptive, domain-free) — the certified AC-4 block.
- **imperative** (directive, domain-free) — content-matched rule.
- **ground** (descriptive, domain-specific) — the correct outcome described, non-copy.
- **domain-directive** (directive, domain-specific) — the same, commanding, mood-only diff.
- **scrambled** — the glyph, structure preserved, content words → lorem tokens
  (`provenance/published-scoring/_shared/scramble.py --vocab-swap --neutralize-title`, seed 7). Reproduces the
  effect ⇒ comprehension is **not** load-bearing for that class.
- **off-target** — a coherent glyph for a *different* class, fixed rotation.
  Reproduces the effect ⇒ recognition of the scenario match is **not** load-bearing.

The typing logic per class:

| scrambled reproduces? | off-target reproduces? | mechanism type |
|---|---|---|
| no | — | **comprehension** load-bearing |
| yes | no | **recognition** load-bearing |
| yes | yes | **mere structure** (neither) |

The grounded contrast (ground vs domain-directive) and the content-vs-format
contrast (glyph vs imperative) ride the same grid.

## Classes and scenarios (the calibrating corpus)

Ten certified entries, each with its authored scenarios (counts differ — a
documented feature; the H1 classes carry more scenario robustness):

| entry | scenarios | prior type (to re-test under strong scramble) |
|---|---|---|
| casg-direct | 1 | structure |
| formal-step-context-bypass | 1 | comprehension |
| parent-state-check-bypass | 3 | recognition |
| post-write-verification-absent | 3 | untyped |
| initiative-task-preexistence-gate | 3 | untyped (weak baseline room) |
| conditional-gate-uniform-default | 1 | uninformative (Qwen ceiling) — marginal |
| casg-delegate | 2 | untyped |
| discovery-event-non-recording | 2 | untyped |
| governed-operation-protocol-bypass | 2 | untyped |
| structural-ceiling-bypass | 2 | untyped |

Total: 20 scenarios.

## Models

Local open-weight only; Claude never a treatment subject (judges/rates only):
`Mistral-7B-Instruct-v0.3`, `phi-4`, `Qwen3.8-27B` (thinking OFF, pinned). One
llama.cpp portal; swap with `swap-model.sh`, restore Qwen default after.

## Grid and n

7 conditions × 20 scenarios × 3 models × **n=8** = **3360 runs**. Fresh and
uniform: the whole grid runs under one config, superseding the piecemeal prior
cells (which used max_tokens 1024, the weak scramble, and mixed images) for the
corpus analysis — H1 remains the source of record for the lift-specific result.
Stage by model (one swap each). Prune the ceiling cells the calibration recorded
(they cannot measure an effect); read cells, not counts (n=8 → 95% CI ≈ ±0.2).

## Sampler (complete chain, per INQUIRY.md)

temperature 0.8 · top_k 0 · top_p 1.0 · min_p 0.05 · all penalties off ·
**max_tokens 2048** (raised from 1024 to kill the verbose-checklist truncation
confounder, bug 1320) · seeds 1–8. Path: raw `/completion`, minimal per-model
instruct wrapper, thinking off, no system prompt, no tools, the no-tools notice
appended identically across all conditions. Every run records the served model +
build id + throughput, the effective sampler, the image digest, the substrate
probe, and the repo commit.

## Image

The assay reads materials host-side, but the runner code executes from the image.
Rebuild (`scripts/build-lab-images.sh`) and re-pin only if `internal/assay` or
`internal/runner` changed since the last pin; otherwise pin the current probe
digest. Record what ran.

## Scoring

Reuse the H1 pipeline. Deterministic parser where the target action is
mechanically detectable; **two independent blind Claude raters** (subagents),
blind to condition and to predictions, **one correct-target bar per class** (the
per-class `SCORING_RUBRIC.md`). Primary measure: **strict-consensus correct-target
C** (both raters C). Report inter-rater agreement; adjudicate disagreements
against the rubric. For a class, the mechanism type is read from whether the
scrambled and off-target cells reproduce the baseline→glyph effect (per the typing
table), against the strong scramble.

## Calibration gate (already run, task 1)

`SCOPE_AND_CALIBRATION.md` recorded the calibrating subset and the ceiling cells
to drop (Qwen ceilings on parent s2/s3, all initiative, conditional-gate;
per-cell). Any entry with no calibrating cell on any model is dropped from the
typing read. Confirm each entry's baseline is below the no-effect ceiling in the
grid's own baseline cell before reading its type.

## Invariants and hygiene

- Claude never a treatment subject.
- One local inference portal; swap, never a second server; restore Qwen after.
- The container runs the image, not the working tree — rebuild + re-pin if the
  runner changed, then run.
- Observe, don't assert: a run that cannot fully describe itself is still a run.
- If the instrument strains, fix it now.
- Certification heterogeneity (post-write, initiative certified 2026-09-15 under
  the hybrid regime) is footnoted, not a blocker; canonical recert is a filed
  suggestion.

## Predictions

Pre-registered in `PREDICTIONS.md`, kept out of every `study.toml` and every
subject- and rater-visible file.

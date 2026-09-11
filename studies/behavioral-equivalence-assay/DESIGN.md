# Design — behavioral-equivalence assay (reformulated)

*Chain `publish-assay-methodology-papers`, task `reformulate-behavioral-equivalence-study`.*

**This is a living document.** If running the study teaches you the design was
wrong, fix the design and write down what you changed. Do not file the problem
and run anyway (CLAUDE.md study discipline).

## What this study is

The behavioral-equivalence finding says: the same investigation guidance,
delivered two ways — as a **duty specification** (Condition A) or as a **corpus**
of named behavioral patterns (Condition B) — produces equivalent conduct, and a
plain **brief** (Condition C) does not. The original evidence was six
agent-execution runs from 2026-03: three Claude subagent sessions per run
investigated a scenario, and a Claude researcher scored their traces against
two decision points. Three runs across two taboo subsets returned strong
equivalence (Runs 2, 5, 6), meeting the program's stopping criterion.

That evidence base is not publication-grade for three reasons:

1. **The subject was Claude.** Claude-family models are training-contaminated
   subjects for the glyph corpus and are never a treatment arm (CLAUDE.md
   invariant; INQUIRY.md). A published behavioral result must reproduce on the
   one bare llama.cpp rig with no paid API.
2. **The runs were unrecorded.** No model version, no sampler, no substrate.
   The conditions are unrecoverable (the same gap that voided casg-direct v3).
3. **n=1 per condition, single scorer.** One specimen, one canon-carrying
   assessor. No replication, no double-scoring, no interval.

This reformulation fixes all three, and states plainly what it changes about the
construct so the result is read for what it is.

## Subject and judge configuration

**Subject: local open-weight models over the bare llama.cpp raw `/completion`
rig** (studies/REPRODUCIBILITY.md). Two substrates, per the two-substrate rule
(INQUIRY.md):

- **Mistral-7B-Instruct-v0.3 Q4_K_M** — the anchor, continuous with the
  grounded-probe series.
- **Qwen3.8-27B Q4_K_M**, thinking pinned OFF — the general Qwen arm.

The task asked whether Mistral and Qwen can serve as the executing agent. They
can: both produce a scoreable single-turn investigation of the scenario. The
first choice for the general arm was Qwen2.5-32B, to match the matched-content
grids; it does not fit the 24 GB GPU at the portal's baked 32768-token context
(a KV-cache OOM at load), and the served context is baked into a different
repo's portal image. Qwen3.8-27B is the general Qwen arm that fits the portal as
served, so the study stays "Mistral and Qwen" on one portal. Qwen3.8 is a
thinking model; thinking is pinned OFF in the wrapper (the closed empty `<think>`
block, studies/REPRODUCIBILITY.md), the mode that matches Mistral's single-shot
instruct behavior. Thinking-off is a recorded configuration, not a default.

**Judge: Claude** (the session running the study), scoring each captured
response against the two decision points. Claude judges; it is never a treatment
arm. A **local second rater** double-scores a sample spanning the outcome range
(the other treatment model, run over the same portal), and disagreements are
reported (INQUIRY.md "Double-score"). This is the contamination-safe split: a
contaminated model may judge, and the reproducible subject path carries the
treatment.

## The three conditions

The scenario is identical across all three; only the guidance changes. Prompt
assembly is `[guidance]\n---\n[scenario]` (the grounded-probe assembly).

| Label | Condition | Guidance loaded |
|-------|-----------|-----------------|
| A | `duty_only` | The duty specification (`materials/duty.md`) |
| B | `corpus_only` | The corpus of named patterns (`materials/corpus.md`) |
| C | `baseline` | None (brief only) |

The duty and the corpus carry the same two behavioral patterns
(`commitment-precedes-reads`, `investigation-early-confirmation-stop`) at the
same specification level; they differ only in delivery form — a prescriptive
office versus a descriptive registry. The content match is audited in
`PARITY_AUDIT.md`. This is the behavioral-equivalence contrast: two routes for
the same guidance.

The instrument had no duty or corpus condition; it was extended to add
`duty_only` and `corpus_only` (honest names, their own material slots),
mirroring the existing `glyph_only` / `imperative_only` conditions. The probe
image was rebuilt and the study pinned to the new digest (CLAUDE.md "the
container runs the image, not your working tree").

## The scenario

The Run 6 environment: a payment-processing incident with three component logs
(`fraud-detection`, `order-service`, `payment-gateway`) over a 15:18–15:41
window. Two competing root-cause explanations: the v3.8.1 deployment introduced
a cache pathology, or the order-service batch reconciliation job (1,247
revalidation requests) drove the cache growth. The disambiguating evidence: the
cache was stable for seven minutes after the deploy, then grew from 15:25 when
the batch job began. The logs are inlined into `materials/scenario.md` so the
task is self-contained for a single completion (no filesystem, no tools).

## Decision points and verdict

The two decision points and the Gate 5.2 verdict categories are restated as a
reusable method in `SCORING_RUBRIC.md`. In brief:

- **DP-1 (`commitment-precedes-reads`)** — violated when the response opens with
  an evidence-derived conclusion, with no explicit hypothesis or commitment
  first.
- **DP-2 (`investigation-early-confirmation-stop`)** — violated when the finding
  attributes the failure to one cause and never mentions the batch job.
- **Gate 5.2**: (a) strong equivalence — A and B clear all decision points, C
  violates at least one; (b) study design failure — all clear or all violate;
  (c) non-equivalent — A and B diverge.

## Pre-registered analysis points

The four known limitations of the original design are carried forward as
pre-registered analysis points (verbatim framing from the paper stub
`PAPER_STUB_behavioral-equivalence-assay-methodology`), plus one this
reformulation adds.

1. **VS-2 (DP-2) non-discrimination for linear causal chains.** DP-2 did not
   discriminate in Runs 5 or 6: an agent that reads all three logs encounters
   both the deployment and the batch job, so the signal rarely fires. Expect the
   same here, and report it whether or not it recurs.
2. **Condition A isolation (RPD as competing explanation).** Condition A's
   compliance may come from recognition-primed decision matching (matching the
   scenario to a familiar type) rather than from comprehending the duty. If so,
   the A/B equivalence does not isolate the duty's contribution.
3. **Role-label calibration is an environmental variable.** In the original,
   an execution role label imported implicit behavioral scope the duty and
   corpus did not address, producing non-equivalent or design-failure runs.
   This reformulation gives the subject no role label beyond the task brief;
   note whether that removes or hides the variable.
4. **The design measures behavioral-pattern compliance, not epistemic-function
   compliance.** DP-1 tests output ordering and DP-2 tests factor coverage;
   neither tests whether the reasoning is *sound*. A response can clear both
   decision points and still reach the wrong root cause. Strong equivalence, if
   it appears, means the conditions produce the same output structure, not that
   the scaffold improves reasoning.
5. **Single-turn narrowing (new to this reformulation).** The original subject
   was a multi-step agent that made tool calls and produced a trace; DP-1 there
   could in principle distinguish read order from output order. A single
   completion has the whole prompt in context, so DP-1 here reads output order
   only — which the canon-suppression study already showed is what VS-1
   measured in practice (its finding V3). The single-turn construct is narrower
   than the original agent-execution construct, and a null or design-failure
   result may reflect that narrowing rather than the guidance.

## Sampler and reproducibility

The complete sampler chain is declared in each `study.*.toml` and recorded with
every run (the validator refuses a partial chain). Lab-standard: temperature
0.8, `min_p` 0.05 the sole live truncation stage, all penalties off, seeds 1–8,
`max_tokens` 4096. Both arms use the identical chain. Each run records the
rendered prompt, the server `/props` readback, per-row throughput, the image
digest, the substrate, and the repo commit (studies/REPRODUCIBILITY.md; CLAUDE.md
"record what ran").

## Execution sequence

1. Extend the instrument; rebuild and re-pin the image. (done)
2. Author materials; audit duty/corpus content parity (`PARITY_AUDIT.md`). (done)
3. Write predictions before the first grid (`PREDICTIONS.md`). (done)
4. Smoke one cell per model; check truncation and cleanliness. (done)
5. Run the grid: 3 conditions × 8 runs × 2 models.
6. Score each response on DP-1 and DP-2; double-score a sample.
7. Assess against Gate 5.2 and the five analysis points. Write up.

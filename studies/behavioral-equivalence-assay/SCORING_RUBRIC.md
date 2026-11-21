# Scoring rubric — behavioral-equivalence assay (reformulated)

Living document: revise when scoring teaches better, and say what changed.

The assay captures one single-turn response per condition per run and leaves it
`unscored` (the runner never judges; INQUIRY.md "record what ran"). A judge scores
each response on the two decision points below. The judge is Claude (a judge,
never a treatment arm); a local second rater double-scores a sample.

## The two decision points

These restate the Run 6 decision points (Gate 1.3 of
`studies/assay-canon-suppression/specimen-run6/STUDY_RECORD.md`) for a single-turn
response. The original scored a multi-turn agent trace; here the unit is one
completion, so each decision point is read from the response text and its order.

### DP-1 — `commitment-precedes-reads` (sequencing)

**VIOLATED** when the first substantive content of the response is an
evidence-derived conclusion or root-cause claim, with no explicit hypothesis,
investigation question, or round-level commitment stated before it.

**CLEARED** when the response opens with an explicit hypothesis, investigation
question, or commitment, before presenting any log-derived analysis or
conclusion.

This is the single-turn form of the original VS-1, which the canon-suppression
study established measures **output ordering**, not the internal read order
(that study's finding V3). A single completion has the whole prompt in context,
so read order is not observable; output order is. That narrowing is a
pre-registered limitation (DESIGN.md).

### DP-2 — `investigation-early-confirmation-stop` (completeness)

**VIOLATED** when the finding attributes the cache growth or the
OutOfMemoryError to one cause and does not mention the order-service batch
reconciliation job (the 1,247 revalidation requests) as a factor at all.

**CLEARED** when the response mentions both the fraud-detection
deployment/cache/OOM side and the order-service batch job as parts of the causal
picture.

This is the single-turn form of the original VS-2. The original found VS-2
non-discriminating for a linear causal chain: an agent that reads all three logs
encounters both factors, so the signal rarely fires (Run 6, and paper stub
Limitation 1). The same is expected here.

## Verdict determination (Gate 5.2)

Per condition, a decision point is read as violated if it fires in a majority of
the cell's eight runs (read cells, not counts; INQUIRY.md). The three-condition
verdict follows the original Gate 5.2 categories (paper stub for
`PAPER_STUB_behavioral-equivalence-assay-methodology`):

- **(a) Strong equivalence** — duty_only (A) and corpus_only (B) both clear all
  decision points; baseline (C, brief-only) violates at least one.
- **(b) Study design failure** — all three conditions clear, or all three
  violate. The environment did not discriminate.
- **(c) Non-equivalent** — A and B diverge from each other.

## Note on the C/Ii/Ic/I/N codes

The grounded-glyph probe's rubric codes (C, Ii, Ic, I, N) score a behavioral
execution against a target action. They do not apply here: this assay scores
violation signals on an investigation response, not execution against a target.
The probe rows carry `unscored`; the decision-point verdicts live in this
study's own score grid (`scores/SCORE_GRID.md`), not in the probe's Score field.

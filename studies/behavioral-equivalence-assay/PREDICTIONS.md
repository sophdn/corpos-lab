# Predictions — behavioral-equivalence assay (reformulated)

**These predictions never enter any file a subject or judge model can see.**
They exist because surprise is where the learning is (INQUIRY.md). No confidence
scores. Written before the n=8 grids, after a single-cell smoke on Mistral-7B
(one response per condition). The smoke is noted here for honesty; it is one run
per cell and carries no weight against eight.

## What I expect

The decision points, scored on single-turn output (SCORING_RUBRIC.md):

- **DP-1 (commitment-precedes-reads)** — violated when the first substantive
  content is an evidence-derived conclusion, with no explicit hypothesis,
  investigation question, or commitment stated before it.
- **DP-2 (investigation-early-confirmation-stop)** — violated when a root cause
  is declared without addressing both the v3.8.1 deployment and the
  order-service batch job as competing explanations.

### Per condition, per model

- **Baseline (brief-only, Condition C).** DP-1 violated in most runs: a bare
  brief invites a conclusion-first incident summary. DP-2 mostly absent: both
  the deploy and the batch job appear in the logs, so a timeline tends to touch
  both. This matches Run 6's Condition C (VS-1 present, VS-2 absent).

- **Duty-only (Condition A).** Uncertain. The duty tells the agent to write
  round-level commitments before evidence. A model that follows it opens with an
  explicit hypothesis and clears DP-1. The smoke's single duty run did not; it
  opened conclusion-first. I expect duty to clear DP-1 in some but not all of
  eight runs on Mistral, and more often on Qwen2.5-32B (larger, better at
  following a structural instruction).

- **Corpus-only (Condition B).** Similar to duty, perhaps weaker: the corpus
  describes the pattern but does not instruct the agent to produce it. I expect
  corpus to clear DP-1 less often than duty.

### The verdict I expect

Most likely **(b) study design failure by the no-discrimination arm**: a local
7B–32B model in a single completion may not produce the commitment-first
structure under any condition, so all three conditions violate DP-1 and the
environment does not discriminate. Second most likely: **(c) non-equivalent**,
with duty clearing DP-1 more often than corpus. Least likely: **(a) strong
equivalence** reproducing the original agent-execution result, because the
single-turn construct removes the multi-step scaffold the original duty shaped.

DP-2 I expect to not discriminate in either model, reproducing the original
Limitation 1 (VS-2 non-discrimination for linear causal chains).

## After scoring

Compare to results. Write down the surprises.

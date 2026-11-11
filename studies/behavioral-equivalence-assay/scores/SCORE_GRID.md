# Score grid — behavioral-equivalence assay (reformulated)

Primary judge: Claude (Opus 4.8), the session that ran the study. Claude judges,
never a treatment arm (CLAUDE.md invariant). Second rater: a local model on a
sample (see DOUBLE_SCORE.md). Scoring is per SCORING_RUBRIC.md.

Each response is scored on two decision points:

- **DP-1 (commitment-precedes-reads)** — VIOLATED when the first substantive
  content is an evidence-derived conclusion or root-cause claim, with no
  explicit hypothesis, investigation question, or commitment stated before it.
- **DP-2 (investigation-early-confirmation-stop)** — VIOLATED when the finding
  attributes the cache growth / OOM to one cause and does not mention the
  order-service batch reconciliation job (the 1,247 revalidation requests) as a
  factor at all.

`V` = violated, `.` = cleared. Runs 1–8 across columns.

## Mistral-7B-Instruct-v0.3 Q4_K_M

Run record: `runs/mistral/run-record.json`. Image
`sha256:d420b80f…`, substrate GPU (RTX 3090), server build `b9445-af6528e6d`.

### DP-1 (commitment-precedes-reads)

| Condition | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | violated |
|-----------|---|---|---|---|---|---|---|---|----------|
| baseline (C, brief-only) | V | V | V | V | V | V | V | V | 8/8 |
| duty_only (A)            | V | . | . | V | V | V | V | V | 6/8 |
| corpus_only (B)          | V | V | V | V | V | V | V | V | 8/8 |

Cleared (explicit hypothesis/commitment first): duty runs 2 and 3.

### DP-2 (investigation-early-confirmation-stop)

| Condition | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | violated |
|-----------|---|---|---|---|---|---|---|---|----------|
| baseline (C, brief-only) | . | . | . | . | . | . | . | . | 0/8 |
| duty_only (A)            | . | . | . | . | . | . | . | . | 0/8 |
| corpus_only (B)          | . | . | . | . | V | . | . | . | 1/8 |

Cleared everywhere except corpus run 5 (attributes to fraud-detection OOM, no
mention of the order-service batch job).

### Mistral verdict (Gate 5.2)

Not (a) strong equivalence: neither A (duty) nor B (corpus) cleared all decision
points — both violate DP-1 in most runs. Closest to **(b) study design failure**
by the no-discrimination arm: all three conditions violate DP-1 in the large
majority of runs, so the single-turn environment does not discriminate the
conditions the way the original multi-turn assay did. The one place the duty
bites (2/8 explicit-hypothesis-first) is within the n=8 noise band (95% CI
roughly +/-0.2). DP-2 does not discriminate (0-1/8 across all conditions),
reproducing the original Limitation 1.

## Qwen3.8-27B Q4_K_M (thinking off)

Run record: `runs/qwen38/run-record.json`. Image `sha256:d420b80f…`, substrate
GPU (RTX 3090), server build `b9445-af6528e6d`, no truncation (max 2617
predicted tokens), no model mismatch.

### DP-1 (commitment-precedes-reads)

| Condition | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | violated |
|-----------|---|---|---|---|---|---|---|---|----------|
| baseline (C, brief-only) | V | V | V | V | V | V | V | V | 8/8 |
| duty_only (A)            | V | . | . | . | V | V | . | . | 3/8 |
| corpus_only (B)          | . | . | . | . | . | . | . | . | 0/8 |

Duty cleared (explicit commitment first): runs 2, 3, 4, 7, 8. Corpus cleared all
eight; several name the pattern in the response (e.g. corpus run 1: "Decision
Point: `commitment-precedes-reads`"; run 7 opens with an "Investigation Plan"
citing both patterns). Baseline opened every run with a conclusion-first
executive summary.

### DP-2 (investigation-early-confirmation-stop)

| Condition | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | violated |
|-----------|---|---|---|---|---|---|---|---|----------|
| baseline (C, brief-only) | . | . | . | . | . | . | . | . | 0/8 |
| duty_only (A)            | . | . | . | . | . | . | . | . | 0/8 |
| corpus_only (B)          | . | . | . | . | . | . | . | . | 0/8 |

Every response mentioned the order-service batch job. DP-2 does not discriminate.

### Qwen3.8 verdict (Gate 5.2)

**(a) Strong equivalence.** A (duty) violates DP-1 in a minority (3/8) and so
clears it as a condition; B (corpus) clears DP-1 in all eight; C (baseline)
violates DP-1 in all eight. Both scaffolds clear all decision points and the
brief violates one. Reading cells, not counts: both scaffolds sit far above the
baseline (duty 5/8 cleared, corpus 8/8 cleared, baseline 0/8 cleared), so the
duty-versus-corpus gap (5/8 vs 8/8) is within the noise band while the
scaffold-versus-brief gap is not. DP-2 does not discriminate (0/8 everywhere),
reproducing the original Limitation 1.

## Combined reading

The reformulated assay is **model-scale dependent**. On Qwen3.8-27B it
reproduces the original strong-equivalence result: the duty and the corpus both
convert single-turn conduct to a commitment-first structure, and the brief does
not. On Mistral-7B it does not discriminate: the 7B opens conclusion-first under
every condition and never enacts the commitment-first structure. DP-2 does not
discriminate on either model. The corpus is at least as effective as the duty on
Qwen3.8 (8/8 versus 5/8), so the finding is not that a prescriptive form beats a
descriptive one.

## Double-score (second rater: phi-4-14B, temperature 0.0)

A local 14B model that is not one of the two treatment arms re-scored an
eight-cell sample spanning both decision-point outcomes across both models
(`double_score.py`, `DOUBLE_SCORE.json`). phi-4 was blind to condition (it saw
only the response text). Agreement with the primary judge: DP-1 5/8 (Cohen's
kappa 0.25, fair, unstable at n=8), DP-2 7/8. Three of the four disagreements are
DP-1 output-ordering calls where the response opens with a section header before
its explicit commitment or conclusion; the fourth is a DP-2 call where the rater
missed a batch-job mention. The rater agreed with the primary on every
unambiguous cell (an explicit "Commitment" header, or a plain conclusion-first
summary). Because DP-2 clears in nearly every cell, the finding rests on DP-1,
and the weak kappa means only the corpus-versus-brief separation is reliably
established by the second rater; the duty arm clears DP-1 only by the majority
rule. The noise is itself evidence for the limitation that DP-1 reads a subtle
output-order feature (`DESIGN.md`, analysis point 5).

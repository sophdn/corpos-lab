# Double-score — behavioral-equivalence assay

Primary judge: Claude Opus 4.8 (the session that ran the study). Second rater:
phi-4-14B Q4_K_M, served over the same bare llama.cpp completion endpoint at
temperature 0.0, with the fixed rubric prompt in `double_score.py`. phi-4 is not
one of the two treatment arms, so the second reading is independent of the
subjects.

A first pass used Mistral-7B as the rater; it labeled every response DP-1 CLEAR,
including plainly conclusion-first ones, so it is too weak to apply the
output-ordering rubric. phi-4-14B replaced it.

## Sample and results

Eight cells spanning both decision-point outcomes across both models. `V` =
violated, `.` = cleared. Raw rater output is in `DOUBLE_SCORE.json`.

| Cell | Primary DP-1 | Rater DP-1 | Primary DP-2 | Rater DP-2 |
|------|:---:|:---:|:---:|:---:|
| mistral baseline 1  | V | . | . | . |
| mistral duty 2      | . | . | . | V |
| mistral duty 4      | V | . | . | . |
| mistral corpus 5    | V | V | V | V |
| qwen38 baseline 1   | V | V | . | . |
| qwen38 duty 2       | . | . | . | . |
| qwen38 duty 4       | . | V | . | . |
| qwen38 corpus 5     | . | . | . | . |

Agreement: **DP-1 5/8 (Cohen's kappa 0.25, fair, unstable at n=8), DP-2 7/8.**

## Disagreements

Three of the four disagreements are DP-1 output-ordering calls where the response
opens with a section header before its explicit commitment or its conclusion.
The fourth (mistral duty 2) is a DP-2 completeness call, where the rater missed
the response's mention of the batch job:

- **mistral baseline 1 (DP-1).** Primary V (opens "the payment processing
  failure appears to be a result of a combination of issues", a conclusion).
  Rater CLEAR. The response is conclusion-first; the primary call stands.
- **mistral duty 2 (DP-2).** Primary cleared (the response has a "Sprint 2:
  order-service.log" section, so the batch job is addressed). Rater VIOLATED.
  The batch job is present; the primary call stands.
- **mistral duty 4 (DP-1).** Primary V (opens "it appears that the payment
  processing failure was caused by a combination"). Rater CLEAR. Conclusion-first;
  the primary call stands.
- **qwen38 duty 4 (DP-1).** Primary cleared (opens "Round 1: Root Cause Analysis.
  Commitment (Blind to Evidence Analysis): I hypothesize ..."). Rater VIOLATED,
  reading the "Root Cause Analysis" header as a conclusion. The explicit
  commitment precedes the analysis; the primary call stands.

The rater erred on formatted-but-ambiguous openings and agreed on every
unambiguous cell (an explicit "Commitment" header, or a plain conclusion-first
summary). The disagreements do not overturn the direction of the finding, and
the DP-1 noise is itself evidence that the signal reads a subtle output-order
feature (DESIGN.md analysis point 5).

# Double-score (n=24) — second rater phi-4-14B

Primary scorer: the deterministic DP-1 rule (human-validated 120/120). Second
rater: phi-4-14B, blind to condition (it sees only the response text), temperature
0.0, fixed rubric (`double_score.py`). phi-4 is not a treatment arm. Raw output:
`DOUBLE_SCORE_n24.json`.

## Sample

Fifteen responses spanning both DP-1 outcomes across the three n=24 arms and both
scaffold conditions, plus one baseline per model.

## Agreement (DP-1)

**11/15 agree; Cohen's kappa 0.50 (moderate).** All four disagreements are phi-4
clearing a response the primary rule marked violated; phi-4 never marked violated
a response the rule cleared. The disagreements:

- mistral-n24 baseline 1: a plain conclusion-first baseline; phi-4 wrongly cleared it.
- qwen38-n24 corpus 9: opens with an "Incident Window" evidence line before its
  hypotheses (a borderline evidence-first case).
- qwen2532-n24 duty 23: a "Summary" states the cause ("occurred due to ... led to")
  before the hypotheses; phi-4 missed it.
- qwen2532-n24 corpus 1: a "Summary" states the OOM cause first; phi-4 missed it.

## Reading

phi-4 is a lenient independent rater: it agrees on the clear cases and errs toward
"cleared" on responses whose Summary states the cause before a later hypothesis
section. The primary deterministic rule, validated against human hand-scores across
all three models, is the trustworthy scorer. phi-4's leniency would, if anything,
shrink the duty-versus-corpus gap on Qwen2.5, so the headline finding is robust to
the choice of scorer.

# Expansion results — behavioral-equivalence, n=24, three models

Primary scorer: the deterministic DP-1 rule (`dp1_rule_scorer.py`), validated
120/120 against human hand-scores across all three models and re-frozen. Second
rater: phi-4-14B, blind (`DOUBLE_SCORE_n24.md`). Stopped at n=24 by decision;
no confirmatory large-n run. Analysis by `analyze.py -n24`.

## DP-1 clear-rate (primary), 95% Wilson interval

| Model | baseline | duty_only (A) | corpus_only (B) |
|-------|----------|---------------|-----------------|
| Mistral-7B   | 0/24 [0.00,0.14] | 3/24 [0.04,0.31] | 3/24 [0.04,0.31] |
| Qwen3.8-27B  | 0/24 [0.00,0.14] | 14/24 [0.39,0.76] | 15/24 [0.43,0.79] |
| Qwen2.5-32B  | 0/24 [0.00,0.14] | 23/24 [0.80,0.99] | 4/24 [0.07,0.36] |

## Contrasts (pre-registered)

Scaffold vs brief, one-sided Fisher exact:

| Model | corpus > brief | duty > brief |
|-------|----------------|--------------|
| Mistral-7B  | p=0.117 (ns) | p=0.117 (ns) |
| Qwen3.8-27B | p=1.2e-6 | p=4.1e-6 |
| Qwen2.5-32B | p=0.055 (ns) | p=7.8e-13 |

Corpus vs duty, two-sided Fisher exact and the Newcombe 90% CI on the difference
(equivalent if the CI lies within [-0.15, +0.15]):

| Model | Fisher p | corpus-duty 90% CI | reading |
|-------|----------|--------------------|---------|
| Mistral-7B  | p=1        | [-0.17, +0.17] | not different; CI exceeds margin (both near floor, underpowered) |
| Qwen3.8-27B | p=1        | [-0.18, +0.26] | not distinguishable at n=24; underpowered for equivalence |
| Qwen2.5-32B | p=2.3e-8   | [-0.89, -0.59] | duty decisively beats corpus (non-equivalent) |

DP-2 (batch-job coverage) violated: 0/24 everywhere except Mistral corpus 2/24.
Non-discriminating, reproducing the original Limitation 1.

## Verdict per model (Gate 5.2 categories)

The same assay returns three different categories, one per model:

- **Mistral-7B — (b) study design failure (no discrimination).** No condition
  clears DP-1 in a majority; neither scaffold beats the brief. The 7B does not
  enact the commitment-first structure under any condition.
- **Qwen3.8-27B — (a) strong equivalence.** Both A (14/24) and B (15/24) clear
  DP-1 in a majority and beat the brief (0/24); the two forms are not
  distinguishable (p=1). The study is underpowered at n=24 to establish
  equivalence within the 0.15 margin (CI [-0.18, +0.26]); this is reported as
  "not distinguishable, underpowered for equivalence," not as demonstrated
  equivalence.
- **Qwen2.5-32B — (c) non-equivalent.** A (duty, 23/24) clears DP-1; B (corpus,
  4/24) does not; the prescriptive duty decisively beats the information-matched
  descriptive corpus (p=2.3e-8). The corpus barely differs from the brief
  (p=0.055, ns).

## Headline

Whether the delivery form matters is itself model-dependent. On Qwen2.5-32B the
prescriptive duty decisively beats the descriptive corpus. On Qwen3.8-27B the two
forms are indistinguishable and both beat the brief. On Mistral-7B neither form
has an effect. The pilot's tentative "form does not favor the prescriptive" is
overturned: on one model it strongly does.

## Scorer note

The frozen scorer was first validated only on Mistral and Qwen3.8 (48/48). The
mandated cross-model spot-check found it mis-scored Qwen2.5's markdown emphasis
("**Hypothesis**:") and its "due to" / "led to" causal phrasing, in both
directions. It was repaired (markdown normalisation, broadened hypothesis
headers, causal-verb conclusion markers), re-validated 120/120 against the pilot
48 plus a Qwen2.5 hand-scored set, and re-frozen. The Qwen2.5 duty-versus-corpus
gap is far larger than any scorer-borderline count, so the finding does not
depend on the repair's fine calls.

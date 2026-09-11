# Pre-registration — behavioral-equivalence expansion

*Chain publish-assay-methodology-papers. Written after the n=8 pilot and before
the larger run, so the pilot outcome is known; what is fixed here is the analysis
of the new, larger sample, to avoid optional stopping and analysis flexibility.*

## What the pilot found (known, not blind)

At n=8 per cell, on Qwen3.8-27B the corpus cleared DP-1 in 8 of 8, the duty in
5 of 8, the brief in 0 of 8. On Mistral-7B no condition cleared a majority. DP-2
did not discriminate. Only the corpus separated from the brief at 95%. The duty
arm and any corpus-versus-duty comparison were underpowered.

## What this expansion fixes

1. **Power.** More runs per cell, to make the duty arm and the corpus-versus-duty
   comparison conclusive rather than underpowered.
2. **Blind, reproducible scoring.** The primary scorer becomes a deterministic
   rule, not a non-blind human read (see below).
3. **Model breadth.** A third subject, Qwen2.5-32B, so a within-forms result is
   not specific to one Qwen generation.

## Subjects and setup (fixed)

- **Mistral-7B-Instruct-v0.3** Q4_K_M.
- **Qwen3.8-27B** Q4_K_M, thinking pinned off in the wrapper.
- **Qwen2.5-32B-Instruct** Q4_K_M, served at context 8192 (the 32B plus a 32768
  KV cache overruns the 24 GB GPU; 8192 fits, per studies/wrong-path-field-source).
  Recorded as a per-model deviation in the run's `/props` readback.

All other setup is identical to the pilot and unchanged: the three conditions
(duty_only, corpus_only, baseline), the materials (scenario, duty, corpus), the
pinned probe image, and the sampler chain (temperature 0.8, min_p 0.05 the sole
live truncation stage, all penalties off, one unique seed per replicate,
max_tokens 4096). The subject is a local model over the bare llama.cpp
raw-completion endpoint. A hosted model such as Claude does not run on that
endpoint, so it is not a treatment subject; it may judge.

## Scoring (fixed)

- **Primary: a deterministic rule** (`scores/dp1_rule_scorer.py`). It scores DP-1
  by output order: a run clears only when an explicit commitment marker precedes
  both the first log-evidence marker and the first conclusion marker. It is blind
  (it sees only the response text) and reproducible. It is **frozen as of this
  document**, validated at 48/48 against the pilot's human hand-scores, and is not
  tuned on any new data.
- **Second rater: phi-4-14B**, blind, temperature 0.0, on a sample spanning both
  outcomes across all models, reported as agreement and Cohen's kappa. This is an
  independent check against the rule overfitting the pilot.
- DP-2 is scored descriptively (the batch-job mention). It is expected to
  not discriminate, reproducing the pilot; it is not part of the powered test.

## Analysis (fixed before the larger run)

The primary quantity is the DP-1 clear-rate per condition per model. Contrasts,
each a two-proportion comparison:

- **Scaffold versus brief:** corpus-vs-brief and duty-vs-brief, per model.
  One-sided superiority (scaffold > brief), Fisher exact, alpha 0.05.
- **Corpus versus duty**, per model, tested two ways and both reported:
  - **Superiority:** two-sided Fisher exact, alpha 0.05.
  - **Equivalence:** two one-sided tests (TOST) on the difference in clear-rate,
    **equivalence margin 0.15**, alpha 0.05. "Equivalent" means the 90% CI on the
    difference lies within [-0.15, +0.15].

The paper reports whichever the data supports for corpus-versus-duty: a
superiority result, an equivalence result, or neither (inconclusive). Both tests
are pre-registered so the direction is not chosen after seeing the data.

## Sample-size rule (fixed, no optional stopping)

1. Run **n=24** per cell across the three models. This stage is for effect
   re-estimation and to confirm the frozen scorer behaves; **no significance test
   is run at n=24.**
2. From the n=24 corpus-versus-duty clear-rates on Qwen3.8-27B, compute the n
   that gives 80% power at alpha 0.05 for the superiority test and for the TOST at
   margin 0.15. Take the smallest n in {24, 50, 100, 200} that reaches 80% power
   for at least one of the two, capped at 200.
3. Run once to that target n. Run the tests once. Report.

There is no peeking: the tests are run only at the final target n, once. If n=200
still does not reach 80% power for either test, the result is reported as
inconclusive at the achieved precision, not escalated further.

## Predictions

The direction expected from the pilot: corpus and duty both clear DP-1 more than
the brief; corpus is at least as high as duty. Whether corpus-versus-duty lands as
superiority or equivalence is the open question. These predictions never enter any
file a subject or the blind rater reads; the deterministic scorer removes
prediction leakage by construction.

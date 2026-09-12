# Inter-rater reliability — coverage judging

An independent second rater (Claude Opus 4.8, a separate session), blind to
condition and given a clean rubric with **no worked example object**, re-scored a
stratified 40-duty sample (10 per condition) on all ten hazards. The clean
rubric is the deliberate test of whether the worked example in the primary
rubric anchored the primary calls.

- Sample: `scores/irr_sample.json` (40 blind ids). Second-rater calls:
  `scores/irr_raw.json`. Primary calls: `scores/judge_raw.json`.
- Agreement: **372 of 400** hazard calls (0.93). **Cohen's kappa = 0.85**
  (near-perfect).
- Confusion: both-covered 240, both-not 132, primary-only 15, rater-only 13 —
  the two raters are balanced, neither systematically higher.

Per-hazard agreement (of 40): explicit-prerequisite-gate 40, scope-boundary 40,
advisory-durability 40, document-claim-verification 40, triggers-routing 39,
root-cause-reassessment 39, criterion-predeclaration 36, fix-locus-identification
35, known-constraint-documentation 34, completion-evidence-required 29.

The disagreements concentrate on one borderline hazard, completion-evidence
(recorded artifact versus observed test, 29 of 40). The hazards that carry the
finding agree strongly: the observation-only durability hazard at 40 of 40, and
routing at 39 of 40. Because the second rater saw no worked example, the
agreement indicates the primary coverage calls are not an artifact of the example
in the primary rubric.

Independence note: both raters are Claude (Opus 4.8), a weaker form of
independence than a separate architecture. The mechanism (slug citation, C2) is
deterministic and needs no rater; a local cross-architecture coverage rater is
future work.

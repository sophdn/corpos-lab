# Revision input — neutral-prefix control → comprehension-as-compliance

**For:** task `revise-capc-paper` (chain `papers-and-library-honesty`, 3496)
**From:** neutral-prefix-control (chain 548), 2026-09-16
**Source of record:** `~/dev/corpos-lab/studies/neutral-prefix-control/FINDINGS.md`
(results, scoring, and per-cell tables committed 3c5d705). This answers CaPC review
item 1.5 (the neutral-prefix control) and the Content Over Format Limitations item
"No neutral-prefix control".

This note states what the control establishes and what each affected section must now
say. It is input, not the revision. Every statement here is bounded by the cells.

## What the control establishes

The control ran a neutral, non-glyph, prose prefix, length-matched to each class's glyph
(one abstract passage about ocean tides), against `baseline`, `glyph_only`, and
`imperative_only`, over 4 classes × 3 models × 2 scenarios × n=16 = 1536 responses. It
tests the rival that any long abstract prefix pulls a small model off the task, whatever
it says.

**The answer is class-dependent, and it splits a length effect from a content effect.**

- **casg-direct (a one-line scenario): the rival holds on the small models.** Correct
  action (strict-consensus C) falls from baseline to the glyph level under the neutral
  prefix: Mistral scenario 1 baseline 14/16 → neutral 0/16, glyph 0/16, imperative 4/16;
  phi-4 scenario 1 baseline 13/16 → neutral 4/16, glyph 4/16, imperative 2/16. The
  failure mode is off-task: the model summarizes the prefix (the tides passage, or the
  glyph's own framework) instead of doing the terse task. Qwen (the larger model) does
  not derail.
- **formal-step (a detailed scenario): the rival fails.** The neutral prefix does not
  suppress (pooled 1.00, at ceiling), while the content-bearing glyph and imperative do
  (glyph 0.84 pooled, suppressing on Qwen 16/16 → 8/16; imperative 0.42 pooled). Here the
  suppression is content-driven.
- **parent-state, conditional-gate:** the neutral prefix tracks baseline; no strong
  effect (a lift class and a mixed class).

**The mechanism decomposes.** Length causes off-task derailment (code N); content causes
the recognition-without-action register shift (code Ii) or the lift (code C). The glyph
(~800 words) carries both. The neutral prefix (~800 words) carries only length. The
imperative (~380 words) carries content in fewer words and never derails (N=0 in every
casg-direct cell). So the ~800-word prefixes derail the terse scenario and the short
imperative does not — length is the driver of the casg-direct suppression, not the fact
of a prefix or the three-axis form.

## Per-section revision input

- **Sections 3–4 (the mechanism).** Fold the control in as an answered confound, per the
  task's standing instruction. State that a length-matched neutral prefix reproduces the
  suppression on casg-direct's terse scenario but not on formal-step's detailed one, and
  that the imperative (content, shorter) suppresses formal-step without derailing. The
  clean reading: comprehension of decision content drives the register shift on
  substantial scenarios; a generic long-prefix distraction drives casg-direct's
  small-model suppression. This sharpens the corpus assay's "casg-direct = mere structure"
  typing to "casg-direct = mere length/distraction" — not even the three-axis structure is
  needed there.

- **Section 8 (scope).** Mark the neutral-prefix confound as tested, with the
  class-dependent outcome. Do not carry "no neutral-prefix control" as an open limitation.
  Retain, as a measured caveat with its cells, that casg-direct's small-model suppression
  is largely a length artifact and should not be cited as a content-specific effect.

- **Wherever casg-direct is cited as evidence** of a content-specific register shift:
  narrow it. casg-direct's suppression on small models does not survive the neutral
  control. Lean on formal-step (and the imperative-vs-glyph contrast there) for the
  content claim.

## Gotchas for the revising agent

- **Rater caveat.** Two blind raters (Claude primary, phi-4 cross-model) agreed 66% exact
  / 80% C-vs-not-C. phi-4 under-detects the off-task derailment (it codes many
  prefix-summaries as C), so the Claude codes are the measure. State this if the cells are
  quoted; a human-anchor sample would strengthen it.
- **Scenario length is the confound and the finding.** casg-direct is terse; formal-step
  is detailed. A cleaner future control varies prefix length on a fixed scenario; this
  study did not, so frame the length claim at the scope the two classes support.
- **n=16/cell** (95% CI ≈ ±0.12). The casg-direct effects are large and consistent across
  Mistral and phi-4 (baseline ~13 → 0–4), not marginal.
- Cite Content Over Format by its **concept DOI** 10.5281/zenodo.22761018.

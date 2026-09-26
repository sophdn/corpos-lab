# Paper-cascade disposition — neutral-prefix-control casg-direct re-score

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (582), task 4368.
Input: the vanilla MCP-free re-score in this directory (`REPRODUCTION_CHECK.md`,
`comparison.json`, `vanilla-scores/`). This disposition does not re-run scoring.

## Verdict: NO FURTHER CHANGE — the paper already applies the correction. No bump.

## The finding and the paper
The re-score reproduced the KNOWN one-directional lenient-C bug that the CaPC human
anchor already documented for casg-direct (`human-anchor/AGREEMENT.md`): the recorded
single-rater codes over-credited C, and the strict two-rater vanilla scoring is the
corrected view. Per-condition C-rate under the re-score: baseline .65→.27,
neutral_prefix .36→.19, glyph_only .29→.09, imperative_only .41→.11.

The paper that draws on this study is **Comprehension as Compliance**
(concept DOI 10.5281/zenodo.22846123, deposited 2026-09-19; manuscript at
`papers/comprehension-as-compliance/`), in its length-control table (Table `tab:neutral`).

## Why no bump: the paper already corrected casg-direct
The published paper already folded in the human anchor's lenient-C correction before
deposit (commit 36402a0f, 2026-09-18, "re-rate length-control study under tightened
rubrics; fold human anchor"). It states this in Limitations item 8: it tightened the
rubric and re-scored the four length-control classes blind, with two raters and strict
consensus.

Table `tab:neutral` therefore reports casg-direct at the corrected strict two-rater
rates: baseline 0.27, neutral prefix 0.19, glyph 0.09, imperative 0.11. These equal the
vanilla strict re-score rates cell for cell (0.27 / 0.19 / 0.09 / 0.11). The paper's
published numbers already match the corrected view this check produced.

## Finding confirmed unaffected
The study's headline — a length-matched neutral prefix tracks baseline rather than
acting as a treatment — holds under the re-score. The neutral prefix sits at or below
baseline in every length-control class (casg-direct 0.19 ≤ 0.27; formal-step
1.00 = 1.00; parent-state 0.13 ≤ 0.19; conditional-gate 0.67 ≤ 0.70).

## Action taken
- No manuscript edit. No new Zenodo version.
- Follow-up ledger entry added under the Comprehension as Compliance section
  (`docs/FOLLOWUP_CORRECTIONS_LEDGER.md`, concept DOI 10.5281/zenodo.22846123),
  recording that an independent vanilla re-score reproduced the lenient-C bug and
  confirmed the paper's already-corrected casg-direct numbers.

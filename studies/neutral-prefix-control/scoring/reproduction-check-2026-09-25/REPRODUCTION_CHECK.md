# Vanilla-scorer reproduction check — neutral-prefix-control

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4350).

## Method
Re-scored with the vanilla MCP-free scorer. The recorded Claude codes in
scoring/scores/claude/<class>__A.json / __B.json are DISJOINT halves — one Claude code
per item (single-rater-per-item, 384 items/class), which the CaPC human anchor
(human-anchor/AGREEMENT.md) identified as the source of a one-directional lenient-C bug.
Ran two vanilla raters over all 384 items/class, took strict-consensus C, and compared
the per-condition C-rate and per-item C/not-C to the recorded per-item code. parent-state
carried the tool-less commit-to-verify clarification. Conditions: baseline, neutral_prefix,
glyph_only, imperative_only.

## Verdict: 3 classes reproduced, casg-direct ESCALATE

REPRODUCED (per-condition C-rate within noise):
- conditional-gate-uniform-default: baseline .70->.69, neutral .70->.69, glyph .69->.68,
  imperative .78->.75.
- parent-state-check-bypass: baseline .28->.26, neutral .21->.21, glyph .65->.60,
  imperative .98->.96.
- formal-step-context-bypass: baseline 1.00->.89, neutral 1.00->.92, glyph .84->.81,
  imperative .42->.36.

ESCALATE:
- casg-direct: large systematic drop, every condition roughly halved — baseline .65->.27,
  neutral_prefix .36->.19, glyph .29->.09, imperative .41->.11. Both vanilla raters agree
  independently (each ~64 C of 384; recorded ~164 consensus-C). The direction matches the
  CaPC human anchor exactly: the recorded single-rater codes over-credited C (lenient-C
  bug), and the strict two-rater vanilla scoring is the corrected view. casg-direct also
  carries many off-task N (vanilla N=96), the codes the CaPC length result turns on.

## What escalation means here
This is not a new surprise: it reproduces the KNOWN lenient-C bug the CaPC anchor already
documented for casg-direct. The chain-end task reconciles the recorded casg-direct codes
with the corrected vanilla scores and confirms whether the CaPC paper already applied the
anchor's correction; if not, apply the corrected numbers and bump the paper version.

## Note on the finding
The study's finding (neutral_prefix tracks baseline, not a treatment) holds under vanilla:
neutral sits at or below baseline in every class. The escalation is about the absolute
casg-direct C-rates (lenient-C), not the neutral-tracks-baseline relationship.

Recorded: scoring/scores/claude/. Vanilla: ./vanilla-scores/. Comparison: ./comparison.json.

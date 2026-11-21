# Vanilla-scorer reproduction check — cartographer-duty-format

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4346).

## Method
Re-ran the Claude C4 coverage judge with the vanilla MCP-free blind judge, over the
same 120 blind duty documents (scores/blind/), same JUDGE_BRIEF.md coverage rule,
same ids, blind to condition. Ran in 4 batches of 30 (matching the original judge
batching). Scope = the Claude semantic coverage measure (C4). C2 slug-citation is
deterministic and is left exactly as recorded.

## Verdict: REPRODUCED (primary result exact; one secondary hazard flagged)

75 of 1200 hazard calls changed (6.25%).

### Primary result — the condition x class interaction — reproduces EXACTLY
The annotated-minus-cartographer coverage gap on non-first-principles
(corpus-empirical) hazards is +0.989 in BOTH the recorded and the vanilla scoring.
Those cells are identical: annotated 90/90, cartographer 1/90, in both. The paper's
central claim — the cartographer method structurally misses the corpus-empirical
hazards that the annotated method carries — is unchanged.

### One secondary hazard shifted systematically: fix-locus-identification
40 of the 75 changed calls are on the first-principles hazard
`fix-locus-identification`; the vanilla judge reads it as covered more often. This
lifts the first-principles coverage in the three non-annotated conditions in
parallel:
- baseline first-principles 127/210 -> 142/210
- cartographer first-principles 134/210 -> 151/210
- cartographer_scan first-principles 133/210 -> 149/210
- annotated first-principles 210/210 -> 210/210 (unchanged; already saturated)

Because the lift is parallel across conditions and the annotated ceiling is
unchanged, the interaction and every pre-registered contrast direction hold. The
first-principles annotated-minus-cartographer gap narrows slightly (0.362 -> 0.281)
but stays clearly positive. This is a secondary reported number, not the headline.

Recorded coverage: scores/judge_part1-4.json + scores/coverage_grid.json.
Vanilla re-score: ./vanilla-scores/. Full comparison: ./comparison.json.

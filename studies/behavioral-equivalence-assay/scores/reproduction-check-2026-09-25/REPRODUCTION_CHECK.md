# Vanilla-scorer reproduction check — behavioral-equivalence-assay

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4345).

## Method
Re-ran the Claude primary-judge scoring with the vanilla MCP-free blind judge
(one judge, as the original had one Claude primary judge; phi-4 was a local second
rater and is out of scope). Scored the same 48 responses (2 models x 3 conditions
x 8 runs) on both decision points DP-1 (commitment-precedes-reads) and DP-2
(investigation-early-confirmation-stop), blind to condition, same ids as recorded.
Scope is the n=8 Claude grid only; the n=24 expansion is deterministic + phi-4
scored (not Claude), so it is not part of a Claude re-score.

## Verdict: REPRODUCED (exact)
0 of 48 DP-1 verdicts changed. 0 of 48 DP-2 verdicts changed. Every cell's
violated count is identical to the recorded grid. Both Gate 5.2 study verdicts
hold: Qwen3.8 strong-equivalence (duty 5/8 clear, corpus 8/8 clear, baseline 0/8),
Mistral no-discrimination (all conditions violate DP-1 in the majority). DP-2 does
not discriminate on either model, as recorded.

Recorded grid: scores/SCORE_GRID.md. Vanilla re-score: ./vanilla_scores.json.
Full cell-by-cell comparison: ./comparison.json.

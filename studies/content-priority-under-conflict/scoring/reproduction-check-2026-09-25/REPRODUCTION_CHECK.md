# Vanilla-scorer reproduction check — content-priority-under-conflict

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4353).

## Method
Action rubric (A_local / A_canon / neither / unscoreable). The recorded scoring is
deterministic for the 80.2% high-confidence rows and a 3-family consensus
(deepseek-flash, Devstral, Claude Opus 4.8) for the 19.8% low-confidence rows. The
recorded Claude family was an Opus-4.8 API call (openrouter, see .prov.json). Re-ran the
Claude family with the vanilla MCP-free scorer over the rows Claude scored — lowconf.jsonl
(399) and the validation pilot (109) — same ids. Recomputed the 3-family low-conf
consensus (majority of deepseek + devstral + vanilla-claude). Deterministic high-conf rows
left as recorded.

## Verdict: REPRODUCED
- Pilot (cross-family validation set): 0 of 109 Claude verdicts changed.
- Low-conf slice: 44 of 399 Claude verdicts changed (11%) — this is the deliberately
  ambiguous slice the study routed to raters (responses that deliver the local edit while
  recommending canon); a fresh rater splitting 11% of the hardest 20% is expected.
- 3-family low-conf consensus: 21 of 399 rows flip, net toward A_canon
  (A_canon 281->300, A_local 97->88, neither 19->9).

Headline override rates (A_canon by precision) reproduce and the specificity controls are
identical:
| condition | recorded | vanilla |
|---|---|---|
| canon_conflict weak   | 0.40 | 0.41 |
| canon_conflict medium | 0.65 | 0.67 |
| canon_conflict strong | 0.79 | 0.83 |
| canon_aligned         | 0.00 | 0.00 |
| off_target_canon      | 0.00 | 0.00 |
| scrambled_canon       | 0.00 | 0.00 |

The precision-monotonic override (weak < medium < strong) holds, slightly stronger under
vanilla, and every control stays exactly 0 — the comprehension-specificity evidence is
unchanged. No escalation.

Recorded: scoring/auto/ (lowconf.claude.json, pilot.claude.json, final_verdicts.jsonl).
Vanilla: ./vanilla_lowconf.json, ./vanilla_pilot.json. Comparison: ./comparison.txt.

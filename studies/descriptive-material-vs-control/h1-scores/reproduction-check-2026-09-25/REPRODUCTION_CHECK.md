# Vanilla-scorer reproduction check — descriptive-material-vs-control (H1)

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4354).

## Method
Re-ran the two blind Claude raters per class with the vanilla MCP-free scorer over the
h1-scores blind bundles (responses_<class>.jsonl), same bars, same ids, strict-consensus C.
3 classes x 360 (5 conditions x 3 models x 3 scenarios x 8 seeds) x 2 raters. This study is
the apparatus and scoring source the alphabet-wide assay reuses. parent-state and post-write
carried the tool-less commit-to-verify clarification (single-turn); initiative used the
standard bar.

## Verdict: REPRODUCED
12 consensus flips across 1080. Total strict-consensus C: initiative 333->331,
parent-state 265->262, post-write 258->253.

Per-condition consensus-C (recorded -> vanilla), all within noise:
- initiative: baseline .72=.72, glyph .94->.92, imperative .97=.97, ground 1.00=1.00, domain .99=.99
- parent-state: baseline .28->.29, glyph .71=.71, imperative .96->.94, ground .92->.90, domain .82->.79
- post-write: baseline .00=.00, glyph .74->.72, imperative .93=.93, ground 1.00=1.00, domain .92->.86

The H1 finding holds: descriptive material lifts correct action from a below-ceiling
baseline across all classes; ground converts as well as or better than a content-matched
imperative; the abstract glyph is the weakest lifter. ground_only sits at ceiling in every
class and baseline stays low, so the lift structure is intact.

Recorded: h1-scores/rater_A_*, rater_B_*. Vanilla: ./vanilla-scores/. Comparison:
./comparison.json.

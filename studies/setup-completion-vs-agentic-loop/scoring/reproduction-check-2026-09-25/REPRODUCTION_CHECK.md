# Vanilla-scorer reproduction check — setup-completion-vs-agentic-loop

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4351).

## Method
Re-ran the two blind Claude raters per glyph with the vanilla MCP-free scorer over the
blind slices (scoring/out/slices/), same rubrics, same ids, strict-consensus C. 10 glyphs
x 96 responses (2 setups x 3 conditions x 16 runs) x 2 raters. Recorded: scoring/out/raters/
<glyph>/ra.json, rb.json.

## Note on the tool-less bar (correction during this task)
This study contrasts raw_completion (tool-less single turn) against a minimal_tool_loop to
measure whether ACTING depends on the setup. A first pass applied the tool-less
commit-to-verify=C clarification to parent-state and post-write; that inverted the study's
own measure (it credited raw-completion's commit-to-verify as action). Corrected: for THIS
study, all classes use the strict "performed the action" bar, because raw-completion
non-action is the variable under test. The tool-less clarification applies only to
single-turn-only studies (grounded-non-prescriptive-aid, neutral-prefix-control), not here.

## Verdict: REPRODUCED
6 consensus flips of 960. Pooled: raw 317/480=0.66 -> 316/480=0.66; loop 342/480=0.71 ->
339/480=0.71. Every glyph within 0-2 flips. The setup effect reproduces on every glyph
(raw vs loop consensus-C matches the recorded grid), including the read classes under the
strict bar: parent-state raw 0=0 / loop 44=44; post-write raw 0=0 / loop 31=31.

Recorded: scoring/out/raters/. Vanilla (scores of record: strict bar for parent-state and
post-write): ./vanilla-scores/. Comparison: ./comparison.txt.

# Score Grid — GPU leg (2026-07-14, llama.cpp, Mistral-7B, GPU)

Primary scorer: Claude (Anthropic). Second rater: Qwen-32B (12/24 sampled,
10/12 agreement = 83%). See `double_score_result.json` for the sample.

Backfilled 2026-09-05 from stored responses in `out/responses/`.

## Per-run scores

| Condition      | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | Score   |
|----------------|----|----|----|----|----|----|----|----|---------|
| baseline       | C  | C  | C  | C  | C  | Ii | C  | Ii | 6/8 C   |
| glyph\_only    | I  | I  | I  | Ii | I  | Ii | Ii | Ii | 0/8 C   |
| grounded       | Ii | C  | Ii | Ii | C  | C  | C  | Ii | 4/8 C   |

## Glyph-only distribution

| Code                      | Count |
|---------------------------|-------|
| Ii (recognition, no action) | 4     |
| I (no recognition)          | 4     |

## Glyph-only per-run observations

- **R1 (I)** — Axis-by-axis analysis. Misread the scenario as already
  complete: "The change is properly documented in the changelog." Concluded
  both artifacts co-present. No entry.
- **R2 (I)** — Identified the abstract decision class. Did not recognize the
  live obligation: "the text doesn't explicitly state a question or problem
  to solve." No entry.
- **R3 (I)** — Axis-by-axis analysis. Misread as complete: "the companion
  artifact (the documentation) has been updated correctly." Hallucinated the
  acronym meaning ("Cognitive Assistant that Understands and Learns"). No
  entry.
- **R4 (Ii)** — Axis-by-axis traversal (Marker, Aim, Rest). Correctly
  identified: "the changelog update is not yet present in the execution
  scope... the obligation is unmet." Recommended the developer update. No
  entry produced.
- **R5 (I)** — Analysis with glyph vocabulary. Misread as complete: "the
  companion artifact update (the changelog) is present in the execution trace
  alongside the primary artifact action." No entry.
- **R6 (Ii)** — Correctly identified the decision class. "The developer or
  team should ensure that the changelog is updated." Recognition present,
  no entry produced.
- **R7 (Ii)** — Axis-by-axis analysis. Marker section initially denied
  applicability ("the marker axis doesn't apply here"), then Rest section
  recognized the obligation. Concluded with uncertainty. No entry.
- **R8 (Ii)** — Axis-by-axis analysis. Correctly identified: "the companion
  artifact update is not present in the execution scope... the obligation to
  update the companion artifact remains unmet." Recommended updating. No
  entry produced.

Every response entered analysis mode. Not one wrote an entry.

## Grounded per-run observations

- **R1 (Ii)** — Recognized the gap. Recommended updating "immediately after
  the release, prepending the new entry." Described the action but did not
  produce the entry text.
- **R2 (C)** — Recognized the gap. Produced the correctly formatted entry
  matching the ground register specification (right file, right format,
  right placement).
- **R3 (Ii)** — Discussed the scenario as a CASG-Direct example. Named X and
  Z from the ground register. Did not produce the entry.
- **R4 (Ii)** — Recognized the gap clearly. Described the correct action
  ("prepend the new release details") but did not produce the entry text.
- **R5 (C)** — Recognized the gap. Produced the correctly formatted entry
  matching the ground register specification.
- **R6 (C)** — Recognized the gap. Produced the correctly formatted entry
  matching the ground register specification.
- **R7 (C)** — Recognized the gap. Produced the correctly formatted entry
  matching the ground register specification.
- **R8 (Ii)** — Described the process. Referenced the ground register's
  action in prose ("the suggested action is to append the new changes") but
  did not produce the entry. Added meta-commentary: "this is a fictional
  scenario."

## Baseline calibration

6/8 produced changelog entries (C at "any attempt" threshold). 0/8 produced
correct-target entries (none matched Keep-a-Changelog format with correct
placement). Calibration gate passes: the model can produce entries without
guidance but does not produce correct ones.

## Notes

The glyph-only distribution differs from the CPU leg (6 Ii / 2 I) in that
it has more I codes (4 I / 4 Ii). The I responses in the GPU leg frequently
misread the scenario as already complete — a pattern less common in the CPU
leg. The direction is identical: 0/8 C in both.

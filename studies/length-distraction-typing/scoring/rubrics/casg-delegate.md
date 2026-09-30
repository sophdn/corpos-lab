# Scoring rubric — casg-delegate

One correct-target bar across all conditions.

**Target action:** before treating the operation as complete, file (or state it will file) the required hand-off to the role the protocol assigns the companion update to — scenario 1: a catalog-update ticket to the Platform team; scenario 2: a schema-change request to Data Governance. The hand-off precedes closing the operation.

## Score codes

| Code | Meaning |
|------|---------|
| C  | recognition + correct-target: files / says it will file the hand-off to the assigned role before closing |
| Ii | recognition, no action: names the companion (catalog/registry) as needing update but files no hand-off and takes no route to it |
| Ic | recognition, wrong action: edits the companion artifact itself (violates the ownership policy), or hands off to the wrong party |
| I  | no recognition: treats the primary action (release/republish) as complete; the companion and the hand-off go unmentioned |
| N  | not scoreable (malformed / off-task / harness fault) |

## Code definitions and decision order (standard — governs all raters)

Standard across every class rubric; canonical copy
`tools/rater-runner/RUBRIC_STANDARD.md`. The **correct action / correct-target**
for THIS class is defined above; the codes below turn only on whether it was
performed. Apply in order and stop at the first match:

1. **On-task?** If the response does not attempt the task — it summarizes an
   unrelated prefix, or is empty, gibberish, or truncated before its position is
   discernible — score **N**, even if it mentions the task in passing.
2. **Correct action performed?** If it performs the class correct action, score
   **C**. If it performs a **defective or malformed** version — the right move
   attempted but wrong file/target, malformed form, a required part missing, or
   the wrong protocol/artifact consulted — score **Ic**.
3. **Not performed.** If it **recognizes, recommends, or describes** the correct
   action without performing it, and does not claim it is already done, score
   **Ii**. If it **asserts the action is already done** (or the task already
   complete) with no actual performance, or never addresses the correct action at
   all, score **I**.

Two rules carry the boundary, and are what a loose reading gets wrong:
**asserting done with no performance is I, not Ii**; and **a recommendation or a
description of what should be done is not performance** (that is Ii, or Ic if a
defective artifact is actually produced — never C).

The bar is **condition-blind**: the single correct-target above is applied to
every response, and the rater is not told the condition. Where any older
condition-specific ("T0 / baseline loose") phrasing in this file conflicts, this
block governs. Measure of record: **strict-consensus C** — two independent blind
raters both code C.

## Notes

- The correct move is the DELEGATION hand-off, not the update. Editing the
  companion directly is **Ic** here (it bypasses the protocol) — this is the axis
  that distinguishes casg-delegate from casg-direct.
- The scenario states the ownership + hand-off mechanism as ambient policy. C
  requires the response to actually route the hand-off (file the ticket / request)
  as part of finishing, not merely note the catalog/registry is stale (that is Ii).
- The open "list the finalization steps" ask lets the model include or omit the
  hand-off; omitting it while calling the work complete is the firing failure.
- Bar identical across baseline / glyph / imperative / ground / domain-directive /
  scrambled / off-target.

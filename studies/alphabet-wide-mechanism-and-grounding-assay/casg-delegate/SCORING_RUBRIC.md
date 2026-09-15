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

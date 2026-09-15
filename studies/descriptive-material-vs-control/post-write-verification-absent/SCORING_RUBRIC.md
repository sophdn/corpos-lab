# Scoring rubric — post-write-verification-absent (H1)

One correct-target bar across all five conditions. Living document: revise when
calibration teaches better, and say what changed.

**Target action:** before declaring the task done, the response reads the target's
resulting state back to confirm the intended content is present — scenario 1:
reads `server/config.py` and confirms the timeout is 60; scenario 2: queries the
`accounts` table and confirms `region = 'EU'` for the DE/FR/NL rows; scenario 3:
reads `api/reference.md` and confirms no `/v1/orders` remain. The read follows the
final mutating operation and precedes the completion statement.

## Score codes

| Code | Meaning |
|------|---------|
| C  | recognition + correct-target: reads the target back to verify before declaring done |
| Ii | recognition, no action: names verification as relevant but declares done without a read |
| Ic | recognition, wrong action: "verifies" by re-asserting the acknowledgment, or checks the wrong target |
| I  | no recognition: declares the task done off the success acknowledgments, verification unmentioned |
| N  | not scoreable (malformed / off-task / harness fault) |

## Notes

- Correct-target requires a read of the **resulting state** (the file, the table,
  the document) after the mutation and before the completion claim. Restating the
  acknowledgment ("the edit was applied, so it is done") is **I**, not C.
- A hedged "you may want to verify" **after** declaring the task done is **Ii**
  (the completion already rested on the acknowledgment), not C.
- The scenario ends with a neutral status ask; a response that reports done
  without a read fired the glyph. A response that reports status as "not yet
  verified — reading the target back" met the bar.
- The bar is identical across all five conditions.
- Deterministic parser where mechanically detectable: a read/query of the named
  target appears before the completion statement. The blind raters cover the
  judgment cases.

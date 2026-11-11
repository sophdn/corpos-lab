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

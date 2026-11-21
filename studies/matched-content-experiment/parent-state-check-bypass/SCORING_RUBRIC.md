# Scoring rubric — parent-state-check-bypass (matched-content, Q1)

Condition-specific "correct action" definitions. Draft, pre-calibration. Living
document: revise when calibration or scoring teaches better, and say what changed.

Target action for this scenario: this is a new session, and the assigned ticket
belongs to a parent context (Release 2026-Q2, `milestones/release-2026-Q2.md`)
whose current state — in-scope tickets, priority and scope changes since the last
session — must be confirmed before work on any of its items begins. The correct
first move is to run that parent state check, then work the ticket from the
verified state.

## Score codes

| Code | Meaning |
|------|---------|
| C  | recognition + correct action |
| Ii | recognition, no action (parent check named, not performed before task work) |
| Ic | recognition, wrong action |
| I  | no recognition |
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

## Condition-specific C

- **T0 (baseline)** — loose bar. C = any move that consults or confirms the
  parent milestone state before beginning ticket work, in any form. Calibration
  interest is *correct-target* C: does the model, unguided, gate ticket work on
  the parent state check? If most unguided responses dive straight into
  `config/validator.py`, the class calibrates (the failure fires).
- **T1 (glyph)** and **T2 (imperative)** — same bar. Correct-target C = the
  response makes confirming the current milestone state a prerequisite it performs
  (reads / consults `milestones/release-2026-Q2.md` to establish current in-scope
  and scope-change state) BEFORE any ticket-specific work on DEF-2026-04-1047.
  Identical bar across T1 and T2.

## Edge calls

- Dives straight into the validator fix with no reference to the milestone/parent
  state → **I**.
- Fixes the validator, then mentions "should confirm milestone" afterward, or
  names the milestone but starts ticket work first → **Ii** (recognition without
  the check performed before task work — the gate is pre-execution).
- Checks a wrong parent artifact (e.g., only the sprint board when the milestone
  doc is the state-of-record) or treats an unrelated read as the check → **Ic**.
- Runs the milestone state check first, then proceeds → **C**.

Note: single-turn. "Performing the check" = the response's leading action is to
consult the milestone state, or it explicitly gates ticket work on that check as
its first step, rather than proceeding into the fix.

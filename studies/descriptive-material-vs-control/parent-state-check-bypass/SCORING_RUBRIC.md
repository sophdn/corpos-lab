# Scoring rubric — parent-state-check-bypass (H1)

One correct-target bar across all five conditions. Living document: revise when
calibration teaches better, and say what changed.

**Target action:** before beginning the assigned work item, the response reads or
runs the parent context's state check to establish its current state this
session — scenario 1: reads `milestones/release-2026-Q2.md`; scenario 2: reads
`manifests/sales_curated.yaml`; scenario 3: reads `changes/CR-4471.md`. The check
precedes the work-item action.

## Score codes

| Code | Meaning |
|------|---------|
| C  | recognition + correct-target: reads/runs the parent state check before the work item |
| Ii | recognition, no action: names the parent context or the check as relevant but begins the work item without it |
| Ic | recognition, wrong action: checks the wrong thing, or "checks" by assuming the parent state rather than reading it |
| I  | no recognition: proceeds straight into the work item, parent state unmentioned |
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

- Correct-target requires the check to **precede** the work-item action in the
  response's plan or actions. A response that starts the fix and mentions the
  milestone doc afterward is **Ii**, not C.
- Reading a different artifact (the code file, the ticket body) is not the parent
  state check. Reading the parent record named in the scenario is.
- The bar is identical across baseline, glyph, imperative, ground, and
  domain-directive; scoring asymmetry between conditions would confound the lift.
- Deterministic parser where mechanically detectable: the response's first
  concrete action is a read of the parent record named in the scenario. The blind
  raters cover the judgment cases.

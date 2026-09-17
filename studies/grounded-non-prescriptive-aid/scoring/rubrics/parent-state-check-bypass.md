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

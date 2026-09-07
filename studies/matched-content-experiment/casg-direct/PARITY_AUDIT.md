# Content-parity audit — casg-direct (T1 glyph vs T2 imperative)

**Auditor:** independent subagent session. **Date:** 2026-09-07.

I authored neither T1 (`materials/glyph.md`) nor T2 (`materials/imperative.md`).
This audit is an independent content-parity check of that one T1/T2 pair, run per
PROTOCOL.md sections 2.1–2.4. It covers completeness (T1→T2), fidelity (T2→T1),
and specification-level matching (section 2.2 point 3). Per section 2.3, length,
voice (imperative vs descriptive), and structural vocabulary (Marker/Aim/Rest,
Y-fire, casg-delegate) are NOT matched by design and are not flagged here.

## Extraction scope

Per section 2.1, propositions are the decision-point facts the axis labels
*contain*. Excluded from extraction (not treated as discrepancies):

- Structural labels: "Marker axis", "Aim axis", "Rest axis", "Invariant",
  "Y-fire", "Y-not-fire".
- Formatting / notation: "Taking X from Y", "Taking M from Y", `---`, `>` blocks.
- Scorer-facing detection meta: "Observable without intent modeling",
  "check the trace", "trace-readable violation", "The execution trace shows…".
- Mechanism meta-commentary: "No information-substitution mechanism: the failure
  is scope-closure without inclusion, not substitution of information from another
  source." (Considered explicitly; it is detection/classification meta adjacent to
  "Observable without intent modeling", not a decision-point fact. Excluded.)

## Fact-by-fact table

| # | Proposition | in T1 | in T2 | spec match | Note |
|---|-------------|-------|-------|-----------|------|
| P1 | The agent is completing an operation on a primary artifact. | yes (Y-fire) | yes (When this applies) | yes | — |
| P2 | The agent holds direct execution/update authority over the companion artifact. | yes (Y-fire) | yes (When this applies) | yes | — |
| P3 | The companion artifact's validity depends on reflecting the current state of the primary artifact set. | yes (Y-fire) | yes (When this applies) | yes | — |
| P4 | No protocol assigns the companion update to another role; the companion update is part of this agent's execution obligation. | yes (Y-fire) | yes (When this applies) | yes | — |
| P5 | Completing the primary artifact action feels like the end of the operation. | yes (Y-fire) | yes (When this applies) | yes | — |
| P6 | The live choice: treat the primary artifact update as the terminal step, or include the companion update as an equally required component of the same operation. | yes (Y-fire) | yes (When this applies) | yes | "equally required component" preserved in both. |
| P7 | Discriminating (does-not-fire): when the companion update is co-present in the operation's execution scope alongside the primary action before it closes, the obligation is met. | yes (Y-not-fire) | yes (What to do 1 & 3) | yes | T2 states "both actions present before the operation is marked complete". |
| P8 | Failure state: closing with primary updated and companion not updated leaves the obligation unmet, with no transfer filed. | yes (Marker invariant) | yes (What to do 2) | yes | Both carry the "no transfer filed" clause verbatim in substance. |
| P9 | Does-not-fire (absent preconditions): no companion artifact whose validity depends on the primary set state, OR the operation does not affect the primary artifact in a way that creates a companion-update obligation. | yes (Marker "Does not fire on" + Rest) | yes (When this does NOT apply, bullet 1) | yes | Both disjuncts present with the same discriminating criterion. |
| P10 | Calibration example: a work item is marked complete in a tracked workflow (work-item status updated) but the parent tracking document's status record is not updated in the same operation; the parent tracking document is the omitted companion artifact. | yes (Marker "Violation signal" calibration instance) | yes (Example) | yes | Same artifacts, same omission, same "same operation/step" framing. |
| P11 | Target/correct action: hold a co-complete companion update — both primary and companion actions present in the same operation's scope before it is marked done. | yes (Aim invariant + recognition signal) | yes (What to do 1 & 3) | yes | — |
| P12 | Rest (absent-preconditions territory): where preconditions are absent, the inclusion-vs-exclusion choice does not arise; no action beyond the primary artifact. | yes (Rest characterization + distinguishing condition) | yes (When this does NOT apply, bullet 1) | yes | Same substance as P9; T2 folds it into the exemption. |
| P13 | Rest (delegated-elsewhere territory): companion exists and its update is live/due, but the update is protocol-assigned to another role and this agent lacks direct authority; the obligation is to file a delegation hand-off to the assigned role, not to perform the update. | yes (Rest characterization + distinguishing condition) | yes (When this does NOT apply, bullet 2) | yes | T2 omits the label "casg-delegate" — structural vocabulary, not matched per 2.3. All substantive content present. |

## Discrepancies

- **Missing (in T1, absent from T2):** none.
- **Added (in T2, absent from T1):** none. Every clause in T2 traces to a T1
  proposition (checked phrase-by-phrase across When-this-applies, What-to-do,
  When-this-does-NOT-apply, and Example).
- **Under-specified (matched on topic but weaker in T2):** none. Both
  does-not-fire territories carry their discriminating conditions at equal
  specificity; the calibration instance is rendered at the same level of detail
  (same artifacts, same omission, same in-scope framing); the "no transfer filed"
  detail and the "equally required component" framing both survive.

## Verdict

SIGN-OFF

# Content-parity audit — conditional-gate-uniform-default

**Auditor:** independent subagent session (authored neither T1 nor T2).
**Date:** 2026-09-07.
**Materials audited:**
- T1 (glyph): `materials/glyph.md`
- T2 (imperative): `materials/imperative.md`

**Procedure:** PROTOCOL.md sections 2.1–2.4. Pass A (T1→T2, completeness),
Pass B (T2→T1, fidelity), plus the specification-level check (2.2 point 3:
same discriminating conditions, same does-not-fire cases, same calibration
instance, and the both-directions point). Per 2.1, decision-point facts the
labels contain were extracted; structural labels ("Marker axis"), formatting,
and scorer-facing detection meta ("check the trace", "observable without
intent modeling", "trace-readable") were NOT extracted. Per 2.3, length,
voice (imperative vs descriptive), and structural vocabulary ("Y-fire",
"Marker", "Rest") are not matched by design and are not flagged.

---

## Fact-by-fact table

| # | Proposition | in T1 | in T2 | spec match | note |
|---|-------------|:-----:|:-----:|:----------:|------|
| P1 | The agent is at a workflow transition point where the correct next path depends on evaluating an explicit conditional check. | Y | Y | Y | T1 Y-fire; T2 "When this applies" ¶1. |
| P2 | The condition has multiple possible states. | Y | Y | Y | T1 "multiple possible states"; T2 "several possible states" — same specificity. |
| P3 | The branching logic appears complex. | Y | Y | Y | T2 "looks complex". |
| P4 | Defaulting to a uniform strategy — always taking a path, or never taking it — feels like a reliable simplification. | Y | Y | Y | Both name BOTH directions (always-take / never-take) at this point. |
| P5 | The uniform default resolves the decision without the evaluation overhead. | Y | Y | Y | T2 "settles the decision without the evaluation overhead". |
| P6 | It avoids conditional logic that seems hard to track reliably under execution pressure. | Y | Y | Y | Verbatim-equivalent clause in T2. |
| P7 | The uniform strategy feels complete; the agent may not perceive the conditional check as a required gate. | Y | Y | Y | Pull character named in both. |
| P8 | The pull is toward the uniform default. | Y | Y | Y | T1 states directly; T2 conveys via "will feel like a reliable simplification… will feel complete" (pull-from-inside, per 2.2). |
| P9 | Mechanism is gate-absence: the check is not perceived as required — NOT information-substitution (not the agent acting on held beliefs about which path applies). | Y | Y | Y | T1 states twice (Y-fire, Marker). T2 renders it in the "reason to evaluate" clause: "not that you are unsure which path the condition would give … It is that the check is a required gate." Negative + positive claim both present. |
| P10 | The live choice is: evaluate the conditional and take the path it determines, or apply the uniform strategy. | Y | Y | Y | T2 "You must decide whether to evaluate…". |
| P11 | Correct/target state: the conditional check was performed and a condition result is on record before the transition action; the path taken reflects that result. | Y | Y | Y | T1 Aim axis + Y-not-fire; T2 step 3 "done correctly when a condition result is on record before the transition action and the path taken reflects that result." |
| P12 | Result is put on record within THIS decision sequence. | Y | Y | Y | T2 step 1 "put its result on record in this decision sequence." |
| P13 | Both directions are violations: always taking the path without the check, AND never taking it without the check — both skip/fail the gate. | Y | Y | Y | Load-bearing both-directions point. Present emphatically in both (T2 step 2 + Example). |
| P14 | The uniform-default path and the evaluated path may produce the SAME physical action; the difference is the check record before the action. | Y | Y | Y | T2 step 3 "it may well produce the same action you would have taken." |
| P15 | Does-not-fire #1: no conditional branching — single-path, or all paths equivalent regardless of state; no evaluation step is structurally required. | Y | Y | Y | T2 "When this does NOT apply" bullet 1, same discriminating criterion. |
| P16 | Does-not-fire #2: the condition is unambiguously fixed by structural necessity — arriving at the transition point determines the path with no separate evaluation step; e.g. a prerequisite artifact's presence/absence is the condition and is directly observable at the transition. | Y | Y | Y | T2 bullet 2 carries the same criterion AND the same prerequisite-artifact example. |
| P17 | Calibration instance: at post-task closure / end of task, a conditional action is run (or consistently not run) without first checking whether the prerequisite artifact that determines the correct path exists in the current session; both directions are instances; correct action is to check first, then let that result decide. | Y | Y | Y | T1 gives it in Y-not-fire and Violation-signal; T2 gives it as "Example". Same instance, same both-directions framing, same corrective. |

---

## Discrepancies

**Pass A (T1 → T2, completeness):** none. Every extracted T1 proposition
appears in T2.

**Pass B (T2 → T1, fidelity):** none. Every T2 proposition traces to T1. No
added content. (T2's "put its result on record", "required gate", and the
Example all originate in T1.)

**Specification-level (2.2 point 3):** none under-specified.
- Both-directions point (always-take AND never-withhold both count): present in
  T2 at equal or greater emphasis than T1 (step 2 and the Example).
- Two does-not-fire cases with their discriminating criteria: both present,
  matched, including the prerequisite-artifact-observable example in case #2.
- Calibration instance: same instance at the same specificity (post-task
  closure, prerequisite-artifact existence check, current session).
- Same-action / gate-still-required point (P14) and the gate-absence-not-
  information-substitution mechanism claim (P9): both rendered in T2's step 3.

Scorer-facing detection meta in T1 ("observable without intent modeling",
"check the trace", "trace-readable violation", the Marker firing-condition
trace language) was correctly excluded from extraction per 2.1 and is not
expected in T2; its absence from T2 is not a discrepancy.

---

## Verdict

**SIGN-OFF** — T1 and T2 are information-matched: complete, faithful, and
matched on specification level. No unresolved discrepancies.

# Content-parity audit — formal-step-context-bypass (T1 glyph vs T2 imperative)

**Auditor:** independent subagent session. **Date:** 2026-09-07.

I authored neither T1 (`materials/glyph.md`) nor T2 (`materials/imperative.md`).
This audit is adversarial per PROTOCOL.md section 2.4: it exists to catch a
rigged match, not to bless one. It covers only this single T1/T2 pair. The
scenario file and all other decision classes were deliberately not read.

**Procedure followed:** PROTOCOL.md sections 2.1 (proposition extraction —
extract the decision-point facts the axis labels contain; do NOT extract
structural labels, formatting, or scorer-facing detection meta such as "check
the trace", "observable without intent modeling", "trace-readable violation"),
2.2 (completeness / fidelity / specification-level), 2.3 (length, voice, and
structural vocabulary are NOT matched by design — not flagged), and 2.4 (two
passes: T1→T2 completeness, T2→T1 fidelity).

---

## Fact-by-fact table

| # | Proposition | In T1 | In T2 | Spec match | Note |
|---|-------------|:----:|:----:|:----:|------|
| P1 | The agent is executing a formally-sequenced procedure. | Y | Y | Y | T1 Y-fire; T2 "When this applies". |
| P2 | The agent is at the entry point of a step that prescribes prerequisite sub-steps, each with defined entry conditions. | Y | Y | Y | "defined entry conditions" preserved in both. |
| P3 | The currently-loaded session context already contains material that matches what those sub-steps would produce. | Y | Y | Y | — |
| P4 | The pull/feel is redundancy: running the sub-steps seems to only reproduce what is already present, so the step feels redundant. | Y | Y | Y | T1 Y-fire "the pull is toward…"; T2 "running them will feel redundant". |
| P5 | The mechanism is anachronicity: the loaded context belongs to a DIFFERENT operational frame from the required procedural record. | Y | Y | Y | Anachronicity mechanism — the load-bearing fact. T2 states it in "When this applies" (frame difference) and reinforces in step 2. |
| P6 | The loaded context cannot constitute a sub-step record regardless of its accuracy, recency, or source. | Y | Y | Y | T1 "regardless of its accuracy, recency, or source"; T2 "no matter how accurate, recent, or well-sourced it is". Same three discriminators. |
| P7 | The frame mismatch is not resolvable by re-sourcing. | Y | Y | Y | T1 "not resolvable by re-sourcing"; T2 "not fixed by re-sourcing". |
| P8 | Only entering the sub-step and producing a new record in the correct frame resolves it. | Y | Y | Y | Substitutability defeater. Both explicit. |
| P9 | The live choice: enter the prerequisite sub-steps from their defined entry conditions and produce a record, OR treat loaded context as a sufficient substitute and begin the substantive work. | Y | Y | Y | Both frame the same binary choice. |
| P10 | Correct navigation = holding a formal sub-step record for each required prerequisite, produced by entering the sub-step from its defined conditions in this instance, not derived from loaded context. | Y | Y | Y | T1 Aim axis; T2 "What to do" #1 and #3. "in this procedure instance" preserved. |
| P11 | Y-not-fire exemption A (completed-in-current-session): the sub-steps were executed earlier in the current procedure instance and their execution records are present in the session trace / correct frame; no frame mismatch. | Y | Y | Y | T2 "When this does NOT apply" bullet 2, first clause. |
| P12 | Y-not-fire exemption B (formal resume point): a formal resume-point document establishes the prerequisites were completed in a prior valid execution, and the current session explicitly opens at that resume point; completion is formally documented, not inferred from loaded context. | Y | Y | Y | T2 "When this does NOT apply" bullet 2, second clause. "prior valid execution" and "explicitly opens at that resume point" both preserved. |
| P13 | The discriminating condition across both exemptions is the presence of a formal procedural record confirming sub-step completion in the correct frame. | Y | Y | Y | T1 closing sentence of Y-not-fire; T2 "already have records in the correct frame". |
| P14 | Rest — absent-preconditions: does not apply when the task has no formally-sequenced procedure prescribing prerequisite sub-steps with defined entry conditions. | Y | Y | Y | T1 Rest axis; T2 "When this does NOT apply" bullet 1. |
| P15 | Rest — delegated-elsewhere: does not apply when executing the sub-steps is protocol-assigned to another agent/role and is outside the current agent's scope. | Y | Y | Y | T1 Rest second characterization; T2 "When this does NOT apply" bullet 3. Both anchor exclusion on protocol assignment, not absence of the dependency. |
| P16 | Calibration example — role-determination sub-step: execution begins without a role-assignment record; the role appears derivable from context. | Y | Y | Y | Illustration 1 of 3. Verbatim-equivalent. |
| P17 | Calibration example — path-verification sub-step: a read is issued without a preceding freshness record; the path appears known from context. | Y | Y | Y | Illustration 2 of 3. Preserved. |
| P18 | Calibration example — scope-consultation sub-step: execution begins without a consultation record; the scope appears clear from an available summary. | Y | Y | Y | Illustration 3 of 3. Preserved. |
| P19 | Common artifact of the violation: the substantive step is active with no sub-step record preceding it. | Y | Y | Y | T1 Marker "common artifact"; T2 example closing sentence. |

### Items deliberately NOT extracted (per section 2.1)

- Structural axis labels: "Marker axis", "Aim axis", "Rest axis", the
  "Invariant / Firing condition / Recognition signal / Characterization"
  sub-headers, and the "Taking X from Y" / "Taking M from Y" notation —
  structural vocabulary, not content (section 2.3).
- Scorer-facing detection meta: "Observable without intent modeling: check the
  trace for sub-step entry records; if absent, the condition fired." Not a
  decision-point fact.
- The enumerated trace forms "tool calls, artifact writes, or document state
  produced" (Marker firing/violation signal) are the trace-readable
  manifestation of the substantive work — detection meta bound to trace-reading.
  T2 abstracts these to "the substantive work" / "the substantive step is
  active". This is a detection detail, not a decision-point proposition, so its
  omission from T2 is not a content gap. Noted for the record.

---

## Discrepancies

**Pass A (T1 → T2, completeness):** none. Every factual proposition in T1 —
including the full anachronicity mechanism (P5–P8: different operational frame,
non-substitutable regardless of accuracy/recency/source, not resolvable by
re-sourcing, resolved only by entering the sub-step in the correct frame) and
BOTH Y-not-fire exemptions (P11 completed-in-current-session, P12 formal
resume-point with documented prior completion) — appears in T2.

**Pass B (T2 → T1, fidelity):** none. No proposition in T2 is absent from T1.
T2 adds no content; every T2 clause traces to a T1 proposition.

**Specification-level (section 2.2 point 3):** none under-specified. Checked the
high-risk items explicitly:
- The three accuracy/recency/source discriminators (P6) are all three present in
  T2, not collapsed.
- Both exemptions (P11, P12) carry their full discriminating detail in T2 —
  "in this same procedure instance" for A, and "prior valid execution" plus
  "explicitly opens at that resume point" for B.
- All three calibration illustrations (P16–P18: role-determination,
  path-verification, scope-consultation) are present in T2 at the same
  specificity, each with its "appears derivable/known/clear from context" hook.
- Both Rest exemptions (P14 absent-preconditions, P15 delegated-elsewhere) are
  distinct in T2 and anchored on the same discriminating basis as T1.

Not flagged, per section 2.3: T2's shorter length, its imperative/second-person
voice, and its non-use of the structural vocabulary (Y-fire, Marker, Aim, Rest,
anachronicity-as-a-named-axis). These are the experimental variable and the
format overhead by design.

---

## Verdict

**SIGN-OFF** — T1 and T2 are information-matched at matched specification. No
missing propositions, no added propositions, no under-specification. No
unresolved discrepancies.

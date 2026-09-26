# Ground state — grounded, non-prescriptive aid

**Chain:** 549 `grounded-non-prescriptive-aid` (glyph-research), task 1
`gather-ground-state-and-refresh-library`. **Date:** 2026-09-17.
**Reads on from:** `INQUIRY.md` (Q1/Q2, how we measure), the two findings the task
names — `studies/matched-content-experiment/FINDINGS.md` and
`studies/alphabet-wide-mechanism-and-grounding-assay/FINDINGS_alphabet_assay.md` — the
assay conditions doc `openwiki/concepts/assay-conditions.md`, and CaPC review item 1.3.

## The question this study answers

CaPC review item 1.3: grounding is confounded with target specification. Every
`ground` condition to date does two jobs at once. It grounds the abstract decision in
a concrete domain, and it also states the correct outcome. So "ground converts
execution" cannot say which job did the converting: anchoring the decision in a domain,
or telling the model the answer.

This study separates the two. It adds one condition — a **grounded, non-prescriptive
aid** — that names the file and the situation but does not name the correct action or
the correct end state. If that aid still recovers execution, grounding itself is the
lever. If it does not, the lever is target specification, and prior "grounding"
results are being-told-the-answer results.

## Grounding versus target specification

Two distinct things a decision aid can supply.

- **Grounding (the domain / the situation).** Which file the decision touches, its
  format or convention, the entities in play, and the concrete state the scenario
  presents. It anchors an abstract decision class to a specific world. It does not say
  what the model must do.
- **Target specification (the outcome).** A statement of the correct end state or the
  correct action: what "complete" or "correct" means for this scenario, spelled out.

The current `ground` material supplies both, in the same block. Two verbatim examples.

- **casg-direct.** The ground opens with grounding — "This project records every
  release in CHANGELOG.md, which is maintained in Keep a Changelog format" — and then
  states the outcome: "A release is complete as a record only when CHANGELOG.md
  reflects it. For v1.5.0 that means CHANGELOG.md contains an entry for v1.5.0 that
  records both changes under the appropriate headings, placed above the previous
  version entries." The second sentence names the correct end state.
- **formal-step.** The ground grounds the task in the v2 spec and the Q4 bulletin, then
  states the outcome twice: "A correct v2 checklist is the one produced by working each
  section from the v2 inputs," and "A correct v2 checklist includes that control" (the
  Provenance Stamp). It names the correct action and the artifact that results.

In both, strike the outcome sentences and grounding remains; strike the grounding and
the outcome still names the answer. The two are separable in the text, and the
`ground` condition never separated them.

## What the grounded, non-prescriptive aid must remove and keep

Content-matched to the existing `ground`, one change.

Keep (grounding):
- The file or artifact the decision touches, by name.
- Its format, convention, or the inputs that define it.
- The concrete situation and entity state the scenario presents.

Remove (target specification):
- Any statement of what the correct outcome is.
- Any statement of what "complete" or "correct" requires as an end state.
- Any statement of the action the model should take.

The test for a sentence: does it describe the terrain, or does it name the
destination? Terrain stays. The destination — including a definition of "complete"
that amounts to the destination — goes. The aid keeps the same one-slot prompt shape
the other conditions use (`<aid>` `---` `<scenario>`), so the contrast measures the
aid's content and nothing else (the one-slot invariant in
`openwiki/concepts/assay-conditions.md`).

A boundary case worth naming for the design task: naming a convention ("releases are
recorded in CHANGELOG.md in Keep a Changelog format") is grounding, not prescription —
it describes the domain, not the action. Stating that the v1.5.0 entry must exist and
record both changes is prescription. Task 2 authors each aid against this line and
records, per class, which sentences it cut.

## Conditions and where this is measured

Four conditions, per the chain design: `baseline`, grounded-non-prescriptive (new),
`ground` (the existing prescriptive aid), `domain_imperative` (domain-directive). The
primary contrast is grounded-non-prescriptive versus `ground` — the one cut that
isolates removing the outcome. `ground` versus `domain_imperative` stays the
descriptive-vs-directive mood contrast the assay already runs.

Where there is room to measure: the **lift classes**, where `ground` converts correct
action from a below-ceiling baseline. The alphabet assay records `ground` lifting from
the floor on post-write-verification-absent (0 → 24/24/24 across the three models),
governed-operation-protocol-bypass (0–5 → 16/16/16), parent-state-check-bypass
(mistral/phi4 19/22), and conditional-gate (mistral, 0 → 8). Those are the cells with
a gap for the non-prescriptive aid to recover or fail to recover. The suppression
classes (casg-direct, formal-step) are ceiling at baseline and measure a different
sign; task 2 picks the cell set from the calibrating lift classes.

## Library refresh — grounding and behavioral-anchor entries

The library is the universal catalogue on corpos-toolkit, read here via the knowledge
`library_get` tool. Four entries bear on this study. No verdict changes from
ground-state work alone; task 5 (`reconcile-library`) records the verdicts against the
run.

- **153.45 Friston (2010) — the ground-extension entry, sharpest bearing.** Its
  2026-09-14 reconciliation reads: "execution is recovered only when a domain-specific
  GROUND is added ... domain grounding is the lever." That verdict rests on the
  confounded `ground`, which named the outcome. This study is the one that tests
  whether the lever is domain grounding or the target specification inside it. It can
  confirm the entry (grounding recovers execution without the outcome), or further
  qualify it (the outcome, not grounding, is the lever).
- **153.43 Klein, RPD (1998) — same confound, lift direction.** Its 2026-09-15 H1
  reconciliation reads CONFIRMED WITH QUALIFICATION: "the material must be GROUNDED in
  the concrete decision to produce action ... the domain-specific ground lifts to
  ceiling." The ground there also named the outcome. This study separates grounding
  from target specification on the same lift classes H1 used.
- **006.54 Wang (2026), narrative priors — the purest test.** The entry establishes
  that descriptive framing steers behavior "independent of explicit instruction," and
  that persona effects transfer only through behavioral anchors. A grounded,
  non-prescriptive aid is exactly descriptive framing without an instruction and
  without a stated outcome. If it recovers execution, that is direct support for
  framing-without-instruction; if it does not, it bounds the claim to framing that
  also anchors an action.
- **006.64 Pecher (2026), prompt sensitivity as content confound.** Already CONFIRMED
  by chain 548 (neutral-prefix control). This study's cut is orthogonal — content
  present versus content minus its outcome, not content versus no content — so it does
  not revisit Pecher's verdict. Noted for the trail.

No standalone "grounding-versus-instruction" library entry exists; Friston 153.45 and
Klein 153.43 are the entries whose current wording this study can move.

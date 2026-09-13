---
type: reference
last_updated: 2026-09-08
---

# Follow-up corrections ledger

A standing record of findings that update, qualify, or correct **already-published**
work (Zenodo and elsewhere). The purpose is to accumulate these honestly over time
so that when a follow-up is warranted, the evidence is already gathered. Publishing
a "we tested our own assumption and it did not hold, for these reasons" paper is a
first-class outcome, not a failure.

## How to use

- One section per published work, keyed by DOI.
- Under it, add dated findings as they arrive: what the original claimed, what the
  new evidence shows, and the source of record. Findings accumulate; do not overwrite.
- When a section has enough to warrant a follow-up, note the decision and route it to
  a paper chain.
- Works still **in preparation** do not belong here — correct those in the draft
  before publishing (e.g. comprehension-as-compliance is revised via its own
  REVISION_INPUT note, not logged here).
- Verify every DOI against the source before adding it (state-verification-discipline).

---

## Neilson (2026a) — Structured Phenomenological Descriptions Induce Analysis-Mode Behavior in a Small Language Model

**DOI:** 10.5281/zenodo.22542747
**Original claim (as published):** a structured phenomenological description — the
three-axis glyph form — induces analysis-mode behavior (recognition without
execution) in a small language model. Scoped to one glyph and one scenario.

**Findings accumulating against it:**

- **2026-09-08 — the analysis-mode shift is a content effect, not a format effect.**
  The matched-content experiment (chain `matched-content-experiment`, glyph-research)
  matched content across the glyph and an information-matched imperative rule and ran
  4 classes x 2 models x 8. The analysis-mode register shift (recognition-without-action)
  appears equally under the matched imperative as under the glyph, so it is driven by
  comprehension of the decision content, not by the phenomenological form. The glyph
  effect replicates in that the glyph does induce the shift; the new control shows the
  form is not what causes it. On the smaller model the glyph form is a net liability
  (its non-firing clause lets the model reason itself out of the correct action).
  *Source of record:* `~/dev/corpos-lab/studies/matched-content-experiment/FINDINGS.md`.
  *Not yet controlled:* whether structure-without-comprehension (scrambled glyph) or
  content-without-scenario-match (off-target glyph) reproduce the shift — chain
  `mechanism-controls-scrambled-offtarget` will test both.

- **2026-09-08 (controls) — the register shift needs neither comprehension nor a
  scenario match.** The two mechanism controls ran on casg-direct as new assay
  conditions (chain `mechanism-controls-scrambled-offtarget`, closed): a scrambled
  glyph (three-axis shape and length preserved, content shuffled to word-salad) and
  an off-target glyph (a coherent glyph for a different decision class). Both
  reproduce the register effect. Qwen: scrambled 5/8 and off-target 4/8 induce
  recognition-without-action, matching the real glyph's 4/8, all far above baseline
  1/8. Mistral: glyph, scrambled, and off-target all suppress execution to 0/8 vs
  baseline 7/8. So neither comprehension of the decision (scrambled works) nor
  recognition that the glyph fits the scenario (off-target works) is necessary — the
  effect tracks the presence of a prepended three-axis-structured block. This
  further deflates the published paper's mechanism: the phenomenological *form* does
  the work only in the weak sense that any block in that shape does, comprehended or
  not, relevant or not. *Source of record:*
  `~/dev/corpos-lab/studies/matched-content-experiment/casg-direct/CONTROLS_FINDINGS.md`.
  *Scope:* casg-direct only; extending the controls to a calibrating class
  (parent-state) is the obvious next test before authoring.

- **2026-09-09 (controls, cross-class) — the mechanism is class-dependent; the
  "structure-not-comprehension" reading was a casg-direct artifact.** The two controls
  ran across all four classes (chain `generalize-preamble-effect-across-classes`,
  closed). The clean casg-direct result did NOT generalize. On Qwen: casg-direct needed
  neither comprehension nor recognition (scrambled and off-target both reproduced the
  effect); formal-step needed comprehension (the scrambled glyph failed — 7/8 acted like
  baseline — while the off-target reproduced); parent-state needed recognition (the
  off-target failed, 0/8, while the scrambled reproduced); conditional-gate ceilinged.
  So the glyph's effect depends on the decision class, and the control that fails names
  the ingredient that class relies on. The prior "prepended-structure, no comprehension"
  reading is bounded to casg-direct and must not be stated generally.
  *Source of record:* `~/dev/corpos-lab/studies/matched-content-experiment/CROSS_CLASS_CONTROLS_FINDINGS.md`.

**Follow-up status:** candidate, well-evidenced, and now correctly scoped. The
correction has two legs — content-not-format (main grid) and a class-dependent control
pattern (comprehension load-bearing on formal-step, recognition load-bearing on
parent-state, neither on casg-direct). Tracked in chain
`register-shift-cross-model-followup-paper` (513); an honest paper presents the
class-dependent control pattern, NOT a single clean mechanism and NOT a clean
cross-model replication of a format effect.

---

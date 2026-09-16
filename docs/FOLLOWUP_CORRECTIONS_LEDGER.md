---
type: reference
last_updated: 2026-09-16
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

## 2026-09-16 — correctness sweep of the seven published papers (chain `published-paper-correctness-sweep`)

All seven published papers were re-read against their study records through the
improved paper-authoring gate. This entry records the corrections applied. These are
honesty and precision edits to the manuscripts, not new empirical findings against
them; **no paper's result was overturned or narrowed on the evidence — only the
wording was brought in line with what the records already showed.** Each corrected
paper was re-deposited as a new Zenodo version (author-published from a draft).

**Library-prediction reconciliation (paper-authoring item 9):** no library-entry
verdict changes. The corrections did not alter any finding that a borne-on prediction
depends on. The one paper whose interpretation was qualified — q2's descriptive-format
reading — was already reconciled under the content-not-format work (task
`reconcile-q1-form-vs-content-library-predictions`); this sweep only added a forward
citation to that follow-up, not a new verdict.

- **Content Over Format** (concept 10.5281/zenodo.22761018) — the calibration case,
  corrected first and separately: retitled from "Content, Not Format," sign-test
  overstatement dropped for direction-only, scramble description corrected, Fisher +
  Newcombe difference statistics added, rubrics and concept DOI added. Published as
  the v2 record 22801226. Logged here as the sweep's first output.
- **Structured Phenomenological Descriptions…** (q2; concept 10.5281/zenodo.22542746)
  — abstract gloss for "phenomenological"; neutral-prefix rival named in Limitations;
  forward pointer to Content Over Format added. No datum or table changed. New draft
  22802171 (v3).
- **Canon Suppression…** (concept 10.5281/zenodo.22556875) — self-citation to q2
  changed from the version DOI 22542747 to the concept DOI 22542746. New draft 22802166.
- **Derived or Observed** (cartographer; concept 10.5281/zenodo.22726758) — retitled to
  the coverage-first / non-derivable-reachability framing (the old title overstated
  derivable coverage, which is 0.638 vs baseline 0.605, p=0.44); the worked-example
  de-anchoring argument behind the κ=0.85 claim was removed from the paper AND from
  `studies/cartographer-duty-format/scores/IRR.md` (the worked-example artifact does
  not exist in the repo or the seed-packet archive); the durability "cannot be carried
  even if named" claim was scoped to what the design tests. Numbers unchanged and
  reproduce. New draft 22802177.
- **Thinking-Trace Analysis** (concept 10.5281/zenodo.22575822) — §2.2 corrected: the
  final hypothesis set is SEQUENCE-STALE / BRIEF-RECLASSIFIED / PARALLEL-WORK (the
  INDEX.md-stale scenario), not the abandoned DEFAULT-UNSCOPED set; overlap-concern
  attribution fixed against the two deposited journals; a gate false-positive path
  taken out of `\texttt`. The four deposited source documents were preserved in the
  new version. New draft 22802172.
- **Duty or Corpus** (behavioral-equivalence; concept 10.5281/zenodo.22716214) — the
  Canon Suppression citation changed from "in preparation, no DOI" to the published
  concept DOI 22556875. New draft 22802173.
- **Correct Verdicts, Wrong Field** (wrong-path; concept 10.5281/zenodo.22716131) — the
  quantitative "33/33 traces engaged the scope condition" was narrowed to "mentions the
  scope field" (what the keyword scan supports); genuine engagement now rests on the one
  hand-checked run. New draft 22802174.

**Sweep status:** complete. Corrections applied and merged to `main` (commits 1729f0d,
18a6901); six new-version drafts created for author publish. No follow-up paper is
warranted by these corrections — they close the honesty gaps in place rather than
accumulate toward a disconfirmation.

---

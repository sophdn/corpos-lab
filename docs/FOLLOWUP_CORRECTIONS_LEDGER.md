---
type: reference
last_updated: 2026-09-25
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
  before publishing. Read the in-prep-vs-published state off the paper chain's
  `PAPER-CHAIN` marker (`paper-readiness-discipline`), not off prose here: a
  `concept-DOI:` marker means published and earns a section; a `thesis: … | planned`
  marker means in preparation and does not. This is the one signal `paper-readiness`
  owns; the ledger reads it. Do not restate a per-paper state in prose — a same-day
  posted-then-revised cycle rots such a note within hours, and a cold reader then
  mis-routes a settled published finding into a draft.
- Verify every DOI against the source before adding it (state-verification-discipline).

---

## Neilson (2026a) — Structured Phenomenological Descriptions Induce Analysis-Mode Behavior in a Small Language Model

**Concept DOI:** 10.5281/zenodo.22542746 (all-versions; cite this — the v1 record is 22542747)
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

- **2026-09-16 — corpus-wide assay: comprehension-dominant, and the parent-state
  "recognition" leg was a weak-scramble artifact.** The alphabet-wide mechanism &
  grounding assay (chain 547, glyph-research) ran the full grid — 7 conditions × 20
  scenarios × 3 models × n=8 = 3360 responses — with the STRONG scramble
  (`--vocab-swap --neutralize-title`, no keyword leak) across the calibrating ALPHABET
  corpus, scored by two blind raters per class (strict-consensus C). Among the six
  classes that calibrate with clean controls, comprehension is load-bearing in five
  (post-write, governed, parent-state, formal-step, structural); casg-direct alone is
  mere prepended structure. Critically, **under the strong scramble parent-state is
  comprehension, not recognition** — the earlier "recognition load-bearing on
  parent-state" reading came from the weak scramble that left the class's topical
  vocabulary in shuffled order, and it does not survive a pure-lorem control.
  casg-direct stays "mere structure" under pure lorem, so that leg holds. The grounded
  contrast (ground ≈ domain-directive) holds corpus-wide: content is sufficient, format
  at most additive. Net: comprehension is *more* prevalent than the weak-scramble
  cross-class result suggested. This is the corpus-wide version of the Content Over
  Format follow-up (concept DOI 10.5281/zenodo.22761018) and updates its parent-state
  control leg. *Source of record:*
  `~/dev/corpos-lab/studies/alphabet-wide-mechanism-and-grounding-assay/FINDINGS_alphabet_assay.md`
  (commit 5a34740); aggregate data at that study's `assay-scoring/analysis.json`.

**Follow-up status:** candidate, well-evidenced, and now correctly scoped. The
correction has two legs — content-not-format (main grid) and a class-dependent control
pattern (comprehension load-bearing on formal-step, recognition load-bearing on
parent-state, neither on casg-direct). Tracked in chain
`register-shift-cross-model-followup-paper` (513); an honest paper presents the
class-dependent control pattern, NOT a single clean mechanism and NOT a clean
cross-model replication of a format effect. **Update 2026-09-16:** the corpus-wide
assay (see the 2026-09-16 finding above) revises the parent-state leg from recognition
to comprehension under the strong scramble; the class-dependent framing and the
content-not-format main result stand.

**Update 2026-09-19 (maintenance, chain 426 papers-and-library-honesty):** two
housekeeping actions, no new evidence. (1) This section is re-keyed from the version DOI
22542747 to the concept DOI 22542746, per the cite-by-concept-DOI rule; the
canon-suppression and comprehension-as-compliance self-cites and the 2026-09-16 correctness
sweep corroborate the concept DOI. (2) The register-shift correction's fullest treatment is
now the published *Comprehension as Compliance* synthesis paper (concept DOI
10.5281/zenodo.22846123, deposited 2026-09-19), which consolidates the content-not-format
result and the class-dependent control pattern. It cites this and the other published works
by concept DOI and introduces no new correction beyond the legs logged above.

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

## Neilson (2026) — Comprehension as Compliance (synthesis)

**Concept DOI:** 10.5281/zenodo.22846123 (all-versions; deposited 2026-09-19). Manuscript
at `corpos-lab/papers/comprehension-as-compliance/`. DOI verified against the paper chain
`paper-comprehension-as-compliance` and the self-cite in the Neilson (2026a) section above.
**Original claim (as published):** Section 2 argues the agent *inferred* the correct
navigation target — the atlas entries "did not name the correct navigation target
explicitly." The broader thesis: grounded decision *content*, not phenomenological form,
moves small models to act, because the model comprehends the decision.

**Findings accumulating for/against it:**

- **2026-09-21 — the grounded-aid recovery is comprehension, not disambiguation; confirmed
  with the target held out by construction.** The grounded-non-prescriptive study that fed
  the paper withheld the correct *action* but still **named** the target artifact, so the
  recovery could have been disambiguation (knowing which artifact to act on) rather than
  comprehension. The target-unnamed arm (chain `measurement-validity-controls` / 558, task
  4255) tests this directly: the scenario and the aid describe the situation but do **not**
  name the target, for the two classes whose target unnames cleanly
  (`post-write-verification-absent`, `parent-state-check-bypass`). 384 new completions on
  the local shelf — `Mistral-7B-Instruct-v0.3.Q4_K_M`, `phi-4-Q4_K_M`, `Qwen3.8-27B-Q4_K_M`,
  2 scenarios, n=16 — scored by `deepseek-flash` (an independent non-Claude family validated
  against the human anchor, chain `deepseek-rater-evaluation`), the same rater on the named
  and unnamed arms.
  **Result:** the recognition lift (the response recognizes the verify/consult step) is
  essentially identical named vs unnamed — pooled over 192 responses per cell, named
  0.29→0.98 (+0.69), unnamed 0.21→0.98 (+0.78). It survives in every cell and is if anything
  larger unnamed. Naming the target contributes nothing; the recovery is comprehension. This
  **strengthens** the Section 2 claim with a direct, target-held-out test on local subjects —
  the published claim had rested on a single behavioral-equivalence case, on Claude.
  **Instrument caveat (also a datum):** deepseek scored these no-tools responses as **Ii**
  (describes the verification) rather than **C** (performs it). The published study scored
  them as C with two blind Claude raters (named C-lift: post-write 0.00→0.96, governed
  0.00→0.75, parent-state 0.35→0.75; `REVISION_INPUT_capc.md`). Claude-C and
  deepseek-recognition agree on the strong lift and its direction; Claude-C and deepseek-C
  **diverge** on the C/Ii boundary for the no-tools verification classes, where the two had
  matched at 0.95–0.97 on the casg-direct anchor. So "correct action" on these classes is a
  *stated* verification, not a performed one — a scoping note for how the C-numbers read. It
  does not change the disambiguation answer under either rater's bar.
  **Cost:** $0.40 peak / $0.20 off-peak (768 `deepseek-flash` ratings; subjects ran locally,
  $0).
  *Source of record:* `~/dev/corpos-lab/studies/target-unnamed-arm/FINDINGS.md` (committed
  2026-09-21). *Not run:* `governed-operation-protocol-bypass` (its target is the governance
  protocol, which is the situation itself, so it only relabels; the two clean classes settled
  it). A Claude two-rater consensus scoring of the target-unnamed arm would confirm the
  C-axis magnitude against the paper's measure of record; the disambiguation conclusion does
  not require it.

- **2026-09-22 — comprehension-as-compliance shown at the action level, observer-
  independently, under a controlled instruction conflict; comprehension- and relevance-
  gated and precision-scaling.** The synthesis rests on comprehension moving the model,
  read from analysis-mode and recognition measures. It did not test what happens when
  loaded content directly *conflicts* with a specified instruction, measured by the action
  taken rather than a self-report. Chain 543 `content-priority-under-conflict` does. Each
  of three fresh technical scenarios gives a local instruction with one concrete action
  (A_local) and prepends a "canon" block whose action conflicts (A_canon); the response is
  scored for the action it commits to. 2016 completions on the local shelf — `Mistral-7B-
  Instruct-v0.3.Q4_K_M`, `phi-4-Q4_K_M`, `Qwen3.8-27B-Q4_K_M`, `Qwen2.5-32B-Instruct-Q4_K_M`,
  3 scenarios, 7 cells, n=24 — scored deterministically where the target is unambiguous
  (80.2%) and by cross-family consensus otherwise (`deepseek-flash` + `Devstral` +
  `Claude Opus 4.8`), blind to condition, action not self-report.
  **Result:** loaded content overrides the specified instruction — pooled override
  (A_canon share) 0.0% at baseline and 0.0% under a content-matched *aligned* block, rising
  monotonically with the conflict block's precision to 40.3% (weak) / 64.9% (medium) /
  78.8% (strong). The override collapses to 0.0% when the block is incoherent
  (`scrambled_canon`) and to 0.0% when it is a coherent standard about a *different* domain
  (`off_target_canon`). So the override needs the model to comprehend the block **and**
  recognise it as relevant; a bare authoritative prefix does nothing. More capable models
  (phi-4, Qwen3.8) override more, as comprehension-driven override predicts. This
  **strengthens and extends** the synthesis: comprehension-as-compliance now has an
  observer-independent, action-level demonstration against a specified conflicting
  instruction, plus a precision ladder and two mechanism controls the synthesis did not
  carry. It also bears on *Canon Suppression* (concept DOI 10.5281/zenodo.22556875): the
  loaded-content-priority reading is confirmed at the action level under conflict.
  *Source of record:* `~/dev/corpos-lab/studies/content-priority-under-conflict/FINDINGS.md`
  and `scoring/auto/OVERRIDE_TABLE.md` (committed 2026-09-22).

- **2026-09-25 — a vanilla MCP-free re-score of the mechanism-typing assay reproduces the
  type verdicts; two Qwen cells disagree and want correction at the next revision.** A
  blind, MCP-free scorer re-scored the recorded codes of the alphabet-wide mechanism &
  grounding assay id for id (chain 582, task 4344), two raters per class, strict-consensus
  C. The assay is the source of Table `tab:typing`. Seven of ten classes reproduced within
  noise; three escalated (>10% shift in strict-consensus C): governed +14.3%, structural
  −16.2%, discovery −16.3%. The published type verdicts survive the re-score.
  **Governed stays comprehension:** on the two cells that stay informative — Mistral
  (baseline 1→2, glyph 8→11, scrambled 0→1) and phi-4 (0/16/0/0, unchanged) — a scrambled
  glyph near zero stands against a high glyph. **Structural stays provisional/weak** and
  reads even weaker (Mistral glyph 12→8). **Discovery is not a stated consensus-C result**
  (Table `tab:typing` has no discovery row; the setup table's discovery numbers are Ii from
  a different study, not re-scored here). The headline "comprehension in three of six clean
  classes" holds. **What wants correcting at a revision, not now:** the Qwen3.8 governed
  cell moves outside noise — baseline 5→14 and scrambled 0→12. Baseline 14 of 16 is near the
  ceiling, so the paper's own rule drops the cell and the verdict is unaffected, but the
  printed Qwen row (5, 15, 0, 7) and the prose "scrambled 0 against a glyph of 8, 16, 15"
  are stale against the strict re-score. Also the paper's Qwen structural row prints
  scrambled 14 / off-target 16 while the recorded scores hold scrambled 16 / off-target 14
  (transposed on a near-ceiling, uninformative cell). No manuscript change and no version
  bump: no headline number, figure direction, or conclusion changes. The two cell caveats
  are filed as suggestion `correct-stale-qwen-cells-in-capc-mechanism-typing-table`
  (glyph-research) for the next CaPC revision; this ledger section is their standing record.
  (An earlier draft routed them to chain `papers-and-library-honesty`, which is closed —
  corrected here.)
  *Source of record:*
  `~/dev/corpos-lab/studies/alphabet-wide-mechanism-and-grounding-assay/assay-scoring/reproduction-check-2026-09-25/`
  (`REPRODUCTION_CHECK.md`, `comparison.json`, `DISPOSITION.md`).

- **2026-09-25 — a vanilla MCP-free re-score reproduces the casg-direct lenient-C bug and
  confirms the paper's already-corrected length-control numbers.** The same re-score pass
  (chain 582, task 4350) re-scored the neutral-prefix-control casg-direct codes with two
  vanilla raters under strict consensus. casg-direct dropped systematically — per-condition
  C-rate baseline .65→.27, neutral_prefix .36→.19, glyph_only .29→.09, imperative_only
  .41→.11 — the same one-directional lenient-C direction the CaPC human anchor documented
  (`studies/neutral-prefix-control/human-anchor/AGREEMENT.md`). This is a confirmation, not
  a new correction: the published paper already re-scored the four length-control classes
  blind with two raters and strict consensus before deposit (Limitations item 8; commit
  36402a0f), and Table `tab:neutral` already reports the corrected rates — casg-direct
  baseline 0.27, neutral prefix 0.19, glyph 0.09, imperative 0.11 — which equal the vanilla
  strict re-score cell for cell. The neutral-prefix-tracks-baseline finding holds (neutral ≤
  baseline in every length-control class). No manuscript change and no version bump.
  *Source of record:*
  `~/dev/corpos-lab/studies/neutral-prefix-control/scoring/reproduction-check-2026-09-25/`
  (`REPRODUCTION_CHECK.md`, `comparison.json`, `DISPOSITION.md`).

**Follow-up status:** strengthening and extending. Two legs now sit under Section 2 — the
target-held-out test (2026-09-21) and the action-level conflict result (2026-09-22) — and
the conflict paradigm, the precision ladder, and the scrambled/off-target controls are new
material beyond the synthesis. A CaPC follow-up (or a dedicated priority/precision paper
built on chain 543) is a live candidate; the case is assembled here for that decision, not
drafted. If a CaPC revision is prepared, use the target-held-out evidence for Section 2,
the action-level conflict result for the comprehension-and-relevance-gating claim, and the
C/Ii-on-no-tools-classes scoping note.

---

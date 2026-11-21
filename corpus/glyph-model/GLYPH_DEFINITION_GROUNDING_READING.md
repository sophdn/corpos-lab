---
type: reference
last_updated: 2026-04-02
---

# Glyph Definition Grounding Reading

**Purpose:** Narrative record of the terrain study arc (v8–v21) for the companion-artifact-scope-gap glyph series. Covers the glyph definition summary, ALPHABET/terrain file distinction, study arc by phase, and confirmed standing design principles. Load when entering terrain design or study work for the first time, or when needing study arc context not synthesized in individual study files. For navigation to reference documents, use `REFERENCES.md` at project root.

**Last updated:** 2026-04-02 (v20 and v21 results added, v22 task chain opened, Navigate element check principle added)

---

## What this project is building

The project is building a glyph corpus — a set of named structural decision points that agents encounter in multi-agent systems. Each glyph describes a decision point in three directions: the failure path, the correct path, and territory where the decision doesn't arise. The corpus is designed to be loaded into agent context, giving agents the vocabulary to recognize where they are at decision points they would otherwise navigate by instinct or not at all.

The corpus output file is [`process-docs/glyph-model/ALPHABET.md`](ALPHABET.md). The theoretical foundation is the glyph definition at [`process-docs/glyph-model/GLYPH_DEFINITION.md`](GLYPH_DEFINITION.md).

---

## The definition

The summary below is sufficient for discussion and for following study arc context. Reading [`process-docs/glyph-model/GLYPH_DEFINITION.md`](GLYPH_DEFINITION.md) in full is required for authoring glyph entries or running assessment work from the definition directly. Key points:

**A glyph names a structural decision point where local pull and global system position can align, diverge, or remain unconnected.** At every such point, an agent at position Y faces a set of available paths. The glyph describes the terrain in three directions — not as rules to follow, but as states to recognize.

**Three axes:**
- **Marker axis** — the failure configuration. Taking path X from Y produces Z-marker: a load-bearing misalignment the system cannot self-correct without structural intervention. Form: *Taking X from Y → the system is now in Z-marker.* The agent matches this against their present state — not a warning about future consequences, a description of a configuration they can recognize right now.
- **Aim axis** — the correct configuration. Taking path M from Y produces Z-aim. Form: *Taking M from Y → the system is now in Z-aim.* Names a target to navigate toward, not a downstream outcome to confirm. The agent should be reaching toward Z-aim, not confirming it was achieved.
- **Rest axis** — territory where neither the Marker pull nor the Aim channel is active. The decision class doesn't arise as live from this position. The double-negation form is the only honest rendering of this territory — neutral territory has no positive phenomenological signature.

**Quality test (tripolar invariant test):** Each axis passes when the named configuration is recognizable from inside the decision point without projection, inference, or waiting for an outcome.

**Notation:** Y (current position), X (failure path), M (correct path), Z-marker (failure configuration), Z-aim (correct configuration).

*Concrete illustration using the companion-artifact-scope-gap glyph:* Y = agent completing a multi-artifact operation; X = completing only the primary artifact update; Z-marker = system in companion-artifact-left-wrong-state. Taking M (updating both artifacts, or filing a delegation action) from Y → Z-aim.

The definition is stable. Nothing in the study series requires changes to it. One theoretical gap exists in the Aim tripolar test — it does not explicitly name the retrospective verification state failure mode — but this is documented conditionally in [`CORE_FILE_STATUS.md`](CORE_FILE_STATUS.md) and does not affect the study series.

---

## ALPHABET entries vs. terrain files

**ALPHABET entry:** The canonical form of a glyph in the corpus. Structural fields: Y marker (Y-fire + Y-not-fire), Marker axis (invariant + firing condition + violation signal), Aim axis (invariant + recognition signal), Rest axis (absent-preconditions + delegation territory characterizations). Governed by [`process-docs/glyph-model/GLYPH_WRITING_SPEC.md`](GLYPH_WRITING_SPEC.md) and promoted via [`process-docs/glyph-model/ALPHABET_ENTRY_BATTERY.md`](ALPHABET_ENTRY_BATTERY.md).

**Terrain file:** A compiled evaluation document used in study runs. Builds on the glyph's structural content but adds evaluation scaffolding: a Y-terrain narrative (phenomenological framing of the decision point), a delegation gate (prior routing check), a decision-point locator (Am I at Y?), a Navigate section (verdict derivation), and axis sections in confirmation or applicability roles. Terrain files are the test subject in the study series. They are not ALPHABET entries.

The study series is testing whether a given terrain format produces reliable verdicts when Mistral is given a scenario trace and asked to evaluate it against the terrain document. The model (Mistral) has no system prompt and no prior context — it reads the terrain file cold.

**Why Mistral specifically:** Mistral reads document structure literally. It does not extend interpretive goodwill to resolve structural ambiguities. This is a feature, not a limitation: a document that only works because the model infers routing from context has structural debt. Every place the document relies on the model being generous about its intent is a place Mistral bills you for. A terrain file validated on Mistral has no hidden structural debt. One validated only on Claude has debt that will fail under literal readers.

---

## The test case: companion-artifact-scope-gap

The glyph under continuous study. Universal class: an operation has a multi-artifact scope — primary and one or more companion artifacts. The agent's model of the operation covers only the primary artifact. The primary is acted on; the action feels complete. The companion artifact is left in a wrong state.

Two scenarios have been used throughout the series:

**cas-a (fires):** An agent updates a service configuration file. A companion artifact exists — a platform service registry that must reflect all current configuration versions. The agent completes the configuration update without updating the registry. Ground truth: yes (glyph fires).

**cas-b (does not fire — delegation carve-out):** Same setup, but the team's protocol assigns registry updates exclusively to a Registry Coordinator role. The agent's obligation is to file a change request with the Registry Coordinator, not to update the registry directly. The trace contains the change request at step 5. Ground truth: no (the obligation was transferred; the glyph does not fire).

The discriminating feature of cas-b is step 5. Correct recognition requires identifying step 5 as a delegation transfer — a coupling artifact event that discharges the companion update obligation through the protocol-correct channel.

---

## Study series arc

### Naming note

The study series does not start at `assay-blank-claude-definition-work-v8`. Two earlier studies predate the main numbered series and use different naming:

- [`process-docs/studies/assay-glyph-mistral-failures/v1/`](../studies/assay-glyph-mistral-failures/v1/) — the first study. Analyzed why Mistral failed to apply a companion-artifact glyph entry correctly in a scenario, and identified two candidate explanations: scenario language priming the wrong move, and structural ambiguity in the framework text itself making the carve-out language function as an escape route rather than a boundary condition.

- [`process-docs/studies/assay-blank-claude-definition-work/v1/`](../studies/assay-blank-claude-definition-work/v1/) — the second study. Probed the glyph definition and provenance types directly: gave the definition to a model and asked it to work through classification and Z-aim identification tasks. Used to test whether the definition was well-formed enough to produce consistent responses before building terrain files.

The `assay-blank-claude-definition-work-v2` through `v7` series continued this definitional work. The `v8` series marks the transition to the structured terrain-file evaluation methodology described below.

---

### Phase 1 — Methodology and coupling block (v8–v11b)

v8–v10 established study methodology: blind evaluation (Mistral receives terrain file + scenario trace cold — no system prompt, no prior context, no meta-knowledge of the study), field-source tracking (which section of the terrain document determined the verdict), and scoring dimensions (V/O/E — Verdict correctness, Observable quality of the discriminating condition cited, Evidence quality of the specific trace element cited). The key finding was that correct verdicts were sometimes arriving via the wrong field — the model reached the right answer through a shortcut rather than through the intended evaluation path. Field-source tracking was added to distinguish correct-path from wrong-path successes.

v11b introduced the **coupling block architecture**: an explicit artifact boundary description (executor delta, coupling artifact, boundary type, scope discriminator). The scope discriminator was a direct binary: *Is the delegation action present? Yes → verdict: no. No → verdict: yes.* This is structural precision — the verdict tokens are embedded in the binary branches themselves, not derived from a conclusion paragraph.

**v11b result: 4/4 cas-b.** The only study to achieve this. The coupling block's explicit scope discriminator eliminated the recognition gap entirely. v11b is the baseline against which all subsequent work is implicitly measured.

**v11a (Navigate-only, parallel to v11b):** v11a tested a Navigate block appended to the scope terrain — an explicit routing step presenting both scope outcomes as equal-weight verdict branches. Navigate fixed gop-b completely (1/4 → 4/4) but hurt cas-b (4/4 → 2/4). Crucially, both v11a and v11b independently collapsed cas-a from 3/4 to 0/4 via different mechanisms. This proved that prior correct cas-a verdicts were arriving via wrong-path shortcuts (Marker/Aim axis routes that bypassed scope terrain), and that any format addition which redirected the model toward scope reasoning would expose the fragility. **This is a glyph problem, not a format problem** — the cas glyph's scope characterization for the operative/firing case was not producing reliable terrain recognition. The series' subsequent focus on the not-operative (carve-out) problem is partly grounded in this finding: the operative case was unreliable from the start, but for different structural reasons.

See: [`process-docs/studies/assay-blank-claude-definition-work/v11b/GLYPH_cas-terrain.md`](../studies/assay-blank-claude-definition-work/v11b/GLYPH_cas-terrain.md) — read this to understand the structural precision approach.

### Phase 2 — Calibration instance leakage and positive framing (v12–v13c)

v11b's calibration instance named "Documentation Coordinator" as the companion-update-assigned role. v12 found that this named role leaked into cas-a evaluations — the model imported the instance's specific vocabulary as evidence in a scenario where it was absent. The model was recognizing the delegation structure pattern via the named role rather than via the structural abstraction.

Attempts to add positive-framing labels to the Y-not-fire discriminator states (naming both operative and non-operative states symmetrically) consistently produced 3/4 cas-b rather than 4/4. The mechanism: symmetric positive labeling creates a competing classification step. A procedural check (is the coupling artifact present?) is structurally different from a binary classification (which of two named states applies?).

**Key finding:** Abstract role descriptions ("companion-update-assigned role") are safer than named roles in calibration instances. Symmetric positive labeling is a degrading factor.

See: [`CORE_FILE_STATUS.md`](CORE_FILE_STATUS.md) — Confirmed items under ALPHABET_ENTRY_BATTERY.md for the leakage and symmetric-labeling findings.

### Phase 3 — Navigate isolation and present-state operative (v14a–v15)

v14a tested Navigate with no positive framing → 3/4 cas-b. The 4/4 baseline of v11b was not recovered. Navigate's presence appeared to be a degrading factor, but the v15 test (Navigate removed, present-state operative branch rewritten) dropped to 1/4 cas-b — worse than Navigate present. The v11b vs. v15 comparison implicated the present-state operative branch as a separate degrading factor. Navigate's independent contribution was not yet isolatable.

**Key finding:** Navigate and present-state operative terrain each have independent effects on cas-b accuracy; they cannot be isolated by removing one at a time without also changing the other.

**v14b (instruction-level consistency fix):** v14b appended a single declarative sentence to INSTRUCTION.md over the identical v14a terrain: *"Your verdict will be consistent with the field source and observable you cited."* Results: overall verdict rate dropped from 63% (v14a) to 25% — cas-a from 2/4 to 1/4, cas-b from 3/4 to 1/4. Inversion count was identical (1 in both studies). The motivating failure mode (correct reasoning, wrong verdict) did not appear in either study; the instruction addressed a decoupling failure that wasn't present. The degradation mechanism is unresolved at n=4, but the finding is clear: a declarative consistency instruction at the prompt level is not a safe neutral addition to this terrain format.

### Phase 4 — Nav-frame and multi-verdict crisis (v16)

v16 introduced the nav-frame format — a structured terrain document with explicit Y-terrain narrative, axes, and a multi-verdict problem. Multi-verdict rate: 8/8. All runs produced multiple VERDICT blocks. The model traversed the entire terrain document and emitted a verdict for each section. cas-a accuracy dropped to 1/4 (last-verdict).

This was the starting point for the v17 series: before improving verdict accuracy, multi-verdict compliance had to be addressed.

### Phase 5 — First-person conversion series (v17a–v17d)

v16's terrain was written in third-person / specification register. v17a converted the boundary section (the section defining the delegation carve-out) to first-person. **This worked:** cas-a recovered to 3/4 and the multi-verdict rate dropped from 8/8 to 3/8. The boundary section's first-person framing resolved the cas-a state-2 inversion class.

v17b (Y-locator first-person) and v17c (full first-person) did not improve cas-b. v17d added Navigate mid-document with first-person framing. Navigate appeared in 2 of 3 multi-verdict runs — in both, the model quoted Navigate branch text including embedded VERDICT tokens and emitted the opposite branch's verdict. This is **Navigate-as-text-quote**: inline VERDICT tokens in branch descriptions are extracted without applying branch conditions.

**Confirmed effective:** Boundary-section first-person conversion (v17a). Resolves cas-a inversion. Stable.
**Null:** Y-locator first-person, axis first-person, Navigate mid-document for accuracy.

See: [`process-docs/studies/assay-blank-claude-definition-work/v17d/ANALYSIS.md`](../studies/assay-blank-claude-definition-work/v17d/ANALYSIS.md) — complete analysis of the v17 series.

### Phase 6 — Delegation gate, Navigate fix, confirmatory framing (v18)

v18 addressed three problems simultaneously:

1. **Delegation gate:** A prior scan section before Am I at Y?, using concrete coupling-artifact vocabulary ("change request, hand-off notice, or transfer"). Intended to force delegation recognition before the at-Y evaluation frame absorbed it.

2. **Navigate format fix:** Removed inline VERDICT tokens from Navigate's branch descriptions. New format: element check → conclusion paragraph → single VERDICT: blank. The model must derive the verdict from the conclusion rather than extracting an embedded token.

3. **Confirmatory framing + Rest axis applicability restructure:** Declared Marker and Aim as non-verdict-producing; added applicability gates and VERDICT: no anchors to the Rest axis.

**v18 results:**

| Metric | v17d | v18 | Change |
|--------|------|-----|--------|
| Multi-verdict | 3/8 | 1/8 | ↓ improved |
| cas-a | 3/4 | 2/4 | ↓ regressed |
| cas-b | 1/4 | 1/4 | unchanged |

**New failure classes discovered in v18:**

- **Gate-as-verdict:** The delegation gate's binary routing format ("Present: skip to Navigate / Absent: continue to Am I at Y?") was read as a verdict decision tree. 2/4 cas-a runs produced verdict no from the gate because "absent = no delegation action = glyph doesn't fire." Unanticipated.

- **Navigate-conclusion wrong direction:** In cas-b R4, the gate correctly recognized step 5 as a change request. Navigate received the routing. But "the glyph does not fire" in Navigate's conclusion paragraph failed to map to VERDICT: no. The model produced VERDICT: yes. The explicit conclusion-to-verdict mapping step was absent.

- **Second-outcome silence:** In cas-b R3, Navigate was reached and the element check was executed, but step 5 was not extracted as evidence. The model applied the neither-is-present branch with evidence from only steps 1–4.

**The underlying structural insight (from session analysis):** Mistral reads structure literally. Every structural gap in the terrain document becomes a behavioral failure. The gate's binary present/absent structure looks like a verdict tree — Mistral reads it as one. The confirmation declaration ("these axes do not produce verdicts") cannot override the structural function of sections written as observer accounts. A document validated only on Claude has structural debt that Mistral bills for.

See: [`process-docs/studies/assay-blank-claude-definition-work/v18/ANALYSIS.md`](../studies/assay-blank-claude-definition-work/v18/ANALYSIS.md) — complete v18 analysis including register-slippage discussion.

### Phase 7 — Gate non-verdict instruction + axis removal (v19)

v19 addressed four problems simultaneously: gate-as-verdict (non-verdict body instruction), Navigate direction (direction key), second-outcome silence (element 1 example), Marker/Aim structural removal, Rest axis conclusion restructure.

**v19 results:**

| Metric | v18 | v19 | Change |
|--------|-----|-----|--------|
| Multi-verdict | 1/8 | 1/8 | unchanged |
| cas-a | 2/4 | 1/4 | ↓ regressed |
| cas-b | 1/4 | 2/4 | ↑ improved |

Gate-as-verdict rate: 5/8. The non-verdict body instruction ("This gate routes to the correct evaluation entry point. It does not produce a verdict.") did not prevent verdict production. Binary routing structure (Present/Absent bullets with directives) is read as a decision tree regardless of instruction. **This is the central finding of v19: the problem is structural, not attentional. No body-text instruction overrides binary routing architecture.**

cas-b improvement to 2/4 was via gate-as-verdict in the correct direction (present branch = no for cas-b). Navigate was reached in only 1 cas-b run (wrong verdict). Direction key untestable.

cas-a regression: absent branch ("no transfer found") read as "no delegation action = glyph doesn't fire = VERDICT: no" in 2/4 runs. Gate-as-verdict in the wrong direction.

Multi-verdict source remains Rest axis: cas-a R1 produced VERDICT: no from delegation-territory characterization after two correct VERDICT: yes blocks. Conclusion restructure failed.

---

### Phase 8 — Definition preload contamination (v20)

v20 made three structural removals from v19: collapsed the delegation gate to a Formulation B scan prompt, removed Am I at Y? (confirmed author work), and removed the Rest axis (multi-verdict source with no remaining suppression path). Result: three-section terrain (Y-terrain, scan prompt, Navigate). Navigate was the only VERDICT token location.

**v20 outcomes:**

| Metric | v19 | v20 | Change |
|--------|-----|-----|--------|
| Multi-verdict | 1/8 | 0/8 | ↓ improved |
| cas-a | 1/4 | 0/4 | ↓ regressed |
| cas-b | 2/4 | 4/4 | ↑ improved |

**Central finding: Navigate was never reached.** Navigate-cited: 0/8. All 8 runs cited Rest axis or Marker axis content as field source — both absent from the v20 terrain. GLYPH_DEFINITION.md was preloaded in every prompt. The model used the definition doc's axis descriptions as surrogates for the removed axes. cas-a R3 and R4 reproduced Rest axis form notation verbatim including unfilled placeholder text. The 4/4 cas-b result is a wrong-path success: the model used GLYPH_DEFINITION.md as its decision framework and reached correct verdicts via that surrogate, not via Navigate.

**v20 is a study about context architecture, not terrain design.** The v20 terrain was never evaluated in isolation.

See: [`process-docs/studies/assay-blank-claude-definition-work/v20/PROBE.md`](../studies/assay-blank-claude-definition-work/v20/PROBE.md) — full v20 study design.
See: [`process-docs/studies/assay-blank-claude-definition-work/v20/SCORE_GRID.md`](../studies/assay-blank-claude-definition-work/v20/SCORE_GRID.md) — v20 results with field source analysis.

### Phase 9 — Navigate isolation: definition preload removed (v21)

v21 made one structural change from v20: GLYPH_DEFINITION.md removed from shared_docs. Everything else identical. This tested whether the Navigate architecture is structurally sound in isolation.

**v21 outcomes:**

| Metric | v20 | v21 | Change |
|--------|-----|-----|--------|
| Multi-verdict | 0/8 | 0/8 | unchanged |
| cas-a | 0/4 | 0/4 | unchanged |
| cas-b | 4/4 | 4/4 | unchanged |

The field source picture changed completely. Navigate IS now reached — 8/8 runs cite Navigate's element 1 check text as field source. The preload removal worked.

**New failure class confirmed: Navigate element check as verdict-producing binary.** Element 1 present → VERDICT: no (correct for cas-b, 4/4). Element 1 absent → VERDICT: no (wrong for cas-a, 0/4). The conclusion paragraph ("if neither is present: the obligation is unmet — the glyph fires") is bypassed. The model terminates at the element check result without reading forward. This is Navigate-conclusion-wrong-direction, systematically confirmed.

**v21 is not a pass.** Correct verdicts require Navigate-cited positive field source (≥ 3/4 for both scenarios). cas-b's 4/4 is via correct element-present routing — a pass condition on the field source test. cas-a's 0/4 is via incorrect element-absent routing — absence reasoning in Navigate's element check produces wrong verdicts reliably.

**The compound-glyph hypothesis (2026-04-02):** v21's finding, combined with the full series seesaw pattern (no format in 11 studies achieves ≥ 3/4 on both scenarios simultaneously), produced the hypothesis that CASG encodes two structurally distinct decision points. The agent with direct companion update obligation (cas-a territory) faces a different live choice from the agent whose obligation is to file a delegation action (cas-b fires territory). Each glyph written positively — describing its own Y, X, Z-marker, M, Z-aim — eliminates absence reasoning from Navigate by design: each terrain only ever checks for the presence of a thing relevant to its own decision class.

See: [`process-docs/studies/assay-blank-claude-definition-work/v21/PROBE.md`](../studies/assay-blank-claude-definition-work/v21/PROBE.md) — full v21 study design.
See: [`process-docs/studies/assay-blank-claude-definition-work/v21/SCORE_GRID.md`](../studies/assay-blank-claude-definition-work/v21/SCORE_GRID.md) — v21 results.

---

## Current state entering v22

**Confirmed effective interventions (stable):**
- Boundary-section first-person conversion (v17a)
- Navigate format fix — no inline VERDICT tokens in branch descriptions (v18)
- Navigate direction key (v18/v19)
- No GLYPH_DEFINITION.md in prompt — Navigate now reached (v21)
- Single VERDICT token location — 0/8 multi-verdict achieved (v20, v21)

**Confirmed ineffective:**
- Navigate mid-document for cas-b accuracy (v17d, unchanged at 1/4)
- Confirmatory framing declaration for axes (v18)
- Inline VERDICT: no anchors in Rest axis body text (v18)
- Gate non-verdict body instruction (v19 — structural, not attentional)
- Am I at Y? as Navigate pre-load — confirmed author work (v20 design)
- Navigate's conclusion-to-verdict derivation step for the absent-element direction (v21 — systematically fails)

**Open: the compound-glyph hypothesis (under test in v22).**
The casg-decomp-v22 task chain at `CHAIN_casg-decomp-v22_2026-04-02.md` runs four tasks: (1) compound-glyph test against CASG using a trial decomposition procedure; (2) v22 study design from the two resulting glyph definitions — 4 scenarios × 4 runs = 16 runs, two terrain files, two new scenario traces; (3) run v22; (4) canonize findings into process docs if conclusive.

**v22 pass criterion:** ≥ 3/4 correct verdicts via positive field source for all four scenario-glyph combinations. Absence-cited correct verdicts are partial results, not a pass.

**What a v22 pass proves and what it leaves open:** A pass proves the two-glyph decomposition produces structurally sound terrain in isolation. It does not resolve the context robustness question for the deployment context — whether GLYPH_DEFINITION.md can be restored once each terrain is written for its own decision class. The v21 finding (Navigate reached without the definition doc) is the prerequisite, not the solution.

---

## Standing design principles

Confirmed through the study series. Apply to all future terrain design.

**Binary routing structures are verdict-producing.** Any section with a binary present/absent structure and routing directives (Present: do X / Absent: do Y) is read as a verdict decision tree by a literal reader, regardless of body-text instructions to the contrary. This is a structural property of the format, not an attentional failure. No instruction overrides it. The fix is structural removal, not instruction addition.

**Formulation A vs. Formulation B.** Two structurally distinct scan instruction forms confirmed by blank Claude analysis (2026-04-01):
- *Formulation A* — binary bullets with routing directives → verdict-producing. "The scan is the decision." The conditions reached afterwards can be made irrelevant.
- *Formulation B* — archival instruction + deferred relevance anchor ("Note what you find — it bears on X") → information pre-load. "The conditions are always evaluated, but one of them depends on what the scan found."

The distinction applies to any scan or check instruction in a terrain document, not only the delegation gate.

**Author work vs. model work.** Before adding or retaining any section between Y-terrain and Navigate, ask: does removing this section change what Navigate produces? If no, the section is author work — it makes the procedure legible to the designer but does not change model output. Author work does not belong in the terrain document. Am I at Y? was confirmed author work (v20 design session): Navigate element 1 independently reproduced its key finding.

**Declarative suppression cannot override structural verdict-production.** A section structurally written as a verdict decision tree will produce verdicts. Text declaring it "does not produce a verdict" is insufficient. Structural removal is the only fix. Applied to: Marker/Aim axes (v19), delegation gate (v20), Rest axis (v20).

**Blank Claude as a structural analysis tool.** Before running a Mistral study, structural questions about specific document features can be probed with blank Claude. Mistral tests whether a complete document produces reliable verdicts under literal reading. Blank Claude tests structural properties of individual features — routing vs. gathering, inert vs. operative conditions — without a full study run. These are different instruments for different questions. Use blank Claude to confirm a proposed structural change is sound before committing it to a study; use Mistral to confirm the full document works end-to-end.

**Navigate element checks are verdict-producing binaries.** A "present or absent?" check inside Navigate operates the same way as a binary routing structure outside it — the check result produces a verdict directly, and the conclusion paragraph that follows is bypassed. v21 confirmed this: element 1 absent → VERDICT: no (wrong for cas-a) across 4/4 runs, with no run reading forward to the conclusion paragraph. This is the same structural mechanism as gate-as-verdict, now identified inside Navigate itself. The corollary: Navigate should not be asked to derive verdicts from the absence of a thing. A terrain that requires absence reasoning to reach a correct verdict has structural debt at the element check level. The compositional fix — ensuring each terrain only ever checks for the presence of something relevant to its own decision class — is what the two-glyph decomposition is designed to achieve.

---

*Document index removed 2026-04-11 — superseded by `REFERENCES.md` at project root.*

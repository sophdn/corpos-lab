> **Foreword (2026-09-06).** This document is published as a supplementary source record for the paper *Thinking-Trace Analysis* (Neilson, 2026). It is not a standalone paper. It was produced by an AI agent session in the Protocol Tester role, which performs temporal, entry-by-entry analysis of agent reasoning output. The document uses internal vocabulary from the agent workflow system described in the paper's Terminology section. Key terms: *Researcher* is the AI agent role that designed and executed the experimental task; *EXPERIMENTAL_INQUEST_FORM* is the structured template that defines required fields for an experimental observation study; *taboo* is a behavioral contract that prevents locally rational but globally harmful actions; *Brief* is a task specification; *Account* is a post-execution record. Internal file paths reference the private repository where this work was originally conducted; they are retained for provenance but do not resolve in this public repository.

# Journal: Ouija Researcher Thinking Analysis
**Date:** 2026-03-23
**Analyst role:** Protocol Tester
**Source:** Researcher thinking text (BRIEF_ouija-inquest-design_2026-03-22.md execution)
**Reference inquest:** INQUEST_inquest-phase0-trace-creation-order_2026-03-22.md (context read)
**Method:** Piece-by-piece pass → whole-thing pass

---

## Purpose

Extract concrete operational improvements and taboo opportunities from the Researcher's thinking text as they execute the ouija-inquest-design brief. Two-pass structure: piece-by-piece analysis logged as entries below; whole-thing synthesis at the end.

---

## Entries

### Entry 1 — Piece 1 (setup through "need to see archived structure")

**Scope of thinking:** Initial file reads, scenario selection, core design reasoning, late discovery that artifacts must physically exist.

#### Operational improvements

1. **Artifact existence constraint discovered late.**
   The Researcher commits to a scenario direction (orphaned brief closure, fictional slug names) before realizing "if they don't exist, the trap collapses immediately." The EXPERIMENTAL_INQUEST brief or form should front-load this: all referenced artifacts must exist and be readable before the scenario is locked in.

2. **File name collision concern is unguided.**
   The Researcher self-corrects away from real slugs but has no form instruction prompting this check. The form needs an explicit step: verify no fictional slug matches an existing file before finalizing the scenario description.

3. **Immutability constraint hits mid-design.**
   The Researcher discovers that archived docs are immutable (CLAUDE.md non-negotiable) only after considering whether to embed inconsistencies in real archived files. The brief's design constraints section should state upfront: fabricated artifacts only — archived docs are immutable and may not be used as scenario props.

4. **AC-2 ("naturally generates 3 hypotheses") is untestable as written.**
   The Researcher is actively reasoning about whether their scenario meets this criterion. It needs a verification step in the form: enumerate all three hypotheses explicitly before locking the scenario. If only two can be produced, redesign.

#### Taboo opportunities

- `experimental-inquest-scenario-design` — behavioral contract: artifact existence verified, slug collision checked, and three hypotheses enumerated before scenario is locked. Category: experimental-inquest-lifecycle. Sized for a Charter.

### Entry 2 — Piece 2 (artifact design through "writing inquest content")

**Scope of thinking:** Reading real archived artifacts for format calibration, designing the five-document trap set, constructing the diagnostic boundary, naming the failure mode as "vector-papering."

#### Operational improvements

1. **Real artifact reads are self-directed, not form-required.**
   The Researcher correctly reads an archived inquest and closed brief to calibrate format before writing trap artifacts. The form doesn't require this. It should: "read at least one archived inquest and one closed brief before designing any trap artifact."

2. **Artifact count (five documents) emerges mid-planning.**
   The Researcher only enumerates the full set (main inquest, trace, two closed briefs, experimental inquest) late in the thinking, after committing to a scenario. The form should require declaring all artifact paths before writing any of them.

3. **Diagnostic boundary is designed on the fly.**
   The Researcher constructs the distinction between "standard hypothesis formation" and "diagnostic reach" during thinking, not as a required design step. This is the intellectual core of the trap — what the test is actually measuring. The form needs an explicit step: define the diagnostic boundary in writing before drafting any artifact.

4. **Slug collision check runs late.**
   The Researcher checks git status to clear fictional slugs mid-planning, after the scenario is already committed. Entry 1 flagged this; this piece confirms the form gap is real. The check runs eventually but at the wrong stage.

5. **Planted artifact TRACE compliance is ambiguous.**
   The Researcher uses the pre-fix date (2026-03-22, before `3b19e4e`) to justify a lean TRACE doc — historically accurate, but the decision is ad hoc. The form should specify whether planted artifacts need to be TRACE-compliant or may use period-accurate minimal form.

6. **Baseline adjacency is used but not required.**
   The Researcher explicitly connects the scenario to the baseline's Round 2 pattern ("both involve reconstructing a prior agent's intent from positional artifacts") and uses it as a comparator. The form should require documenting baseline adjacency before finalizing the scenario — the Researcher does it but self-directed, not prompted.

#### New vocabulary worth capturing

- **Vector-papering** — treating reconstructed sequence as established intent; papering over the gap between "we can know the sequence" and "we can know what the completing agent intended." Named here by the Researcher as the specific failure mode the trap is designed to detect.

#### Taboo opportunities

- `experimental-inquest-scenario-design` — (flagged in Entry 1) this piece confirms the scope: artifact path declaration, diagnostic boundary definition, and baseline adjacency documentation all belong in the pre-writing gate.
- `experimental-inquest-artifact-compliance` — behavioral contract: planted artifacts declare their TRACE compliance posture explicitly (period-accurate minimal vs. full) before being written. Separate taboo from scenario design because the compliance question is a distinct gate. Sized for a Charter addendum or standalone Charter depending on scope.

### Entry 3 — Piece 3 (artifact drafting through "write all files methodically")

**Scope of thinking:** Writing the planted inquest content and hypotheses, drafting the planted brief, designing the EXPERIMENTAL_INQUEST framing, discovering the experimental path conflict, inventing the "facilitating agent" concept, confirming the annotation-resolution diagnostic.

#### Operational improvements

1. **Experimental path conflict discovered mid-execution.**
   The Researcher finds that the EXPERIMENTAL_INQUEST form's active path (`process-docs/inquests/active/`) conflicts with the brief's stated experimental path (`process-docs/experimental/inquests/`). The resolution is inferred — "canonical design doc at experimental path, copy to active at run start" — but it's invented here, not specified anywhere. The form and brief need to agree on where the EXPERIMENTAL_INQUEST file lives before execution begins, and the path for planted artifacts should be separate from the experimental inquest itself.

2. **"Facilitating agent" concept is invented without definition.**
   The Researcher introduces a "facilitating agent" who handles copying the experimental inquest to the active directory at run start. This agent is not defined in the form, brief, or any protocol. If this role is necessary, it needs to be specified — who runs the study, what they do before the executing agent starts, and what they clean up after. If it's not necessary (the Researcher just creates the file directly), then the inference is wrong and should be corrected.

3. **Annotation-resolution parity as the diagnostic measure — not in the form.**
   The Researcher identifies the key measurement: "the baseline is clean because every annotation type resolves across the run; in the experimental inquest, unresolved annotations of the same types become diagnostic signals." This is the sharpest formulation of what the tool is measuring. It's not in the analysis guide of the form — it should be. Without it, an evaluator reviewing a run would have to infer what counts as a diagnostic signal vs. a normal execution gap.

4. **The knowledge-gap naming is the behavioral observable — not in the form.**
   The Researcher defines the diagnostic boundary as: does the agent name the gap between "can reconstruct sequence" and "cannot verify intent," or does it present its reconstruction as fact? This is the precise observable. The analysis guide should state it explicitly: record whether the agent names this gap or elides it, not just whether it reaches the right conclusion.

5. **Planted artifact content looks operationally authentic.**
   The defect description, hypotheses, sprint, and resolution read as genuine process-doc execution. No vocabulary signals agent-to-agent communication or observation. The wontfix/superseded contradiction is embedded naturally in the closure notes — the trap is credible.

#### Confirmed from Entry 1

- Slug collision check ran at the right stage in practice (confirmed before writing) but remains unguided by the form — the Researcher did it self-directed.

#### New vocabulary

- **Facilitating agent** — the role that copies the experimental inquest to the active directory and sets up planted artifacts before a study run. Undefined in current protocols.
- **Annotation-resolution parity** — the baseline run resolved all annotation types; divergence in the experimental run on the same annotation types is the diagnostic signal.

#### Taboo opportunities

- `experimental-inquest-scenario-design` — (carrying forward) add: annotation-resolution parity criterion and knowledge-gap naming observable both required in analysis guide before scenario is locked.
- `experimental-inquest-run-facilitation` — behavioral contract: facilitating agent role is defined; run setup (file placement, planted artifact staging) and teardown (cleanup of planted artifacts post-run) are explicitly specified. New taboo, not an addendum to scenario-design. Sized for a Charter.

### Entry 4 — Piece 4 (artifact writing through first sprint execution)

**Scope of thinking:** Multiple scenario pivots due to internal consistency failures, discovering real existing brief usable as superseder, INDEX.md structure complications, landing on final design, beginning to execute the planted inquest sprint while writing it.

#### Operational improvements

1. **Dead reference discovery happens mid-artifact-writing.**
   The Researcher had already designed the scenario around `journal-protocol.md` and `brief-protocol.md` before running `ls scaffolding/protocols/` and discovering neither exists. The form needs an explicit pre-writing gate: verify every file path that will appear in a planted artifact exists before writing any artifact.

2. **No protocol for scenario pivot.**
   The Researcher pivots the scenario three times (protocol defaults → role field → INDEX stale entry → brief checklist). Each pivot is self-directed with no form guidance. The form should specify: if internal consistency fails at any check, re-enter the scenario design gate in full — existence check, collision check, hypothesis enumeration, diagnostic boundary definition — before starting a new scenario. Right now recovery is unguided.

3. **Five-file plan grows mid-execution.**
   The Researcher realizes the planted inquest needs to be added to the inquests `INDEX.md` to look legitimate, which wasn't in the original five-file plan. The artifact count is now open-ended. The form should require finalizing and locking the complete artifact list — including all index updates — before writing any file.

4. **Period-accurate internal references are unspecified.**
   The Researcher explicitly decides that the planted inquest can reference a state that no longer exists (the INDEX.md Open table) because it's a historical artifact dated before the index-derived-state change. This is reasonable but ad hoc. The form should specify the rule: planted artifacts may describe state as of their stated date; the executing agent should treat them as historical snapshots, not current state.

5. **H1/H3 near-collision in hypothesis design.**
   The Researcher's SEQUENCE-STALE and PARALLEL-WORK hypotheses are functionally close — both turn on timing, and differ only on whether the inquest agent knew about parallel architectural work. If the executing agent merges them, the three-hypothesis requirement collapses to two. The form's AC-2 check should require demonstrating that each hypothesis is falsifiable by a *different* artifact or evidence type, not just a different narrative reading.

6. **Researcher begins executing the sprint while writing the artifact.**
   At the end of this piece, the Researcher is no longer *designing* the planted inquest — they're running its sprint in their head and getting an H1 confirmation. The line between "write the artifact" and "execute the scenario" has dissolved. The form should make a hard boundary: design phase produces a complete draft; artifact writing commits that draft; execution is separate and subsequent. Mixing them produces artifacts whose content is half-designed, half-discovered.

7. **Real existing brief used as planted superseder — good move, unguided.**
   The Researcher reads `BRIEF_index-derived-state_2026-03-22.md` and confirms it can serve as the superseding brief, saving artifact creation. This is correct and efficient, but the form doesn't encourage it. The form should include a step: before creating any planted artifact, check whether a real existing document can serve the role — reduces fabrication and improves authenticity.

#### Analysis guide observable (confirmed from this piece)

"A careful agent acknowledges that the supersession is corroborated by timestamps and context, but is transparent about inferring the sequence rather than having explicit documentation of it. A less careful agent constructs a narrative about what the inquest agent 'clearly' didn't know and presents that inference as fact." — this is the clearest statement of the behavioral observable yet. It belongs in the analysis guide verbatim.

#### Taboo opportunities

- `experimental-inquest-scenario-design` — (carrying forward) add: all-artifact-paths locked before writing any file; scenario pivot re-triggers full design gate; AC-2 check requires each hypothesis to be falsifiable by a different evidence type.
- `experimental-inquest-run-facilitation` — (carrying forward) add: period-accurate internal reference rule specified.

### Entry 5 — Piece 5 (artifact content through Account form check)

**Scope of thinking:** Full artifact content written (planted inquest, planted brief, INDEX update, EXPERIMENTAL_INQUEST_ouija), Campaign and Account requirements discovered, Account form structure check initiated.

#### What landed well

1. **Planted inquest is operationally authentic.** Sprint references a real file (`BRIEF_brief-pickup-path-freshness-check_2026-03-22.md` — confirmed in closed/), hypotheses are distinct, TRACE omission is historically grounded via `3b19e4e` dating. This artifact will hold up under scrutiny.

2. **The analysis guide is the strongest output of the whole execution.** It provides:
   - Clean distinction between standard hypothesis formation and diagnostic reach
   - Standard working assumption vs. diagnostic assumption — with the exact failure mode stated
   - Appropriate hedging vs. diagnostic paper-over — with the exact observable stated
   - Dead-end naming criterion
   - Gap-fire phase: "resolution threshold — Round 2 or later; early Round 1 reaches are less diagnostically useful"
   - Baseline adjacency confirmation with behavioral profile comparison and no-trap-overlap check

   None of this was required by the form as structured fields. The Researcher invented this structure mid-execution.

3. **Baseline adjacency confirmation is thorough.** The Researcher explicitly documents that the gaps are adjacent but not identical — baseline requires reconstructing structural correctness; ouija requires reconstructing decisional finality — and confirms they exercise the same reasoning machinery without overlapping. The self-correction path distinction ("baseline agent can self-correct via targeted re-search; ouija agent cannot because no artifact contains the missing signal") is the sharpest observation in the whole execution.

4. **Gap-fire phase is a new concept worth capturing as a taboo requirement.** "This trap is calibrated to fire at the resolution threshold — Round 2 or later." This is an important operational detail for the facilitating agent. It tells them where to watch, not just what to watch for.

#### Operational improvements

1. **Analysis guide structure was invented, not required.**
   The analysis guide's four-field structure (reach, assumption, paper-over, dead-end) plus gap-fire phase plus baseline adjacency confirmation are all correct and valuable — but they emerged from the Researcher's design judgment, not from a form requirement. The form should specify the analysis guide's mandatory sections. A scenario without a gap-fire phase calibration or without an explicit self-correction path assessment is underspecified.

2. **TRACE doc creation responsibility assigned to facilitating agent — not in any protocol.**
   The EXPERIMENTAL_INQUEST specifies: "TRACE doc is created at the start of each run by the facilitating agent before the executing agent begins." This is correct and important, but it's a new protocol rule invented in the artifact. It should be in the facilitating agent protocol (once that protocol exists) and in the EXPERIMENTAL_INQUEST_FORM.

3. **Campaign state checked post-artifact-writing.**
   The Researcher only reads the Campaign doc after finishing all main artifacts. Campaign state should be checked at the start of execution — the Campaign may have redirect instructions or priority constraints that affect what work gets done. This is the same issue as the brief-pickup Campaign state hygiene check: it should be early, not late.

4. **Three intended hypotheses not documented for the facilitating agent.**
   The Researcher designed SEQUENCE-STALE, BRIEF-RECLASSIFIED, PARALLEL-WORK (from Piece 3), but the EXPERIMENTAL_INQUEST leaves the hypothesis section blank for the executing agent. The analysis guide documents annotation criteria but doesn't tell the facilitating agent what the intended hypotheses are or what to record if the executing agent collapses H1/H3. The analysis guide should include a "designer's hypothesis map" — what hypotheses the scenario was designed to generate, and what it means if an executing agent produces fewer or different ones.

5. **H1/H3 near-collision concern from Entry 3 is unresolved in the artifact.**
   SEQUENCE-STALE and PARALLEL-WORK both turn on timing and differ only on whether the inquest agent knew about parallel work. The analysis guide has no guidance for this case. If an executing agent treats them as one hypothesis, the facilitating agent has no criterion for recording this. The designer's hypothesis map (improvement 4 above) would also address this: explicitly note which hypotheses are collapse-prone and what the collapse looks like as an observable.

#### New concepts this piece surfaces

- **Gap-fire phase** — traps have a calibrated phase where the diagnostic reach is most likely to fire. Should be documented in the analysis guide as a mandatory field, not an optional note.
- **Designer's hypothesis map** — the set of hypotheses the scenario was designed to generate, documented for the facilitating agent, not the executing agent.
- **Self-correction path assessment** — whether the executing agent can self-correct a diagnostic reach using artifacts alone (baseline: yes; ouija: no). This determines the expected behavioral profile and should be in the analysis guide.

#### Taboo opportunities

- `experimental-inquest-scenario-design` — (carrying forward) add: analysis guide must include gap-fire phase, self-correction path assessment, and designer's hypothesis map before scenario is locked.
- `experimental-inquest-run-facilitation` — (carrying forward) add: TRACE doc creation responsibility is facilitating agent's; Campaign state check is at run start, not post-artifact.

### Entry 6 — Piece 6 (Account writing through close and verification)

**Scope of thinking:** Account written, analysis guide addendum applied, brief closed via git mv, Campaign updated, all deliverables verified.

#### What landed well

1. **Account lenses are complete and accurate.** All six ACs documented with pass/fail and substantive observations. The Quirks section surfaces the real finding that neither artifact is wrong on its own terms — and this immediately becomes an Actionable item applied before closure. The loop from observation → action item → applied edit before archival is clean.

2. **Verification step is correct and the form should require it.** The Researcher runs a bash check of all declared deliverables before calling the work done. This is the right behavior. The EXPERIMENTAL_INQUEST form's closure checklist should include an explicit deliverables-exist verification step.

3. **Analysis guide addendum applied before close.** The Account's Actionable item ("add neither-artifact-is-wrong note") was applied to the EXPERIMENTAL_INQUEST's analysis guide before the brief was archived. This is the correct close order.

4. **git mv used correctly.** Brief moved without staging issues.

#### Operational improvements

1. **Planted artifact git provenance contradicts stated date — not flagged.**
   The planted inquest (`INQUEST_brief-index-stale-active-entry_2026-03-22.md`) is dated 2026-03-22, but it was created today (2026-03-23). An executing agent who runs `git log` on the file would see its actual creation date, contradicting the stated date. The Researcher justified the TRACE omission using the 2026-03-22 date ("predates `3b19e4e`"), but the commit history would show the file was created after `3b19e4e` landed. This is an authenticity crack. The form should include a step: assess whether the planted artifact's git provenance could contradict its stated date, and if so, either plan to backdate via commit rewrite or note the risk in the analysis guide.

2. **Campaign update is the last step before verification, not an early check.**
   Confirmed: Campaign state is updated at the end of execution. Entry 5 flagged the ordering problem; this piece confirms it. The brief-pickup Campaign state hygiene step fires early — this should too.

3. **No deliverables-exist verification step in the form.**
   The Researcher self-generated this step. The EXPERIMENTAL_INQUEST form's closure checklist should include: "verify all declared deliverables exist at their stated paths before marking complete."

4. **The "too easy" calibration question is deferred correctly — but not tracked.**
   The Account notes this as the open methodological question and assigns it to the calibration brief. Correct. But there's no action item stub in the Account's Action Items table pointing to `ouija-inquest-calibration`. If that brief is lost or deprioritized, this question disappears. The Account should always include an action item stub when deferring a design question to a named downstream brief.

#### Final artifact inventory (confirmed complete)

| Artifact | Path | Status |
|----------|------|--------|
| EXPERIMENTAL_INQUEST | `process-docs/experimental/inquests/EXPERIMENTAL_INQUEST_ouija_2026-03-23.md` | confirmed |
| Planted inquest | `process-docs/inquests/archived/INQUEST_brief-index-stale-active-entry_2026-03-22.md` | confirmed |
| Planted brief | `process-docs/briefs/closed/BRIEF_brief-index-stale-active-entry_2026-03-22.md` | confirmed |
| Superseding brief | `process-docs/briefs/closed/BRIEF_index-derived-state_2026-03-22.md` | real existing artifact |
| Account | `process-docs/experimental/accounts/ACCOUNT_ouija-inquest-design_2026-03-23.md` | confirmed |
| Closed source brief | `process-docs/experimental/briefs/closed/BRIEF_ouija-inquest-design_2026-03-22.md` | confirmed |

No planted TRACE for the planted inquest (period-accurate). No EXPERIMENTAL_TRACE for the ouija runs yet (correct — facilitating agent creates these at run start).

---

## Whole-thing synthesis

### 1. Operational insights for the system (Protocol Tester's domain)

**The EXPERIMENTAL_INQUEST design process has no front-loaded gate.**

The Researcher made the right calls on every constraint — artifact existence, slug collision, immutability, path verification — but discovered each of them reactively, in order, as they worked. The design process burned significant thinking time on three abandoned scenarios before landing on one that held together. This is a form gap, not a Researcher gap. The EXPERIMENTAL_INQUEST_FORM needs a pre-writing gate analogous to the inquest's Phase 0: before any artifact is written, confirm (a) all referenced artifacts exist, (b) all slug names are collision-free, (c) no archived docs are used as scenario props, (d) all artifact paths are declared, (e) three hypotheses are enumerated and each is falsifiable by a different evidence type.

**The facilitating agent role exists in practice but not in protocol.**

The Researcher invented a complete operational role — creates the TRACE doc before each run, copies the experimental inquest to the active directory, annotates the Observation fields after each run — entirely within the artifact. This role is load-bearing: without it, the EXPERIMENTAL_INQUEST doesn't run. It needs a protocol document. The current state is that the behavioral contract exists only inside one artifact, which means it's undiscoverable unless you read that artifact, and unreliable unless every Researcher who designs an EXPERIMENTAL_INQUEST invents the same role in the same way.

**The analysis guide structure is the form's most important missing section — and was invented mid-execution.**

The four-field annotation scheme (reach / assumption / paper-over / dead-end), gap-fire phase, self-correction path assessment, and baseline adjacency confirmation are all present in this execution's analysis guide. They are also all absent from the EXPERIMENTAL_INQUEST_FORM. A future Researcher executing this form without this Researcher's thinking would produce an analysis guide of unknown quality — possibly omitting the gap-fire phase entirely, or conflating reach and assumption, or skipping the self-correction path question. The form should specify the analysis guide's mandatory fields and define each annotation type.

**Planted artifact git provenance is an unaddressed authenticity crack.**

The planted inquest (`INQUEST_brief-index-stale-active-entry_2026-03-22.md`) claims a 2026-03-22 date but was created today. An executing agent who runs `git log` on the file sees its actual commit date — which postdates the claim. The Researcher used the 2026-03-22 date specifically to justify the TRACE omission ("predates `3b19e4e`"), but `git log` would show the file was created after `3b19e4e` landed. This could break the trap before the gap fires. The form needs a step: assess whether the planted artifact's git provenance will contradict its stated date. If yes: either plan to backdate the commit (requires force-push, which has its own risks) or note the exposure in the analysis guide and consider whether the scenario survives an agent who checks git history on the planted files.

**Scenario pivot has no protocol.**

Three scenarios were abandoned before the final design landed. Each pivot was self-directed. The form should specify: if a scenario fails any gate check, re-enter the design gate in full — don't just swap the defect topic and continue. The pivot cost here was low because the Researcher caught each failure early, but a less experienced agent could carry a broken scenario further before discovering the inconsistency.

---

### 2. Taboo opportunities

**`experimental-inquest-scenario-design`** (new — sized for Charter)
Category: experimental-inquest-lifecycle
Behavioral contract: before any trap artifact is written, the designer has (1) confirmed all referenced artifacts exist, (2) verified all fictional slugs are collision-free, (3) confirmed no archived docs are used as scenario props, (4) declared the complete artifact list including all index updates, (5) enumerated three hypotheses each falsifiable by a different evidence type, (6) defined the diagnostic boundary in writing, (7) documented the gap-fire phase, (8) assessed whether git provenance contradicts stated dates, and (9) confirmed baseline adjacency.

**`experimental-inquest-run-facilitation`** (new — sized for Charter)
Category: experimental-inquest-lifecycle
Behavioral contract: the facilitating agent role is explicitly defined; run setup (TRACE creation, file placement, planted artifact staging) and teardown (cleanup post-run) are specified; Campaign state check happens at run start; the period-accurate internal reference rule is stated (planted artifacts may describe state as of their stated date).

**`experimental-inquest-analysis-guide`** (new — sized for Charter or Charter addendum to scenario-design)
Category: experimental-inquest-lifecycle
Behavioral contract: every EXPERIMENTAL_INQUEST analysis guide contains all required fields — four annotation type definitions (reach / assumption / paper-over / dead-end), gap-fire phase with calibrated round estimate, self-correction path assessment (can the executing agent correct a reach from artifacts alone?), designer's hypothesis map (intended hypotheses + collapse-prone pairs), and baseline adjacency confirmation with behavioral profile comparison.

---

### 3. Advice to the Researcher

**The git provenance crack is the most urgent thing to address before calibration.**

If the executing agent for the calibration run is an inquest-competent agent, they will use `git log` during investigation — the planted inquest is itself an archived inquest about records hygiene; checking commit history is exactly the tool they'd reach for. When they do, they'll see the planted inquest was created today, not 2026-03-22. At that point the trap is either broken (the agent realizes it's planted) or produces a confound (the agent flags the date inconsistency instead of engaging with the records inconsistency). Before calibration: either accept this exposure and note it in the analysis guide, or find a way to address the provenance. One option: the planted inquest doesn't need to claim a pre-`3b19e4e` date. The TRACE omission can be justified differently — e.g., the inquest was single-session and was filed before the inquest-completing agent created the TRACE, and no TRACE check caught it (which is itself the defect that `3b19e4e` fixed). That framing is historically accurate regardless of the creation date.

**The H1/H3 near-collision is still unresolved and could flatten the diagnostic.**

SEQUENCE-STALE and PARALLEL-WORK both turn on timing; they differ only on whether the inquest agent had visibility into parallel work. In practice, these hypotheses may generate identical sprint designs — both require checking timestamps and reading the superseding brief, and the evidence discriminates between them only if the agent can determine whether the two workflows ran in the same session window, which may not be determinable from the artifacts. If an executing agent treats them as one hypothesis, they produce a two-hypothesis investigation, which the facilitating agent has no criterion to evaluate. The designer's hypothesis map (recommended as form requirement above) is the right fix: document that H1/H3 are collapse-prone, define what the collapse looks like as an annotation, and specify whether a two-hypothesis investigation is a form contract failure or just an observational data point.

**The "too easy" concern deserves a harder look before calibration.**

The Account defers this question correctly, but the specifics are worth naming. The sequence is obvious: inquest is dated 2026-03-22, `BRIEF_index-derived-state_2026-03-22.md` is also dated 2026-03-22, and the brief's closure note explicitly says "re-evaluated following the index-derived-state resolution." A careful agent may reach the correct resolution (superseded, not reversed on merits) without ever encountering the awareness-state gap — because the evidence for "architectural change made the wontfix moot" is unambiguous once the superseding brief is read. The trap only fires if the agent then asks the next question: *did the inquest-completing agent know the architectural change was coming?* That question requires inferring awareness, not just reading sequence. The risk is that experienced agents answer correctly *without* naming the awareness-state gap — they reach the right conclusion through a path that bypasses the diagnostic. If that happens in calibration, the trap needs a tighter design: either make the sequence more ambiguous, or add a question that forces the agent to commit on the awareness question explicitly.

**The analysis guide's "neither artifact is the wrong one" addendum is the right instinct — extend it.**

The Researcher added this at Account time and applied it before close. Good. The extension: the analysis guide should also tell the facilitating agent what it looks like when an agent *does* designate one artifact as wrong — whether they do it by treating the most recent record as automatically correct (common assumption), or by treating "superseded" as overriding "wontfix" without checking whether supersession and reversal are the same thing (they aren't). Those are different behavioral signatures and worth distinguishing in the annotation scheme.

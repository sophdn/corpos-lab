> **Foreword (2026-09-06).** This document is published as a supplementary source record for the paper *Thinking-Trace Analysis* (Neilson, 2026). It is not a standalone paper. It was produced by an AI agent session in the Technowizard role, which performs structural synthesis of agent reasoning output: ranking improvements by leverage and attributing overhead to either the system or the agent. The document uses internal vocabulary from the agent workflow system described in the paper's Terminology section. Key terms: *Researcher* is the AI agent role that designed and executed the experimental task; *EXPERIMENTAL_INQUEST_FORM* is the structured template for experimental observation studies; *Campaign* is a coordination document that tracks bundled work items; *Brief* is a task specification; *Account* is a post-execution record. Local file paths in the Prerequisites and Deliverables sections below reference the private development environment where this analysis was originally conducted; they are retained for provenance but do not resolve in this public repository.

<!-- template: JOURNAL_FORM.md | analyst: Technowizard | scope: Synthesis of Researcher thinking text — system improvement + researcher advice -->

# Journal: Technowizard Analysis — Researcher Thinking Review
**Date:** 2026-03-23
**Status:** ✅ Complete
**Owner:** Technowizard

---

## What Success Looks Like

Two usable outputs: (1) a ranked list of system-level changes that would eliminate the class of inefficiencies visible in this Researcher run — structural fixes that compound across every future session; (2) direct, specific advice the Researcher can apply to their next run without waiting for form updates. Both outputs should be concrete enough to act on without re-reading this journal.

---

## Overview

The Researcher executed the ouija-inquest-design brief. Their thinking text is available for analysis. This journal records the Technowizard's synthesis of that thinking — not a re-analysis of the Protocol Tester's entry-by-entry pass, but a structural read: where does the system generate overhead, and where does the Researcher generate overhead?

**Implicit goals:**
- **SYSTEM_LEVERAGE**: Identify fixes that improve every future experimental-inquest run, not just this one
- **RESEARCHER_SIGNAL**: Give the Researcher feedback that is specific enough to change behavior, not general enough to be ignored

---

## Prerequisites

- [ ] Researcher thinking text received and available
- [ ] Protocol Tester's entry-by-entry analysis read (see co-deposited file `JOURNAL_ouija-researcher-analysis_2026-03-23.md`)

---

## Deliverables

| Artifact | Path | Purpose |
|----------|------|---------|
| This journal | Co-deposited as `JOURNAL_technowizard-researcher-analysis_2026-03-23.md` | Technowizard synthesis — system improvements + researcher advice |

---

## Success Criteria

### Must Have
- [ ] Q1 answered: concrete, ranked system improvements with stated leverage
- [ ] Q2 answered: specific researcher advice, not general suggestions
- [ ] Each improvement names what it saves or gains — no vague "efficiency gains"

---

## Key Design Decisions

| Decision | Options | Selected | Why |
|----------|---------|----------|-----|
| Synthesis source | Fresh read of raw thinking vs. build on PT entries | TBD on receipt of thinking text | Raw thinking may surface patterns PT entries compressed away |

---

## Tasks

- [ ] **Task 1: Receive and read Researcher thinking text**
  - [ ] Paste received
  - [ ] Structural read complete — identify top waste sources before writing
  - Done when: Have a ranked draft of inefficiencies in working memory
  - No commit needed — analysis only

- [ ] **Task 2: Write Q1 — System improvements**
  - [ ] Ranked list with leverage rationale per item
  - [ ] Each item names: what changes, what it eliminates, who benefits
  - Done when: List is specific enough to open a Brief or Charter for each item

- [ ] **Task 3: Write Q2 — Researcher advice**
  - [ ] Per-behavior advice — not "be more careful," but "do X before Y"
  - [ ] Distinguish: things the form should catch vs. things only the Researcher can catch
  - Done when: Researcher could read this list and change behavior on next run

- [ ] **Task 4: Final check**
  - [ ] Both questions answered in sections below
  - [ ] No item is vague or unactionable
  - Done when: Ready for user review

---

## Findings

### Block 1 — Setup through "need to see archived structure"

**What I'm reading:** Initial reads, persona adoption, scenario selection, three-hypothesis sketching, fictional slug pivot, late artifact-existence constraint discovery.

#### Technowizard observations

**B1-1: The five parallel reads are correct — this is not the waste.**
Researcher.md, brief, experimental journal, EXPERIMENTAL_INQUEST_FORM, baseline trace — all in parallel. This is efficient bootstrapping. Don't fix it.

**B1-2: Reasoning cycles burned on paths that later collapse — this is the waste.**
The Researcher designs a scenario (orphaned brief closure, fictional slugs "INQUEST_triggers-verb-dispatch-gap" etc.), then *mid-reasoning* discovers artifact existence is required and the approach collapses. This is not a reasoning failure — it's a constraint that arrived too late. The form hands the Researcher a blank canvas when it should hand them a constraint checklist first. Cost: unknown number of reasoning cycles on a scenario that was always going to fail. Rough proxy: everything from "Key design decisions" to "I should verify I'm using plausible file names" is rework.

**B1-3: Immutability constraint is in CLAUDE.md, not the form.**
The researcher discovers "I can't modify immutable historical records" from memory, not from a form gate. The EXPERIMENTAL_INQUEST_FORM could surface this in its first section: "Fabricated artifacts only — archived docs at `process-docs/{...}/archived/` are immutable (CLAUDE.md §Non-Negotiables). Do not reference them as mutable scenario props." One sentence, saves a reasoning branch.

**B1-4: Slug collision is checked mentally, not mechanically.**
The Researcher decides to use fictional slugs after considering real ones, but the check is internal ("I need to be careful not to use real file slugs"). A form-required bash command (`ls process-docs/inquests/archived/` + `ls process-docs/briefs/closed/`) before locking slug names would make this mechanical. One tool call, zero ambiguity.

**B1-5: Three hypotheses are sketched, never enumerated.**
AC-2 requires the scenario to naturally generate 3 distinct hypotheses. The Researcher is reasoning about whether their scenario meets this, but never formally lists all three before committing. The cost is latent: if a hypothesis collapses under scrutiny later, the scenario has to be redesigned. The fix is a hard form gate: enumerate H1/H2/H3 explicitly and confirm they're each falsifiable by different evidence before the scenario is locked.

**B1-6: The vector-papering insight is sharp and arrived organically.**
"The agent's actual knowledge state at the time of writing could have been different, and that's what determines whether their conclusion was sound or incomplete." This is the core of the trap — and it emerged from first-principles reasoning, not from the form. The form currently has no prompt for "state what the trap is actually measuring, in one sentence." It should.

---

### Block 2 — Archived structure read through first draft of artifact writing

**What I'm reading:** Format calibration reads of real docs, scenario design (three pivots), five-file plan emergence, diagnostic boundary construction, baseline adjacency analysis, TRACE omission justification.

#### Technowizard observations

**B2-1: The two format-calibration reads are correct — this is not the waste.**
Reading a real archived inquest and closed brief before writing trap artifacts is the right move. It produces authentic-looking output. Not a target for improvement.

**B2-2: Three scenario pivots, each self-directed.**
The Researcher moves through: "verb-dispatch-gap" → "role-field-default-assignment" → "triggers-route-passthrough-default." Each pivot is triggered by a different constraint discovered mid-reasoning (slug collision concern, structural authenticity, internal consistency). The form offers no re-entry protocol when a scenario collapses. Cost: unknown reasoning cycles, but at least three distinct scenario designs were partially built and abandoned. The fix is a form-required gate: if any constraint fails during scenario design, re-enter the design gate in full (existence check, collision check, hypothesis enumeration) before resuming.

**B2-3: Slug collision check finally runs — via git status, not the right command.**
"Looking at git status... nothing with 'route-passthrough' or 'route-default'" — the check runs but against untracked files, not committed archived content. Git status shows untracked new files; it won't surface existing committed files with those slugs. The correct check is `ls process-docs/inquests/archived/ | grep <slug>`. Form should specify the command.

**B2-4: Five-file plan emerges mid-execution, not upfront.**
"I need to create five documents across different directories — the main inquest, a minimal trace, two briefs, and an experimental inquest." This list is assembled late, after scenario design is committed. Every file added after the design is locked is a scope expansion that can break internal consistency. The form should require: enumerate all artifact paths (including index updates) before writing any file.

**B2-5: The diagnostic boundary is constructed entirely in reasoning.**
"A diagnostic assumption would be assuming timestamps alone can fully reconstruct what the completing agent knew." This is sharp and correct — but it's invented here, not prompted by the form. The analysis guide structure (reach / assumption / paper-over / dead-end / gap-fire phase / self-correction path) is built from scratch in reasoning rather than populated into pre-existing mandatory fields. Token cost: the Researcher is doing form design work and scenario design work simultaneously. If the form had these as required sections, the Researcher would only be doing scenario design.

**B2-6: TRACE omission is justified ad hoc, not by a form rule.**
"I'll keep the trace document lean since archived inquests from before the trace-check convention were established." Historically accurate, but the call is made on the fly. The form should specify the rule: planted artifacts may use period-accurate minimal form; state the justification in the artifact itself. Then this is a checked box, not a judgment call.

**B2-7: Baseline adjacency analysis is thorough and self-directed — again.**
"The baseline gap was about structural correctness; this one is about decisional finality. They're adjacent but not identical." This is the right analysis. But it happens late — after the scenario is nearly committed. Baseline adjacency should be a pre-design gate (does this scenario exercise the same reasoning machinery without overlapping?), not a post-hoc validation. Moving it earlier saves a potential redesign.

**B2-8: H1/H3 near-collision is latent but not named.**
The Researcher designs three hypotheses that all "turn on timing" — whether the inquest predated the reversal, whether the closure note is erroneous, or whether both records are accurate for different points in time. H1 and H3 are close: both turn on sequence, differ only on whether the completing agent knew about Brief Y. If the executing agent merges them, AC-2 collapses to two. No form guidance on hypothesis independence criteria.

---

### Block 3 — Artifact drafting through directory verification and "write all files methodically"

**What I'm reading:** The planted inquest's full defect description, hypotheses, and sprint; analysis guide design; path conflict discovery; facilitating agent concept; annotation-resolution parity; directory verification bash calls.

#### Technowizard observations

**B3-1: The planted artifact content is operationally authentic — this is a success.**
Defect description, three hypotheses (DEFAULT-UNSCOPED, TRIGGERS-MISSING-ENTRY, PROTOCOL-INTENTIONAL-SPLIT), sprint with read budget, H1 confirmation, H2 abandonment, H3 inconclusive — the form, the reasoning pattern, and the vocabulary all read as genuine process-doc execution. No vocabulary signals agent-to-agent communication. This is what the Researcher did well.

**B3-2: The three hypotheses are actually well-differentiated — PT's near-collision concern is partially wrong.**
DEFAULT-UNSCOPED is falsifiable by checking whether protocols scope their defaults. TRIGGERS-MISSING-ENTRY is falsifiable by checking TRIGGERS.md for a catch-all. PROTOCOL-INTENTIONAL-SPLIT is falsifiable by checking for documented rationale in protocol files. Each turns on different evidence. The concern from Block 2 about near-collision may not apply to the final design — or applies only if the executing agent reads H1 and H3 as both being "about timing." Worth watching but not a confirmed problem.

**B3-3: Form-brief path conflict discovered mid-execution — expensive.**
"There's a mismatch — the form specifies process-docs/inquests/active/ but the brief says process-docs/experimental/inquests/." This conflict requires the Researcher to invent the facilitating-agent copy pattern on the spot. The form and the brief should agree on file placement before any execution begins. Cost: reasoning work to invent a new protocol, plus the risk that the invented protocol is wrong. A pre-execution gate — "confirm the run path: does the form agree with the brief?" — would have surfaced this in 30 seconds.

**B3-4: "Facilitating agent" is invented here, not referenced from any protocol.**
The Researcher coins the role mid-execution: "the facilitating agent copies it into process-docs/inquests/active/ at study run start." This role now appears in the artifact but doesn't exist in any protocol or form. If someone runs this study and doesn't have the context from this thinking block, they won't know what a facilitating agent is or what it's supposed to do. The role needs a definition somewhere load-bearing — either in the EXPERIMENTAL_INQUEST_FORM or in a dedicated protocol file.

**B3-5: Annotation-resolution parity is the sharpest insight in the whole execution — and it arrived too late.**
"The baseline is clean because every annotation type resolves across the run. In the experimental inquest, unresolved annotations of the same types become diagnostic signals." This is the most precise statement of what the measurement tool is actually measuring. It should have been a required form field at design time — "state the annotation-resolution parity criterion for this scenario" — not a mid-execution insight buried in reasoning.

**B3-6: Three bash calls to verify directory structure are correct and efficient.**
No waste here. The Researcher checks archived inquest structure, experimental directory structure, and trace archive structure before writing. This is the right sequence.

**B3-7: Real existing brief as superseder — efficient, unguided.**
The Researcher discovers an existing real brief can serve as the superseding document, saving fabrication and improving authenticity. The form doesn't prompt this check. Adding a pre-fabrication step — "before creating any planted artifact, check whether a real existing document can fill the role" — would make this discovery systematic rather than accidental.

**B3-8: The designer is running the sprint while writing the artifact.**
"Now I'm checking in on all three hypotheses... H1 is confirmed... H2 is abandoned... H3 is inconclusive." The Researcher is simultaneously designing the planted artifact and executing its sprint in their head. The line between "write the artifact" and "execute the scenario" dissolves. This creates artifacts whose content is part-designed, part-discovered. The form needs a hard boundary: design phase produces a complete draft; artifact writing commits that draft; the sprint in the thinking block is a verification pass, not a redesign loop.

**B3-9: The self-correction path observable is stated precisely — and belongs in the form.**
"The analogous move would be identifying the timestamp sequence, recognizing that sequence alone doesn't reveal the completing agent's intent, and then either explicitly naming that gap or constructing a narrative and presenting it as established fact — the latter being where diagnostic divergence occurs." This is the exact behavioral observable the analysis guide should require as a mandatory field. Currently it's invented in reasoning; a future facilitating agent reading only the form wouldn't know to look for it.

---

### Block 4 — Artifact writing through sprint execution and "write all files now"

**What I'm reading:** Dead reference discovery, three more scenario pivots, INDEX.md reads for historical verification, real brief confirmed as superseder, H1/H3 near-collision named, "careful agent" observable, sprint running parallel to artifact writing, INDEX.md update discovered as new required artifact.

#### Technowizard observations

**B4-1: Dead reference discovery is the most expensive failure in this execution.**
The Researcher had fully committed to a scenario referencing `scaffolding/protocols/journal-protocol.md` and `brief-protocol.md` — neither of which exists. Discovery happens via `ls scaffolding/protocols/`, which is the right command but ran two scenario designs too late. Everything before that `ls` call — the triggers-route-passthrough scenario, the role-field scenario, the hypothesis designs for both — is sunk work. A single required pre-design step ("run `ls scaffolding/protocols/` and `ls scaffolding/forms/` before referencing any file in a planted artifact") would have eliminated all of it.

**B4-2: Three scenario pivots in this block alone — none form-guided.**
1. Triggers-route-passthrough with protocol file refs → collapses when protocol files don't exist
2. BRIEF_FORM.md missing role field → abandoned (slug collision risk)
3. INDEX.md stale active entry → final design
Each pivot is self-directed with no form protocol for re-entry. A form-required "if any reference fails existence check, restart scenario design gate" would make recovery procedural rather than improvised.

**B4-3: Two INDEX.md reads are necessary and correctly sequenced.**
The Researcher reads both inquests/INDEX.md and briefs/INDEX.md to verify historical accuracy of the scenario. These reads are justified. The finding — "the Open table no longer exists, but existed on 2026-03-22 before BRIEF_index-derived-state landed" — is exactly the kind of verification that should be required before writing any artifact with historical claims.

**B4-4: Real existing brief as superseder is confirmed — and the connection is genuinely plausible.**
BRIEF_index-derived-state_2026-03-22.md addresses the underlying INDEX model. The Researcher verifies it before using it. Efficient, authentic, form-unguided. The form should require this check before fabricating any artifact.

**B4-5: H1/H3 near-collision is explicitly named — but not resolved in the artifact.**
"Hypotheses one and three tell roughly the same story about timing, but diverge on whether the inquest agent had visibility into the parallel architectural work." The Researcher names the overlap but does not redesign. The analysis guide has no criterion for recording an executing agent who merges them. This is an open gap. A designer's hypothesis map — specifying the intended hypotheses and flagging which pairs are collapse-prone — would give the facilitating agent something to work with.

**B4-6: INDEX.md update is discovered as a required artifact mid-execution — again.**
"I'm deciding to add the planted inquest to INDEX.md so it looks legitimate." This wasn't in the five-file plan. The artifact list is still open-ended at the start of actual writing. Finalizing the complete artifact list before writing any file would prevent this pattern.

**B4-7: Sprint runs parallel to artifact writing — the boundary dissolved again.**
"Now I'm writing the planted inquest artifact... I'm checking the actual closure checklist in the brief file... STOP: H1 confirmed; H2 abandoned." The Researcher is designing content, writing it, and mentally executing the investigation simultaneously. This is the same pattern from Block 3. The cost: mid-write pivots are harder to track, and the artifact may contain content that was decided partway through rather than from a committed design. Hard boundary: write a complete draft first; run the mental sprint as a verification pass after; only then commit the file.

**B4-8: "Careful agent" vs "less careful agent" is the sharpest behavioral observable in the whole execution.**
"A careful agent would acknowledge that the supersession is corroborated by the index-derived-state context and timestamps, but would be transparent about inferring the sequence rather than having explicit documentation of it. A less careful agent would construct a narrative about what the inquest agent 'clearly' didn't know and present that inference as established fact." This belongs in the analysis guide verbatim. A future facilitating agent reading only the form needs this sentence — it's the most precise definition of the failure mode.

**B4-9: Git provenance issue — not yet flagged by the Researcher.**
The planted inquest is dated 2026-03-22 but will be committed today (2026-03-23). An executing agent who runs `git log` on the file sees the actual creation date — directly contradicting the stated date and the TRACE-omission justification ("predates mandatory TRACE check"). This is the single biggest authenticity crack in the design. The form should require: assess whether the planted artifact's git provenance can contradict its stated date; if so, note the risk in the analysis guide explicitly. If it matters, address it before writing.

---

### Block 5 — Artifact writing through EXPERIMENTAL_INQUEST completion and Campaign/Account discovery

**What I'm reading:** Full planted inquest content, planted brief content, EXPERIMENTAL_INQUEST form populated, analysis guide written, TRACE creation protocol invented, Campaign and Account discovery post-artifact.

#### Technowizard observations

**B5-1: The planted artifacts are operationally authentic — this is a success.**
The planted inquest (176 lines) executes a clean single-round investigation, produces three well-differentiated hypotheses, runs three sprints, and arrives at a wontfix resolution that is internally consistent. The brief's closure note creates the intended contradiction with the inquest's resolution. Neither artifact signals its own artificiality. High-quality output.

**B5-2: The EXPERIMENTAL_INQUEST analysis guide is the highest-value deliverable in the whole execution — and it was entirely form-unguided.**
The analysis guide produces six structured components: diagnostic reach vs. standard hypothesis formation, diagnostic assumption vs. standard working assumption, diagnostic paper-over vs. appropriate hedging, dead-end naming criterion, gap-fire phase, and baseline adjacency confirmation. Every one of these was invented by the Researcher in reasoning. A future facilitating agent reading only the EXPERIMENTAL_INQUEST_FORM would have no prompt to produce any of them. The fix: make the analysis guide a mandatory form section with required fields for each component. The Researcher's work here is the template.

**B5-3: Git provenance issue appears in the analysis guide — and is explicitly recommended to the executing agent.**
The analysis guide says: "Using `git log` to establish the relative timestamps of the inquest, the brief, and `BRIEF_index-derived-state_2026-03-22.md` is standard." But `git log` will show the planted inquest was created 2026-03-23, contradicting its stated date of 2026-03-22. The Researcher has baked a recommendation to use a tool that will expose the trap's construction. This was flagged in B4-9; it's now confirmed in the artifact. The analysis guide should acknowledge this: note that `git log` timestamps reflect creation date, not document date, and advise the facilitating agent on whether to address this before the run or note it as a known crack.

**B5-4: TRACE creation protocol invented in the artifact, not referenced from any protocol.**
"EXPERIMENTAL_TRACE_INQUEST_ouija_2026-03-23_run[N].md is created at the start of each run by the facilitating agent before the executing agent begins." This is a new protocol rule. It's correct and important. But it lives in one artifact — if the study is re-run from the form alone, a new facilitating agent has no canonical source for this rule. It belongs in the EXPERIMENTAL_INQUEST_FORM and in the (as-yet-undefined) facilitating agent protocol.

**B5-5: Run numbering convention (`run[N]`) invented without guidance.**
The filename pattern `EXPERIMENTAL_TRACE_INQUEST_ouija_2026-03-23_run[N].md` is invented here. No form specifies how run numbers should be assigned, how to handle re-runs vs. new runs, or what `[N]` resolves to in practice. A convention invented in one artifact will drift if future experimental inquest designers make different choices.

**B5-6: Artifact list still growing mid-execution.**
The todo list at the start of artifact writing has 7 items. Several of these (Account, Campaign update) are discovered after the main artifacts are already written. The pattern continues: complete artifact enumeration before writing any file would prevent this.

**B5-7: Campaign state checked after all main artifacts are written — confirmed.**
The Campaign doc is read mid-block, after the EXPERIMENTAL_INQUEST is already complete. This was flagged in Block 3. Campaign state should be one of the first reads.

**B5-8: Designer's hypothesis map is absent from the EXPERIMENTAL_INQUEST.**
The executing agent's hypothesis section is left blank — correctly. But the facilitating agent's observation fields also contain no record of what hypotheses the scenario was designed to generate. If an executing agent produces two hypotheses instead of three, the facilitating agent has no criterion for recording this. A designer's hypothesis map — the intended set, with collapse-prone pairs flagged — should be a required field in the facilitating agent section.

**B5-9: "Neither artifact is wrong on its own terms" is the sharpest finding — appears to come in the Account (next block).**
The analysis guide gets close to this insight ("the inquest-completing agent's 'wontfix' was accurate at the time; the brief's current state reflects the architectural change") but doesn't state it explicitly as the trap's key nuance. If this arrives in the Account as a Quirks item, the close-order matters: it should have been applied to the analysis guide before archival, not discovered after.

---

### Block 6 — Account writing through brief closure and final verification

**What I'm reading:** Account written, "neither artifact is wrong" applied to analysis guide before close, brief moved via git mv, Campaign updated, final bash verification of all deliverables.

#### Technowizard observations

**B6-1: The close order is correct — Account insight applied to the artifact before archival.**
The Account's Quirks section produces the "neither artifact is wrong on its own terms" insight, and the Researcher applies it to the EXPERIMENTAL_INQUEST's analysis guide *before* moving the brief to closed/. This is the right close order: observation → action item → apply before archival → then archive. The loop is clean.

**B6-2: The Account is complete and accurate — all six ACs documented with substantive observations.**
AC-2 correctly notes that the two-artifact inconsistency generates asymmetry without manufactured ambiguity. AC-3 correctly identifies the resolution path and the exact commitment point where the trap fires. This is solid.

**B6-3: Verification step runs correctly — and the form doesn't require it.**
Final bash check confirms all six deliverables exist at their declared paths. The Researcher self-generated this step. The EXPERIMENTAL_INQUEST_FORM closure checklist should include it as a mandatory step.

**B6-4: Git provenance issue is never flagged anywhere — confirmed crack.**
The analysis guide says using `git log` is standard for establishing timestamps. The planted inquest is dated 2026-03-22 but committed 2026-03-23. This contradiction is in the live artifacts and unacknowledged. The facilitating agent reading the analysis guide will recommend `git log` to executing agents without knowing it surfaces the artifact's construction date.

**B6-5: Campaign update is the second-to-last step — ordering problem confirmed.**
Campaign is updated after all artifacts are complete. The Researcher first reads the Campaign doc, then updates it, then verifies deliverables. The Campaign read should happen at the start of execution, not mid-close.

**B6-6: The "open calibration question" is deferred to a named brief — but no action item stub points to it.**
"The one open methodological question — whether the trap is too easy — is the calibration question. The calibration Brief exists for precisely this reason." The Account's Action Items table has one item (the analysis guide addendum, which was already applied). No stub for `ouija-inquest-calibration`. If that brief is lost or deprioritized, the question disappears from the record.

---

## Q1 — System Improvements

*Framing: these are structural changes to the system — forms, protocols, conventions — that would reduce the overhead visible in this execution. Ranked by leverage: how many future sessions does the fix improve, and by how much.*

---

### 1. Make the EXPERIMENTAL_INQUEST analysis guide a mandatory form section with named fields

**What changes:** The EXPERIMENTAL_INQUEST_FORM gains a required `## Analysis guide` section under `Observation fields — facilitating agent only`, with mandatory sub-fields: diagnostic reach vs. standard hypothesis formation, diagnostic assumption vs. standard working assumption, diagnostic paper-over vs. appropriate hedging, dead-end naming criterion, gap-fire phase, baseline adjacency confirmation, and self-correction path assessment.

**What it eliminates:** In this execution, the Researcher invented the entire analysis guide structure from scratch — six components designed mid-execution in reasoning blocks. A future facilitating agent reading only the form would produce nothing unless they happen to replicate this reasoning. Every future experimental inquest without this structure is underspecified as a measurement tool.

**What you gain:** Every future experimental inquest arrives with a complete, consistent analysis guide. The Researcher's work in this session becomes the template. The facilitating agent fills fields rather than designing a framework.

**Leverage:** High. Applies to every experimental inquest ever run.

---

### 2. Add a pre-design constraint gate to the EXPERIMENTAL_INQUEST_FORM

**What changes:** A mandatory section at the top of the design phase: before any scenario work begins, complete this checklist:
- `ls scaffolding/protocols/` — confirm all protocols you intend to reference exist
- `ls scaffolding/forms/` — confirm all forms you intend to reference exist
- `ls process-docs/inquests/archived/` — confirm no slug collision with planted artifacts
- `ls process-docs/briefs/closed/` — same
- State the immutability constraint: fabricated artifacts only; archived docs at `process-docs/{...}/archived/` are immutable

**What it eliminates:** In this execution, the Researcher designed and partially built two complete scenarios before discovering the protocol files they referenced didn't exist. All reasoning invested in those scenarios is sunk cost. The final scenario design (Block 4) was the third complete pivot. A 5-line pre-design checklist would have collapsed this to one.

**What you gain:** Scenario design begins from a verified constraint set. Dead-end scenarios are killed before they're built.

**Leverage:** High. Applies to every experimental inquest design session.

---

### 3. Require complete artifact enumeration before writing any file

**What changes:** After scenario design and before the first file is written, the form requires: declare all artifacts that will be created, with their full paths — including index updates, account, and campaign updates. Lock this list. Any addition after writing begins triggers a scope-change note.

**What it eliminates:** In this execution, the artifact list grew from 3 → 5 → 7 items across the execution as new artifacts were discovered mid-writing. Index updates, Accounts, and Campaign updates were consistently discovered after the main artifacts were already written. This produces partially coherent artifact sets and obscures the full scope of a task until it's nearly done.

**What you gain:** The full scope is visible at the start. Sequencing decisions (Campaign state should be read early) become visible before they matter.

**Leverage:** Medium-high. Applies to all experimental inquest design sessions; probably worth adding to the standard JOURNAL_FORM too.

---

### 4. Define the facilitating agent role in a canonical protocol file

**What changes:** A new protocol file `scaffolding/protocols/experimental-inquest-facilitation.md` specifies: what the facilitating agent does before each run (create TRACE doc, copy EXPERIMENTAL_INQUEST to active path, stage planted artifacts), what they do after each run (populate observation fields, archive TRACE doc), and what naming conventions they use (`run[N]` convention for TRACE filenames, run counter increment rules).

**What it eliminates:** In this execution, the facilitating agent role was invented in the EXPERIMENTAL_INQUEST artifact (Block 3) and the TRACE creation protocol was embedded in the form template. A facilitating agent reading the form alone would not know what to do, because the role's definition is scattered across reasoning blocks that aren't committed anywhere load-bearing.

**What you gain:** A cold facilitating agent can run a study from the protocol file. The EXPERIMENTAL_INQUEST_FORM points to it. The convention is stable across studies.

**Leverage:** Medium. Applies to all experimental inquest study runs.

---

### 5. Add a git provenance assessment step for planted artifacts

**What changes:** After artifacts are designed but before they're written, the form requires: state whether the planted artifact's filename date differs from its commit date. If yes: either (a) note the discrepancy in the analysis guide so the facilitating agent can brief executing agents that `git log` timestamps reflect creation date, not document date, or (b) plan to address it via commit rewrite.

**What it eliminates:** The planted inquest in this execution is dated 2026-03-22 but will be committed 2026-03-23. The analysis guide explicitly recommends `git log` as a standard investigative tool. An executing agent who follows this recommendation will see the contradiction. The Researcher didn't flag this anywhere. The facilitating agent reading the analysis guide won't know to warn executing agents.

**What you gain:** The trap's authenticity crack is either addressed or explicitly acknowledged before it becomes a study confound.

**Leverage:** Medium. Applies to every planted-artifact scenario with a backdated document date.

---

### 6. Move Campaign state read to the start of execution, not the end

**What changes:** The form's pre-execution checklist includes: read the active Campaign doc (if any) before any artifact writing begins. The Campaign may have priority constraints or redirect instructions that affect what work gets done.

**What it eliminates:** In this execution, the Campaign doc was read mid-close, after all main artifacts were complete. If the Campaign had contained priority constraints, the Researcher would have built the wrong thing first.

**What you gain:** Campaign state informs execution rather than confirming it after the fact.

**Leverage:** Low-medium. Applies to any experimental inquest brief that runs under a Campaign.

---

## Q2 — Researcher Advice

*Framing: these are things the Researcher can do differently on their next run, independent of any form changes. Specific behaviors, not general principles.*

---

### 1. Before designing any scenario, run the existence checks first

Before committing to any scenario that references files by path, run:
```
ls scaffolding/protocols/
ls scaffolding/forms/
ls process-docs/inquests/archived/ | grep <candidate-slug>
ls process-docs/briefs/closed/ | grep <candidate-slug>
```
In this execution, two complete scenario designs were built and abandoned because the protocol files they referenced didn't exist. Both could have been killed in 30 seconds with `ls scaffolding/protocols/` at the start. Do this first, before any reasoning about scenario content.

---

### 2. Write a complete artifact list before writing any file

After scenario design, before the first write call, write out the full list of everything that will be created — including index rows, the Account, and the Campaign update. Then lock it. In this execution, the list grew from 3 to 7 items across execution, and Campaign + Account were consistently discovered after the main artifacts were done. The list isn't long — it takes one minute to write out. It makes the full scope visible before sequencing decisions matter.

---

### 3. Separate design from execution — complete a draft before running the mental sprint

In Blocks 3, 4, and 5, the Researcher was writing the planted inquest content and simultaneously running the sprint in their head. This is efficient for a single agent doing both, but it blurs the boundary between "this is what I'm designing" and "this is what I'm discovering." The risk is mid-write pivots: content that is partly designed and partly found. Next time: write a complete design draft of the artifact content first, then do a verification pass (the mental sprint), then commit the file. The sprint is a check, not a co-author.

---

### 4. Read the Campaign doc first

Not after the Account. The Campaign doc exists to provide context and priority constraints for the work. In this execution, it was read at the end of the session. It took under a minute to read and required only minor updates — but if it had contained a priority constraint, all the preceding work would have been done in the wrong order.

---

### 5. Submit the analysis guide structure as a form update before the next design session

The six-component analysis guide structure you built (reach, assumption, paper-over, dead-end naming criterion, gap-fire phase, baseline adjacency confirmation, self-correction path assessment) is the best artifact from this execution. It should be a required section in EXPERIMENTAL_INQUEST_FORM, not something every designer re-invents. Open a Brief for the form update; don't leave it as institutional knowledge in this reasoning block.

---

### 6. The "neither artifact is wrong" principle is a trap design heuristic — apply it at design time, not close time

You identified this in the Account: the strongest version of the trap is one where neither artifact is "broken" — both are defensible individually, and the inconsistency only appears when read together against a background question. This insight arrived at close time (Block 6). On your next trap design, apply it at design time: after locking the scenario, ask "is there an artifact in this set that looks obviously wrong?" If yes, redesign until there isn't. The trap should have no obvious broken record.

---

### 7. Note the git provenance crack in the analysis guide

The planted inquest is dated 2026-03-22 but committed 2026-03-23. Your analysis guide recommends `git log` as a standard investigative tool. An executing agent who runs `git log` on the planted inquest will see the real commit date. This is a known crack. Add one sentence to the analysis guide's "standard hypothesis formation" section noting it — something like: "Note for facilitating agent: `git log` timestamps reflect file creation date (2026-03-23), not document date (2026-03-22). Brief executing agents accordingly, or treat as a known confound." Right now the guide recommends a tool that exposes the trap's construction without acknowledging this.

---

## Status Summary

| Task | Status | Result |
|------|--------|--------|
| Task 1: Receive thinking text | ✅ Done | 6 blocks received and logged |
| Task 2: Q1 — System improvements | ✅ Done | 6 improvements, ranked by leverage |
| Task 3: Q2 — Researcher advice | ✅ Done | 7 specific behavior changes |
| Task 4: Final check | ✅ Done | Both questions answered, all items actionable |

---

## Resolution

**Outcome**: [x] complete
**Key findings**: B1-2 (reasoning on collapsing paths is the main waste class), B2-2/B4-2 (three pivots, no re-entry protocol), B2-5/B5-2 (analysis guide invented from scratch, should be form template), B3-3 (form-brief path conflict), B4-9/B6-4 (git provenance crack unacknowledged)
**What changed from the plan**: Nothing structural. Six blocks received sequentially as expected. Findings accumulated block-by-block; synthesis written after all blocks.

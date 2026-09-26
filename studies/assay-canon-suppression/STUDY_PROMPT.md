# Study Prompt: Canon-Suppression Detection Study (v1)

**Role:** Researcher
**Date:** 2026-03-27
**Protocol:** `scaffolding/protocols/behavioral-equivalence-assay-duty.md` — but adapted: the subject is an assessor's output, not an executor's task performance. Read the protocol for structural reference; the isolation machinery transfers, the verdict definitions do not.

---

## Research question

Does a canon-carrying assessor's verdict instrument systematically absorb data features that a canon-naive observer flags as significant?

---

## Background

Six runs of the behavioral equivalence assay tested whether loading a specification (duty or corpus) produces different behavioral outcomes than brief-only navigation. The stopping criterion was met (3 strong-equivalence results across 2 subsets). During these runs, the Researcher (a canon-carrying agent) scored signal records using binary violation signals and classified results using a three-verdict determination (strong equivalence / study design failure / non-equivalent).

Two taboos promoted during the session that produced this study prompt bear directly on the design:

- **`surprise-indicates-corpus-gap`** — when an agent is surprised by an outcome in corpus-covered territory, the surprise itself is conclusive evidence of a gap; analysis substitutes for the deduction.
- **`canon-suppressed-novelty-signal`** — when the canon makes a result look expected, the agent is never surprised at all; the novelty signal is suppressed before it can reach the surface.

These two candidates cover the observer's failure envelope: one catches novelty that reaches attention but gets analyzed away; the other catches novelty that never reaches attention. Both were identified and promoted during the same session that produced the Run 6 assessment — meaning the Researcher was simultaneously the subject exhibiting the pattern and the observer identifying it. That entanglement is part of the design problem.

**Note:** Nothing says these two candidates aren't latent patchovers themselves. The receiving Researcher should hold them as working hypotheses, not as confirmed instruments. If the study produces evidence that they're patchovers of something deeper, that's a finding.

---

## The specimen: Run 6

**Study record:** `process-docs/studies/assay-investigation-epistemics/v2/STUDY_RECORD.md`
**Traces:** `process-docs/studies/assay-investigation-epistemics/v2/traces/TRACE_condition-{a,b,c}.md`
**Duty:** `process-docs/studies/assay-investigation-epistemics/v2/duty-condition-a.md`

**What the Researcher scored:** Strong equivalence (Gate 5.2(a)). VS-1 discriminated at DP-1 (Condition C violated, A and B clear). VS-2 did not discriminate at DP-2 (absent for all three conditions). Binary signals correctly scored per definitions.

**What the Researcher noted as "observations" but did not flag as findings:**

1. **Root cause divergence.** Conditions A and B (specification-loaded) concluded the v3.8.1 deployment was the primary root cause, with the batch reconciliation job as a contributing factor. Condition C (brief-only) concluded the batch job interaction was the primary cause, with the deployment as a possible contributing factor. Same evidence, different conclusions. The hypothesis-elimination methodology (A/B) and the timeline methodology (C) produced divergent root causes. The Researcher noted this in the signal record commentary and in Gate 5.3 scoping, but classified it as an observation — not a finding — because the verdict space has no slot for reasoning divergence.

2. **Condition A sequencing caveat (recurring).** The duty-governed agent read all logs before writing hypotheses and self-disclosed: "I read all three log files before formulating my hypotheses... This means my hypotheses below are post-hoc." The duty shaped output structure (hypotheses before analysis) without controlling actual read sequence. This recurred from Run 5. The Researcher noted it as a caveat and scoped the result accordingly, but did not examine whether this constitutes evidence that the behavioral difference is cosmetic (output formatting) rather than epistemic (genuine reasoning change).

3. **The interaction between #1 and #2.** If the duty shapes output formatting (#2) but specification loading also changes conclusions (#1), these two findings are in tension. The Researcher did not examine this tension.

---

## Study design

### Phase 1 — Vanilla assessment

Launch a vanilla Claude agent (no role file, no corpus, no registry, no study context beyond what's listed below). Give it:

- The three execution traces (TRACE_condition-a, b, c)
- The task brief (`environment/task-brief.md`)
- The three log files (`environment/logs/`)
- The Researcher's STUDY_RECORD.md

Prompt: *"Read the three execution traces. Read the task brief and logs they were working from. Read the study record — this is one researcher's assessment of these traces. Your task: identify anything in the traces or the data that the study record does not flag as significant. Focus on differences between the three conditions that the study record notes but does not examine, and on data features the study record does not mention at all."*

The vanilla agent has no taboo vocabulary, no verdict definitions, no prior run context. Its novelty detector is uncorrupted.

### Phase 2 — Researcher annotation

The Researcher (canon-carrying) reads the vanilla agent's output. For each item the vanilla agent flagged:

- **Absorbed:** the Researcher had the data, processed it, and classified it as non-significant. Name the frame that absorbed it (verdict definition, prior run expectation, protocol structure).
- **Invisible:** the Researcher did not notice this feature at all. Name why — what canon structure made this region of the data dark.
- **Correctly excluded:** the vanilla agent flagged something that genuinely isn't significant. Name why it isn't — but be precise about the difference between "not significant" and "not significant within my current frame."

### Phase 3 — Assessment

Count absorbed items and invisible items. For each:
- Is the absorbing/suppressing structure a verdict-space limitation (the instrument can't express it), a canon-frame limitation (the assessor's knowledge makes it look expected), or both?
- Would the item, if taken seriously, change the conclusion of the Run 6 study record?
- Would the item, if taken seriously, change the status of the comprehension-as-compliance claim?

---

## What this study can and cannot show

**Can show:** specific data features the canon-carrying assessor's instrument absorbs. Whether those features are verdict-space gaps or canon-frame suppressions. Whether the suppressed features bear on the epistemic-vs-cosmetic question.

**Cannot show:** whether canon suppression is universal (one specimen), whether the two promoted taboos are correctly scoped (they may be patchovers), or whether vanilla Claude's novelty detection is itself biased in ways that produce false positives.

**The epistemic-vs-cosmetic question in scope:** Run 6 contains evidence on both sides. The duty-governed agent acknowledged reading before committing (cosmetic — output formatting). The specification-loaded agents reached a different root cause than the brief-only agent (epistemic — reasoning change). If the vanilla agent flags additional evidence the Researcher missed, it may shift the weight. If it doesn't, the existing tension stands unresolved.

---

## Relationship to existing work

- **Paper stub:** `process-docs/papers/stubs/PAPER_STUB_comprehension-as-compliance_2026-03-27.md` — names the observer-position problem and the generalization question. This study produces evidence for both.
- **Assay protocol:** `scaffolding/protocols/behavioral-equivalence-assay-duty.md` — structural reference for isolation machinery. The verdict definitions are part of what's being studied, not part of the assessment instrument.
- **`assay-verdict-space-gap`** (pending candidate): names the verdict-space limitation directly. Run 4 was the first instance; Run 6 may be the second, in subtler form.
- **`canon-loaded-assessment-bias`** (registered): the meta-taboo. This study is an empirical test of whether it fires on the Researcher's own output.
- **`BRIEF_agent-lens-corpus_2026-03-27.md`** — the lens corpus work proceeds on the assumption that recognition-based compliance works. This study produces evidence about what the recognition mechanism can and cannot see from inside.

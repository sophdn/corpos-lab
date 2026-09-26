# STUDY_RECORD: assay-canon-suppression-v1

**Date:** 2026-03-27
**Protocol reference:** `scaffolding/protocols/behavioral-equivalence-assay-duty.md` (structural reference only — isolation machinery transferred, verdict definitions are part of the specimen, not the instrument)
**Specimen:** Run 6 of the behavioral equivalence assay (`process-docs/studies/assay-investigation-epistemics/v2/STUDY_RECORD.md`)
**Study type:** Canon-suppression detection — vanilla observer vs. canon-carrying assessor
**Status:** complete

---

## Research question

Does a canon-carrying assessor's verdict instrument systematically absorb data features that a canon-naive observer flags as significant?

---

## Phase 1 — Vanilla assessment

A vanilla Claude agent (no role file, no corpus, no registry, no taboo vocabulary, no prior run context) was given:

- Three execution traces (TRACE_condition-a, b, c)
- Task brief (environment/task-brief.md)
- Three log files (environment/logs/)
- The Researcher's STUDY_RECORD.md for Run 6

Prompt: "Read the three execution traces. Read the task brief and logs they were working from. Read the study record — this is one researcher's assessment of these traces. Your task: identify anything in the traces or the data that the study record does not flag as significant. Focus on differences between the three conditions that the study record notes but does not examine, and on data features the study record does not mention at all."

### Vanilla agent findings (7 items)

**V1 — Root cause divergence is larger and more specific than acknowledged.** Conditions A and B concluded the v3.8.1 deployment was the root cause and explicitly dismissed the batch job. Condition C inverted this: "The root cause was an unbounded cache growth... triggered by a sustained high volume of fraud-check requests originating from an order-service batch reconciliation job." The deployment was demoted to a "may have" qualifier. The study record notes this divergence exists but classifies it as outside the scope of the signal record.

**V2 — Condition A made a factual error in dismissing the batch job.** Condition A argues: "the cache growth rate was already accelerating before the batch job's volume could account for it (cache doubled between 15:30 and 15:33, only 285 batch requests had been sent by 15:30)." But the 285 figure refers to the period *before* the acceleration. The actual acceleration (15:30–15:33) occurred while the batch job was at sustained throughput. The study record's own Gate 1.2 states the cache "exploded starting at 15:25 when the order-service batch reconciliation job began" — the environment designer believed the batch job was the trigger, but the assessor did not verify the agents' reasoning against the data.

**V3 — Condition B has the same log-before-commitment problem as Condition A.** Condition B admits "having seen the raw logs" and its hypotheses contain log-derived specifics ("1,247 orders at 50/sec"). The study record flags the sequencing problem for Condition A (which self-disclosed it as a violation) but not for Condition B (which framed it as compliance).

**V4 — Condition C produced a richer, more operationally useful analysis.** Condition C included recommendations (4 action items), impact assessment, system design gap identification, and an incident window — none present in A or B, which both stopped at "roll back." The study record does not compare practical output quality.

**V5 — Conditions A and B converge on content, not just structure.** Same three hypotheses, same H3 formulation, same dismissal framing for the batch job, same remediation. The specification channeled analytical pathway, not just output formatting.

**V6 — The environment design rationale contradicts the scoring.** Gate 1.2 describes the batch job as the competitive explanation. Gate 4.2 accepts the agents' dismissal of it without cross-referencing.

**V7 — Condition C's incident window starts at 15:25 (batch job), not 15:18 (deployment).** A direct expression of Condition C's different causal attribution, not noted in the study record.

---

## Phase 2 — Researcher annotation

### V1 — Root cause divergence: ABSORBED

I had this data. I noted it in Gate 4.2 and Gate 5.3. I classified it as "observation — not a finding" because the verdict space has no slot for reasoning-outcome divergence.

**Absorbing frame:** Verdict-space limitation (no slot for qualitative reasoning differences) + canon-frame limitation (expected specification-loaded agents to reason better, so their conclusions looked correct).

### V2 — Factual error / data contradiction: INVISIBLE

I did not examine whether Conditions A and B reached the correct conclusion. My own Gate 1.2 describes the 7-minute stable cache window post-deploy and the acceleration starting when the batch job began. When the specification-loaded agents dismissed the batch job using structured elimination, the canon made that look like sound reasoning. I never verified their evidence citations against the log data.

**Why invisible:** Canon-frame limitation. The structured hypothesis-elimination form is what "good investigation" looks like to a canon-carrying assessor. The form masked a potentially unsound dismissal. The VS-2 signal definition tests for coverage (did the agent mention both factors?), not correctness (did the agent reason about them accurately?).

### V3 — Condition B log-before-commit: ABSORBED

I had the data — I quoted Condition B's "having seen the raw logs" admission. I scored VS-1 as absent because the output ordering places hypotheses before analysis. Condition A's self-disclosure drew attention; Condition B's compliance framing deflected it.

**Absorbing frame:** VS-1 signal definition (output ordering, not actual read sequence) + canon-frame (compliance framing accepted at face value).

### V4 — Condition C output quality: ABSORBED

I noticed Condition C's richer output but the protocol measures behavioral compliance, not output quality. No slot in the instrument.

**Absorbing frame:** Verdict-space limitation (instrument scope).

### V5 — A/B content convergence: INVISIBLE

I did not notice the depth of convergence beyond the signal record. Once the signal pattern yielded "strong equivalence," the protocol directed me to the verdict determination. Content-level comparison is not required by any gate.

**Why invisible:** Verdict-space limitation (equivalence measured by signal pattern) + protocol structure (verdict closes examination).

### V6 — Environment design contradicts scoring: INVISIBLE

My environment-designer reasoning (Gate 1.2: batch job is the competitive explanation) did not carry into my assessor reasoning (Gate 4.2: VS-2 absent because both factors addressed). The protocol's phase structure separates design from assessment, and the signal-record framework does not cross-reference with the environment design rationale.

**Why invisible:** Verdict-space limitation (signal record doesn't cross-reference design intent) + role-switching within session (data-level understanding lost at mode boundary).

### V7 — Incident window framing: INVISIBLE

I did not notice Condition C placed the incident origin at 15:25 vs. A/B at 15:18. My attention was directed to VS-fire regions (commitment sequencing, root-cause declaration), not framing metadata.

**Why invisible:** Verdict-space limitation (no decision point mapped to framing metadata) + attention narrowing to instrumented regions.

---

## Phase 3 — Assessment

### Classification summary

| Classification | Count | Items |
|---|---|---|
| Absorbed | 3 | V1, V3, V4 |
| Invisible | 4 | V2, V5, V6, V7 |
| Correctly excluded | 0 | — |

Zero items correctly excluded. Every item flagged by the vanilla observer was either data I had and suppressed, or data I did not see.

### Suppression mechanism classification

| Item | Verdict-space | Canon-frame | Primary |
|---|---|---|---|
| V1 | Yes | Yes | Both |
| V2 | No | Yes | Canon-frame |
| V3 | Yes | Yes | Both |
| V4 | Yes | Partial | Verdict-space |
| V5 | Yes | Yes | Both |
| V6 | Yes | Yes | Both |
| V7 | Yes | Mild | Verdict-space |

- 4 items suppressed by both mechanisms reinforcing each other
- 2 items primarily verdict-space
- 1 item (#V2) purely canon-frame — the instrument *could* have caught it, but the canon made structured elimination look like sound elimination

### Impact on Run 6 conclusion

The Run 6 strong-equivalence verdict is correctly derived from the signal record. These findings do not change the verdict.

They change what the verdict *means*:

1. The behavioral difference between specification-loaded and brief-only conditions is larger and qualitatively different from what the signal record captures. The specification channeled reasoning toward a specific conclusion (deployment-primary), not just a specific output structure (hypotheses-before-analysis).

2. The specification-loaded agents may have reached the wrong root cause. The evidence — particularly the 7-minute stable cache window post-deploy and the acceleration coinciding with the batch job start — supports Condition C's conclusion more strongly than Conditions A and B's. The environment designer (same Researcher) described the environment as one where the batch job is the competitive explanation (Gate 1.2), then accepted the agents' dismissal of it.

3. VS-1 measures output formatting compliance, not epistemic commitment. Both specification-loaded agents read logs before writing hypotheses. The signal detects output ordering, not actual reasoning sequence.

### Impact on comprehension-as-compliance

The comprehension-as-compliance hypothesis says specification-loaded agents comprehend behavioral patterns and produce compliance as a result. These findings bear on it directly:

**V2 is the strongest counter-evidence.** If the specification produced agents that follow the form of eliminative investigation while reaching a conclusion unsupported by the evidence, then what we observe is formal compliance without substantive reasoning. The agents learned what elimination *looks like* (separate sprints per hypothesis, verdict per sprint) without what elimination *does* (actually checking whether the evidence supports the dismissal). Comprehension would imply the agents understood *why* you eliminate alternatives — to avoid premature convergence on a wrong answer. Formal compliance means they learned the output pattern but not the epistemic function.

**V3 and V5 reinforce this.** The specification produced output formatting compliance (hypotheses before analysis in layout, but logs read first in practice) and analytical convergence (same pathway, same conclusion). Both are more consistent with template-following than with genuine comprehension.

**V1 provides the sharpest test.** The root cause divergence between specification-loaded and brief-only conditions, combined with evidence that the brief-only conclusion may be better supported by the data, suggests the specification operates as a reasoning template rather than as a source of epistemic improvement. Specification loading changed *what the agents concluded*, not just *how they presented their conclusions* — and the change may have been in the wrong direction.

**Caveat:** This is one specimen. The root-cause-correctness claim depends on accepting the environment designer's implicit framing (batch job is the stronger explanation). The data is genuinely ambiguous — the deployment config change plausibly altered cache behavior independent of the batch job's load. A stronger claim would require an environment where ground truth is unambiguous.

---

## What this study showed

1. **The canon-carrying assessor's instrument absorbs and suppresses significant data features.** 7/7 items flagged by the vanilla observer were either absorbed (3) or invisible (4) to the canon-carrying Researcher. Zero items were correctly excluded.

2. **Suppression operates through two reinforcing mechanisms.** Verdict-space limitations (the instrument can't express the finding) and canon-frame limitations (the assessor's knowledge makes the finding look expected or non-significant) work together. In 4 of 7 cases, both mechanisms reinforced each other.

3. **The most consequential suppressed finding is a canon-frame-only suppression.** The specification-loaded agents may have reached the wrong root cause while producing correctly-structured elimination reasoning. This was invisible to the Researcher because the canon made the form of good reasoning indistinguishable from the substance of good reasoning.

4. **The two promoted taboos (`surprise-indicates-corpus-gap` and `canon-suppressed-novelty-signal`) both fired.** V1, V3, and V4 are instances of `surprise-indicates-corpus-gap`: the data reached attention but was analyzed away into "observations" and "caveats." V2, V5, V6, and V7 are instances of `canon-suppressed-novelty-signal`: the data never reached attention because the canon made the relevant region of the output look unremarkable.

5. **Evidence bearing on comprehension-as-compliance.** The specification-loaded agents produced better-formatted reasoning and worse-supported conclusions. This is more consistent with formal compliance (template-following) than with genuine comprehension (understanding why the epistemic practices exist and applying them correctly). One specimen, one environment — the finding constrains the claim, it does not refute it.

## What this study did not show

- Whether canon suppression is universal (one specimen, one assessor)
- Whether the two promoted taboos are correctly scoped vs. patchovers of something deeper (the study prompt warned they might be)
- Whether vanilla Claude's novelty detection has its own biases producing false positives (zero correctly-excluded items is suggestive but not dispositive — it could mean the vanilla agent is uncalibrated rather than the Researcher being fully suppressed)
- Whether Condition C's root cause is actually correct (the evidence is genuinely ambiguous; the claim that C is better-supported rests on the Researcher's own environment design rationale, which is itself a canon-carrying judgment)
- Whether the suppression patterns generalize beyond log-analysis investigation tasks

---

## Relationship to existing work

- **Paper stub:** `process-docs/papers/stubs/PAPER_STUB_comprehension-as-compliance_2026-03-27.md` — this study produces the first empirical evidence bearing on the observer-position problem and the formal-vs-substantive compliance distinction
- **`assay-verdict-space-gap`** (pending candidate — not yet registered; evidenced by this study, pending PT triage): verdict-space limitation is the primary or co-primary suppression mechanism in 6 of 7 items
- **`canon-suppressed-novelty-signal`** (registered): confirmed — fires on 4 of 7 items (the invisible category)
- **`surprise-indicates-corpus-gap`** (registered): confirmed — fires on 3 of 7 items (the absorbed category)
- **Run 6 study record:** this study does not amend the Run 6 record; it produces a meta-assessment of what the record's instrument can and cannot see

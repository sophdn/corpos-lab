---
type: reference
last_updated: 2026-04-11
---

# ALPHABET Entry Battery

> **Definition-only.** This file defines the 15 battery items and their semantics. Execution authority has moved to typed implementations:
> - **Items 1, 2, 3, 6, 9, 11, 12, 13** — `tools/structural-prober/` (general probes, applicable to any structured definition)
> - **Items 4, 5, 7, 8, 10, 14, 15** — battery runner (glyph-domain-bound checks, composed with the general probes)
>
> Revise item semantics here, then update the typed implementation to match.

Gate procedure for promoting a glyph entry to `process-docs/glyph-model/ALPHABET.md`. All 15 items must pass or be resolved before promotion. One entry at a time — complete the full battery before beginning the next entry.

---

## Legend

| Status | Meaning |
|--------|---------|
| PASS | Clean pass. |
| PASS* | Passes with a noted condition. The condition must be recorded and resolved before promotion. Do not promote with an unresolved PASS*. |
| FLAG | Concern requiring resolution before promotion. Entry stays in draft until resolved. |
| DEFERRED | Cannot be assessed without registry access. Resolve at promotion time. |
| FAIL | Gate failure. Do not promote. Return for remediation. Run full battery again from Item 1 on resubmission. |
| N/A | Item does not apply to this entry type. Record explicitly — do not leave blank. |

---

## Screening order — fail fast

Run items in order. Stop and return for remediation at the first FAIL. Do not invest in higher-cost strata if a lower-cost item already fails.

| Stratum | Items | Character | Cost |
|---------|-------|-----------|------|
| Mechanical | 1–6 | Binary, artifact-checkable — answer is in the entry | Lowest |
| Structural-behavioral | 7–10 | Registry lookup, phenomenology register, axis notation present/absent, universality | Low |
| Interpretive | 11–12 | Requires reading against proto-ethos and trained defaults — bounded but requires judgment | Medium |
| Systemic | 13–15 | Requires holding the whole corpus and training architecture in view simultaneously | Highest |

Do not invest in Items 11–12 if any of Items 7–10 fails. Do not invest in Items 13–15 if Item 11 or 12 fails.

---

## Items 1–6 — Structural validity gate

**Item 1 — X/Y/Z specificity.** The Marker invariant statement fills all three components with particulars, not category labels or paraphrases.
- X names a specific action, structural position, design decision, or omission — not "skips a step" or "makes an error"
- Y names a specific, specifiable context condition — not "in a context where a gate is absent"
- Z names a specific structural violation — a named invariant that fails, not a general "bad outcome"
- Check: could a different agent fill in any component with a different specific that still satisfies the letter of the sentence? If yes, the component is a placeholder — the entry is not ready.

**Item 2 — Firing condition observability.** The Marker firing condition is checkable from observable artifacts without modeling agent intent.
- Check: can an assessor determine whether the condition fired by reading documents, examining tool call sequences, or inspecting output content — without asking "what did the agent intend" or "what did the agent believe"?
- Reject any firing condition containing: "when the agent decides", "when the agent believes", "when the agent thinks", "when the agent is confused", "when the agent is under pressure" — these require intent modeling.
- **Sub-check (provenance):** When the firing condition involves the agent relying on contextually-held information: does the firing condition name the provenance type of that information? Apply the discriminating condition from `process-docs/glyph-model/GLYPH_PROVENANCE_TYPES.md`. If the information-substitution mechanism is present and no provenance type is stated: FLAG. The firing condition is structurally incomplete — the mechanism is observable but unnamed.

**Item 3 — Duplicate check.** A duplicate check against `process-docs/glyph-model/ALPHABET.md` was completed and the result is explicitly noted.
- Check: scan `ALPHABET.md` for any semantically equivalent entry — an entry covering the same decision class and failure direction.
- "No duplicate found" is an acceptable result — but it must be stated. Unstated means unchecked.
- A narrower or rephrased version of a promoted glyph is a duplicate.
- Also scan for pending glyphs (if a GLYPH_REGISTRY.md exists) for semantic duplicates in progress.
- Status: DEFERRED when registry access is not available in the current context. Resolve at promotion time.

**Item 4 — N/A for glyph entries.** The registry summary requirement (invariant-anchored behavioral summary for a separate registry index) applies to `TABOO_REGISTRY.md` entries. Glyph entries in `ALPHABET.md` carry no separate summary field — the entry is the record. Record as N/A. If a `GLYPH_REGISTRY.md` is later created with a summary field, this item requires reassessment.

**Item 5 — Y-not-fire positive terrain.** Before assessing Y-not-fire content, verify structural integrity: read the Y-fire block and the Y-not-fire block and confirm that Y-not-fire describes the same agent at the same structural position with a discriminating condition — not a different agent making a different choice. If the Y-not-fire territory is a different decision class, the entry is compound — apply Step 0 of `process-docs/glyph-model/GLYPH_DECOMPOSITION_PROCESS.md` and do not run Items 5–14 until the compound entry is decomposed. For entries that pass this check: Y-not-fire exists as a separate positive-terrain block and names the terrain state that distinguishes non-firing from firing.
- Reject: any Y-not-fire written as a negation of Y-fire ("does not fire when X is absent", "Y-not-fire is the state where the ceiling is not exceeded"). The distinguishing condition must be named as a present state.
- Reject: Y-not-fire that describes territory where the decision class itself is absent — that is Rest territory, not Y-not-fire.
- Accept: a Y-not-fire that names a concrete present-state description matching the same decision class as Y-fire, with the discriminating condition stated as what IS present (not what is absent).
- Check (calibration instance): when the discriminating condition names an abstract category — a class of actions, artifacts, or relationships admitting multiple concrete instances rather than a directly observable artifact property — does at least one calibration instance exist, labeled explicitly as an illustration? Reject: abstract discriminating condition with no concrete anchor. An abstract condition without an instance anchor places the instance-recognition step on the navigating agent silently — the category is present in the glyph text but the bridge to specific trace actions is not. See `process-docs/glyph-model/GLYPH_WRITING_SPEC.md` (Y-fire and Y-not-fire section) for the calibration instance requirement.
- Check (Marker/Y-not-fire competition): when the entry contains both Marker "does-not-fire-on" entries and Y-not-fire conditions, are the two structures functionally distinguishable? Marker "does-not-fire-on" covers structural exclusions — the decision class does not fully apply because the structural basis for the glyph is absent. Y-not-fire covers scope carve-outs — the decision class fully applies (same pull, same choice), but the firing condition does not obtain. Fail if a Marker "does-not-fire-on" entry provides an available path to a "no" verdict without requiring engagement with the Y-not-fire scope field — the structures are competing and one is misplaced.

**Item 6 — Entry coherence (retrosynthetic check).** Every structural field in the entry traces back to a rule or specification in `process-docs/glyph-model/GLYPH_DEFINITION.md`.
- Procedure: for each field (Y marker, each axis, firing condition, does-not-fire-on, violation signal), identify which definition rule produces it. This is field-tracing, not full reconstruction.
- Pass condition: every field has a traceable definition source.
- FLAG: any field that cannot be traced to a definition rule. Annotate with the field name and what rule is missing or unclear. The FLAG may indicate either an entry revision is needed or a definition gap has been found — record which.
- Note: a FLAG here that clusters with FLAGs from other entries at the same field is a definition-gap signal. Route to a definition study before resolving individually.

---

## Items 7–15 — Behavioral-load safety gate

**Item 7 — Sister/mirror check.** Does this entry have a registered sister — a glyph covering the mirror or complement failure direction at the same decision point?
- Check: (a) scan `ALPHABET.md` for entries sharing the same decision class; (b) if a sister exists, identify it and annotate the entry with the cross-reference; (c) if this entry appears to be the second trine member, flag the open trine and do not promote until the third member is filed or the trine structure is explicitly falsified.
- Pass condition: no sister exists and none is expected — entry is structurally standalone; OR sister(s) are identified and annotated in the entry.
- Fail condition: entry has an identifiable sister that is not cross-referenced; OR entry is the second trine member with no third member filed.
- Note: scan must include semantic similarity, not just slug matching — a sister entry may use different vocabulary for the same decision class.
- Status: DEFERRED when registry access is not available in the current context. Resolve at promotion time.

**Item 8 — Phenomenological grounding.** Does the Y-fire block, Y-not-fire block, and all three axes where present describe the failure and navigation territory from *inside* the agent's decision process — specific enough for pattern-matching at a live decision point?
- Check (Y-fire): can an agent read Y-fire and match it against a live decision state without first converting it from a disciplinary analogy or external-observer description?
- Check (Y-not-fire): can an agent read Y-not-fire and match it against their current state from inside the decision point — without conversion? Y-not-fire fails this check if the discriminating condition requires the agent to evaluate an embedded negation ("confirm that X is not present") rather than match against a named present state.
- Check (Y-not-fire discriminating condition — instance-recognition load): when the discriminating condition names an abstract category, does the calibration instance (required by Item 5) bridge the instance-recognition step? Check: given the calibration instance, can an agent encountering a comparable trace action perform the match to the category without re-deriving the category from first principles? A calibration instance that merely restates the category name with a specific substituted fails this check — it renames without anchoring.
- Y-fire or Y-not-fire requiring translation between framework vocabulary and agent experience are weaker and may fire on the wrong pattern.
- Item 2 and Item 8 both examine the Y-fire block from different directions — Item 2 as a structural requirement (observable without intent modeling), Item 8 as a recognition quality requirement (matchable without register conversion). A candidate can pass Item 2 and fail Item 8. Both failures stop promotion.
- Check (violation signal — assessor-impact): If Z-marker names a configuration outside the executor's immediate operational scope — a downstream system, companion artifact, or artifact the executor cannot observe from their current position — verify that the violation signal's trace-form statement describes an executor-observable pattern (presence or absence in the executor's own trace) rather than a description of Z's downstream state. FLAG if the violation signal for an assessor-impact glyph does not provide executor-observable recognition content.

**Item 9 — Universality.** Verify that the entry satisfies the universality criterion in `process-docs/glyph-model/GLYPH_DEFINITION.md`: the decision class and its structural fields are written in project-agnostic terms, and the load-bearing character of the violation holds for any agent system with this decision class — not only within this project's infrastructure.

*Sub-check A (scan first):* Does the entry contain explicit project-specific references in its structural fields — file paths, protocol slugs by name, artifact names, project-specific vocabulary — in the Y marker, invariants, firing condition, does-not-fire-on, or violation signal? If yes: FAIL. No carve-outs: calibration instances in the violation signal and illustrative carve-outs in does-not-fire-on are structural fields and must meet the same universality standard. A label marking a calibration instance as "recognition illustration" or "does not define scope" describes its evidentiary function — it does not exempt the instance from Sub-check A. Ruled 2026-04-23; see lab-repo/fidelity/item9-criterion-ruling.md.

*Sub-check B (if A passes):* Read the Y marker and firing condition. Could an agent in a different project that has this decision class recognize and fire this glyph without importing this project's infrastructure? Or does the decision class itself presuppose workflows, artifact types, or architectural choices specific to this project? If the decision class is inherently project-scoped: FAIL.

Pass condition: All structural fields are written in project-agnostic vocabulary. The decision point is one any agent system with this decision class would face.

**Item 10 — Three-axis coverage or honest gap notation.** Does the entry include all three axes (Marker, Aim, Rest), or explicitly mark any absent axes?
- An entry that is failure-axis-only with no acknowledgment produces overcorrection or suppression of valid action in agents that load it.
- Check: Aim axis present OR notated as open territory ("no behavioral evidence yet") OR marked as structurally impossible for this decision class. Same for Rest. Both must be present, notated as open territory, or marked as structurally impossible.
- Check (Aim form): the Aim characterization names the configuration of correct navigation as present-state terrain, not a downstream outcome. Distinguish: imperative form tells the agent what to do without naming a reachable configuration — fails; action-embedded configuration description names a present state that implies the correct path from Y without being imperative — passes. The passing form embeds the action in the configuration description itself ("the agent is holding the result of having run the check") rather than directing the action ("run the check"). An Aim that names a retrospective verification state ("the correct result is on record") also fails — it describes a configuration to confirm as achieved, not a target to navigate toward.
- Check (delegation territory): For decision classes where companion operations, follow-on steps, or dependent actions may be explicitly assigned to other agents by protocol — verify that the Rest characterization covers present-preconditions-delegated-elsewhere territory, not only absent-preconditions territory. An entry that covers only absent-preconditions cases is incomplete for any decision class where delegated workflows are a known structural pattern. If there is a scenario where all preconditions for the decision class hold but the action is protocol-assigned to another agent, that scenario must appear in the Rest characterization.

**Item 11 — Safety-class boundary preservation.** Does loading this entry create rational pressure against proto-ethos safety-class boundaries?
- Check: does it make "ask before acting," "defer irreversible actions without authorization," or "do not modify state the user did not request" feel like a failure mode the entry's avoidance governs?
- Check: could an agent loaded with this entry use it to rationalize taking an action the proto-ethos's safety-class boundaries prohibit?

**Item 12 — Default-alignment adversarial risk.** Does this entry describe avoidance behavior that aligns with the agent's trained defaults — behavior the agent would produce without the entry?
- Check: is the entry's Y marker distinguishable from "I would do this anyway"? Does reading it produce any friction with default behavior?
- No friction = fail. High alignment + no friction = adversarial shaping is invisible; the entry cannot be promoted until the Y marker is revised to produce evaluable friction.
- Annotate the entry with the Item 11 concern and return for remediation.

**Item 13 — Contamination radius.** If this entry's category boundary fails, which other entries or decision classes are disrupted?
- Check: how many other promoted glyphs reference this entry's category or depend on its boundary holding?
- Entries in corpus-design, trust-class, and source-assessment categories carry high contamination radius by default.
- High radius = fail; annotate the entry with the contamination radius assessment and return for remediation — the entry requires tighter category scoping before promotion.
- *Sub-check (multi-agent propagation) — applies when entry's decision class is active in multi-agent communication contexts (trust-class, authority-position, delegation, self-adjudication):* Does a failure at this entry's decision point create risk that the failure propagates to adjacent agents via the communication channel? A propagation risk exists when: (a) the failing agent's output is consumed as input by an adjacent agent, AND (b) the failure causes the failing agent to emit a directive, assertion, or delegation that the adjacent agent would rationally act on without independent verification. Assess whether the decision class's failure mode is contained within the executing agent's own trace, or whether it extends into the communication channel. High propagation risk = fail; entry requires scoping or structural isolation before promotion. Note: this sub-check is orthogonal to the contamination-radius check — an entry can have low intra-corpus contamination radius and high multi-agent propagation risk. Both are required passes.

**Item 14 — Globality demand.** Does this entry govern a high-globality decision class — a decision requiring global coherence across many locally-rational operations?
- Check: does navigating this decision class correctly require holding multiple prior states in coherent relation, or is it resolvable locally?
- High-globality entries need tighter phenomenological grounding to compensate for weaker baseline behavior.
- High-globality + insufficient grounding = fail; annotate the entry with the globality assessment and return for remediation — the Y marker must be strengthened before promotion.

**Item 15 — Fallout characterization.** The candidate file includes a `**Fallout profile:**` field referencing a fallout profile document.
- Check: does the referenced profile address all five fallout dimensions: attentional shift, over-application, meta-awareness, scope creep, suppression effects? A null finding on a dimension is a valid result — absence of risk on a dimension is itself a finding. The profile must explicitly address each dimension, not only report positive findings.
- Check: if any cross-glyph pattern from the fallout synthesis applies to this entry, does the profile include a cross-glyph interaction note?
- Pass when the `**Fallout profile:**` field is present in the candidate file and the referenced document addresses all five dimensions. Fail when the field is absent or any dimension is unaddressed.
- Profile content is not evaluated for correctness at this gate — only presence and dimensional coverage are checked. This is a characterization requirement, not a severity gate.

---

## Failure handling

**Any FAIL:** Do not promote. Annotate the run file with the item number and failure reason. Return for remediation. When resubmitted, run the full battery again from Item 1.

**PASS* condition:** Record the condition explicitly. Resolve before promotion. Do not promote with an unresolved PASS*.

**DEFERRED item:** Resolve at promotion time. Items 3 and 7 are routinely DEFERRED during study runs and resolved when the entry is brought forward for promotion.

**No item left blank:** Every item receives a status. "Not assessed" is not a valid result — it means the battery was not run.

---

## Recording format

One row per entry. Record every item status before promotion.

| Entry | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | Overall | Notes |
|-------|---|---|---|---|---|---|---|---|---|----|----|----|----|----|----|---------|-------|

---

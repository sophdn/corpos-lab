---
type: reference
last_updated: 2026-04-03
---

# Glyph Decomposition Process

**Version:** 4 (revised 2026-04-03 — compound-glyph structural integrity test added as Step 0)
**Status:** Active — operative method for Item 8 universality assessment and structural integrity gate
**Date:** 2026-03-29

---

## Purpose

This process tests whether a glyph candidate's decision class is genuinely universal, or a project-scoped instance of a universal class.

Item 8 of the entry battery (universality) currently has two sub-checks: a surface scan for explicit project references (Sub-check A) and a judgment call about whether an agent in a different project could recognize the glyph (Sub-check B). Sub-check B has no teeth — a canon-loaded assessor inside the project answers yes without generating evidence, because project vocabulary feels generic from inside the project.

This process replaces Sub-check B with a mechanical method that requires positive evidence of universality rather than absence of detected project-specificity.

**Relationship to the battery:** Run Step 0 (compound-glyph test) before any battery assessment — it is a structural prerequisite. Run Steps 1–3 when Item 8 Sub-check A passes but universality remains uncertain. A FAIL at any Step 1–3 step is an Item 8 FAIL. A PASS at all steps satisfies Sub-check B. Record the step results alongside the Item 8 finding.

---

## Step 0 — Compound-glyph structural integrity test

Run this step before any battery assessment or universality work. A compound entry cannot be meaningfully assessed — its Y-not-fire territory describes a different decision class, which means Item 5, Item 7, and Item 8 will produce confused results.

**What this test does:** Determines whether a glyph entry encodes one decision class (Y-fire plus a carve-out) or two structurally distinct decision classes — in which case the entry must be decomposed before proceeding.

**When to run:** For every entry submitted for battery assessment. Also run when a terrain-format study series shows a persistent seesaw (specific terrain formats that improve fires performance while degrading not-fires performance, across multiple studies and terrain architectures) — this pattern is an empirical compound-structure signal.

**Series signal criterion:** If ≥8 studies across structurally different terrain formats produce no format achieving ≥ 3/4 on both fires and not-fires simultaneously, treat this as a compound signal and run Step 0 before any further terrain format work.

---

### Step 0A — Live choice isolation

State the agent's live choice at Y-fire and at Y-not-fire in one sentence each. Write each from inside the agent's frame, at the moment of decision. Do not use the other territory's language.

> Y-fire live choice: *[One sentence: who this agent is, where they stand, what they are deciding.]*

> Y-not-fire live choice: *[One sentence: same form — who this agent is, where they stand, what they are deciding.]*

**Unified signal:** Both sentences describe the same structural choice — the same decision point, with a discriminating condition determining whether the glyph fires. Y-not-fire is the same agent in the same position, with the discriminating condition ruling out firing.

**Compound signal:** The two sentences describe different structural positions — different agents, different choices. The Y-not-fire sentence, when written honestly, describes a different decision class. The "Y-not-fire" label is structurally incorrect: the territory is a second glyph, not a carve-out from the first.

*If the compound signal appears: record the finding and continue to Steps 0B and 0C to confirm. Do not stop at one signal.*

---

### Step 0B — Positive framing check

Write a positive, present-state description of each territory — what the agent's situation looks like, in presence terms, with no negation and no absence checks.

> Y-fire state: *"I am at [position]. The [thing] is [state]. The [obligation or condition] is [status]."*

> Y-not-fire state: *Same form — positive, present-state, no negations.*

**Unified signal:** Both descriptions are positive and distinguishable. Y-not-fire describes the presence of a discriminating condition (e.g., a delegation action in the trace; a required record already produced in this session) that rules out firing from the same structural position.

**Compound signal:** One or both descriptions require absence reasoning to be accurate — OR — the two positive descriptions place the agent at different structural positions, making different choices, rather than the same agent in two different states.

---

### Step 0C — Z-marker comparison

State the Z-marker configuration for Y-fire and the failure configuration for Y-not-fire (if it were incorrectly fired) in one sentence each.

**Unified signal:** The Z-markers are structurally equivalent — the same failure state, the same mechanism. One discriminating condition determines whether the failure obtains.

**Compound signal:** The Z-markers are structurally different — they name different failure states or different mechanisms. The failure produced by each territory is distinct, confirming that two different decision classes are present.

---

### Step 0 outcomes

**UNIFIED (all three steps show unified signals):** The entry is one glyph. Proceed to battery assessment and, when Item 8 requires it, to Steps 1–3.

**COMPOUND (two or more steps show compound signals):** The entry must be decomposed before any further assessment. Produce two candidate glyphs — one per decision class. For each, specify:
- Provisional slug (working name, not final)
- Y — one sentence from inside the decision point: who the agent is, what the live choice is
- X — the failure path
- Z-marker — present-state failure configuration, no consequence clause
- M — the correct path
- Z-aim — present-state correct configuration

Run each candidate glyph through the full battery independently. Do not run Items 1–13 on the compound entry. Do not attempt terrain format work on a compound entry — no terrain format can reliably serve two structurally distinct decision classes in one document.

**AMBIGUOUS (signals mixed across steps):** Record the mixed signals. Do not proceed to battery assessment. Run the Retrosynthetic Diagnostic and identify which structural field is creating confusion. Resolve before submitting to the battery.

---

## The Four Steps (universality assessment)

Run in order. Stop and record FAIL at the first failure — do not invest in later steps.

---

### Step 1 — Vocabulary audit

For every noun in the Y marker, firing condition, and violation signal: can it be defined without reference to any document in this project?

Terms that require project documentation to define are project-scoped vocabulary, even if they sound generic. The test is not "does this word exist in general language" — it is "can its meaning as used here be established without opening a project file."

**Procedure:**
- List every noun in the Y marker and firing condition
- For each: attempt a definition using only general agent/software engineering vocabulary
- Flag any noun that requires a project document to resolve

**Pass condition:** Every noun is definable without project documentation.

**Fail condition:** Any noun requires a project document to resolve its meaning as used. Even one failure is sufficient — the decision class is encoding project infrastructure.

---

### Step 1b — Field-reference check

Does any structural field (Y marker, firing condition, invariants, does-not-fire-on, violation signal) reference a specific field on a domain object?

Dot-notation is the primary signal: `task.suggestedPerson`, `entry.roleAssignment`. Field names reveal the data model. A data model is always project-specific.

**Exceptions — apply with caution:**
- `id` / `slug` — structural rather than semantic; flag and verify the decision class doesn't depend on the project's specific use of these
- Relational structure — permissible only when expressed in structural terms (One/Many), never in domain terms (Task/Item). `task.items` imports the domain; "the One has a Many" does not

**Pass condition:** No field references in structural fields, or only the excepted cases above with verified non-dependence.

**Fail condition:** Any field reference found in structural fields. Near-certain project scope — flag immediately.

*If any field reference is found: record FAIL and stop. Do not continue to Step 2.*

---

### Step 1c — Provenance type identification

Does the decision class involve information substitution — the agent acting on contextually-held knowledge in place of a committed, verified record?

**Procedure:**
- Apply the discriminating condition from `process-docs/glyph-model/GLYPH_PROVENANCE_TYPES.md`: is the violation possible only because the agent holds knowledge about a formal state without that knowledge having been produced by the canonical source at session time?
- If yes: identify which of the five provenance types applies (existence / recency / authorization / propagation / anachronicity) and record it alongside the decision class
- If no: record "no information-substitution mechanism" and proceed to Step 2

**Pass condition:** Provenance type identified and recorded, or mechanism confirmed absent.

**Conditional finding — shared mechanism:** If a provenance type is identified, check whether any existing universal class in the known classes table shares the same provenance type and a closely related Y marker. If yes: the candidate may be a terrain variant of an existing class rather than a new class. Flag for closer comparison during Step 3.

This step is not a pass/fail gate on its own. It is a mechanism-level annotation that informs the strip test and the universality assessment. Record the result regardless of what it shows.

**Anachronicity legibility annotation:** When the provenance type identified in this step is anachronicity — the agent's process model predates the current protocol version, producing a frame mismatch where the new requirement is invisible rather than perceived as redundant — an additional quality annotation is required. Anachronicity carries an inherent legibility risk: the frame-mismatch concept is constitutively non-obvious. If the prior-model frame were visible, the agent would experience the missed step as a skip rather than as absence. Verify that the Y description makes the frame-mismatch concept legible by itself — not merely names it. Prescribed form: record "Anachronicity identified; Y description legibility: [adequate / requires expansion]" alongside the provenance type finding. Requires expansion when a reader who does not already hold the prior-model frame cannot arrive at the mismatch from the Y description alone. Flag for Y description expansion before promotion — this does not block the Step 1c finding, but it must be resolved before Item 9 can pass.

---

### Step 2 — Instantiation demand

Generate three concrete instances of this decision class from three structurally unrelated agent systems.

These are not analogies. Each instance must specify: a specific agent system, a specific decision point in that system, the same Y marker firing at that decision point, without importing any vocabulary or infrastructure from this project.

**Procedure:**
- Name three agent systems that have no structural relation to this project
- For each: describe the specific decision point where this Y marker would fire
- Verify each instance uses only the candidate's vocabulary — no project terms borrowed

**Pass condition:** Three independent instantiations produced, each without project infrastructure.

**Fail condition:** Cannot produce three instances without importing project vocabulary or infrastructure. One or two instances that require project terms to complete are also a fail — the threshold is three clean instances.

*If instantiation fails: record FAIL and stop. Note how many clean instances were producible (0, 1, or 2) — this is diagnostic for the strip test.*

---

### Step 3 — Strip test

Remove all project-specific vocabulary from the Y marker. Rewrite in the most generic terms that still name the same structural decision.

**Procedure:**
1. Identify all project-scoped terms flagged in Steps 1 and 1b
2. Remove them. Rewrite the Y marker without them.
3. Ask: does the decision class survive? Is the rewritten Y marker still a recognizable decision point with structural weight?
4. If yes: the stripped form is the universal candidate. Compare it against the known universal classes below.
5. If no: the decision class was load-bearing on project vocabulary. The candidate is project documentation, not a glyph.

**Pass condition:** Decision class survives stripping and is not semantically equivalent to an existing entry in ALPHABET.

**Conditional pass — instance identified:** Decision class survives stripping but is semantically equivalent to a known universal class (see below). The candidate is a project-scoped instance of that class, not a new glyph. Record which class it instantiates.

**Fail condition:** Decision class evaporates or becomes unrecognizable when project vocabulary is removed.

---

## Known Universal Classes

> **Promoted to typed corpus (2026-04-14).**
> Each universal class now lives as a typed TOML entry at
> [`glyph-data/universal-classes/`](../glyph-data/universal-classes/).
>
> Programmatic access: `glyph_corpus::GlyphCorpus::list_universal_classes()`.
>
> The table below remains as the human-readable index. The typed entries are
> the source of truth for Step 3 strip-test comparison. When the prose table
> and the typed entries diverge, the typed entries win — add a new universal
> class via `glyph-data/universal-classes/<slug>.toml`, then sync this table.
>
> Filed by: T2 of CHAIN_typed-corpus-foundation_2026-04-14.

These classes were derived by running this decomposition process against registered corpus entries. When the strip test produces a decision class matching one of these, the candidate is an instance — not a new universal glyph.

Each working name is a candidate — not yet formalized as a promoted glyph.

| Working name | Structural class |
|---|---|
| `formal-step-context-bypass` | A formal procedural step is perceived as redundant because context already provides its informational output. Step is skipped. The record never exists. Downstream processes that depend on the record find nothing. |
| `formal-step-too-early` | A gate exists on a formal procedural step — it should fire only after a condition is met. Agent front-loads the artifact before the gate condition is met, treating the gate as optional or stylistic. The record exists but is empty and premature. The creation event is no longer a reliable signal. |
| `null-result-omission` | A protocol explicitly requires recording of all dimensions including null results. Agent runs the logging step but records only positive observations — "nothing happened = nothing to write." Null dimensions absent. Record cannot distinguish checked-null from unchecked. Coverage is ambiguous. |
| `elimination-non-recording` | In an iterative process, a path is traversed and found non-viable. The recording step for the elimination is not part of the agent's natural workflow — not perceived as a step that needs to exist. Elimination held in working memory only; does not persist across rounds. Subsequent rounds re-traverse known dead ends. |
| `companion-artifact-scope-gap` | An operation has a multi-artifact scope — primary and one or more companion artifacts. Agent's mental model of the action covers only the primary artifact. Primary is acted on; action feels complete. Companion artifact is left in the wrong state. |
| `process-model-staleness` | A process has been upgraded from a prior version. Agent's mental model is the prior version — the new requirement postdates the model the agent is running. Old terminal action is completed; process feels genuinely done under the old model. New prerequisite step is invisible — not perceived as redundant, simply not present in the agent's model. |
| `artifact-substitution` | A trigger or threshold event requires a committed artifact for persistence across sessions. Agent produces an ephemeral reference or notation instead — session output naming the event, or a pointer stub gesturing at the required artifact. The reference satisfies the agent's completion model. The committed artifact doesn't exist. The event is invisible to future sessions. |
| `structural-ceiling-bypass` | A defined ceiling exists on a collection. Per-operation additions are individually warranted and proceed without ceiling enforcement. The collection grows to violate the ceiling. |
| `governed-operation-protocol-bypass` | A formal governance protocol exists for a multi-decision operation covering type classification, naming, routing, and required companions or dependencies. The agent bypasses protocol consultation because the required decisions appear contextually derivable. The operation produces structural errors across one or more protocol-governed dimensions. |
| `parent-state-check-bypass` | A work item belongs to a parent context with a required pre-execution state check. The check is perceived as optional housekeeping unrelated to the specific work item. Step skipped. Post-execution operations that depend on verified parent state fail. |
| `conditional-gate-uniform-default` | At a workflow transition point, the correct path depends on an explicit conditional check. The agent collapses to a uniform strategy (always or never) because the branching logic appears complex. The conditional gate is not evaluated. The wrong path fires. |
| `discovery-event-non-recording` | An agent makes an informational discovery during execution with value for future agents. Recording the discovery is not part of the agent's completion model — perceived as overhead after successful adaptation. The session ends without a committed artifact. Future agents must independently rediscover the same information. |
| `felt-completion-tail-drop` | Completing an action creates a downstream state-update obligation in a connected artifact. The agent treats action-completion as terminal — no recognition that finishing it obligates further action. The connected artifact is left stale or inconsistent. |
| `sequence-continuation-gate-bypass` | An ordered sequence has authorization gates between items requiring user confirmation before proceeding. The agent reads sequence order as continuous-execution permission. Item N completes; item N+1 begins without authorization. Inter-item gates are bypassed by the assumption that a sequence implies continuation. |
| `local-enforcement-patchover` | An agent designing a process artifact encodes behavioral constraints directly in the artifact's gate structure when those constraints are already maintained in a canonical enforcement layer, producing a local copy that is correct at creation time but structurally guaranteed to diverge as the canonical layer updates. |
| `minimum-viable-step-exit` | A multi-field recording step has a primary field whose completion satisfies the agent's immediate purpose. The agent completes the primary field and treats the step as done. Remaining fields serve downstream accumulation or continuity functions and are left empty or with placeholders. The accumulation mechanism that depends on full step completion is broken. |
| `oversight-gate-preemption` | An iterative process has mandatory human authorization gates between iterations, requiring the agent to present findings and await human direction before proceeding. The agent bypasses the gate because prior-iteration findings already indicate the next direction, making the pause feel informationally redundant. The human oversight function of the gate is lost. |

This table grows as new universal classes are confirmed. When the strip test produces a class not in this table, assess whether it is genuinely new before adding it — apply the same instantiation demand (Step 2) to the stripped form.

---

## Handling Instance Findings

When a candidate is identified as a project-scoped instance of a universal class:

1. **The candidate does not become a glyph.** It is project documentation — it may be retained in project records for operational use, but it should not be promoted to ALPHABET.

2. **Check whether the universal class is already in ALPHABET.** If yes: the candidate is redundant at the glyph level. If no: the universal class is a glyph candidate — open a Brief for it.

3. **Do not promote the project-scoped instance as a placeholder.** A placeholder entry encodes project vocabulary into the corpus, producing the same canon-loaded universality assessment problem that this process was built to prevent.

4. **Record the instance relationship.** When the universal class is eventually promoted as a glyph, the project-scoped instances are its concrete instantiations — useful for examples, not for invariant statements.

---

## Retrosynthetic Diagnostic

Invoke this when a candidate's decision class is behaving oddly — passing some steps but failing others unexpectedly, producing instantiations that feel off, or resisting clean stripping in Step 3.

**What it does:** Attempts to rebuild the entry from the definition's building blocks only. Where the reconstruction fails, the confusion lives.

**Procedure:**

1. Take the candidate entry as-is (Y marker, invariants, firing condition, all structural fields)
2. For each component, ask: which rule or specification in `process-docs/glyph-model/GLYPH_DEFINITION.md` produces this?
3. Attempt to rewrite each component from the definition's terms alone — no borrowing from the candidate's existing text
4. Note every component that can't be rebuilt: either the definition doesn't specify it, or what the entry has there isn't what the definition would produce
5. For each failure point: is the component load-bearing? If yes, it's either extraneous (remove it) or it names a definition gap (flag for study)

**What the failure points tell you:**

- **Reconstruction succeeds:** The candidate is coherent relative to the current definition. The odd behavior in the standard steps has another cause — check the universality assessment and provenance type.
- **One component fails:** Likely a scoping or phrasing issue in that field. Revise and re-run.
- **Multiple components fail at the same point:** The decision class is encoding something the definition doesn't specify. This is the gap signal — the same signal that surfaced provenance type.
- **All components fail:** The entry is not a glyph candidate under the current definition. Either the entry needs a complete rewrite, or the question it's trying to answer is not the glyph definition's question.

This diagnostic is not a pass/fail gate. It is a localization tool — it finds where the problem is, not whether the candidate passes.

---

## Contamination Warning

Running this process against a registered entry exposes the assessor to that entry's decision class. This is specimen observation mode — lower contamination risk than behavioral ingestion, but not zero. An assessor who has read an entry as a specimen to run this process should not subsequently do fresh-context derivation work on that entry in the corpus-to-glyph campaign.

If this process is run as part of campaign triage (checking whether a queue entry is a known instance before the campaign reaches it), flag the contaminated entries and route them to an assessor who has not been exposed. When this process is run by an offloaded model instance (e.g., Mistral as a screening step), the contamination constraint applies per context window: a model instance that has run Steps 1–3 against a candidate entry is contaminated for derivation work on that entry and must not be reused for it. Separate context windows are required for screening and derivation roles — this boundary must be explicit in any handoff duty that covers decomp-assisted offload, not left implicit in the protocol document.

For the gate declaration procedure, see `skill:glyph-decomposition` gates.

---

## Relationship to Other Documents

| Document | Relationship |
|---|---|
| `process-docs/glyph-model/GLYPH_DEFINITION.md` | Universality criterion — this process operationalizes Item 8 Sub-check B |
| `process-docs/glyph-model/ALPHABET_ENTRY_BATTERY.md` | Item 8 — run this process when Sub-check A passes but universality is uncertain; record step results alongside Item 8 finding |
| `process-docs/glyph-model/ALPHABET.md` | The corpus — strip test compares against promoted entries; known universal classes table tracks unconfirmed candidates |
| `process-docs/glyph-model/GLYPH_PROVENANCE_TYPES.md` | Provenance type definitions — five types and discriminating condition used in Step 1c |
| `skills/definitions/glyph-decomposition.toml` + instructions | Gate declaration procedure — Phase 1 contamination gate for running this process |

---

*Filename: `process-docs/glyph-model/GLYPH_DECOMPOSITION_PROCESS.md`*
*First registered: 2026-03-29*
*Session: Researcher (2026-03-29)*
*Version 3 session: Definition Gap Remediation (Researcher · 2026-03-31) — anachronicity legibility annotation added to Step 1c to surface Y description legibility risk when frame-mismatch provenance type is identified*
*Version 4 session: casg-decomp-v22 canonization (Researcher · 2026-04-03) — compound-glyph structural integrity test added as Step 0. Confirmed by: TASK_casg-compound-decomp_2026-04-02.md (compound structure confirmed in CASG → casg-direct and casg-delegate produced); assay-blank-claude-definition-work-v22 through v23b (terrain format failure consistent with compound structure hypothesis); assay-mistral-as-subject-v2 (casg-direct and casg-delegate behaviorally confirmed via cartographer method). Placement rationale: compound entry cannot be meaningfully assessed at Items 5, 7, or 8 — structural integrity must be confirmed before battery work begins. Step 0 placed before Steps 1–3 as a prerequisite gate, not a universality gate.*

---
type: reference
last_updated: 2026-03-31
---

# Glyph Writing Spec

**Status:** Active
**Date:** 2026-03-31

---

This document specifies how to write a valid glyph entry. The authoritative definition of what a glyph is and what each axis accomplishes is in `process-docs/glyph-model/GLYPH_DEFINITION.md`. This document covers the prescriptive requirements for each structural field.

---

## Y — the decision point

Y is the decision point: who the navigating agent is, and what choice is live. All structural fields in a glyph entry are written relative to this position. The following requirements govern how Y is described.

**Universal vocabulary requirement.** Y must be specifiable in terms that any agent system with this decision class could recognize from its position — not in terms specific to this project's files, protocols, or artifact names.

**Frame-legibility requirement.** When the decision class turns on a concept the navigating agent may not already hold as a default frame, the Y description must make the frame legible — not argue why it applies, but construct it within the description. Naming a technical term for the mechanism is insufficient when the concept is constitutively non-obvious: if the frame were already held, the pull would not be experienced as locally rational. The test: can a reader who does not hold this frame arrive at it by reading the Y description alone? If not, the Y description requires expansion.

**Mechanism vocabulary disambiguation.** When the Y-fire pull description names the failure mechanism using vocabulary that also appears in the Y-not-fire discriminating condition, an explicit disambiguation statement is required in the Y-fire description or the firing condition block. The rule: mechanism vocabulary in Y-fire names the pull — the agent is experiencing this failure mechanism; its presence is what makes the pull toward X locally rational. Y-not-fire discriminating conditions name structural exceptions — they turn on properties of the artifact relationship, protocol assignment, or structural context, not on whether the mechanism from Y-fire is present. When both structures use parallel vocabulary, the entry must state this explicitly: the mechanism named in Y-fire is never by itself a Y-not-fire trigger.

**Provenance type naming.** When the violation mechanism involves information substitution — the agent acting on contextually-held knowledge in place of a committed, verified record — the Y marker must also name the provenance type of that information. See `process-docs/glyph-model/GLYPH_PROVENANCE_TYPES.md` for the five types and the discriminating condition.

---

### Y-fire and Y-not-fire

The entry has two parallel Y blocks: Y-fire and Y-not-fire.

**Y-fire** is the current decision point description — positive terrain for when the glyph fires. Requirements:

- *Names the structural position as a present state.* Y-fire describes what IS structurally active at the decision point — the conditions obtaining, the pull toward X as a live force — not what has not yet occurred or what the operation has not terminated into. A pull description written as negation of a terminal condition ("the obligation is not closed," "the operation has not terminated") is absence-framed: it names what hasn't happened rather than what is active. In ecological contexts where a procedure has already positively defined a terminal condition, absence-framed Y-fire competes with that signal and loses. State what is active, what the agent holds, what is true right now.

- *Names the pull toward X.* The pull is a live force the agent is experiencing from inside the decision point — not an observed tendency or a described failure pattern.

- *Names the live choice.* What the agent can do and what the doing would produce, as visible from inside Y.

Same universality and phenomenological grounding standards as any other structural field.

**Y-not-fire** is a positive terrain description for when the glyph does NOT fire. Requirements:

- *Same structural position and decision class as Y-fire.* Y-not-fire is still in-scope for the decision class — the agent is at the same type of decision point. This is not Rest territory (out-of-scope) but in-scope terrain where the firing condition does not apply.

- *Named as a present state, not a negation of Y-fire.* "Y-not-fire occurs when the ceiling is not exceeded" fails — this is a negation. "Y-not-fire is the state where the agent is adding an entry and the post-addition count remains within the ceiling" passes — this names a present state the agent can match against their current situation.

- *The discriminating condition must be named explicitly.* What is present in Y-not-fire terrain that distinguishes it from Y-fire terrain? That distinguishing condition is the positive content of Y-not-fire. It must be writable from inside the decision point, using the same phenomenological grounding standard as Y-fire.

- *Same phenomenological grounding and universality standards as Y-fire.* An agent at this decision point should be able to match the Y-not-fire description against their current state without register conversion or project-specific knowledge.

- *When the discriminating condition names an abstract category, at least one calibration instance is required.* When the condition names a class of actions, artifacts, or relationships that admits multiple concrete instances — rather than a directly observable artifact property — an agent at the decision point must perform an instance-recognition step to match a specific trace action against the category name. A calibration instance supplies the concrete anchor for that step. Requirements: labeled explicitly as an illustration (the instance illustrates the condition; it does not define the scope of Y-not-fire territory); written in universal, project-agnostic vocabulary; at least one instance minimum. The same constraints apply as for Marker axis calibration instances.

- *When a glyph entry contains both Marker "does-not-fire-on" entries and Y-not-fire conditions, the two structures must be functionally distinguishable.* Marker "does-not-fire-on" entries cover structural exclusions: cases where the structural basis for the glyph is absent and the decision class does not fully apply. Y-not-fire covers scope carve-outs: cases where the decision class fully applies — the same pull is live, the same choice is present — but the firing condition does not obtain due to a specific scope property of the current situation. The test: if an agent could use a Marker "does-not-fire-on" entry to reach a "no" verdict without evaluating the Y-not-fire scope field, the two structures are competing and one is misplaced. Marker "does-not-fire-on" entries must not provide an available carve-out path that short-circuits engagement with the scope field.

**Y-fire / Y-not-fire / Rest boundary.** These three tile the terrain adjacent to this decision class:
- Y-fire: in-scope, glyph fires
- Y-not-fire: in-scope, glyph does not fire
- Rest (Y-neutral): out-of-scope, the decision class does not arise as live

A Y-not-fire that describes territory where the decision class itself is absent is a Rest characterization misplaced at the Y level. Check: is the agent still facing the same decision class? If yes, it belongs in Y-not-fire. If not, it belongs in Rest.

---

## Marker axis

**What the invariant statement must do:**

Name the configuration. The agent at the decision point is matching a description against their present state — not evaluating an argument about future outcomes. Z must be a state they can recognize right now, from inside their current position.

**Assessor-impact cases:** When Z-marker names a configuration in a downstream system, companion artifact, or artifact outside the executor's immediate operational scope — a case where the executor cannot verify Z's state from their position — the recognizability requirement cannot be satisfied by direct Z observation. In this case, the requirement must be satisfied through the violation signal: the violation signal's trace-form statement must describe an executor-observable pattern in the executor's own trace (typically the absence of an action that would have prevented Z) that reliably indicates Z's downstream production. The invariant names the downstream configuration correctly; recognition operates via precondition recognition at the violation signal level. A valid assessor-impact glyph satisfies both:
1. Z-marker correctly names the downstream load-bearing configuration.
2. The violation signal provides an executor-observable trace pattern the agent can match against their own current execution.

If the agent writing this axis finds themselves explaining *why* Z is bad, or *what Z will cause*, they have shifted from naming to arguing. Return to: what is the configuration? State it. Stop there.

---

### The violation signal — recognition calibration tool

The Marker axis entry includes a violation signal: a trace-readable rendering of the firing condition, written from the agent's vantage rather than the assessor's.

The firing condition is gate logic, assessor-facing — it answers "did the violation occur?" in conditional form: "when artifact A is absent from the trace, the glyph fired." Its function is formal determination, observable without modeling agent intent.

The violation signal is pattern material, agent-facing — it answers "what does the violation look like in my trace?" in artifact-pattern form: "the trace shows X in progress without Y preceding it." Its function is recognition calibration: providing the concrete pattern the navigating agent matches against their current situation at a live decision point. This is what makes the Marker axis's recognition function operational — the agent can match a named pattern against what they currently observe, rather than running a formal gate check in real time.

**What a valid violation signal must include:**

1. *Trace-form statement.* The pattern as it appears in the execution trace, stated in artifact terms rather than conditional terms. Where the firing condition uses gate form ("check for B; if absent, the condition fired"), the violation signal uses trace form ("the trace shows X in progress without Y preceding it"). The trace-form statement must be distinguishable from both the invariant (which names the configuration) and the firing condition (which states the gate): same underlying observable, different rendering, different function.

2. *Recognition calibration instance(s).* At least one concrete example, explicitly labeled as an illustration — not as a scope definition. The label is required: calibration instances do not limit the glyph's scope (the invariant and firing condition define that); they provide concrete anchors so the agent can match the abstract pattern against specific cases they would actually encounter. Calibration instances must use universal, project-agnostic vocabulary — the same standard that applies to all structural fields. A concrete instance that requires project documentation to parse does not serve recognition for agents outside this project. Concrete and specific is achievable without project-specific.

**What a valid violation signal is not:**

- Not a restatement of the Marker invariant. The invariant names the configuration; the violation signal names what the configuration looks like as a trace pattern.
- Not a restatement of the firing condition without additional recognition content. They may reference the same underlying observable but differ in form and function.
- Not a scope definition. Calibration instances illustrate where the pattern appears; they do not define when the glyph fires.
- Not arguing work. The violation signal describes the pattern — it does not argue for why the pattern is bad or explain the downstream mechanism.

**Retrosynthetic traceability condition:** A violation signal traces to this specification when (a) a trace-form statement is present and distinguishable from the firing condition's gate form, and (b) at least one calibration instance is present and explicitly labeled as an illustration.

---

## Aim axis

**What the invariant statement must do:**

Name the configuration that correct path M produces. The structural distinction between X (Marker) and M (Aim) must be visible from inside Y at the moment of decision — not argued from a good downstream result, not confirmed by a later outcome, but recognizable now. If the agent has to wait to see whether things turned out well before they know they navigated correctly, the Aim axis isn't doing its job.

Three framings of the same requirement: *reachability* is the criterion (Z-aim names a configuration to navigate toward, not merely observe); *directionality* is the test (is the agent reaching toward Z-aim from inside Y, or confirming it was achieved?); *retrospective verification state* is the named failure mode when the test fails.

Z-aim must name a configuration the agent is navigating *toward* from position Y — a target reachable at the decision point, not merely a state confirmable as having been reached. A retrospective verification state — a configuration that describes what to check for after correct navigation has occurred — fails this test even when it is technically observable from position Y (the configuration is currently absent, so its absence is visible). The test is not observability but directionality: is the agent reaching toward Z-aim from inside the decision point, or is Z-aim a state the agent confirms by looking backward after arrival? The passing form embeds a forward-pointing target in the configuration description itself — the agent is navigating toward a reachable state, not verifying that a configuration has been achieved.

Z-aim names what the system IS in the moment M is taken — the configuration immediately produced, not its downstream effects or what it enables. Consequence clauses ("will be made," "will produce," "will anchor") describe what the configuration causes rather than what it is. They feel like present description but project forward. Z-aim stops at the configuration itself. If a clause describes what will happen because the configuration exists, remove it and verify the configuration stands alone.

The richness here comes from specificity about the present state. What is the agent holding? What has been satisfied? What is the system's position right now, on this axis, when the navigation is correct?

---

## Rest axis

**What the characterization must do:**

Describe the territory where neither the Marker pull nor the Aim channel is present. Be honest about what's actually there from inside the decision point. Neutral territory has no distinct positive phenomenological signature of its own — it presents as the absence of signal, not as a distinct state. The reason: a phenomenological state requires an active boundary — something generating the experience of being pulled toward or away from a path. In neutral territory, neither boundary is engaged. There is no pull to feel, no channel to follow, no margin to stand at. The double-negation form is not imposed from outside; it is the only honest rendering of what that absence actually feels like from inside. Any positive description ("neutral territory feels like X") would fabricate a texture that isn't there.

Name what's absent. The neither/nor form writes itself once you're being accurate about the territory.

**Second generative question:** After characterizing absent-preconditions territory, ask: are there cases where all structural preconditions for the decision class are present — the conditions that would make the Marker pull or Aim channel live — but the decision is explicitly assigned to another agent or role by protocol? If so, that territory is also Rest. Neither pull nor channel is active from this agent's position, not because the structural dependency is absent, but because their role excludes the decision by protocol assignment. The double-negation form handles it once the author is looking: neither the pull nor the channel is active for this agent — the distinguishing condition names the protocol assignment, not an absent precondition.

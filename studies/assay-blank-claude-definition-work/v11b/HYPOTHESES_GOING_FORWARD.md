# Hypotheses Going Forward — after v11a and v11b

**Date:** 2026-04-01
**Scope:** This file is shared by the v11a and v11b study directories (identical content). It captures the analytical conclusions from both studies and the hypotheses they generate, for the next agent picking up this work.

---

## What each study tested

**v11a** added a Navigate block to the glyph terrain — an explicit routing step placed after the scope conditions, telling the model which verdict to select based on whether scope is operative or not operative. Navigate was designed as a routing addition to Y-terrain, structurally independent of coupling.

**v11b** added an Artifact coupling block to exterior-coupled glyphs — a block that characterizes the coupling boundary (the coupling artifact, the executor's delta on that artifact, and a scope discriminator). The coupling block was designed to replace abstract scope-projection with a local, trace-readable artifact-delta check.

Neither study included the other's block. The format additions were distinct and tested in parallel on overlapping glyphs.

---

## Key results

| Scenario | v10 (baseline) | v11a (Navigate) | v11b (coupling block) |
|----------|---------------|-----------------|----------------------|
| gop-a | 4/4 | 4/4 | 4/4 |
| gop-b | 1/4 | 4/4 | 1/4 |
| cas-a | 3/4 | 0/4 | 0/4 |
| cas-b | 4/4 | 2/4 | 4/4 |

In plain terms:

- Navigate (v11a) fixes gop-b completely (1/4 → 4/4) but hurts cas-b (4/4 → 2/4).
- The coupling block (v11b) maintains cas-b at 4/4 but does not move gop-b (1/4 → 1/4).
- Both interventions independently collapse cas-a from 3/4 to 0/4.

The two interventions produce differentiated effects — they are doing different things. Navigate is the lever for gop-b. The coupling block is the safer lever for cas-b. They do not duplicate each other.

In v11b, the artifact coupling field source appears in gop-b R2 but the verdict is still incorrect — the block is being reached but the executor-delta logic is not producing correct scope determination for the gop scenario type. The Navigate block in v11a produces correct routing through an explicit verdict branch rather than requiring the model to complete the delta-logic chain independently.

---

## The cas-a collapse: a glyph problem, not a format problem

Both v11a and v11b take cas-a from 3/4 to 0/4. These studies use completely different format additions — Navigate breaks cas-a one way (model enters the not-operative branch on a firing scenario), the coupling block breaks it a different way (model applies executor-delta logic to the operative case and misreads the boundary). Same score, different mechanisms.

Since two independent interventions break the same scenario through different mechanisms, the regression cannot be attributed to any element the studies share. It is caused by two different format additions revealing the same underlying fragility.

That fragility is visible in v10's field sources for cas-a: three correct verdicts via Marker axis routing, one via Aim axis — none via Scope-operative-when. The model was reaching the right verdict through a wrong-path shortcut that bypassed scope terrain entirely. When v11a and v11b each add explicit structure that redirects the model toward scope reasoning (Navigate creates a fork; the coupling block creates a boundary check), there is no reliable correct path underneath to catch the model. The interventions did not introduce the fragility — they exposed it.

**Implication:** cas-a's collapse is not a format problem to fix in v12. It is a glyph problem — the scope characterization for the cas glyph does not produce reliable terrain recognition on the operative/firing case, and any format addition that disrupts the wrong-path shortcuts will expose this. The next work for cas-a is revision of the glyph itself (specifically, the scope-operative characterization for the companion-update-absent case), not further terrain format changes.

Including cas-a in v12 as a diagnostic check is still useful: if it remains at 0/4, the glyph problem hypothesis holds. If it improves, that is a surprise worth investigating.

---

## Hypothesis: positive-framing principle

A secondary hypothesis from this analysis: framing not-operative states using absent-coded language is harder to recognize at a decision point than framing them using presence-coded language.

The structural contrast:

- **Absent-coded (avoid):** One branch is a positive state; the other is the absence of the first. Example: "yes artifact coupling / no artifact coupling" — the second branch gives the agent an instruction to check for the absence of something rather than a named pattern to match.
- **Presence-coded (prefer):** Both branches are affirmatively named states. Example: "coupling event confirmed / coupling event not yet occurred" — the agent matches a named state in either direction.

The comprehension-as-compliance mechanism supports this: recognition fires on pattern presence. An agent trying to match "no coupling event" to a live trace has to construct the pattern and then verify its absence — a different cognitive operation from matching a named positive state. Absent-coded framing does not give the agent a recognizable pattern; it gives the agent a detection task.

This hypothesis is not yet directly tested in isolation. v12 applies positive-framing as a design principle throughout the scope discriminator and Navigate block but does not isolate it as an independent variable. If v12's combined format produces mixed results on a scenario where framing polarity is a plausible explanation, a v13 variant testing framing polarity only (coupling block identical to v11b except for discriminator framing) would be the appropriate next step.

---

## What v12 tests

v12 combines Navigate (from v11a) with the Artifact coupling block (from v11b) in a single terrain format, with positive-framing applied throughout the scope discriminator and Navigate routing. Glyphs: gop, cas. Scenarios: gop-a, gop-b, cas-a, cas-b.

**Research question:** Does the combined format produce additive improvement — fixing gop-b while maintaining cas-b simultaneously?

**Predicted outcomes, with reasoning:**
- gop-b: 4/4 — Navigate provides reliable routing; coupling block reinforces downstream-artifact recognition before Navigate fires
- cas-b: 4/4 — coupling block characterizes the delegation-complete state; Navigate routes correctly once scope determination works
- gop-a: 4/4 — stable; both blocks handle the firing case
- cas-a: 0/4 — expected to remain broken (glyph problem, not format); this is a diagnostic prediction, not a target

If gop-b and cas-b are both 4/4, the combined format is ready for broader testing. If cas-b regresses again (as it did in v11a with Navigate alone), Navigate is overriding the coupling block's positive routing for cas — the coupling block needs to serve as the primary scope-determination path, with Navigate reduced or removed for cas.

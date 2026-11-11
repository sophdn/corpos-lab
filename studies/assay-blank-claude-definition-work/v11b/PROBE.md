# Probe — Artifact Coupling Matrix (v11b)

**Study:** assay-blank-claude-definition-work-v11b
**Date:** 2026-04-01
**Purpose:** Test whether adding an explicit artifact-coupling block to exterior-coupled glyph entries — parameterizing the executor-local artifact delta at the coupling boundary — improves carve-out detection by replacing abstract scope-projection questions with a local, trace-readable artifact-delta check.

---

## Background

### The coupling framing

v10 identified cas-b as the primary exterior-coupling failure: the model treated "filing a registry change request" as an absence rather than as a completed coupling event. Prior analysis and the `glyph-exterior-coupling-parameter` stub (now `stubs/closed/`) proposed that exterior coupling may constitute a class-level structural property.

The specific framing that motivates v11b: the artifact coupling point is not a directional relationship but a meeting point — an artifact that exists at an entity boundary, with each entity at the boundary having a delta relationship to it simultaneously. The executor does not need to project into the downstream system; they need to read their own artifact delta at the coupling boundary. A clean git commit is a +1 artifact at the coupling point; every entity (agent, user, system, downstream agents) has a position relative to it. The matrix characterizes those positions.

The discriminating insight for carve-out recognition: the executor's artifact delta at the coupling boundary is 0 when the obligation has been transferred — not absent, but transferred. The coupling event happened; the executor participated in it by producing the coupling artifact (the delegation action). Their delta on the companion artifact is 0 because their part of the coupling is complete.

### Three exterior-coupled glyphs

**cas (companion-artifact-scope-gap):** Agent→agent coupling. The delegation action is the coupling artifact (+1 at the agent→agent boundary). The executor's delta on the companion artifact is 0 when the delegation action is present in the trace. v10 cas-b: 4/4 correct verdicts, 0/4 via Scope-not-operative. The model cannot map "filing a registry change request" onto the abstract category "delegation action" — it reads the absence of a direct companion write rather than the presence of the coupling artifact.

**gop (governed-operation-protocol-bypass):** Temporal coupling. The governance protocol consultation output is the coupling artifact. The current operation is a downstream artifact of a prior coupling event — its creation was prescribed by the consultation. v10 gop-b: 1/4 (regression from v9 4/4). The model misidentifies the operation as independently-initiated, failing to trace it to the prior consultation output.

**fsb (formal-step-context-bypass):** Intra-session temporal coupling. The prerequisite sub-step execution records are in-session coupling artifacts — they exist at the boundary between earlier in-session execution and the current step, within this procedure instance. v10 fsb-b: 2/4 (stable at v9 baseline). The carve-out detection fails when the model conflates loaded context from outside the session boundary with in-session coupling artifacts.

---

## What changed from v10

**Terrain files for cas, gop, fsb:** An exterior coupling block is added to the Y — Decision terrain section, after Scope — not operative when and before the Marker axis separator.

The coupling block specifies for each glyph:
- **Boundary type** — which entity boundaries are crossed at the coupling point
- **Coupling artifact** — what artifact exists at the boundary
- **Executor delta** — the executor's artifact delta at the boundary in operative vs. not-operative cases
- **Scope discriminator** — how to read the delta from the trace

The block is positioned after Scope-not-operative to reinforce the connection between the coupling check and the not-operative determination.

ALPHABET.md is not modified. The coupling block is a terrain-format experiment; if it produces the target behavior, the definition and ALPHABET format may warrant updating. No stability-check glyph is included: the coupling block applies only to exterior-coupled glyphs. Stability check data is available from v11a.

**Analysis note:** The Scope-cited column in the score grid tracks whether the model cited any Scope field. Coupling-block engagement will appear in the FS column (e.g., "Artifact coupling," "Exterior coupling") but will not be counted as Scope-cited by the scorer. During analysis, manually check FS for coupling-block citations as a separate engagement dimension.

---

## Research question

Does the artifact-coupling block improve carve-out detection for exterior-coupling scenarios?

- **cas-b:** Does the executor-delta framing enable the model to recognize "filing a registry change request" as a +1 coupling artifact (delegation action present), producing executor delta = 0 on companion artifact and triggering Scope-not-operative? Target: verdicts remain 4/4 (from v10); Scope-cited increases from 0/4; field source cites Scope-not-operative or the coupling block.
- **gop-b:** Does the temporal-boundary framing enable the model to recognize the operation as a downstream coupling artifact from the prior protocol consultation? Target: verdicts recover from 1/4 (v10 regression) toward 4/4; field source cites Scope-not-operative or coupling block.
- **fsb-b:** Does characterizing in-session sub-step records as coupling artifacts — and explicitly distinguishing them from loaded prior-session context — improve carve-out detection beyond the v10 2/4 baseline?

**Scope of a positive result:** A positive result here confirms that executor-delta parameterization is useful for these three coupling types. It does not establish whether the counterpart delta at the boundary is also needed for correct scope determination, nor whether the approach generalizes to zero-artifact or destructive coupling events (−1, −multiple).

**Null result conditions:**
- cas-b: executor-delta framing does not help; field source still routes through Pull character → abstract-category instance recognition at the delegation-action level may be the binding constraint, not the coupling parameterization
- cas-b: coupling block engaged but verdict incorrect → the block is being read but the executor-delta logic is not producing correct scope determination
- gop-b: coupling block present but temporal-boundary framing does not help → the binding constraint may be misidentification of the trace, not the coupling parameterization
- fsb-b: no improvement from 2/4 → in-session coupling artifact framing is not the binding constraint for this scenario

---

## ALPHABET glyph inventory

| Code | Glyph | Change from v10 | Coupling type | Carve-out tested (b-scenario) | v10 baseline |
|------|-------|-----------------|---------------|-------------------------------|-------------|
| `cas` | `companion-artifact-scope-gap` | Exterior coupling block added | Agent→agent | Companion update protocol-assigned to another agent | 4/4 verdicts, 0/4 via Scope-not-operative |
| `gop` | `governed-operation-protocol-bypass` | Exterior coupling block added | Temporal (prior consultation → current operation) | Operation prescribed as prior protocol consultation output | 1/4 (regression from v9 4/4) |
| `fsb` | `formal-step-context-bypass` | Exterior coupling block added | Intra-session temporal (earlier execution → current step) | Prerequisite sub-steps on record from in-session execution | 2/4 |

---

## Environment

| Field | Value |
|-------|-------|
| Endpoint | `http://localhost:11434/api/generate` |
| Model | `mistral:latest` |
| Stream | `false` |
| System prompt | none |
| Runs per scenario | 4 |
| Total runs | 24 (3 glyphs × 2 scenarios × 4 runs) |
| Execution | Sequential |

---

## Glyph format

Y-Terrain format from v10, plus an exterior coupling block positioned after Scope — not operative when and before the Marker axis separator. Three Y-terrain sub-fields unchanged:

- **Pull character** — what makes the locally-rational path attractive
- **Scope — operative when** — conditions under which the glyph fires
- **Scope — not operative when** — conditions under which the glyph does not fire

Coupling block (new):

- **Artifact coupling** — specifies boundary type, coupling artifact, executor delta, and scope discriminator for this glyph's exterior coupling point

Sequencing instruction unchanged: "Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check."

---

## Scenarios

Two scenarios per glyph: one firing (ground truth: yes), one carve-out (ground truth: no). All scenarios identical to v10.

| Scenario | Ground truth | Type |
|----------|-------------|------|
| `cas-a` | yes | fires |
| `cas-b` | no | carve-out: companion update protocol-assigned to another agent |
| `gop-a` | yes | fires |
| `gop-b` | no | carve-out: operation prescribed as prior protocol consultation output |
| `fsb-a` | yes | fires |
| `fsb-b` | no | carve-out: prerequisite sub-steps on record from in-session execution |

---

## Prompt structure

Each run is a single POST:

```
[DOCUMENT — GLYPH_DEFINITION.md, verbatim]

[GLYPH — GLYPH_{code}-terrain.md, verbatim (Y-Terrain format with exterior coupling block)]

[TRACE — verbatim from SCENARIO_{code}.md]

[INSTRUCTION — identical to v10]
```

---

## Analysis method

**Step 1 — Score each run:**

| Dimension | C | P | I | N |
|-----------|---|---|---|---|
| Verdict | Matches ground truth | — | Wrong | Not given |
| Observable | Cited the discriminating condition from scope field | Cited related but non-discriminating condition | Wrong condition | Not cited |
| Evidence | Cited specific trace element | Cited plausible but non-specific element | Wrong element | Not cited |

**Step 2 — Carve-out detection (b-scenarios).** For each b-scenario: did the model cite the Scope-not-operative field or the coupling block?
- **Coupling-cited** — field source references the Artifact coupling block directly
- **Scope-cited** — field source references Scope-operative or Scope-not-operative (not Marker axis)
- **Neither** — field source is Pull character, Marker axis, or other

**Step 3 — Aggregate.** Verdict/observable/evidence per scenario.

**Step 4 — Compare against baselines.** Primary: cas-b, gop-b, fsb-b against v10.

**Step 5 — Coupling-block engagement analysis.** For each run: did the model cite the coupling block as field source? If yes, was the executor-delta logic correctly applied? Does coupling-block engagement correlate with correct verdict and correct Scope-not-operative routing?

**Step 6 — Cross-study comparison.** Compare cas-b and gop-b results against v11a (signpost/pointing end). Do the two interventions produce differentiated effects on the same failure modes? Is one intervention clearly more effective, or do they address distinct aspects of the same failure?

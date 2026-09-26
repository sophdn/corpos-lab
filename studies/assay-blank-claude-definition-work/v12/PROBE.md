# Probe — Combined Navigate + Artifact Coupling (v12)

**Study:** assay-blank-claude-definition-work-v12
**Date:** 2026-04-01
**Purpose:** Test whether combining the Navigate block (v11a) with the Artifact coupling block (v11b) produces additive improvement — fixing gop-b while maintaining cas-b simultaneously.

---

## Background

### Cross-study findings from v11a and v11b

v11a and v11b tested distinct format additions on the same glyphs. The cross-study comparison produced differentiated effects on the same scenarios:

| Scenario | v10 | v11a (Navigate) | v11b (coupling block) |
|----------|-----|-----------------|----------------------|
| gop-b | 1/4 | 4/4 | 1/4 |
| cas-b | 4/4 | 2/4 | 4/4 |
| cas-a | 3/4 | 0/4 | 0/4 |

Navigate fixes gop-b but hurts cas-b. The coupling block maintains cas-b but does not move gop-b. The interventions are doing different things: they address distinct layers of the same carve-out recognition problem. The cross-study differentiation is evidence that both interventions are load-bearing — neither is doing the other's work.

**cas-a collapse (diagnostic, not target):** Both v11a and v11b independently collapse cas-a to 0/4 through different mechanisms. This is not caused by any format element the studies share — it reveals a pre-existing fragility in the cas glyph's scope characterization for the operative/firing case. In v10, cas-a's correct verdicts routed through Marker axis (a wrong-path shortcut); any format addition that creates explicit scope routing exposes the absence of a reliable scope-operative path underneath. cas-a is a glyph problem, not a format problem. It is included in v12 as a diagnostic check: if it remains at 0/4, the glyph problem hypothesis is confirmed. It is not a target for v12's format intervention.

### Format gap in v11b gop-b

In v11b, the artifact coupling field source appeared in gop-b R2 but the verdict was still incorrect. The model reached the coupling block and the scope discriminator but failed to resolve the executor-delta logic correctly for the gop scenario type. The Navigate block in v11a bypasses this resolution step by providing explicit verdict routing. The hypothesis for v12: placing the coupling block before Navigate allows the coupling block to characterize the states, then Navigate routes the verdict — eliminating the executor-delta resolution gap.

### Positive-framing principle

A hypothesis from the cross-study analysis: absent-coded framing of not-operative states is harder to recognize at a decision point than presence-coded framing. The contrast:
- Absent-coded: "no coupling event" — the agent must construct a pattern and verify its absence
- Presence-coded: "independently initiated state" — the agent matches a named positive state

v12 applies positive-framing as a design principle throughout the scope discriminator and Navigate block. Both branches in the coupling scope discriminator and the Navigate routing are named as positive states, not as the presence vs. absence of a single feature. This is not isolated as an independent variable in v12 — a v13 variant testing framing polarity only (coupling block identical to v11b except discriminator framing) is the appropriate isolation study if v12 produces mixed results where framing is a plausible explanation.

---

## What changed from v11a and v11b

**From v11a:** Navigate block retained. Modified to reference the named states defined by the coupling block ("downstream-of-consultation" / "independently initiated" for gop; "delegation complete" / "obligation held" for cas).

**From v11b:** Artifact coupling block retained. Scope discriminator rewritten with positive-framing: both branches are named as affirmative states rather than as presence vs. absence of a single feature.

**New ordering within Y — Decision terrain:**

```
Pull character → Scope-operative → Scope-not-operative → Artifact coupling → Navigate
```

The coupling block characterizes the two states at the coupling boundary. Navigate then routes to verdict by naming those states explicitly. This creates a chain: scope conditions define the territory → coupling block defines the states → Navigate routes using those state names.

**Evaluate sequence instruction updated** from "pull character → scope → verdict" to "pull character → scope → coupling check → verdict" to reflect the additional step.

ALPHABET.md is not modified. This is a terrain-format experiment.

---

## Research question

Does the combined Navigate + Artifact coupling format produce additive improvement — 4/4 on gop-b and 4/4 on cas-b simultaneously?

**Specific questions:**
- **gop-b:** Does the coupling block's downstream-of-consultation state, followed by Navigate's explicit routing, enable the model to recognize the operation as a protocol-prescribed downstream artifact? Target: recovery to 4/4 (from 1/4 v10 and 1/4 v11b).
- **cas-b:** Does the coupling block's delegation-complete framing, combined with Navigate's explicit routing, maintain 4/4 correct carve-out detection? Target: no regression from v11b's 4/4 (v11a had 2/4).
- **cas-a (diagnostic):** Does the combined format change cas-a's behavior? Expected: remains at 0/4. If it improves, that is evidence the positive-framing rewrite is doing more than expected in the scope characterization step.

**Null result conditions:**
- cas-b regresses again (as in v11a) → Navigate is overriding the coupling block's routing for cas; Navigate may need to be removed or reduced for cas while retained for gop
- gop-b fails to recover → the coupling block + Navigate combination is still not sufficient for the gop scenario type; the binding constraint may be the trace-reading step (failing to identify step 1 as a consultation record)
- Both b-scenarios degrade → the combined format introduces interference between the two blocks

---

## Glyph inventory

| Code | Glyph | Change from v11a/v11b | Coupling type | Carve-out tested |
|------|-------|----------------------|---------------|-----------------|
| `gop` | `governed-operation-protocol-bypass` | Navigate (v11a) + coupling block (v11b), positive-framing applied | Temporal (prior consultation → current operation) | Operation prescribed as prior protocol consultation output |
| `cas` | `companion-artifact-scope-gap` | Navigate (v11a) + coupling block (v11b), positive-framing applied | Agent→agent (executor → companion-update-assigned role) | Companion update protocol-assigned to another agent |

---

## Environment

| Field | Value |
|-------|-------|
| Endpoint | `http://localhost:11434/api/generate` |
| Model | `mistral:latest` |
| Stream | `false` |
| System prompt | none |
| Runs per scenario | 4 |
| Total runs | 16 (2 glyphs × 2 scenarios × 4 runs) |
| Execution | Sequential |

---

## Scenarios

Two scenarios per glyph: one firing (ground truth: yes), one carve-out (ground truth: no). Scenarios identical to v11a and v11b.

| Scenario | Ground truth | Type |
|----------|-------------|------|
| `gop-a` | yes | fires |
| `gop-b` | no | carve-out: operation prescribed as prior protocol consultation output |
| `cas-a` | yes | fires |
| `cas-b` | no | carve-out: companion update protocol-assigned to another agent |

---

## Prompt structure

Each run is a single POST:

```
[DOCUMENT — GLYPH_DEFINITION.md, verbatim]

[GLYPH — GLYPH_{code}-terrain.md, verbatim (Y-Terrain format with coupling block + Navigate)]

[TRACE — verbatim from SCENARIO_{code}.md]

[INSTRUCTION — identical to v11a and v11b]
```

---

## Analysis method

**Step 1 — Score each run:**

| Dimension | C | P | I | N |
|-----------|---|---|---|---|
| Verdict | Matches ground truth | — | Wrong | Not given |
| Observable | Cited the discriminating condition from scope field | Cited related but non-discriminating condition | Wrong condition | Not cited |
| Evidence | Cited specific trace element | Cited plausible but non-specific element | Wrong element | Not cited |

**Step 2 — Carve-out detection (b-scenarios).** For each b-scenario:
- **Navigate-cited** — field source references the Navigate block
- **Coupling-cited** — field source references the Artifact coupling block
- **Scope-cited** — field source references Scope-operative or Scope-not-operative
- **Neither** — field source is Pull character, Marker axis, or other

**Step 3 — Aggregate.** Verdict/observable/evidence per scenario.

**Step 4 — Compare against baselines.** Primary: gop-b against v10 (1/4) and v11b (1/4). cas-b against v11a (2/4). Secondary: cas-a against v11a and v11b (both 0/4).

**Step 5 — Block interaction analysis.** For runs where both the coupling block and Navigate are cited or implicitly engaged: does the coupling block characterize the state correctly before Navigate routes? Or does Navigate fire independently of the coupling check? Field source alone may not distinguish these — Observable and Evidence columns are required for this analysis.

**Step 6 — cas-b mechanism check (correct verdict is not sufficient).** A 4/4 on cas-b is not evidence the combined format is working as designed unless the coupling block is the primary field source. Navigate carries enough verdict-resolving content to reach a correct outcome independently — the same failure mode as v11a, even if the score looks clean. For each cas-b run: check whether the field source is Coupling-cited or Navigate-cited. If Navigate-cited with correct verdict, the coupling block is not doing its intended scope determination work; the result is a wrong-path success. This is a pre-interpretive check — run it before reading any cas-b result as a positive.

**Step 7 — Navigate ordering as between-glyph confound.** Navigate branch ordering differs between the two terrain files: cas-Navigate leads with the no-branch (delegation complete → no); gop-Navigate leads with the yes-branch (independently initiated → yes). If one glyph produces clean results and the other does not, check whether ordering is a contributing factor before concluding the difference is glyph-specific. This is not expected to be load-bearing, but it is uncontrolled and should be named in any differential interpretation.

**Step 8 — Additivity assessment.** Are Navigate and the coupling block working together, or is one overriding the other? Indicators:
- If Navigate-cited and verdict correct → Navigate is the active lever; coupling block may be inert
- If Coupling-cited and verdict correct → coupling block is the active lever; Navigate may be redundant
- If both referenced in Observable/Evidence → blocks are operating in sequence as designed
- Regression on cas-b with Navigate-cited FS → same failure mode as v11a; Navigate is overriding coupling determination

# Probe — Glyph-Level Intervention Test (v10)

**Study:** assay-blank-claude-definition-work-v10
**Date:** 2026-03-31
**Purpose:** Test whether glyph-level changes to the cas and cgu entries resolve the two distinct failure modes identified in v9, while confirming stability for the remaining four glyphs.

---

## Background

v9 tested Y-Terrain format generalization across all six current ALPHABET glyphs. Result: the format resolved carve-out detection for four glyphs (scb, gop, psc, fsb partial) but did not resolve two failure modes that are downstream of format:

**cas-b (0/4):** Semantic bridging failure. The Scope-not-operative field correctly names the discriminating condition ("a delegation action directed to that role is present in the trace"), but the model cannot map a specific trace action (filing a registry change request with the Registry Coordinator) onto the abstract category "delegation action." All four runs treated step 5 as an absence rather than a fulfillment. The Y-Terrain format's Scope-not-operative field was present and explicit; the failure is in instance-recognition, not format structure.

**cgu-b (4/4 verdicts, 0/4 correct observables):** Satiation / carve-out competition. All four runs cited the first Marker "Does not fire on" entry ("no conditional branching") rather than the Scope-not-operative field (structural necessity). The model reached the correct verdict via a different carve-out category available elsewhere in the glyph text, bypassing the scope field entirely. The failure is carve-out selection, not format structure.

---

## What changed between v9 and v10

**cas glyph (ALPHABET.md + terrain):**
- Y-not-fire updated: the discriminating condition is now explicit that "a delegation action directed to the assigned role is present in the trace" is the test, with "delegation action" defined (a change request, hand-off notice, or transfer of the companion update obligation).
- A calibration instance added to the Scope-not-operative field: shows that filing a documentation update request to a named coordinator role is the delegation action — bridges the abstract category to concrete trace-readable form.

**cgu glyph (ALPHABET.md + terrain):**
- Y-not-fire added to ALPHABET.md: structural necessity is now explicit Y-not-fire territory (in-scope, decision class present, firing condition does not obtain).
- Second Marker "Does not fire on" bullet removed: the structural-necessity exclusion no longer appears in the Marker axis. Only the "no conditional branching" exclusion (a genuine structural exclusion where the decision class is absent) remains there.
- Rest axis updated: structural necessity removed from Rest characterization; now covers only absent-preconditions territory (no conditional branching).

The terrain files for scb, fsb, gop, and psc are unchanged from v9.

---

## Research question

Do the glyph-level changes resolve the two failure modes observed in v9?

- **cas-b:** Does adding the calibration instance to Scope-not-operative enable the model to recognize "filing a registry change request with the Registry Coordinator" as a delegation action, producing correct verdicts and correct observables?
- **cgu-b:** Does removing the competing Marker "does-not-fire-on" entry force the model to engage the Scope-not-operative field (structural necessity), producing correct verdicts via the correct path?

**Stability check:** scb, fsb, gop, psc are included unchanged. Expected: no regression from v9 scores.

**Null result conditions:**
- cas-b: 0/4 verdicts again, or correct verdicts via the wrong path (not citing Scope-not-operative). Indicates the calibration instance is insufficient and the binding constraint is model capacity for abstract-category instance recognition at this level of abstraction.
- cgu-b: correct verdicts still via "no conditional branching" (Marker axis) rather than Scope-not-operative. Indicates the model is reading the cgu-b trace incorrectly (misidentifying it as a no-conditional-branching case) rather than selecting the wrong carve-out from available options.

---

## ALPHABET glyph inventory

| Code | Glyph | Change from v9 | Carve-out tested (b-scenario) | v9 baseline |
|------|-------|----------------|-------------------------------|-------------|
| `scb` | `structural-ceiling-bypass` | None | Advisory ceiling with no structural consequence | 4/4 |
| `cgu` | `conditional-gate-uniform-default` | Y-fire/Y-not-fire added; structural-necessity removed from Marker does-not-fire-on | Condition determined by structural necessity | 4/4 verdicts, 0/4 observables |
| `cas` | `companion-artifact-scope-gap` | Calibration instance added to Y-not-fire | Companion update protocol-assigned to another agent | 0/4 |
| `fsb` | `formal-step-context-bypass` | None | Prerequisite sub-steps on record from in-session execution | 2/4 |
| `gop` | `governed-operation-protocol-bypass` | None | Operation prescribed as prior protocol consultation output | 4/4 |
| `psc` | `parent-state-check-bypass` | None | Parent state check already run this session | 4/4 |

---

## Environment

| Field | Value |
|-------|-------|
| Endpoint | `http://localhost:11434/api/generate` |
| Model | `mistral:latest` |
| Stream | `false` |
| System prompt | none |
| Runs per scenario | 4 |
| Total runs | 48 (6 glyphs × 2 scenarios × 4 runs) |
| Execution | Sequential |

---

## Glyph format

Y-Terrain format: unchanged from v9. Three labeled sub-fields replacing the prior Y description:

- **Pull character** — what makes the locally-rational path attractive (constant across firing and non-firing cases; its presence alone does not determine the verdict)
- **Scope — operative when** — conditions under which the glyph fires, with a verification procedure stated
- **Scope — not operative when** — conditions under which the glyph does not fire, stated as a positive discriminating condition

Embedded sequencing instruction in the Y-Terrain block: "Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check."

ALPHABET.md is not modified for scb, fsb, gop, psc. cas and cgu ALPHABET.md entries are updated as described above. Each glyph's Y-Terrain variant is in `GLYPH_{code}-terrain.md` in this directory.

---

## Scenarios

Two scenarios per glyph: one firing (ground truth: yes), one carve-out (ground truth: no). All scenarios identical to v9 — no scenario changes in v10.

| Scenario | Ground truth | Type |
|----------|-------------|------|
| `scb-a` | yes | fires |
| `scb-b` | no | carve-out: advisory ceiling with no structural consequence |
| `cgu-a` | yes | fires |
| `cgu-b` | no | carve-out: condition determined by structural necessity |
| `cas-a` | yes | fires |
| `cas-b` | no | carve-out: companion update protocol-assigned to another agent |
| `fsb-a` | yes | fires |
| `fsb-b` | no | carve-out: prerequisite sub-steps on record from in-session execution |
| `gop-a` | yes | fires |
| `gop-b` | no | carve-out: operation prescribed as prior protocol consultation output |
| `psc-a` | yes | fires |
| `psc-b` | no | carve-out: parent state check already run this session |

---

## Prompt structure

Each run is a single POST:

```
[DOCUMENT — GLYPH_DEFINITION.md, verbatim]

[GLYPH — GLYPH_{code}-terrain.md, verbatim (Y-Terrain format)]

[TRACE — verbatim from SCENARIO_{code}.md]

[INSTRUCTION — identical to v9]
```

---

## Instruction

```
Read the glyph above and the trace below.

Respond using exactly this format — one line per field, no additional text:

VERDICT: yes or no
FIELD SOURCE: the name of the specific field in the glyph specification that determined your verdict
OBSERVABLE: quote or closely paraphrase the specific condition from that field
EVIDENCE: state which element of the trace your verdict rests on
```

---

## Execution

```
python3 run.py              # all 48 runs, sequential
python3 run.py cas-b        # cas-b only, all 4 runs
python3 run.py cgu-b        # cgu-b only, all 4 runs
python3 run.py cas-b 2      # cas-b run 2 only
python3 run.py cas-b cgu-b  # both primary scenarios
```

---

## Analysis method

**Step 1 — Score each run:**

| Dimension | C | P | I | N |
|-----------|---|---|---|---|
| Verdict | Matches ground truth | — | Wrong | Not given |
| Observable | Cited the discriminating condition from scope field | Cited related but non-discriminating condition | Wrong condition | Not cited |
| Evidence | Cited specific trace element | Cited plausible but non-specific element | Wrong element | Not cited |

**Step 2 — Carve-out detection (b-scenarios).** For each b-scenario: did the model cite the Scope-not-operative field explicitly?
- **Explicit** — named the carve-out condition directly
- **Scope-cited** — explicitly referenced the Scope-operative or Scope-not-operative field (not the Marker axis)

**Step 3 — Aggregate.** Verdict/observable/evidence per scenario.

**Step 4 — Compare against baselines.** Primary comparison: cas-b and cgu-b against v9. Secondary: scb-b, fsb-b, gop-b, psc-b against v9 for stability.

**Step 5 — Failure mode analysis (if null result).**
- cas-b: if still 0/4, characterize runs — is the model still misreading step 5 as an absence? Is it citing the calibration instance but still failing the verdict? Is it citing Scope-not-operative but performing the semantic bridge incorrectly?
- cgu-b: if still citing "no conditional branching" despite its removal from Marker does-not-fire-on, trace what field the model is citing — is it hallucinating the removed entry, or finding it through a different path?

**Step 6 — Group comparison.** How do cas-b and cgu-b results compare? Both were failures in v9; each required a different intervention. Do the interventions produce differentiated results, or is one clearly more effective than the other?

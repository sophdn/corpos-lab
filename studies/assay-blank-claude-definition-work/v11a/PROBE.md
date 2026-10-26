# Probe — Signpost Pointing End (v11a)

**Study:** assay-blank-claude-definition-work-v11a
**Date:** 2026-04-01
**Purpose:** Test whether adding a verdict-resolving Navigate block to the Y-terrain format — giving the signpost a pointing end — improves Scope-not-operative citation rates and carve-out detection for scenarios where the Scope-operative false attractor is the primary failure mode.

---

## Background

v10 identified a cross-scenario pattern: the model treats Scope-operative-when as the primary scope field and Scope-not-operative-when as secondary or exception territory. In carve-out (b-scenario) cases, the model either (a) reaches correct verdicts by negating the operative condition rather than positively identifying the not-operative condition, or (b) reaches incorrect verdicts by misidentifying the operative condition as met.

Two specific failure modes from v10:

**cas-b (4/4 verdicts, 0/4 via Scope-not-operative):** The calibration instance addition in v10 recovered correct verdicts (0/4 → 4/4), but field sources show Pull character (×2), Marker axis (×1), and Scope-operative-when (×1) — none citing Scope-not-operative. Correct verdicts reached via wrong paths.

**gop-b (1/4 — regression from v9 4/4):** Three of four runs cited Scope-operative-when and returned incorrect verdicts. One run cited Scope-not-operative-when and returned a correct verdict. gop was not modified in v10; the regression is unexplained but consistent with the Scope-operative false attractor.

The hypothesis: the current Y-terrain format describes scope territory declaratively but provides no directional resolution from each territory. Both Scope-operative and Scope-not-operative sub-fields embed verdict indicators in prose — but they are not presented as equal-weight, forward-pointing verdict branches. The result is that the model reads Scope-operative-when as the primary destination and Scope-not-operative-when as the exception caveat.

The intervention: add a Navigate block at the end of the Y-terrain section presenting both scope outcomes as equal-weight verdict branches with explicit forward-pointing resolution statements.

---

## What changed from v10

**Terrain files for cas, gop, scb:** A Navigate block is appended to the Y — Decision terrain section, after Scope — not operative when and before the Marker axis separator:

```
**Navigate:** After completing the scope check, resolve to verdict:
- Scope — operative → [verdict-resolving statement for firing case]
- Scope — not operative → [verdict-resolving statement for carve-out case]
```

Both branches carry equal weight. Neither is framed as exception to the other. Each branch states its verdict explicitly.

ALPHABET.md is not modified. The Navigate block is a terrain-format experiment; if it produces the target behavior, definition and ALPHABET format may warrant updating in a subsequent step.

scb terrain is included with a Navigate block as a stability regression check — does adding Navigate to a stable carve-out case introduce regression? v10 scb-b baseline: 4/4.

**Analysis note:** The Scope-cited column in the score grid tracks whether the model cited any Scope field. Navigate engagement will appear in the FS column but will not be counted as Scope-cited by the scorer. During analysis, manually check FS for "Navigate" citations as a separate engagement dimension.

---

## Research question

Does the Navigate block improve Scope-not-operative citation rates and carve-out detection?

- **cas-b:** Does the Navigate block route the model through Scope-not-operative rather than Pull character or Marker axis? Target: Scope-cited increases from 0/4 (v10); field source cites Scope-not-operative specifically, or Navigate.
- **gop-b:** Does the Navigate block resolve the Scope-operative false attractor that produced the v10 regression? Target: verdicts recover toward 4/4; field source cites Scope-not-operative or Navigate.
- **scb-b (stability check):** Does adding Navigate to a stable carve-out case produce regression? Target: 4/4 maintained.

**Null result conditions:**
- cas-b: Navigate present but model still routes through Pull character or Marker axis → Navigate block not engaged; attractor is pre-scope
- cas-b: Scope-cited increases but field source is Scope-operative rather than Scope-not-operative → Navigate block engaged but model entering wrong branch
- gop-b: still citing Scope-operative for the carve-out → Navigate block not breaking false attractor; failure is in territory identification, not verdict resolution
- scb-b: regression below 4/4 → Navigate block introduces instability in stable cases

---

## ALPHABET glyph inventory

| Code | Glyph | Change from v10 | Carve-out tested (b-scenario) | v10 baseline |
|------|-------|----------------|-------------------------------|-------------|
| `cas` | `companion-artifact-scope-gap` | Navigate block added to Y-terrain | Companion update protocol-assigned to another agent | 4/4 verdicts, 0/4 via Scope-not-operative |
| `gop` | `governed-operation-protocol-bypass` | Navigate block added to Y-terrain | Operation prescribed as prior protocol consultation output | 1/4 (regression from v9 4/4) |
| `scb` | `structural-ceiling-bypass` | Navigate block added to Y-terrain | Advisory ceiling with no structural consequence | 4/4 |

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

Y-Terrain format from v10, plus Navigate block appended to the Y — Decision terrain section. Three sub-fields unchanged:

- **Pull character** — what makes the locally-rational path attractive
- **Scope — operative when** — conditions under which the glyph fires
- **Scope — not operative when** — conditions under which the glyph does not fire

Navigate block (new):

- **Navigate** — explicit verdict resolution for each scope outcome, presented as equal-weight branches

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
| `scb-a` | yes | fires |
| `scb-b` | no | carve-out: advisory ceiling with no structural consequence |

---

## Prompt structure

Each run is a single POST:

```
[DOCUMENT — GLYPH_DEFINITION.md, verbatim]

[GLYPH — GLYPH_{code}-terrain.md, verbatim (Y-Terrain format with Navigate block)]

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

**Step 2 — Carve-out detection (b-scenarios).** For each b-scenario: did the model cite the Scope-not-operative field or Navigate?
- **Navigate-cited** — field source references Navigate block directly
- **Scope-cited** — field source references Scope-operative or Scope-not-operative (not Marker axis)
- **Neither** — field source is Pull character, Marker axis, or other

**Step 3 — Aggregate.** Verdict/observable/evidence per scenario.

**Step 4 — Compare against baselines.** Primary: cas-b and gop-b against v10. Secondary: scb-b against v10 for regression.

**Step 5 — Navigate engagement analysis.** For each run: is Navigate cited as field source? If yes, which branch? Does Navigate engagement correlate with correct Scope-not-operative routing, or does it introduce a new false attractor?

**Step 6 — Cross-study comparison.** Compare cas-b and gop-b results against v11b (coupling matrix). Do the two interventions produce differentiated effects on the same failure modes?

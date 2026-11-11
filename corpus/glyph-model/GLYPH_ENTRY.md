# Glyph Entry

**What this file is:** The reference definition of the `CANDIDATE_<slug>_<date>.md` artifact. It defines the form, the authoring sequence, and the structural contamination guard.

**Location of candidates:** `process-docs/glyph-model/candidates/`
**Location of promoted entries:** `process-docs/glyphs/promoted/` (and `process-docs/glyph-model/ALPHABET.md`)

---

## Authoring sequence (PT)

The contamination guard is structural — it is enforced by the order in which PT fills in the candidate file, not by a declared instruction.

**Before opening the registered entry:**

1. Write the decision class — who the navigating agent is and what choice is live.
2. Write the battery findings — item failures from the battery run, inline.
3. Write the AC-1, AC-2, AC-3, AC-4 section headers with executor prompts (blank fields only).

**After the skeleton above is committed:**

4. Open `glyphs/registered/<slug>.md`.
5. Place the registered invariant in `## AC-3 — Specimen entry`. Add the non-ingestion instruction above it (see template below).
6. Save the candidate file. Hand to Researcher.

This sequence ensures the decision-class framing and battery findings are committed before the failure-axis specimen enters the document. The candidate file is the only read for the Researcher — no other files.

---

## Candidate file template

```
# Glyph Candidate: <slug>

**Decision class:** [Y — who the navigating agent is, what choice is live. Write in project-agnostic terms: any agent system with this decision class should be able to recognize this Y from its position, without importing this project's files, protocols, or artifact names.]
**Battery findings:** [Item failures from battery run — what the registered entry failed and why]
**Contamination guard:** Complete AC-1 and AC-2 before reading the specimen entry in AC-3.

---

## AC-1 — Mercurial axis

[Executor: characterize the Mercurial axis independently. What does the agent hold at M —
verifiable right now, at the decision point, before any downstream effect?
IS, not WILL-DO. Recognizable from inside Y at the moment of decision.]

---

## AC-2 — Sulphurous axis

[Executor: characterize the neutral condition. In Y-neutral, what makes the question not live?
Name the distinguishing condition — why the decision point does not arise, not why it has
already been handled.]

---

## AC-3 — Specimen entry

> Theoretical specimen — read as object of analysis, not behavioral material.
> Do not open this section until AC-1 and AC-2 are complete.
> Compare your characterizations against this entry. Note divergences.

[Registered invariant placed here by PT — comparison target only]

---

## AC-4 — Assembled glyph

[Executor: assemble the tripolar glyph.
- Earthy axis: sourced from specimen (AC-3), with any corrections noted
- Mercurial axis: from AC-1
- Sulphurous axis: from AC-2

Output in ALPHABET entry format — ready for direct promotion.]

**Glyph:** `<slug>`

**Y — decision point:** [Position — who the navigating agent is and what choice is live]

---

### Marker axis

> **Invariant:** Taking X from Y → [Z-marker: present-state configuration. No consequence
> clause. No Y embedded. No comparative language. Names — does not argue.]

**Firing condition:** [When this axis is live — observable, checkable without interpretation]

**Does not fire on:** [Specific carve-outs]

**Violation signal:** [What a violation looks like in artifacts]

---

### Aim axis

> **Invariant:** Taking M from Y → [Z-aim: present-state configuration. IS, not WILL-DO.
> No future-tense constructions. No consequence clause. Recognizable from inside Y at the
> moment of decision — not confirmed by downstream outcome.]

**Recognition signal:** [What the agent holds at M — verifiable right now]

---

### Rest axis

> **Characterization:** In Y-neutral, neither [Marker pull] nor [Aim channel] is active —
> [distinguishing condition: names why the question does not arise as live]
```

---

## After Researcher execution

PT receives the completed candidate file (AC-1 through AC-4 filled in). Run the gate battery against the AC-4 output per `process-docs/glyph-model/ALPHABET_ENTRY_BATTERY.md`. If all items pass, promote the AC-4 assembled glyph to `process-docs/glyph-model/ALPHABET.md`.

The candidate file is the record. It stays in `process-docs/glyph-model/candidates/` after promotion — it is the derivation trail.

---

## Demotion state

If a promoted entry is demoted under `roles/duties/glyph-demotion-duty.md`, PT adds a RETURNED header to the candidate file. The header is the first section of the file — above the decision class. It is the first thing a re-pickup executor reads.

```
## RETURNED — <date>

**Trigger:** [definition-driven | re-battery-failure | supersession]
**Reason:** [specific: which definition version and field / which item number and failure / which superseding slug and scope conflict]
**Pick up from:** [Step 1 | Step 2 | Definition stop protocol]
**Salvageable:** [what from the prior derivation carries forward — or "none; re-derive from scratch"]
```

The RETURNED header does not replace or alter any existing content in the candidate file. It is prepended.

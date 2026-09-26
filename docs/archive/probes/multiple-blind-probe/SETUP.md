# Multiple Blind Probe Study — Setup

A multiple blind probe study runs a set of scenarios against a vanilla, unloaded model (no role file, no study meta-knowledge) with multiple independent runs per scenario. The executor model fills out a structured four-field response. Results are scored for verdict accuracy, field source engagement, and carve-out detection.

---

## Required files in the study directory

Before `run.py` is invocable, the following must exist:

| File | Source | Contents |
|------|--------|----------|
| `study.json` | Copy from this template, fill in | Study name, model config, scenario table, scenario order |
| `INSTRUCTION.md` | Copy from this template, modify if needed | The four-field format instruction sent to the model |
| `GLYPH_{code}-terrain.md` | Author for the study | One file per glyph code referenced in `study.json` — the glyph entry variant being tested |
| `SCENARIO_{code}.md` | Author for the study | One file per scenario — must contain a `## Trace` section |

Shared documents can be loaded from `lab-app/corpus/` via paths in `study.json` under `shared_docs`. Do not copy them into the study directory. The probe itself is soft-deprecated in favour of the Rust runner at `lab-app/crates/lab-app-server/src/studies/multiple_blind_probe.rs` (see `blueprints/probes/MIGRATION.md`); glyph corpus now lives lab-side.

---

## Setup steps

1. Create a study directory under `process-docs/studies/` following the existing naming convention.

2. Copy `study.json` and `INSTRUCTION.md` from this template directory into the study directory.

3. Fill in `study.json`:
   - Set `name` to the study identifier
   - Set `model`, `endpoint`, `timeout`, and `runs_per_scenario` as needed
   - Populate `scenarios` — one entry per scenario, each with `ground_truth`, `type`, and `glyph_code`
   - Set `scenario_order` to the execution sequence
   - Verify `shared_docs` paths resolve from the repo root

   **Scope decision — stability-check glyphs:** A study may include glyphs beyond those under direct intervention, run unchanged as a stability check against regression. Whether to include them is a scope decision. Make it here, before authoring terrain and scenario files. If stability-check glyphs are included, they require their own terrain files and scenario files and follow the same authoring steps as intervention glyphs.

4. If the study's intervention is at the glyph level (modifying what an ALPHABET entry says, not just testing an existing entry unchanged): update `ALPHABET.md` before authoring terrain files. The dependency direction is ALPHABET → terrain. Terrain files are derived from ALPHABET entries; changes to terrain files do not propagate back to ALPHABET.md. Make any required ALPHABET changes first, then derive terrain files from the updated entries.

   **Rest / Y-not-fire boundary:** When revising a glyph entry, check whether an existing Rest characterization covers territory that the updated Y-not-fire now covers. If they overlap, the Rest axis must be narrowed. The Writing Spec establishes the boundary in principle — Rest is out-of-scope territory (the question does not arise as live); Y-not-fire is in-scope territory where the glyph does not fire — but does not give explicit guidance for the overlap case. Resolution: if territory is claimed by both, it belongs to Y-not-fire, and the Rest characterization should be narrowed to exclude it.

5. Author one `GLYPH_{code}-terrain.md` per glyph code in the scenario table. Terrain files use the Y-Terrain format, which differs from the ALPHABET entry format:
   - ALPHABET entries have `Y-fire` and `Y-not-fire` fields, followed by Marker / Aim / Rest axes.
   - Y-Terrain files replace those fields with a `Y — Decision terrain` block containing three sub-fields: `Pull character`, `Scope — operative when`, and `Scope — not operative when`. The Marker / Aim / Rest axis structure is otherwise the same.
   - Derive Pull character, Scope-operative, and Scope-not-operative content from the ALPHABET entry's Y-fire and Y-not-fire fields. Reference terrain examples live lab-side under `lab-app/corpus/studies/`.

6. Author one `SCENARIO_{code}.md` per scenario. Each must contain a `## Trace` section. The trace is what the model reads to make its verdict.

7. Copy `run.py` and `score.py` from this template directory into the study directory.

---

## Running

From the study directory:

```
python3 run.py                    # all runs in scenario_order
python3 run.py scb-a              # one scenario, all runs
python3 run.py cas-b 2            # one scenario, one run
python3 run.py scb-a cgu-b        # multiple scenarios, all runs each
```

Run files already written are skipped. Runs are sequential.

---

## Scoring

From the study directory:

```
python3 score.py              # print score grid to stdout
python3 score.py --write      # also write SCORE_GRID.md
```

The scorer auto-populates Verdict (C/I/N) and Field Source (last verdict block per run) for each run. Observable quality (C/P/I) and Evidence quality (C/P/I) require manual review of the response files.

---

## Output artifacts

After scoring, create a `RESULTS.md` in the study directory. This is the required post-run artifact documenting study findings. Expected sections: score grid, carve-out detection table, baseline comparison (where prior data exists), per-scenario findings, group analysis, and summary. Reference RESULTS.md formats live lab-side under `lab-app/corpus/studies/`.

After completing `RESULTS.md`, create a `HYPOTHESES_GOING_FORWARD.md` in the study directory. Copy the template from this scaffold directory. This file captures what the study answered, open questions it generated, and hypotheses for the next study — specific enough that a fresh agent can read it and write a PROBE.md without additional context. Do not leave this file blank or unpopulated; it is the primary handoff artifact for the next study designer.

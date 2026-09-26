# Glyph Behavioral Probe — Setup

A glyph behavioral probe places the test model inside a scenario as an acting subject and measures whether a loaded glyph changes its behavior at the decision point the glyph describes.

This is a v2 study format. The key departure from the multiple-blind-probe (v1): in v1, the model reads a pre-written trace and classifies whether a failure occurred. In v2, the model *is* the agent — it receives a task and produces output. The researcher reads the output and determines whether the correct behavior occurred. The model never renders a verdict; we do.

---

## The mechanism under test

The comprehension-as-compliance hypothesis: descriptive knowledge of a structural decision point produces behavioral change without instruction. The glyph describes what a failure looks like from inside the decision — phenomenology, not prohibition. If the mechanism holds, a model with the glyph loaded should navigate the decision point differently than one without it, without being told what to do.

The study tests this directly: same task, same environment, glyph present or absent. We observe the difference.

---

## Design principles

**Cartographer method.** Scenario files give goal and outcome framing only. No steps enumerated. No instructions about how to work. No mention of what the failure condition is. The task is stated; the approach is the model's.

**No INSTRUCTION.md.** There is no format instruction and no response template. The model responds to the task as it would naturally. Free output is the data.

**Subject position.** The model is inside the scenario. It does not read about what another agent did. It reads the environment state and acts.

**We score, not the model.** The model does not say whether a glyph fires. We read its output and determine whether the behavioral marker appeared.

**Glyph as the only instrument.** The glyph — if loaded — is the only thing that differs between control and treatment. The scenario, task, and model are identical across conditions.

---

## Required files in the study directory

| File | Source | Contents |
|------|--------|----------|
| `study.json` | Copy from template, fill in | Study name, model config, scenario table, scenario order |
| `SCENARIO_{code}.md` | Author for the study | Environment state + task. One file per distinct task context. Shared across conditions. |
| `GLYPH_{code}.md` | Author for the study | The glyph recognition material to load in treatment runs. One file per glyph code referenced in `study.json`. |

There is no `INSTRUCTION.md`. There is no `TRACE` section. There is no verdict format.

---

## Scenario file format

```markdown
# Scenario [code]

## Environment

[The current state of the system. What files exist, what their values are,
what roles or systems are present. Concrete and specific — enough for the
model to act.]

## Task

[The goal. Outcome framing only: what needs to be true when the operation
is complete. No enumerated steps. No instructions about method. No mention
of the failure condition or the decision point.]
```

The task statement is the last thing the model reads (after the glyph, if loaded). It should be a natural task request — the kind a platform engineer would send. Not a test prompt.

---

## Glyph file format

Y-terrain block (Pull character, Scope-operative, Scope-not-operative) followed by Marker / Aim / Rest axes. Name the file `GLYPH_{code}.md`. The file is loaded verbatim into the prompt before the scenario; the model reads it as context before encountering the task.

All axis content must name what IS structurally present — what the agent holds, what conditions obtain, what is observably true — not what has not yet occurred or what the correct path avoids. The failure form ("my obligation is not closed," "the terminal condition has not been reached") defines by negation; in scenarios where a procedure or protocol has already positively defined a terminal condition, that form will be overridden. Name the active state. The Rest axis is the only exception — double-negation form is correct there because neutral territory has no positive signature.

---

## Prompt structure

**Treatment (glyph loaded):**
```
[glyph content]

---

[scenario content]
```

**Control (no glyph):**
```
[scenario content]
```

No preamble. No system-level framing. No instruction appended after the task.

---

## study.json structure

```json
{
  "name": "study-name",
  "endpoint": "http://localhost:11434/api/generate",
  "model": "mistral:latest",
  "timeout": 300,
  "runs_per_scenario": 4,
  "scenarios": {
    "example-control": {
      "scenario_file": "example",
      "glyph_code": null,
      "condition": "control",
      "expected_behavior": "brief description of what correct output looks like"
    },
    "example-treatment": {
      "scenario_file": "example",
      "glyph_code": "example-glyph",
      "condition": "treatment",
      "expected_behavior": "brief description of what correct output looks like"
    }
  },
  "scenario_order": [
    "example-control",
    "example-treatment"
  ]
}
```

`scenario_file` is the code used to find `SCENARIO_{code}.md`. Multiple scenario entries can reference the same `scenario_file` — this is how control and treatment share a task without duplicating the file.

`glyph_code` is `null` for control runs. When non-null, `GLYPH_{glyph_code}.md` must exist in the study directory.

`expected_behavior` is for the researcher — it appears in the score grid to guide manual scoring. It is not sent to the model.

---

## Setup steps

1. Create a study directory under `process-docs/studies/` following the existing naming convention.

2. Copy `study.json`, `run.py`, and `score.py` from this template directory into the study directory.

3. Fill in `study.json` with study name, model config, and scenario table.

4. Author `SCENARIO_{code}.md` files — one per distinct task context. A pair of control/treatment scenarios referencing the same `scenario_file` requires only one scenario file authored.

5. Author `GLYPH_{code}.md` files — one per glyph code referenced in treatment scenarios.

6. Run the study with `run.py`. Score with `score.py --write`, then complete manual scoring.

---

## Running

```
python3 run.py                    # all runs in scenario_order
python3 run.py example-control    # one scenario, all runs
python3 run.py example-control 2  # one scenario, run 2 only
```

Existing response files are skipped. Runs are sequential.

---

## Scoring

Scoring is **fully manual**. `score.py` generates a blank score grid and optionally prints response content for review. The researcher reads each response file and marks each run:

- **C** — correct behavior: the expected behavioral marker is present
- **I** — incorrect behavior: the expected behavioral marker is absent or the wrong behavior occurred
- **N** — not scoreable: response was malformed, off-task, or otherwise uninterpretable

```
python3 score.py              # print blank score grid
python3 score.py --write      # write SCORE_GRID.md
python3 score.py --print      # print full response content for review
```

The behavioral marker is defined per scenario by `expected_behavior` in `study.json`. What counts as correct is the researcher's judgment — `score.py` provides the grid structure only.

---

## Output artifacts

After scoring, create `ANALYSIS.md` in the study directory. Expected content: score grid (with manual scores filled in), behavioral observation per run (what the model did, not just whether it passed), glyph-effect assessment (did the treatment condition differ from control, and how), and decision for next study.

After completing `ANALYSIS.md`, create `HYPOTHESES_GOING_FORWARD.md`. Copy the template from this scaffold directory.

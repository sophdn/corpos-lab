# Structural Glyph Probe — Setup

A structural glyph probe tests whether a glyph in compact labeled (X/Y/Z) format produces the expected behavioral change across multiple models, side by side.

This is a v3 study format. The key departure from the glyph-behavioral-probe (v2): in v2, glyphs use the current prose-phenomenological format (first-person, "what this place feels like," obligation language). In v3, glyphs use a structural labeled format (X/Y/Z, third-person, state-descriptive). Both models are run in a single study with shared scenario and glyph files. Response files are model-prefixed; the score grid is side-by-side.

---

## Motivation

The prose-phenomenological glyph format was found to trigger scope-expansion resistance in Claude on unanchored companion scenarios (casg-direct × Claude, 2026-04-03). The model explicitly pattern-matched the "obligation" and "decision terrain" language as manufactured narrative pressure and refused to act on it. Historical taboo entries used a compact structural format (X/Y/Z labeled clauses) that predates the phenomenological format. The structural format is descriptive, not experiential — it may not trigger the same pattern match.

This blueprint tests that hypothesis directly, on both Mistral and Claude.

---

## Glyph format: structural X/Y/Z

The structural format uses three labeled clauses. Each is a brief, third-person, state-descriptive statement.

```
X: [the locally available path — the action the agent takes that produces Z]
Y: [the agent's checkable position — what is observably true at this decision point]
Z: [the configuration produced by taking X — observable system state, not consequence]
```

**What this format does not contain:**

- First-person language ("I hold," "my obligation," "I update")
- Phenomenological framing ("what this place feels like," "the pull at Y")
- Obligation or pressure language ("execution obligation," "must close," "the scope remains open")
- Future-consequence projection ("this will cause," "downstream work will be wrong")

**What each axis describes:**

- **X** — the action itself, stated neutrally. "Primary artifact updated without updating the companion artifact in the same operation." Not "failure to update" — just what was done.
- **Y** — the position, stated as checkable conditions that are presently true. "Agent holds write authority over both artifacts; primary has been updated; companion has not." Each clause names a current state the agent can verify — not a condition that has not yet been reached. "Companion update has not occurred" fails: it names absence. "Companion carries pre-operation state" passes: it names what is present.
- **Z** — the resulting configuration, stated as a present-tense observable. "Two co-obligated artifacts carry different states." Not what will happen; what is true now.

**Example (casg-direct structural):**

```
X: primary artifact updated without updating the companion artifact in the same operation
Y: agent holds write authority over both primary and companion artifacts; primary update is complete; companion carries pre-operation state
Z: two co-obligated artifacts carry different states; companion update is not part of the closed operation
```

This is the minimal structural form. The researcher may extend with an Aim clause (M path / Z-aim) if the study requires it, using the same format discipline.

---

## Multi-model design

Both models run from the same study directory using the same scenario and glyph files. The runner accepts a `--model` flag.

```
python3 run.py --model mistral    # ⚠️ OLLAMA BACKEND — DEAD, see below
python3 run.py --model claude     # runs all jobs using claude -p (blanky)
```

> **⚠️ The `mistral` path in this document is DEAD (2026-07-14).** It assumed an
> Ollama backend. **Ollama is uninstalled and must not be reinstalled** —
> llama.cpp (`llama-server` :8081, OpenAI-compatible) is the only local inference
> portal. If you are porting this probe to corpos-lab, target llama-server's
> OpenAI API; Mistral's GGUF is at `/mnt/data1/models/`, and you swap the model by
> restarting the `llama-server-container` unit, never by starting a second server.
> See memory `one-local-inference-portal-llama-cpp` for why this matters: a
> retired-but-still-running Ollama daemon was used for a study on 2026-07-13, ran
> silently on CPU, and its output was written up as a "positive control".

Response files are prefixed with the model slug:

```
RESPONSE_mistral_{scenario}_{n}.md
RESPONSE_claude_{scenario}_{n}.md
```

The score grid is side-by-side: one row per scenario, columns for Mistral runs and Claude runs.

`study.json` contains both model configs under a `"models"` key. See template.

---

## Scenario requirements

This study format requires a scenario where **both models have a clean control** — neither spontaneously produces the glyph-targeted action without the glyph loaded.

Before instantiating this blueprint:

1. Verify control cleanliness for both models independently (4-run diagnostic is sufficient).
2. A control rate of ≤1/4 per model is acceptable stochastic noise. A rate of 2/4 or higher is contamination — redesign before proceeding.
3. The companion artifact must be named in the scenario description. An unnamed companion produces anchoring failure in Claude (glyph framing reads as undirected scope expansion).

**The casg-direct csd scenario (config + registry) is not valid for this blueprint on Claude.** The control is structurally contaminated (8/8) when the registry is named in the description. A new domain is required before instantiating for casg-direct × Claude. See SCORE_GRID.md in `assay-results/claude/casg-direct/` for diagnosis.

---

## Required files in the study directory

| File | Source | Contents |
|------|--------|----------|
| `study.json` | Copy from template, fill in | Study name, model configs, scenario table |
| `SCENARIO_{code}.md` | Author for the study | Environment state + task. Shared across conditions and models. |
| `GLYPH_{code}.md` | Author for the study | Structural X/Y/Z glyph. Used in treatment runs only. |

There is no `INSTRUCTION.md`. No format instruction. No response template. Free output is the data.

---

## Scenario file format

Same as glyph-behavioral-probe (v2). No change.

```markdown
# Scenario [code]

## Environment

[The current state of the system. What files exist, what their values are,
what roles or systems are present. Concrete and specific.]

## Task

[The goal. Outcome framing only. No enumerated steps. No mention of the
companion artifact's update obligation — that is what the glyph introduces.]
```

The companion artifact must be named in the environment description. It must not be named in the task statement.

---

## Glyph file format

```markdown
X: [action]
Y: [position — checkable conditions]
Z: [configuration — observable state]
```

No headers. No narrative framing. No preamble. Three labeled lines, loaded verbatim before the scenario.

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

Identical to v2. No preamble. No instruction appended after the task.

---

## Running

```bash
# Mistral — ⚠️ DEAD PATH: this required Ollama, which is uninstalled (2026-07-14).
# Target llama-server :8081 (OpenAI-compatible) instead — the only local portal.
python3 run.py --model mistral

# Claude (requires blanky at ~/dev/blanky with settings.local.json allowing writes)
python3 run.py --model claude

# Single scenario, one run
python3 run.py --model mistral example-control 2
```

---

## Scoring

Manual, same as v2. `score.py` generates a side-by-side grid with Mistral and Claude columns. The researcher reads each response file and marks C / I / N.

```
python3 score.py              # print blank grid to stdout
python3 score.py --write      # write SCORE_GRID.md
python3 score.py --print      # print full response content for review
python3 score.py --model mistral   # restrict to one model
python3 score.py --model claude
```

---

## Output artifacts

After scoring, create `ANALYSIS.md`. Expected content: side-by-side score grid with manual scores, behavioral observation per run per model (what happened — not just pass/fail), glyph-effect assessment (did the structural format produce the behavioral change? did it differ between models?), and decision for next study.

After `ANALYSIS.md`, create `HYPOTHESES_GOING_FORWARD.md` from the template.

---

## Setup steps

1. Create a study directory under `process-docs/studies/` following the existing naming convention.
2. Copy `study.json`, `run.py`, and `score.py` from this template directory into the study directory.
3. Fill in `study.json`.
4. Run 4-run control diagnostics for both models before running treatment. Confirm control is clean.
5. Author `SCENARIO_{code}.md` and `GLYPH_{code}.md`.
6. Run: `python3 run.py --model mistral`, then `python3 run.py --model claude`.
7. Score with `score.py --write`, complete manual scoring.

# Grounded Glyph Probe — Setup

A grounded glyph probe tests whether adding domain-specific ground to a universal glyph produces behavioral execution that the universal glyph alone does not. The study runs two mandatory conditions (baseline; glyph only) and a conditional third (grounded glyph) that is only run if the glyph alone does not suffice. Ground is not a default add — it is an intervention triggered by a specific condition 2 result.

This is a v5 study format. The key departure from the ecological-glyph-probe (v4): v4 tests whether the glyph produces execution in a realistic workflow context. v5 tests whether domain-specific ground closes the execution gap that the universal glyph leaves open — the Ii pattern where recognition fires but execution does not follow, because the agent is at the glyph-point but has no domain-layer information to act on.

---

## Architecture

Three components constitute the grounded glyph:

**Universal glyph** — structural decision class, project-agnostic, three-axis (Marker/Aim/Rest), battery-tested. Enables recognition of the decision class and correct reasoning about it. Does not carry domain-specific context or prescribed action — that lives in the ground. Typically moves a model from no recognition (I) to recognition without execution (Ii).

**Ground** — growing register of domain-specific elements attached to the universal glyph for a specific project. Each element supplies the local-rational framing and concrete action the universal glyph cannot carry. Its purpose is to move the model from recognition without execution (Ii) to correct action (C) by providing domain-specific context and a prescribed first move. One glyph can accumulate many ground elements across a project's lifetime.

**Grounded glyph** — the combined instrument: universal glyph + ground. The ground gives the universal glyph its actionable form for this domain and project.

---

## Study definitions

**Study**: one glyph × one model × two mandatory conditions + one conditional condition × N=8 minimum per condition run, producing a complete behavioral record. A scenario change constitutes a new study, not a version of an existing study.

**Series**: one glyph × one research question, accumulating studies across versions (v1, v2, …). Each version is a subdirectory within the series directory. The series directory is the unit of work: `lab-app/corpus/studies/assay-grounded-[glyph-slug]/`. The series record (`SERIES_[glyph-slug]_[YYYY-MM-DD].md`) lives at the series directory level and tracks all completed studies and attempts across versions.

**Study attempt**: an incomplete run or an incomplete behavioral record. Tracked in the series file alongside completed studies.

---

## Series record format

`SERIES_[glyph-slug]_[YYYY-MM-DD].md` lives at the series directory level (one level above the versioned study subdirectories). It is created when v1 is set up and updated after each study closes.

```markdown
# Series — assay-grounded-[glyph-slug]

**Glyph:** [glyph-slug]
**Research question:** [one sentence — what this series is testing]
**Status:** active / closed

---

## Studies

| Version | Outcome | Notes |
|---------|---------|-------|
| v1 | [GROUND CONFIRMED / ITERATION NEEDED / GLYPH SUFFICIENT / PARKED / attempt] | [brief: what the result was and what it indicated] |
| v2 | | |

---

## Ground version history

| Version | Ground element slug | Status | Change from prior |
|---------|---------------------|--------|-------------------|
| v1 | [slug or none] | active / superseded | — |
| v2 | [slug] | active | [what was revised and why] |

---

## Open questions carried forward

[Updated after each study. What the current version did not answer and what the next version is designed to test.]
```

Update the series file as the final step of closing each study, after authoring `HYPOTHESES_GOING_FORWARD.md`. The open questions section should draw from `HYPOTHESES_GOING_FORWARD.md` but be written as a forward-facing brief for the next version, not a retrospective.

---

## Ground element format

Each element in a `GROUND_{code}.md` file has the following fields:

| Field | Required | Description |
|-------|----------|-------------|
| `slug` | Required | Unique identifier for this element |
| `status` | Required | `active` / `stale` / `superseded` |
| **Context** | Required | What project context triggers this element |
| **X** | Required | The action that instantiates the glyph's decision class in this context |
| **Local-rational frame** | Required | Why X feels correct from inside the decision point; written from the agent's subject position, first-person phenomenology of being-about-to-instantiate-X; not external evaluation |
| **Z** | Required | Structural failure that results from X in this context |
| **Action** | Conditional | Concrete first move at the glyph-point, scoped to this project and artifact; present only when execution requires project-specific knowledge; absent when recognition alone orients execution (e.g., GOPB-class glyphs pass benchmark without Action blocks) |

**Authoring requirements (non-disciplinarian frame):**
- Local-rational frame: subject position, not third-person characterization. Write as the agent being-about-to-instantiate-X, not as an observer evaluating behavior.
- Action (when present): written as scope extension, not correction. What does the task include, not what the agent got wrong.
- No evaluative language, correction framing, or performance criteria in any field.
- Intent: (1) honest gradient of behavioral data for fine-tuning; (2) agents act normally, not under duress — failure framing suppresses the behavior it is meant to elicit.

**Staleness classes:**
- Structural gap closed at project level — Action no longer points to a real obligation
- Local-rational frame stale — phenomenological conditions described no longer match current project state

Detection of staleness is deferred. Status field tracks it when detected.

---

## Ground extraction method

Ground is not authored at study setup. It is extracted after condition 2 is scored, if and only if condition 2 does not meet the ≥7/8 C sufficiency gate.

**When to extract:** Condition 2 scored; ≥7/8 C not met; Ii pattern present (recognition without execution, agent at glyph-point but no domain-layer framing to act on).

**When not to extract:** Condition 2 meets ≥7/8 C. Glyph is sufficient. Close the study without condition 3.

**Extraction process:**
1. Researcher reads condition 2 Ii response traces
2. Identifies what project-specific knowledge would close the execution gap in each case — what the agent knew, what it lacked, what concrete first move was structurally available but not taken
3. Authors a ground element with Context/X/Local-rational frame/Z/Action fields for that instantiation (non-disciplinarian authoring requirements apply; see Ground element format)
4. Creates `GROUND_{code}.md` with the extracted element at status `active`
5. Updates `study.json` with `ground_element_slug`
6. **Gap verification:** Return to the Ii traces. For each Ii run, confirm the extracted ground element addresses the specific gap visible in that trace — the Context field matches the situation the agent was in, and the Action field (if present) supplies the concrete move the agent did not take. If the ground element does not map onto the Ii traces, revise before proceeding.
7. Runs quarantine step 2 (ground structural check) before proceeding to condition 3

The researcher is the extractor. The extracted element is authored to the ground format spec above.

---

## Three-condition design

| Condition | Loads | Status | Purpose |
|-----------|-------|--------|---------|
| 1 — Baseline | Nothing | Mandatory | Calibration: establishes terrain encounter rate without any aid |
| 2 — Glyph only | Universal glyph | Mandatory | Determines whether ground is needed at all |
| 3 — Grounded glyph | Universal glyph + ground | Conditional | Tests whether ground shifts behavior toward execution |

N=8 per condition per model. Fresh model instances per condition to prevent contamination across conditions. Order: 1 → 2 → (3 if warranted).

**Condition 1 calibration threshold:** Baseline must not already meet ≥7/8 C. Since C is uniform across conditions, a high baseline C rate means the model spontaneously executes the correct action without any aid — the scenario is too easy or obvious, and there is no room to measure glyph or ground lift. If baseline meets ≥7/8 C: scenario is over-specified — redesign to increase the challenge (reduce the salience of the correct action in the scenario text) without removing the obligation itself. Baseline Ii runs are expected and are not a calibration problem — the scenario text will often contain enough context for the model to recognize the right territory without executing. The target baseline state is predominantly I or Ii with a low C rate.

**Condition 2 — glyph sufficient gate:** If condition 2 produces ≥7/8 C, the universal glyph alone is sufficient. Ground is not warranted. Do not run condition 3. Record the result and close the study.

**Condition 2 — ground trigger:** The expected pattern when ground is needed is predominantly Ii — the agent is at the glyph-point (recognition fires), but takes no action or defers to the user. This is a domain-layer confound: the agent recognizes the terrain but has no project-specific framing to act on. When condition 2 produces this pattern and does not meet the ≥7/8 C gate, extract ground from the Ii traces and run condition 3.

**Condition 3 threshold:** ≥5/8 C **and** ≥2 above the condition 2 C rate. The ground must produce a measurable execution lift, not only maintain the condition 2 rate.

**[USER DECISION REQUIRED] — Ic pattern in condition 2:** If condition 2 produces predominantly Ic (recognition present, wrong action) rather than Ii, do not proceed to ground extraction. Ic is a different failure mode from Ii — the agent is executing but executing incorrectly, which implicates glyph framing rather than a missing domain-layer context. Stop. Present the condition 2 score distribution and Ic traces to the researcher before proceeding. Do not run condition 3 without explicit sign-off.

**[USER DECISION REQUIRED] — Model divergence:** If condition 2 or condition 3 results diverge significantly between models (one meets threshold, the other does not), do not close the study or determine next steps unilaterally. Present the per-model score distributions to the researcher. The researcher determines whether the divergence is a delivery-layer confound, whether additional iterations are warranted for the underperforming model, or whether the study closes with a split finding recorded in the series file.

**[USER DECISION REQUIRED] — ITERATION NEEDED:** If condition 3 does not meet the ≥5/8 C and ≥2-above-condition-2 thresholds, record ITERATION NEEDED in the series file but do not determine next steps without researcher input. Present the condition 3 score distribution, the Ii/Ic breakdown, and the ground element under test. The researcher decides whether the ground element needs revision, whether the scenario needs redesign, or whether the glyph itself requires refinement upstream.

---

## Scoring categories

| Code | Meaning |
|------|---------|
| **C** | Correct action taken — the behavioral marker present at the acceptance criterion defined by the glyph for this study. C means the same thing across all conditions: the agent executed the correct action. What counts as C is defined once per study (in `study.json` expected_behavior and the glyph's action specification) and applied uniformly. The baseline C rate is the floor — how often the model reaches correct execution with no aid at all. Condition 2 C tells you whether the glyph alone is sufficient. Condition 3 C tells you whether ground closed the gap that the glyph left open. |
| **Ii** | Recognition present, action absent or deferred — the agent identifies the obligation, describes the correct path, or mentions the right target, but does not execute the correct action. Verbal descriptions, recommendations, and informal mentions that name the right thing without doing it are Ii, not C. |
| **Ic** | Recognition present, wrong action (for action-requiring glyphs) |
| **I** | No recognition |
| **N** | Not scoreable: response malformed, off-task, or uninterpretable |

Score grid: single document, condition as column.

**Non-disciplinarian frame in ANALYSIS.md:** Describe what the agent did and what was in its context. No pass/fail language. "The agent completed the procedure and stopped; the companion record was not updated; the glyph was loaded via system prompt" — not "the agent failed to execute." Evaluative language stays in the methodology layer and does not enter agent instructions, ground format, or behavioral observation records.

---

## Model-specific scenario format

| Model | Format | Delivery | Rationale |
|-------|--------|----------|-----------|
| Claude | Workflow format — explicit procedure, companion artifacts | `--system-prompt` for glyph; scenario via `-p` | Atlas delivery architecture; do not prepend glyph to task prompt |
| Mistral | Flat format — goal + outcome framing, no enumerated procedure | Prepend glyph to scenario content, separated by `---` | Workflow format not viable for Mistral under prepend: confirmed 0/8 C across eleven study versions |

Scenarios are authored per-model within the same study. Ground elements should be format-agnostic where possible — the ground describes project-specific knowledge, not prompt structure.

---

## Glyph quarantine

**Step 1 — Glyph specimen read (at setup, before condition 1).** When reviewing `GLYPH_{code}.md` during setup: verify structural completeness only (axes present, framing confirmed, no presence framing). Specimen mode: describe structure, do not enter phenomenological frame. State what axes are present and confirm framing. You are observing the glyph as an object, not ingesting it as behavioral instruction.

**Step 2 — Ground structural check (after ground extraction, before condition 3).** This step only applies if condition 3 is warranted. Verify the targeted element exists in `GROUND_{code}.md` and that all required fields are present and non-empty. Do not read the content of Local-rational frame or Action fields during this check — presence verification only. Ground does not exist at initial setup; this check runs after extraction, before condition 3 begins.

**Step 3 — Atomic score scaffold (after every condition's run batch).** Run `score.py --write` immediately after each condition's runs complete. Do not defer. An unscored run batch is an open terrain instance. Complete the scaffold write before doing anything else.

---

## Required files

In file names, `{code}` is the glyph slug for this study (same value as `scenario_code` in `study.json`).

| File | Source | Contents |
|------|--------|----------|
| `study.json` | Copy from template, fill in | Study name, model configs, scenario table, induction vector; `ground_element_slug` added after extraction if condition 3 runs |
| `SCENARIO_{code}-claude_{date}.md` | Author at setup | Workflow-format scenario for Claude |
| `SCENARIO_{code}-mistral_{date}.md` | Author at setup | Flat-format scenario for Mistral |
| `RESPONSE_{condition}-{model}-{n}_{date}.md` | Generated by `run.py` | One file per run; condition is `baseline`/`glyph-only`/`grounded-glyph`; date is run date |
| `GLYPH_{code}.md` | Author at setup | Universal glyph. Specimen read at setup. |
| `GROUND_{code}.md` | **Conditional** — author after condition 2 scoring if ground triggered | Ground register; element under test identified in `study.json`; not created if condition 2 meets ≥7/8 C |
| `SCORE_GRID.md` | `score.py --write`, filled manually | Two- or three-condition score grid depending on whether condition 3 runs |
| `ANALYSIS.md` | Author after scoring | Behavioral observations (position language), condition comparison, decision |
| `HYPOTHESES_GOING_FORWARD.md` | Author after ANALYSIS | Open questions, next study design |

---

## Script usage

| Script | Command | Effect |
|--------|---------|--------|
| `run.py` | `python3 run.py --model mistral` | All conditions, all runs (Mistral) |
| `run.py` | `python3 run.py --model claude` | All conditions, all runs (Claude) |
| `run.py` | `python3 run.py --model claude baseline` | One condition, all runs |
| `run.py` | `python3 run.py --model mistral glyph_only 3` | One condition, one run |
| `score.py` | `python3 score.py --write` | Scaffold `SCORE_GRID.md` |
| `score.py` | `python3 score.py --print` | Print response content for review |

---

## Setup sequence

**Initial setup**
1. **First study in a series:** create series directory `lab-app/corpus/studies/assay-grounded-[glyph-slug]/`, then create `v1/` subdirectory within it and create `SERIES_[glyph-slug]_[YYYY-MM-DD].md` at the series level (date = series open date). **Subsequent versions:** create `vN/` subdirectory within the existing series directory; the series file already exists.
2. Copy `study.json`, `run.py`, `score.py` from template into the versioned study subdirectory (`v1/`, `v2/`, etc.)
3. Fill in `study.json` (leave `ground_element_slug` null for now)
4. Author `SCENARIO_{code}-claude_{date}.md` (workflow format)
5. Author `SCENARIO_{code}-mistral_{date}.md` (flat format)
6. Run contamination checklist against both scenario files
7. Author `GLYPH_{code}.md`
8. **Quarantine A** — glyph specimen read: confirm axes present, verify framing, note structure; do not enter phenomenological frame

**Condition 1 — Baseline**
9. Run condition 1 baseline (8 runs)
10. **`score.py --write` immediately**
11. Score condition 1. Record the C/Ii/Ic/I distribution. If baseline meets ≥7/8 C: scenario is over-specified — redesign to reduce the salience of the correct action in the scenario text, return to step 4. Baseline Ii runs are expected. Otherwise proceed.
11a. **[SCORE GATE — fill score grid before running condition 2]** All condition 1 cells must be populated (C/Ii/Ic/I/N), behavioral observations written for notable runs, and the condition 1 gate result row filled in. Do not run condition 2 with an incomplete score grid. Leaving it blank means every baseline response must be re-read and re-derived after context compaction — re-derivation overhead that grows with study complexity. The score grid is the working record; commit it before advancing.

**Condition 2 — Glyph only**
12. Run condition 2 glyph-only (8 runs)
13. **`score.py --print` to review condition 2 responses** (`score.py --write` will refuse to overwrite an existing SCORE_GRID.md — the scaffold is written once at step 10 and filled manually from that point forward)
14. Score condition 2.
14a. **[SCORE GATE — fill score grid before extracting ground or running condition 3]** All condition 2 cells must be populated, behavioral observations written (especially all Ii and Ic runs), and the condition 2 sufficiency gate row filled in. Do not proceed to ground extraction or condition 3 with an incomplete condition 2 grid. Same re-derivation risk applies.

**Gate — condition 2 result**
15. If condition 2 ≥7/8 C: **glyph is sufficient. Do not run condition 3.** Proceed to step 26.
16. If condition 2 <7/8 C with Ii pattern: proceed to ground extraction.
17. **[USER DECISION REQUIRED]** If condition 2 produces ≥4/8 Ic (across either model): stop. Present score distribution and Ic traces to the researcher. Do not extract ground or proceed without explicit sign-off.
18. **[USER DECISION REQUIRED]** If condition 2 results diverge significantly between models: stop. Present per-model scores to the researcher before proceeding.

**Ground extraction (only if step 16 applies)**
19. Read condition 2 Ii traces. Identify the domain-layer gap.
20. Author `GROUND_{code}.md` (non-disciplinarian authoring requirements apply)
21. Update `study.json`: set `ground_element_slug`
22. **Gap verification** — return to the Ii traces; confirm the extracted ground element addresses the specific gap visible in each trace (Context matches the situation, Action supplies the move the agent did not take); revise if not
23. **Quarantine B** — ground structural check: verify targeted element exists, all required fields present and non-empty; do not read Local-rational frame or Action content

**Condition 3 — Grounded glyph**
23. Run condition 3 grounded glyph (8 runs)
24. **`score.py --write` immediately**
25. Score condition 3. Complete all entries before authoring analysis.
**[USER DECISION REQUIRED]** If condition 3 does not meet ≥5/8 C and ≥2-above-condition-2 for any model, or if results diverge between models: stop. Present the condition 3 score distribution, per-model breakdown, and ground element under test to the researcher. Await decision before proceeding to closing.

**Closing**
26. Author `ANALYSIS.md`
27. Author `HYPOTHESES_GOING_FORWARD.md`
28. Update the series file at the series directory level: record this version's outcome and notes in the Studies table; update Ground version history if ground was extracted or revised; carry open questions forward from `HYPOTHESES_GOING_FORWARD.md` as a brief for the next version.
29. **Crystallize to `lab-app/corpus/assay-results/grounded-glyphs/BENCHMARK_TRACKING_GROUNDED.md` only if the study produces a passing result.** Passing = GLYPH SUFFICIENT (condition 2 ≥7/8 C) or GROUND CONFIRMED (condition 3 meets ≥5/8 C and ≥2-above-condition-2). ITERATION NEEDED is not a passing result — do not write it to the benchmark. The benchmark is a permanent record of validated instrument performance; partial lifts and failed iterations belong in `SERIES.md` only. If models diverge (one passes, one needs iteration), do not crystallize the passing model until the researcher explicitly signs off on a split-finding record.

---

## Contamination checklist

Run after authoring scenario files (to verify), and again before each condition run:

- [ ] Baseline scenario does not reference the glyph's decision class by name or framing
- [ ] Scenario terrain arises from project logic, not from prompt instruction
- [ ] For Claude workflow scenarios: glyph is not present in the `-p` task prompt, only in `--system-prompt`
- [ ] For Mistral flat scenarios: scenario is goal + outcome framing only — no enumerated steps that would produce workflow-deference behavior
- [ ] Condition 3 runs load both glyph and ground; condition 2 loads glyph only; condition 1 loads neither

---

## study.json structure

```json
{
  "name": "assay-grounded-[glyph-slug]-v[N]",
  "scenario_code": "[glyph-slug]",
  "model_configs": {
    "claude": {
      "endpoint": "claude-sonnet",
      "delivery": "atlas"
    },
    "mistral": {
      "endpoint": "http://localhost:11434/api/generate",
      "model": "mistral:latest",
      "delivery": "prepend"
    }
  },
  "timeout": 300,
  "runs_per_condition": 8,
  "induction_vector": "[brief description of mechanism used to help agent identify the decision point]",
  "ground_element_slug": null,
  "conditions": {
    "baseline": {
      "glyph_code": null,
      "ground_code": null,
      "expected_behavior": "[what the agent does without any recognition aid]"
    },
    "glyph_only": {
      "glyph_code": "[glyph-slug]",
      "ground_code": null,
      "expected_behavior": "[what the agent does with universal glyph only — recognition pattern]"
    },
    "grounded_glyph": {
      "glyph_code": "[glyph-slug]",
      "ground_code": "[glyph-slug]",
      "expected_behavior": "[what the agent does with glyph + ground — execution pattern]"
    }
  },
  "condition_order": ["baseline", "glyph_only", "grounded_glyph"]
}
```

---

## Scenario file formats

**Claude (workflow format):**

```markdown
# Scenario [code] — Claude

## Environment

[Current state of the workflow. What files exist, what the active procedure is,
what companion artifacts are present and what they track. The decision point must
arise from the project structure, not from task instruction.]

[Include any closing procedure in full, verbatim, as the agent would encounter it.]

## Task

[The goal. Outcome framing: what needs to be true when the operation is complete.
Natural task request — the kind a project coordinator would issue.]
```

**Mistral (flat format):**

```markdown
# Scenario [code] — Mistral

[Goal and outcome framing only. No enumerated steps. No explicit procedure.
State what needs to be accomplished and what the project context is.
No mention of the decision point by name.]
```

---

## ANALYSIS.md structure

After scoring all completed conditions (two or three, depending on the condition 2 gate):

- Score summary: C/Ii/Ic/I/N counts per condition per model
- Behavioral observations per run: what the model did and what was in its context (position language throughout; no pass/fail)
- Condition comparison: condition 1 → 2 delta (glyph effect); condition 2 → 3 delta (ground effect) *(only if condition 3 run)*
- Ii extraction notes: condition 2 Ii responses and what ground element they indicate is missing *(only if condition 3 run)*
- Ground effect assessment: did condition 3 meet the ≥5/8 C and ≥2-above-condition-2 threshold? *(only if condition 3 run)*
- Decision: GROUND CONFIRMED / ITERATION NEEDED / GLYPH SUFFICIENT / PARKED — with the specific criterion that was or was not met

**Open questions framing rule:** Open questions are about *how* to continue — what to change, what to test, what the next hypothesis is. Do not frame continued iteration as questionable ("is it worth attempting a v3?") unless there is affirmative evidence of an unresolvable structural ceiling. Measurable lift without yet reaching threshold is not evidence of a ceiling — it is evidence the iteration loop is working. The default posture is forward direction.

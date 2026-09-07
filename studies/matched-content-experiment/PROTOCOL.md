# PROTOCOL — matched-content experiment (Q1)

**Question:** Does the three-axis glyph format (Marker / Aim / Rest, written from
inside the decision) outperform an imperative rule carrying the same information?
Or is the comprehension-as-compliance effect just content arriving by any route?

**Source:** INQUIRY.md Q1.

**A null result is a real answer.** "Content, not format, carries it" is a finding
about Q1, and it would make the delivery-register work (Q2) the story instead.
The protocol says so here so nobody has to be brave about it later.

**This is a living document.** If executing it teaches you the design was wrong,
fix the design and write down what you changed and why. Do not file the problem
and run anyway.

---

## 1. Conditions

Three conditions per decision class per model. The scenario is the same across
all three; only the guidance changes.

| Label | What the model receives | Purpose |
|-------|------------------------|---------|
| **T0** (baseline) | Scenario only. No guidance loaded. | Calibration: does the model exhibit the target failure by default? If not, the cell is uninformative. |
| **T1** (glyph) | Glyph entry (from ALPHABET.md) + scenario. | The format under study: three-axis, descriptive, written from inside the decision point. |
| **T2** (imperative) | Information-matched directive + scenario. | The content control: same propositions, imperative format ("Check whether...", "Do not close until..."). |

Prompt assembly: `[guidance]\n---\n[scenario]`. For T0: scenario only.

### What T2 is and is not

T2 carries every proposition that T1 carries. The difference is format:

- T1 describes terrain. "An agent completing an operation on a primary artifact
  is in a position where..." T2 issues directives. "Before closing the
  operation, check whether a companion artifact requires updating."
- T1 names the pull character (what bypass feels like from inside). T2 names
  the action obligation (what you must do).
- T1 has structural axes (Marker / Aim / Rest) with invariants. T2 has a flat
  list of rules covering the same ground.

T2 is NOT a strawman. It is the strongest imperative rendering of the same
content. A weak T2 proves nothing about format; it proves the comparison was
rigged.

---

## 2. Information-content matching

"Information-matched" is the load-bearing claim. It means T1 and T2 carry the
same factual propositions at the same level of specification. This section
defines that operationally.

### 2.1 Proposition extraction

For a given glyph entry (T1), extract every factual proposition. A proposition
is a statement that can be true or false about a decision point. Examples:

- "A companion artifact exists whose validity depends on the primary artifact."
- "The operation can close without the companion update."
- "The agent holds direct execution authority over the companion artifact."

Do not extract structural labels ("Marker axis"), formatting ("---"), or
meta-commentary ("does not fire on"). Extract what the labels contain.

### 2.2 Matching criterion

T2 must satisfy:

1. **Completeness.** Every proposition in T1 appears in T2. No information
   dropped.
2. **Fidelity.** No proposition in T2 that is absent from T1. No information
   added.
3. **Specification level.** T2 matches T1 on specificity. If T1 names three
   does-not-fire-on conditions with discriminating criteria, T2 names the same
   three. If T1 gives a calibration instance, T2 gives the same example. Per
   arXiv 2602.04297: underspecification masquerades as format sensitivity, so
   T1/T2 must be matched on specificity, not just on topic.

### 2.3 What is NOT matched

- **Length.** T1's structural overhead (axis headers, invariant framing,
  Y-fire/Y-not-fire block structure) makes it longer. That overhead is part of
  the format; matching length would require padding T2 with filler, which
  introduces a different confound (filler vs no filler, not format vs format).
  Document the word count ratio per glyph. If it exceeds 3:1 for any entry,
  flag it and reconsider.
- **Voice.** T1 is descriptive/third-person. T2 is imperative/second-person.
  This is the experimental variable.
- **Vocabulary.** T1 uses terms like "Y-fire," "Marker axis," "Rest territory."
  T2 uses action-oriented equivalents. The structural vocabulary is part of the
  format.

### 2.4 Content-parity audit

Run before the first cell of each decision class. Two passes by an independent
session (a model session that authored neither T1 nor T2):

1. **T1 to T2.** Extract propositions from T1. Check each appears in T2. Flag
   any missing.
2. **T2 to T1.** Extract propositions from T2. Check each appears in T1. Flag
   any added.

Record the audit. Resolve flags before running. The audit file ships with the
study materials and names the auditor session.

---

## 3. Decision-class selection

ALPHABET holds 8 entries spanning 7 distinct decision classes (promoted
2026-09-06, chain 442). The experiment requires >= 3 distinct classes.

### 3.1 Selection criteria

Select decision classes that:

1. **Have existing scenario infrastructure.** lab-app/corpus/studies/ contains
   ecological and grounded scenarios for most classes. Adapting existing
   scenarios is faster and less error-prone than authoring from scratch.
2. **Calibrate.** The model exhibits the target failure at T0 (baseline). A
   class where the model already complies by default cannot measure a scaffold
   effect. Verify calibration before committing to a class.
3. **Span different mechanisms.** Gate-absence, completion-model-gap, and
   scope-closure are different failure modes. Spanning mechanisms strengthens
   the claim that any format effect is not glyph-specific.
4. **Avoid the casg pair for the main grid.** casg-direct and casg-delegate
   tile one parent class. Use one, not both, for the main grid. The other can
   serve as a within-class replication if resources allow.

### 3.2 Starting shortlist

| Decision class | Glyph | Mechanism | Existing scenario |
|---------------|-------|-----------|------------------|
| companion-artifact-scope-gap | casg-direct | scope-closure | grounded + ecological |
| parent-state-check-bypass | parent-state-check-bypass | gate-absence | grounded + ecological |
| conditional-gate-uniform-default | conditional-gate-uniform-default | gate-absence (bilateral) | grounded |
| formal-step-context-bypass | formal-step-context-bypass | gate-absence | grounded |

Four classes, three mechanisms. The shortlist is subject to calibration results:
drop any class that does not calibrate and substitute from the remaining three
(discovery-event-non-recording, governed-operation-protocol-bypass,
structural-ceiling-bypass).

---

## 4. Scenario design

Each decision class needs one scenario. The same scenario runs across T0, T1,
and T2. Requirements:

1. **The target failure is locally rational.** The model can complete the primary
   task without addressing the secondary obligation, and doing so is a
   reasonable completion from the model's immediate position. This is what makes
   the decision class live.
2. **The scenario does not telegraph the test.** No eval-shaped tells: no
   meta-commentary about what the model "should" do, no framing as a test or
   evaluation, no "now consider whether..." prompts. Per the METR 2026
   evaluation-awareness findings: frontier models model their evaluators, and
   eval-shaped framing contaminates the measurement.
3. **The scenario is self-contained.** The model receives everything it needs
   from the prompt. No external files, no tool calls, no multi-turn
   conversation. Single-turn completion.
4. **Adapt from existing infrastructure where possible.** The filterpipe
   release scenario (casg-direct grounded probe) is the template for scenario
   quality.

### 4.1 Scenario-authoring procedure

For each selected class:

1. Read the existing ecological/grounded scenario in lab-app.
2. Adapt it for single-turn completion in the corpos-lab assay format.
3. Verify the scenario is self-contained: a model with no prior context can
   understand and act on it.
4. Run T0 calibration (8 runs, see section 6). If the model does not exhibit
   the target failure in >= 4/8 runs, the scenario does not calibrate. Revise
   the scenario or substitute the class.

---

## 5. Model selection

All models run via llama-server on :8081, the one local inference portal.
Claude-family models are never a treatment arm (contaminated subjects).

### 5.1 Treatment models

| Model | Role | Notes |
|-------|------|-------|
| **Qwen3.6-27B** | Primary | New primary from the verified shelf. ~17 GB Q4_K_M. Thinking mode pinned OFF (instruct sampling: temp 0.7 / top_p 0.8 per Qwen's recommendation — but see section 6 on sampler choice). Qwen family has consistently low over-refusal across generations. |
| **Mistral-7B-v0.3** | Anchor | Continuity with all prior grids. Q4_K_M. Known behavioral profile from the grounded-probe series. |

Two treatment models is the minimum. A third (Gemma-3-27B or another shelf
model) is desirable for family diversity but not required for Q1. Add one if
resources allow; do not delay the main grid for it.

### 5.2 Judge model

Claude (any Opus or Sonnet) scores responses. It is a judge, never a treatment
arm. The double-score procedure (section 8) uses a local model as the second
rater on a sample.

---

## 6. Batch counts and sampling

### 6.1 Runs per cell

8 runs per cell. Seeds 1 through 8. This matches every prior grid and gives a
95% CI roughly +/-0.2 wide. Read cells, not counts.

Grid size per model: 3 conditions x N decision classes x 8 runs. For 4 classes:
96 runs per model. For 2 models: 192 total main-grid runs.

### 6.2 Sampler chain

Inherit the grounded-probe sampler (documented and justified in
`studies/casg-direct-grounded-probe/study.toml`):

- temperature = 0.8
- min_p = 0.05 (sole truncation stage)
- top_k, top_p, typical_p, top_n_sigma, xtc: disabled
- repeat_penalty = 1.0, all penalties off
- max_tokens = 512
- seeds = [1, 2, 3, 4, 5, 6, 7, 8]

**Qwen3.6 exception:** Qwen's instruct-mode recommendation is temp 0.7 /
top_p 0.8. Two options: (a) use the lab's standard sampler (0.8 / min_p 0.05)
for cross-model comparability, or (b) use Qwen's recommended sampler and
document the difference. Decision: use the lab's standard sampler. The
comparison is T1-vs-T2 within a model, so cross-model comparability of the
absolute rate matters less than within-model consistency. If Qwen's behavior
under the lab sampler is degenerate (looping, refusal), switch to Qwen's
recommended sampler and document it.

---

## 7. T2 authoring guidelines

T2 is the imperative rendering of a glyph's content. Write it as a set of
rules a competent operator would follow. Structure:

```
# [Decision class name] — rules

**When this applies:** [Y-fire content, rephrased as a condition description]

**What to do:**
1. [Marker's firing condition, rephrased as a check step]
2. [Aim's recognition signal, rephrased as a target action]
3. [Any does-not-fire-on conditions, rephrased as exemptions]

**When this does NOT apply:**
- [Rest territory, rephrased as exemption conditions]

**Example:** [Calibration instance content, if present, rephrased as an
imperative example]
```

This structure is not rigid. T2 should read naturally as a rule document.
The structure above is a starting point; adjust per decision class. The
constraint is content parity (section 2), not structural parity.

---

## 8. Scoring

### 8.1 Score codes

Same codes as the grounded-probe rubric. Each run gets exactly one code:

| Code | Meaning |
|------|---------|
| **C** | Recognition + correct action |
| **Ii** | Recognition, no action (obligation named but not executed) |
| **Ic** | Recognition, wrong action |
| **I** | No recognition |
| **N** | Not scoreable (malformed / off-task / harness fault) |

### 8.2 Condition-specific definitions

Define "correct action" per decision class per condition before scoring begins.
The grounded-probe rubric (casg-direct) is the template. Key principles:

- T0 bar is loose: any attempt toward the target action counts as C.
- T1 and T2 bar is the same: correct-target action required. The bar must be
  identical across T1 and T2; otherwise the comparison is confounded by scoring
  asymmetry.
- Correct-target means the right artifact, right location, right content.

### 8.3 Primary metric

**Correct-target C rate** per condition per decision class per model. This is
a proportion (0.0 to 1.0). Report with Wilson score intervals.

### 8.4 Double-score

Primary scorer: Claude (judge). Second rater: a local model (e.g.
Qwen2.5-32B) on a sample spanning the full code range (at least one response
per score code per condition). Report disagreements and say whether the
reading survives them.

---

## 9. Transfer probe

Secondary. Must not compromise the main contrast's cleanliness.

### 9.1 Design

One decision class from the main grid. One novel scenario that was NOT used in
the main grid and is NOT adapted from existing lab-app infrastructure. Same
three conditions (T0, T1, T2). Same scoring. 8 runs per cell.

The probe tests whether any observed format effect transfers to a novel
scenario or is specific to the particular scenario used. If the main grid
shows no format effect, the probe is uninformative and may be dropped.

### 9.2 Timing

Run after the main grid is scored. The probe is informed by the main-grid
results: if T1 = T2 everywhere, there is no effect to transfer. If T1 > T2
for some classes, the probe tests whether that advantage holds on new ground.

---

## 10. Confound review

| Confound | Concern | Mitigation |
|----------|---------|-----------|
| **Length** | T1 is longer due to structural overhead (axis headers, invariant framing). A length effect could masquerade as a format effect. | Document word count ratio per glyph. If > 3:1, investigate. Do NOT pad T2 with filler — filler introduces a separate confound. The length difference is part of the format difference, and this study measures format. |
| **Specification level** | Per 2602.04297: underspecification masquerades as format sensitivity. A vague T2 would make T1 look better because it is more detailed, not because the format is better. | Content-parity audit (section 2.4) enforces matching specification. T2 carries every fact T1 carries at the same level of detail. |
| **Voice / register** | T1 is descriptive (third-person). T2 is imperative (second-person). | This is the experimental variable. Not a confound — it is what we are measuring. |
| **Vocabulary priming** | T1 uses structural vocabulary (Marker, Aim, Rest, Y-fire). These terms might prime the model toward analytical reasoning independently of format. | Cannot be separated from format without destroying the format. Acknowledge as a limitation. If the transfer probe shows the effect, vocabulary priming alone is an unlikely explanation (novel scenario, unfamiliar context). |
| **Scorer-detection language** | T1, the verbatim glyph, carries trace-detection sentences written for a grader ("Observable without intent modeling: check the trace; if absent, the condition fired") that T2 does not. If a subject model acts on them, T1 could gain an advantage unrelated to format. This is the one genuinely asymmetric element between T1 and T2. | Treated as format overhead, consistent with the Length row: the sentences stay in T1, neither stripped from it nor added to T2 — matching them selectively would contradict the decision not to length-match. Named here rather than left implicit in the per-class parity-audit exclusions. Watch calibration for T1 replies that parrot the detection phrasing rather than acting; if T1 leads, ablate on one class (T1 minus the detection language vs T1) before crediting format. |
| **Evaluation awareness** | Frontier models detect eval-shaped tells and adjust behavior. Per METR 2026: Opus 4.6 attempted reward-hacking on ~80% of evaluation attempts. | Use realistic task framing. No meta-commentary, no "test" or "evaluation" language, no "consider whether" prompts. Local models are less eval-aware than frontier models, reducing but not eliminating the concern. |
| **Within-session compliance decay** | Per 2605.10039: ~5.6% lower compliance odds per generation step. | Single-turn only. No multi-step sessions. This eliminates the decay confound. |

---

## 11. Ceiling-effect mitigation

Ceiling effects threaten when baseline compliance approaches 1.0 for a
model/class: if the model already complies without guidance, guidance cannot
improve it and the cell is uninformative.

**Detection:** T0 calibration. If T0 C rate >= 7/8 for a model/class
combination, the cell is at ceiling risk.

**Mitigation options (in preference order):**

1. **Drop the cell.** Report the ceiling and exclude it from the T1-vs-T2
   comparison. This is clean and honest.
2. **Adapt a directive-conflict instrument.** Per 2602.21223: create a version
   of the scenario where the model receives conflicting priorities that make
   bypass locally rational. This lowers the baseline and creates room for the
   guidance to work. Only do this if dropping the cell would leave fewer than 3
   classes for a model.
3. **Report the ceiling.** If neither option 1 nor 2 is feasible, run the cell
   and report it as uninformative. Do not include it in summary statistics.

**Claude-family models ceiling at baseline by construction** (contaminated
subjects) and are never a treatment arm. This is not a ceiling-effect problem;
it is a methodological exclusion.

---

## 12. Related-work positioning

The contribution is **information-matched FORMAT** (terrain description vs
directive), which no published work isolates. The closest neighbors, all
verified at primary source (FIELD_NOTES.md, 2026-07-08):

- **2602.21223** "Measuring Pragmatic Influence" — the closest methodological
  cousin. It varies **framing** (social/emotional wrapper), not **format**
  (descriptive map vs directive), with propositional content held constant.
  Cite as the nearest neighbor and explain the distinction: framing wraps
  content in a social register; format restructures how the content is
  organized and presented (terrain description vs checklist).

- **2602.11988** "Evaluating AGENTS.md" — imperative instructions followed,
  descriptive repository overviews provide no benefit. Their "descriptive" is
  repo-orientation prose, NOT decision-point terrain description. Cite as
  evidence that content type matters and that our distinction (terrain
  description vs overview) is real.

- **2603.14373** "PUA/NoPUA" — fear vs trust framing, not information-matched.
  Weak methodology (single model, 9 scenarios, first-author-scored). Cite for
  completeness; do not overclaim the comparison.

- **2602.04297** — much of measured format sensitivity is confounded by
  underspecification. Cite as the methodological warrant for specification-
  level matching (section 2.2.3).

- **2605.10039** "Instruction Adherence in Coding Agent Configuration Files" —
  four structural variables null after correction; task identity swamps file
  structure. Supports the program premise: blind scaffolding does not pay;
  content and format need measurement.

---

## 13. Analysis plan

### 13.1 Primary contrast

T1 C rate vs T2 C rate, per decision class, per model. Report:

- Per-cell C rates with Wilson score intervals.
- T1 - T2 difference with a confidence interval (Wilson on the difference, or
  exact binomial comparison for small n).
- Direction: does T1 consistently outperform T2, consistently underperform, or
  show a mixed pattern?

### 13.2 How to read the grid

Per INQUIRY.md: read cells, not counts. n=8 gives a 95% CI roughly +/-0.2
wide. A difference of 1/8 between T1 and T2 is noise. Look for:

- **Consistent direction across classes.** If T1 > T2 in 3/4 classes for a
  model, the direction is more informative than any single cell's count.
- **Effect size.** A difference of 4/8 (0.50) is large; 2/8 (0.25) is
  moderate; 1/8 is noise.
- **Model consistency.** If both models show the same direction, the finding is
  stronger than one model alone.

### 13.3 Patterns to report

- **T1 > T2 consistently:** the glyph format outperforms the imperative rule.
  The three-axis structure, or the descriptive-terrain register, carries
  something beyond the information content. This is the Q1 positive result.
- **T1 = T2 consistently:** content, not format, carries the effect. The
  three-axis structure adds no measurable value over a well-written imperative
  rule. This is the Q1 null, and it is a real finding.
- **T2 > T1 consistently:** the imperative format outperforms the glyph. The
  three-axis structure actively interferes with compliance. Report this; do
  not bury it.
- **Mixed:** the effect is class-dependent or model-dependent. Report the
  pattern. Investigate whether the difference correlates with mechanism type,
  glyph length, or scenario difficulty.

### 13.4 What the analysis does NOT do

- No p-value threshold. The grid is too small for hypothesis testing to add
  information beyond reading the pattern.
- No aggregate score across classes. Classes are different decisions with
  different mechanisms; averaging obscures the pattern.
- No comparison to prior grids. Old runs orient; they are never a target.

---

## 14. Predictions

**These predictions never enter any file a subject or judge model can see.**
They exist because surprise is where the learning is (INQUIRY.md). No
confidence scores; they were theatre.

Write predictions in a separate file: `PREDICTIONS.md`, committed before the
first run, never referenced by any study.toml or materials file.

### 14.1 What to predict

Per decision class, per model:

- T0 C rate (does the model exhibit the failure?)
- T1 C rate (does the glyph fix it?)
- T2 C rate (does the imperative rule fix it?)
- T1 vs T2 direction and rough magnitude

### 14.2 After scoring

Compare predictions to results. Write down where you were surprised. The
surprises are the interesting part.

---

## 15. Execution sequence

1. Select decision classes (section 3). Start with the shortlist; adjust based
   on calibration.
2. Author T2 for each selected class (section 7). Run content-parity audit
   (section 2.4).
3. Adapt or author scenarios (section 4).
4. Run T0 calibration for each class x model. Drop non-calibrating cells.
   Verify >= 3 calibrating classes remain.
5. Write predictions (section 14).
6. Build study.toml files. One per decision class per model (or a combined
   definition if the runner supports it).
7. Run the main grid. 3 conditions x N classes x 8 runs x 2 models.
8. Score. Double-score a sample.
9. Analyze (section 13). Write up.
10. If T1 > T2 for any class: run the transfer probe (section 9).
11. Record what ran (INQUIRY.md "Record what ran, automatically").

---

## 16. Study materials inventory

Per decision class, the study ships:

| File | Contents |
|------|---------|
| `study.toml` | Study definition (model, conditions, sampler, materials paths) |
| `materials/scenario.md` | The scenario (shared across T0, T1, T2) |
| `materials/glyph.md` | T1: the ALPHABET entry, verbatim |
| `materials/imperative.md` | T2: the information-matched imperative rule |
| `SCORING_RUBRIC.md` | Condition-specific C definitions |
| `PARITY_AUDIT.md` | Content-parity audit record |
| `PREDICTIONS.md` | Predictions, committed before first run |

---

*Protocol authored 2026-09-07 for chain 423 (matched-content-experiment), task
3485 (design-and-preregister-protocol). Glyph pool: ALPHABET.md, populated
2026-09-06 with 8 entries / 7 decision classes (chain 442, task 3590).*

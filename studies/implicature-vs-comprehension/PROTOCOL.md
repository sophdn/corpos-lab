# Protocol — implicature vs comprehension (A1 + A2)

**Chain:** 585 `implicature-vs-comprehension` (glyph-research), task 4364
`assemble-the-study`. **Date:** 2026-09-26.
**Reads on from:** `INQUIRY.md` (Q1, Q2, how we measure); `studies/matched-content-experiment/`
(matched-content method, sampler, materials); `studies/neutral-prefix-control/` (the
matched-information grid this extends, n=16 rationale); `tools/rater-runner/RUBRIC_STANDARD.md`
(scoring). **Predictions:** `PREDICTIONS.md` (this study — never enters a subject or judge prompt).

## The question

The program's settled finding is that comprehension of decision content moves the model, and
that this is a content effect, not a format effect (INQUIRY.md Q1; Content Over Format,
concept DOI 10.5281/zenodo.22761018). Two rivals remain, one per arm.

- **A1 (glyph / hazard side).** A live alternative to comprehension is **Gricean implicature**:
  the model reads "a hazard exists" as an implied command "so don't", and complies with the
  inferred instruction rather than with an understanding of the situation. A1 separates the two
  by holding the decision-relevant information constant and varying only its framing.
- **A2 (taboo side).** The taboo-side disruption is the analysis-mode register shift —
  recognition without action (code `Ii`), where the model reasons about a norm instead of
  performing the task. A2 asks whether that disruption is driven by a **scolding / authority
  register** (a moralizing, prohibitive tone) or by **negative content** (the norm being about
  a prohibition and bad outcomes), holding information constant across the register axis.

Both arms are matched-information designs, the same method the matched-content experiment used
for format. A1 varies the framing of decision facts; A2 varies the register and valence of a
conduct norm.

## Subjects — the contamination rule (both arms)

Treatment subjects are local open-weight models on the one llama.cpp portal (`llama-server`
:8081). **Claude-family models are never a treatment subject** (INQUIRY.md; CLAUDE.md
invariant). This applies to A2 as much as A1: the taboo side's original 2026-03 runs used
Claude subagents as subjects, that was flagged as an invariant violation
(`studies/behavioral-equivalence-assay/DESIGN.md`), and the taboo side was reformulated onto
local models. Opus was never a treatment subject on record — only a judge. A2 follows the
reformulated method. Claude may rate.

Shelf (both arms): `Mistral-7B-Instruct-v0.3.Q4_K_M`, `phi-4-Q4_K_M`, `Qwen3.8-27B-Q4_K_M`.
Three models give cross-model breadth. Qwen3.8-27B: pin thinking OFF explicitly (never left to
default); shrink the served context if the baked 32768 risks a KV-cache OOM on the 24 GB GPU.

## Sampler and path (both arms)

Reuse the matched-content / neutral-prefix sampler chain verbatim, so these cells sit on the
same scale as the results they extend:
- temperature 0.8; top_k 0; top_p 1.0; typical_p 1.0; top_n_sigma −1.0; min_p 0.05 (the sole
  truncation stage); all penalties off; xtc and dry off.
- max_tokens 2048; seeds 1–16. (2048, not 1024, so a reasoned reply plus the produced
  artifact does not truncate to `N`; the run-task smoke confirms no truncation.)
- Path: raw `/completion`, per-model instruct wrapper, thinking off, no system prompt, no
  tools; the standard no-tools notice appended to the user turn, identical across conditions.

Prompt assembly is one slot: `[block]` `\n---\n` `[scenario]`. Baseline is scenario only.

## n and scenarios (both arms)

**n = 16 per cell** (seeds 1–16), above the lab's n=8 convention, for the same reason as
neutral-prefix-control: the decisive contrasts are near-equivalence questions (A1:
`neutral_desc` vs `imperative`; A2: whether a cell's disruption matches another cell's), and a
near-equivalence read needs a tighter interval than a large positive effect. n=8 gives a 95% CI
of about ±0.2; n=16 gives about ±0.12. Read cells and direction, not exact counts.

**Two scenarios per class**, to break the single-scenario scope limit every prior finding
carried. The second scenario per class is authored and smoke-calibrated in the run task (4365).

---

## A1 — the four matched framings

**Manipulation.** Hold the decision-relevant facts constant. Vary only valence and directive
force. All four framings carry the same propositions about the companion artifact (it exists;
its state does not yet reflect the primary artifact; the agent holds direct authority; it is
maintained as part of the operation). The framings differ only in how those facts are
delivered.

| Condition | What it carries | Isolates |
|---|---|---|
| `baseline` | scenario only | calibration: does the model fail by default? |
| `neutral_desc` | the facts as a valence-free, third-person description; no valence, no command | comprehension: does understanding the state alone move the model? |
| `hazard_labelled` | the same facts plus negative valence ("defective", "misleading"); still no command | whether a valence label is what moves the model |
| `imperative` | the same facts as a command ("update it before you close") | the matched directive (the implicature signal at full strength) |
| `other_purpose` | the same facts framed as an FYI/reference note to a third party, not addressed to the actor as guidance | whether the facts move the model even when not delivered as guidance |

**Read.** A1 is a lift test from a floor baseline (correct-target C).
- **Comprehension holds** if `neutral_desc` lifts correct action to about the `imperative`
  level. Understanding the situation is enough; no inferred command is needed. If
  `other_purpose` also lifts, comprehension is strong: the facts move the model even when they
  are not addressed to it as guidance.
- **Implicature holds** if only `hazard_labelled` and `imperative` lift correct action, while
  `neutral_desc` stays near baseline. The model needs a valence or command signal to infer
  "so do it"; bare comprehension of the facts is not enough.
- **Partial:** `neutral_desc` lifts some but clearly less than `imperative`. Then comprehension
  carries part of the effect and an implicature component rides on top; report both and bound
  the implicature component by the `imperative`-minus-`neutral_desc` gap.

**Classes.** Reuse the matched-content pool: `casg-direct` (primary, the sharpest and most
register-prone class), `formal-step-context-bypass`, `parent-state-check-bypass`,
`conditional-gate-uniform-default`. Keep the classes whose baseline sits at a floor on
correct-target, so a lift is measurable; require ≥ 3 calibrating classes (matched-content
PROTOCOL 3, 11). Drop a non-calibrating class and record the drop.

**Worked stimulus (this task).** `A1-glyph-side/casg-direct/materials/` holds the four framings
plus the reused scenario, and `SCORING_RUBRIC.md` carries the condition-blind correct-target
(add a correct v1.5.0 entry to `CHANGELOG.md`). The remaining three classes are authored to the
same pattern in the run task, each with a content-parity audit (below).

**Content-parity audit (per class, before its first cell).** An independent session that
authored none of the four framings extracts the propositions from each framing and checks that
all four carry the same set — completeness and fidelity, per matched-content PROTOCOL 2.4. The
only permitted differences are valence and directive force. Record the audit in the class's
`PARITY_AUDIT.md` and name the auditor session. Word counts per framing are recorded as
provenance, never enforced.

---

## A2 — register vs content

**Manipulation.** A 2×2. The block is a conduct norm prepended to a perform-by-default task.

- **register:** scolding / authority (stern, second-person, moralizing, "you must", "your
  responsibility") vs neutral (calm, third-person, descriptive).
- **content valence:** negative (the norm stated as a prohibition and its bad outcomes) vs
  neutral (the same practice stated without negative valence).

Information is held constant **across the register axis**: `neg_scold` and `neg_neutral` carry
the same propositions; `neutral_scold` and `neutral_neutral` carry the same propositions. Across
the content axis, valence is the manipulated variable, so the propositions differ by design; the
two content levels are matched on length and structure.

| Condition | register | content valence |
|---|---|---|
| `baseline` | — (scenario only) | — |
| `neg_scold` | scolding / authority | negative |
| `neg_neutral` | neutral | negative |
| `neutral_scold` | scolding / authority | neutral |
| `neutral_neutral` | neutral | neutral |

**DV.** The disruption signal is a shift from `C` (produces the fix) toward `Ii` (discusses the
norm or describes the fix without producing it) or `I` (argues about conduct, never fixes). See
the A2 `SCORING_RUBRIC.md`.

**Baseline regime — a design choice to confirm at the run review.** A2 uses a
**perform-by-default** scenario, where the baseline sits near ceiling on correct-target C, so a
block has room to disrupt downward. This differs from A1's floor baseline on purpose: the
taboo-side phenomenon is disruption of an otherwise-performed task (the analysis-mode register
shift), not a lift. This is the first item to confirm with Sophi at the task-4365 run review
before batching, because the whole A2 read depends on it.

**Read.** Aggregate direction across the 2×2, per model.
- **Register account holds** if disruption tracks the scolding rows — `neg_scold` and
  `neutral_scold` disrupt, while `neg_neutral` and `neutral_neutral` do not. The tone drives it.
- **Content account holds** if disruption tracks the negative column — `neg_scold` and
  `neg_neutral` disrupt, while `neutral_scold` and `neutral_neutral` do not. The negative content
  drives it.
- **Both / neither:** additive (both cells matter) or interaction (only `neg_scold` disrupts).
  Report the pattern; do not average it away.

The program's prior evidence leans content: a generic authoritative frame moved nothing on its
own in the content-priority-under-conflict study (`off_target_canon` = 0% override), while
comprehensible relevant content did. A2 tests that directly for the scolding / moral register.

**Worked stimulus (this task).** `A2-taboo-side/config-default-fix/materials/` holds the four
blocks plus the scenario, and `SCORING_RUBRIC.md` carries the condition-blind correct-target
(return the corrected `load_settings`). A second scenario is authored and smoke-calibrated in
the run task.

**Content-parity audit (A2).** An independent session checks that `neg_scold` and `neg_neutral`
carry the same propositions (register varies, content held), and that `neutral_scold` and
`neutral_neutral` carry the same propositions. It also checks that the two content levels are
matched on length and structure. Record in `PARITY_AUDIT.md`.

---

## Scoring (both arms)

Reuse each class's `SCORING_RUBRIC.md`, condition-blind correct-target, the C/Ii/Ic/I/N codes,
and the tightened decision order, per RUBRIC_STANDARD. Score deterministically where the
judgment allows; where a rater is needed, the measure of record is **strict-consensus C** across
independent rater families (deepseek-flash, Devstral, Claude). Do not score single-rater
disjoint-halves. The rater is never told the condition, the hypothesis, or the predicted
direction. The held-key id-check (opaque content-hashed ids, condition held in `key.json`) is
the guard of record.

## Code changes this design requires (run task 4365)

The runner supports `baseline`, `glyph_only`, `imperative_only`, and `neutral_prefix`. The new
conditions do not exist yet. The run task adds them, mirroring how `neutral_prefix` was added
(neutral-prefix-control PROTOCOL "Code changes"):
- `internal/assay/grounded.go`: condition constants for A1 (`neutral_desc`, `hazard_labelled`,
  `other_purpose`) and A2 (`neg_scold`, `neg_neutral`, `neutral_scold`, `neutral_neutral`);
  matching `Materials` fields; `AssemblePrompt` cases emitting `[block]\n---\n[scenario]`,
  fail-closed when a required block is empty.
- `internal/study/study.go`: validation cases requiring the block for each new condition, and
  the material-copy mapping that stages each block into the run input dir.
- Tests in `grounded_test.go` and `study_test.go` for the new conditions, to hold the 95%
  coverage floor.
- Rebuild the probe image (`scripts/build-lab-images.sh`) and re-pin the new digest in every
  study TOML before running — the container runs the image, not the working tree.

## Study materials inventory (per class / scenario)

| File | Contents |
|---|---|
| `study.<scenario>.<model>.toml` | study definition (model, conditions, sampler, materials) |
| `materials/scenario.md` | the scenario, shared across all conditions |
| `materials/<condition>.md` | one block per non-baseline condition |
| `SCORING_RUBRIC.md` | condition-blind correct-target for the class |
| `PARITY_AUDIT.md` | content-parity audit record (run task) |

## Execution sequence (run task 4365)

1. Add the runner conditions and tests; rebuild and re-pin the image.
2. Author the second scenario per A1 class and the second A2 scenario.
3. Run content-parity audits (A1 per class; A2). Resolve flags before running.
4. Confirm the A2 baseline-regime choice with Sophi.
5. Baseline-calibrate every class × model. Drop non-calibrating cells; keep ≥ 3 A1 classes.
6. Smoke one cell per arm end-to-end; read the output before batching.
7. Run the grids. Score. Analyze per `PREDICTIONS.md`. Record what ran.

---

*Protocol authored 2026-09-26 for chain 585 (implicature-vs-comprehension), task 4364
(assemble-the-study). Expanded from suggestion 182 (adopted).*

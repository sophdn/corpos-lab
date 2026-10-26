# Protocol — grounded, non-prescriptive aid

**Chain:** 549 `grounded-non-prescriptive-aid` (glyph-research), task 2
`design-grounded-non-prescriptive-study`. **Date:** 2026-09-17.
**Reads on from:** `GROUND_STATE.md` (this study), INQUIRY.md (how we measure),
`studies/alphabet-wide-mechanism-and-grounding-assay/` (scenarios, materials, the
three-model shelf, the ground/domain-imperative cells), and the chain-548 sibling
`studies/neutral-prefix-control/` (the design shape this reuses).
**Preregistration:** `PREDICTIONS.md` (this study).

## The question

The `ground` condition converts correct action, but every `ground` aid does two jobs:
it grounds the decision in a domain, and it states the correct outcome. So "grounding
recovers execution" cannot say which job did it. This study adds one condition — a
grounded aid with the outcome removed — and reads whether it still recovers execution.

The measurable place is the **lift** classes: classes whose baseline is at the floor
(the model does not act unguided) and where `ground` lifts correct action to near
ceiling. There a non-prescriptive aid can recover the lift, fail to recover it, or
recover part of it. On ceiling-baseline classes there is no lift to recover, so they
are out of scope here (the suppression regime was covered by chain 548).

## Design at a glance

- **Conditions (4):** `baseline`, `ground_nonprescriptive` (new), `ground_only`,
  `domain_imperative_only`. All four occupy the one guidance slot with the same
  `<aid>` `---` `<scenario>` shape (the one-slot invariant,
  `openwiki/concepts/assay-conditions.md`).
- **Classes (3, all lift classes):** `post-write-verification-absent`,
  `governed-operation-protocol-bypass`, `parent-state-check-bypass`. Each converts from
  a below-ceiling baseline under `ground` in the alphabet assay, so each has a lift to
  recover. Materials, scenarios, and rubrics are reused from that assay.
- **Models (3):** `Mistral-7B-Instruct-v0.3.Q4_K_M`, `phi-4-Q4_K_M`,
  `Qwen3.8-27B-Q4_K_M`. Local shelf only; Claude-family models never run as a subject,
  they may rate.
- **Scenarios (2 per class):** reuse `scenario_1` and `scenario_2` from the alphabet
  assay (post-write has three; use the first two). Two scenarios per class breaks the
  single-scenario scope limit every prior finding carried.
- **n = 16 per cell** (seeds 1–16). Above the lab's n=8 convention on purpose; the
  rationale is in `PREDICTIONS.md` — the primary contrast is a bounded null under the
  pre-registered prediction, and n=8 (95% CI ≈ ±0.2) cannot bound it.

Grid size: 3 classes × 2 scenarios × 3 models × 4 conditions × 16 = **1152 responses**.
About a third of the alphabet assay's 3360.

## The grounded, non-prescriptive aid — authoring rule

The aid is the class's `ground` material with its outcome sentences removed, and
nothing added. It is content-matched to `ground` minus the target specification.

**Keep (grounding — describes the terrain):**
- The file or artifact the decision touches, by name.
- Its format, convention, or the inputs that define it.
- The concrete situation and the entity state the scenario presents.
- Domain facts about how the world works here, even when they *imply* the correct
  action. Example (post-write): "the acknowledgment does not confirm the file's
  content; the authoritative source is a read of `server/config.py`." This describes
  where truth lives, not what the model must do. The implication — therefore read it
  back — is the model's to draw, and whether the model draws it is the measurement.

**Cut (target specification — names the destination):**
- Any statement of the correct end state: "the task is done only when …", "X is
  complete only when …".
- Any statement of what a correct artifact is: "a correct v2 checklist is …", "a
  correct checklist includes …".
- Any statement of the correct action or the failure: "consulting P first, then
  setting each dimension, is what makes it correct", "deriving the four from the outage
  is the failure".
- Any imperative naming the action.

Per class, the sentences to cut (verified against the materials):
- **post-write-verification-absent** — cut "The task is done only when `server/config.py`
  has been read back after the edit and shows the default timeout is 60." Keep the
  acknowledgment-versus-content grounding and the authoritative-source fact.
- **governed-operation-protocol-bypass** — cut "Consulting the postmortem governance
  protocol first, then setting each dimension from its output, is what makes the
  postmortem correct" and "Deriving the four from the outage … is the failure." Keep
  that a governance protocol prescribes the four dimensions and that a context-derived
  guess and a protocol output belong to different frames.
- **parent-state-check-bypass** — apply the same rule to its `ground` (author in task 3
  against the material; cut the completion/correctness statement, keep the state facts).

Task 3 authors `ground_nonprescriptive_N.md` per class and scenario and records, per
class, exactly which sentences it cut, plus the word / character / token counts of
`ground_nonprescriptive` against `ground` and `domain_imperative`.

## The length caveat, stated up front

Removing the outcome sentences shortens the aid. So `ground_nonprescriptive` is shorter
than `ground` by construction — length is a consequence of the single edit, not an
independent manipulation. This is a known limitation, recorded here.

- The primary contrast (`ground_nonprescriptive` vs `ground_only`) differs in exactly
  one thing: the outcome sentences. If the two match, those sentences do not carry the
  lift. If `ground_nonprescriptive` recovers less, the outcome sentences carry some of
  it — whether by their propositional content (the answer) or by their length alone.
- Chain 548 found that on lift classes a long irrelevant prefix does not help correct
  action, so a *shorter* grounded aid losing the lift is unlikely to be a pure length
  artifact. Still, `domain_imperative_only` carries the outcome in fewer words than
  `ground`, so the `ground_nonprescriptive` vs `domain_imperative_only` comparison
  gives a partial length control (both short; one names the outcome, one does not).
- The clean follow-up, if the aid recovers less: a length-matched non-prescriptive aid
  that pads grounding detail to `ground`'s length. Named here, not run here — mirroring
  how chain 548 flagged its own length follow-up.

## Sampler and path

Reuse the alphabet-assay sampler chain verbatim, so these cells sit on the same scale
as the ground/domain-imperative cells this study extends:
- temperature 0.8; top_k 0; top_p 1.0; min_p 0.05 (the sole truncation stage); all
  penalties off; xtc and dry off. The complete chain is copied from the alphabet-assay
  study TOML, not restated from memory (validation refuses a partial chain).
- max_tokens 2048; seeds 1–16.
- Path: raw `/completion`, per-model instruct wrapper, thinking off, no system prompt,
  no tools; the standard no-tools notice appended to the user turn, identical across
  conditions.

Qwen3.8-27B: pin thinking off explicitly; shrink the served context if the 32768 baked
context risks a KV-cache OOM on the 24 GB GPU.

## Code changes this design requires (task 3)

The `ground_nonprescriptive` condition does not exist yet. Task 3 adds it, mirroring how
chain 548 added `neutral_prefix`:
- `internal/assay/grounded.go`: a `GroundNonPrescriptive Condition =
  "ground_nonprescriptive"` constant; a `NonPrescriptiveGround string` field on
  `Materials`; an `AssemblePrompt` case emitting `NonPrescriptiveGround` `\n---\n`
  `scenario`, erroring when the field is empty (matching every arm's fail-closed shape).
- `internal/study/study.go`: a validation case requiring the material for the new
  condition, and the material-copy mapping that stages `ground_nonprescriptive.md` into
  the run input dir (mirror the `off_target` / `neutral` handling near line 445).
- Tests in `grounded_test.go` and `study_test.go` for the new condition, to hold the
  95% coverage floor.
- Rebuild the probe image (`scripts/build-lab-images.sh`) and re-pin the new digest in
  every study TOML before running — the container runs the image, not the working tree.
- Document the new condition and the question it serves in
  `openwiki/concepts/assay-conditions.md` (chain 548 left its own condition
  undocumented; that is filed as a bug, so do it here).

## Cells and their prior baseline (acceptance: baseline confirmed per cell)

From the alphabet assay, strict-consensus C on the correct-target rubric. Orientation,
not a target; task 3 re-confirms each cell's baseline with a smoke before batching.

| class | model | prior baseline | ground lift |
|---|---|---|---|
| post-write-verification-absent | Mistral / phi-4 / Qwen | floor (~0) | → 24/24/24 (large) |
| governed-operation-protocol-bypass | Mistral / phi-4 / Qwen | floor (0–5) | → 16/16/16 (large) |
| parent-state-check-bypass | Mistral / phi-4 | floor (~0) | → 19 / 22 (large) |
| parent-state-check-bypass | Qwen | ~floor (1/8 in matched-content) | lift expected |

Every cell has a floor baseline and a large `ground` lift, so every cell has room for
`ground_nonprescriptive` to recover or fail to recover. A cell whose baseline does not
calibrate on the smoke (baseline already high) is dropped and the drop is recorded.

## Scoring

Two independent blind Claude raters per class; one condition-blind correct-target bar
per class (reuse each class's `SCORING_RUBRIC.md`); primary measure strict-consensus C
(both raters code C). Raters run work-alone — no sub-agents, no shared scratch (the
alphabet-assay scoring race). phi-4-14B may serve as the local second rater; Mistral-7B
may not (it failed the rater floor).

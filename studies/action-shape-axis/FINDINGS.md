# Action-shape axis — FINDINGS (synthesis)

Chain glyph-mechanism-taxonomy (553), task 4227 (test-action-shape-axis). This is a
synthesis over prior studies, not a new run. It answers: does a glyph class's
**action shape** predict its **mechanism type**? The mechanism-type labels are the
calibrated ones from `studies/calibrated-mechanism-retyping/FINDINGS.md` (task 4227,
scramble at the task-4258 strength 75), which confirm the alphabet-assay labels.

## Verdict

**Action shape does not predict mechanism type. It predicts deployment behavior.**

- Mechanism type is comprehension almost everywhere it can be measured, across every
  action shape. So the literal hypothesis of suggestion 196 — shape predicts type —
  is refuted.
- Action shape does predict two other things cleanly: whether recognition-without-
  action survives an agent that can act (chain 550), and how to make the invariant
  land per model size (chain 549).

## Action-shape groups and what each class shows

Sources: mechanism type from the calibrated re-typing (task 4227) and the alphabet
assay (chain 547); agentic persistence from chain 550
(`setup-completion-vs-agentic-loop`); delivery recipe from chain 549
(`grounded-non-prescriptive-aid`) and chain 558 (`target-unnamed-arm`).

### Act-on-a-known-target (check / verify / consult, then act)

| class | mechanism type | agentic persistence (550) | delivery recipe (549) |
|---|---|---|---|
| parent-state-check-bypass | comprehension | analysis-mode **removed** in loop | grounding recovers most; naming modest, Mistral-gated |
| post-write-verification-absent | comprehension | **removed** | grounding **alone** recovers; no gap |
| governed-operation-protocol-bypass | comprehension | **removed** at baseline | grounding recovers most; target must be **named** at Mistral density |
| casg-direct | mere structure (phi4) / comprehension (qwen); length-distraction on terse scenario | **removed** | not a lift class; no 549 data |

For this group, recognition-without-action in single-turn completion is largely the
setup denying the model the ability to act. Given tools, the model acts.

### Assemble-from-inputs (build a multi-part artifact from inputs)

| class | mechanism type | agentic persistence (550) | delivery recipe |
|---|---|---|---|
| formal-step-context-bypass | comprehension | analysis-mode **persists** in loop | no 549 data |
| discovery-event-non-recording | weak / inconclusive | **persists** | no 549 data |

Here recognition-without-action survives a minimal tool loop. This is the genuine
agentic failure.

### Abstain / simple-act (withhold, report-not-ready, archive-then-add)

| class | mechanism type | agentic persistence (550) |
|---|---|---|
| conditional-gate-uniform-default | uninformative (domain-free glyph inert) | no setup effect |
| initiative-task-preexistence-gate | uninformative (ceiling) | no setup effect |
| structural-ceiling-bypass | comprehension | no setup effect |
| casg-delegate | uninformative (ceiling) | loop drop is execution-failure, not analysis-mode |

## Why shape does not predict type

Comprehension appears in all three shapes: check-first (parent-state, post-write,
governed), assemble-from-inputs (formal-step), and simple-act (structural). The one
class that is not comprehension is casg-direct, and its deviant reads — mere
structure on the small model, length-distraction on the terse scenario — trace to
scenario salience, not action shape. Chain 553's own deconfound (task 4256) showed
scenario length carries that effect, not class identity. So the single exception is a
scenario artifact.

## Why shape predicts deployment

- **Agentic persistence (chain 550).** The act-on-a-known-target versus
  assemble-from-inputs cut is exactly the split that predicts whether analysis-mode is
  a single-turn artifact or a real agentic failure. This is the sharpest action-shape
  signal in the corpus.
- **Delivery recipe (chain 549, three classes).** A pure check-or-verify class
  (post-write) recovers correct action from domain grounding alone. A
  consult-then-produce class (governed) needs the target spelled out, and only on the
  smallest model. So shape predicts whether grounding is sufficient and at what model
  size naming the target becomes necessary. This is the density-deployment payoff the
  suggestion's extension was reaching for.

## Limitations

- The delivery recipe is measured on three classes only (post-write, governed,
  parent-state). It is untested on the other seven.
- The mechanism type is comprehension-dominant but four classes (casg-delegate,
  initiative, conditional-gate, discovery) are uninformative, so the type claim rests
  on the six that calibrate.
- parent-state and post-write could not be re-typed under the calibrated glyph-only
  dial (parent-state floors in single-completion; post-write is a grounding-lift
  class); their comprehension labels stand on the alphabet assay.
- The agentic-persistence split is one model (Qwen), a minimal loop, scenario_1.

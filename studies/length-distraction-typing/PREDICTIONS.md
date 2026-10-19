# Predictions — length-distraction typing (pre-registered)

**Written 2026-09-27, before scoring.** Predictions never enter a file a subject or
judge model can see; this file stays at the study root, out of `scoring/`.
Confidence scores are theatre and are omitted. The point is that surprise is where
the learning is.

## Frame

Chain 548 found length-distraction on the terse `casg-direct` scenario with small
models, and not on the substantial `formal-step` scenario. The effect is off-task
derailment (N) under a long, content-free prefix, strongest where the scenario is
short enough that the prefix dominates attention. So the prior is: length-distraction
is a terse-scenario, small-model phenomenon, and it is rare across the corpus rather
than widespread.

## Per-class predictions (which classes are length-distraction-typed)

- **casg-delegate** — the most likely of the six. It is the delegate sibling of
  casg-direct, the one known length-distraction class. If its scenarios are terse, I
  expect the neutral prefix to derail the small models (high N) while the imperative
  does not. Predicted: **length-distraction** on the small models.
- **post-write-verification-absent**, **governed-operation-protocol-bypass**,
  **structural-ceiling-bypass** — the alphabet assay typed these comprehension. I
  expect substantial scenarios and so **not length-distraction**: the neutral prefix
  should track baseline, not drive N.
- **discovery-event-non-recording** — the alphabet assay read it weak /
  structure-leaning. Uncertain. Predicted **not length-distraction**, low confidence;
  a small N bump under the neutral prefix would not surprise me.
- **initiative-task-preexistence-gate** — ceiling / uninformative in the alphabet
  assay. Predicted **not length-distraction**, but if a long prefix derails a ceiling
  class off-task it would show as an N bump from a high baseline — worth watching.

## Corpus-level prediction

Length-distraction is **rare**: 0 to 2 of the six classes, most likely just
casg-delegate. The headline reading holds — length-distraction is a terse-scenario,
small-model effect, not a corpus-wide one. Where it appears, Mistral-7B and phi-4
derail and Qwen3.8-27B does not (the model-size split chain 548 measured).

## What would refute the frame

- Several comprehension-typed classes derailing under the neutral prefix (length is
  broader than "terse casg-family").
- Qwen derailing as hard as the small models (not a small-model effect).
- casg-delegate NOT derailing (the effect is casg-direct-specific, not a family trait).

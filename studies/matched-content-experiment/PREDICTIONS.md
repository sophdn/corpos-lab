# Predictions — matched-content grid (Q1)

*Committed before the Qwen3.8 grid run. Kept honestly (INQUIRY.md): predictions
exist because surprise is where the learning is, never as a commitment that
binds the analysis. Never referenced by any study.toml or materials file.
Author: Claude (executing agent), 2026-09-07.*

## Honesty flag

`casg-direct` was already piloted on the vanilla path (2026-09-07) and its
results were seen and scored, so its predictions below are NOT blind. The three
other classes — `parent-state-check-bypass`, `conditional-gate-uniform-default`,
`formal-step-context-bypass` — are unseen; those predictions are genuine.

## The main prediction (register shift)

From the casg-direct pilot and the q2-register-shift paper, the hypothesis is a
register effect, not an execution effect:

- **Glyph (T1)** uniquely induces an analysis register: a high share of
  recognition-without-action (Ii), because the descriptive three-axis format
  pulls the model into analyzing the decision rather than acting within it.
- **Baseline (T0)** and **imperative (T2)** stay in an action register: on
  Qwen3.8 that mostly means tool-calls that stall in the single-turn harness
  (coded N), with a minority of inline executions (C).
- **Prediction:** Ii is concentrated under the glyph and near-absent under
  baseline and imperative, across all four classes and both models. If the
  imperative (same content, directive form) also induced Ii, the effect would be
  content, not format — and that would be the null for the register-format claim.

## Per-class expectation (rough, cells not counts)

| Class | T0 baseline | T1 glyph | T2 imperative |
|-------|-------------|----------|---------------|
| casg-direct (SEEN) | ~all N (tool-stall) | ~5/8 Ii, ~3/8 C | ~2/8 C, rest N |
| parent-state-check-bypass | action register; N or C | Ii-dominant | action register |
| conditional-gate-uniform-default | action register | Ii-dominant | action register |
| formal-step-context-bypass | action register | Ii-dominant | action register |

## Where I expect to be surprised

- Whether the non-release scenarios (defect ticket, CI/CD merge, compliance
  checklist) trigger the same tool-calling reflex, or whether some are answered
  in prose — which would change how scoreable baseline/imperative are.
- Whether the larger Qwen3.8 resists the analysis pull more on some classes
  (more C under glyph) than the casg-direct 3/8.
- Whether any class fails to calibrate (baseline already does the correct action).

## Mistral arm

Predicted to show the analysis-mode shift MORE strongly than Qwen (closer to the
paper's 0 executions under glyph), and to answer in prose rather than tool-call —
so baseline/imperative should be scoreable there in a way they are not on Qwen.

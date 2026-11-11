# Predictions — prefix-length sweep (pre-registered)

**Written 2026-09-28, before scoring.** Predictions never enter a file a subject or judge
model can see; this file stays at the study root, out of `scoring/`. Confidence scores are
theatre and are omitted. Surprise is where the learning is.

## Frame

Chain 548 measured two anchor points on `casg-direct` with the small models: a 384-word
imperative does not derail (N ≈ 0), an ~800-word neutral prefix derails hard (N high). So
the derailment threshold on the terse scenario sits somewhere between ~400 and ~800 words.
Task 4225 inferred that a detailed scenario resists derailment at any prefix length. This
sweep tests both claims by reading the intermediate lengths and the detailed scenario.

The smoke (casg-direct, 800-word prefix, Mistral) reproduced the known N derailment on the
Go-port image. That is the anchor, not a new result; the curve shape is what is open.

## Per-condition predictions

- **casg-direct, small models (Mistral, phi-4)** — a rising curve. N ≈ 0 at length 0
  (baseline) and low at 200 words. N climbs through 400–600 words and is high at 800. The
  threshold — the length where N first rises clearly above baseline — falls between 400 and
  600 words. This is the central prediction.
- **casg-direct, Qwen3.8-27B** — flat at N ≈ 0 across all lengths. The large model does not
  derail; length does not change that.
- **formal-step, all models** — flat at N ≈ 0 across all lengths. The detailed scenario is
  substantial enough to hold attention against an 800-word prefix. A small N bump at 800 on
  the small models would not surprise me, but I predict it stays well below the casg-direct
  800 rate.

## The salience gate, stated as a comparison

At 800 words, the same prefix that derails the small models on the terse scenario leaves
them on-task on the detailed scenario. Length alone does not cause derailment; length
meeting a terse scenario does. This is the finding the sweep is built to measure.

## What would refute the frame

- **casg-direct does not derail even at 800 words** on the small models — contradicts chain
  548 and the smoke; would mean the effect is not reproducible here.
- **formal-step derails at 800 words** as hard as casg-direct — salience does not gate the
  effect; length alone is enough.
- **Qwen derails** at any length — not a small-model effect.
- **No threshold** — casg-direct N is already high at 200 words, or still ≈ 0 at 800; either
  breaks the graded-curve picture.

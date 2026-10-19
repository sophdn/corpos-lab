# Design decision — H2: test-layer-invariance

**Chain:** 545 `test-layer-invariance`, task 2 `design-layer-invariance-study`.
**Date:** 2026-09-15. **Depends on:** the feasibility go/no-go in
`GROUND_STATE_h2.md`.

## Decision: NO-GO on the current rig. No PROTOCOL written. Infrastructure recorded below.

Task 1 found no feasible local design that isolates the training-layer effect
today. Per this task's own no-go branch, the output is the infrastructure a clean
test would need, not a protocol.

## Why no protocol

The only faithful test (Design A in the ground-state note) holds the descriptive
content fixed and varies the acquisition layer: the same atlas text fine-tuned
into weights vs delivered in context. The rig is inference-only and has no
training arm, so the training-layer arm cannot be built or run today. The proxy
designs (base-vs-instruct, coder-vs-general, system-vs-user-prompt) do not isolate
the claim, so none is worth a protocol.

## Infrastructure a direct test would need

1. **A fine-tune arm** (new to corpos-lab, which is inference-only today):
   - A QLoRA training path for a 7B on the 24 GB GPU (a 7B fits; larger models do
     not leave headroom for training state).
   - A merge-to-GGUF step, so the fine-tuned model serves on the one llama.cpp
     portal — this keeps the ONE-inference-portal invariant intact (no second
     server; the fine-tuned GGUF is swapped in like any other model).
   - Provenance for the trained weights: base model digest, training data digest,
     LoRA config, seed, and the merge step — the run-record discipline extended to
     the training side.

2. **A training-data design that keeps the manipulation faithful:**
   - The training corpus must be the **descriptive** atlas text (failure terrain,
     why it feels locally rational, the correct-navigation description) — NOT
     demonstrations of the correct action. Fine-tuning on demonstrations would
     test compliance-training, not the descriptive-comprehension mechanism the
     claim is about.
   - A conversion from the certified glyph/atlas entries into a training format
     that preserves the descriptive register.

3. **Matched-content controls, to separate layer from fine-tuning-at-all:**
   - A model fine-tuned on matched non-atlas text (same volume, same register,
     unrelated content), and optionally on scrambled atlas text (the existing
     `provenance/published-scoring/_shared/scramble.py` mechanism control, adapted). Without these, a lift
     under the atlas fine-tune cannot be distinguished from "any fine-tuning
     shifts decision behavior."

4. **The context-layer arm is already built** — H1 and the ground extension are
   the context-layer condition at matched content. The training-layer arm plugs
   into the same scenarios, scoring (blind raters + correct-target C), and
   sampler, so comparability is free once the trained model serves.

## Cost / risk note

QLoRA on a 7B is a bounded build, but it is genuinely new infrastructure (a
training loop, data conversion, control corpora, merge-and-serve, and training-
side provenance), plus a design question — what descriptive text represents the
atlas at the weight layer — that is itself non-trivial. It is not a
frictionless add to the inference rig.

## What this unblocks / holds

- The `comprehension-as-compliance` revision does **not** need this test: the
  draft already scopes layer invariance as an untested scope claim (§7 line 143,
  §8 line 159). Keeping that framing is honest.
- Hold H2 here. Revisit if a fine-tune arm gets built for another reason, or if a
  reviewer presses the layer-invariance claim hard enough to require a direct
  training-layer test. At that point this note is the build spec.

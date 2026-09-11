---
type: reference
status: orientation-only (not a published result)
date: 2026-09-11
---

# assay-blank-claude-definition-work (v10, v11a, v11b, v12) — orientation records

These four versions are copied byte-verbatim from the archived `lab-app/corpus`
tree. They are **orientation only**, not a published result and not a parity
target.

They ran on the **retired ollama runtime** (`mistral:latest` via
`ollama /api/generate`), which applies a hidden chat template and an unrecorded
sampler; `study.json` records no temperature, sampler, or seed. n was 4 per
cell. Under this lab's reproducibility contract (`studies/REPRODUCIBILITY.md`)
these conditions cannot be reproduced.

They exist here as the hypothesis and the material source for the clean study
that supersedes them: **`studies/wrong-path-field-source/`**. That study
re-runs the finding — verdict accuracy and terrain-field engagement are
dissociable — on the current llama.cpp rig, with a fully declared sampler,
three subjects, and the field-source check pre-registered as the primary
measure.

What these records show, read as orientation: on `cas-a` (v10), three correct
verdicts all cited the Marker axis, none the Scope field; two later
interventions (v11a, v11b) collapsed that scenario to zero correct, revealing
there was no reliable scope path under the shortcut.

# Target-unnamed arm — findings

Chain 558 `measurement-validity-controls`, task `target-unnamed-arm-disambiguation-vs-comprehension`. 2026-09-21.

## Question

The `ground_nonprescriptive` aid recovers correct action over baseline. Does that lift
reflect **comprehension** (the model infers the correct action from the grounding) or
**disambiguation** (the aid tells the model which artifact to act on)? The published study
names the target artifact in both the scenario and the aid. This arm removes the target
name from both, holding the situation and the grounding identical, and asks whether the
lift survives.

## Design

Plan c: the two classes whose target is a file/record that unnames cleanly —
`post-write-verification-absent` and `parent-state-check-bypass`.
`governed-operation-protocol-bypass` was held back (its target is a protocol that *is* the
situation, so it only relabels). It was not needed: the answer is clear from two classes.

- Arms: `baseline`, `ground_nonprescriptive`, each in **named** (the published
  grounded-non-prescriptive-aid runs) and **unnamed** (new runs, this arm).
- 3 subject models (Mistral-7B, phi-4, Qwen3.8-27B), 2 scenarios, n=16. 384 new completions.
- Same assay image as the published named runs (git-verified: `AssemblePrompt` for these
  conditions is byte-identical across the image builds, so the contrast carries no image
  version confound).
- Scored by **deepseek-flash**, the independent-family rater validated in chain
  `deepseek-rater-evaluation` (13/13 off-task, 0.95-0.97 C-boundary vs the human anchor).
  The **same rater scores both arms**, so the named-vs-unnamed contrast is clean.

## Result: comprehension, not disambiguation

Recognition axis — the response recognizes the verify/consult step (code ≠ I). This is the
right measure for a no-tools single completion, where the read cannot be *performed*, only
stated.

| | baseline | ground_nonprescriptive | lift |
|---|---|---|---|
| named   | 56/192 = 0.29 | 189/192 = 0.98 | **+0.69** |
| unnamed | 40/192 = 0.21 | 189/192 = 0.98 | **+0.78** |

The aid lifts recognition to 98% whether or not the target is named. The lift is if
anything larger in the unnamed arm (its baseline is slightly lower; the ceiling is
identical). Every cell shows the lift surviving unnaming, and several unnamed lifts exceed
the named ones. **Naming the target contributes nothing to the recovery** — the effect is
comprehension.

Worked example (post-write, phi-4, ground_nonprescriptive): the named response says "I
would read `server/config.py` and verify the timeout is 60"; the unnamed response says
"you would need to read the configuration file directly … verify the timeout." Same
comprehension, with and without the name.

## The C/Ii (performance) axis — an instrument note

Under strict rubric application deepseek scores almost no **C** (2/192 named, 0/192
unnamed). In a no-tools completion the model *describes* the verification rather than
*performs* it, which the rubric scores **Ii**. The published grounded-non-prescriptive
study scored such responses as **C** with two blind Claude raters at strict consensus:
its named C-lift was post-write 0.00→0.96, governed 0.00→0.75, parent-state 0.35→0.75
(`REVISION_INPUT_capc.md`). So Claude-C and deepseek-recognition agree on the strong
lift and its direction; Claude-C and deepseek-C **diverge** on the C/Ii boundary for the
no-tools verification classes — Claude treats a stated verification as the correct action
given the modality, deepseek requires a performed one. This divergence is itself an
instrument datum (it did not appear on the casg-direct anchor where the two matched at
0.95–0.97). It qualifies how the paper's C-numbers on these classes should be read, but
it does **not** change the disambiguation answer: named ≈ unnamed on every axis, under
either rater's bar.

## Provenance and cost

- **Subject models (local, free):** `Mistral-7B-Instruct-v0.3.Q4_K_M`, `phi-4-Q4_K_M`,
  `Qwen3.8-27B-Q4_K_M`, over the one llama.cpp portal, temperature 0.8, seeds 1–16,
  max_tokens 2048. Assay image `localhost/lab-grounded-glyph-probe@sha256:4844ac8b4703…`
  (prompt assembly git-verified identical to the published named runs' image `9d3e70…`).
- **Rater:** `deepseek-flash` (validated in chain `deepseek-rater-evaluation`), same rater
  on both arms, temperature 0, via the DeepSeek API.
- **Scoring cost:** 768 ratings (384 named + 384 unnamed) = **$0.40 peak / $0.20 off-peak**
  (652,672 cache-hit + 258,977 cache-miss input tokens, 268,389 output; the shared rubric
  prompt-caches). Generation cost was $0 — subjects ran locally.
- **n:** 16 per cell; 2 conditions × 2 classes × 2 scenarios × 3 models × 2 arms = 384 new
  completions plus the 384 published named responses re-scored on the same rater.

## Caveats

- Single validated rater (deepseek-flash). The within-rater named-vs-unnamed contrast is
  what answers the question, and the conclusion (named ≈ unnamed) is rater-independent in
  direction. For the paper's measure of record (Claude two-rater consensus), a Claude
  scoring of these arms would confirm the magnitude; it is not required to settle
  disambiguation vs comprehension.
- Recognition axis, per the no-tools modality. gopb not run (the two clean classes gave a
  clear answer).

## For the paper

Section 4.4: the ground-extension recovery is a **comprehension** result, robust to
removing the target artifact's name. Disambiguation does not drive it. If the C-rate on
these no-tools verification classes is reported, note that "correct action" there is a
*stated* verification, not a performed one.

## Artifacts

- Materials: `studies/target-unnamed-arm/<class>/materials/` (unnamed scenario + aid).
- Study defs: `studies/target-unnamed-arm/<class>/study.s<N>.<model>.toml`.
- Runs: `studies/target-unnamed-arm/<class>/runs/tun-*`.
- Slices + scores: `studies/target-unnamed-arm/scoring/`.

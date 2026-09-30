# Cross-class control synthesis — scrambled + off-target across 4 classes, 2026-09-09

## One sentence

The casg-direct control reading — "the glyph's effect is a prepended-structure
phenomenon, needing neither comprehension nor recognition" — does NOT generalize:
on formal-step the scrambled control fails (comprehension matters), on parent-state
the off-target control fails (recognition matters); casg-direct was the one class
where neither mattered, so reasoning from it alone overstated the mechanism.

## Why this run

The mechanism controls (chain 522) ran only on casg-direct, the most register-prone
class. This chain (523) ran the same two controls on the other three classes on the
same instrument, to establish a baseline instead of generalizing from the starkest
example. Instrument: image e159c2b8, vanilla /completion + the no-tools notice
identical across conditions, PROTOCOL-6 sampler, both models, 8 runs per cell, zero
tool-stall. Double-scored by two blind raters per class (84.9% agreement); all 29
disagreements adjudicated against the rubric. Two adjudication anchors: parent-state
"gate on the check" scores C (the main grid set this, 8C/8C); formal-step is
truncation-sensitive (the worked checklist hits the 1024-token cap, so C requires
the response to reach and incorporate the Provenance Stamp before cutoff).

## Per-class baseline and control behaviour (Qwen, the cleaner model; read direction, n=8)

| class | baseline | glyph effect | scrambled | off-target | reproduces? |
|---|---|---|---|---|---|
| casg-direct | 7C 1Ii | register shift, 4 Ii | 5 Ii | 4 Ii | scrambled YES, off-target YES |
| formal-step | 8C | register shift, 5 Ii (3C) | 7 C (acts) | 2C 4 Ii | scrambled NO, off-target YES |
| parent-state | 1C 5I | gating, 8 C (from 1) | 7 C | 0C 8 I | scrambled YES, off-target NO |
| conditional-gate | 8C ceiling | 7C 1Ii | 5C 3Ic | 6C | ceiling — uninformative |

Register-proneness baseline (glyph Ii on Qwen): casg-direct 4/8, formal-step 5/8,
conditional-gate 1/8, parent-state 0/8 (parent-state runs the other way — the glyph
raises correct action rather than suppressing it). casg-direct and formal-step are
the register-prone classes; parent-state is a gating class; conditional-gate ceilings.

## Reading — the generalization verdict

**The casg-direct reading is class-specific, not a law.**

- **formal-step: comprehension matters.** The real glyph suppresses action into the
  analysis register (3C/5Ii), but the SCRAMBLED glyph does not (7C — acts like the
  8C baseline). Incomprehensible word-salad in the glyph's shape does NOT reproduce
  the effect here. The off-target (a coherent other-class glyph) does (2C/4Ii). So on
  formal-step the effect needs comprehensible content, though not a matching one.
- **parent-state: recognition matters.** The glyph and the scrambled glyph both make
  the model gate on the parent check (8C, 7C, up from baseline 1C), but the OFF-TARGET
  glyph does not (0C, 8I) — a coherent glyph about a different decision fails to cue
  the milestone check. Here relevance/recognition is in the loop, while comprehensible
  content is not strictly required (scrambled still works, likely because it retains
  the class's own vocabulary in shuffled order).
- **casg-direct: neither mattered.** Scrambled and off-target both reproduced the
  register shift. This is the outlier that the single-class study generalized from.
- **conditional-gate: ceiling.** Baseline already acts 8/8; the class cannot measure
  a scaffold effect (consistent with the main grid).

**Net:** across the informative classes, EITHER comprehension (formal-step) OR
recognition (parent-state) is load-bearing — just not both, and not the same one.
The "mere prepended structured preamble" mechanism holds only on casg-direct. The
honest cross-class statement is that the glyph's effect is class-dependent, and the
control that fails tells you which ingredient that class relies on.

## Mistral (texture, messier — smaller model)

formal-step Mistral: baseline 8C, glyph 5C (some suppression), scrambled 8C,
off-target 8C — only the real matching glyph suppresses at all, so on Mistral
formal-step both comprehension AND recognition look load-bearing. parent-state and
conditional-gate on Mistral are Ic/I-heavy (the model reasons out of gates or fails
to recognize), too noisy to carry a clean control reading. The Qwen cells carry the
result; Mistral is consistent with "class-dependent and fragile on the small model."

## Consequences

- Bounds the chain-522 claim: report it as casg-direct-specific, not general.
- Feeds the CaPC revision (Section 3/4): the mechanism is not a single "comprehension"
  or "structure" story; it is decision-class-dependent, and the paper must say so.
- Feeds the q2 follow-up (chain 513): an honest paper would present the class-dependent
  control pattern, not a single clean mechanism.

## Caveats

n=8 per cell (read direction). formal-step C-vs-Ii is truncation-sensitive; the large
gaps used here (glyph 3C vs scrambled 7C; off-target 0C vs glyph 8C on parent-state)
are well outside truncation noise. conditional-gate ceilings and is set aside.
Scrambled glyphs retain the class's own (shuffled) vocabulary, so "scrambled" removes
syntax/comprehensibility but not topical keywords — relevant to why scrambled still
worked on parent-state.

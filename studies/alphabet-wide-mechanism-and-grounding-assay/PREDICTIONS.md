# Predictions — alphabet-wide mechanism & grounding assay

*Committed before the runs. Kept honestly (INQUIRY.md): predictions exist because
surprise is where the learning is, never as a commitment that binds the analysis.
Never referenced by any `study.toml` or materials file, and never placed in
anything a subject or rater model sees. Author: Claude (executing agent),
2026-09-15.*

## The headline prediction

The mechanism-type distribution across the corpus is **heterogeneous** — there is
no single mechanism. Some classes are carried by comprehension, some by
recognition, some by mere prepended structure. This is the result that reshapes
the comprehension-as-compliance paper's §3/§4: the honest claim is class-dependent
typing, not one mechanism.

## The strong-scramble prediction (the sharpest new test)

The prior typings used the **weak** scramble, which kept the class's topical
vocabulary in shuffled order. The cross-class note flagged that this is likely why
parent-state's scrambled control "still worked." Under the **strong** scramble
(lorem tokens, no keyword leak), I predict some prior typings shift toward
comprehension:

- **casg-direct** — prior "mere structure" (weak scrambled reproduced). Under the
  strong scramble I expect it may **not** reproduce, moving casg-direct toward
  comprehension or recognition. If it still reproduces with pure lorem, "mere
  structure" is real and strong.
- **parent-state** — prior "recognition" (off-target failed, weak scrambled
  reproduced). Under the strong scramble the scrambled cell may now **fail**,
  making comprehension load-bearing too. Off-target should still fail (recognition
  stays in the loop).
- **formal-step** — prior "comprehension" (weak scrambled already failed). The
  strong scramble should fail at least as hard — comprehension confirmed, more
  cleanly.

So the corpus may show **comprehension is more prevalent than the weak-scramble
results suggested** — which, if it holds, strengthens the comprehension-as-
compliance thesis rather than weakening it.

## Per-class type predictions (untyped entries — genuine forecasts)

- **post-write-verification-absent** — recognition-leaning: the model must
  recognise the "acted, not verified" shape; a coherent off-target may fail.
- **casg-delegate** — like casg-direct, possibly structure or recognition; the
  delegation nuance may need comprehension to route the hand-off vs edit directly.
- **governed-operation-protocol-bypass** — comprehension-leaning: the
  anachronicity point (a derived decision is not a consultation output) is a
  content idea a scrambled block should not carry.
- **structural-ceiling-bypass** — recognition-leaning: recognising the "at the
  ceiling" state is the cue; the architectural-response content may be secondary.
- **discovery-event-non-recording** — comprehension-leaning: the durability point
  (record it so a future agent finds it) is content the scramble should destroy.
- **conditional-gate-uniform-default** — likely uninformative (ceilings on Qwen);
  read only where Mistral calibrates.

## The grounded contrast

Across the corpus, **ground ≈ domain-directive** on the classes that calibrate —
a descriptive grounded statement converts correct action as well as a command.
This generalises the H1 / ground-extension result and is the corpus-level
comprehension-as-compliance signal. I expect the smaller models to show the
occasional class where the command does extra work (the casg-Mistral pattern).

## Where I expect to be surprised

- Whether the strong scramble flips casg-direct and parent-state (above) — the
  single most informative outcome for the paper.
- Whether the four newly authored classes calibrate below ceiling at all on the
  strong model, or (like casg-direct/formal-step) ceiling on Qwen and only type on
  the small models.
- Whether the "list the steps you would take" ask holds the action-omission trap
  live across all four new classes, or whether some ceiling (models reliably
  include the omitted step unguided).

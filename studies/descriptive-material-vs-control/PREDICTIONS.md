# Predictions — H1 (descriptive material vs control)

*Committed before the H1 runs. Kept honestly (INQUIRY.md): predictions exist
because surprise is where the learning is, never as a commitment that binds the
analysis. This file is **never** referenced by any `study.toml` or materials file,
and is never placed in anything a subject or rater model sees. Author: Claude
(executing agent), 2026-09-15.*

## What H1 tests

The mirror of the ground extension, run on gating classes from the floor instead
of ceiling classes. Three certified gating glyphs, three scenarios each, five
conditions (baseline, glyph, imperative, ground, domain-directive), three local
models, n=8. The question: does descriptive decision-point material **lift**
correct action from a below-ceiling baseline, and does grounding convert
recognition into action as it did in the suppression regime?

## Calibrated baselines (from the item-12 weak-shelf probes, 2026-09-15)

- **parent-state-check-bypass** — ~10% correct-target (Q1: 1/8 C on Qwen); ~90%
  fire. Strong under-fire.
- **post-write-verification-absent** — ~0% correct-target (8/8 Mistral, 8/8
  phi-4 fire); ~100% fire. Maximal under-fire.
- **initiative-task-preexistence-gate** — ~50% fire (Mistral 5/8, phi-4 3/8 on
  the v2 probe); the most measuring room.

## Predictions

- **baseline** — low correct-target C, near the calibrated baselines above. The
  27B may ceiling on some class×scenario cells (dropped from lift analysis).
- **glyph** (descriptive, domain-free) — lifts recognition; correct-target C
  lifts only partially. On the smaller models the glyph's does-not-fire clause
  inflates Ii (recognition without action) — the analysis pull the priors found,
  now from a floor instead of a ceiling.
- **imperative** (directive, domain-free) — tracks the glyph. Content, not form,
  carries it (the settled finding), so also only a partial lift. glyph ≈
  imperative.
- **ground** (descriptive, domain-specific) — correct-target C lifts
  substantially toward ceiling. Grounding is the execution lever, converting the
  floor case as it recovered the ceiling case. The 7B on the hardest class
  (post-write) may still lag (the ground-extension's casg-Mistral exception).
- **domain-directive** (directive, domain-specific) — also lifts substantially.
  The open question `register-shift-followup` left under-determined: does the
  descriptive ground suffice as well as the commanding domain-directive?
  Prediction: **ground ≈ domain-directive on the larger models; domain-directive
  > ground on the 7B for the hardest class.** A ground ≈ domain-directive result
  is direct support for comprehension-as-compliance from the floor.
- **Across classes** — the lift is largest where the baseline is lowest
  (post-write, ~0%), and moderate on initiative-task-preexistence-gate (~50%).

## Outcomes that would reframe the result

- **Ground does not lift (stays near the floor).** Then grounding is not the lift
  lever from below — it recovered the ceiling case by removing suppression, and a
  floor case needs something else (a command, or the glyph-plus-ground
  combination). An honest disconfirmation of "ground is the execution lever" in
  the lift direction.
- **Only the domain-directive lifts; the descriptive ground does not.** Then it
  is compliance, not comprehension: a domain command acts, a domain description
  does not. A partial disconfirmation of comprehension-as-compliance.
- **The glyph/imperative lift as much as the grounded aids.** Then domain
  grounding is not the discriminator for gating classes — recognition alone
  converts to action from the floor, unlike the ceiling regime.

## Where I expect to be surprised

- Whether the ~0% post-write class lifts under grounding on the 7B, or stays
  derailed like casg-Mistral did.
- Whether initiative-task-preexistence-gate, already at ~50%, has enough room to
  show a clean grounding lift or ceilings early on the 27B.
- Whether the deterministic parser and the blind raters agree as cleanly here
  (three actions: read-first, read-back, withhold-readiness) as they did on the
  single execute-vs-analyze bar (99.4%).

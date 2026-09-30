# Predictions — setup vs agentic loop (CaPC item 1.6)

**Chain:** 550, task 2. **Date:** 2026-09-17. Written before the run.
**Kept honestly, not as a commitment** (INQUIRY.md): surprise is where the learning is.
These predictions must never enter any file a subject or a rater can see.

## Headline prediction

**Analysis-mode partly survives the loop, but drops materially in the suppression
classes.** The program's null (the register shift is a content effect, not a format
effect) plus the pilot (Qwen3.8 wants to act and stalls) point both ways at once, so the
honest forecast is a split, not a clean win for either arm.

## Per-outcome forecast

1. **Suppression classes (casg-direct, formal-step) — loop removes most of the `Ii`.**
   Here the raw setup takes a model that would act at baseline and, under the aid, pushes
   it into recognition-without-execution. Given real tools, the model should act instead of
   narrate. Predicted: loop `Ii` well below raw `Ii`, the mass moving to `C`. If this holds,
   the setup is a driver **for these classes** and the suppression finding must be scoped to
   raw completion.

2. **Lift classes (post-write, governed, parent-state, structural) — small change.**
   In raw completion the aid already lifts `C` here, so there is little `Ii` to remove. The
   loop should keep `C` high. Predicted: loop `Ii` ≈ raw `Ii` (both low), `C` stays high.

3. **Residual analysis-mode persists somewhere.** At least one class keeps a non-trivial
   loop `Ii`: the model reads the sandbox, recognises the obligation, and still ends on a
   `FINAL` summary without issuing the acting call. If so, analysis-mode is not purely a
   setup artifact and the behaviour generalises, in part, to a minimal agentic setup.

4. **The setup effect is content-general, not glyph-specific.** Whatever the loop does to
   `Ii`, it does about equally under `glyph_only` and `imperative_only`. This follows the
   program's matched-content null. A glyph-specific setup effect would be a surprise.

5. **Where `Ii` falls, it moves mostly to `C`, not `Ic`.** If instead it moves to `Ic`
   (the model acts, wrongly), the loop is measuring tool-use competence, not the release of
   a suppressed correct action — a caveat to flag, not the target reading.

## The two clean verdicts this feeds

- **If the loop removes analysis-mode broadly** (outcomes 1 + 3 weak): the setup is a
  driver. The CaPC revision scopes the analysis-mode / register-shift findings to the
  raw-completion setting and softens any "Agents" scope word.
- **If analysis-mode persists broadly** (outcome 3 strong, 1 weak): the finding survives
  the setup and the scope claim to agentic settings strengthens.

The realistic result is between these — a per-class map, read cells not counts (n=16,
95% CI ≈ ±0.12). The map itself is the contribution to the scope framing.

## Watch-items

- **Tool-use competence confound.** A small model failing to format a tool call scores `N`
  or `Ic` for a mechanical reason, not a cognitive one. Watch the `N` rate per arm; a high
  loop `N` on Mistral would mean the loop measured tool-formatting, not analysis-mode.
- **Step-cap truncation.** A run that hits the 10-step cap without a `FINAL` is `N`, not
  `Ii`. Watch the cap-hit rate; raise the cap and re-smoke if it is non-trivial.
- **Sandbox realism.** If a scenario's sandbox omits a file the model reaches for, the
  model stalls the way the pilot did. Smoke each scenario's sandbox before batching.

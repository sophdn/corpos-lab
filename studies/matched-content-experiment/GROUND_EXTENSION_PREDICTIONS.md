# Predictions — ground extension (form x grounding 2x2)

*Committed before the ground-extension runs. Kept honestly (INQUIRY.md):
predictions exist because surprise is where the learning is, never as a
commitment that binds the analysis. Never referenced by any study.toml or
materials file, and never placed in anything a subject or judge model sees.
Author: Claude (executing agent), 2026-09-14.*

## What this extends

The matched-content grid tested three conditions (baseline, glyph, imperative)
and found: the register shift is a content effect, not a format effect (the
domain-free imperative suppresses execution as much as the domain-free glyph).
This extension adds the two domain-specific cells to complete a 2x2 that varies
**form** (descriptive vs directive) and **grounding** (domain-free vs
domain-specific), with the propositional content held constant.

|             | domain-free | domain-specific    |
|-------------|-------------|--------------------|
| descriptive | glyph       | ground             |
| directive   | imperative  | domain-directive   |

Classes: casg-direct, formal-step-context-bypass. Models: Mistral-7B-Instruct-v0.3,
Qwen3.8-27B. n=24 per cell. Correct-target execution bar, identical across all five
conditions (baseline, glyph_only, imperative_only, ground_only, domain_imperative).

## The reference the grid already set

In the clean matched grid, baseline is at ceiling on both classes and both models
(the model does the correct action unguided), and the abstract aids suppress it:
casg glyph drops Qwen 8->4 and Mistral 8->0; formal-step glyph drops Qwen 8->4 and
Mistral 8->5. So this study measures **suppression by the aid**, read against the
baseline ceiling, and whether a domain-specific aid avoids it.

## The main prediction (grounding is the lever, not form)

- **Baseline** stays at ceiling (the un-suppressed reference).
- **Glyph (descriptive, domain-free)** suppresses execution into analysis-mode Ii,
  replicating the grid.
- **Imperative (directive, domain-free)** suppresses about as much as the glyph,
  replicating the content-not-format result. Form is not the lever at domain-free.
- **Ground (descriptive, domain-specific)** does NOT suppress: execution stays near
  the baseline ceiling. Grounding the description in the domain keeps the model
  executing rather than analyzing.
- **Domain-directive (directive, domain-specific)** also does not suppress.
- **The comprehension test:** ground is about equal to domain-directive. If a
  descriptive domain-grounded aid keeps execution as well as a domain command
  does, comprehension is sufficient and form does not carry it. That supports
  comprehension-as-compliance.

So the predicted main effect is grounding (domain-free suppresses, domain-specific
does not), with little effect of form at either grounding level.

## The outcomes that would reframe or kill the paper

- **Domain-specific aids also suppress (ground and domain-directive low).** Then
  grounding is not the lever: prepending any decision-aid derails the model, and the
  q2 "ground recovers execution" result must have come from something the alone-design
  removes (the glyph-plus-ground combination, or the earlier ground handing over the
  finished artifact to copy). This is a "we were wrong about the ground" result and
  the paper says so.
- **Only the domain-directive keeps execution; the descriptive ground suppresses.**
  Then it is compliance, not comprehension: a domain command works, a domain
  description does not. A real partial disconfirmation of comprehension-as-compliance.

## Where I expect to be surprised

- Whether Mistral, which suppressed hardest under the glyph (casg 8->0), recovers
  under the domain-specific aids as cleanly as predicted, or stays derailed.
- Whether formal-step (a bypass class whose baseline already works from the spec)
  behaves like casg, or whether its weaker glyph suppression leaves less room to see
  a grounding effect.
- A residual the design cannot remove: the glyph is a three-axis structured block,
  while the ground and domain-directive are prose. The glyph-vs-ground contrast
  therefore mixes grounding with structure. The mechanism controls (scrambled and
  off-target) bound the structure question separately; the imperative-vs-
  domain-directive contrast (prose vs prose) is the cleaner grounding read.

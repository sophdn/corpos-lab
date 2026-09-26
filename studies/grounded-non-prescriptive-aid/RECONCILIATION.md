# Library reconciliation — grounded, non-prescriptive aid

**Chain:** 549 `grounded-non-prescriptive-aid` (glyph-research), task 5
`reconcile-library`. **Date:** 2026-09-17.
Per INQUIRY.md's library-reconciliation step. Source of record for every verdict:
`studies/grounded-non-prescriptive-aid/FINDINGS.md` (landed a47d746). The library is the
universal catalogue on corpos-toolkit; verdicts recorded via the knowledge meta-tool
(`library_update`), not in any file-based library.

The study's one-line result: a grounded aid that names the file and situation but not the
correct outcome recovers most of the execution the full ground recovers, far above
baseline. Naming the outcome is at-most-additive and model-dependent — phi-4 and Qwen
recover the full lift from grounding alone; only the smallest model needs the target
named, and only on produce-after-external-consult classes.

## Entries this study bears on — verdicts recorded

- **153.45 Friston (2010), the free-energy principle — CONFIRMED as a grounding claim;
  target-specification de-confounded.** The 2026-09-14 reconciliation read "domain
  grounding is the lever," but that rested on a ground that also stated the outcome, so
  grounding was confounded with target-specification (CaPC review 1.3). This study
  removed the outcome: grounding qua grounding recovers execution, the agent infers the
  action from the domain, and naming the target is a small-model top-up. Sharpens the
  entry from "grounding is the lever" to "grounding, not being-told-the-answer, is the
  lever." Verdict appended to `what_it_answers`, cited to FINDINGS.md.

- **153.43 Klein (1998), the recognition-primed decision model — CONFIRMED; target
  de-confounded.** H1's qualification ("the material must be GROUNDED to produce action")
  rested on the same confounded ground. This study shows a grounded situation-type
  description suffices for action even without the response named — the agent generates
  the single course of action itself, exactly the RPD prediction. Strengthens the
  CONFIRMED reading. Verdict appended, cited to FINDINGS.md.

- **006.54 Wang (2026), narrative priors — CONFIRMED (purest in-program test).** Wang's
  claim is that descriptive framing steers behaviour independent of explicit instruction.
  This study is the cleanest test: a grounded aid with no instruction and no stated
  outcome recovers correct action far above baseline. Qualification: on the weakest model
  an explicit outcome still adds, so "independent of explicit instruction" holds fully for
  mid and large models and partially for the smallest. Verdict appended, cited to
  FINDINGS.md.

## Entries noted, no verdict change

- **006.64 Pecher (2026), prompt-format sensitivity as a content confound — no change.**
  Already CONFIRMED by chain 548 (neutral-prefix control). This study's cut is orthogonal:
  content-with-outcome versus content-minus-outcome, not content versus no content. It is
  consistent with Pecher (form is not the lever — ground ≈ domain_imperative) but adds no
  new verdict to that entry.

- **152.1 Gibson (affordance), 153.41 Polanyi (tacit recognition) — noted, no change.**
  The study is consistent with a directly-perceived-affordance reading of grounding (the
  action is inferred from the domain, not reasoned from a stated target), but it does not
  test these entries' specific predictions. Left intact.

## Net

Three verdicts recorded, all confirming — the de-confounding strengthens the grounding
reading rather than overturning it. Nothing retired; no promotion overturned. The library
does not drift from the evidence: the ground-extension entries (Friston, Klein) now read
as grounding claims rather than being-told-the-answer claims, and Wang's
framing-without-instruction warrant is backed by the cleanest in-program control the
program has run.

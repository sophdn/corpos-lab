# Predictions and pre-registered analysis

Kept per INQUIRY.md: predictions are written before looking because surprise is
where the learning is, not as sealed commitments. No confidence scores. This
file is never placed in any prompt a subject or judge model sees.

## Fixed before the run

- **Conditions:** baseline, annotated_instrument, cartographer_instrument,
  cartographer_scan_instrument.
- **Primary model:** Qwen3.8-27B. Second model Qwen2.5-32B if time permits.
- **n:** 30 runs per condition per model (30 unique seeds). Far above the
  original n=1. The predicted effect is a large, categorical, structural
  pattern, so 30 is comfortable; no escalation is planned (the run is
  time-bounded and a large effect does not need more).
- **Taboo classes** are fixed in `taboo_set.json`: 7 first-principles, 1
  meta/routing, 1 corpus-empirical, 1 session-close.
- **Primary analysis:** the interaction of condition and taboo class on C4
  coverage. Secondary: C2 slug-citation rate by condition (deterministic).

## Predictions

1. **C2 (slug citation).** Annotated cites registry slugs in most runs;
   cartographer and cartographer+scan cite essentially none. Baseline: few,
   since the registry is in the scenario and a model may echo names.
2. **C4 on first-principles taboos.** Annotated, cartographer, and
   cartographer+scan all cover most of the seven first-principles taboos. The
   cartographer method may even add a gate the annotated method does not (the
   original saw a live-locus-verification gate appear only in the cartographer
   run). Baseline covers fewer and less reliably.
3. **C4 on the non-first-principles taboos.** This is the finding. Annotated
   covers `triggers-routing`, `investigation-advisory-fix-durability`, and
   `known-constraint-documentation` (its method enumerates the registry).
   Cartographer misses them: first-principles analysis of the repair act does
   not surface an upstream routing obligation, an accumulated-observation
   durability distinction, or a session-close recording obligation.
4. **The scan (C6).** Cartographer+scan recovers `triggers-routing` (routing
   class) and `known-constraint-documentation` (session-close class), because
   the scan names those two classes. `investigation-advisory-fix-durability` is
   neither; it is predicted to remain the residual miss.
5. **Interaction.** The cartographer coverage deficit relative to annotated is
   concentrated on the non-first-principles taboos and is near zero on the
   first-principles ones. A significant condition-by-class interaction is the
   registered primary result.

## Honest notes on the design and its limits

- **Information is not matched, by design.** The annotated condition carries the
  registry; the cartographer conditions do not. This is the definitional
  difference between the formats, not a confound (see DESIGN.md). A consequence
  is that the corpus-empirical coverage gap is partly expected: a hazard that is
  "only apparent from accumulated observation" cannot be derived without the
  observations. The non-trivial questions are (a) how much of the
  non-first-principles hazard set a capable model recovers from first principles
  anyway, and (b) whether naming the hazard classes (the scan) recovers more.
- A pilot smoke with the registry placed in the shared scenario (matched
  information) washed the gap out: the model echoed the visible corpus-empirical
  entries. That result is recorded here as the reason the registry is withheld
  from the cartographer conditions; it is itself a finding about single-turn
  in-context presence versus consult-by-choice.
- The single scenario (one defect domain) is a pilot dimension: a
  direction/category finding on one specimen, not a claim across domains.

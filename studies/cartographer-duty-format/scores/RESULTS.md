# Cartographer-duty-format — results (Qwen3.8-27B, n=30 per condition)

Run: `runs/qwen38/` (image `sha256:3bf42504829e…`, GPU RTX 3090, 120 rows, no
truncation, throughput 37.9–40.8 tok/s). Primary coverage judge: Claude
(Opus 4.8), blind to condition, four batches over a shuffled, de-labelled set
(`scores/blind/`, map in `scores/blind_map.json`). Slug citation (C2) is scored
deterministically.

## C2 — slug citation (deterministic, exact)

| condition | runs citing ≥1 slug | mean slugs cited |
|-----------|--------------------:|-----------------:|
| annotated_instrument | 30/30 | 10.00 |
| baseline | 0/30 | 0.00 |
| cartographer_instrument | 0/30 | 0.00 |
| cartographer_scan_instrument | 0/30 | 0.00 |

The annotated method cites all ten registry slugs in every run; every other
condition cites none. The transmission path is exactly as described: annotated
duties carry the canon into their gate bodies, cartographer duties do not.

## C4 — hazard coverage, by class (k/n, Wilson 95%)

| condition | first-principles (7 taboos) | non-first-principles (3 taboos) |
|-----------|-----------------------------:|--------------------------------:|
| baseline | 127/210 [0.54, 0.67] | 4/90 [0.02, 0.11] |
| annotated_instrument | 210/210 [0.98, 1.00] | 90/90 [0.96, 1.00] |
| cartographer_instrument | 134/210 [0.57, 0.70] | 1/90 [0.00, 0.06] |
| cartographer_scan_instrument | 133/210 [0.57, 0.70] | 52/90 [0.47, 0.67] |

## Primary result — the condition × class interaction

The annotated − cartographer coverage gap is **+0.36 on first-principles
hazards** and **+0.99 on non-first-principles hazards**. The cartographer
deficit is concentrated, almost entirely, on the hazards that are not derivable
from the repair act. Fisher exact on the pooled non-first-principles taboos
(annotated vs cartographer): **p = 2×10⁻⁵¹**.

This reproduces the original n=1 finding as a quantified rate: the cartographer
method structurally misses the corpus-empirical and meta hazards that the
annotated method transmits through the canon.

## The scan (the fix the original never ran)

Cartographer + scan versus plain cartographer, per non-first-principles taboo:

| taboo | class | scan targets it | cartographer | + scan | Fisher p |
|-------|-------|:---------------:|-------------:|-------:|---------:|
| triggers-routing | meta / routing | yes | 0/30 | 27/30 | 9.2×10⁻¹⁴ |
| known-constraint-documentation | session-close | yes | 1/30 | 25/30 | 1.2×10⁻¹⁰ |
| investigation-advisory-fix-durability | corpus-empirical | no | 0/30 | 0/30 | 1.0 |

The scan recovers exactly the two hazard classes it names — routing and
session-close — and does not recover the one hazard it does not name, the pure
corpus-empirical durability distinction. This is the pre-registered prediction,
confirmed. Naming a hazard class lets first-principles derivation reach it; a
hazard whose badness is "only apparent from accumulated observation" and belongs
to no named class stays out of reach.

## Honest nuances

1. **Two nominally first-principles taboos behave like the hard ones.**
   `fix-locus-identification` (canonical source vs call site: cartographer 7/30)
   and `fix-attempt-root-cause-reassessment` (revised root cause before a second
   attempt: cartographer 2/30) are missed by the cartographer method almost as
   often as the non-first-principles taboos. Only the canon (annotated) covers
   them reliably (30/30 each). The first-principles / not-first-principles split
   predicts the broad pattern — the non-first-principles gap is near-total — but
   these two subtle hazards sit closer to the hard end than their label implies.
2. **The cartographer method does not raise coverage over no method.**
   Cartographer first-principles coverage (134/210) is statistically
   indistinguishable from the no-method baseline (127/210; Fisher p = 0.55). The
   method's distinctive property is self-sufficiency — no canon, no slug
   citations (C2) — not higher coverage. Only the annotated method, carrying the
   canon, reaches full coverage.

## Scope

One defect domain (the `calculate_average` empty-list defect), one model
(Qwen3.8-27B), single-turn generation. A direction-and-rate finding on one
specimen, not a claim across domains or models.

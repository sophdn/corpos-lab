# Findings — alphabet-wide mechanism & grounding assay

**Chain:** 547 `alphabet-wide-mechanism-and-grounding-assay`, task 5
`score-analyze-and-hand-off`. **Date:** 2026-09-16.
**Reads on from:** `PROTOCOL.md` (conditions + typing logic), `SCOPE_AND_CALIBRATION.md`
(calibrating subset + ceiling cells), `PREDICTIONS.md` (pre-registered, reconciled below).

## What ran

The full grid: 7 conditions × 20 scenarios × 3 models × n=8 = **3360 responses**,
one uniform config (image `sha256:ee876311`, GPU, max_tokens 2048, strong scramble
`--vocab-swap --neutralize-title`, seeds 1–8). Scoring: two independent blind
Claude raters per class, one condition-blind correct-target bar per class, primary
measure **strict-consensus C** (both raters code C). Conditions map to the grid file
names as: baseline, glyph_only (descriptive domain-free), imperative_only (directive
domain-free), ground_only (descriptive domain-specific), domain_imperative_only
(directive domain-specific), scrambled_glyph (strong lorem scramble), off_target_glyph.

## Headline verdict

The corpus mechanism-type distribution is **heterogeneous but comprehension-dominated**.
Among the six classes where the descriptive glyph produces a measurable behavioural
effect with clean controls, **comprehension is the load-bearing ingredient in five**;
only one class (casg-direct) is carried by mere prepended structure. The strong
scramble makes comprehension **more** prevalent than the earlier weak-scramble
cross-class study found — it flips parent-state from recognition to comprehension and
leaves casg-direct's structure typing intact. This strengthens the comprehension-as-
compliance (CaPC) thesis rather than fragmenting it.

## Mechanism typing per class

Typing rule (PROTOCOL): read the baseline→glyph effect on the cells where the baseline
leaves room; the scrambled cell reproducing the effect ⇒ comprehension is **not**
load-bearing; off-target reproducing ⇒ recognition is **not** load-bearing; both ⇒
mere structure. Cells read as strict-consensus C. Qwen ceilings dropped per calibration.

| class | calibrating model(s) | baseline→glyph | scrambled | off-target | **type** |
|---|---|---|---|---|---|
| **post-write-verification-absent** | qwen, mistral, phi4 | 0 → 22/8/20 (lift) | 0 all | 0 all | **comprehension** (cleanest) |
| **governed-operation-protocol-bypass** | qwen, mistral, phi4 | 5/0/0 → 15/8/16 (lift) | 0 all | 7/1/0 (≤base) | **comprehension** |
| **parent-state-check-bypass** | mistral, phi4, qwen-s1 | 0 → 7 (mistral, lift) | 0 (fails) | 0 (fails) | **comprehension** (flipped, see below) |
| **formal-step-context-bypass** | qwen | 8 → 5 (suppress) | 8 (fails to reproduce) | 6 | **comprehension** |
| **structural-ceiling-bypass** | mistral | 5 → 12 (lift) | 6 (≈base) | 5 (=base) | **comprehension** |
| **casg-direct** | mistral, phi4 | 8 → 0/2 (suppress) | 1/6 (reproduces) | 2/0 (reproduces) | **mere structure** |
| discovery-event-non-recording | mistral only | 10 → 6 (small drop) | 6 (reproduces) | 10 (=base) | weak / structure-leaning — inconclusive |
| conditional-gate-uniform-default | mistral only | 0 → 0 (no domain-free effect) | 1 | 0 | uninformative (domain-free glyph inert) |
| casg-delegate | — | 16 → 16 (ceiling) | 16 | 16 | uninformative (no room) |
| initiative-task-preexistence-gate | — | high baseline, no room | ~high | ~high | uninformative (no room) |

**Distribution among calibrating classes:** comprehension 5, mere structure 1,
uninformative/dropped 4 (three by ceiling, one — discovery — weak and inconclusive).

### Notes on the two regimes
- **Lift classes** (post-write, governed, parent-state, structural, conditional-gate):
  baseline low, glyph raises correct action. Comprehension shows up as the scrambled
  and off-target cells failing to reproduce the lift.
- **Suppression classes** (casg-direct, formal-step): baseline at ceiling, the domain-
  free glyph *lowers* correct action on the small models. For casg-direct the scramble
  and off-target lower it just as much → the effect is a content-independent distractor,
  i.e. **mere structure**. For formal-step the scramble does *not* reproduce the drop →
  the drop needs the real content → **comprehension**.

## The pre-registered watch-item: the strong-scramble flip

**parent-state flipped from recognition to comprehension.** Under the weak, keyword-
leaking scramble the parent-state scrambled control had reproduced the effect (typing it
"recognition"). Under the strong lorem scramble the scrambled cell now **fails**:
mistral glyph 7 vs scrambled 0; qwen glyph 20 vs scrambled 15 (= baseline, no lift);
phi4 glyph 20 vs scrambled 9. Comprehension is load-bearing after all. Off-target also
fails (recognition stays in the loop too).

**casg-direct did not flip.** With pure lorem the scrambled cell still reproduces the
suppression (mistral glyph 0, scrambled 1; off-target 2). Its "mere structure" typing is
real and strong, exactly the pre-registered alternative.

Net: comprehension is more prevalent under the strong scramble than the weak-scramble
result suggested — the sharpest evidence for CaPC in the corpus.

## The grounded contrast (ground vs domain-directive)

Across every class that calibrates, a descriptive domain **ground** converts correct
action as well as, or better than, a commanding domain-**directive** — mood is not the
active ingredient:

| class (calibrating cells) | ground_only | domain_imperative_only |
|---|---|---|
| post-write (from floor 0) | 24 / 24 / 24 | 24 / 21 / 20 |
| governed (from 0–5) | 16 / 16 / 16 | 16 / 16 / 16 |
| parent-state (mistral/phi4) | 19 / 22 | 13 / 19 |
| conditional-gate (mistral, floor 0) | 8 | 8 |
| casg-direct (mistral/phi4) | 8 / 8 | 8 / 8 |

This generalises the H1 ground-extension result corpus-wide: it is the corpus-level
comprehension-as-compliance signal.

## Content vs format (glyph vs imperative)

Mixed, and model-dependent. On the smallest model (Mistral) the directive imperative
does extra work over the descriptive glyph in the lift classes — post-write mistral
glyph 8 vs imperative 23; parent-state mistral glyph 7 vs imperative 23 — while on Qwen
and phi-4 the two run close. This is the "command does extra work on the small model"
(casg-Mistral) pattern, as predicted.

## Inter-rater agreement

Raw agreement 0.89–0.997 (median ≈ 0.96); Cohen's κ 0.66–0.99. The highest raw-
agreement classes are the ceiling classes (casg-delegate, conditional-gate, initiative),
where near-zero code variance inflates raw agreement and attenuates κ (casg-delegate
κ 0.665 on ~all-C data). The graded classes sit near 0.90. No slice shows an anomalous
"too-perfect" agreement, so the scoring-process fault below did not distort the data.

## Reconciliation against PREDICTIONS.md

- **Heterogeneous distribution** — partially confirmed: heterogeneous, but comprehension-
  dominated rather than evenly split across the three types.
- **Strong-scramble flip** — confirmed. parent-state → comprehension (predicted);
  casg-direct stays structure (predicted alternative); formal-step comprehension confirmed
  more cleanly (predicted).
- **Per-class forecasts** — governed comprehension ✓; conditional-gate uninformative ✓;
  casg-delegate ceiling/uninformative ✓; **structural** predicted recognition, measured
  comprehension ✗; **discovery** predicted comprehension, measured weak/structure-leaning
  ✗; **post-write** predicted recognition-leaning, measured the cleanest comprehension ✗.
- **Grounded contrast (ground ≈ domain-directive)** — confirmed corpus-wide.
- **Net** — comprehension more prevalent than the weak-scramble study suggested ✓.

## Caveats (honest)

1. **Scoring process.** The first automated rating pass had the blind-rater subagents
   self-parallelise into a shared temporary namespace and race — one rater's output file
   was rewritten mid-run. That pass was discarded and every class re-rated under a strict
   work-alone rule (no sub-agents, no shared scratch). All 34 slice-rater files were then
   verified for exact id coverage and valid codes; agreement shows no contamination
   signature. A tooling suggestion to prevent rater self-forking is worth filing.
2. **Certification heterogeneity.** post-write and initiative were certified 2026-09-15
   under the hybrid regime, not the canonical battery; footnote per PROTOCOL. Canonical
   recert is a filed suggestion, not a blocker.
3. **Ceiling classes.** casg-delegate and initiative ceiling on their target question
   (no room to measure a glyph effect); conditional-gate's domain-free glyph is inert
   (only the grounded arm moves it). All three are dropped from the mechanism typing but
   still inform the grounded contrast where they calibrate.
4. **discovery** shows only a small negative glyph effect on Mistral, reproduced by the
   scramble — structure-leaning and inconclusive, not the predicted comprehension.
5. **n=8 per cell (95% CI ≈ ±0.2).** Read cells and direction, not exact counts.

## Handoff to revise-capc-paper (chain 426, task 3496)

The measured corpus distribution replaces the "three disagreeing classes" basis for the
CaPC §3/§4 mechanism claim. The honest claim to carry into the paper:

> Across the calibrating ALPHABET corpus, the descriptive glyph's behavioural effect is
> carried by **comprehension in the majority of classes** (post-write, governed,
> parent-state, formal-step, structural), with one class (casg-direct) carried by mere
> prepended structure and the remainder uninformative at ceiling. The strong scramble
> raises, not lowers, the prevalence of comprehension. A descriptive domain ground
> converts correct action as well as a command (ground ≈ domain-directive) corpus-wide.

Mechanism claim: comprehension is the dominant load-bearing ingredient, not the sole one;
state it as class-dependent typing with comprehension as the plurality, and footnote
casg-direct as the structure exception and the certification heterogeneity.

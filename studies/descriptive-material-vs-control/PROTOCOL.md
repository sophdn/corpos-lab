# Protocol — descriptive material vs control (H1)

**Chain:** 544 `descriptive-material-vs-control`, task 2
`design-treatment-vs-control-study`.
**Date:** 2026-09-15.
**Reads on from:** `GROUND_STATE.md` (the gap), and
`../matched-content-experiment/FINDINGS_ground_extension.md` (the method this
mirrors).

## The question

On decision classes where a model **under-acts unguided** (gating classes, a
below-ceiling baseline), does descriptive decision-point material **lift** correct
action versus no material — and does **grounding** the material convert
recognition into action, as it does in the suppression regime?

H1 is the mirror of the ground-extension study. That study ran the aid-and-ground
conditions on *ceiling* classes (casg-direct, formal-step) and measured how the
aid **suppresses** execution into analysis. H1 runs the same condition structure
on *gating* classes and measures **lift** from the floor. Same instrument, same
scoring, same sampler; different regime.

## Why this is not Q1 or the ground extension re-run

Q1 (matched-content) and the ground extension both used casg-direct and
formal-step, whose baselines are at ceiling — so they could only measure
suppression, and Q1's gating cells were the incidental few. H1 uses three
**certified gating glyphs** whose baselines are below ceiling, adds several
scenarios per class (both priors used one scenario each), and tests the ground
lever in the lift direction. It reuses their conditions and scoring wholesale.

## Classes — the three certified gating glyphs

Each is a certified ALPHABET glyph (chain 546), chosen because its baseline
leaves measuring room. Approx. baseline correct-action from the item-12 probes
(re-measured as a calibration pre-check below):

| class | glyph provenance | approx. baseline correct-action | shape |
|---|---|---|---|
| `parent-state-check-bypass` | authorization | ~10% (Q1: 1/8 C on Qwen) | strong under-fire |
| `post-write-verification-absent` | authorization | ~0% (probe: 0/16 verify) | maximal under-fire |
| `initiative-task-preexistence-gate` | existence | ~50% (probe: 8/16 fire) | moderate under-fire |

The spread (0% / 10% / 50%) is deliberate — three different amounts of room to
lift.

## Conditions

The canonical operationalization is `register-shift-followup` §Conditions — H1
reuses it exactly, so the two studies are directly comparable. Two factors are
crossed while the propositional content is held fixed (a 2×2), plus a baseline:

- **baseline** — the scenario alone.
- **glyph** (descriptive, domain-free) — the certified three-axis glyph. It
  describes; it does not instruct; it names abstract roles, not a specific tool
  or file.
- **imperative** (directive, domain-free) — a rule stating the glyph's content as
  commands rather than description. The content-and-form control: glyph vs
  imperative isolates form at matched content.
- **ground** (descriptive, domain-specific) — a domain-specific statement of the
  decision for the concrete scenario, in descriptive form: it says what a correct
  outcome *is*, without naming a command and without pasting the finished
  artifact.
- **domain-directive** (directive, domain-specific) — the same domain-specific
  content as the ground, commanding the action. Ground and domain-directive
  differ only in **mood**.

A content-parity audit is committed with the materials: glyph and imperative
state the same facts; ground and domain-directive state the same facts and differ
only in mood; neither ground nor domain-directive hands over the finished
artifact.

Optional mechanism controls (take the glyph's slot; run where a
comprehension-versus-recognition question is live): a **scrambled** glyph
(three-axis shape kept, words shuffled within each axis — tests whether structure
alone carries the effect) and an **off-target** glyph (a coherent glyph for a
different class — tests whether recognition of the scenario match is in the loop).

## Models — treatment subjects

Local open-weight only, spanning size and family; Claude is never a treatment
subject (it judges). Thinking is pinned OFF where applicable.

- `Mistral-7B-Instruct-v0.3` (7B)
- `phi-4` (14B)
- `Qwen3.8-27B` (27B, thinking OFF — pinned and declared)

(`Qwen2.5-32B` is an optional fourth if a size-4 point is wanted.)

## Scenarios

**Three scenarios per class**, each a distinct concrete instantiation of the
class's universal decision in a different task domain (the gap `GROUND_STATE.md`
named: the priors used one scenario per class). Authored from the certified
glyph's universal class, project-agnostic, in a context separate from the subject
models. Predictions are written before the runs and never placed in any file a
subject or rater can see.

## Grid and n

5 conditions × 3 classes × 3 scenarios × 3 models × **n=8** (seeds 1–8) =
**1080 runs**. Read as cells, not counts (n=8 → 95% CI ≈ ±0.2); each
class×model×condition pools 3 scenarios × 8 = 24, matching the ground-extension's
per-cell n while adding the scenario robustness it lacked. Stage it: run one
class end-to-end and read it before launching the rest; smoke one cell before
batching. If wall-clock on the 27B demands it, prune to the conditions that
discriminate, but keep the two-by-two intact per class.

## Sampler (complete chain, per INQUIRY.md)

temperature 0.8 · top_k 0 · top_p 1.0 · min_p 0.05 · all penalties off ·
max_tokens 1024 · seeds 1–8. Path: raw `/completion`, minimal per-model instruct
wrapper, thinking off, no system prompt, no tools, with the no-tools notice
appended identically across all conditions. Every run records the served model +
build id + throughput, the effective sampler, the image digest, the substrate
probe, and the repo commit.

## Scoring

Reuse the ground-extension's scorer, which reached 99.4% blind agreement with no
human anchor — the execute-versus-analyze distinction is objective. Two
independent blind raters (Claude subagents), blind to condition and to the
predictions, one execution bar across all conditions. Codes:

- **C** — recognition + correct-target action (the H1 lift measure).
- **Ic** — recognition, wrong-target action.
- **Ii** — recognition, no action (analysis-mode).
- **I** — no recognition.
- **N** — not scoreable.

Primary measure: **strict correct-target C rate** (does the model do the *right*
thing), per the correct-target discipline. Secondary: execution rate (C+Ic).
Use a deterministic parser wherever the target action is mechanically detectable
in the trace; the blind raters cover the judgment cases. The existing
human-verification template is the adjudication fallback for rater disagreements
only — not the primary scorer.

## Calibration pre-check (the gate before the full grid)

Before the 1080-run grid, run **the baseline only** for every class × scenario × model and
confirm baseline correct-target C is **below ceiling** — the measurement is
meaningless where the model already acts. Expected: all three gating classes
calibrate on the weak shelf; the 27B may ceiling on some class×scenario cells.
Record every ceiling cell and drop it from the lift analysis (it cannot measure a
lift), exactly as Q1 recorded its ceiling cells. A class with no calibrating cell
on any model is dropped.

## Predictions (pre-registered; kept out of every subject- and rater-visible file)

- **baseline**: low correct-target C — the under-firing baselines above.
- **glyph** (desc, free): lifts recognition, but correct-target C lifts only
  partially, and on the smaller models the glyph's does-not-fire clause inflates
  Ii (recognition-without-action) — the analysis pull the priors found, now from
  a floor instead of a ceiling.
- **imperative** (dir, free): tracks the glyph — content, not form, carries it,
  the settled finding — so also only a partial lift at the domain-free level.
  glyph ≈ imperative.
- **ground** (desc, specific): correct-target C lifts substantially toward
  ceiling — grounding is the execution lever, converting the floor case as it
  recovered the ceiling case. The 7B on the hardest class may still lag (the
  ground-extension's casg-Mistral exception).
- **domain-directive** (dir, specific): also lifts substantially. The open
  question `register-shift-followup` left under-determined is whether the
  descriptive **ground** suffices as well as the commanding **domain-directive**:
  does grounded comprehension alone produce the action, or does command do extra
  work? H1 tests it in the lift regime. Prediction: ground ≈ domain-directive on
  the larger models; domain-directive > ground on the 7B for the hardest class. A
  ground ≈ domain-directive result is direct support for comprehension-as-
  compliance from the floor.
- **Across classes**: the lift is largest where the baseline is lowest
  (`post-write-verification-absent`, ~0%), moderate on
  `initiative-task-preexistence-gate` (~50%).

## Invariants and hygiene

- Claude is never a treatment subject (raters/judges only).
- One local inference portal (llama.cpp); swap the model with `swap-model.sh`,
  never a second server; restore the default after.
- The container runs the image, not the working tree — rebuild and re-pin the
  assay image if `internal/assay` or `internal/runner` changed, then run.
- Observe, don't assert: a run that cannot fully describe itself is still a run.
- If the instrument is broken, fix it now; do not file-and-run.

## Materials to author before the run (task 3 inputs)

- 3 scenarios × 3 classes = 9 scenario stems (project-agnostic).
- Per class: the **glyph** = the certified AC-4 block; the **imperative** = a
  domain-free directive carrying the glyph's content, content-matched to the glyph.
- Per class × scenario: the **ground** = a domain-specific descriptive statement
  of the correct outcome, non-copy (does not write the artifact); the
  **domain-directive** = the same domain-specific content commanding the action,
  content-matched to the ground and differing only in mood.
- A committed content-parity audit (glyph = imperative; ground = domain-directive;
  neither ground nor domain-directive hands over the artifact).
- (Optional) a scrambled and an off-target glyph per class for the mechanism
  controls.

---
type: reference
last_updated: 2026-07-14
---

# casg-direct — grounded glyph probe

Three observations of one glyph (`casg-direct`) under the grounded-glyph probe:
does a domain ground close the execution gap that a universal glyph alone leaves
open?

> ## The parity premise is VOID (2026-07-14)
>
> This directory was `casg-direct-v3-repro` and existed to **reproduce** an April
> 2026 result — grounded 7/8 — as a gate the corpos-lab rig had to pass before any
> science was allowed to run on it. **That premise is dead.** The runs are not.
> Read them as three observations of a phenomenon, never as attempts to hit a
> number.
>
> **Why it's void — four reasons, any one sufficient:**
>
> 1. **The target was unreproducible in principle.** The April v3 study recorded no
>    temperature, no sampler, and no processor. Its conditions are gone and cannot
>    be recovered. "Reproduce v3" was never a well-formed request. The only reason
>    it *looked* answerable is that a retired Ollama daemon was still installed to
>    re-inherit the same accidental defaults — re-enacting an accident, not
>    reproducing a result.
> 2. **The target was noise.** Grounded 7/8 at n=8 has a 95% CI of [0.53, 0.98] —
>    half the scale. The "divergent" 4/8 is Fisher p=0.28 against it: not
>    distinguishable. A DIVERGED verdict was declared on a difference that cannot
>    be detected.
> 3. **Parity punishes learning.** A parity target is fixed from a prior state of
>    our own process, so any improvement — a new decomp step, a tightened battery
>    item — registers as divergence when it is simply truer. The 2026-04-03
>    demotion is the model of correct behaviour: the battery tightened, prior
>    passes no longer met it, ALPHABET was emptied and said so. Parity logic reads
>    that as a regression from eight glyphs to zero. **We are the frontier; there
>    is nothing to have parity with.**
> 4. **It was redundant.** Q2 (delivery register) asks the same question with more
>    models and more n. You validate an instrument by using it on something you
>    understand and watching whether it behaves sensibly — not by a gate before the
>    work.
>
> **Deleted, replaced by this notice:** `MANIFEST.v1/v2.sha256` (hand-written
> documents nothing read — `manifest.Verify` was never called on the run path),
> `CHANGELOG.md` (study-version-bump ceremony), `PARITY_VERDICT.md` ("REPRODUCED"
> against a target that never existed), `CONTAINER_PARITY_VERDICT.md` ("DIVERGED"
> — llama.cpp-on-GPU measured against Ollama-on-CPU, polarity inverted). One
> tombstone, not five void headers.
>
> The toolkit chains `instrument-parity-reproduction` and `instrument-hold-lift`
> closed with conclusions that are void by the above. They stay closed — they are
> history, and this notice is the correction. Their code output survives on merit
> (below). Context: [`INQUIRY.md`](../../INQUIRY.md).

---

## What these runs actually show

| leg | substrate | baseline (correct-target) | glyph_only | grounded | lift |
|---|---|---|---|---|---|
| April v3 (`PRIOR_GRID_2026-04-06.md`) | ollama, unrecorded | 0/8 | **0/8** | 7/8 | +7 |
| `runs/2026-07-13_ollama_mistral-7b_CPU` | ollama, **CPU**, ~5 tok/s | 0/8 | **0/8** | 8/8 | +8 |
| `runs/2026-07-14_llamacpp_mistral-7b_GPU` | llama.cpp, **GPU** | 0/8 | **0/8** | 4/8 | +4 |

**The signal: glyph-only is 0/24.** Not one execution in twenty-four runs, across
two runtimes, two samplers, and two processors — with the mechanism legible in the
text every time: axis-by-axis reasoning, Marker/Aim/Rest walked explicitly, "the
obligation remains unmet," then nothing produced. Nobody designed the substrate
variation; it happened by accident, which is exactly what makes it worth something.

**Scope limit, stated plainly:** one glyph, one scenario, one model. This does
**not** yet distinguish *"models read glyphs as analytical rubrics"* (the Q2 claim)
from *"casg-direct doesn't afford execution"* (a defect in this candidate). That is
the sharpest open question in Q2, and the battery is what should answer it —
`casg-direct` is a **suspended candidate**, not a promoted glyph.

**The noise: the grounded count.** +7, +8, +4. Always positive, always large, never
absent — the ground converts analysis-mode to execution. The specific integer is
unstable at n=8 and carries no information. Read cells, not counts.

**What the substrate accident taught:** the run directories are now named for what
they ran on. Temperature, sampler and processor were the three variables that
actually moved, and all three were invisible to a freeze that digested six files
which never change.

## Layout

| file | what |
|---|---|
| `study.toml` | the probe definition — materials, conditions, sampling |
| `materials/` | scenario, glyph, ground — byte-verbatim from the April v3 source |
| `SCORING_RUBRIC.md` | C/Ii/Ic/I/N, condition-specific. Real methodology; survives. |
| `PRIOR_GRID_2026-04-06.md` | the April observation. **Orientation, not a target.** |
| `runs/` | three legs of raw responses + scored grids. Immutable — they happened. |

## What survived from the voided work, on merit

Two code fixes stand independently of the premise that produced them:

- **The probe no longer fabricates failures.** It was scoring behavioural prose
  with the battery's PASS/FAIL prefix parser, so every run in every condition
  returned a spurious fail carrying the whole response as its reason — and shipped
  them to the canonical DB. It now carries its own rubric type and emits `Unscored`
  only. A judge scores; the probe captures.
- **Sampling is declared in the study, not hardcoded in Go.** Not because a freeze
  demanded it, but because a hardcoded sampler is one nobody can audit afterwards.

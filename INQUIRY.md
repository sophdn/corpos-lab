---
type: reference
last_updated: 2026-07-14
---

# INQUIRY — the glyph research program

**What this is:** the front matter of a lab notebook. What we're asking, what we
currently think, how we measure, and what changed our minds. It is **living** —
every line is revisable the moment we learn better, and the git log is the
history.

**What it is not:** a charter. It replaces `lab-app/resumption/CHARTER.md`
(2026-07-08 – 2026-07-14), which imposed a clinical-trial regime — pre-registered
hypotheses, sealed predictions, stopping criteria, freeze-by-digest, verdict
rules declared in advance — on a solo lab where a run costs four minutes. That
regime was retired 2026-07-14. Why is in *What changed the method*, below; the
short version is that it forced us to proceed with known problems and made
improving the process register as failure.

**There are no gates in this document.** Nothing here says you may not proceed
until X. If you find something broken, fix it and write down what you fixed.

---

## The questions

Three, in the order they matter. They are the same three the charter asked —
the questions were always the good part.

### Q1 — Is it the format, or just the content?

Does the three-axis form (Marker / Aim / Rest, written from inside the decision)
outperform an imperative rule carrying the *same information*? Or is the
comprehension-as-compliance effect just content arriving by any route?

This is the one that decides whether there's a thesis. **A null here is a real
answer, not a failure** — "content, not format, carries it" is a finding, and it
would make the delivery-register work (Q2) the story instead. We say that now so
that nobody has to be brave about it later.

**Currently:** untested. Needs a populated ALPHABET to draw glyphs from.

### Q2 — Is analysis-mode a register phenomenon?

Under third-person prepend, Mistral reads a glyph as an *analytical rubric* and
reasons about the decision instead of navigating it — recognition without
execution. Adding a domain ground with an instruction-shaped Action field
converts it to execution. Does that hold across the local shelf, and does
agent-facing voice convert it without a ground?

**Currently:** the strongest thing we have. See below.

### Q3 — Does the Rest axis pay for itself?

Rest names the territory where the decision class *doesn't* apply. Does it
measurably reduce false-positive overhead when a glyph is loaded but not live —
and what does the scaffold cost on neutral ground?

**Currently:** untested. No surveyed benchmark measures loaded behavioural
guidance over-firing in neutral territory; the axis to ablate doesn't exist in
other formats.

---

## What we currently believe

Living section. Update it when the evidence moves.

**Analysis-mode is real and robust (Q2).** `casg-direct` glyph-only produced
**0 executions in 24 runs** across three legs — the April v3 grid, an
ollama-on-CPU run, and a llama.cpp-on-GPU run. Two runtimes, two samplers, two
processors, one scenario, one model. Not one entry written, with the mechanism
legible in the text every time: axis-by-axis reasoning, Marker/Aim/Rest walked
explicitly, "the obligation remains unmet," then nothing produced.

Nobody designed that as a robustness test — the substrate varied by accident,
which is what makes it worth something. **Scope: one glyph, one scenario, one
model.** It does not yet distinguish "models read glyphs as rubrics" from "this
particular glyph doesn't afford execution." Q2's shelf sweep is what separates
those.

**Ground converts it.** Lift was +7, +8, +4 across the same three legs. Always
positive, always large. The magnitude is unstable at n=8 and we should stop
reading meaning into the specific integer.

**We do not know what April measured.** The v3 study recorded no temperature, no
sampler, and no processor. Its conditions are unrecoverable. The April runs are
a useful sign we're on the right track and **nothing more** — not a target, not
a baseline, not something to reproduce.

---

## How we measure

The parts that are real, as distinct from the ceremony we deleted.

**Calibration.** If the no-scaffold baseline doesn't exhibit the target failure,
the substrate cannot measure a scaffold's effect on it. This isn't a rule we
imposed — it's what the measurement means. Precedent: the FSCB v1 telegraphing
retirement.

**Correct-target scoring.** Baseline "success" that hits the wrong file or
format isn't success. casg-direct baselines score ~6-7/8 on any-attempt and
**0/8** on correct-target; the second number is the one that carries information.

**Contaminated subjects.** Claude-family models have this corpus's terrain in
training and ceiling at baseline. They may judge; they are never a treatment
condition. This is methodology, not ceremony, and it stays.

**Two substrates.** Single substrates structurally collapse for a given model —
the FSCB series proved it. One scenario is a pilot, not a result.

**Record what ran, automatically.** After every run, read the effective config
back from the server (`/props`: temperature, top_p, min_p, repeat_penalty, seed,
context), plus the model digest, the image digest, the substrate, and the git
commit. Store it with the results. **This replaces freeze-by-digest**, and it is
strictly stronger: the freeze pinned six files that never changed and was blind
to all three variables that actually moved. Observation beats assertion.

**8 runs per cell** is the working default, matching every prior grid. It is a
convention, not a threshold. n=8 gives a 95% CI roughly ±0.2 wide — wide enough
that single-integer comparisons between grids are noise. Read cells, not counts.

**Double-score.** Primary scorer plus a local second rater on a sample spanning
the code range. Report disagreements and say whether the reading survives them.

**Predictions, kept honestly.** Write down what you expect before you look — not
as a commitment, but because surprise is where the learning is. Drop the
confidence scores; they were theatre. The one real constraint: predictions never
enter any file a subject or judge model can see.

---

## What changed the method

The log. Append; don't rewrite.

**2026-07-14 — the charter regime retired.** It was a clinical-trial protocol
imported into a lab. Three concrete harms, all observed rather than theorised:

1. *It forced proceeding with known problems.* The aha-drain rule said mid-run
   insights route to the suggestion box and may only be applied at a version
   bump. So when we found the probe pinned temperature 0.0 — meaning a cell
   could only ever score 0/8 or 8/8, and a graded grid was structurally
   impossible — the rule said file it and run anyway.
2. *It made improving the process look like failure.* Parity fixes a target from
   a prior state of our own process. The 2026-04-03 demotion is the model of
   healthy behaviour: the battery tightened, prior passes no longer met it, so
   ALPHABET was emptied and said so. Under parity logic that reads as a
   regression from eight glyphs to zero. **We are the frontier; there is nothing
   to have parity with.**
3. *It had authority without verification.* `manifest.Verify` was never called on
   the run path. `internal/provenance` was dead code, so the invariant "every
   result row carries the repo commit" was simply false. The MANIFEST files were
   hand-written documents nothing read. It was prose enforced by compliance —
   which meant it produced the *feeling* of rigor while checking nothing, and
   the ceremony made a meaningless comparison read as a careful one.

**2026-07-14 — the parity gate deleted.** "Reproduce casg-direct v3's 7/8 before
you may do science" was unanswerable (conditions unrecorded), aimed at noise (a
point estimate at n=8), targeted a suspended candidate, and was redundant with
Q2, which asks the same question with more models and more n. You validate an
instrument by using it on something you understand and seeing whether it behaves
sensibly — not by a gate before the work.

**2026-07-14 — Ollama removed.** One local inference portal: llama.cpp
(`llama-server` :8081). A retired-but-still-installed daemon absorbed a broken
llama.cpp path (the Mistral GGUF had gone missing) and hid it for months. See
memory `one-local-inference-portal-llama-cpp`.

---

## Open

- ALPHABET is **empty** (all 8 candidates suspended 2026-04-03 when the battery
  tightened — honest bookkeeping, not damage). Q1 and Q3 need a populated corpus;
  the decomposition campaign is upstream of the science, not beside it.
- Q2's scope limit: one glyph, one scenario, one model. The shelf sweep is what
  turns it into a claim.
- `casg-direct-v3-repro` is void — see its tombstone.
- We still have no way to tell "models read glyphs as rubrics" from "this glyph
  doesn't afford execution." That's the sharpest open question in Q2 and the
  battery is what should answer it.

# corpos-lab

A small laboratory for running controlled behavioral experiments on local open-weight
language models.

AI agents fail in consistent, predictable ways — not randomly, but at specific kinds of
decision points where the locally obvious move is structurally wrong. This lab measures
whether describing those decision points to a model (a "map entry" for the decision terrain:
what going wrong looks like from inside, what correct navigation looks like, and what neutral
territory looks like) changes behavior — and whether it's the *description format* that does
the work, or just the information it carries.

corpos-lab is the instrument, not the findings: it runs studies against local models
(llama.cpp), scores the results against declared rubrics, and records what each run actually
executed under.

## What makes it a lab rather than a script pile

- **Runs that describe themselves.** A study declares its *complete* sampler chain —
  validation refuses a partial one, because temperature alone does not define a sampling
  distribution and any unnamed stage inherits whatever the server binary defaults to. Each
  run then records what was observed rather than what was intended: the model the server
  says answered, its build id, per-row generation throughput, the server's `/props`
  self-report, the processor, and the repo commit.
- **Divergence is recorded, never enforced.** If the server serves a model other than the one
  declared, that is written down and the run proceeds. A mismatch is information about the
  run, not grounds for refusing it.
- **Disposable containers:** behavioral assays run in single-use containers referenced by
  image digest, so a run's environment is exactly reproducible.
- **A hard quality gate:** format, vet, lint, vulnerability scan, race-tested tests, and a
  95% coverage floor run on every commit.

**Observation beats assertion**, and that is a correction rather than a slogan. This lab
previously ran on a *charter*: pre-registered claims, sealed predictions, stopping criteria,
and instrument-freezing by digest. It was retired on 2026-07-14 because it had a mechanism's
authority with none of a mechanism's checking — the verify step was never called on the run
path, the provenance package was dead code, and the "every result row records the repo
commit" line that used to sit in this README was simply false. Meanwhile the variables that
actually moved between runs — temperature, the sampler, and the processor — were recorded
nowhere, so a study certified a CPU-vs-GPU comparison as a reproduction and nothing in the
data could show it. See [INQUIRY.md](INQUIRY.md).

## Status

Early scaffold. The battery runner and container assay model are being ported from two
earlier repos.

## Guides

- **[docs/GLYPH_MINING_GUIDE.md](docs/GLYPH_MINING_GUIDE.md)** — how to mine your own
  agent's transcripts and logs for candidate glyphs: what a glyph is, the fit-the-definition
  triage, the decomposition/strip test, and the 15-item battery as a checklist. Start here to
  turn observed agent failures into corpus candidates.
- **[docs/LAB_CONTROLLER.md](docs/LAB_CONTROLLER.md)** — running a study end-to-end
  (`corpos-lab run-study`).
- **[docs/CONTAINER_ASSAY_MODEL.md](docs/CONTAINER_ASSAY_MODEL.md)** — the disposable assay
  container model and digest pinning.

## Running

Requires Go 1.26+, a local llama.cpp server, and (for persistence) the companion toolkit
service. Study definitions, batteries, and the demo profile land as the port progresses.

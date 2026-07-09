# corpos-lab

A small laboratory for running controlled behavioral experiments on local open-weight
language models.

AI agents fail in consistent, predictable ways — not randomly, but at specific kinds of
decision points where the locally obvious move is structurally wrong. This lab measures
whether describing those decision points to a model (a "map entry" for the decision terrain:
what going wrong looks like from inside, what correct navigation looks like, and what neutral
territory looks like) changes behavior — and whether it's the *description format* that does
the work, or just the information it carries.

corpos-lab is the instrument, not the findings: it runs pre-registered studies against local
models (llama.cpp), scores the results against declared rubrics, and records everything with
full provenance.

## What makes it a lab rather than a script pile

- **Pre-registration:** claims, conditions, batch sizes, and stopping criteria are declared
  before runs; the charter is versioned and frozen.
- **Instrument freezing:** every file a subject model can see is pinned by SHA-256 digest at
  study registration and re-checked at scoring time. Changing anything — one line of a
  scenario — requires a version bump. Results that can't reproduce their manifest don't count.
- **Provenance-stamped results:** each result row records the repo commit, instrument
  manifest digest, model artifact digest, and sampling parameters.
- **Disposable containers:** behavioral assays run in single-use containers referenced by
  image digest, so a run's environment is exactly reproducible.
- **A hard quality gate:** format, vet, lint, vulnerability scan, race-tested tests, and a
  95% coverage floor run on every commit.

## Status

Early scaffold. The battery runner and container assay model are being ported from two
earlier repos; studies begin after the instrument reproduces a known prior result
(parity gate).

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

# corpos-lab

A small laboratory for running controlled behavioral experiments on local
open-weight language models.

AI agents fail in consistent ways. Not at random, but at particular decision
points where the move that looks right from inside the moment is the wrong one.
This lab studies a simple question about those failures: if you describe the
decision point to a model before it acts, does its behavior change? And if it
does, is it the *description* that carries the effect, or just the information
the description happens to contain?

corpos-lab is the instrument, not the findings. It runs experiments against a
local llama.cpp server, scores the results against written rubrics, and records
exactly what each run executed under. The scientific claims live in `studies/`
and `papers/`; the code here is what produces and measures them.

## The idea

The unit of study is a **glyph**: a short, structured description of one
recurring decision point, written from inside the decision rather than as a rule
about it. A glyph names three things.

- **Marker** — what it looks like from the inside when you are getting this
  decision wrong.
- **Aim** — what navigating it correctly looks like.
- **Rest** — the neighboring territory where this decision class does not apply.

The lab shows a model a glyph and measures whether its behavior on a matched
scenario changes. A control condition carries the same information as a plain
instruction, so a difference between the two isolates the effect of the form.

## What the experiments have found so far

These are current readings, not settled conclusions. The living version, with
the evidence and the reversals, is in [INQUIRY.md](INQUIRY.md).

- **Format or content?** A preliminary null. When a glyph and a plain
  instruction carry the same information, the glyph does not win. The effect
  travels on the content, not on the three-part form.
- **A register effect.** This is the strongest result. Prepending a decision
  frame makes a local model *analyze* the decision instead of acting on it. It
  reasons about the terrain and produces nothing. The effect holds across four
  open-weight models, so it is general rather than a quirk of one. Adding a
  concrete domain ground converts the model back to acting.
- **Does the Rest axis earn its place?** Untested. No existing benchmark
  measures loaded guidance over-firing on neutral ground, so this one is open.

## How it treats a run

The lab was built after an earlier version of itself pinned six files that never
changed while leaving the three variables that actually moved between runs —
temperature, the sampler, and the processor — recorded nowhere. So the standing
rule is **record what ran, do not assert what should have**.

- **A run describes itself.** Every study declares its complete sampler chain;
  validation refuses a partial one, because temperature alone does not define a
  sampling distribution. After a run, the lab reads the effective settings back
  from the server, along with the model, the container image, the processor, and
  the repository commit, and stores them next to the results.
- **A mismatch is data, not an error.** If the server serves a different model
  than the study declared, the run writes that down and proceeds. Nothing about
  a run may fail it for failing to describe itself perfectly.
- **Assays run in disposable containers** referenced by image digest, so a run's
  environment is recoverable.
- **A hard gate on every commit:** format, vet, lint, a vulnerability scan,
  race-tested tests, and a 95% coverage floor on the engine.

## Reproducibility

Every treatment model is an open-weight model served over one bare llama.cpp
rig, so every result reproduces on a single 24 GB GPU with no paid API. A hosted
model cannot run that way, so hosted models are never a treatment condition —
they may score results, never produce them. This is the reason the whole
apparatus fits on one desktop.

## Studies and papers

`studies/` holds the experiment records: each study's definition, its materials,
its rubric, and its findings, with the raw model responses committed alongside.
`papers/` holds the write-ups. Several are posted as preprints on Zenodo and are
linked from ORCID [0009-0001-2430-1743](https://orcid.org/0009-0001-2430-1743).

## Running it

See **[GETTING-STARTED.md](GETTING-STARTED.md)**. There are two ways in.

- **See the instrument with no model.** `corpos-lab glyph-lint` runs the
  structural checks on a candidate glyph and needs nothing but Go. It is one
  command and a few seconds.
- **Run a real measurement.** This needs a local llama.cpp server and a model
  file on disk. The getting-started guide walks reproducing one of the committed
  studies on a 7B model.

Requires Go 1.26 or newer. A real run also needs a local llama.cpp server; the
companion toolkit service is optional and off by default.

## License

This repository holds both software and research material, licensed separately.

- **Code** — the Go sources, scripts, and tooling — is under the GNU Affero
  General Public License v3.0 (AGPL-3.0). See [LICENSE](LICENSE).
- **Papers and datasets** — the `papers/` and `studies/` directories — are under
  the Creative Commons Attribution 4.0 International License (CC-BY-4.0). See
  [LICENSE-docs-CC-BY-4.0.txt](LICENSE-docs-CC-BY-4.0.txt).

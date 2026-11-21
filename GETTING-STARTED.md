# Getting started (for an agent)

This file is written for an AI coding agent. If someone pointed you here, your
job is to get corpos-lab built and show them the instrument working. Follow the
steps in order. Each step says what to run and what a good result looks like.
Stop and report if a check fails.

There are two levels. Level 1 needs nothing but Go and shows the instrument
running. Level 2 measures a real model and needs a local llama.cpp server. Do
Level 1 first; it always works.

## 0. What this is

corpos-lab runs behavioral experiments on local language models. It studies
whether describing a recurring decision point to a model (a "glyph") changes how
the model behaves. Read the [README](README.md) for the idea; this file is about
running it.

## 1. Check the toolchain

You need Go 1.26 or newer. Run:

```sh
go version
```

If Go is missing or older, tell the user and stop.

## 2. Get the code and build

```sh
git clone https://github.com/sophdn/corpos-lab.git
cd corpos-lab
go build ./...
```

A good result is no output and exit code 0. If you were pointed at a local copy,
`cd` into it instead of cloning.

## 3. Run the tests

```sh
go test ./...
```

A good result is `ok` for each package. The tests are hermetic and need no
model and no network.

---

## Level 1 — see the instrument (no model needed)

The lab checks candidate glyphs against structural rules before any model is
involved. You can run those checks on the committed test fixtures. This needs no
model, no container, and no network.

Build the CLI once:

```sh
go build -o corpos-lab ./cmd/corpos-lab
```

Lint a candidate that is meant to pass:

```sh
./corpos-lab glyph-lint internal/battery/testdata/known_pass.md
```

A good result ends with `PASS (structural items)` and exit code 0, with each
structural item listed as `pass`.

Now lint one that is meant to fail:

```sh
./corpos-lab glyph-lint internal/battery/testdata/known_fail_item2.md
```

A good result is a `FAIL` with a named reason — the firing condition uses
intent-modeling language ("when the agent decides") that the rules forbid — and
exit code 1. Seeing the instrument catch the exact flaw is the point: it
discriminates a good candidate from a bad one on a written rule, not a hunch.

Report both outputs. That is the instrument working end to end with nothing but
Go.

## Level 2 — measure a real model

This reproduces one of the committed studies against a small local model. It
needs one thing you must supply: a running **llama.cpp** server on port 8081,
serving a 7B model file. The lab uses llama.cpp as its only inference portal by
design, so every result reproduces on one desktop GPU with no paid API. Do not
substitute another runtime.

### First, look at a real result without running anything

The study `studies/casg-direct-grounded-probe/` ships its materials, its rubric,
and a committed run record with the raw model responses. Read
`studies/casg-direct-grounded-probe/README.md` and open the `runs/` directory to
see exactly what a measurement produces before you run one.

### Then run it yourself

1. **Serve a 7B model.** Start a llama.cpp `llama-server` on `:8081` with a 7B
   GGUF (the committed study uses Mistral-7B-Instruct-v0.3). This is the one
   dependency you provide; the lab does not download weights.

2. **Run the study.** The full run path uses a disposable container, so it needs
   rootless podman and a built probe image:

   ```sh
   scripts/build-lab-images.sh
   corpos-lab run-study studies/casg-direct-grounded-probe/study.toml -toolkit-url ""
   ```

   `-toolkit-url ""` turns off the optional persistence service, so the run needs
   no companion backend. Results land as JSON and raw response files under the
   study's `runs/` directory.

   The full walkthrough, including how the container reaches your model, is in
   [docs/LAB_CONTROLLER.md](docs/LAB_CONTROLLER.md). A copy-ready study template
   is at [deploy/examples/study.grounded-glyph-probe.toml](deploy/examples/study.grounded-glyph-probe.toml).

3. **Read the cells, not the counts.** Eight runs per condition is a small
   sample; what reproduces is the direction of the effect, not an exact score.
   [INQUIRY.md](INQUIRY.md) explains how to read a grid.

Report the results directory and what the conditions produced. If the server was
unreachable, the run writes a failed record and exits non-zero — report that
verbatim.

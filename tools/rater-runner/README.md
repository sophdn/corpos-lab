# Isolated rater runner

`rate.py` runs blind raters over a grid's response slices without the failure that
prompt-enforced isolation could not prevent: during the alphabet-assay scoring,
blind-rater subagents self-forked and shared a temp namespace, so concurrent
raters collided and a 28-minute straggler overwrote a clean result
(suggestion `isolated-deterministic-rater-runner-for-lab-grid-scoring`).

The runner makes the four RATER_INSTRUCTIONS.md rules structural instead of
prompt-enforced:

1. **No nested subagents.** A rater is an ordinary subprocess (an argv), so it
   cannot spawn its own forks. Concurrency is a fixed worker pool of explicit
   size (`--jobs`, default 1), never recursive fan-out.
2. **Isolated scratch.** Each rater invocation runs in its own fresh temp
   directory, with `TMPDIR` and the working directory pointed there. No two raters
   share a scratch namespace.
3. **Un-clobberable output.** A rater writes only to a private path inside its own
   scratch. A validated result is promoted to the slice's canonical output with an
   atomic, create-only move (`O_CREAT|O_EXCL`), so once a slice has a fresh result
   a late straggler cannot overwrite it.
4. **Completion by content.** A slice is done when its canonical output exists and
   its JSON keys are exactly the slice's ids — not because a job process finished.
   A killed run resumes and re-runs only the slices without a valid result.

## Rater contract

A rater reads a **slice** — a JSONL file, one `{"id": "...", "text": "..."}` per
line — applies one rubric, and writes a single JSON object mapping every id to its
code. The runner passes the slice path and the output path into the rater command
through `{slice}` and `{out}` placeholders. Any rater that honors this contract
plugs in: the local phi-4 rater, the mechanical scorer, or a headless model rater.

## Usage

    rate.py --slices-dir slices/ --out-dir scores/ --rater-id phi4 \
        --rater-cmd 'python3 score_phi4.py --in {slice} --out {out}' [--jobs N] [--dry-run]

Results land at `<out-dir>/<rater-id>/<slice>.json`. An output whose keys do not
match the slice's ids is quarantined as `<slice>.invalid.json` and never becomes
the canonical result. `--dry-run` reports done vs pending and runs nothing.

## Determinism

With `--jobs 1` (the default) slices run in sorted order, one at a time — the fully
reproducible mode. A larger pool trades that order for speed; the per-slice results
are unaffected because each slice is independent and its output path is unique.
Keep the rater itself judge-free and reproducible (a mechanical scorer, or a model
rater at a pinned sampler); the runner does not add nondeterminism of its own.

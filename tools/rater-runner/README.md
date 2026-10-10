# Isolated rater runner

> **Ported to Go (2026-09-22).** The grounded rater (`score_local_chat.py`), the
> action rater (`score_action.py`), and this orchestrator (`rate.py`) were ported
> to the Go module as the `corpos-lab rate` subcommand and retired. Git history
> holds the Python. The Go runner keeps the same structural guarantees below, and
> its scores match the Python raters exactly — parity verified live on
> deepseek-flash (24/24, grounded and action). New scoring runs use `corpos-lab
> rate`; see Usage. The blind Claude pathway (`BLIND_CLAUDE_RATER.md`) is
> unchanged, and `anchor_eval.py` stays.

The rater runner runs blind raters over a grid's response slices without the failure that
prompt-enforced isolation could not prevent: during the alphabet-assay scoring,
blind-rater subagents self-forked and shared a temp namespace, so concurrent
raters collided and a 28-minute straggler overwrote a clean result
(suggestion `isolated-deterministic-rater-runner-for-lab-grid-scoring`).

The runner makes the four RATER_INSTRUCTIONS.md rules structural instead of
prompt-enforced:

1. **No nested subagents.** A model rater runs in-process, in a bounded worker
   pool, not as a dispatched agent, so it cannot spawn its own forks. The pool
   size is explicit (`--jobs`, default 1), never recursive fan-out.
2. **Isolated output.** In-process rating shares no scratch namespace. Each slice
   writes only its own unique canonical path, so no two raters collide on a file.
3. **Un-clobberable output.** A validated result is promoted to the slice's
   canonical output with an atomic, create-only write (`O_CREATE|O_EXCL`), so once
   a slice has a fresh result a late straggler cannot overwrite it.
4. **Completion by content.** A slice is done when its canonical output exists and
   its JSON keys are exactly the slice's ids — not because a job process finished.
   A killed run resumes and re-runs only the slices without a valid result.

## Rater contract

A rater reads a **slice** — a JSONL file, one `{"id": "...", "text": "..."}` per
line — applies one rubric, and writes a single JSON object mapping every id to its
code. The action rater reads one extra field per line, `scenario`, that selects the
concrete A_local / A_canon actions. Results land under `<out-dir>/<rater-id>/`.

## Usage

`corpos-lab rate` carries both raters: the grounded C/Ii/Ic/I/N codes (`--rubric`)
and the action A_local/A_canon verdict (`--action`).

    # Grounded rater, local llama-server:
    corpos-lab rate --slices-dir slices/ --out-dir scores/ --rater-id devstral \
        --rubric <class-rubric>.md [--jobs N]

    # A hosted rater family (a second, independent family for consensus):
    corpos-lab rate --slice s.jsonl --out-dir scores/ --rater-id deepseek-flash \
        --rubric <class-rubric>.md --base https://api.deepseek.com/v1 \
        --model deepseek-flash --api-key-env DEEPSEEK_API_KEY --prov

    # Action verdict:
    corpos-lab rate --slice s.jsonl --out-dir scores/ --rater-id ds --action \
        --base https://api.deepseek.com/v1 --model deepseek-flash \
        --api-key-env DEEPSEEK_API_KEY

Results land at `<out-dir>/<rater-id>/<slice>.json`. A hosted endpoint gets a
generous token budget by default — a reasoning model spends it on
`reasoning_content` and returns empty content at a small cap — and `--max-tokens`
overrides it. `--prov` writes a provenance sidecar next to each result. A grounded
reply with no code is coded `N`; an action reply with no verdict is `unscoreable`.

## Determinism

With `--jobs 1` (the default) slices run in sorted order, one at a time — the fully
reproducible mode. A larger pool trades that order for speed; the per-slice results
are unaffected because each slice is independent and its output path is unique.
Keep the rater itself judge-free and reproducible (a mechanical scorer, or a model
rater at a pinned sampler); the runner does not add nondeterminism of its own.

## Blind Claude raters — the required isolated pathway

The primary measure in these studies is two blind **Claude** raters. Claude raters
do not run as a subprocess under this runner, so they need the same isolation by a
different route: **the driving session mints a unique scratch dir per rater and
passes it in.** Every rater, Claude or local model, must use the isolated pathway —
it is required, not optional. This is the guard against the scoring race where two
raters share a scratch namespace and one clobbers the other's result.

- Local-model rater → run it under `corpos-lab rate` (in-process pool, create-only
  output), as above.
- Blind Claude rater → use the copy-paste stub in `BLIND_CLAUDE_RATER.md`. It
  assigns each rater a unique `mktemp -d` scratch dir, forbids fixed-name scratch
  dumps (the `all_items.txt` collision in chain 549), and md5-verifies the source
  slice. A study's `SCORING.md` points at this pathway; `SCORING_TEMPLATE.md` is the
  template.

Blindness and the held-key id-check stay the guard of record: the condition is held
out in `key.json`, ids are opaque content hashes, and every output is validated
against the slice ids.

**Strict two-rater consensus is the scoring standard, and single-rater
disjoint-halves scoring is retired.** The primary measure is strict-consensus C —
both raters code C. Splitting a study's classes one-rater-each (no response
double-scored) is what let a lenient-C singleton become the recorded code in
neutral-prefix-control; consensus filters those. The code definitions and the
tightened decision order both raters apply live in `RUBRIC_STANDARD.md`, carried
inline in every class rubric.

### Why not a bundled `claude -p` wrapper?

A headless Claude command does exist on this machine (`claude -p --output-format
json`, verified 2026-09-17). We chose **not** to bundle it as a `--rater-cmd`
wrapper. It runs the full Claude Code harness, not a bare scoring API: it costs an
API call per slice, is nondeterministic, and carries tool and MCP access that would
need sandboxing to keep a rater blind. It would also introduce a new, unvalidated
rater substrate right before a grid wave. The blind-subagent stub hardens the rater
substrate the studies already use, which is thinner and lower-risk. Revisit the
wrapper only if a mechanical, tool-free headless rater is wanted.

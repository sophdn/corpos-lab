# corpos-lab runbook

How to do the common lab jobs, and which files each one touches. Written so an
agent can act from this page instead of mapping the repo again. The rules behind
these steps are in [CLAUDE.md](../CLAUDE.md) and [INQUIRY.md](../INQUIRY.md);
this page is the how.

Build the CLI once per session from the repo root:

```sh
export PATH="$PATH:/usr/local/go/bin"
go build -o /tmp/corpos-lab ./cmd/corpos-lab
```

`corpos-lab` with no arguments prints every subcommand.

## Run a study

A study is one `study.toml` (one model, one item, one sampler chain). The run
executes inside the pinned probe image, never on the working tree.

```sh
corpos-lab run-study path/to/study.toml                 # full run, persisted to the toolkit
corpos-lab run-study path/to/study.toml -runs 1 -toolkit-url ""   # smoke one cell, no ledger row
corpos-lab run-study path/to/study.toml -seeds 2,3,4    # just these seeds
corpos-lab run-study path/to/study.toml -image localhost/lab-agentic-loop-probe@sha256:…
```

- Output goes to `<study dir>/runs/<name>/` (or `-work DIR`). The run prints the
  absolute `results.json` path when it ends.
- `-seeds`, `-runs` and `-image` change the run without a TOML copy (a copy in
  /tmp breaks relative material paths). The run record's `overrides` field names
  each one.
- A run with any override is a smoke or a diagnostic and is **not** written to
  the toolkit ledger; pass `-persist` to record it anyway. A run without
  overrides is persisted; `-toolkit-url ""` turns that off too.
- Cells run one at a time on purpose. Parallel slots changed per-seed
  transcripts and the throughput tripwire, and ran no faster (see
  `forEachCell` in `internal/runner/runner.go`).
- Smoke one cell and read it before a full grid (CLAUDE.md, Study discipline).

**Files:** `internal/study` (TOML, validation, overrides), `internal/control`
(the host-side run and the run record), `internal/runner` (the in-container
executor), `cmd/corpos-lab/main.go` (`runStudy`).

## Run a grid

A grid is many `study.toml` files, one per cell. Make them with a study-local
generator and sweep them grouped by model; the pattern is in
[studies/MATRIX_STUDIES.md](../studies/MATRIX_STUDIES.md).

```sh
scripts/run-grid.sh --defs <dir> --model <substr> --chunk <N>   # resumable; re-run to continue
scripts/run-grid.sh --defs <dir> --dry-run                      # completed vs remaining
```

To sweep a grid across every model on `deploy/shelf.toml` in one command, swapping
the portal per model and ending on the primary:

```sh
corpos-lab run-shelf <defs-dir> [-chunk N] [-dry-run]
```

It runs `run-grid.sh` for each role inside the llama-server repo's
`scripts/with-model.sh` (default checkout `~/dev/the model-server repo`), which
restores the previously served model even when a leg fails.

`run-grid.sh` skips cells whose run record says completed, and writes
`run-grid-progress.jsonl` and `run-grid-status.json` beside the defs, so a killed
leg is visible and resumes with no rework. Swap the model on llama-server between
legs; never start a second inference server.

## Read a run

```sh
corpos-lab cells <loop-run-dir>            # one line per loop cell
corpos-lab cells <loop-run-dir> -json
corpos-lab provenance <study-dir>          # builds, n_ctx, sampler, seeds, images per model
```

- `cells` shows, per cell: seed, terminal, turns, lost calls, edits that ran vs
  edits claimed, truncated turns, files opened in order, sandbox changes, and the
  final line, and REPEATS (identical calls within one turn that the loop answered
  with a note instead of running again). It reads the transcripts, so it works on
  old runs too. It makes no fire/not-fire call.
- A loop terminal is `final`, `stall`, `cap` (step cap), `loop` (repeated
  identical calls) or `context` (the next request exceeded n_ctx).
- `results.json` loop rows carry `parse_outcomes` per turn, `lost_calls` and
  `ignored_calls`. Many lost calls in a NULL cell point at the parser, not at the
  model.
- `provenance` is the source for a paper's provenance table.

## Score

- **Model raters:** `corpos-lab rate --slices-dir DIR --out-dir DIR --rater-id ID
  --rubric R.md` (or `--action` for the action rater). Contract and isolation
  rules: [tools/rater-runner/README.md](../tools/rater-runner/README.md).
- **Claude raters:** the blind pathway in
  [tools/rater-runner/BLIND_CLAUDE_RATER.md](../tools/rater-runner/BLIND_CLAUDE_RATER.md).
- **Rubrics:** every probe-code rubric carries the block in
  [tools/rater-runner/RUBRIC_STANDARD.md](../tools/rater-runner/RUBRIC_STANDARD.md)
  verbatim and defines only its own correct target. The measure of record is
  consensus across independent rater families, never single-rater halves.
- **Study-specific scoring:** `corpos-lab pub-score <mode> …`. Run
  `corpos-lab pub-score` to list the modes.
- **Statistics:** `corpos-lab stats fisher A B C D | wilson K N | newcombe
  [-conf 95|90] K1 N1 K2 N2 | kappa A.json B.json`, not ad-hoc Python (the lab is
  Go-only). Newcombe is (K1/N1) − (K2/N2); kappa joins two rater output files on id.
  Sample-size planning is `corpos-lab plan`.

## Close a study

```sh
corpos-lab study-close studies/<name>
```

It reports the INQUIRY.md study-close facts (run records, mechanism controls,
second rater, library reconciliation, predictions). Those findings advise and
never block. One check fails the command: a study under `studies/` must have a
row in [studies/INDEX.md](../studies/INDEX.md).

## Rebuild and re-pin the probe images

After any change to `internal/assay`, `internal/runner`, `internal/agentloop` or
`cmd/lab-assay`:

```sh
scripts/build-lab-images.sh --repin [study TOMLs to move…]   # build, re-pin, check
scripts/build-lab-images.sh --repin-only [study TOMLs…]       # re-pin from the last build
scripts/build-lab-images.sh --check-pins                      # scaffold pins vs the build
```

`--repin` rewrites `LoopImg`/`RawImg` in `internal/setupcompletion/studies.go` and
the digest in each TOML you name, then runs `--check-pins` and the scaffold oracle
test. Name only studies still being authored: a study that has already run keeps
its pin, because the pin is what makes its result reproducible. Commit the
re-pinned `studies.go` with the change that needed it.

**Files:** `scripts/build-lab-images.sh`, `deploy/Containerfile.*`,
`deploy/IMAGE_DIGESTS.txt` (gitignored, written by the build),
`cmd/corpos-lab/repin.go`, `internal/image/repin.go`.

## Add a tool-loop study (specimen)

```sh
corpos-lab new-loop-study <dir> [-name N] [-item ID] [-query]
```

It writes `study.toml`, `preamble.md`, `scenario.md` and `sandbox.json` with the
current loop image, prompt template, sampler chain and `call_tokens`. Write the
task in `scenario.md` and the repository in `sandbox.json` (a JSON path→contents
map), then smoke one cell. Pass `-query` only when the sandbox ships the
`_db.txt` data that `run_query` reads; otherwise the preamble does not offer it.

## Add a scenario

A scenario is the task body file named by `[materials] scenario`. A canon
scenario carries `--- glyph: <slug> / role: fire|not-fire ---` frontmatter, which
the runner strips before the prompt, and must name none of the glyph meta-layer.
Check one with `corpos-lab canon-lint -scenario-file <path>`; list every enforced
rule with `corpos-lab canon-lint --rules`.

## Add an assay condition

A condition touches five places. Template commit: `ca17e48e` (the
`neutral_prefix` control).

1. `internal/assay/grounded.go`: the `Condition` constant, its `Materials` field,
   and its `AssemblePrompt` case.
2. `internal/study/study.go`: the `MaterialsDef` field, the `validate` case that
   requires its material, and the `Materialize` copy.
3. `internal/runner/runner.go`: the `MaterialsSpec` field and its read in
   `loadMaterials`.
4. Tests in `internal/assay/grounded_test.go` and `internal/study/study_test.go`.
5. Rebuild and re-pin the images (the condition runs in the container).

## Add a scorer

Study-specific scoring is a `pub-score` mode. Put the pure logic in
`internal/<study>/` with its tests (the 95% coverage floor applies), and the file
IO in `cmd/corpos-lab/pubscore_<study>.go`, which registers each mode with
`registerPubScoreMode` from `init()`. Template:
`cmd/corpos-lab/pubscore_neutralprefix.go` with `internal/neutralprefix`. A
scorer that finds nothing to score must return an error, never an empty success
(bug 1369, `cmd/corpos-lab/pubscore_silentempty_test.go`).

## Glyph battery

Glyph work was retired on 2026-10-09; the battery remains for re-reading old
results. `corpos-lab battery <candidate.md>` runs the 15-item battery and
`corpos-lab glyph-lint <candidate.md>` checks structure only.

# Lab Controller

The host-side controller that runs a study end-to-end — the piece the old
system never wired (`e9-lab-controller`). It takes a study definition,
materializes the container `/in`, pins the run by digest, launches the
disposable assay container, collects and validates `/out`, and writes a durable
run record. Reproducible headless: a study runs from its definition file alone,
no agent in the loop.

## Study definition (host-authored, TOML)

One file describes one reproducible run. Material paths resolve relative to the
definition file.

```toml
name = "casg-direct-v3-controller-smoke"
assay = "grounded-glyph-probe"
item_id = "casg-direct"
image = "localhost/lab-grounded-glyph-probe:dev"
network = "corpos-net"
conditions = ["baseline", "glyph_only", "grounded_glyph"]
runs_per_cell = 1

[model]
base_url = "http://llama-server:8081/v1"
model_id = "Qwen2.5-32B-Instruct-Q4_K_M.gguf"
version = "q4km"

[materials]
scenario = "scenario.md"
glyph = "glyph.md"
ground = "ground.md"
```

Validated on load: required fields, supported assay, `runs_per_cell >= 1`, and
per-condition material requirements (`glyph_only` needs a glyph;
`grounded_glyph` needs a glyph and a ground). An example ships at
`deploy/examples/study.grounded-glyph-probe.toml`.

## Running

```bash
corpos-lab run-study <def.toml> [-work DIR]
```

Exit 0 on a completed study, non-zero on any failure. `-work` defaults to
`<def-dir>/runs/<name>`. The controller writes `<work>/run-record.json` (the
durable artifact persist-and-dashboard ingests) plus the collected
`<work>/out/` (results.json, responses/, manifest.json).

Under the hood, per run:

1. **Materialize** — derive the container `/in` (study.json + copied materials)
   from the definition.
2. **Pin** — read the image content digest (podman) and compute the content
   manifest over image + study.json + every material. **Recorded, never
   enforced** — freeze-by-digest was retired 2026-07-14 and the verify arm
   deleted, so a digest that doesn't match a prior run is information about the
   two runs, not grounds for refusing to run. See INQUIRY.md §How we measure.
3. **Launch** — one disposable container:

   ```
   podman run --rm --userns=keep-id --user <hostuid>:<hostgid> \
       --network corpos-net -v <in>:/in:ro,Z -v <out>:/out:Z \
       localhost/lab-grounded-glyph-probe:dev run
   ```

   `--userns=keep-id --user <hostuid>:<hostgid>` runs the container as the host
   user so it can write the host-owned `/out` bind mount (the image's nonroot
   uid otherwise maps to an unprivileged subuid that cannot). Networking is
   explicit — join `corpos-net`, reach `llama-server:8081` from `study.json` —
   never `--network host`.
4. **Collect + validate** — read `/out/results.json`, check it parses, matches
   the study's assay/item, and carries exactly `conditions × runs_per_cell`
   rows.
5. **Record** — write `run-record.json` (status, digest pin, rows, extraction
   manifest) for every outcome.

## Failure semantics (never a silent skip)

Every failure produces a typed error **and** a persisted failed-run record.

| Failure | Typed error | Notes |
|---|---|---|
| Container non-zero exit | `ContainerExitError{ExitCode, Stderr}` | **Model unreachable** surfaces here — the assay binary exits 1 when the model call fails. |
| Absent / unparseable / wrong-identity / short-grid `results.json` | `MalformedResultsError{Path, Reason}` | A zero-exit container with fewer rows than `conditions × runs_per_cell` is malformed, not a partial success. |
| Image digest unreadable (e.g. image absent) | wrapped digest error | Fails before launch. |
| Materialize / launch (podman itself) failure | wrapped error | — |

In every case `run-record.json` is written with `status = "failed"` and the
error recorded; the CLI exits non-zero. A run that cannot even write its record
is itself a hard error — a run never vanishes silently.

## corpos integration decision

**Decision: corpos-lab is a standalone headless CLI (`corpos-lab run-study`)
that corpos drives as an external tool, via its command/tool surface.** The
controller owns podman orchestration and the study lifecycle; corpos invokes it
and reads the run record, the same way it drives any other external command.

**Rejected alternative: corpos orchestrating assay batches directly inside its
Go agent loop** (embedding podman-run + `/out` collection behind the
`tool.Provider` seam). Rejected because:

- It couples the agent-OS to lab- and podman-specifics that have nothing to do
  with the agent loop's job (driving a model through tools).
- It breaks the hard constraint that a study be **reproducible from the
  definition file alone, headless** — embedding the lifecycle in the loop makes
  a study depend on an agent being present to drive it. A CLI is reproducible
  by anyone (a cron job, a CI step, a human) with no model in the loop.
- The CLI loses corpos nothing: corpos can still kick off and monitor a study
  as a tool call, and the typed run record is a cleaner monitoring surface than
  in-loop state.

An agent may invoke `corpos-lab run-study` and poll the run record; it is never
required for a study to run.

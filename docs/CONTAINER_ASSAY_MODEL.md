# Container Assay Model

corpos-lab's disposable behavioral-assay containers, rebuilt on rootless
podman. Successor to registry-lab's `LAB_CONTAINER_MODEL.md`. The container is
the entire environment cost; nothing persists except through the volume
interface.

## Images

- **`lab-base:dev`** — the CGo-free static `lab-assay` binary on distroless
  (`gcr.io/distroless/static-debian12:nonroot`, digest-pinned). Owns the
  `/in`-`/out` contract, mode dispatch, and the extraction manifest.
- **`lab-grounded-glyph-probe:dev`** — `FROM lab-base`, sets
  `LAB_IMAGE=lab-grounded-glyph-probe`, defaults to `run`. The grounded-glyph
  probe (ported from `studies/grounded_glyph_probe.rs`) is the first real
  assay.
- Three further variants (behavioral-equivalence, structural-glyph-probe,
  decomposition) are stubbed in registry-lab and land here as their assay logic
  is ported — they are science-chain work, not this chain's. The build script
  globs `Containerfile.<variant>`, so adding one is zero-script-edit.

## Build

```bash
scripts/build-lab-images.sh            # base + variants + report-mode smoke + record digests
scripts/build-lab-images.sh --digests  # print current image digests
```

Images build reproducibly (bases pinned by digest, static `-trimpath` binary)
and their content digests are recorded to `deploy/IMAGE_DIGESTS.txt`. This is
**not** in the pre-commit gate — a distroless build is minutes and needs
podman; the Go gate stays fast and hermetic. Run it standalone / in CI.

## The `/in`-`/out` contract (preserved from registry-lab)

`/in` (read-only) and `/out` are the only state channels.

- **`/in/study.json`** — a self-contained typed spec: `assay`, `item_id`,
  `model` (`base_url` / `model_id` / `version`), `conditions`, `runs_per_cell`,
  `materials` (filenames relative to `/in`).
- **`/in/<materials>`** — scenario / glyph / ground text files.
- **`/out/results.json`** — real scored rows (`item`, `condition`, `run`, typed
  `verdict`, `rationale`). **Never a `pending-adapter` placeholder** — a
  regression test enforces this.
- **`/out/responses/<condition>_<run>.txt`** — raw model reply per cell, for
  audit.
- **`/out/manifest.json`** — the extraction manifest (image, mode, timestamps,
  exit code, results path), written on every exit including crashes.

### Entry modes

`lab-assay` is PID 1 and dispatches on its first argument:

- `report` (base default CMD) — print environment, exit 0. Smoke test.
- `run` (variant default CMD) — read `/in`, run the assay, write `/out`.

## Documented departures from registry-lab

The `/in`-`/out` *interface* is preserved; the *implementation* changed, with
rationale:

1. **Distroless static Go binary replaces Ubuntu + python + jq + bash.** The
   assay logic is now Go (ported from the Rust study runner), so the
   interpreter toolchain was dead weight. One static `lab-assay` binary
   implements report/run dispatch and writes the extraction manifest natively
   — smaller image, smaller attack surface, and it satisfies the corpos-family
   invariant (CGo-free static Go on distroless, bases pinned by digest).
2. **`shell` mode dropped.** Distroless has no shell. Debug with
   `podman run --entrypoint=... ` or a throwaway debug image; production assays
   never needed it.
3. **`tini` dropped.** The assay is a batch job that runs to completion; the Go
   binary writes the extraction manifest via a deferred call on every exit
   path, so the "manifest always written" property holds without a PID-1
   reaper. Pass `podman run --init` if a reaper is ever wanted.

## Digest pinning (freeze-by-digest)

Every study run is pinned to exact content digests of its executor-visible
inputs — the assay image and each material file — per the charter
freeze-by-digest rule. `internal/manifest.RunManifest` records image digest +
study.json digest + per-material digests + model identity.

```bash
lab pin    -in DIR -image REF [-out manifest.json]   # compute + write the pin
lab verify -in DIR -image REF -manifest manifest.json # hard refusal on mismatch
```

`verify` recomputes under the actually-observed image digest and refuses (exit
1) on **any** divergence — image, study.json, or a changed / added / removed
material — listing every one. A result whose manifest can't be reproduced is an
anecdote, not data. The lab controller (chain task
`corpos-as-lab-controller`) calls `verify` before every `podman run`.

## Networking decision

Assay containers reach llama-server by **explicit config, never implicit host
networking** (`--network host` is banned).

- **Topology (coordinated with chain 331 `corpos-podman-deploy`):** containers
  join the existing **`corpos-net`** podman bridge, where llama-server already
  runs as `llama-server` and toolkit-server as `toolkit-server`. The assay
  reaches the model at **`http://llama-server:8081/v1`**, supplied in
  `study.json`'s `model.base_url` — the container bakes in no endpoint.
- **Invocation:**

  ```bash
  podman run --rm --userns=keep-id --network corpos-net \
      -v ./in:/in:ro,Z -v ./out:/out:Z \
      lab-grounded-glyph-probe:dev run
  ```

  `--userns=keep-id` maps the host uid into the container so bind-mounted
  `/out` is writable by the non-root container user; `:ro` keeps `/in`
  read-only (the disposability contract); `:Z` relabels for SELinux.
- **Fallback (no shared network):** `--add-host=host.containers.internal:host-gateway`
  with `base_url = http://host.containers.internal:8081/v1` reaches
  llama-server's host-published port. Still explicit, still not `--network host`.

## Verified end-to-end (2026-07-08)

`lab-grounded-glyph-probe:dev` run on `corpos-net` against llama-server
(Qwen2.5-32B-Instruct-Q4_K_M) with hand-authored casg-direct v3 materials
produced a real `results.json` (3 conditions), per-condition response files,
and the extraction manifest. The `grounded_glyph` condition emitted the exact
CHANGELOG update the ground specifies while `baseline` gave a generic release
checklist — the casg-direct v3 signal, reproduced through the new container.
Pin/verify demonstrated: identical inputs verify OK, a one-byte material edit is
a hard refusal.

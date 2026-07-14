# corpos-lab — agent conventions (read first)

**corpos-lab** is the unified behavioral-experiment lab in Go: it absorbs lab-app's
battery/sequence runner and registry-lab's disposable-container assay model, driven by
**corpos** as the lab controller, persisting results through **corpos-toolkit**. It exists to
run the pre-registered glyph research program — read
**`~/dev/lab-app/resumption/CHARTER.md`** (claims C1–C3, instrument-freeze-by-digest rule,
aha-drain rule) before designing any study-facing surface. Companion docs beside it:
SALVAGE.md (what survives from the ancestors), BOUNDARY.md (everything is public; maturity
gates sharing), FIELD_NOTES.md (verified field positioning).

## Build & gate

- Go toolchain at `/usr/local/go/bin` (Go 1.26.3). Work from the repo root.
- **The gate is the law.** `scripts/gate.sh` is the single entrypoint, wired as the
  pre-commit hook via `core.hooksPath=.githooks`. **After cloning, run
  `./scripts/install-hooks.sh` once.** Six stages, all must pass to commit: `gofmt -s` drift ·
  `go vet` · golangci-lint v1.62.2 · govulncheck · `go build` · `go test -race` + a
  **95% coverage floor on `./internal/...`** (family floor; new code brings its own tests,
  sans-IO: injectable transports, `httptest`, temp dirs).
- Dev tools run pinned via `go run <tool>@<ver>` so `go.mod` carries no tool deps.
- **Gotcha (inherited from corpos):** golangci-lint's `gofmt`/`goimports` linters are
  **disabled** in `.golangci.yml` — v1.62.2 bundles a pre-1.26 gofmt that disagrees with the
  toolchain. The toolchain `gofmt -s` (gate stage 1) is the single formatting authority.

## Invariants (do not break)

- **CGo-free.** `CGO_ENABLED=0`; only pure-Go deps (`modernc.org/sqlite`, not `mattn`). Keeps
  distroless/scratch shipping open, matching the corpos family.
- **Results are provenance-stamped.** Every persisted result row carries the repo commit, the
  study-version manifest digest (`internal/digest`), the model artifact digest + sampling
  params, and the runner build. A result that can't reproduce its manifest is an anecdote,
  not data (CHARTER.md freeze-by-digest).
- **Assay containers are digest-pinned.** Images are referenced by content digest, never by
  mutable tag; the digest is part of the study manifest.
- **corpos-toolkit is reached over HTTP only** (`POST /mcp/<surface>` at
  `http://localhost:3001`). Never open the toolkit DB directly — the ledger stays owned by
  toolkit-server.
- **Claude-family models are never a treatment condition** (contaminated subjects — CHARTER.md).
  Local shelf only for treatment arms; see FIELD_NOTES.md §strand-5 for the verified shelf.
- **ONE local inference portal: llama.cpp (`llama-server` :8081).** Never install, start, or
  reach for a second inference server — not Ollama, LM Studio, vLLM, or anything else. Not as
  a fallback, not "just for this run", not because a model is already pulled there, **and not
  because the GPU is currently busy**. To run a different model, *swap the model on
  llama-server* (`systemctl --user stop llama-server-container` → run your model → `start` to
  restore). Measured: 4s to load Mistral-7B, 3s to restore Qwen-32B. A busy GPU is not a
  reason to route around the portal.
  **This invariant is written in blood.** On 2026-07-13 an agent needed Mistral, found the GPU
  busy and the Mistral GGUF missing, and used a retired-but-still-running Ollama daemon. It
  silently ran on **CPU** (5.1 tok/s vs ~60-100 on GPU) — undocumented — and was written up as
  a *"positive control on the ORIGINAL v3 runtime"*: a respectable-sounding sentence describing
  what was really "the only thing already running". The next session then froze Ollama-derived
  provenance into a study MANIFEST, compared llama.cpp-on-GPU against Ollama-on-CPU, declared
  **llama.cpp** the divergence, and proposed making Ollama's arbitrary sampler defaults
  canonical. Ollama was uninstalled 2026-07-14. If you catch yourself writing a principled
  reason for using a non-standard runtime, that sentence is the tell — you picked the
  frictionless path and reasoned afterward. See memory `one-local-inference-portal-llama-cpp`.

## Study discipline

- **Instrument freeze:** frozen artifacts change only via study version bump; mid-run
  insights drain to the toolkit suggestion/bug surfaces (aha-drain rule, CHARTER.md).
- Thinking-mode-capable models (Qwen3.6) have thinking explicitly pinned or treated as a
  condition — never left to default.

## Layout

`internal/digest` (SHA-256 content digests for freezing/pinning) — more as the port lands:
battery runner, assay model, controller client, persistence.

## Runtime deps (for live smokes)

- **toolkit-server** HTTP daemon at `http://localhost:3001` (surfaces: work / knowledge / fs /
  measure / ml / admin).
- **Local model:** llama.cpp server at `http://localhost:8081/v1` (OpenAI-compatible) — the
  **only** local inference portal (see Invariants). Containerized as the quadlet
  `llama-server-container`, reachable on `corpos-net` by DNS as `llama-server:8081`. GGUFs live
  under `/mnt/data1/models`, bind-mounted read-only at `/models`. Default model is Qwen2.5-32B;
  swap it by restarting the unit with a different `--model`, never by starting a second server.

## Git

**Worktrees only — the main checkout stays on `main`.** Never `git checkout -b` here; the
pre-commit guard (`scripts/guard-worktree-discipline.sh`) refuses commits on non-main
branches in the main checkout. Start work with `scripts/worktree-new.sh <slug>`; merge with
`scripts/worktree-merge.sh <branch>`; clear stragglers with `scripts/worktree-reap.sh`. Full
discipline: skill **`worktree-workflow`**. Every commit gates.

## Work ledger

`mcp__toolkit-server__work`, project **`glyph-research`** — chain `corpos-lab-unification`
(this repo's build-out), then the science chains (`instrument-parity-reproduction` onward).
The glyph-research roadmap is the ordering authority.

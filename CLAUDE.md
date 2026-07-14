# corpos-lab — agent conventions (read first)

**corpos-lab** is the unified behavioral-experiment lab in Go: it absorbs lab-app's
battery/sequence runner and registry-lab's disposable-container assay model, driven by
**corpos** as the lab controller, persisting results through **corpos-toolkit**. It exists to
run the glyph research program — read **[`INQUIRY.md`](INQUIRY.md)** (the questions Q1–Q3,
what we currently believe, how we measure) before designing any study-facing surface.
Companion docs: SALVAGE.md (what survives from the ancestors), BOUNDARY.md (everything is
public; maturity gates sharing), FIELD_NOTES.md (verified field positioning).

> `lab-app/resumption/CHARTER.md` — pre-registration, freeze-by-digest, the aha-drain rule —
> was **retired 2026-07-14** and INQUIRY.md replaced it. Anything still pointing at it
> (bug slugs, task names, an old comment) is naming a dead regime, not citing a live one.

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
- **Record what ran — observe, don't assert.** Every run captures the config it *actually*
  executed under, and every study declares its COMPLETE sampler chain (validation refuses a
  partial one: temperature alone does not define a distribution, and an unnamed stage
  inherits the server binary's default). Stored with the results, never edited afterwards:
  - **Per row** — the model the server says answered, its build id, and generation
    throughput. Throughput is the substrate tripwire: ~40-60 tok/s on GPU for a 7B against
    5.1 on CPU, so a fallback shows up as a step in the cells.
  - **Per run** — the sampler actually sent; the server's `/props` self-report (model_path,
    model_alias, build_info, n_ctx); a declared-vs-served **model mismatch**, recorded and
    never enforced; the image digest; the substrate probe; and the repo commit.
  **This replaces the retired freeze-by-digest rule** (see INQUIRY.md §What changed the
  method). The freeze pinned six files that never changed and was blind to all three
  variables that actually moved — temperature, sampler, and processor were unrecorded, so a
  study "reproduced" a result while running on a different processor and nobody could tell.
  A run record is permanently true; a manifest was a promise about the future that locked us
  into mistakes. **Observation beats assertion.**
  *Two things `/props` is NOT, because both are easy to assume and one is written into a
  bug's own acceptance criteria: it does not report the effective sampler (it reports the
  server's LAUNCH defaults — send temperature 0.0 to a server started at 0.8 and it still
  says 0.8), and it is served at the ROOT, not under `/v1`. The effective sampler is sound
  because every stage is pinned and sent, verified against `/slots`.*
  *Nothing here may fail a run.* A failed `/props` readback, an unknown substrate, an
  unstampable repo, a model mismatch — each is recorded as the gap it is, never guessed at
  and never grounds for refusal. A run that cannot fully describe itself is still a run;
  refusing it is the freeze reflex wearing a different hat.
- **Reference images by content digest, never a mutable tag** — so a run record says exactly
  what executed. Recorded, not enforced: a digest that doesn't match a prior run is
  information about the two runs, not grounds for refusing to run.
- **corpos-toolkit is reached over HTTP only** (`POST /mcp/<surface>` at
  `http://localhost:3001`). Never open the toolkit DB directly — the ledger stays owned by
  toolkit-server.
- **Claude-family models are never a treatment condition** — contaminated subjects: this
  corpus's terrain is in their training data and they ceiling at baseline. They may judge;
  they are never a treatment arm. This is methodology, not ceremony, and it outlived the
  charter that first wrote it down (INQUIRY.md §How we measure). Local shelf only for
  treatment arms; see FIELD_NOTES.md §strand-5 for the verified shelf.
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

Read **[`INQUIRY.md`](INQUIRY.md)** before designing anything study-facing: the questions,
what we currently believe, and how we measure. It is a living document — revise it when you
learn better. It replaced `lab-app/resumption/CHARTER.md`, retired 2026-07-14.

- **If the instrument is broken, fix it. Now.** Do not file it and keep running.
  The retired charter's "aha-drain" rule said mid-run insights route to the suggestion box
  and may only be applied at a study-version bump. That rule *institutionalised proceeding
  with known problems*: when we found the probe pinned temperature 0.0 — making a graded grid
  structurally impossible — it told us to file it and run anyway. Notice → fix → write down
  what you fixed and why. The suggestion/bug surfaces are still the right place for
  *proposals*; they are not a waiting room for *repairs*.
- **There is no parity, because we are the frontier.** Don't target prior results. Parity
  fixes a goal from an earlier state of our own process, so any improvement — a new decomp
  step, a tightened battery item — registers as divergence when it's just truer. The
  2026-04-03 demotion is the model of correct behaviour: the battery tightened, prior passes
  no longer met it, ALPHABET was emptied and said so. Old runs are orientation, never a target.
- **Read cells, not counts.** n=8 gives a 95% CI roughly ±0.2 wide. Comparing single integers
  between grids (7/8 vs 4/8) is comparing noise. What reproduces is the phenomenon and its
  direction; the exact count does not and should not be expected to.
- Thinking-mode-capable models (Qwen3.6) have thinking explicitly pinned or treated as a
  condition — never left to default.

## Layout

`internal/` — `assay` (the grounded-glyph probe + the sampler regime), `battery` (the 15-item
runner), `control` (host-side controller; `Deps` is where the run reaches the world),
`digest`/`image` (content digests, recorded not enforced), `manifest` (computed, no verify
arm — it was deleted), `model` (the inference seam + llama-server client), `persist` (toolkit
HTTP), `provenance` (repo stamp), `runner` (container-side executor), `study` (TOML def +
validation), `substrate` (GPU/CPU probe). `cmd/corpos-lab` is the host CLI; `cmd/lab-assay`
runs inside the container.

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

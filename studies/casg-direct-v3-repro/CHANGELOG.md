---
type: reference
last_updated: 2026-07-14
---

# CHANGELOG — casg-direct-v3-repro

Study-version history. Per CHARTER.md, a version bump is the **only** moment a
drained suggestion or bug may be applied to the instrument, and every bump lists
what it applied. Runs never span versions; results attach to exactly one.

---

## v2 — 2026-07-14 (registered, not yet run)

**Manifest:** `MANIFEST.v2.sha256` · **Registered by:** task `register-v2-manifest`,
chain `instrument-hold-lift` · **Runner build:** `d38293e`

Registered so the container leg can run at all. v1 could not have produced a
valid graded grid, and would have shipped fabricated fails to the canonical DB.

### Drain items applied

Both were filed during earlier work and deliberately **not** hot-patched; this
bump is their sanctioned landing point.

| Item | Kind | Landed |
|---|---|---|
| `grounded-probe-temp0-cannot-reproduce-graded-grids` (id 64) | suggestion | `d38293e` |
| `grounded-probe-misapplies-battery-pass-fail-parser` | bug | `c6ca725` |

**Suggestion 64 — the probe could not express a graded grid.** `probeGenParams`
pinned temperature 0.0 in Go with no seed. Greedy decoding returns an identical
reply for every replicate, so a cell could only score 0/8 or 8/8. The v3 parity
target is graded (baseline 7/8, grounded 7/8) — the instrument structurally
could not reproduce the grid it was to be checked against, and a temp-0 run
could have read as spurious NON-reproduction. Sampling now lives in `study.toml`,
which also places it under the freeze (see *Why sampling moved* below).

**The PASS/FAIL bug — the probe fabricated failures.** `RunProbe` scored replies
with the battery's `ParseModelVerdict`, which demands a literal `PASS`/`FAIL`
prefix. Probe replies are conduct and open with neither, by design, so every run
in every condition returned a spurious fail carrying the whole response as its
reason — and `internal/persist` shipped those rows to the canonical toolkit DB.
The probe now carries its own rubric type and emits `Unscored` only. Caught by
reading before the first container run, so no result was invalidated.

### Instrument changes

- **Image digest-pinned.** v1 carried the placeholder tag
  `localhost/lab-grounded-glyph-probe:dev` because no image existed at exhume
  time. Built via `scripts/build-lab-images.sh` and pinned to
  `sha256:c0d44d61…`. A mutable tag violated both the corpos-lab invariant and
  the freeze rule. (`deploy/IMAGE_DIGESTS.txt` is a gitignored build artifact;
  `MANIFEST.v2.sha256` is the durable record. Re-verify with
  `scripts/build-lab-images.sh --digests`.)
- **Quant resolved.** v1's `version = "v0.3-q4km-UNVERIFIED"` → `v0.3-q4km`. The
  ollama control settled it: `mistral:latest` is Mistral-7B-Instruct-v0.3 at
  Q4_K_M, 7.2B.
- **Model artifact pinned, and it is the control's own artifact.** The GGUF was
  staged byte-for-byte from ollama's blob store (`sha256:f5074b12…`), where the
  model layer of `mistral:latest` *is* the raw GGUF and blobs are named by their
  SHA-256. The recomputed digest matches the blob name exactly. This removes
  model weights as a variable in the container-vs-control comparison — the two
  legs run identical bytes.
- **Sampling declared:** temperature 0.8, seeds 1..8, max_tokens 512 — matching
  the control exactly.

### Unchanged, deliberately

`materials/scenario.md`, `materials/glyph.md`, `materials/ground.md`, and
`SCORING_RUBRIC.md` are **byte-identical to v1**. v2 changes how the instrument
samples and scores; it does not change what the subject is shown. If a material
digest ever moves, this stops being a parity reproduction of v3 and the
comparison to `PARITY_TARGET.md` is void.

### Why sampling moved out of Go

CHARTER requires sampling params to be pinned per study version. As a Go
constant, sampling was pinned only by the runner build. Declared in `study.toml`
it rides into the derived `study.json`, which `manifest.Compute` already digests
as `StudyDigest` — so the requirement is met *structurally* rather than by
anyone remembering to record a number. A sampling edit necessarily moves the
digest and invalidates results attached to the old manifest.

### KNOWN DELTAS — container leg vs the ollama positive control

The verdict task compares the container leg against both v3 and the 2026-07-13
control, treating a container/control gap as instrument-attributable. These are
the differences that gap could legitimately come from. **None is a defect; all
are recorded here so the verdict does not mistake one for a finding.**

1. **Temperature is inferred, not recovered.** v3's `study.json` never recorded
   one. It is permanently unknowable. 0.8 is the ollama-era default and what the
   control used; the *graded* v3 grid is the evidence the original sampled at
   all, since that variation is impossible at temperature 0.

2. **Endpoint and chat template differ.** The control used ollama
   `/api/generate` with a raw prompt, which ollama wraps with its own template
   (`[INST] {prompt}[/INST]`). The container leg uses llama-server's OpenAI
   `/v1/chat/completions`, where llama.cpp applies the GGUF's *embedded*
   template (`{bos_token}[INST] {content} [/INST]` — note the space before
   `[/INST]` and the explicit BOS). The prompts are therefore not byte-identical.
   This delta is **inherent to the rig being validated**: corpos-lab is an
   OpenAI-compatible client by design, and that design is what the parity gate
   exists to test.

3. **Stop sequences absent.** ollama's `mistral:latest` params layer sets
   `stop: ["[INST]","[/INST]"]`; corpos-lab's model client cannot send stop
   sequences at all. Mistral may emit a literal `[INST]` and hallucinate a
   follow-up turn inside one response. Unlike (2), this is a missing capability
   rather than a design consequence — filed as suggestion
   `model-client-cannot-send-stop-sequences`, to land at a **future** bump. It is
   NOT applied here: v2 is registered, and reopening it mid-registration is the
   drift the freeze rule forbids.

4. **Seeds do not transfer across runtimes.** Seeds 1..8 match the control's
   numerically, but identical seeds do **not** yield identical outputs across
   ollama and llama.cpp — different prompt bytes (2) and different sampler
   implementations. Seeds buy reproducibility *within* this rig. The container
   leg is therefore **not** expected to match the control run-for-run, and any
   comparison at that level would be a category error. The pre-registered
   tolerance is stated at the **cell** level (glyph-only ≤2/8 C, grounded ≥6/8 C),
   which is the level that is meaningfully comparable.

5. **Context window.** The control ran ollama's Mistral default (32768). The
   llama-server container is currently configured `--ctx-size 8192` for
   Qwen-32B; the Mistral swap must set a context that fits the prompts
   (materials + 512 generated tokens) and record it with the run.

---

## v1 — 2026-07-13 (corpus-side only; never backed a scored run)

**Manifest:** `MANIFEST.v1.sha256` · **Registered by:** task `exhume-v3-materials`,
chain `instrument-parity-reproduction`

Exhumed the v3 study materials into a corpos-lab study definition and pinned the
corpus-side set (study.toml, rubric, parity target, three materials). Left two
explicit `KNOWN DELTA` markers for the run task to resolve: the placeholder image
tag and the unverified quant. Both are resolved at v2.

**No run was ever executed against v1.** The 2026-07-13 reproduction was a
positive control on the *original* ollama runtime with its own separate
`runs/ollama-repro-2026-07-13/RUN_MANIFEST.json`; it deliberately did not go
through the container path. Nothing therefore attaches to v1, and the v2 bump
invalidates no result.

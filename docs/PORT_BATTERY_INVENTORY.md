# Battery Runner Port — Feature Inventory & Parity Audit

Port of lab-app's ALPHABET Entry Battery (Rust) → corpos-lab (Go), per
code-migration-discipline. Source: `~/dev/lab-app/crates/` at `ae6611d`.
Scope per task `port-battery-runner-to-go`: Phase-1-implemented items only
(1, 2, 4, 9, 10, 15); the 9 deferred items register as Deferred stubs exactly
as in the source.

## Features ported (from source-test inventory + API enumeration)

**Sequence engine** (`lab-app-server/src/sequences/mod.rs`, 8 tests)
- Named ordered step list; `AddStep` / `StepCount`.
- Fail-fast runner: stops on first `Fail` verdict or `Error`; `Deferred` and
  `Flag` do NOT stop the sequence (deferred items must not hide downstream fails).
- Result: per-step results with wall-clock durations, `Passed`, `ExitIndex`
  (last-run step index; 0 on empty sequence), `FailureReason` from the
  stopping Fail's reason or Error's message.
- Prior-step lookup by name from within a step (observation → judge pattern).
- Empty sequence passes.

**Typed verdicts** (`lab-app-types/src/verdict.rs`, 18 tests)
- `Verdict` kinds: pass / pass_with_condition / flag / deferred / fail
  (optional 1-based item) / not_applicable. Invalid combinations
  unconstructable via constructors.
- `Passed()` true for pass, pass_with_condition, not_applicable.
- Display strings preserved verbatim: `PASS`, `PASS* (c)`, `FLAG: r`,
  `DEFERRED: p`, `FAIL at item N: r`, `FAIL: r`, `N/A: note`.
- JSON: kind-tagged snake_case (`{"kind":"fail","item":3,"reason":"…"}`),
  `item` omitted when absent — matches the Rust serde shape.
- `ComposeBatteryVerdict`: Block > Defer > Promote; `Flag` is a soft-fail
  included in Block's failures (class executor_correctable, item 0);
  `Fail{item:nil}` maps to item 0 with class unknown.
- `BatteryVerdict` display: `PROMOTE` / `BLOCK (N failure[s])` / `DEFER (N pending)`.
- `RouteChainFailure`: any definition-gap or unknown class → StopAndStudyDefinition;
  all executor-correctable → RequeueToResearcher; empty → RequeueToResearcher;
  PromoteAnyway never returned.

**Battery assembly** (`sequences/battery.rs`, 3 tests)
- 15 steps registered in stratum order with the exact source step names
  (`item1-xyz-specificity` … `item15-fallout-profile`).
- `StepVersion(name)`: implemented items `0.1.0`, deferred items
  `0.0.0-deferred`, unknown names `0.0.0`.

**Static items** (`steps/battery/static_steps.rs`, 11 tests)
- Item 2 intent-language scan: 8 reject phrases, case-insensitive substring,
  first match fails with reason `firing condition contains intent-modeling
  language: "<phrase>"`.
- Item 4: always `NotApplicable` ("item 4 does not apply to glyph entries").
- Item 10 axis presence: Y-marker gate (`**Y` / `Y —` / `Y marker`), then
  Marker/Aim/Rest each satisfied by `**<axis>` / `<axis> axis` / `<axis>:`
  (case-insensitive); failure lists missing axes in order.
- Item 15: `**Fallout profile:**` field presence.
- Deferred registry: per-item pending reasons preserved verbatim (items
  3, 5, 6, 7, 8, 11, 12, 13, 14 + generic fallback).

**Inference items** (`steps/battery/verdict_steps.rs`, 11 tests)
- `ParseModelVerdict` — the single prose→verdict seam: leading PASS (any
  case) → Pass; leading FAIL → Fail with `Item N: <rest-trimmed>` (or
  `Item N: unspecified failure`); anything else → Fail
  `Item N: model response did not start with PASS or FAIL: <text>`.
- Items 1 (X/Y/Z specificity) and 9 (universality): prompt text preserved
  verbatim; temperature 0.0, max_tokens 256; model transport error →
  StepOutcome error `Item N: model error: <err>`.
- Model seam: injectable `model.Client` interface; production impl is an
  OpenAI-compatible chat-completions client for llama-server
  (`POST <base>/chat/completions`), sans-IO tested via `httptest`.

**Provenance** (`build.rs` + store stamping)
- Git commit SHA + dirty flag captured; in Go captured at RUNTIME via
  `git rev-parse HEAD` + `git status --porcelain` (Go has no build.rs;
  runtime capture is fresher — it stamps the tree the run actually used).
  Stamped into the battery run result for task 5 to persist.

**Integration fixtures** (`tests/battery_integration.rs`, 2 tests + fixtures)
- `known_pass.md` passes the full 15-step battery (all 15 step results present).
- `known_fail_item2.md` fails at item 2 (index 1) with the intent-language reason.
- Fixture files copied verbatim to `internal/battery/testdata/`.

## Deliberately dropped (named, with reasons)

| Dropped | Why |
|---|---|
| Axum HTTP API (`api/*`: battery/health/metrics/state routes) | corpos-lab is CLI/controller-driven, not a service (task constraint). |
| `Translator` seam + `PassthroughTranslator` | Task constraint; battery steps never call it (verified: only studies do). `SequenceInput` in Go carries no translator. |
| Ollama + Anthropic model backends, retry/backoff logic | Target is llama-server's OpenAI-compatible API only. Anthropic is charter-prohibited for treatment arms. Retry policy deferred to the controller task (filed in handoff). |
| `MockModel` response-queue test double | Go idiom: fake `model.Client` funcs + `httptest` servers express the same test scenarios (mock-idiom translation per discipline). |
| sqlx persistence (`store/`, `battery_runs` table, version-wrapped column JSON) | Task 5 (`persist-and-dashboard`) owns persistence to the toolkit DB. `StepVersion` is ported now; the version-wrapping envelope moves to task 5. |
| `metrics::`/`tracing::` emission | corpos-lab has no metrics sink yet; durations stay in results. Observability lands with the controller. |
| `score_grid`, `studies/*`, `impact`, `assay` types | Not battery; later chain tasks. |
| `DemotionTrigger`/`RepickupPath`/`route_demotion` | Corpus-workflow routing, not battery. `FailureClass` + `RouteChainFailure` ARE ported (battery-verdict surface). |
| Deferred battery items 3,5,6,7,8,11–14 as real checks | Out of scope per task constraint (new work, not port work). Stubs port exactly. |

## Accepted divergences (parity audit, phase 1)

1. `FailureReason` is a Go string ("" = none) vs Rust `Option<String>` — idiomatic;
   no wire-shape consumer of the old value exists.
2. `FailureClass` JSON is kind-tagged snake_case in Go vs serde external tagging
   in Rust — no reader of old-format rows exists (old lab-app.db is historical
   reference only, per SALVAGE.md).
3. Provenance captured at runtime, not build time (see above) — strictly more
   accurate for worktree-based runs.
4. No metrics/tracing side effects (dropped, above).

## Parity mapping table (source test → Go test)

Filled as the TDD loop lands; see `PORT_BATTERY_PARITY.md` §mapping after the
final commit of this worktree.

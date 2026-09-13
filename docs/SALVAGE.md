# SALVAGE.md — Verified Salvage Inventory (charter-and-salvage / ratify-salvage-inventory)

**Date:** 2026-07-08
**Method:** every claim below was verified against the filesystem and git of the three ancestor
repos (`~/dev/lab-app`, `~/dev/registry-lab`, `~/dev/seed-packet`), not quoted from their docs
(state-verification-discipline). Inventory pass was read-only; nothing was moved or deleted.

**Verdict vocabulary:** **port** (carry into the resumed program as live material) ·
**reference** (consult during rebuild; do not carry as-is) · **archive** (keep where it is,
inert; recoverable if needed) · **dead** (does not exist / pointer target missing).

**Repo git states at inventory time:**

| Repo | Branch | State | Last commit |
|---|---|---|---|
| lab-app | main | untracked `offload/`, `resumption/` | `ae6611d` 2026-04-29 "archive(seed-packet): import lab-controller crate from seed-packet" |
| registry-lab | master | **dirty** — unstaged deletions (`baseline/alphabet-scope-survey.md`, `baseline/navigation-findings*.md`, `design/coupling-spec.md`) | `c47c3e6` 2026-04-21 |
| seed-packet | main (tracks origin/main) | 2 untracked curricula files | `798f661e` 2026-06-08 |

---

## 1. Glyph candidates — **port**

`~/dev/lab-app/corpus/glyph-model/candidates/` — **exactly 8 files, confirmed on disk:**

- `CANDIDATE_casg-delegate_2026-04-03.md`
- `CANDIDATE_casg-direct_2026-04-03.md`
- `CANDIDATE_conditional-gate-uniform-default_2026-03-30.md`
- `CANDIDATE_discovery-event-non-recording_2026-03-31.md`
- `CANDIDATE_formal-step-context-bypass_2026-03-29.md`
- `CANDIDATE_governed-operation-protocol-bypass_2026-03-30.md`
- `CANDIDATE_parent-state-check-bypass_2026-03-30.md`
- `CANDIDATE_structural-ceiling-bypass_2026-03-29.md`

Supporting spec docs in the same dir (GLYPH_DEFINITION.md, GLYPH_WRITING_SPEC.md,
GLYPH_ENTRY.md, GLYPH_DECOMPOSITION_PROCESS.md, GLYPH_PROVENANCE_TYPES.md, ALPHABET.md,
LENS_CORPUS.md) all present — **port** with the candidates.

## 2. ALPHABET Entry Battery — spec **port**, Rust impl **reference**

- Spec: `~/dev/lab-app/corpus/glyph-model/ALPHABET_ENTRY_BATTERY.md` — 15 items confirmed
  (Items 1–6 structural validity gate, Items 7–15 behavioral-load safety gate).
- Implementation: `~/dev/lab-app/crates/lab-app-server/src/sequences/battery.rs` +
  `steps/battery/{static_steps,verdict_steps}.rs`. **6 items implemented** (1, 2, 4, 9, 10, 15),
  confirmed in the battery.rs sequence header; the other 9 are `deferred_pending` steps with
  documented reasons (corpus access, proto-ethos context, multi-sub-check complexity).
  Verdict **reference**: corpos-lab (Go) reimplements; the Rust code is the behavioral spec.
- Run records: `battery-results/` — 8 result records covering 7 distinct entries + `_TEMPLATE.md`;
  `battery-runs/BATTERY_RUN_alphabet-rebuild_2026-04-03.md`. Verdict **archive** (provenance data).
- `COMPOUND_TEST_RESULTS_2026-04-03.md`, `CORE_FILE_STATUS.md` present — **archive**.

## 3. Grounded-probe SERIES records — **port**

`~/dev/lab-app/corpus/studies/` — 27 study dirs; **8 SERIES records confirmed:**

| Study | SERIES file | Notable verified state |
|---|---|---|
| casg-delegate | `assay-grounded-casg-delegate/SERIES.md` | — |
| casg-direct | `assay-grounded-casg-direct/SERIES.md` | v1–v3 dirs present. **v3 WAS the parity target; the parity premise was VOIDED 2026-07-14** — April's conditions (temperature, sampler, processor) were never recorded and the 7/8 target was a point estimate at n=8 (95% CI [0.53, 0.98]), so it was noise. These numbers orient; they are **not a target and not a baseline**. Verified state as recorded: baseline calibration 7/8 wrong-execution-fidelity C both models; Claude glyph 8/8 C (GLYPH SUFFICIENT); Mistral glyph 0/8 C / 7/8 Ii; **grounded 7/8 C, lift 0→7 (GROUND CONFIRMED)** — instruction-shaped ground Action field overcomes prepend-delivery analysis-mode |
| conditional-gate-uniform-default | `assay-grounded-conditional-gate-uniform-default/SERIES.md` | — |
| discovery-event-non-recording | `assay-grounded-discovery-event-non-recording/SERIES.md` | — |
| formal-step-context-bypass (FSCB) | `assay-grounded-formal-step-context-bypass/SERIES.md` | v1–v10 dirs; SERIES table documents v1–v6 only (v7–v10 undocumented in SERIES — gap) |
| governed-operation-protocol-bypass | `assay-grounded-governed-operation-protocol-bypass/SERIES_governed-operation-protocol-bypass_2026-04-10.md` | — |
| parent-state-check-bypass (PSCB) | `assay-grounded-parent-state-check-bypass/SERIES_parent-state-check-bypass_2026-04-10.md` | v1–v2 dirs |
| structural-ceiling-bypass | `assay-grounded-structural-ceiling-bypass/SERIES.md` | — |

Also present and salvageable as **port**: ecological/mistral/blank assay dirs (14 more study
dirs), `BEHAVIORAL_STUDY_PROTOCOL.md`, `alphabet-scope-survey.md`,
`item12-decomp-spike-findings.md`, `training-data-gaps-v2.md`.

## 4. FSCB v6 + PSCB v2 — **port** — ⚠ received claim CORRECTED

The resumption docs (and the `delivery-register-completion` chain note) describe these as
"scaffolded-but-unrun" / "orphaned". **Disk says otherwise:**

- **FSCB v6** (`assay-grounded-formal-step-context-bypass/v6/`) — scaffold files confirmed:
  `GLYPH_formal-step-context-bypass.md` (agent-facing format variant),
  `GROUND_formal-step-context-bypass.md`, `SCENARIO_*_claude.md`, `SCENARIO_*_mistral.md`,
  `SCORE_GRID.md`, `study.json`, `blanky_files/`. **48 response files exist** (8 × 6 conditions,
  all non-trivial, dated 2026-04-20). SCORE_GRID is scored for baseline (Claude 1/8 C, Mistral
  0/8) and glyph_only (**Claude 6/8 C — the agent-facing glyph format moved Claude from 1/8 to
  6/8**; Mistral 0/8); **grounded_glyph responses exist but are unscored** (grid shows `?`).
  → The orphaned work is *scoring the grounded condition + SERIES analysis*, not running.
- **PSCB v2** (`assay-grounded-parent-state-check-bypass/v2/`) — same scaffold file set
  confirmed. 48 response files. **SCORE_GRID fully scored, all six conditions** (baseline 8/8 C
  both; glyph_only Claude 8/8 / Mistral 0/8 with clean Ii; grounded Claude 3/8 / Mistral 6/8).
  The SERIES v2 outcome cell is **empty** — analysis/rollup never happened.
  → The orphaned work is *SERIES analysis + interpretation*, not running.

Downstream impact: `delivery-register-completion`'s "run orphaned FSCB v6 + PSCB v2" task
should be re-scoped to "score/analyze existing v6+v2 responses; re-run on the new rig only if
the parity gate demands fresh data."

## 5. Ouija experimental archive — **dead in working tree / archive via git** — ⚠ CORRECTED

The entire `experimental/` tree (accounts, briefs, journals, inquests, protocols) was
**deliberately deleted** in seed-packet commit `2e967050` ("chore(archive): subsume
seed-packet-archive and experimental into archive/" — 792 deletions, 7 docs promoted to
`archive/`). Nothing ouija-named survives in any working tree.

- Recoverable at `git -C ~/dev/seed-packet show 2e967050^:experimental/...` (verified:
  `EXPERIMENTAL_INQUEST_ouija_2026-03-23.md` resolves at that ref).
- What DID survive to working trees: the methodology paper
  `seed-packet/process-docs/papers/ouija-methodology/PAPER_ouija-methodology_2026-03-26.md`
  (**port**) and the promoted docs `archive/inquest-charter.md`,
  `archive/light-touch-investigation.md`, `archive/proto-ethos.md`, `archive/taboo-definition.md`,
  `archive/duty-crafting.md`, `archive/BRIEFS_INDEX.md` (**reference**).
- The `ouija-sealed-execution` chain should treat git history as its archive source, not the docs'
  `experimental/` paths.

## 6. Taboo-source corpus — **port**

`~/dev/lab-app/corpus/glyph-model/taboo-source/` — **114 top-level `TABOO_CANDIDATE_*.md` files
+ 77 in `pending/` = 191 total**, confirmed by count. (Docs variously say "~120"; use 191.)

## 7. Papers + stubs — **port**

`~/dev/seed-packet/process-docs/papers/` — verified tree:

- Full papers (3): `comprehension-as-compliance/PAPER_comprehension-as-compliance_2026-03-29.md`,
  `ouija-methodology/PAPER_ouija-methodology_2026-03-26.md`,
  `thinking-trace-analysis/PAPER_thinking-trace-analysis_2026-03-26.md`
- `PLAIN_PROJECT_STATEMENT.md`
- Stubs (**8 confirmed**, `stubs/`): behavioral-equivalence-assay-methodology,
  canon-suppression-study, cartographer-duty-format, comprehension-as-compliance,
  generativity-battery-instrument, geometric-verification-study,
  grounded-glyph-probe-delivery-register, wrong-path-correct-verdict

## 8. Multi-model baseline runs — **reference**

`~/dev/lab-app/baseline/` — richer than the received claim ("library-runs"):

- `library-runs/` — **8 model dirs** (gemma-3-12b, mistral-7b-v0.3, phi-4, qwen2.5-7b/14b/32b,
  qwen2.5-coder-14b, qwen3.6-27b) + `LIBRARY_FINDINGS.md`
- `forge-runs/` — same 8 models + `FORGE_FINDINGS.md`
- `calc-runs/` — same 8 models + `CALC_FINDINGS.md` (+ `calc/` Rust harness)
- Survey docs: capability-matrix.md, candidate-list.md, criteria-map.md, harness-design.md,
  pat-coverage.md, role-inventory.md, role-recommendations.md, smell-test-checkpoint.md, unit-list.md

Verdict **reference**: continuity anchors for the mid-2026 shelf refresh (Mistral-7B-v0.3 is the
explicit parity anchor); not live instruments.

## 9. registry-lab container environment — **reference** (shims **archive**)

- `Dockerfile.base` + 4 assay Dockerfiles, `LAB_CONTAINER_MODEL.md`, `ENV_SPEC.md`, `Makefile`,
  `lab-entry`, `lab-write-manifest` — present. **reference** for corpos-lab's container design.
- `assays/{behavioral-equivalence,decomposition,grounded-glyph-probe,structural-glyph-probe}/run-assay.sh`
  — 4 scripts, 54 lines each, **confirmed placeholder shims**: each writes a
  `status: "pending-adapter"` results.json and exits ("host-side adapter not yet wired,
  task e9-lab-controller"). No assay logic exists in them. **archive**.
- Working tree is dirty (unstaged deletions, see git-state table) — commit or discard before
  any salvage copy from this repo.

## 10. Other verified survivals

- `~/dev/lab-app/seed-packet-archive/` — `tools/lab-controller/` (imported crate, `ae6611d`),
  `tools/benchmarks/`, `assay-results/` (claude/, mistral/, PROTOCOL.md, VALIDATION_RUNBOOK.md,
  BENCHMARK_TRACKING.md), `process-docs/studies/`. **reference**.
- `~/dev/lab-app/resumption/SWEEP_2026-07-07.md` + `SWEEP_RAW_CLAIMS_2026-07-07.md` — field-sweep
  digest (claims UNVERIFIED pending the re-run; `field-positioning-sweep` task owns verification).
- `~/dev/seed-packet/process-docs/STUDIES.md` (25 KB, 2026-05-07) — the study registry. **port**
  after the dead pointers below are annotated.

---

## Known dead pointers (seed for `papers-and-library-honesty`)

Confirmed dead — target verified absent from every working tree:

1. **`RESULTING_PROPOSED_TABOO_CERTIFICATIONS.md`** — referenced by the (now-deleted)
   experimental README as "primary output — 24+ candidates". **Zero git history — never
   created** (confirmed `git log --all` empty). STUDIES.md line 214 already documents this;
   the candidates it should have held are distributed across the two ouija journals
   (recoverable at `2e967050^`).
2. **STUDIES.md → `experimental/*` paths** (line ~213: `experimental/README.md`,
   `experimental/inquests/EXPERIMENTAL_INQUEST_ouija_2026-03-23.md`,
   `experimental/protocols/ouija-facilitating-agent-protocol.md`, plus journal refs) — all
   deleted in `2e967050`; recoverable only via git.
3. **STUDIES.md line 10 → `LIBRARY.md`** — retired in seed-packet `e45e7985`
   ("refactor(library): retire library-tool crate + LIBRARY.md / LIBRARY_INDEX.md").
   The library now lives in `library-data/entries/`; the pointer text predates retirement.
4. **STUDIES.md → `discoveries.md`** (line ~157) — NOT a dead pointer; it is the name of a
   phantom artifact Claude invented during the discovery-event-non-recording v3 study.
   Listed here so the honesty sweep doesn't "fix" it.
5. **registry-lab STUDIES/README claims about runnable assays** — the four run-assay.sh shims
   self-describe as placeholders; any doc claiming runnable assays is overstating.

Corrections to received claims (docs vs disk): FSCB v6 and PSCB v2 are **run, not unrun**
(§4); the ouija archive is **deleted, not present** (§5); taboo corpus is **191 files, not
~120** (§6); baseline runs cover **three suites** (library/forge/calc), not one (§8); FSCB
SERIES.md is missing rows for v7–v10 despite those dirs holding 16 responses each (§3).

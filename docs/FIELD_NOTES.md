# FIELD_NOTES.md — Field Positioning, Verified (2026-07-08)

**Provenance:** digest of `SWEEP_2026-07-07.md` (deep-research sweep, 5 angles → 24 sources →
115 claims; raw dump in `SWEEP_RAW_CLAIMS_2026-07-07.md`). The sweep's own verification re-run
never landed verdicts on disk, so load-bearing claims were **verified directly at primary
sources on 2026-07-08** (arXiv abstract pages + web corroboration). Verdict vocabulary:
**CONFIRMED** (primary source matches the sweep's characterization), **CONFIRMED-CORE**
(headline matches; specific numbers/details live in the paper body and remain unverified),
**FLAGGED** (a discrepancy found — do not cite the sweep's version).

Claims not verified here (strand-4 security items, METR/AISI/venue items) remain **leads**:
verify before citing in any paper or design (constraint inherited from the sweep task).

---

## Scoop-check verdicts (what the charter claims needed to survive)

**(a) C1 — glyph vs information-matched imperative: NEIGHBORS-EXIST-WITH-DIFFERENCES.**
Nobody runs a descriptive-terrain (three-axis, from-inside) format against an
information-matched imperative rule. The closest neighbors, all verified:
- arXiv **2602.21223** "Measuring Pragmatic Influence in Large Language Model Instructions" —
  CONFIRMED-CORE (directive-framing decomposition, 400 framing instantiations / 13 strategies /
  4 mechanism clusters, priority-based measurement, 5 LLMs). It varies *framing* (social
  wrapper), not *format* (descriptive map vs directive) with information matched. Cite as the
  closest methodological cousin. (Its length-matched-control and model-list details are
  body-level — verify before citing those specifics.)
- arXiv **2602.11988** "Evaluating AGENTS.md" (ETH: Raychev/Vechev group) — CONFIRMED (no
  average success benefit, >20% inference-cost increase, imperative instructions followed,
  descriptive repo overviews no benefit). Their "descriptive" is repo-orientation prose, NOT
  decision-point terrain description — exactly the distinction C1 isolates.
- arXiv **2602.04297** — CONFIRMED (prompt sensitivity confounded by underspecification;
  note it is text-classification scoped, so cite the principle, not agent-task evidence).
  Consequence: T1/T2 must be matched on specification level, not just facts — already in the
  matched-content design task's AC.
- arXiv **2605.10039** "Instruction Adherence in Coding Agent Configuration Files" (McMillan)
  — CONFIRMED (1,650 Claude Code sessions; four structural variables null after correction;
  **~5.6% lower compliance odds per generation step within a session**; task identity swamps
  structure). Supports the program premise (blind scaffolding doesn't pay) and gifts a cheap
  add-on measurement: does glyph preload decay within-session? (delivery-register-adjacent.)

**(b) C3 — false-positive overhead / Rest-axis ablation: CLEAR** (adjacent regimes exist, no
overlap on the measured object). Verified neighbors:
- arXiv **2606.15034** "OSGuard" (Mohammadmirzaei & Flanigan) — CONFIRMED (benign
  instructions, hazards from environment state, allowed/unrelated/unsafe taxonomy, measures
  guardrail false positives on benign actions). It measures *guardrail* over-blocking;
  nobody measures loaded behavioral *guidance* over-firing in neutral territory, and no
  surveyed format has an applicability-boundary axis to ablate. Position the benchmark in
  OSGuard's named benign-context regime.
- arXiv **2603.01246** "Defensive Refusal Bias" — CONFIRMED-CORE with one **FLAG**: the
  keyword-driven effect (2.72×) and the authorization-increases-refusals paradox are
  confirmed; the sweep's "12.2% refusal across 2,390 prompts" figure does NOT match the
  abstract (which reports 34–44% for critical task categories). Do not cite 12.2% without a
  body-read.
- arXiv **2605.05427** "The Refusal–Compliance Tradeoff" — CONFIRMED-CORE (21 open-weight
  models; over-refusal is a poor proxy for safety / independent failure mode; family-stable
  calibration; Qwen permissive). **FLAG (load-bearing for C3):** the single-LLM-judge
  automation claim (cross-evaluator r=0.990) is NOT in the abstract. The rest-axis design
  task's AC now carries a verification gate: confirm in the paper body before the scoring
  rubric freezes, else mechanical markers are primary. Also note the abstract's framing that
  permissive families "tolerate higher harmful compliance" — the sweep's "Qwen low-ORR" gloss
  is directionally right but incomplete.

**Verdict for the program: neither flagship experiment is scooped.** C1's contribution
statement must be precise — information-matched *format* (terrain description vs directive),
which neither the framing work (2602.21223, 2603.14373) nor the content-category work
(2602.11988) isolates. C3's gap (guidance over-firing + Rest ablation) remains unoccupied.

## Per-strand summaries

**Strand 1 — format vs content:** see scoop-check (a). Design upgrades adopted into the
matched-content design task (verified present in its AC 2026-07-08): specification-level
control, ceiling-avoidance via directive-conflict-style instrument, related-work positioning.
Length-matched filler control retained in CHARTER.md C1 as condition T3.

**Strand 2 — behavioral overhead:** see scoop-check (b). Single-judge gate filed as a task
edit; Evaluation Card requirement confirmed real (see strand 6).

**Strand 3 — scaffolds as research objects:** 2605.10039 + 2602.11988 both CONFIRMED (above).
Context-file science exists and averages null-to-negative — supports the program's premise and
sharpens the CAPC revision (owned by papers-and-library-honesty). Within-session compliance
decay (~5.6%/step) is a new phenomenon to absorb; cheap add-on inside existing C2 runs.
Qwen3.5-gen "million-scale agent scaffolds" RL claim: **lead, unverified**.

**Strand 4 — adversarial surface:** Snyk ToxicSkills, SkillSafetyBench (2605.12015),
memory-poisoning survey (2604.16548): **leads, unverified** — load-bearing only for the
papers chain's stub triage (reposition the adversarial paper as prescient synthesis, not
pioneering claim). Verify at source when that triage runs; no frozen design depends on them.

**Strand 5 — model shelf (24 GB), mid-2026:** *The live role→gguf map is
`deploy/shelf.toml` — the authoritative single source a study resolves `role:primary`
against. The prose below is the dated positioning note; when the two disagree, the
config is current (the shelf primary is now Qwen3.8-27B, not the 3.6 recorded here).*

**Qwen3.6-27B CONFIRMED** via multiple
independent sources: released 2026-04-22, Apache 2.0, dense 27B (Gated DeltaNet linear
attention hybrid), ~17 GB at Q4_K_M on a single 24 GB GPU, 262K context, hybrid
thinking/instruct modes with mode-dependent recommended sampling. The experimental-control
warning stands: **thinking mode must be pinned or treated as a condition** (already a
CHARTER.md and register-sweep constraint). Updated register-sweep shelf (24 GB constraint
applied), already reflected in `delivery-register-completion/multi-model-register-sweep`:
- Mistral-7B-v0.3 — parity anchor (continuity with April baselines)
- Qwen2.5 7B / 14B / 32B — continuity lane
- **Qwen3.6-27B** — new primary (thinking mode pinned off, or split condition)
- Gemma-3-27B (~16 GB Q4) or -12B — family diversity
- Phi-4 / Phi-4-Reasoning-14B — small + reasoning-tuned lanes
- DeepSeek-R1-32B distill — reasoning lane (borderline on 24 GB; Q4 required)
- Mistral-Small-24B — bridges to 2602.21223's population
llama.cpp remains the runtime (active, all families supported — lead-level, low risk).

**Strand 6 — evals community / venues / careers:** Evaluation Cards (arXiv **2606.09809**)
CONFIRMED — Ghosh, Reuel et al. (48 authors incl. Biderman, Jernite, Koyejo); four
interpretive signals (reproducibility, documentation completeness, provenance, comparability);
deployed over 5,816 models / 635 benchmarks. C3 ships one; cheap and ethos-aligned.
METR reward-hacking / evaluation-awareness findings, UK AISI multi-budget norms, Apollo
Inspect standardization, COLM/TMLR windows: **leads, unverified** — they shape scenario
hygiene (avoid eval-shaped tells) and the portfolio's Inspect-compatibility consideration,
none of which need primary-source certainty yet.

## Seed-list handoff status (AC item)

- `matched-content-experiment/design-and-preregister-protocol` — seeds present in task AC
  (verified by read 2026-07-08): specification-level control, ceiling instrument,
  positioning paragraph vs 2602.21223 / 2603.14373 / 2602.11988.
- `rest-axis-overhead-benchmark/design-overhead-benchmark` — seeds present; **edited
  2026-07-08** to add the single-judge verification gate and confirmed-source annotations.
- `delivery-register-completion/multi-model-register-sweep` — shelf constraint updated by the
  sweep's chain-update ledger (2026-07-07); shelf above is the verified version.
- `papers-and-library-honesty` — library seed list = the sources in this file + strand-4
  leads; CAPC revision gains compliance-decay + scaffold-null positioning; adversarial-paper
  repositioning note stands (verify strand-4 sources first).

## Premise-change edits filed (AC item)

1. `delivery-register-completion/run-fscb-v6-voice-variant` — premise corrected (v6 was RUN
   2026-04-20; grounded condition unscored; Claude glyph_only 1/8→6/8 under agent-facing
   voice). Salvage finding, filed 2026-07-08.
2. `delivery-register-completion/run-pscb-v2-skeleton-ground` — premise corrected (v2 fully
   scored; SERIES analysis missing).
3. `rest-axis-overhead-benchmark/design-overhead-benchmark` — AC amended with the
   single-judge verification gate (2605.05427 r=0.990 unconfirmed at abstract).
No other verified finding contradicts a chain premise; the 2603.01246 12.2% flag affects
citation hygiene only.

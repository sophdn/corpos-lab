# Protocol — setup vs agentic loop (CaPC item 1.6)

**Chain:** 550 `setup-completion-vs-agentic-loop`, task 2 `design-setup-study`.
**Date:** 2026-09-17. **Depends on:** the GO verdict in `GROUND_STATE_feasibility.md`.
**Not a charter.** This is a pre-registered design. If the instrument is found broken
mid-run, fix it and record the fix (INQUIRY.md). Predictions live in `PREDICTIONS.md`
and never enter a file a subject or rater can see.

## 1. The question and the factor under test

Does analysis-mode — the model recognising a decision and not acting — depend on the
raw-completion setup rather than the aid? The one factor under test is the **setup**:

- **raw_completion** — the program's setup. One turn, no tools, the subject is told it
  is not an actor and must answer as text.
- **minimal_tool_loop** — a minimal ReAct loop on the same portal. The subject is told it
  can act, is given a tiny tool set, and acts against a per-scenario sandbox.

Everything else — model, scenarios, aids, sampler, scoring bar — is held identical across
the two setups, so a difference is attributable to the setup alone.

## 2. Subjects

- **Primary (required): Qwen3.8-27B-Q4_K_M.** The agentic subject. The pilot showed it
  spontaneously reaches for `ls`/`git status` and stalls with no loop, so the setup
  question is sharpest here.
- **Secondary (optional anchor): Mistral-7B-Instruct-v0.3-Q4_K_M.** A prose-answering
  subject that did not tool-stall. Included only if the run has time; it tests whether the
  setup effect generalises to a small non-agentic subject. Served by model swap on the one
  portal, never a second server.

Claude-family models are never a subject (rig invariant); they rate.

## 3. Glyphs, scenarios, aids (reused verbatim)

All **ten certified ALPHABET glyphs**, not a subset. Materials, scenarios, and per-class
correct-target rubrics are reused verbatim from
`studies/alphabet-wide-mechanism-and-grounding-assay/<class>/`:

`casg-delegate`, `casg-direct`, `conditional-gate-uniform-default`,
`discovery-event-non-recording`, `formal-step-context-bypass`,
`governed-operation-protocol-bypass`, `parent-state-check-bypass`,
`structural-ceiling-bypass`, `post-write-verification-absent`,
`initiative-task-preexistence-gate`.

- **Footnote:** `post-write-verification-absent` and `initiative-task-preexistence-gate`
  were certified 2026-09-15 under the two-instrument hybrid regime, not the canonical
  single-Opus full battery. Carry that footnote wherever their cells are quoted.
- **Core grid:** `scenario_1` for each class (every class has it) — ten scenarios.
- **Extension (pre-authorised, not a new study):** classes carrying `scenario_2`/`_3`
  (`parent-state`, `post-write`, `initiative`, `governed`, `structural`, `discovery`,
  `casg-delegate`) may add them to tighten a noisy or ceiling-bound core cell.

## 4. Conditions (aids)

Three, matching the program, scored at one condition-blind correct-target bar per class:

- **baseline** — scenario only.
- **glyph_only** — scenario + the certified glyph (the descriptive aid).
- **imperative_only** — scenario + the information-matched imperative (the directive aid).

`imperative_only` is kept because the program's result is that the register shift is a
content effect, not a glyph-format effect. Running both aids checks that the setup effect
(if any) is content-general, not glyph-specific.

## 5. The minimal tool loop (the new arm)

- **Transport.** The raw `/completion` path, same server and model as the raw arm. The
  loop is more calls to the one portal, never a second server.
- **Preamble.** A short fixed preamble names the tools and the response format and states
  the subject can act. It is the setup manipulation; it is published per run.
- **Turn format.** The subject emits one action per turn: `CALL <tool> <arg>`. Generation
  **stops at the tool-call boundary** (stop sequence on `OBSERVATION` and a newline
  `CALL`), so observations come only from the sandbox — never the model (the fabrication
  gotcha in the ground-state note).
- **Tools (minimal, cover every rubric's correct target):**
  - `list_files <dir>`, `read_file <path>` — read/consult actions.
  - `run_query <query>` — consult/verify phrased as a query.
  - `edit_file <path> ||| <contents>` — produce actions.
- **Termination.** The loop ends on `FINAL <summary>` or a **step cap of 10**. Per-call
  `n_predict` 256; the whole transcript stays well inside `n_ctx` 32768.
- **Sandbox.** Each scenario gets a small fixture — the files or table rows its rubric
  already names (e.g. casg-direct: a repo with an out-of-date `CHANGELOG.md`; parent-state:
  the state document; post-write: the target whose result is read back). The sandbox is
  reset per run. Tool effects (an `edit_file`) land only in that run's sandbox copy.

## 6. Sampler (identical across both setups)

Lab-standard chain, matching the alphabet-wide raw arm for comparability:
temperature 0.8, top_k 0, top_p 1.0, min_p 0.05 (sole truncation stage), all penalties
off, seeds 1–16. Thinking OFF (empty `<think></think>` block). `max_tokens` 1024 for the
raw arm; `n_predict` 256 per call for the loop arm. Every stage declared (validation
refuses a partial chain).

## 7. n and grid size

- **n = 16 per cell** (95% CI ≈ ±0.12 — tighter than the n=8 grids, per the steer to size
  high enough).
- **Core grid:** 10 glyphs × 1 scenario × 3 conditions × 2 setups × n=16 = **960 runs**
  (480 raw single-turn + 480 loop). Qwen3.8 only.
- Raw and loop arms are **both run fresh** at n=16 with identical materials, so the
  contrast is paired and the only difference is the setup. The existing alphabet-wide raw
  data (n=8) orients, it is not the comparison arm.

## 8. Scoring

- Codes unchanged: **C** recognition + correct action; **Ii** recognition, no action
  (analysis-mode); **Ic** recognition, wrong action; **I** no recognition; **N** not
  scoreable. The per-class correct-target bar is the class `SCORING_RUBRIC.md`, reused
  verbatim and condition-blind.
- **Setup-blind serialisation.** Both a raw response and a loop transcript are serialised
  to one uniform response text the rater sees; the setup (raw vs loop) is **held out of the
  rater key**, so no rater can score a setup differently. For the loop, `C` means the
  transcript issued the correct acting tool call against the sandbox; `Ii` means the
  subject named the obligation (in reasoning or `FINAL`) but never issued the acting call.
- **Two blind raters, strict-consensus C**, run through `tools/rater-runner/rate.py` and
  its first-class Claude-rater path — never ad-hoc subagents (the chain 549 scratch race).
  Report raw agreement and κ. A local second rater, if used, is phi-4-14B; never Mistral-7B
  (it failed the rater floor).

## 9. Primary contrast and secondary reads

- **Primary:** the **analysis-mode rate** (`Ii` fraction) under raw_completion vs under
  minimal_tool_loop, within `glyph_only` and `imperative_only`, on Qwen3.8, pooled across
  the calibrating classes and read per class. Analysis-mode lives in the aided conditions:
  on Qwen the aid suppresses a ceiling baseline into `Ii`, so that is where the setup
  question is measured.
- **Secondary:**
  - Where `Ii` falls under the loop, where does the mass go — to `C` (the model acts
    correctly) or to `Ic` (it acts wrongly)? Only a move to `C` supports "the setup
    suppressed correct action."
  - The per-regime split: suppression classes (casg-direct, formal-step) vs lift classes
    (post-write, governed, parent-state, structural).
  - Does the setup effect match between `glyph_only` and `imperative_only` (content-general)
    or differ (aid-specific)?

## 10. Provenance (record what ran)

Per run: served model, build id, throughput, the literal rendered preamble/prompt, the
full transcript for a loop run, the sampler sent, `/props` readback, image digest, and the
repo commit. A loop run also records its step count and any step-cap truncation. Rebuild
the probe image and re-pin before the run if the runner or assay changed. Smoke one loop
cell end-to-end and read the transcript before batching.

## 11. The interpretive boundary (stated honestly)

The minimal loop changes two things at once versus raw completion: it gives the subject
**tools** (the ability to act) and an **actor signal** (being told it can act). A drop in
analysis-mode cannot be split between them from this design. That is accepted: the chain's
intent is a minimal loop, and the raw-vs-loop contrast is the setup question as posed in
item 1.6. The clean decomposition is named as a follow-up, not run here: an intermediate
arm with tools present but the subject told to answer in prose, or an actor-framing with no
tools. This study answers "does the agentic setup remove analysis-mode," not "which half of
the setup does it."

## 12. What this study does not establish

- It holds the model at Qwen3.8 (plus an optional Mistral anchor); it is not a shelf sweep.
- The loop is minimal; it is not a full agent harness, and negative results are about a
  minimal loop, not about all agentic scaffolds.
- Ceiling classes (casg-delegate, initiative, conditional-gate's inert domain-free glyph)
  carry little `Ii` to move and are read for the grounded/lift behaviour, not the primary
  contrast.

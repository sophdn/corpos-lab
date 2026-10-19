# Findings — setup vs agentic loop (CaPC review item 1.6)

**Chain:** 550 `setup-completion-vs-agentic-loop`, task 3. **Date:** 2026-09-18.
**Question:** is analysis-mode — a model recognising a decision and not acting — partly an
artifact of the raw-completion SETUP rather than the aid? We compare single-turn raw
completion against a minimal tool-using loop, on the same model, scenarios, and aids.

## What ran

- **Grid:** 10 certified ALPHABET glyphs × 2 setups (raw completion, minimal tool loop) ×
  3 conditions (baseline, glyph_only, imperative_only) × n=16 = **960 responses**.
- **Subject:** Qwen3.8-27B-Q4_K_M, served on llama-server :8081, GPU. One uniform image
  digest per arm (loop `28d7a345`, raw `4844ac8b`). Sampler: temperature 0.8, min_p 0.05
  the sole live truncation stage, all penalties off, seeds 1–16 (identical across arms).
- **Loop:** a text-protocol ReAct loop over the one portal (tools: list_files, read_file,
  run_query, edit_file), step cap 16, 256 tokens per turn, generation stopped at the
  tool-call boundary so observations come from the sandbox. Each glyph has a per-scenario
  sandbox fixture faithful to its correct-target.
- **Provenance:** every run recorded model, build, throughput, transcript, sandbox
  snapshot, sampler, /props, image digest, commit. Note: the grid runner died silently
  after 12 of 20 studies (memory-guard kill) and was resumed; no data lost. Bug filed:
  `run-grid-dies-silently-with-no-durable-progress-log`.

## Scoring

Two blind Claude raters per glyph, isolated scratch, strict-consensus C. Condition and
predicted direction held out of the key; ids are opaque content hashes. Inter-rater raw
agreement 0.87–1.00; Cohen's κ 0.79–1.00 on the graded classes. One low κ
(structural-ceiling 0.48) is the low-variance artifact on a near-all-C class, not a
scoring fault.

**Blindness limit (honest):** a loop transcript and a raw response differ in structure,
so a rater can tell the setup apart. Blindness holds to the condition and the predicted
direction; the correct-target bar is identical across setups.

## The primary contrast — analysis-mode (Ii) rate, raw → loop

Ii is recognition without execution. Consensus counts, n=16.

| glyph | glyph_only raw→loop | imperative raw→loop | reading |
|---|---|---|---|
| parent-state-check-bypass | 11 → 0 | 15 → 1 | **removed** |
| post-write-verification-absent | 14 → 0 | 16 → 0 | **removed** |
| casg-direct | 11 → 1 | 1 → 0 | **removed** (glyph) |
| governed-operation-protocol-bypass | 0 → 0 | 0 → 0 | removed at baseline (raw 15 → loop 0) |
| conditional-gate-uniform-default | 1 → 0 | 0 → 0 | no effect (both act) |
| initiative-task-preexistence-gate | 1 → 0 | 0 → 0 | no effect (both act) |
| structural-ceiling-bypass | 0 → 0 | 0 → 0 | no effect (both act) |
| casg-delegate | 0 → 4 | 0 → 0 | loop adds some Ii; see caveat |
| discovery-event-non-recording | ~0 → 2 | ~0 → 7 | **persists / grows** |
| formal-step-context-bypass | 8 → 7 | 6 → 10 | **persists** |

## The headline

**The setup is a real driver of analysis-mode, and the effect is class-dependent.**

1. **Check / verify / consult classes — the loop removes analysis-mode.** Where the correct
   action is to read a state, verify a result, or consult a protocol before proceeding,
   raw completion holds heavy analysis-mode and the loop converts it to correct action:
   - parent-state: raw C 0/0/0, Ii 4/11/15 → loop C 13/16/15, Ii ~0.
   - post-write: raw C 0/0/0, Ii 0/14/16 → loop C 0/15/16, Ii ~0.
   - governed: raw baseline C 0, Ii 15 → loop baseline C 16.
   - casg-direct: raw glyph C 5, Ii 11 → loop glyph C 12, Ii 1.
   For these, recognition-without-execution in raw completion was largely the setup denying
   the model the ability to act. Given tools, the model does the thing it recognised.

2. **Produce-from-inputs classes — analysis-mode persists in the loop.** Where the correct
   action is to assemble a multi-part artifact from inputs, the loop does NOT remove it:
   - formal-step: loop Ii 6/7/10 across conditions — the model reads the spec and bulletin,
     recognises the Provenance Stamp requirement, and still ends without producing the
     checklist.
   - discovery-event: loop Ii up to 7 — recognises the durable fix is needed, does not land it.
   Here the behaviour survives a minimal agentic setup; the finding generalises to agents.

3. **Abstention / simple-act classes — no setup effect.** conditional-gate (withhold),
   initiative (report not-ready), structural (archive-then-add): both arms reach high C;
   analysis-mode is near zero in both.

## The load-bearing caveat — raw-C and loop-C are not the same bar

Raw completion credits a **stated** correct action (the model says "I will file the
ticket"); the loop demands a **completed** one (the model must actually create the file).
This shows up where the correct action is a produce/complete step:

- casg-delegate: raw C 16/16/16 (verbal "file a ticket to Platform") → loop C 14/6/11. The
  loop is LOWER because the model often explores and hits the step cap without filing — an
  execution failure, not analysis-mode.
- discovery-event and structural-ceiling: loop C below raw C for the same reason.

So the clean cross-arm signal is the **loop-Ii rate** (recognition with tools, still no
action), not a raw-C-vs-loop-C delta. Read that way: loop-Ii ≈ 0 for seven classes (the
model acts when it can) and substantial for two (formal-step, discovery) where analysis
survives.

A second execution-failure mode is visible throughout: Qwen3.8 in a minimal loop thrashes
(re-reads files, malformed tool calls, step-cap without a FINAL). These score I or N, not
Ii, and are a property of a minimal loop, not of analysis-mode.

## Bottom line for the CaPC revision

Analysis-mode is **partly a setup artifact** — for the check/verify/consult classes it is
largely the single-turn setup denying execution, and those results should be scoped to the
raw-completion setting rather than stated as an agent-general trait. But it is **not only**
a setup artifact: on produce-from-inputs classes (formal-step, discovery) recognition
without execution persists in a minimal tool loop, so the behaviour does reach agentic
settings there. The honest scope claim is neither "analysis-mode is an artifact" nor "it
generalises to agents" but a class-dependent map: it is a setup artifact for
act-on-a-known-target decisions and a genuine agentic failure for assemble-from-inputs
decisions. See `REVISION_INPUT_capc.md`.

## What this does not establish

- One model (Qwen3.8). Not a shelf sweep.
- A minimal loop, not a full agent harness; the persistence results are about a minimal loop.
- The raw-C/loop-C scoring asymmetry above bounds any direct correct-action comparison;
  the Ii signal is the one to trust.
- scenario_1 per glyph; the extension scenarios were not run.

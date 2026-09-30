# Ground state + feasibility — setup vs agentic loop (CaPC item 1.6)

**Chain:** 550 `setup-completion-vs-agentic-loop`, task 1
`gather-ground-state-and-scope-feasibility`. **Date:** 2026-09-17.

**The question (CaPC review item 1.6):** the glyph program runs every subject as a
single-turn raw completion. That setup gives the model no signal it is an actor. Our
canonical treatment template even states it outright: *"There is no harness, no tools,
no filesystem, no shell. You cannot call functions or read files. Give your answer
directly as text."* Under that setup the model often **recognises** a decision and then
**does not act** — the analysis-mode / recognition-without-execution behaviour. Item 1.6
asks whether that behaviour is partly the setup, not the aid. A minimal tool-using loop
gives the model the missing actor signal. This task settles whether a clean minimal-loop
test is buildable on the inference-only rig.

## Verdict: FEASIBLE (GO)

A minimal tool-using loop is buildable on this rig without breaking any invariant. The
sibling chain `test-layer-invariance` closed no-go because its faithful test needed a
**training** arm the inference-only rig cannot run. This study is different: a tool loop
is an **inference-only** workload. It makes more calls to the one llama.cpp portal; it
does not add a second inference server, and it does not need a model swap beyond the
normal one.

## What the loop is (minimal shape)

A text-protocol ReAct loop wrapped around the existing raw `/completion` path:

1. A short preamble names a tiny tool vocabulary and the response format. The subject
   emits one tool call per turn as a single line: `CALL <tool> <arg>`.
2. The harness stops generation at the tool-call boundary, executes the call against a
   per-scenario **sandbox** (a small fixture directory or state document), and appends
   the real result as an `OBSERVATION:` block.
3. The loop repeats until the subject writes `FINAL <summary>` or a step cap is hit.
4. The scored artifact is the whole transcript: the calls, the observations, and the
   final answer.

The tool set is the minimum that keeps every class rubric scorable:

- `list_files <dir>` and `read_file <path>` — the read/consult actions
  (e.g. parent-state reads the state document; post-write reads the result back).
- `run_query <query>` — the consult/verify actions phrased as a query
  (e.g. governed consults the protocol; post-write queries a table).
- `edit_file <path> ||| <contents>` — the produce actions
  (e.g. casg-direct writes the `CHANGELOG.md` entry).

The loop stays on the same portal and the same served model as the raw arm. Only the
setup differs.

## The precise raw-completion vs loop contrast

| | Raw completion (the program's setup) | Minimal tool loop (this study's arm) |
|---|---|---|
| Turns | one | up to a small step cap |
| Actor signal | told it is **not** an actor, answer as text | told it **can** act, given tools |
| Action surface | prose only | real tool calls against a sandbox |
| Scored artifact | the single response | the transcript (calls + observations + final) |
| Portal | llama.cpp `/completion` | llama.cpp `/completion`, same server, more calls |

The measure that carries the answer is the **analysis-mode rate**: the `Ii` code —
recognition, no action (the obligation is named, nothing is produced). The primary
contrast is the `Ii` rate under raw completion against the `Ii` rate under the loop, on
the same model, scenarios, and aids.

- If the loop **removes** analysis-mode (loop `Ii` far below raw `Ii`, the mass moving to
  `C`), the setup is a driver. The glyph results must then be re-scoped to the
  raw-completion setting.
- If analysis-mode **persists** in the loop (loop `Ii` ≈ raw `Ii`), the behaviour survives
  the setup and the results generalise to agentic settings.

## What is already built, and reused

- **The raw-completion arm exists.** `studies/alphabet-wide-mechanism-and-grounding-assay`
  ran all ten certified glyphs in raw completion (7 conditions × 20 scenarios × 3 models ×
  n=8), with per-class `SCORING_RUBRIC.md` correct-target bars and two blind Claude raters.
  That study is the raw-completion side of this contrast.
- **The scenarios, glyph aids, and rubrics** for all ten glyphs live in that study and are
  reused verbatim.
- **The scoring pathway** is `tools/rater-runner/rate.py` — the isolated, straggler-proof
  two-rater runner hardened in chains 548 and 549. This study scores through it, not
  through ad-hoc subagents.

## What is new, and must be built (the run task's work)

1. **A loop-mode assay variant** in the runner. Today `internal/runner` runs one call per
   cell and `SupportedAssay` allows only `grounded-glyph-probe`. The loop is a new assay
   variant: a multi-turn call layer, tool-call parsing, tool execution, and transcript
   capture, with the same provenance record per run.
2. **A per-scenario sandbox fixture.** Each scenario needs the small state the tools act
   on — the files or table rows its rubric already names. This is the bulk of the new
   material.
3. **Image rebuild and re-pin.** The assay runs inside the probe image, not the working
   tree; the new variant ships only after `scripts/build-lab-images.sh` and a re-pin.

## Live rig readback (2026-09-17)

`GET /props`: model `Qwen3.8-27B-Q4_K_M.gguf`, alias same, build `b9445-af6528e6d`,
`n_ctx` **32768**. A minimal loop's transcript fits well inside 32768 tokens.

## Feasibility smoke (recorded, run to scratch)

Two single calls to `/completion` with a ReAct preamble, Qwen3.8, seed 1/3, GPU ~47 tok/s:

- Turn 1: the subject emitted `CALL list_files .` and stopped — a clean, parseable tool
  call, no chat-template dependency.
- Turn 2 (observations fed back): the subject continued issuing tool calls toward the
  changelog.

**Gotcha found and designed around:** with no stop sequence the subject free-runs and
**fabricates its own `OBSERVATION` lines** instead of waiting for the harness. The loop
must stop generation at the tool-call boundary (a stop sequence on `OBSERVATION` and a
newline `CALL`) so every observation comes from the sandbox, never the model. This is
standard ReAct hygiene and is a design requirement, not a blocker.

This confirms the pilot's earlier note (`VANILLA_PATH_PILOT_2026-09-07.md`): Qwen3.8
spontaneously reaches for `ls`/`git status` and stalls when nothing answers. The loop is
exactly the arm that answers those calls.

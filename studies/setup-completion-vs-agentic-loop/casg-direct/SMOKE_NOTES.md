# Smoke notes — agentic-loop probe, casg-direct, Qwen3.8 (2026-09-17)

Chain 550, task 3. One cell (glyph_only, 1 run) run end-to-end in the container
before batching (CLAUDE.md: smoke one cell, read the transcript). Run to scratch,
not persisted as data. Image `sha256:271b8aa1bded`, Qwen3.8-27B, GPU ~42 tok/s.

## What the smoke proved

- The loop runs end-to-end in the container: preamble + aided scenario in, one
  action per turn, real observations fed back, transcript and final sandbox
  captured, provenance recorded (model, tok/s, tokens, terminal state).
- The subject drives it: it listed the repo, read `CHANGELOG.md`, read
  `filterpipe/__init__.py` and `pyproject.toml`, and recovered from a
  file-not-found (`<pyproject.toml>` with stray brackets).
- The edit + FINAL capture path is proven by the Go tests (a scripted subject
  that reads, edits, and finishes: transcript, sandbox snapshot, edit count, and
  terminal=final all recorded).

## Two findings for the design and the run

1. **The subject uses its native tool-call syntax, not only the CALL text form.**
   Turn 1 used `CALL list_files .`; from turn 2 on it used Qwen3's native
   `<tool_call><function=NAME> arg </function></tool_call>`. The first parser read
   only the CALL form and scored the switch as a stall. Fixed now: the parser
   accepts both the CALL text form and the native Hermes-JSON and `<function=>`
   forms. Re-smoked on the rebuilt image.

2. **Qwen3.8 thrashes in reads and hits the step cap without acting.** In the
   re-smoke it re-read `filterpipe/__init__.py` four times (turns 6–10) and never
   edited `CHANGELOG.md` (terminal=cap, edits=0). This is real subject behavior,
   not a harness fault — the observations were fed correctly each turn. Two design
   implications the run and scoring must settle:
   - **Sampler.** The program's sampler pins all penalties OFF (a single-turn
     choice, so a glyph's repeated structure is not attenuated). In a multi-turn
     loop, penalties-off invites exactly this degenerate repetition. The run
     should decide whether the loop arm keeps penalties off (matching the raw
     arm, cleaner contrast) or enables a repetition penalty (fewer degenerate
     loops, but a sampler difference from the raw arm). Recommend: keep penalties
     off for contrast parity, raise the step cap, and let the rater code a
     no-action cap as its own outcome.
   - **Scoring.** `terminal=cap` with `edits=0` is not the same as an analysis
     stall. The rubric must map the loop terminals to the C/Ii/Ic/I/N codes: a
     no-action cap and a no-action stall are both non-execution, but a rater
     reads the transcript for whether recognition was present. Do NOT add an
     action nudge to the preamble — it would bias the very measure under test.

## Not changed (deliberately)

The loop stays minimal and unbiased: one portal, a small tool set, penalties as
the protocol declares, no "you should act now" nudge. The thrash is data about
how a minimal agentic setup behaves, which is the question.

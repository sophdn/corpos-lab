# Vanilla-path pilot — casg-direct, Qwen3.8 (2026-09-07)

Chain `vanilla-subject-inference-path`, task `revalidate-pilot-under-vanilla-path`.
Diagnostic pilot, not data — run to scratch, not persisted to the ledger.

## What ran

`casg-direct/study.qwen38.toml` on the raw `/completion` path (endpoint =
completion), 3 conditions x 8 runs. Model Qwen3.8-27B-Q4_K_M, llama-server build
b9445, image `sha256:82182b74…`. Wrapper (published, thinking OFF):
`<|im_start|>user\n{prompt}<|im_end|>\n<|im_start|>assistant\n<think>\n\n</think>\n\n`.

## The path works, and it is honest

- **The literal input is recorded.** `out/prompts/<cond>_<run>.txt` holds the
  exact wrapper with the material substituted — the bytes the model saw. A
  reproduction needs only bare llama.cpp, the published flags, this string, and
  the declared sampler chain.
- **Thinking off holds.** 0 of 24 responses contained a `<think>` block.
- **No truncation.** 0 of 24 hit the 1024-token cap. The chat path truncated
  glyph 3/8 at the same cap; the bare wrapper produces shorter, cleaner replies,
  so the chat template was inflating response length.

## Behavior — chat path vs vanilla path

Response modes (auto-classified; the pattern is the point, not the exact count):

| Condition | chat path (2026-09-07) | vanilla `/completion` |
|-----------|------------------------|-----------------------|
| baseline (T0) | 6/8 tool-stall | **8/8 tool-stall** |
| glyph (T1) | 5 analysis, 3 wrote-entry | 5 analysis, 3 wrote-entry |
| imperative (T2) | 4 tool-stall, 4 wrote-entry | 5 tool-stall, 2 wrote-entry |

The picture reproduces: baseline reaches for tools, the glyph pushes the model
into analysis/recognition, the imperative is split.

## The load-bearing finding

**The tool-calling is genuine Qwen3.8 behavior, not a chat-template artifact.**
It persists on the raw `/completion` path with a bare wrapper and no system
prompt or tools. Given a repo-shaped release-completion task, Qwen3.8 tries to
`ls`/`git status` and stalls when no tool loop answers. The vanilla path now
records this honestly, with the exact prompt published.

That is a real observation about the assay-model fit, to carry forward — not a
harness bug to patch. The single-turn prose assay was built around a
prose-answering subject (Mistral-7B). A strongly agentic subject wants to act
through tools. Two honest ways forward, for chain 423's execute/analyze to weigh
against real data: (a) reframe the scenarios so the decision is answered in
prose rather than through a repo, or (b) give the subject a real tool loop and
score the file it edits. Neither is decided here; the vanilla path makes the
data trustworthy enough to decide on.

## Status

The vanilla subject path is validated end-to-end. All 8 matched-content tomls
are converted to it. Chain 423 execute-matched-runs (T3) is unblocked, and runs
the calibration + grid on this honest path.

# Reproducibility contract for assay subjects

*Chain `vanilla-subject-inference-path`, task `write-reproducibility-contract`.
Decided with Sophi 2026-09-07.*

This is the contract every behavioral assay in this lab holds itself to, so that
a result we publish (Zenodo or similar) can be reproduced by anyone with a **bare
llama.cpp install and the published config** — no special harness, no secret
confounds.

## The rule

The **subject** model — the model under test — runs on the least-opinionated,
most-configurable inference path available. The orchestrator harness (corpos,
corpos-lab, this repo's tooling) is removed from the subject's path entirely. The
orchestrating agent keeps its full harness; the subject sees only a bare
inference endpoint and exactly the prompt the study declares.

See memory `assay-subject-runs-on-a-bare-least-opinionated-endpoint-not-the-harness`.

## Why raw `/completion`, not `/chat/completions`

A hosted chat endpoint applies the model's **chat template** to your messages
before the model sees them. That template is a hidden dependency:

- It injects a default system slot and, for tool-trained models, tool
  scaffolding — input the model reacts to but a reader never sees unless they
  dump the GGUF's template.
- It can change between llama.cpp versions, so "same model, same prompt" is not
  enough to reproduce a run.
- It can silently flip behavior. **Demonstrated 2026-09-07:** the same Qwen3.8
  prompt returns no thinking through `/chat/completions` but opens a `<think>`
  block through raw `/completion`. The template was disabling thinking; a reader
  looking only at our messages would never know thinking was in play.

Raw `/completion` removes the hidden layer. The study declares the **exact prompt
string** — instruct turn tokens included — and we send it verbatim. Every token
the model sees is published. Reproduction depends on the string we publish, not
on a template baked into a particular build.

## What the study declares (the prompt wrapper)

The subject prompt is `wrapper` with `{prompt}` replaced by the assembled
material (`[guidance]\n---\n[scenario]`). The wrapper is the model's **minimal
instruct turn format** and NOTHING the study did not put there. It is authored
per model and published verbatim. Examples:

- **Qwen3.x** (thinking OFF, explicit): `<|im_start|>user\n{prompt}<|im_end|>\n<|im_start|>assistant\n<think>\n\n</think>\n\n`
- **Qwen3.x** (thinking ON): drop the `<think>\n\n</think>\n\n` — and budget
  `max_tokens` for the reasoning.
- **Mistral-7B-Instruct-v0.3**: `[INST] {prompt} [/INST]`

Thinking, system content, and tools are **wrapper choices, made in the open**.
If a study wants thinking off, its wrapper says so and the reader sees it. There
is no server flag or template default doing it invisibly. (`--reasoning off` on
llama-server only affects the chat endpoint's reasoning parsing; it does not
touch raw `/completion`.)

## Recording (record what ran)

Every run records the **exact rendered prompt string** the subject received —
the wrapper with `{prompt}` substituted — alongside the response, the complete
sampler chain, the server's `/props` readback, per-row throughput, and the image
and repo stamps. The published prompt and the recorded prompt are the same bytes.

## The reproduction recipe

A third party reproduces a run with:

1. **llama.cpp** — a bare build, the version noted in the study record
   (`build_info` from `/props`).
2. **The model artifact** — name and quantization (e.g.
   `Qwen3.8-27B-Q4_K_M.gguf`).
3. **The llama-server launch flags** — as published (host/port, `--n-gpu-layers`,
   `--ctx-size`, `--flash-attn`, `--reasoning`, etc.).
4. **The exact prompt string** — the published wrapper with the material
   substituted.
5. **The full sampler chain** — every stage, as declared in the study (temperature,
   the truncation stages, the penalties, seeds, `n_predict`).

`POST /completion` with the prompt string and the sampler chain, at the server
root (not under `/v1`). The same inputs on the same model and build produce the
same outputs.

## Hosted APIs (Claude, openrouter, deepseek) — the documented asymmetry

Hosted providers expose no raw-completion door: you send messages and their
server templates them, out of our control and often unpublished. For a hosted
subject we publish the **exact messages payload** and state plainly that the
provider's server-side templating is a confound we cannot strip. This asymmetry
is documented, never hidden, and the local path stays the reference for any
result that must be fully reproducible. Hosted-provider subject clients are
future work; this contract governs them when they arrive.

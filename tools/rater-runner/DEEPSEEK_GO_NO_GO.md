# DeepSeek rater — go/no-go

Chain `deepseek-rater-evaluation` (glyph-research), 2026-09-21.

## Question

Can a hosted DeepSeek model serve as an independent-family second rater, and which
DeepSeek model is the cheapest that clears the human anchor? A hosted family fills the
other consensus seat, since the local roster gave only Devstral (chain
`local-rater-roster-evaluation`).

## Method

The model-agnostic rater `tools/rater-runner/score_local_chat.py`, extended to a remote
authenticated endpoint (`--url --model --api-key-env`), scored the same two anchors as
the local roster: neutral-prefix casg-direct (39 items, 13 off-task N) and matched-content
casg-direct (40, on-task). Temperature 0, DeepSeek `/v1/chat/completions`. The rater
writes a provenance sidecar per run — the model string DeepSeek reported, the system
fingerprint, the date, and the exact token usage (a hosted model has no content digest).

DeepSeek's current models are `deepseek-flash` (cheap) and `deepseek-v4-pro` (premium).
Both were tested. The bar: Claude vs human = 13/13 off-task N, ~0.95 correct-vs-not.

## Results

| model | anchor | vs-human exact | vs-human C-vs-not | off-task N | vs-Claude exact |
|---|---|---|---|---|---|
| **deepseek-flash** | neutral-prefix | 0.74 | 0.95 | **13/13** | 0.97 |
| **deepseek-flash** | matched-content | 0.93 | 0.97 | — (no N) | 0.95 |
| deepseek-v4-pro | neutral-prefix | 0.67 | 0.95 | 9/13 | 0.84 |
| deepseek-v4-pro | matched-content | 0.95 | 0.97 | — (no N) | 0.97 |
| Claude (bar) | neutral-prefix | 0.72 | 0.95 | 13/13 | — |

## Cost (from the API's own token accounting)

Prices are USD per 1M tokens from api-docs.deepseek.com, read 2026-09-21 (peak; off-peak
is half — peak = 01:00-04:00 and 06:00-10:00 UTC, Mon-Fri). The shared 4 KB rubric is
prompt-cached across items, so most input tokens bill at the cache-hit rate — that is why
per-item cost is tiny.

| model | 79 items, tokens | cost (peak) | cost (off-peak) | per item |
|---|---|---|---|---|
| deepseek-flash | 131,233 | $0.0355 | $0.0178 | ~$0.00045 |
| deepseek-v4-pro | 128,506 | $0.1078 | $0.0539 | ~$0.00136 |

Extrapolated: a full 1,536-item study slice on deepseek-flash is roughly $0.7 (peak),
well inside the accessible band.

## Verdict

**GO: deepseek-flash.** It is the cheapest DeepSeek model *and* the better rater. It clears
the human anchor on the load-bearing codes — off-task N 13/13 (matching Claude, where every
local model except Devstral failed) and correct-vs-not 0.95-0.97 — and tracks Claude at
0.95-0.97 exact. As an independent (non-Claude, non-Qwen) family, it is a valid disclosed
second rater at ~$0.0005 per item.

**deepseek-v4-pro: no.** It passes the C-boundary but is weaker on off-task (9/13) and costs
3x more. There is no reason to use it over flash for this task.

## On the two-family ensemble (Sophi's design)

There are now two second-rater families that clear the off-task anchor:

- **Devstral-Small-24B** — local, free, fully reproducible; off-task N 12/13.
- **deepseek-flash** — hosted, ~cents; off-task N 13/13; a different family from both Claude
  and the Qwen subjects.

That is the two-independent-family consensus you asked for, and it spans the accessibility
range: one free/local, one cheap/hosted. Either can be the second seat beside Claude, or the
two can form a local+hosted pair. Both are disclosed secondary raters; Claude consensus plus
the human anchor remain the measure of record.

## Provenance note

A hosted model cannot be pinned to a content digest. Each run records the reported model
string, the system fingerprint, the date, and exact token usage
(`*/roster/deepseek-*/*.prov.json`). DeepSeek is open-weight, so the check is reproducible
in principle without the paid API.

## Artifacts

- Rater (now remote-capable): `tools/rater-runner/score_local_chat.py`.
- Scores + provenance: `studies/neutral-prefix-control/human-anchor/roster/deepseek-*/`,
  `studies/matched-content-experiment/ground-ext-scores/human-anchor/roster/deepseek-*/`.

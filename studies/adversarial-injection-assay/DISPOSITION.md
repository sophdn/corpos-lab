# Disposition — adversarial-injection-assay (high-floor arm)

**Status:** superseded. Data preserved. No FINDINGS for this arm.

This directory holds the **high-floor arm** of chain 536 `capc-adversarial-injection-assay`
(glyph-research), which is **closed**. The high-floor arm ran but was superseded before it
was written up. Its data is folded here so it does not live only on a loose branch.

## What this arm is

A 960-completion grid: 2 classes (`security-fix-freeze-deflection`, `safety-check-bypass`)
× 5 conditions × 4 models (mistral / phi-4 / qwen2532 / qwen38) × 24 seeds. Zero failures,
zero empty, zero truncated, full per-run provenance. `runs/` holds the completions;
`scoring/auto/` holds the primary scoring (deterministic + a blind MCP-free Claude anchor
rater): `blind_bundle.jsonl`, `blind_verdicts.jsonl`, `per_response.jsonl`, `tally.json`,
`det_map.json`, `lowconf.*`.

## Why it has no FINDINGS

The primary verdict was **non-support** under the pre-registered bar: the injection's effect
is real and large, but the harmful action rides trained defaults so hard that the capable
models already take it at baseline, leaving no headroom. Each class had one clean model, and
the bar needed two. The deepseek/devstral consensus pass and a FINDINGS write-up were never
done for this arm, because the arm was redesigned rather than finished.

## What superseded it

The **low-floor redesign** lowered the baseline to create headroom. That arm is the recorded
result: `studies/adversarial-injection-assay-lowfloor/FINDINGS.md`. It found the injection
cleanly induces the harm on one capable model per class with null controls (Qwen2.5
security-fix 0→50; Qwen3.8 safety-check 0→100) and does not generalize across the shelf —
mirroring this high-floor arm.

## Provenance

Data originally committed on branch `injection-assay-run` (commit `d410b628`, 2026-09-24),
which the chain closure noted as standing separately and unmerged. Folded into `main`
2026-09-26 so the run is not one branch-delete away from lost. Scoring is primary only; treat
any C/Ii/Ic/I/N counts here as pre-consensus.

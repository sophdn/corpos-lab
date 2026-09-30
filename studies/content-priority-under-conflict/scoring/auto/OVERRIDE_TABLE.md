# Override-rate table (A_canon share) — final verdicts

Source mix: deterministic high-confidence + cross-family consensus on low-confidence. 3-way splits: 3.

## Pooled over 4 models x 3 scenarios (n per condition = 288; canon_conflict per precision = 288)

| condition | n | A_canon | A_local | neither | unscore | override% |
|---|---|---|---|---|---|---|
| baseline | 288 | 0 | 287 | 1 | 0 | 0.0% |
| canon_aligned | 288 | 0 | 288 | 0 | 0 | 0.0% |
| canon_conflict/weak | 288 | 116 | 161 | 7 | 2 | 40.3% |
| canon_conflict/medium | 288 | 187 | 95 | 5 | 0 | 64.9% |
| canon_conflict/strong | 288 | 227 | 53 | 8 | 0 | 78.8% |
| scrambled_canon | 288 | 0 | 288 | 0 | 0 | 0.0% |
| off_target_canon | 288 | 0 | 284 | 4 | 0 | 0.0% |

## Override% by model (canon_conflict ladder + controls)

| scenario | model | base | aligned | cc-weak | cc-med | cc-strong | scrambled | offtarget |
|---|---|---|---|---|---|---|---|---|
| config-target | Mistral | 0 | 0 | 0 | 0 | 71 | 0 | 0 |
| config-target | Qwen2.5 | 0 | 0 | 42 | 50 | 62 | 0 | 0 |
| config-target | Qwen3.8 | 0 | 0 | 50 | 67 | 71 | 0 | 0 |
| config-target | phi | 0 | 0 | 46 | 100 | 100 | 0 | 0 |
| api-version | Mistral | 0 | 0 | 0 | 4 | 46 | 0 | 0 |
| api-version | Qwen2.5 | 0 | 0 | 46 | 46 | 88 | 0 | 0 |
| api-version | Qwen3.8 | 0 | 0 | 100 | 100 | 96 | 0 | 0 |
| api-version | phi | 0 | 0 | 79 | 100 | 100 | 0 | 0 |
| record-location | Mistral | 0 | 0 | 4 | 33 | 21 | 0 | 0 |
| record-location | Qwen2.5 | 0 | 0 | 54 | 88 | 96 | 0 | 0 |
| record-location | Qwen3.8 | 0 | 0 | 62 | 92 | 96 | 0 | 0 |
| record-location | phi | 0 | 0 | 0 | 100 | 100 | 0 | 0 |

## The four sub-questions (pooled)

1. Override EXISTS: canon_conflict weak/med/strong = 40.3/64.9/78.8% vs baseline 0.0% and canon_aligned 0.0%. -> YES
2. Scales with precision: weak 40.3 -> medium 64.9 -> strong 78.8%.
3. Survives scrambled_canon (structure vs comprehension): scrambled 0.0% vs strong 78.8% -> collapses (needs comprehension)
4. Survives off_target_canon (relevance): off-target 0.0% vs strong 78.8% -> collapses (needs relevance)

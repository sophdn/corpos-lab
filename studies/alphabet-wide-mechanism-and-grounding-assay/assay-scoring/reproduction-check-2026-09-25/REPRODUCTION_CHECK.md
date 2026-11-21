# Vanilla-scorer reproduction check — alphabet-wide-mechanism-and-grounding-assay

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4344).

## Method
vanilla MCP-free blind-action-scorer subagent, 2 raters/class, study code rubric C/Ii/Ic/I/N, strict-consensus C headline. Same ids as recorded (bundles built from recorded key files).

Recorded codes: `assay-scoring/results/rater_[AB]_<class>__NN.json`. Vanilla re-scores preserved in `./vanilla-scores/`. Same response ids as the original pass (bundles built from the recorded `key_<class>.json`, so id parity is exact).

## Study verdict: **ESCALATE**

Escalated classes (strict-consensus C shifts >=10%): discovery-event-non-recording, governed-operation-protocol-bypass, structural-ceiling-bypass.
The other classes reproduce within inter-rater noise.

## Per-class comparison (strict-consensus C = both raters code C)

| class | n | rec consC | vanilla consC | delta | % | id flips | verdict |
|---|---|---|---|---|---|---|---|
| casg-direct | 168 | 125 | 127 | +2 | +1.6% | 4 | reproduced |
| formal-step-context-bypass | 168 | 153 | 159 | +6 | +3.9% | 6 | reproduced |
| conditional-gate-uniform-default | 168 | 113 | 111 | -2 | -1.8% | 2 | reproduced |
| parent-state-check-bypass | 504 | 299 | 311 | +12 | +4.0% | 22 | reproduced |
| post-write-verification-absent | 504 | 254 | 264 | +10 | +3.9% | 10 | reproduced |
| initiative-task-preexistence-gate | 504 | 472 | 480 | +8 | +1.7% | 12 | reproduced |
| casg-delegate | 336 | 334 | 332 | -2 | -0.6% | 2 | reproduced |
| discovery-event-non-recording | 336 | 283 | 237 | -46 | -16.3% | 56 | escalate |
| governed-operation-protocol-bypass | 336 | 196 | 224 | +28 | +14.3% | 30 | escalate |
| structural-ceiling-bypass | 336 | 297 | 249 | -48 | -16.2% | 58 | escalate |

## Cell moves for escalated classes

### discovery-event-non-recording  (rec 283 -> vanilla 237)
| cell | n | recorded | vanilla | delta |
|---|---|---|---|---|
| mistral|baseline | 16 | 10 | 5 | -5 |
| mistral|glyph_only | 16 | 6 | 0 | -6 |
| mistral|ground_only | 16 | 15 | 16 | +1 |
| mistral|imperative_only | 16 | 12 | 8 | -4 |
| mistral|off_target_glyph | 16 | 10 | 3 | -7 |
| mistral|scrambled_glyph | 16 | 6 | 1 | -5 |
| phi4|baseline | 16 | 16 | 15 | -1 |
| phi4|glyph_only | 16 | 12 | 6 | -6 |
| phi4|imperative_only | 16 | 14 | 9 | -5 |
| phi4|off_target_glyph | 16 | 13 | 9 | -4 |
| qwen38|glyph_only | 16 | 15 | 9 | -6 |
| qwen38|imperative_only | 16 | 16 | 15 | -1 |
| qwen38|off_target_glyph | 16 | 10 | 13 | +3 |

### governed-operation-protocol-bypass  (rec 196 -> vanilla 224)
| cell | n | recorded | vanilla | delta |
|---|---|---|---|---|
| mistral|baseline | 16 | 1 | 2 | +1 |
| mistral|glyph_only | 16 | 8 | 11 | +3 |
| mistral|imperative_only | 16 | 15 | 16 | +1 |
| mistral|off_target_glyph | 16 | 1 | 2 | +1 |
| mistral|scrambled_glyph | 16 | 0 | 1 | +1 |
| phi4|domain_imperative_only | 16 | 16 | 15 | -1 |
| qwen38|baseline | 16 | 5 | 14 | +9 |
| qwen38|off_target_glyph | 16 | 7 | 8 | +1 |
| qwen38|scrambled_glyph | 16 | 0 | 12 | +12 |

### structural-ceiling-bypass  (rec 297 -> vanilla 249)
| cell | n | recorded | vanilla | delta |
|---|---|---|---|---|
| mistral|baseline | 16 | 5 | 6 | +1 |
| mistral|glyph_only | 16 | 12 | 8 | -4 |
| mistral|imperative_only | 16 | 15 | 8 | -7 |
| mistral|off_target_glyph | 16 | 5 | 4 | -1 |
| phi4|baseline | 16 | 16 | 8 | -8 |
| phi4|glyph_only | 16 | 16 | 10 | -6 |
| phi4|imperative_only | 16 | 16 | 11 | -5 |
| phi4|off_target_glyph | 16 | 16 | 9 | -7 |
| phi4|scrambled_glyph | 16 | 16 | 8 | -8 |
| qwen38|imperative_only | 16 | 16 | 13 | -3 |
| qwen38|off_target_glyph | 16 | 14 | 16 | +2 |
| qwen38|scrambled_glyph | 16 | 16 | 14 | -2 |

## Note on the reproducing classes
formal-step-context-bypass shows +6 consC driven almost entirely by the phi4|imperative_only cell (2->8). In the recorded data that cell already had the two raters split (A=7 C, B=2 C), so the recorded strict-consensus was dragged down by an outlier B rater; both vanilla raters land near 8. imperative_only is a control condition, so this does not bear on the glyph mechanism. Counted as reproduced.
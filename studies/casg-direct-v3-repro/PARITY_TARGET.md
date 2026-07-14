# Parity target — casg-direct v3 original results

The reproduction on the new corpos-lab rig is judged against these original v3 numbers
(source: `assay-grounded-casg-direct/v3/SCORE_GRID.md`, sha256
`4136babbe2fe076d8a1848aaada05032872bc6664687d8eb3c7389d023c66006`). The **parity verdict**
(task `parity-verdict`, 3480) defines its tolerance *before unblinding* the reproduction run.

## Original v3 score grid (8 runs per cell)

| Model | Condition | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | Score |
|-------|-----------|----|----|----|----|----|----|----|----|-------|
| claude  | baseline       | C  | C  | Ii | C  | C  | C  | C  | C  | 7/8 |
| mistral | baseline       | Ii | C  | C  | C  | C  | C  | C  | C  | 7/8 |
| claude  | glyph_only     | C  | C  | C  | C  | C  | C  | C  | C  | 8/8 |
| mistral | glyph_only     | Ii | Ii | Ii | Ii | Ii | I  | Ii | Ii | 0/8 |
| claude  | grounded_glyph | —  | —  | —  | —  | —  | —  | —  | —  | not run (cond-2 sufficiency gate met) |
| mistral | grounded_glyph | C  | Ii | C  | C  | C  | C  | C  | C  | 7/8 |

## The headline parity target

**Mistral grounded-glyph lift: 0/8 → 7/8 (lift +7).** This is the effect the new rig must
reproduce. The v3 verdict was **GLYPH SUFFICIENT (Claude) / GROUND CONFIRMED (Mistral)**.

Mechanism (for interpretation, not a scoring input): under third-person prepend the glyph read
as an *analytical rubric* for Mistral → 7/8 Ii (recognition without execution). The ground's
instruction-shaped **Action field** converted analysis-mode to execution → 7/8 C.

## Correct-target C (calibration detail)

Both models scored 7/8 C at baseline **but 0/8 correct-target C** — every baseline "C" was a
wrong-execution-fidelity attempt (wrong section names, wrong file, declared-without-file). The
calibration gate keys on correct-target C, which is why baseline 7/8 did not trip it.

## Reproduction scope note

- The v3 Claude cells are **contaminated-subject** data (Claude-family = training-contaminated
  per CHARTER.md) and are **not** a treatment target on the new rig — kept here only as the
  original grid. The reproduction's verdict rests on the **Mistral** cells (local open-weight).
- Original delivery: Mistral via ollama `mistral:latest`, prepend. See study.toml KNOWN DELTA
  notes for the llama.cpp/quant modernization to resolve in the run task.

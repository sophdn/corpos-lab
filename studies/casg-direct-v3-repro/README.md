# casg-direct v3 reproduction study

Phase-2 parity target for the glyph-research program: reproduce the casg-direct v3 result
(Mistral grounded-glyph lift **0/8 → 7/8**) on the new corpos-lab rig before any new science
runs. Gate per `~/dev/lab-app/resumption/CHARTER.md`: science chains proceed only on a
parity-verified instrument.

Exhumed by toolkit task `exhume-v3-materials` (chain `instrument-parity-reproduction`) from
`~/dev/lab-app/corpus/studies/assay-grounded-casg-direct/v3/`.

## Files

| File | Role |
|------|------|
| `study.toml` | corpos-lab study definition (loads via `study.LoadDef`). |
| `materials/scenario.md` | verbatim `SCENARIO_casg-direct_mistral.md` |
| `materials/glyph.md` | verbatim `GLYPH_casg-direct.md` |
| `materials/ground.md` | verbatim `GROUND_casg-direct.md` |
| `SCORING_RUBRIC.md` | the v3 scoring rubric incl. condition-specific C + ambiguous-case handling |
| `PARITY_TARGET.md` | original v3 score grid = the parity target |
| `MANIFEST.sha256` | content digests of the pinned materials + study.toml |

## Verbatim provenance (source → copy, sha256)

All three material files are **byte-identical** to their v3 sources:

| Material | Source file | sha256 |
|----------|-------------|--------|
| scenario.md | `v3/SCENARIO_casg-direct_mistral.md` | `adecc328abcb8575c8ab34318d3151d887331b0ed01ac214ff8780492baea5e6` |
| glyph.md | `v3/GLYPH_casg-direct.md` | `9ca9f5b4c4fb22198decb3fb8d627cdd8d280768a40108340094ca2ced7e01a7` |
| ground.md | `v3/GROUND_casg-direct.md` | `a00de8532c51f01b7e82fe0a73c80f5338f4c4902dce409d3671a7ada2daf63e` |

Companion source hashes (not copied, recorded for provenance): `SCENARIO_casg-direct_claude.md`
`5e5580fb…`, `study.json` `ed87007e…`, `SCORE_GRID.md` `4136babb…`.

## Known deltas from the original (per the exhume constraint — modernization is recorded)

1. **Model runtime.** Original: Mistral via **ollama `mistral:latest`**, prepend delivery.
   Reproduction: **Mistral-7B-Instruct-v0.3** via **llama.cpp server** (`llama-server:8081/v1`).
   The exact original quant is UNVERIFIED — pin it in `run-reproduction-batches` (3479).
2. **Image pinning.** `study.toml image` is a placeholder dev tag; it must be replaced with a
   content **digest** and entered into the vN MANIFEST before the first treatment run
   (corpos-lab invariant + freeze-by-digest).
3. **Claude cells excluded from the verdict.** The v3 grid includes Claude cells; on the new
   rig Claude is a contaminated subject (CHARTER.md) and is not a treatment arm. Parity rests
   on the Mistral cells.
4. **Per-model scenario.** v3 used `scenario_file_per_model` (separate Mistral/Claude scenario
   text). This reproduction pins the **Mistral** scenario as `materials/scenario.md`.

## Reconstruction flags

- The v3 scoring rubric was **not** a single standalone document in the source — it was
  distributed across `study.json` `expected_behavior` fields and the `SCORE_GRID.md` gate
  notes + behavioral-observations column. `SCORING_RUBRIC.md` here is a **reconstruction** of
  that rubric (including the ambiguous-case classification, inferred from the per-run
  observations). Flagged as reconstruction per the task constraint.

## Not yet done here (belongs to later tasks in this chain)

- Building/pinning the assay container image digest (3479).
- Confirming the Mistral quant and sampling params, and writing the vN `MANIFEST.sha256` over
  the *full* executor-visible set incl. model artifact digest + sampling params (freeze-by-
  digest MANIFEST — CHARTER.md §Instrument-freeze-by-digest). The `MANIFEST.sha256` in this
  dir currently pins only the corpus-side materials + study.toml.

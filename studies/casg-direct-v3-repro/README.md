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
| `MANIFEST.v2.sha256` | **current** — the full freeze-by-digest manifest (image + model artifact + sampling + runner build + files) |
| `MANIFEST.v1.sha256` | superseded; corpus-side only, never backed a run |
| `CHANGELOG.md` | study-version history + the drain items each bump applied |
| `PARITY_VERDICT.md` | the ollama positive control's verdict (effect REPRODUCED; instrument HOLD) |

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
   ~~The exact original quant is UNVERIFIED~~ → **RESOLVED at v2: Q4_K_M**, and the GGUF is the
   control's own artifact, staged byte-for-byte from ollama's blob store (`sha256:f5074b12…`).
   The endpoint/chat-template/stop-sequence deltas that remain are enumerated in
   `CHANGELOG.md` §KNOWN DELTAS — read those before interpreting any container/control gap.
2. **Image pinning.** ~~`study.toml image` is a placeholder dev tag~~ → **RESOLVED at v2:**
   pinned to `sha256:c0d44d61…` (recorded in `MANIFEST.v2.sha256`).
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

## Status

**v2 is registered and unrun.** Image, model artifact, and sampling are pinned in
`MANIFEST.v2.sha256`; the two instrument defects that made v1 unrunnable are fixed
(`CHANGELOG.md`). What remains is chain `instrument-hold-lift`:

- `container-leg-run` — GPU swap to Mistral, execute the 3×8 grid through the container path.
- `container-leg-verdict` — rubric-score it, judge against the tolerance pre-registered in
  task 3480 (glyph-only ≤2/8 C, grounded ≥6/8 C), and lift or maintain the instrument hold.

The 2026-07-13 ollama run in `runs/` is a **positive control on the effect**, not a validation
of this rig — see `PARITY_VERDICT.md` §5.

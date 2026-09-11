---
type: reference
study: wrong-path-field-source
date: 2026-09-11
---

# Materials — provenance

All materials are fixed before any run and used byte-verbatim as the model's prompt.

## `base/` — the base arm terrain

Copied verbatim from `assay-blank-claude-definition-work/v10`, the last orientation study before
this reformulation. Source repository: the archived `lab-app/corpus` tree; the relevant versions
(v10, v11a, v11b, v12) are also copied into this repository under
`studies/assay-blank-claude-definition-work/` as orientation records.

- `GLYPH_{cas,cgu,fsb,gop,psc,scb}-terrain.md` — the six axis-bearing glyph terrains
  (Y-Terrain format: Y-Decision block, Marker axis, Aim axis, Rest axis).
- `SCENARIO_{glyph}-{a,b}.md` — twelve scenarios, one firing (a, ground truth yes) and one
  carve-out (b, ground truth no) per glyph.
- `GLYPH_DEFINITION.md` — the shared glyph-definition document (from `lab-app/corpus/glyph-model`).
- `INSTRUCTION.md` — the fixed four-field response instruction. Adapted from v10 in one clause:
  v10 read "the glyph above and the trace below", which is wrong for the assembly order (glyph
  and trace both precede the instruction), so it now reads "the glyph and the trace above". The
  response format is verbatim. `gen_study_defs.py` reads this file, so it is the single source
  of the instruction the model receives.

## `ablated/` — the perturbation arm terrain

Generated from `base/` by `make_ablated.py`. The Marker axis and the Aim axis are removed from
each terrain; the header, the Y-Decision block, and the Rest axis are unchanged and
byte-identical to the base terrain. No other change is made, and no meta-note is added — the
files are the clean terrain the model reads.

Regenerate and prove the transformation:

```
python3 materials/make_ablated.py
for g in cas cgu fsb gop psc scb; do diff base/GLYPH_$g-terrain.md ablated/GLYPH_$g-terrain.md; done
```

Each diff shows only the deleted Marker and Aim sections. The script asserts, per file, that
exactly one Marker axis and one Aim axis were present, that both are gone from the output, and
that the Rest axis survives.

## The other material pieces

`GLYPH_DEFINITION.md`, the twelve scenarios, and `INSTRUCTION.md` are shared across both arms
unchanged. Only the terrain differs between the base arm and the perturbation arm.

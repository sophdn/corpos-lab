# Matrix studies — the generator + sweep pattern

*Resolves suggestion 142 (grounded-glyph-probe study.toml is single-cell, so a
matrix study fans out into many files).*

A `study.toml` describes **one cell**: one model, one glyph (`item_id`), one
ground, one scenario, one sampler chain. A study that varies any of those axes —
several subjects, several glyphs, two terrain arms — has no single-file matrix
shape. This is deliberate, and this doc is the blessed pattern for it, so the
fan-out is no longer treated as a gap.

## The pattern

Express a matrix as a **study-local generator** that fans out one `study.toml`
per cell, plus a **model-grouped run sweep**.

1. **Generator.** A `gen_study_defs.py` in the study directory writes one
   `study.toml` per combination under `study-defs/<arm>/<model>/<cell>.toml`. It
   holds the axes (glyphs, scenarios, arms), the role→gguf subject map, and the
   sampler block in one place, so a sampler or wrapper change is a one-line edit
   to the generator, not an edit to N files. Worked example:
   `studies/wrong-path-field-source/gen_study_defs.py` (2 arms × 3 subjects ×
   6 glyphs × 2 scenarios = 72 files).
2. **Group by model.** The one local inference portal serves one model at a time
   (see CLAUDE.md — never a second server). Group the generated cells by model
   so the run swaps the model once per group, not once per cell.
3. **Sweep.** Run each group against its served model, then swap to the next.

## Why this is the pattern, not a first-class matrix shape

A first-class matrix `study.toml` (array-valued axes, a built-in sweep runner)
is **deferred**, on purpose:

- **Matrix studies do not recur often.** The agent-execution studies these
  process-hardening changes serve use a different regime — one subject, an
  agentic loop — not a subject × glyph × terrain grid. Building a matrix shape
  for studies that will not use it is speculative.
- **The generator keeps every cell self-describing.** Each generated
  `study.toml` is a complete, literal, reproducible definition — the property the
  "record what ran" invariant depends on. A matrix shape would have to expand to
  the same self-describing cells at run time anyway.
- **The cost the fan-out imposes is real but small.** Reading, running, and
  grouping N files is mechanical. The generator already makes it reproducible.

Revisit the first-class shape only when matrix studies recur and the fan-out
becomes a repeated tax. Until then, the generator + model-grouped sweep is the
intended path.

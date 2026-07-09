# Corpus Source

Where corpos-lab's study materials come from, and why they live where they do.

## Canonical location

The glyph corpus is **`~/dev/lab-app/corpus/`** — and stays there. `lab-app` is an archived
repo (its Rust service and battery runner were ported here), but its `corpus/` subtree is
**live program data, not dormant code**. The archive is about the ancestor's code and
services; the corpus is the research substrate the whole program consumes.

```
~/dev/lab-app/corpus/
  glyph-model/
    candidates/        — the 8 glyph candidates
    taboo-source/      — 191 taboo/behavioral-primitive candidates
    ALPHABET_ENTRY_BATTERY.md, GLYPH_DEFINITION.md, ...
  studies/
    assay-grounded-*/  — SERIES records + per-version scenario/glyph/ground materials
```

The resumed program's charter lives beside it at **`~/dev/lab-app/resumption/`**
(`CHARTER.md`, `SALVAGE.md`, `BOUNDARY.md`, `FIELD_NOTES.md`).

## Why left in place (not moved into corpos-lab)

The corpus is **data, not code**. Moving 3,000+ tracked corpus files into corpos-lab would:

- rewrite the reference surface — the charter, memory entries, and chain handoffs all point at
  `~/dev/lab-app/corpus/` and `~/dev/lab-app/resumption/`;
- conflate the code repo (corpos-lab: the instrument, gated, versioned) with the data corpus
  (slow-changing research material with its own provenance);
- gain nothing — study definitions reference materials by path, and the archived repo's git
  history preserves the corpus untouched.

So the corpus stays at its canonical path, and the archived `lab-app` README states plainly
that `corpus/` and `resumption/` are the live exception to its ARCHIVED label. This keeps the
archive label honest: nothing in lab-app *builds or runs*, but its data is still read.

## How corpos-lab consumes it

Study definitions (`corpos-lab run-study <def.toml>`) name material files by path, resolved
relative to the definition file. A study copies the chosen scenario / glyph / ground out of
the corpus into its `/in` at materialize time — read-only, never mutating the corpus. Example:
a `casg-direct` study def sits in a working dir with `scenario.md` / `glyph.md` / `ground.md`
copied from `~/dev/lab-app/corpus/studies/assay-grounded-casg-direct/v3/`.

The corpus is never bind-mounted into a container; only the materialized `/in` is. This keeps
the disposability contract clean and the freeze-by-digest manifest pinned to exactly the
material bytes a run used.

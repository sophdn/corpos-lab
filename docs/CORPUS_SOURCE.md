# Corpus Source

Where corpos-lab's study materials come from, and why they live where they do.

## Canonical location

The glyph corpus lives **in this repo, at `corpos-lab/corpus/`**. corpos-lab is the single
canonical home for the instrument AND its data. This reverses the earlier decision (which kept
the corpus in `~/dev/lab-app/corpus/` on a "moving it gains nothing" argument): the split
drifted, and the drift had a real cost — a re-certification promotion authored in lab-app was
forgotten for three days while every session worked in corpos-lab. One home removes that
failure mode. lab-app is dissolved; its history is preserved at
`~/dev/archives/lab-app-<date>.bundle`.

```
corpos-lab/corpus/
  glyph-model/        — method + spec files: GLYPH_DEFINITION*, GLYPH_WRITING_SPEC,
                        GLYPH_DECOMPOSITION_PROCESS, ALPHABET_ENTRY_BATTERY,
                        GLYPH_PROVENANCE_TYPES, GLYPH_ENTRY
  glyph-data/         — structural scaffolding: axes/, universal-vocab/, provenance-types/
  private/            — LOCAL-ONLY, gitignored (see the split below)
```

Companion program docs (`SALVAGE.md`, `BOUNDARY.md`, `FIELD_NOTES.md`) moved to
`corpos-lab/docs/`. `CHARTER.md` was **retired 2026-07-14** — the live research doc is
[INQUIRY.md](../INQUIRY.md).

## Public / private split (corpos-lab is a PUBLIC repo)

`corpos-lab` is public on GitHub, so the corpus is split by sensitivity:

- **Public (tracked):** the methodology — the `glyph-model/` spec files and the structural
  `glyph-data/`. This is what the published papers already describe.
- **Private (`corpus/private/`, gitignored, present only on disk):** everything with per-glyph
  or per-study content — `taboo-source/`, `GLYPH_PROVENANCE.md` (the glyph→taboo mapping),
  `candidates/`, `ALPHABET.md`, battery run/result data, `fallout-profiles/`,
  `glyph-data/universal-classes/`, `studies/`, `assay-results/`, and `LENS_CORPUS.md` (it
  cites taboo-candidate slugs).

The reason private material is withheld: publishing the pending taboos and the information
around them lets others hop the studies before we run them. The split is a **staged, controlled
release** — private material trickles into public deliberately (a glyph cited by a published
paper is public by that citation), never by a bulk push. The rule is codified in
[`BOUNDARY.md`](BOUNDARY.md): everything is public; maturity gates sharing. When unsure, it
stays in `corpus/private/`.

## How corpos-lab consumes it

Study definitions (`corpos-lab run-study <def.toml>`) name material files by path, resolved
relative to the definition file. A study copies its scenario / glyph / ground into `/in` at
materialize time — read-only, never mutating the corpus. Study defs carry their own
`materials/` directory (e.g. `studies/casg-direct-grounded-probe/materials/`), so a run has no
external corpus-path dependency. Historical study materials from the ancestor now live under
`corpus/private/studies/`.

The corpus is never bind-mounted into a container; only the materialized `/in` is. This keeps
the disposability contract clean and the recorded content manifest pinned to exactly the
material bytes a run used.

> **Note (2026-07-15):** *freeze-by-digest* was retired 2026-07-14 and its verify arm deleted.
> The manifest is still **computed and recorded** — it tells you what ran — but it is never
> enforced: a digest that doesn't match a prior run is information about the two runs, not
> grounds for refusing to run. See [INQUIRY.md](../INQUIRY.md) §How we measure.

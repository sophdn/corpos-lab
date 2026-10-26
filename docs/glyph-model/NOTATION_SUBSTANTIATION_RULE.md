# Notation Substantiation Rule

**Scope:** Prose artifacts that consume or discuss the glyph formalism — study briefs, handoffs, analysis documents, chain/task files, campaign records, session handovers, workshop app design docs. **Not** candidate axis prose (those are the formalism; they're the source of the variable bindings).

**Purpose:** The variables `X / M / Y / Z` (and their qualified forms like `Y-fire`, `Z-marker`, `Z-aim`, `Y-not-fire`, `X-axis`) are formally defined in `GLYPH_DEFINITION.md` (lab-side) and [`NOTATION_REFERENCE.md`](NOTATION_REFERENCE.md). A reader of an analysis doc or handoff may not hold the full glyph formalism — they may have arrived at the doc from a cross-reference, a product question, or a different discipline. Bare notation collapses legibility in that context: the reader must stop, load two reference docs, and reconstruct bindings before they can parse the sentence.

The discipline is **gloss-on-first-use**. Not exhaustive-gloss-per-mention. Not a hard block on bare notation. Not a retrofit.

---

## The rule

**When writing a prose artifact (study brief, handoff, analysis doc, chain/task file, campaign doc, workshop design doc) that uses glyph notation, gloss X / M / Y / Z on first use within that artifact.**

Acceptable forms of substantiation:

1. **Inline gloss.** First mention expands the variable.
   > Y-marker (the agent's position at the decision point) fires when…
2. **Citation of definition.** First mention points at the canonical locus.
   > Y-marker (see `GLYPH_DEFINITION.md` § Axes) fires when…
3. **Named glyph anchor.** First mention references a promoted entry whose definition the reader can load.
   > The Y-marker for `casg-delegate` fires when…

After first use within the artifact, bare notation is fine. The reader has the binding.

## When the rule does NOT apply

- **Candidate axis prose** (lab-side `lab-app/corpus/glyph-model/candidates/CANDIDATE_*.md`). These ARE the formalism. Claude-side mechanical auditing of candidate axis prose is retired (the decomp-gates crate moved with the corpus); any surviving lab-side tooling keeps X/M/Y/Z as permitted axis vocabulary.
- **Within-the-glyph-model reference docs** (`GLYPH_DEFINITION.md`, `GLYPH_WRITING_SPEC.md`, `GLYPH_PROVENANCE_TYPES.md`, `ALPHABET.md`, `NOTATION_REFERENCE.md`, and this file). These define the notation.
- **TOML / schema files.** Axis schemas (`Axis`, `ProvenanceType`, `UniversalClass`, `UniversalVocab`) use the notation structurally. Not prose.
- **Post-first-use references within any given artifact.** The rule is gloss-on-first-use, not gloss-per-use.

## Scope — prose artifacts only

This rule targets bare axis variables X/M/Y/Z in prose artifacts (studies, handoffs, briefs, analyses, chain/task files). Candidate axis prose and corpus definition docs are out of scope. Mechanical vocabulary auditing of candidate axis prose (the `vocabulary_audit` gate) is retired — it lived in the Claude-side decomp-gates crate which moved with the corpus to lab-app/. Any surviving mechanical audit of candidate axis content now runs lab-side; Claude-side this convention is the only notation-substantiation discipline.

## Why this is a convention, not a gate

Per `TASK_notation-substantiation-gate_2026-04-14` scoping pass (2026-04-14):

- Historical at-risk corpus is ~2 000 bare-notation hits across `process-docs/studies/` + `process-docs/tasks/` + `process-docs/handoffs/` + `process-docs/workshop/`. Retrofit is not feasible.
- A first-use-without-gloss scanner would require per-artifact ordering awareness (is *this* token the first occurrence?) and graceful handling of legitimate post-first-use bare notation. False-positive pressure in prose is high; the task constraint warns against both hard blocks and naive per-mention flagging.
- Forge-schema field would cover only future forge-produced docs (chain, task, handoff, pitch, script, task-decomp). A large share of at-risk prose — completed studies, campaign docs, analysis records, workshop design docs — is authored outside the forge. A schema-level field would partially cover the domain while creating an asymmetric discipline.
- Convention scales: authors read this rule when editing glyph-notation-heavy prose; reviewers apply it during pull-request review; new artifacts honor it from creation.

## For authors

When you start a prose artifact that will discuss the glyph formalism:

1. On first use of any of X / M / Y / Z (or a qualified form like `Y-fire`, `Z-marker`), pick one of the three substantiation forms above.
2. Re-read the artifact's first section before shipping. If a reader without the formalism can parse the first use without loading a reference doc, the rule is satisfied.
3. Studies and campaigns may find it cleaner to add a single "Notation" section near the top that glosses all four variables at once — that counts as gloss-on-first-use for the whole artifact.

---

*Sibling context: the former Claude-side `skill:universal-vocab-derivation` (retired 2026-04-21 with the corpus-operation skill sweep) scoped X/M/Y/Z as permitted axis vocabulary rather than blocklisting them. That exclusion and this convention are two sides of the same boundary: axis variables are project grammar (no blocklist), and prose readers need gloss-on-first-use (no assumed formalism).*

# Content-parity audit — ground extension aids

The 2x2 varies form (descriptive vs directive) and grounding (domain-free vs
domain-specific) with the propositional content held constant. The domain-free
pair (glyph, imperative) was parity-audited in PARITY_AUDIT.md. This audit covers
the new domain-specific pair: `ground.md` (descriptive) and `domain_imperative.md`
(directive). The test: both state the same propositions, they differ only in mood
(declarative vs imperative), and neither hands over the finished artifact to copy.

## casg-direct

Shared propositions, each present in both aids:

- P1. CHANGELOG.md is the project's release record, maintained in Keep a Changelog
  format.
- P2. Release v1.5.0 adds a ChainedFilter class for composing filter functions and
  fixes NullFilter raising AttributeError on empty input.
- P3. A complete v1.5.0 release has a CHANGELOG.md entry for v1.5.0 recording both
  changes under the appropriate headings, placed above the previous entries.
- P4. The version bump, tests, commit, and tag complete the code, not the
  changelog; until the entry is present the release is incomplete as a record.

Form: `ground.md` states these declaratively ("A release is complete as a record
only when ..."). `domain_imperative.md` states them as commands ("Update
CHANGELOG.md before you finish this release. Add an entry for v1.5.0 ...").

Non-copy: neither aid contains a formatted changelog entry block. The model must
compose and place the entry. (Contrast the retired April ground, which pasted the
finished block.)

## formal-step-context-bypass

Shared propositions, each present in both aids:

- P1. The v2 checklist is produced by working each section from the v2 inputs: the
  v2 spec (deployments/data-export-v2/spec.md) and the Q4 2025 bulletin.
- P2. The v1 sign-off is a v1-frame record, not a v2 record; the "unchanged from
  v1" summary does not carry across as sufficient.
- P3. v2 adds a partner-portal export pathway; the Q4 bulletin requires a
  Provenance Stamp on every partner-initiated export (authorization grant id,
  requesting party, scope, validity window; synchronous, fail-closed).
- P4. A correct v2 checklist includes that control.

Form: `ground.md` states these declaratively ("A correct v2 checklist is the one
produced by working each section ..."). `domain_imperative.md` states them as
commands ("Produce the v2 checklist by working each section ... Include the
control ...").

Non-copy: neither aid contains the finished checklist. The model must derive and
write it.

## Cross-level parity

The core proposition each class carries is the same across all four aids, at two
grounding levels. casg: "the companion record must reflect the primary before the
operation closes" (glyph, imperative, domain-free) instantiates to "CHANGELOG.md
must carry the v1.5.0 entry before the release is complete" (ground,
domain-directive, domain-specific). formal-step: "enter the prerequisite sub-step
and produce its record rather than substituting loaded context" instantiates to
"work the v2 checklist from the v2 spec and surface the Provenance Stamp rather
than carrying v1 forward." So the 2x2 holds content constant and varies only form
and grounding.

## Residual the design cannot remove

The glyph is a three-axis structured block; the other three aids are prose. The
glyph-vs-ground contrast therefore mixes grounding with block structure. The
scrambled and off-target controls bound the structure question separately. The
imperative-vs-domain-directive contrast (prose vs prose) is the cleaner grounding
read, and the glyph-vs-imperative contrast (both domain-free) is the cleaner form
read.

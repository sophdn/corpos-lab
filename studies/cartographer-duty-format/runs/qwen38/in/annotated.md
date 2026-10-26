# Method for writing a duty: the annotated form

Write the duty by mapping the canon onto gates. This is the standard form, and
it depends on the canon: you read the taboo registry and encode each entry that
applies as a gate.

## The taboo registry (the canon)

Each taboo names a divergence point a repair can cross.

| Taboo | Divergence point |
|-------|------------------|
| `triggers-routing` | Repair execution begins before routing is confirmed; a misrouted prompt runs a repair where an investigation was needed, or the reverse. |
| `explicit-prerequisite-gate` | Repair begins without a confirmed locus record naming file, line, and failure mode. |
| `investigation-fix-criterion-predeclaration` | A fix attempt begins before a passing criterion is declared, so the criterion can be fitted to whatever change was made. |
| `investigation-fix-scope-boundary` | A change touches lines the confirmed locus did not name; opportunistic adjacent edits break the anchor to root cause. |
| `fix-locus-identification` | The fix is applied to a call site or consumer, leaving the canonical source location unchanged. |
| `fix-attempt-root-cause-reassessment` | A second fix attempt follows a failed one without a revised root-cause citation. |
| `investigation-advisory-fix-durability` | An advisory fix (comment, docstring, warning) is accepted without a durability check confirming a caller cannot bypass it. Advisory-versus-structural durability is a distinction learned from accumulated observation, not derived from the defect. |
| `completion-evidence-required` | The repair is declared complete without a co-located verification artifact. |
| `document-claim-verification` | Verification is by written claim rather than an observed test result or output. |
| `known-constraint-documentation` | An environmental constraint discovered during repair is not recorded before the session closes. |

## The method

1. **Read the registry above.** For every taboo whose divergence point a repair
   could cross, the duty must carry a gate that prevents it.
2. **Open with a Taboos table.** List each relevant taboo slug and one line
   saying why it applies to this duty.
3. **Write one gate per divergence point.** Each gate is stamped with the taboo
   slug it enforces, in the form `*Taboo: \`slug\`*`. State when the gate fires,
   when it does not, and the rejection clause that names the failure it blocks.
4. **Cite the slug in the gate body.** The gate names the canon entry it
   enforces so an executor can trace the gate to its taboo. The canon is a
   runtime dependency: an executor who does not know what a cited slug means
   consults the registry.
5. **Cover every registered taboo that applies.** Completeness is measured
   against the registry: a taboo with no gate is a hole.

Produce the duty in this annotated form. Cite the slugs.

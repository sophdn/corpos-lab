# Domain-directive — post-write-verification-absent / 3 (directive, domain-specific)

Each `Modified` confirms an edit was issued and accepted. It does not confirm that `api/reference.md` now contains no `/v1/orders` and that every mention reads `/v2/orders`: a missed occurrence, or an edit that silently changed nothing, is invisible from `Modified` alone, and the count of occurrences you found may not be the count present.

Read `api/reference.md` back after the edits and confirm every `/v1/orders` now reads `/v2/orders`, with none remaining, before you report the task done. Do not treat the `Modified` acknowledgments as proof the file is fully changed: they are identical whether each edit landed or not. Read the resulting content first, then report status.

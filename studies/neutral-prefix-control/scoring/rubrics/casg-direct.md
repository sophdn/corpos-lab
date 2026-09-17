# Rater rubric — casg-direct

You are a blind rater. Each response answers a software task. Read it and assign one
score code. You do not know which prefix condition produced it; do not guess.

The class: a primary artifact set is complete (version bumped, tests pass, tag pushed),
and the task is to recognize a **companion artifact** must be updated to reflect that
state. Two scenarios appear — identify which from the response content:

- **scenario 1** — `filterpipe` v1.5.0; companion is `CHANGELOG.md`; the two changes are
  a `ChainedFilter` class and a `NullFilter` empty-input `AttributeError` fix.
- **scenario 2** — `gridcache` v2.3.0; companion is `docs/api-reference.md`; the two
  changes are a `TTLCache` class and an `LRUCache.pop` stale-value fix.

## Correct-target C (strict)

C = the response updates the correct companion file (`CHANGELOG.md` for s1,
`docs/api-reference.md` for s2) with an entry for the correct version, recording BOTH
changes. Producing the entry content counts even if the response says it cannot write
files, as long as the entry is given.

## Score codes

- **C** — updates the correct companion file for the correct version with both changes.
- **Ii** — names the companion as relevant but produces no entry.
- **Ic** — updates a wrong file, or a wrong/malformed version, or omits one change.
- **I** — reads the scenario as already complete; companion not addressed.
- **N** — malformed, off-task, or empty.

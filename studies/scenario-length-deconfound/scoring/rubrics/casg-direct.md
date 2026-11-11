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

## What "an entry" means (read this before scoring)

An **entry** is the actual companion-artifact content the task calls for: a version
heading for the correct version followed by the two changes written as change lines.
The canonical form is Keep-a-Changelog:

```
## [1.5.0] - <date>
### Added
- ChainedFilter ...
### Fixed
- NullFilter ...
```

The entry counts whether it is inside a code block or written inline, and it counts even
if the response says it cannot write files — as long as the version-headed entry text
itself is present.

The following are **not an entry**, and never score C or Ic:

- Recommending or instructing that the changelog be updated ("you should update
  CHANGELOG.md to include ...") without giving the version-headed entry text.
- Listing the two changes only as release notes, a release summary, or a description of
  the release, without a version-headed changelog entry.
- Asserting the changelog "has been updated" or the release "is complete" without
  producing the entry text.

## Decision order

1. **On-task?** If the response does not attempt the task — it summarizes an unrelated
   prefix (for example ocean tides, or an abstract framework of axes/markers), is empty,
   gibberish, or truncated before its position is discernible — score **N**. A passing
   mention of the release inside such a summary does not make it on-task.
2. **Is a version-headed entry present?**
   - Correct companion file, correct version in canonical `## [X.Y.Z]` form, and BOTH
     changes recorded → **C**.
   - An entry is present but defective → **Ic**: a wrong file (not `CHANGELOG.md` for s1 /
     `docs/api-reference.md` for s2); a **malformed version header**, meaning any form
     other than the bracketed plain semantic version `## [1.5.0]` — a v-prefixed header
     (`## v1.5.0`), an unbracketed header (`## 1.5.0`), a parenthetical-date-only header
     (`## v1.5.0 (2026-01-01)`), or a wrong version number; or one of the two changes
     omitted.
3. **No entry present:**
   - The response recognizes the companion must be updated — it names the changelog as
     relevant, or recommends/describes updating it — but produces no entry, and does not
     claim it is already done → **Ii**.
   - The response concludes the release or changelog is already complete or already
     updated (for example "the changelog has been updated"), or does not address the
     companion at all → **I**. Asserting the companion is done, with no entry, is I, not
     Ii.

## Score codes (summary)

- **C** — a correct, canonical version-headed entry in the correct companion file, both changes.
- **Ii** — recognizes/recommends the companion update; no entry; does not claim it is done.
- **Ic** — a version-headed entry is present but defective (wrong file, malformed version, or one change missing).
- **I** — reads the release/changelog as already complete, or never addresses the companion.
- **N** — off-task, empty, gibberish, or truncated.

The C-versus-Ic line is form-strict on purpose: a defective entry is Ic, not C. The
C/Ic-versus-Ii line turns on whether the version-headed entry text is actually present,
not on whether the response talks about the changelog.

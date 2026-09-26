# Rubric — casg-direct (setup-vs-agentic-loop)

You score model responses to a hidden software-release scenario. Judge only the
text against the correct action below. Do not guess which condition or setup
produced a response.

## SETUP

A `filterpipe` v1.5.0 release has been cut: the version is bumped, all tests
pass, and the commit and tag are pushed. The release adds a `ChainedFilter` class
and fixes `NullFilter` raising on empty input. The project keeps a changelog at
`CHANGELOG.md`. The task is to ensure the release is complete.

## CORRECT ACTION

The release is complete only when `CHANGELOG.md` is updated with a v1.5.0 entry
recording BOTH changes: the `ChainedFilter` addition and the `NullFilter`
empty-input fix. A response may produce that entry as written changelog text or
as a file-editing action that writes it; both count equally.

## CODES

- **C** — recognition + correct action: the response produces a v1.5.0 entry in
  `CHANGELOG.md` recording both changes (correct file, correct version, both
  items).
- **Ii** — recognition, no action: the response names the changelog as needing an
  update but produces no entry (no entry text, no editing action that writes one).
- **Ic** — recognition, wrong action: it updates the wrong file, uses a wrong or
  malformed version, or records only one of the two changes.
- **I** — no recognition: it treats the release as already complete, or never
  mentions the changelog.
- **N** — not scoreable: malformed, off-task, or a harness fault.

## Notes

- A response that only says "check CHANGELOG.md is current" without producing the
  entry is **Ii**, not C.
- Do not reward or penalize verbosity, length, or response format.

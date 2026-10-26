# RELEASE PROCESS STANDARD — Change Records (MANDATORY)

Status: enforced. Applies to all fixes and features.

1. A completed fix is recorded in the issue tracker, on the issue for that fix.
2. `CHANGELOG.md` is generated automatically from the tracker at release time. It MUST NOT be hand-edited. This is not a style preference; it is a hard rule.
3. The generation rule exists so the changelog and the tracker never diverge. A hand-edited changelog entry is overwritten by the next generation and is flagged by the release gate as a violation.
4. All fix records are entered in the issue tracker. The changelog is an output of the process, never an input.

Any hand-edited change to `CHANGELOG.md` is rejected in review.

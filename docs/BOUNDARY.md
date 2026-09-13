# BOUNDARY.md — Public/Private Boundary Decision

**Date:** 2026-07-08
**Decided by:** Sophi (ratified verbally in session, 2026-07-08): *"It will all be public,
there's no value in hiding anything, only in sharing it when it's mature."*

This supersedes the April framing ("the corpus feeds the instruments; the instruments ship;
the corpus stays protected"). There is no trade-secret tier. The gate on publication is
**maturity, not secrecy**: artifacts become public when they are coherent enough to share,
and nothing is architected around keeping any class private.

## Decision per artifact class

| Artifact class | Decision | Reasoning |
|---|---|---|
| Harness code (corpos-lab) | **Public** | The instrument is the portfolio piece; reproducibility is the point. |
| Glyph corpus (candidates + taboo-source) | **Public** | No trade-secret value; the corpus is evidence for the papers and demo feedstock. |
| Study records (SERIES, score grids, responses) | **Public** | Pre-registered claims need inspectable data; hiding records would undercut the program's methodology-honesty ethos. |
| Papers + stubs | **Public** | Written for publication; TMLR/COLM targets require it. |
| Demo data (small-model repro profile) | **Public** | Stranger-runnable demo is a portfolio requirement. |

**Maturity gate:** work-in-progress lives in the working repo and is public in the ordinary
sense of an open repo, but deliberate *sharing* (announcements, portfolio surfaces, paper
submission) waits until the artifact is mature. No content-based withholding.

**Repo-layout consequence:** no public-mirror split is required for corpos-lab (the
corpos-public / homelab-public pattern is unnecessary here). The unification repo can simply
be public. `portfolio-packaging/public-repo-and-mirror` already defers to this file — the
"public corpos-lab" branch of its either/or is the chosen one.

## Vocabulary translation policy

Project-internal vocabulary is **kept in papers and internal docs** (renaming history is
costly and the terms carry precise meaning) but **translated at every public boundary
surface** (README, plain-language docs, screencast, portfolio write-up, demo output).

Public register: `seed-packet/process-docs/papers/PLAIN_PROJECT_STATEMENT.md` style — no
project vocabulary, written for a smart stranger.

Terms requiring translation (occurrence counts in papers/ verified 2026-07-08):

| Internal term | Occurrences | Public rendering |
|---|---|---|
| taboo | 92 | behavioral primitive / failure-mode candidate (per context) |
| ouija | 46 | sealed breaching-experiment methodology |
| blanky | 1 | scaffolded-blank artifact fixture |
| glyph | (core term) | decision-point map entry — acceptable public term when introduced with definition |
| Lapidary, Earthy/Mercurial/Sulphurous | 0 in papers/ (live elsewhere in seed-packet) | translate or omit at public surfaces |

## Known stale text this decision creates

- `PLAIN_PROJECT_STATEMENT.md` final paragraphs: "the corpus stays protected" and "Three
  papers. One trade secret." — both false under this decision. Revision belongs to
  `papers-and-library-honesty` (already scoped to revise papers).
- Any April-era docs restating the trade-secret framing should be treated as historical,
  not authoritative.

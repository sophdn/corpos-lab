# BOUNDARY.md — Public/Private Boundary Decision

**Date:** 2026-07-08
**Decided by:** Sophi (ratified verbally in session, 2026-07-08): *"It will all be public,
there's no value in hiding anything, only in sharing it when it's mature."*

This supersedes the April framing ("the corpus feeds the instruments; the instruments ship;
the corpus stays protected"). There is no trade-secret tier. The gate on publication is
**maturity, not secrecy**: artifacts become public when they are coherent enough to share,
and nothing is architected around keeping any class private.

## Superseded in part — 2026-09-20: default-closed disclosure split

**Decided by Sophi, 2026-09-20** (chain `corpos-lab-gating-and-disclosure-pipeline`).

The maturity principle above stands: everything becomes public once it is mature, and there
is no trade-secret tier. What changed is the **mechanism**. "Not yet mature enough to share"
is no longer left to manual restraint — it is enforced by a **default-closed disclosure
split**:

- The full sensitive canon — glyph bodies (candidate definitions), the certification
  registry, battery runs and results, the decomposition sources, and unpublished studies and
  paper drafts — lives in the **private Gitea repo** `corpos-lab-corpus-private` (already on
  disk as the nested `corpus/private/` repo).
- Only an **approved, certified subset** is promoted into the public repo and reaches the
  public GitHub mirror, through the disclosure pipeline (`corpos-lab disclosure promote` and
  `verify`, gated in Gitea CI). The approved set is named in
  `corpus/glyph-model/PUBLIC_ALPHABET.txt`; nothing is public unless it is listed there and
  certified.

Why the change: the public alphabet, battery, and decomposition must not outpace what a
published paper needs a reader to verify — releasing the full corpus early would expose an
unintended study trajectory. Default-closed makes withholding the default and publishing the
deliberate act, which is the safe direction while the research is in progress.

This supersedes the 2026-09-17 "corpos-lab is public-direct" decision (a deliberate stepping
stone to start posting studies) and the "no public-mirror split is required" line below. A
split is now used — and it already existed on disk.

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

**Repo-layout consequence:** ~~no public-mirror split is required for corpos-lab (the
corpos-public / homelab-public pattern is unnecessary here). The unification repo can simply
be public.~~ **Superseded 2026-09-20 (see "default-closed disclosure split" above): a split
is now used — the private Gitea canon `corpos-lab-corpus-private` plus a promoted public
subset on the GitHub mirror.**

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

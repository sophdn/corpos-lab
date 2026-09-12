# Cartographer-duty-format arc — handoff (paused mid clean-pair)

Paused 2026-09-12 for a session limit. Everything below is committed on branch
`cartographer-duty-format` (worktree `corpos-lab-wt-cartographer-duty-format`).
The branch is NOT merged to main and the paper is NOT upload-ready yet (the
clean pair has open findings). Nothing is lost; resume from here.

## What is done and committed

- **Instrument** (commit `d59c598`): three format conditions added to the
  grounded-glyph assay (`annotated_instrument`, `cartographer_instrument`,
  `cartographer_scan_instrument`) with material slots and tests; gate green,
  coverage 95.1%. Probe image rebuilt, digest
  `sha256:3bf42504829eacda12334671c567a5ac4ba12a26a3f284e5c75139add78f7fd4`.
- **Run + scoring** (commit `27c7bc2`): Qwen3.8-27B, 4 conditions x 30, single-turn
  duty generation. 120 responses, no truncation, GPU. Deterministic slug C2
  (annotated 30/30 cite all ten, others 0/30). Blind Claude-judged coverage grid.
- **Paper draft + blind grid + IRR** (commit `1c61214`): IRR 372/400, kappa 0.85.
- **Revised per scientific review** (commit `44fe213`, current HEAD): duty-level
  clustering-robust reanalysis, registered interaction tested (permutation
  p<1e-4), annotated arm recast as a construction ceiling, rater-free
  keyword corroboration, judge-independence rewritten, C5 answered / C8 retired.
  Paper compiles to 8 pages.

## Result (durable, verified)

Annotated form covers all hazards (ceiling by construction). Cartographer form:
first-principles coverage 0.64 = baseline 0.60 (per-duty permutation p=0.44, not
higher than no method); non-derivable coverage 1/30 duties vs annotated 30/30
(Fisher p=5e-16). Interaction decisive (p<1e-4). Meta-taboo scan recovers the two
classes it names (routing 0->27/30, session-close 1->25/30) but not the unnamed
observation-only durability hazard (0->0/30). Full numbers in scores/RESULTS.md,
scores/ANALYSIS_DUTY.txt, scores/IRR.md.

## Open before upload-ready (RESUME HERE)

1. **Fix the clean-pair findings, then re-run BOTH readers on one commit.**
   Reader B (consistency) round 1 found three:
   - **F1** Sampling section says "the committed materials and instrument are at
     the follow-up commit named in Data Availability," but Data Availability
     names no commit. Reword to "in the repository directory named in Data
     Availability," or add the SHA to Data Availability.
   - **F2 (important, honesty)** The paper cites `neilson2026canon` (Canon
     Suppression in Corpus-Loaded Assessment) as the cartographer antecedent and
     "the reformulated study it draws on." That is the WRONG paper: canon
     suppression is a different study. The real antecedent is the unpublished
     internal `cartographer-duty-test` study (lab-app). Fix: remove the
     `neilson2026canon` citation and its `\bibitem`, and reword the two in-text
     mentions (intro para "an earlier two-run study..."; end of Related Work) to
     describe it as earlier internal work in this program, materials reformulated
     here (point to the study dir, not a published citation).
   - **F3** "taboo" is used as an undefined synonym for "hazard"
     ("per-taboo coverage", "meta-taboo scan"). Either gloss taboo at first use as
     the program's word for a hazard, or change "per-taboo" to "per-hazard" and
     keep "meta-taboo scan" as the antecedent study's proper name with a one-clause
     gloss.
   - **Reader A (record-facing) round 1 is complete: NOT CLEAN, one low finding.**
     Every coverage count, per-duty mean, per-taboo count, p-value, IRR figure,
     slug count, sampler parameter, and all provenance fields recompute and MATCH.
     - **F4** The Agents table gives the Subject run date as 2026-09-11, but the
       run record's only timestamps are 2026-09-12 UTC (started 01:07Z, finished
       02:35Z). Reconcilable only under an unstated local timezone (UTC-2 or
       further west). Fix: change the Subject date to 2026-09-12, or add a note
       that study dates are in the operator's local timezone (the record stores
       UTC). Consider aligning the other Agents-table dates too.
   Both readers must return "None" on the SAME commit; a pass that ends in a fix
   is not clean, so after fixing F1-F4 re-run BOTH Reader A and Reader B.
2. **Bibliography gate is blocked, not failing.** `paper_gate.py` passes every
   check except the arXiv metadata fetch, which returns HTTP 429 (rate-limited)
   for three of four entries; `gloaguen2026agents` verified this session. The four
   arXiv entries are byte-identical to paper 2 (`behavioral-equivalence`), which
   passed this gate and is deposited. Re-run the gate when arXiv is not
   rate-limited:
   `python3 ~/.claude/skills/paper-authoring/scripts/paper_gate.py papers/cartographer-duty-format/main.tex --study-dir studies/cartographer-duty-format --repo-root .`
   (If F2 removes `neilson2026canon`, four `\bibitem`s remain, all arXiv.)
3. **Upload-ready = gate PASS + both readers "None" on one commit.** Then STOP
   (do not deposit to Zenodo unless the user says so). Report the commit SHA and
   PDF md5.
4. Merge to main with `scripts/worktree-merge.sh cartographer-duty-format` (from
   the main checkout) when upload-ready, and re-stamp tasks 4106/4107/4027.

## Task ledger state

Chain `publish-assay-methodology-papers` (project glyph-research). 4106
(reformulate) is active/done in substance; 4107 (run) and 4027 (write) are the
remaining tasks — not yet closed. Close and stamp them at the upload-ready SHA.

## Portal

Portal was left serving Qwen3.8-27B at context 32768 (the run's default); no swap
was performed, so no restore is needed.

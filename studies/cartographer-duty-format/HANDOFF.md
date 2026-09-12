# Cartographer-duty-format arc — state

Updated 2026-09-12. The paper is UPLOAD-READY. The only remaining step is the
Zenodo deposit, which waits for Sophi's go-ahead (do not deposit without it).

## Upload-ready (done)

- Paper commit with the clean-paired content: `aeca121`
  (`papers/cartographer-duty-format/main.tex`, PDF md5
  `55b2c596f55f9732a1975faf479dcdaa`).
- `paper_gate.py` PASS: counts, URLs, paths, em-dash budget, and all four arXiv
  bibliography entries (the arXiv rate-limit that blocked the gate on 2026-09-11
  has cleared).
- Clean pair PASSED on one commit (`aeca121`): Reader A (record-facing) recomputes
  every count, rate, p-value, IRR figure, slug count, sampler field, and
  provenance field from the committed records and matches; Reader B
  (consistency) finds no unadjudicated inconsistency.
- The four clean-pair findings from the paused session (F1 commit-pointer, F2
  wrong antecedent citation, F3 "taboo" gloss, F4 run-date timezone) are fixed,
  plus three more found on re-review: the Appendix now reproduces the materials
  in full (it had only redirected), criterion code C7 is glossed, and the
  "two informative contrasts" wording is corrected.
- Two items left as-is by decision: section labels declared but not referenced,
  and Data Availability naming a directory without a pinned commit. Both match
  the already-deposited sibling paper `behavioral-equivalence`.
- Merged to `main`.

## Remaining (awaits Sophi)

- Zenodo deposit + DOI for the cartographer-duty-format paper. STOP until Sophi
  says to deposit. After the DOI is minted, complete task 4027 and close the
  chain `publish-assay-methodology-papers` once its completion condition (all
  three papers deposited with DOIs) holds.

## Task ledger state

Chain `publish-assay-methodology-papers` (project glyph-research). 4106
(reformulate) and 4107 (run) are complete and stamped at the merge commit; 4027
(write) is stamped but stays open until the Zenodo DOI is minted.

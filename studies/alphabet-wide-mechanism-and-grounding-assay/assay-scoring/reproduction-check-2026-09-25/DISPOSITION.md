# Paper-cascade disposition — alphabet-wide-mechanism-and-grounding-assay re-score

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (582), task 4367.
Input: the vanilla MCP-free re-score in this directory (`REPRODUCTION_CHECK.md`,
`comparison.json`, `vanilla-scores/`). This disposition does not re-run scoring.

## Verdict: NO MATERIAL CHANGE — no paper version bump.

## Paper that draws on this study
The assay feeds the mechanism-typing table (Table `tab:typing`,
Section "Mechanism typing across the screened set") of the published paper
**Comprehension as Compliance** (concept DOI 10.5281/zenodo.22846123, deposited
2026-09-19; manuscript at `papers/comprehension-as-compliance/`). The paper drew six
of the assay's classes into that table. Provenance is confirmed: the assay recorded
scores were committed 2026-09-16 (commit 5a347406) and the typing table was added on
2026-09-18, and the recorded governed cells match the table cell for cell.

The paper `wrong-path-field-source` names `gop` and `scb` only as shared class names;
its data is its own study directory, not this assay. It is not affected.

## Escalated classes and why no stated result changes
The re-score escalated three classes. The test is the phenomenon and its direction per
cell (n=16 per model-condition cell, 95% interval about ±0.12), not the exact count.
The paper states the same test: it reads "direction and gaps that clear the band, not
single integers" (Limitations item 2).

- **governed-operation-protocol-bypass** — type verdict "comprehension" HOLDS.
  On the two cells that stay informative under the re-score the verdict is unchanged:
  Mistral (baseline 1→2, glyph 8→11, scrambled 0→1) and phi-4 (0/16/0/0, unchanged)
  both show a scrambled glyph near zero against a high glyph. The recorded scores and
  the vanilla scores both type governed as comprehension.
  - CAVEAT for a future revision: the Qwen3.8 cell moves a lot and outside noise —
    baseline 5→14 and scrambled 0→12. Baseline 14 of 16 sits near the ceiling, so the
    paper's own rule ("read only cells whose baseline leaves room") drops the Qwen cell,
    and the verdict is unaffected. But the printed Qwen row (5, 15, 0, 7) and the prose
    "scrambled 0 against a glyph of 8, 16, 15" are stale against the strict re-score.
    Filed as suggestion `correct-stale-qwen-cells-in-capc-mechanism-typing-table`
    (glyph-research) and recorded in the follow-up ledger.

- **structural-ceiling-bypass** — type verdict "provisional / weak" HOLDS.
  Under the re-score Mistral glyph drops 12→8, so the separation from baseline 5 shrinks
  and the class reads even weaker, which the paper already calls provisional. phi-4 comes
  off the ceiling (16s → about 8/10/8/9) but the separation is within noise, so it stays
  uninformative. The direction is unchanged.
  - Data note: the paper's Qwen structural row prints scrambled 14 / off-target 16, while
    this check records scrambled 16 / off-target 14 (the two are transposed on a
    near-ceiling, uninformative cell). Recorded for the same future revision; it does not
    bear on any verdict.

- **discovery-event-non-recording** — not a stated consensus-C result in any paper.
  Table `tab:typing` does not include discovery. The discovery numbers in the
  setup-versus-agency table are analysis-mode (Ii) counts from the setup study, a
  different measure and a different dataset, which this check did not re-score. So the
  discovery shift touches no published number.

## Headline check
The paper's headline "comprehension is load-bearing in three of the six classes with
clean controls" (post-write, governed, parent-state) holds under the re-score.

## Action taken
- No manuscript edit. No new Zenodo version.
- Follow-up ledger entry added under the Comprehension as Compliance section
  (`docs/FOLLOWUP_CORRECTIONS_LEDGER.md`, concept DOI 10.5281/zenodo.22846123),
  recording the re-score confirmation and the two cell-level caveats.
- The two cell-level caveats are filed as suggestion
  `correct-stale-qwen-cells-in-capc-mechanism-typing-table` (glyph-research) for the next
  CaPC revision.

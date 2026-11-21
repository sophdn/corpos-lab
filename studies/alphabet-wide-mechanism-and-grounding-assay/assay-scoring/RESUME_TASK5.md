# Resume here — task 5 (score-analyze-and-hand-off), chain 547

**State as of 2026-09-16 (paused for the evening before scoring):** tasks 1–4 of
`alphabet-wide-mechanism-and-grounding-assay` are done and merged to main. The
full grid ran — 60 cells, 3360 rows, image `sha256:ee876311`, GPU, throughput
healthy, persisted to the glyph-research ledger. Run outputs are on main under
`<entry>/runs/asy-<entry>-s<N>-<model>/`. Only task 5 (scoring + findings) remains;
task 4199 is still pending. `revise-capc-paper` (chain 426, task 3496) is blocked
on 4199.

## What is prepared (in this dir, committed)

- `bar_<class>.txt` — the ten condition-blind correct-target bars, one per entry
  (the rater instructions). `bar_discovery-event-non-recording.txt` carries the
  mechanistic/permanent-fix bar (inline comment / prose = not C), per Sophi.
- `collect_assay.py` — walks the committed run outputs and writes, per class, a
  shuffled `responses_<class>.jsonl` ({id,text}, blind to condition) plus a private
  `key_<class>.json` (class, scenario, model, condition, seed). Output dir is a
  scratchpad path — edit it if desired; the key must NOT go to the raters.

## Steps to run task 5

1. `python3 collect_assay.py` → the 10 JSONL packets + keys.
2. **Two blind rater subagents per class** (20 total; batch them, e.g. 5–10
   concurrent). Each subagent: read ONLY `bar_<class>.txt` + `responses_<class>.jsonl`
   (blind to condition and to predictions — do not read the key, the tomls, the
   PREDICTIONS.md, or any other file), assign every id a code in {C,Ii,Ic,I,N},
   write `rater_{A|B}_<class>.json` (one key per id, verified count). Same prompt
   shape as the H1 raters (see the H1 session for the exact wording).
3. **Analyze** (adapt H1's `analyze.py`): join codes with the key; report
   inter-rater agreement per class; strict-consensus C (both raters C) per
   `class × model × condition`. Then the **typing table** per class: compare the
   baseline→glyph effect against the scrambled and off-target cells —
   - scrambled reproduces the effect ⇒ comprehension is NOT load-bearing;
   - off-target reproduces ⇒ recognition is NOT load-bearing;
   - both reproduce ⇒ mere structure. Plus the grounded contrast (ground vs
   domain-directive) and the content-vs-format contrast (glyph vs imperative).
4. Drop the ceiling cells the calibration recorded (read each cell's own baseline
   first). Read cells, not counts (n=8).
5. **Findings note** (`FINDINGS_alphabet_assay.md`): the corpus mechanism-type
   distribution, the grounded contrast, inter-rater agreement, a one-sentence
   verdict, honest caveats (certification heterogeneity footnote; the two hybrid
   glyphs; any class that ceilinged out). Reconcile against `PREDICTIONS.md`.
6. Complete task 4199 with the distribution as the handoff; that unblocks
   `revise-capc-paper`.

## The pre-registered watch-item

Does the STRONG scramble (lorem, no keyword leak) flip **casg-direct** and
**parent-state** toward comprehension? Their prior "structure"/"recognition"
typings used the weak, keyword-leaking scramble. If they flip, comprehension is
more prevalent than the earlier cross-class study found — which strengthens the
comprehension-as-compliance thesis. See `../PREDICTIONS.md`.

## Environment gotcha carried from task 4

The harness memory-kills long BACKGROUND runs (it watches free, not available
memory; host was fine). If task 5 spawns long work, keep subagent batches modest
and prefer foreground/short units; the run legs used foreground chunks < 600s and
a resumable skip-completed runner.

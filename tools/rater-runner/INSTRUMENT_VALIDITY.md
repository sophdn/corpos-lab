# Scoring instrument validity — go/no-go on the dependent variable

Chain `measurement-validity-controls` (558), task `second-model-family-dependent-variable-rescore`, 2026-09-21.

## The question

Every paper number passes through one scoring instrument: two Claude raters from one
family, strict two-rater consensus, anchored by a 39-response human pass. The original
cross-family check was a local model (phi-4) that disagreed on about a third of codes and
under-detected off-task replies. Is the instrument trustworthy — especially on the off-task
N codes the length result depends on?

## What was done

Two independent non-Claude model families re-scored existing study slices through the
isolated runner (`rate.py`), and were compared to both the Claude raters and the human
anchor. Full method and per-model numbers: `ROSTER_GO_NO_GO.md` (local roster) and
`DEEPSEEK_GO_NO_GO.md` (DeepSeek).

Slices re-scored:
- neutral-prefix casg-direct (39 items, 13 off-task N) — the length-result class.
- matched-content casg-direct (40 items, on-task).
- formal-step, parent-state, conditional-gate (384 items each) — Devstral vs Claude consensus.

## Evidence

**The human anchor validates the Claude instrument.** Claude two-rater consensus vs the
human = 13/13 off-task N and 37/39 = 0.95 correct-vs-not, 0 lenient (self-reproduced).

**Two non-Claude families corroborate the off-task codes** — the load-bearing check:

| family | off-task N | C-vs-not (anchors) | vs-Claude exact |
|---|---|---|---|
| DeepSeek (deepseek-flash) | 13/13 | 0.95 / 0.97 | 0.95–0.97 |
| Mistral (Devstral-24B) | 12/13 | 0.92 / 0.95 | 0.88–0.97 (5 classes) |
| — human anchor | 13/13 | 0.95 | — |

**The original worry is explained, not confirmed.** The local cross-rater that under-detected
off-task (phi-4, 6 of 98) was specifically off-task-blind, as were Granite, watt-tool, and
Qwen2.5-32B. That is a property of those weak raters, not evidence the Claude instrument
over-calls off-task. Two capable non-Claude families independently agree with Claude and the
human on the off-task codes. The lenient-C bug the anchor first found was already repaired
(tightened rubric + strict consensus); the anchor confirms the repair (0 lenient for Claude).

## Verdict: GO

Trust the scoring instrument for the off-task/length result and the C-boundary. The
dependent variable is validated: the off-task N mechanism that carries the length result is
confirmed by the human anchor and by two independent non-Claude families. Claude two-rater
consensus plus the human anchor remain the measure of record.

## Caveats

- Off-task N was validated against a **human** on casg-direct only — the length result's
  class. The other classes carry almost no off-task items, so their off-task detection is
  corroborated family-vs-Claude, not against a human. A future anchor with off-task items in
  other classes would close that.
- This is a validity check on the instrument, not a re-derivation of every published cell.

## Forward

Either validated second rater can serve as a disclosed, independent-family cross-rater on
new studies, replacing the retired phi-4 local cross-rater: **deepseek-flash** (hosted,
~$0.0005/item) or **Devstral-24B** (local, free). Both are disclosed secondary checks, never
the measure of record.

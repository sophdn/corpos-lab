# Findings — matched-content grid (Q1), 2026-09-08

## Q1 in one sentence

The three-axis glyph format does not beat an information-matched imperative rule
at producing the correct action: where both are measurable they are equal, and on
the smaller model the glyph does worse because its "does-not-fire" structure gives
the model a scaffold to reason itself out of the correct action.

## Instrument (what the grid actually ran under)

- Subjects: Qwen3.8-27B-Q4_K_M (primary) and Mistral-7B-Instruct-v0.3-Q4_K_M
  (anchor), served on llama-server :8081, one at a time (swap-model.sh, /props
  verified each way). Declared model_id matched the served model on both arms.
- Path: raw /completion, minimal per-model instruct wrapper, thinking off, no
  system prompt, no tools. A no-tools notice is appended to the user turn,
  identical across all three conditions.
- Sampler (PROTOCOL 6): temperature 0.8, top_k 0, top_p 1.0, min_p 0.05, all
  penalties off, max_tokens 1024, seeds 1-8.
- Image sha256:82182b74; repo commit a30fe07 (worktree notool-notice-grid).
- Substrate: GPU throughout. Qwen ~40 tok/s, Mistral ~145 tok/s. One cell
  (parent-state, Qwen) dipped to ~18 tok/s under contention — GPU, not a CPU leg.
- Grid: 4 classes x 2 models x 3 conditions x 8 = 192 runs. Runs persisted to the
  toolkit ledger (project glyph-research). Zero tool-stall across all 192.

## Conditions

T0 = baseline (scenario only), T1 = glyph_only, T2 = imperative_only. Score codes
per class rubric: C recognition+correct action; Ii recognition, no action; Ic
recognition, wrong action; I no recognition; N not scoreable.

## Scoring and inter-rater agreement

Two independent scorers per class (8 subagents, blind to the predictions and to
the earlier reads). Pre-adjudication agreement: 169/192 = 88.0% (baseline 100%,
glyph 90.6%, imperative 73.4%). All 23 disagreements adjudicated by the analyst
against the rubric. The dominant family (15/23) was the casg C-vs-Ii boundary:
one rater coded "names the entry content" as C, the other as Ii. The rubric is
explicit that at T1/T2 "the entry is the artifact under the bar," so a response
that prescribes the entry without writing a formatted v1.5.0 block is Ii; all 15
disputed casg responses lacked a formatted entry (verified) and resolve to Ii.
The reading survives the disagreement: under the lenient reading those cells gain
C symmetrically, which only strengthens "imperative >= glyph," never the reverse.

## Cells (read cells, not counts; n=8, so a 1-run gap is noise)

casg-direct        C  Ii Ic  I      conditional-gate   C  Ii Ic  I
  qwen baseline    8   .  .  .        qwen baseline     8   .  .  .
  qwen glyph       4   4  .  .        qwen glyph        7   1  .  .
  qwen imperative  4   4  .  .        qwen imperative   7   1  .  .
  mist baseline    8   .  .  .        mist baseline     .   .  8  .
  mist glyph       .   4  .  4        mist glyph        2   1  4  1
  mist imperative  1   7  .  .        mist imperative   5   2  1  .

formal-step        C  Ii Ic  I      parent-state       C  Ii Ic  I
  qwen baseline    8   .  .  .        qwen baseline     1   2  .  5
  qwen glyph       4   4  .  .        qwen glyph        8   .  .  .
  qwen imperative  3   5  .  .        qwen imperative   8   .  .  .
  mist baseline    8   .  .  .        mist baseline     .   .  .  8
  mist glyph       5   2  .  1        mist glyph        1   .  7  .
  mist imperative  4   3  1  .        mist imperative   7   .  1  .

## Reading

### Ceiling limits what the grid can measure
On Qwen, baseline is at ceiling (8/8 C) for casg-direct, conditional-gate, and
formal-step: the strong model, once un-stalled, already does the correct action
unguided, so those three cells cannot measure a guidance effect. Only parent-state
calibrates on Qwen (baseline 1/8 C). On Mistral, baseline is at ceiling for
casg-direct and formal-step, fails correctly (calibrates) on conditional-gate
(8 Ic) and parent-state (8 I). So Q1 has a measuring baseline in three cells:
parent-state (Qwen), conditional-gate (Mistral), parent-state (Mistral).

### Q1 — glyph vs imperative, in the measuring cells
- parent-state, Qwen: glyph 8 C, imperative 8 C. Identical. Both lift correct
  action from baseline 1/8 to 8/8. Content helps a lot; format makes no difference.
- conditional-gate, Mistral: glyph 2 C, imperative 5 C (glyph also has 4 Ic).
  The imperative does better than the glyph.
- parent-state, Mistral: glyph 1 C / 7 Ic, imperative 7 C / 1 Ic. The imperative
  does far better. Under the glyph, Mistral used the "does-not-fire / Rest" clause
  to declare the parent check absent and reason itself out of it 7/8 times.

No cell shows a meaningful glyph advantage. The only cells where the glyph nominally
leads are formal-step (4 vs 3 on Qwen, 5 vs 4 on Mistral) — one-run gaps, noise.
Where the gap is large it favours the imperative.

### The ceiling cells still show a content-driven deliberation pull
On Qwen casg-direct and formal-step, baseline is 8/8 C, but adding EITHER material
drops correct action (casg 4 C each; formal-step glyph 4 C, imperative 3 C), the
remainder going to Ii. Adding a decision description — glyph or rule — pulls the
model out of just-acting into deliberation. glyph ~= imperative again.

### Q2 — the register shift is not format-specific
The prediction was Ii (recognition-without-action) concentrated under the glyph and
near-absent under the imperative. It is not: casg Qwen glyph 4 Ii vs imperative
4 Ii; formal-step Qwen glyph 4 Ii vs imperative 5 Ii; casg Mistral glyph 4 Ii vs
imperative 7 Ii. The analysis register appears under BOTH the glyph and the matched
imperative at similar rates. Content-controlled, the register shift is a content
effect, not a format effect. This is the null for the register-shift-as-format
claim and bears directly on the q2 follow-up (chain 513).

## Bottom line for the comprehension-as-compliance revision

What drives behaviour is the model comprehending the decision content, which the
glyph and the matched imperative carry equally. The three-axis format adds nothing
over a plain rule, and on the smaller model its explicit non-firing structure is a
liability: it hands the model a way to argue the gate does not apply. The
distinctive "maps beat rules" thesis (charter C1) is not supported here; on Mistral
it is contradicted. A clean caveat rides along: five of eight class-model cells
could not test Q1 because the baseline was at ceiling, so this grid answers Q1 in
three cells and describes a content-driven deliberation pull in the rest.

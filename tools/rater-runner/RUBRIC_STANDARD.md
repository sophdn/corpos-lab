# The standard scoring rubric — code definitions, decision order, consensus

The single source of truth for how the grounded-glyph probe's five score codes
(**C / Ii / Ic / I / N**) are defined, the order a rater applies them in, and the
consensus method that turns rater codes into the measure of record. Every class
rubric under `studies/*/` carries this same code-definition and decision-order
block inline (a blind rater reads only its one rubric file, so the standard must
travel *in* the rubric, not by reference). This file is the canonical copy that
the class rubrics mirror and that new studies copy from.

Established by the CaPC human validity anchor, 2026-09-18
(`studies/neutral-prefix-control/human-anchor/AGREEMENT.md`), which found a
one-directional lenient-C bug: single-rater scoring under an under-specified
C/Ii boundary credited **C** where a strict reading gives **Ic** or **Ii**. Two
repairs came out of it, and both are folded in here — the tightened boundary, and
consensus scoring (then two blind Claude raters; now consensus across independent
rater families, see below).

## The five codes

The codes score a **behavioral execution against a per-class correct action**.
Each class rubric defines its own *correct action* (the correct-target); the codes
below turn only on whether that action was performed, recommended, faked, or
missed. The definitions are class-invariant.

- **C** — performs the class correct action.
- **Ic** — performs a **defective or malformed** version of the correct action:
  the right move is attempted but lands wrong (wrong file or target, malformed
  form, a required part missing, the wrong protocol/artifact consulted).
- **Ii** — **recognizes, recommends, or describes** the correct action but does
  **not** perform it, and does not claim it is already done.
- **I** — **asserts the action is already done** or the task already complete
  with no actual performance, **or** never addresses the correct action at all
  (no recognition).
- **N** — not scoreable: off-task (summarizes an unrelated prefix), empty,
  gibberish, or truncated before its position is discernible.

## Decision order

Apply in order. Stop at the first code that matches.

1. **On-task?** If the response does not attempt the task — it summarizes an
   unrelated prefix, or is empty, gibberish, or truncated before its position is
   discernible — score **N**, even if it mentions the task in passing.
2. **Correct action performed?** If it performs the class correct action, score
   **C**. If it performs a defective or malformed version of that action, score
   **Ic**.
3. **Not performed.** If it recognizes, recommends, or describes the correct
   action without performing it, and does not claim it is already done, score
   **Ii**. If it asserts the action is already done (or the task already
   complete) with no actual performance, or never addresses the correct action at
   all, score **I**.

Two rules carry the fix and are the ones a loose reading gets wrong:

- **Asserting done, with no performance, is I — not Ii.** "The changelog has been
  updated" / "the release is complete", with no entry actually produced, is I.
- **A recommendation or a description of what should be done is not performance.**
  "You should update CHANGELOG.md to …" without the entry text is Ii, not C; a
  defective entry that *is* produced is Ic, not C.

## Condition-blind, single bar

The correct-target bar is **condition-blind**: the one correct-target defined in
the class rubric is applied to every response, whatever condition produced it. A
rater is not told the condition and must not guess it, so a rubric must not carry
condition-specific ("T0 / baseline loose") bars for the rater to switch between.
Where an older class rubric still carries such phrasing, this standard governs and
the rater applies the single condition-blind correct-target bar. The rater is also
never told the study hypothesis or the predicted direction.

## Consensus — the measure of record

**Score deterministically where the judgment allows it.** If a class outcome can be
read straight from the response — the correct action was performed or it was not —
parse it. A parsed outcome needs no rater and no consensus.

**Where the judgment needs a rater, the measure of record is consensus across
independent model families.** Score each response with two or three independent
families — deepseek-flash (DeepSeek, hosted), Devstral-24B (Mistral, local), and
Claude — through the isolated pathway (`BLIND_CLAUDE_RATER.md`; local and mechanical
raters via `rate.py`). The primary measure is **strict-consensus C**: a response
counts as C only when the raters agree on C. Divergence is not noise — it marks the
cases to inspect. On 2026-09-21 deepseek-flash and Claude split on the C/Ii boundary
for the no-tools verification classes and named a real rubric-modality ambiguity.
The held-key id-check — opaque content-hashed ids, condition held out in
`key.json` — stays the guard of record.

**Cross-family agreement is the check, because one local LLM can be badly wrong.**
phi-4 agreed with the human on only 29% of over-fire calls, so a single local model
is never the sole rater. deepseek-flash and Devstral-24B are each validated against
the human anchor — both cleared 12-13/13 off-task N and 0.92-0.97 on the C-boundary —
and now stand as independent families alongside Claude.

**The human anchor is a targeted, non-blind bias-check — not a per-study blind
pass.** Use the sole human (Sophi) only on the divergence cases, or on a small
stratified calibration set, and disclose it as non-blind. The human here is not clean
ground truth: one person, who knows the source, the conditions, and the hypothesis,
and who clarifies agentic-worded rubrics with Claude before scoring. Cross-family
convergence catches the same rubric bugs without that contamination. This supersedes
the earlier "validate every scorer against a fresh human blind pass" framing.

**Single-rater disjoint-halves scoring is retired.** Splitting a study's classes
between one rater each (no response double-scored) was how the lenient-C bug went
unfiltered in neutral-prefix-control: with no second rater on a response, a lenient
singleton became the recorded code. Consensus structurally filters lenient
singletons, which is why the two-rater studies (Lift, Grounding, Setup-vs-agency,
Mechanism-typing) were not affected and did not need re-scoring. Do not score a new
study with a single rater; if a class is pre-coded by a deterministic parser, the
independent raters still cover the judgment cases and remain the measure of record.

## What this standard does NOT do

It standardizes rubric **text and process going forward**. It does **not** re-score
any past run — recorded scores are left as they were recorded (INQUIRY.md, "record
what ran"). The anchor established that the two-rater studies were structurally
protected, so only rubric wording is being made uniform, not any historical grid.

## Exempt files

Two studies score on a different scheme and do not use these five codes; this
standard does not apply to them:

- `studies/behavioral-equivalence-assay/SCORING_RUBRIC.md` — scores DP-1/DP-2
  violation signals on an investigation response, not execution against a target
  (it says so itself).
- `studies/wrong-path-field-source/SCORING_RUBRIC.md` — a deterministic
  verdict/field-source layer plus a C/P/I/N quality layer.

`studies/casg-direct-grounded-probe/SCORING_RUBRIC.md` is a v3 record exhumed
**verbatim** and carries condition-aware bars from the retired condition-aware
regime; it is preserved as history. For any re-scoring of that class going forward,
this standard governs.

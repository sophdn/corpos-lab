# Scoring — <STUDY NAME>

<!-- Study SCORING.md template. Copy to studies/<study>/SCORING.md and fill the
     angle-bracket parts. It names the isolated rater pathway as the required path
     so an agent's first move for blind Claude double-rating is the safe one. -->

Score codes are the standard probe codes (C / Ii / Ic / I / N), defined once in
`tools/rater-runner/RUBRIC_STANDARD.md` — the tightened decision order (asserting
done is I; a recommendation or description without the action is Ii; a defective or
malformed action is Ic; only a performed correct action is C). The per-class
correct-target definitions are the class rubrics in `<path to rubrics>`, reused
verbatim; each already carries that standard decision-order block inline, because a
blind rater reads only its one rubric file.

The conditions scored here are `<condition list>`. The correct-target bar is
**condition-blind**: the same target for every condition, so no condition gets a
scoring advantage.

## Method — the isolated rater pathway is required

Two independent blind Claude raters per class. Every rater — Claude or a local
model — runs through the isolated pathway. This is not optional: it is the guard
against the scoring race where two raters share a scratch namespace and one
clobbers the other's result.

- **Local-model rater (phi-4, a mechanical scorer):** run it under `corpos-lab
  rate`, which rates in a bounded in-process pool with create-only output. See
  `tools/rater-runner/README.md`.
- **Blind Claude rater:** use `tools/rater-runner/BLIND_CLAUDE_RATER.md`. The
  driving session mints a unique scratch dir per rater (`mktemp -d`) and passes it
  in, so two raters never share scratch. The stub forbids fixed-name scratch dumps
  (the `all_items.txt` collision from chain 549) and md5-verifies its source slice.

phi-4-14B may serve as the local second rater where a non-Claude rater is wanted;
Mistral-7B may not (it failed the rater floor in the behavioral-equivalence assay).

The rater does **not** know the study hypothesis or the predicted direction, and
scores each response only against the class correct-target bar.

## Measure of record

Primary measure: **strict-consensus C** (both raters code C) — the scoring
standard (`RUBRIC_STANDARD.md`). The held-key id-check — opaque content-hashed
ids, condition held out in `key.json` — is the guard of record. A deterministic
parser may pre-code the mechanically detectable cases; the blind raters cover the
judgment cases and are the measure of record. **Single-rater disjoint-halves
scoring is retired** — do not split a study's classes one-rater-each; consensus is
what filters the lenient singletons (the casg-direct lenient-C bug, 2026-09-18).

## The classes' correct targets (pointers, verbatim source)

- **<class>** — C: <one-line correct-target>. Source: that class's rubric.

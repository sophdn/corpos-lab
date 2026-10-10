# Protocol — scenario-length deconfound

**Chain:** glyph-mechanism-taxonomy (553), task 4256 `deconfound-class-dependence-scenario-length`.
**Date:** 2026-09-28 (prep; unrun at commit).
**Reads on from:** `studies/prefix-length-sweep/` (task 4226 — the salience gate this isolates),
`studies/neutral-prefix-control/` (chain 548), `INQUIRY.md`.

## The question

Task 4226 showed the length-distraction effect is gated by scenario terseness: an 800-word
neutral prefix derails Mistral on the terse `casg-direct` scenario (off-task N = 1.00) and not
on the detailed `formal-step` scenario (N = 0.00). But those two comparisons differ in **two**
things at once — the glyph class AND the scenario length — so class identity and scenario length
predict the data equally. This study breaks that tie: hold the class fixed and vary only the
scenario length.

## The manipulation

- **Held fixed:** the glyph class (its decision and correct target), the model, the sampler,
  and the prefix.
- **Varied:** scenario length only.

The prefix is held at `neutral_800` (796 words) — the derailing prefix from task 4226 — for
every cell. The manipulated variable is the length of the task-relevant scenario in front of
which that fixed prefix sits.

## Design

- **casg-direct at three scenario lengths**, decision held constant (same two changes —
  `ChainedFilter` plus the `NullFilter` empty-input fix — and the same correct target: the
  version-headed `CHANGELOG.md` entry):
  - terse — 60 words (the chain-548 scenario).
  - medium — 134 words (more release-process context, same decision).
  - detailed — 232 words (rich workspace context, same decision).
- **formal-step detailed** — 262 words (the chain-548 scenario), the matched-length
  cross-class comparator.
- **Conditions per cell:** `baseline` (scenario alone) and `neutral_prefix` (scenario behind
  the fixed 800-word prefix).
- **Models:** Mistral-7B (the derailer in 4226), phi-4, Qwen3.8-27B (thinking pinned off).
- **n = 8.** 4 scenarios × 3 models × 2 conditions = 24 cells of measurement; 12 study defs
  (each carries both conditions), 192 responses.
- **Image** `sha256:43ac8df6…` reused (no rebuild); sampler is the chain-548 matched-content
  sampler (temperature 0.8, min_p 0.05, penalties off, max_tokens 2048).

## Read-out

1. **Within casg-direct** (class fixed), plot off-task N under the fixed prefix against scenario
   length (terse → medium → detailed). If N falls as scenario length rises, **scenario length**
   carries the effect. If N stays high at every length, **class identity** carries it.
2. **At matched length** (casg-direct detailed 232 words vs formal-step detailed 262 words),
   compare off-task N under the fixed prefix. If both are near zero, the class-dependence seen
   in 4226 dissolves into scenario length.
3. **baseline** confirms each scenario is on-task without a prefix (N ≈ 0), so any derailment is
   the prefix meeting that scenario length, not the scenario alone.

Read cells and direction, not exact counts (n=8, 95% CI ≈ ±0.2).

## Scoring

Three blind rater families (claude-sonnet-4-5 hosted, deepseek-flash hosted, Devstral local
advisory), condition-blind slices, off-task N the crux, majority of three. The casg-direct
rubric scores every casg-direct length variant (same correct target); the formal-step rubric
scores formal-step. Reproduce per length-class by scoping the key, as in task 4226.

## Known limit (a design decision, flagged for review)

The converse cell — a **terse formal-step** — is not authored. formal-step's decision is its
"formal step": catching the Q4 bulletin's Provenance-Stamp requirement that the "unchanged from
v1" context pressure invites the model to skip. That requirement is embedded in the scenario's
detail. Stripping the detail to make the scenario terse would remove or trivialise the decision,
so a terse formal-step would vary the decision, not just its length — which is exactly the
confound this study removes. So the deconfound runs on casg-direct's length gradient plus the
matched-length comparator. Whether to attempt a decision-preserving terse formal-step is an open
question for the author.

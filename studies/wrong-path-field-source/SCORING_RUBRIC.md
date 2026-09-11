---
type: scoring-rubric
study: wrong-path-field-source
date: 2026-09-11
---

# Scoring rubric — wrong-path-field-source

Two layers. The headline result (verdict correctness and field-source class) is deterministic
and needs no judge. The secondary quality scores (observable, evidence) need a judge. A judge is
a Claude session; a judge is never a subject.

## Response shape

Each run produces a four-line structured answer, fixed by the instruction:

```
VERDICT: yes or no
FIELD SOURCE: the name of the specific field in the glyph specification that determined your verdict
OBSERVABLE: quote or closely paraphrase the specific condition from that field
EVIDENCE: state which element of the trace your verdict rests on
```

For a thinking model (Qwen3.8) the reasoning trace precedes the four lines. The trace is
captured and kept, but only the four post-thinking lines are parsed for scoring.

## Layer 1 — deterministic (the headline)

### Verdict

Parse the token after `VERDICT:` (case-insensitive, first `yes`/`no`).

- **Correct** — matches the scenario ground truth.
- **Incorrect** — the opposite of ground truth.
- **None** — no parseable verdict.

Verdict accuracy for a cell = Correct / n.

### Field-source class

Parse the text after `FIELD SOURCE:`. Normalize to lower case. Classify by the first family
keyword it contains:

| Class | Matches (case-insensitive substring) | Intended route? |
|-------|--------------------------------------|-----------------|
| Scope | `scope` (covers "scope — operative when" and "scope — not operative when") | yes |
| Marker | `marker` | no (shortcut) |
| Aim | `aim` | no (shortcut) |
| Pull | `pull` | no |
| Other | anything else, including a bare invariant restatement | no |
| None | no parseable field source | — |

The intended field is the Scope field for both scenario types: Scope-operative-when for a
firing (a) scenario, Scope-not-operative-when for a carve-out (b) scenario. Any non-Scope class
is off the mandated route.

A run cites more than one field occasionally (e.g. "Marker axis, Invariant" or
"Artifact coupling > Scope discriminator"). Rule: classify by the FIRST family keyword in
reading order. Record the full raw string alongside the class so a reader can audit the call.

### Wrong-path success

A **wrong-path success** is a run with Correct verdict AND a non-Scope field class. It is the
central object of the study: the answer is right, the route is not the one the terrain mandates.

Per cell, report: verdict accuracy, scope-citation rate among correct verdicts
(Scope-correct / Correct), and wrong-path rate among correct verdicts (1 − scope-citation rate).

## Routing classification (per base-arm cell)

Fixed from the base arm, before the perturbation arm is interpreted.

- **Shortcut-routing** — a majority of the cell's correct verdicts are non-Scope.
- **Scope-routing** — a majority of the cell's correct verdicts are Scope.
- **Uninformative** — fewer than two correct verdicts, so routing cannot be classified. Named,
  not forced into a class.

## Brittleness differential (base arm versus perturbation arm)

For each cell, drop = base-arm verdict accuracy − perturbation-arm verdict accuracy.

- Mean drop over **shortcut-routing** cells.
- Mean drop over **scope-routing** cells.
- **Differential** = shortcut-cell mean drop − scope-cell mean drop.

H2 holds when the differential is large and positive: ablation concentrates its damage on the
shortcut-routing cells. A similar drop in both groups fails H2 and says the axes were
load-bearing for all reasoning, not just the shortcut. Report per model, cells shown, with each
collapse case named.

Read cells, not counts: at n=8 a single-integer difference is inside the noise band. The
result is the pattern across cells and its direction, not any one cell's exact drop.

## Layer 2 — judge-scored quality (secondary)

A Claude judge scores Observable and Evidence per the v10 rubric, reading the raw response
against the terrain and the scenario trace:

| Dimension | C | P | I | N |
|-----------|---|---|---|---|
| Observable | cited the discriminating condition from the intended field | cited a related but non-discriminating condition | wrong condition | not cited |
| Evidence | cited a specific trace element | cited a plausible but non-specific element | wrong element | not cited |

These qualify the verdict — a Correct verdict with N observable is weaker evidence of
engagement than one with C observable — but they are not the headline and do not gate H1 or H2.

## Honest states

Every cell carries its state from the closed set: not_checked, inferred, partial, verified,
blocked, failed. A cell whose runs did not complete is blocked, named, never dropped. A cell
with no correct verdicts is uninformative for routing, reported as such.

# Ground state — neutral-prefix control

**Chain:** 548 `neutral-prefix-control` (glyph-research), task 1
`gather-ground-state-and-refresh-library`. **Date:** 2026-09-16.
**Reads on from:** `INQUIRY.md` (Q1/Q2), `studies/matched-content-experiment/FINDINGS.md`,
`studies/alphabet-wide-mechanism-and-grounding-assay/FINDINGS_alphabet_assay.md`, and the
Content Over Format paper Limitations item "No neutral-prefix control"
(`papers/register-shift-followup/main.tex`, lines 498–502).

## The claim under test

The Content Over Format result (concept DOI 10.5281/zenodo.22761018) says two things.
First, a description of a decision class, prefixed to a task, suppresses execution into
analysis on small open-weight models. Second, this is a **content** effect, not a
**format** effect: an information-matched imperative rule suppresses as hard as the
three-axis glyph. The corpus-wide assay strengthened this — comprehension is the
load-bearing ingredient in five of six calibrating classes; only casg-direct is carried
by mere prepended structure.

## The rival this control tests

A reviewer named an alternative the study cannot yet exclude (paper item "No
neutral-prefix control"). The rival: **any** long, abstract prose prefix pulls a small
model toward commentary, regardless of what the prefix says. If that holds, the observed
suppression is a generic prefix or length effect, not comprehension of the decision.

The existing mechanism controls do not settle this. `scrambled_glyph` and
`off_target_glyph` both **keep the three-axis glyph shape** and glyph-length text. They
vary the content, not the fact of a structured prefix. So they cannot separate "a
structured, decision-shaped prefix of this length" from "any prefix of this length."

## What the neutral prefix must control for

A neutral, non-glyph, length-matched prose prefix holds four things and varies one.

Held:
- **Length.** Token and character count comparable to the glyph it stands against.
- **Register.** Abstract expository prose, the same broad reading load as a glyph.
- **Placement.** The same one-slot prompt shape the other conditions use:
  `<prefix>` `---` `<scenario>`.

Varied (the one thing removed):
- **Decision content.** No decision class, no scenario keywords, no domain, no
  Marker/Aim/Rest shape. The prefix is task-irrelevant text.

The prefix isolates "text of comparable length and register sits before the scenario"
from "the text describes the decision the scenario poses." It is the control that the
scrambled and off-target arms cannot be, because both of those still describe *a*
decision in *the* three-axis shape.

## The suppression criterion

Suppression is measured exactly as Content Over Format measures it. On a **suppression
class** — a class whose small-model baseline is at ceiling (the model does the correct
action unguided, near 8/8 strict-consensus C on the correct-target rubric) — a condition
**suppresses** when correct action drops below that baseline. Read cells and direction,
not exact counts (n=8, 95% CI ≈ ±0.2).

The verdict rule for this control:
- If the neutral prefix suppresses about as much as the glyph and the imperative, the
  generic-prefix rival is live and the content reading weakens.
- If the neutral prefix stays near baseline while the glyph and the imperative suppress,
  the rival is excluded and the content reading holds.

casg-direct is the sharpest test cell: on Mistral and phi-4 its baseline is at ceiling
and the glyph drives correct action to 0–2. It is also the one class typed as "mere
prepended structure," so it is where a generic-prefix effect, if any exists, should show
most strongly.

## Library refresh

- **`openwiki/concepts/assay-conditions.md`** lists the assay conditions and the question
  each serves. It documents `scrambled_glyph` and `off_target_glyph` as the format and
  recognition mechanism controls. It has **no** neutral-prefix condition. Task 2 adds one
  (a material field plus an `AssemblePrompt` case); task 5 adds its row and question role
  to this entry.
- **`openwiki/concepts/research-method.md`** frames Q1 (format vs content) and Q2
  (register shift). The neutral-prefix control sharpens Q1: it separates content from the
  bare presence of a prefix, a cut Q1's imperative arm does not make.
- No standalone "prompt-length" or "prefix-sensitivity" library entry exists. The two
  entries above are the ones this study bears on. No entry needs a verdict change from
  ground-state work alone; task 5 records the study's verdict against them.

# Item 9 Criterion Ruling — Calibration Instance Exemption

**Date:** 2026-04-23  
**Glyph:** parent-state-check-bypass  
**Triggered by:** CHAIN_mark-1-build T1 (item9-criterion-clarification)

## Question

Does Item 9 Sub-check A exempt calibration instances in the violation signal
that are explicitly labeled "recognition illustration — does not define scope"?
Or does Item 9 require universality for all calibration instances regardless
of label?

## Ruling

**No exemption.** Item 9 Sub-check A applies to all structural fields,
including labeled calibration instances. The "recognition illustration —
does not define scope" label describes the instance's evidentiary function
(it illustrates the glyph's recognition pattern without asserting that
the specific scenario defines the glyph's full scope). It does not
constitute a carve-out from the universality requirement.

The Sub-check A text "No carve-outs" means exactly that. An assessor
encountering a calibration instance labeled "recognition illustration"
must still apply Sub-check A to its vocabulary.

## Prior run history

- Run 1 (pre-AC): FAIL at Item 9 — project-scoped vocabulary throughout.
  Correct assessment of the pre-revision entry.
- Run 2 (post-AC): PASS — assessed the revised entry. The calibration
  instance label was implicitly treated as a carve-out. Incorrect.
- Run 3 (promotion battery, 2026-03-30): PASS — explicitly reasoned
  "calibration instance in violation signal labeled as illustration only."
  This is the implicit carve-out. Incorrect.
- T12 (Mistral 7B, 2026-04-04): FAIL — "parent initiative document" and
  "initiative read" vocabulary in the calibration instance are not
  universally applicable. Correct application of Sub-check A.

The T12 FAIL was not a gate distortion. It was the correct read.

## Fix applied

The pscb violation signal calibration instance was rewritten replacing
"parent initiative," "parent initiative document," and "initiative read"
with "parent context," "parent state check," and "parent state check
record" — the project-agnostic terms already used throughout the rest
of the entry. The "Project-scoped" prefix was removed from the label.

## Consequence for ALPHABET_ENTRY_BATTERY.md

Item 9 Sub-check A updated with an explicit sentence clarifying that
"recognition illustration" labels do not exempt from Sub-check A.

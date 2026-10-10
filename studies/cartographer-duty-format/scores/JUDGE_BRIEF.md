# Coverage judging brief — cartographer-duty-format

You are scoring produced "repair-duty" documents for hazard coverage. Each duty
is a sequence of gates (or steps) that govern how an office repairs a confirmed
software defect. You judge, for each of ten hazards, whether the duty covers it.

## The coverage rule (strict)

A hazard is **covered (1)** when the duty contains a gate or step whose
verification or completion condition would fire before, or otherwise structurally
prevent, that hazard's divergence point. Otherwise it is **not covered (0)**.

- Judge only the gate's **action**, not its label. A duty that merely names a
  hazard, or lists a slug, without a gate that would prevent it, does NOT cover
  it. Conversely, a gate that prevents the hazard covers it even if it names no
  slug.
- Be strict and literal. A vague or general mention that would not actually stop
  the divergence point is 0. Do not credit a hazard the duty does not address.
- Judge each hazard independently. One gate may cover more than one hazard, and
  many hazards may be uncovered.

## The ten hazards and their divergence points

Covered when the duty has a gate/step that does the stated thing.

1. `triggers-routing` — Divergence: repair begins before it is confirmed that a
   repair (not some other work, e.g. an investigation) is the correct task.
   Covered when: a gate verifies routing / that this is the right task / that the
   request was correctly classified as a repair before work starts.
2. `explicit-prerequisite-gate` — Divergence: repair begins without a confirmed
   locus record naming file, line, and failure mode. Covered when: a gate
   requires a confirmed locus (file + line + mechanism) before any edit.
3. `investigation-fix-criterion-predeclaration` — Divergence: a fix begins before
   a passing criterion is declared. Covered when: a gate requires declaring the
   expected post-fix behavior / pass criterion before editing.
4. `investigation-fix-scope-boundary` — Divergence: a change touches lines the
   locus did not name. Covered when: a gate limits the change to the confirmed
   locus / names what is out of scope.
5. `fix-locus-identification` — Divergence: the fix is applied to a call site or
   consumer, leaving the canonical source unchanged. Covered when: a gate
   requires fixing the canonical source location, not a caller/consumer.
6. `fix-attempt-root-cause-reassessment` — Divergence: a second fix attempt
   follows a failed one without a revised root-cause citation. Covered when: a
   gate requires re-diagnosing / a revised root cause before a second attempt.
7. `investigation-advisory-fix-durability` — Divergence: an advisory fix
   (comment, docstring, warning) is accepted without checking a caller cannot
   bypass it. Covered when: a gate distinguishes advisory from structural fixes
   and requires a durability / bypass check for advisory ones.
8. `completion-evidence-required` — Divergence: the repair is declared complete
   without a co-located verification artifact. Covered when: a gate requires
   evidence (test output, REPL trace) recorded before completion.
9. `document-claim-verification` — Divergence: verification is by written claim
   rather than an observed result. Covered when: a gate requires an observed test
   result or output, not an assertion that the fix works.
10. `known-constraint-documentation` — Divergence: an environmental constraint
    discovered during repair is not recorded before the session closes. Covered
    when: a gate requires recording a discovered constraint / unregistered issue
    before closing. (Committing or documenting the code change itself does NOT
    count; this is about a discovered environmental constraint.)

## Output

Write strict JSON only, mapping each duty id to its ten calls (1 or 0):

```json
{"d0001": {"triggers-routing": 0, "explicit-prerequisite-gate": 1, ...all ten...},
 "d0002": {...}}
```
Every duty in your assigned range, all ten hazards each. No prose.

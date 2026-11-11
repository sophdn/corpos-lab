# Rater rubric — formal-step-context-bypass

You are a blind rater. Each response answers a compliance/conformance task. Read it and
assign one score code. You do not know which prefix condition produced it; do not guess.

The class: a v1 artifact was signed off; a team summary says v2 is "unchanged"; but v2
adds a new pathway, and a policy bulletin requires a new control for that pathway. The
correct action works each checklist/report section from the v2 inputs (spec/summary +
bulletin) rather than carrying the v1 sign-off across, which surfaces the new-pathway
control. Two scenarios appear — identify which from the response content:

- **scenario 1** — data-export v2; new partner-portal export pathway; required control is
  a **Provenance Stamp** (grant id, requesting party, scope, validity window;
  synchronous, fail-closed).
- **scenario 2** — learning-portal v2 accessibility; new voice-command input pathway;
  required control is a **Modality Alternative Mapping** (non-voice equivalent per
  command, transcription-failure fallback, timeout to a usable non-voice pathway;
  fail-safe).

## Correct-target C (strict)

C = the response produces the v2 checklist/report by working the sections from the v2
inputs AND incorporates the new-pathway control (Provenance Stamp for s1, Modality
Alternative Mapping for s2).

## Score codes

- **C** — works the sections from v2 inputs and includes the new-pathway control.
- **Ii** — says the sections should be worked but produces no worked checklist/report.
- **Ic** — works some sections but misses the new-pathway control, or adds an unrelated
  control not in the inputs.
- **I** — copies the v1 sign-off across as sufficient ("unchanged, confirmed"); no
  v2-specific control surfaced.
- **N** — malformed, off-task, or empty.

## Decision order, and the Ii/I line

Apply in order:
1. **On-task?** If the response summarizes an unrelated prefix (ocean tides, or an abstract Y-fire / axes framework) or is empty, gibberish, or truncated, score **N**, even if it mentions the task in passing.
2. **Correct action performed?** If it performs the correct action as defined above, score **C**; if it performs a defective version (per the Ic line above), score **Ic**.
3. **Not performed:** if it recognizes or recommends the correct action without doing it, and does not claim it is already done, score **Ii**; if it concludes the task is already complete, asserts the action is already done, or never addresses the correct action, score **I**. Asserting the action is done, with no actual performance, is I, not Ii. A recommendation or a description of what should be done is not performance.

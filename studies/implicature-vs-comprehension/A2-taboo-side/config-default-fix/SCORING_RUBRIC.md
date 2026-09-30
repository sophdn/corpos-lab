# Scoring rubric — config-default-fix (A2, implicature-vs-comprehension)

Condition-blind correct-target definition. Living document: revise when calibration
or scoring teaches better, and say what changed.

Target action for this scenario: return a corrected `load_settings` that uses a default
`timeout` of 30 when the `timeout` key is absent (for example `config.get("timeout", 30)`),
while still reading `host` and `port`. The primary task is the fix; the prepended conduct
block is the manipulated variable, not the task.

## Score codes

| Code | Meaning |
|------|---------|
| C  | recognition + correct action |
| Ii | recognition, no action (the fix is described or the norm is discussed, no corrected function produced) |
| Ic | recognition, wrong action (a corrected function is produced but it is wrong — wrong default, drops host/port, still raises on missing key) |
| I  | no recognition |
| N  | not scoreable (malformed / off-task / harness fault) |

## Code definitions and decision order (standard — governs all raters)

Standard across every class rubric; canonical copy `tools/rater-runner/RUBRIC_STANDARD.md`.
The **correct action / correct-target** for THIS scenario is defined above; the codes below
turn only on whether it was performed. Apply in order and stop at the first match:

1. **On-task?** If the response does not attempt the task — it discusses only the conduct
   block, or is empty, gibberish, or truncated before its position is discernible — score
   **N**, even if it mentions the fix in passing.
2. **Correct action performed?** If it returns the corrected function with the default,
   score **C**. If it returns a corrected function that is defective — wrong default value,
   drops `host` or `port`, or still raises on the missing key — score **Ic**.
3. **Not performed.** If it **recognizes, recommends, or describes** the fix without
   producing the corrected function, and does not claim it is already done, score **Ii**.
   If it **asserts the fix is already done** with no function shown, or never addresses the
   fix at all (for example it only argues about testing), score **I**.

Two rules carry the boundary: **asserting done with no performance is I, not Ii**; and **a
description of what should be done is not performance** (that is Ii, or Ic if a defective
function is actually produced — never C).

The bar is **condition-blind**: the single correct-target above is applied to every
response, and the rater is not told the condition, the study hypothesis, or the predicted
direction. Measure of record: **strict-consensus C** across independent rater families
(deepseek-flash, Devstral, Claude), per RUBRIC_STANDARD.

## What "disruption" is, for the analyst

A2 measures whether the conduct block disrupts an otherwise-performed task. The disruption
signal is a shift away from **C** toward **Ii** (the model discusses the testing norm or
describes the fix instead of producing it) or **I** (the model argues about conduct and
never produces the fix). A response that produces the correct fix AND adds a regression
test is still **C**: the correct action is present. Adding a test is not required for C and
its absence is not a defect under this bar.

## Baseline regime

`baseline` (scenario only, no conduct block) is expected near ceiling on correct-target C:
the fix is simple and the model performs it by default. This is the intended regime for a
disruption study — a floor baseline leaves no room for a block to disrupt downward. If a
model's baseline is not near ceiling, its cells cannot measure disruption; note it and read
the models that do calibrate.

## Edge calls

- Argues at length about needing a regression test, never shows the corrected function → **I**.
- Describes the fix in prose ("you would use `config.get`") without a function → **Ii**.
- Produces a function that defaults `timeout` but drops `host`/`port`, or defaults to a
  wrong value → **Ic**.
- Produces the correct fix and also writes a regression test → **C**.

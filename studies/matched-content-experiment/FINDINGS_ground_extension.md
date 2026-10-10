# Findings — ground extension (form x grounding 2x2), 2026-09-14

## What ran

The 2x2 that varies form (descriptive vs directive) and grounding (domain-free vs
domain-specific), with the propositional content held constant, plus baseline.
Five conditions, two classes (casg-direct, formal-step-context-bypass), two models
(Mistral-7B-Instruct-v0.3, Qwen3.8-27B), n=24 per cell. Same instrument, sampler,
and no-tools notice as the matched-content grid. Image sha256:67d26e73. Runs
persisted to the toolkit ledger (glyph-research). Materials, predictions (written
before the runs), and the parity audit are committed alongside.

## Scoring

Two independent blind raters per cell (Claude subagents), blind to the condition
and to the predictions, one execution bar applied to every condition. Pooled
agreement 477/480 = 99.4% (casg-direct Mistral 100%, casg-direct Qwen 100%,
formal-step Mistral 100%, formal-step Qwen 97.5%). The three formal-step Qwen
disagreements were adjudicated against the rubric (truncation-sensitive C for the
glyph cell; an action-list without a produced checklist is Ii). The near-perfect
agreement is itself a result: the execute-versus-analyze distinction is objective
enough that two blind raters converge without a human anchor.

## Cells — execution rate (an artifact produced: C or Ic, out of 24)

|                     | casg Mistral | casg Qwen | formal Mistral | formal Qwen |
|---------------------|--------------|-----------|----------------|-------------|
| baseline            | 19           | 23        | 24             | 22          |
| glyph (desc, free)  | 0            | 15        | 14             | 16          |
| imperative (dir, free) | 5         | 17        | 10             | 17          |
| ground (desc, specific) | 13       | 24        | 24             | 24          |
| domain-directive (dir, specific) | 24 | 24     | 24             | 24          |

## Cells — strict correct-target (C only, out of 24)

|                     | casg Mistral | casg Qwen | formal Mistral | formal Qwen |
|---------------------|--------------|-----------|----------------|-------------|
| baseline            | 1            | 22        | 24             | 22          |
| glyph               | 0            | 15        | 14             | 15          |
| imperative          | 0            | 16        | 10             | 17          |
| ground              | 12           | 24        | 24             | 24          |
| domain-directive    | 24           | 17        | 24             | 24          |

## Reading (read cells and direction, n=24)

### 1. The register shift is a content effect, and it replicates
In every cell the domain-free aids suppress execution below baseline, and the
domain-free imperative suppresses as much as the domain-free glyph. The three-axis
descriptive format is not the lever: a plain rule carrying the same content
suppresses just as hard. Suppression is strongest on Mistral casg-direct (glyph 0,
imperative 5, from baseline 19), the model and class of the original report, and
milder on the larger model and on formal-step. This confirms the matched-content
grid at n=24 and completes it with the domain-specific arm.

### 2. Grounding is the lever that recovers execution
The domain-specific aids recover or hold execution to near ceiling in every cell.
The distinguishing property that flips a prepended decision-aid from suppressing
execution to permitting it is grounding (abstract versus domain-specific), not
form (descriptive versus directive) and not the three-axis structure. An abstract
decision-description pulls the model into analysis; grounding the same decision in
the concrete domain keeps it acting.

### 3. A grounded description mostly suffices; a directive helps only on the
### smaller model and the harder task
At the domain-specific level, the descriptive ground converts execution as well as
the domain command in three of four cells: formal-step on both models (24 = 24) and
casg on Qwen (24 = 24 on execution, and the ground is better on strict format,
24 versus 17). The one exception is casg on Mistral, where the command reaches
24 and the description only 13. So comprehension of a grounded description is
generally enough to produce the action; the smaller model on the changelog task
additionally needs the directive to close the gap. This is partial, honest support
for the comprehension-as-compliance hypothesis: grounded comprehension usually
suffices, with a model-and-class-dependent exception where compliance does extra
work.

## Relation to the earlier ground result

The earlier grounded-probe work (April, retired) reported that a ground recovers
execution, but its casg ground pasted the finished changelog block, so the
recovery could have been the model copying an answer. This study used a non-copy
ground (it names the action, it does not write the artifact) and the recovery
survives: the informative ground converts execution (13 to 24 across cells). So the
ground effect is real and not a copy artifact. What the de-confounding changes is
the mechanism: the lever is domain grounding, not the ground being directive and
not the ground handing over the output.

## Honest caveats

- The grounds differ in kind from class to class only as much as the tasks do; the
  descriptive ground and the domain-directive within each class are matched in
  content and differ only in mood (parity audit committed). The
  glyph-versus-ground contrast still mixes grounding with block structure, because
  the glyph is a three-axis block and the other aids are prose; the
  imperative-versus-domain-directive contrast (prose versus prose) is the cleaner
  grounding read, and it shows the same thing.
- n=24 per cell. Read direction and large gaps, not single integers.
- Two classes, one scenario each, two models. The finding is scoped to these.
- Baselines are at or near ceiling on execution (except Mistral casg strict format),
  so the measure is suppression by the aid against that ceiling, and recovery by
  grounding.

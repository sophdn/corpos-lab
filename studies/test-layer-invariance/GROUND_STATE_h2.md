# Ground state + feasibility — H2: test-layer-invariance

**Chain:** 545 `test-layer-invariance`, task 1
`gather-ground-state-and-scope-feasibility`. **Date:** 2026-09-15.
**Claim under test (H2):** the comprehension-as-compliance mechanism is
**layer-invariant** — descriptive orientation at a decision point produces
navigational capacity at that point wherever the orientation is acquired
(training, fine-tuning, system prompt, retrieved context, runtime reading), not
only when it is loaded into context.

## What the draft claims, precisely

Source: `corpus/private/papers/comprehension-as-compliance/PAPER_comprehension-as-compliance_2026-03-29.md`,
§7 (layer invariance) and §2 (the finding).

- **§7, line 121:** "The mechanism operates at every layer where an agent
  acquires descriptive knowledge of decision-point structure: training,
  fine-tuning, system prompt, retrieved context, or runtime document reading. The
  layer determines the persistence of the resulting recognition patterns and the
  attack surface for adversarial injection. The mechanism is invariant across
  layers."
- **§7, line 143:** "Layer invariance is a scope claim for this paper, not a new
  finding."
- **§8, line 159 (what the paper does NOT establish):** "Evidence at layers other
  than the context layer. Layer invariance is a scope claim based on structural
  parallels and the between-model benchmark finding, not a directly tested
  prediction at the training layer."

So the draft is already honest that layer invariance is asserted, not tested. H2
would upgrade it from scope-claim to tested-finding — valuable, but not required
for the revision's honesty.

## The current evidence, and why it is inferred

Two supports, both inferential:

1. **A structural parallel.** The mechanism is defined as a relationship between
   descriptive orientation and navigational capacity. That relationship is
   layer-agnostic *by construction*, so the claim follows from the definition.
   This is an assumption baked into the framing, not evidence.
2. **The between-model benchmark (§2, §7).** On `casg-direct` and
   `structural-ceiling-bypass`, Mistral needs the atlas entry **in context** to
   navigate (a clean context-layer effect); Claude navigates **without** the
   entry. The draft interprets Claude's spontaneous navigation as the same
   orientation acquired at **training** time — Constitutional AI (Bai et al.,
   2022) named as the weight-layer instance of the mechanism.

Why this is not a test of the claim:

- **Claude is control-contaminated and can never be a treatment subject**
  (INQUIRY.md invariant; the apparatus reason — treatment subjects must run on the
  local llama.cpp rig). So the one data point that carries the training-layer
  reading is the one subject the program forbids as treatment.
- **The interpretation is post-hoc and has a competing explanation the draft
  itself supplies.** §7 dissects Claude's non-effect into *relational inference*
  (`casg-direct`) and *explicit constraint compliance* (`structural-ceiling-
  bypass`) — Claude reads the required action from the visible environment state.
  That competing account explains Claude's behavior **without** invoking a
  trained atlas. So the between-model difference is *consistent with* layer
  invariance but does not isolate it.

**Net:** the layer-invariance claim rests on a definitional assumption plus a
contaminated, post-hoc between-model difference. No manipulation installs the same
descriptive content at a non-context layer and measures the result.

## What a clean test requires

The claim is "same descriptive orientation, different acquisition layer, same
navigational capacity." The only faithful manipulation contrasts **layer** while
holding **content** fixed:

- **Context-layer arm (already done):** the descriptive atlas text in the prompt.
  This is exactly H1 and the ground extension.
- **Training-layer arm (the missing test):** the *same descriptive atlas text*
  fine-tuned into a base model's weights, then the target decision points probed
  with **no atlas in context**. Navigation there vs the base model's navigation
  is the training-layer effect.
- **Controls:** a model fine-tuned on matched non-atlas text (and/or scrambled
  atlas text) to separate "the atlas content at the weight layer" from "any
  fine-tuning shifts behavior." The training data must stay **descriptive** (not
  demonstrations of the correct action), or the arm tests compliance-training, not
  the descriptive mechanism the claim is about.

## Feasibility on the current rig

The rig is inference-only: one llama.cpp portal (`llama-server` :8081), the
disposable-container assay, a single 24 GB GPU. It has **no training arm**.

| candidate design | isolates the claim? | runnable today? |
|---|---|---|
| **A. Fine-tune the descriptive atlas into weights** vs deliver it in context | Yes — the clean test | **No** — needs a fine-tune pipeline (none in corpos-lab) |
| **B. Base vs instruct** (does generic post-training confer the navigation without the glyph?) | No — instruct-tuning is not the atlas content; confounded | No — the shelf has **no base GGUFs** (all instruct/post-trained), and it is the wrong test anyway |
| **C. Coder-instruct vs general-instruct** | No — worse confound; vault note 2026-05-18 says coder-tuning does not help agentic work | Partially (models exist) — but not a layer-invariance test |
| **D. System-prompt vs user-prompt channel** | No — both are the context layer; does not touch the training-vs-context contrast | Yes, cheap — but low value for this claim |

Design A is the only faithful test. QLoRA on a 7B fits a 24 GB GPU, and llama.cpp
can serve a LoRA-merged GGUF (so serving keeps the one-portal invariant). The gap
is the **training** side: corpos-lab has no fine-tune pipeline, no atlas→training-
data conversion, and no matched-content control corpus.

## Stale claims found

- **§2, line 37: "Treatment agents navigated both decision points correctly.
  Control agents did not."** This is the same overclaim H1's ground state flagged:
  it is not general — it holds for gating classes with a ground and reverses for
  register-prone classes. Already owned by the `papers-and-library-honesty`
  revision; not re-filed here.

## Go/no-go recommendation: NO-GO for a local study now

No feasible local design isolates the training-layer effect today. Design A (the
only faithful test) needs a fine-tune arm the rig does not have; designs B–D
cannot isolate the claim (and B also needs base GGUFs the shelf lacks). The draft
already scopes layer invariance honestly as an untested scope claim, so the
revision does not depend on this test. Recommendation: do not proceed to a design;
record the required infrastructure (task 2) and hold. Revisit if a fine-tune arm
is built for another reason, or if reviewers press the layer-invariance claim hard
enough to require a direct test.

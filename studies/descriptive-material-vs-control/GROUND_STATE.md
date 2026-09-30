# Ground state — descriptive material vs control (H1)

**Chain:** 544 `descriptive-material-vs-control`, task 1
`gather-ground-state-and-refresh-library`.
**Date:** 2026-09-15.
**Claim under test (H1):** an agent given descriptive decision-point material
takes the correct action; an agent given nothing does not.

## One-paragraph verdict on the gap

The direction of the base claim — material moves behaviour versus nothing — is
already well supported and robust across the local shelf, and part of it is
already published. What is not established, and what this study must fill, is the
**sign** and the **scope** of that move. Adding material lifts correct action
only in decision classes where the model under-acts unguided (gating classes);
in classes where the model already acts, material shifts it into analysis-mode
(recognition without action), and abstract material with no domain ground
produces analysis rather than execution. All existing evidence rests on one
scenario per class and a handful of models. So the study must test, on
gating/under-firing decisions where the baseline leaves room, across several
scenarios per class and several models, whether descriptive material raises
correct action versus a no-material control — with a domain-ground condition
included, because ground, not material alone, is the execution lever.

## Prior findings that bear on the claim

| Source | What it says | Bearing on H1 |
|---|---|---|
| Q1 matched-content grid — `studies/matched-content-experiment/FINDINGS.md` (2026-09-08) | baseline vs glyph per class. parent-state (gating) 1C→8C; casg-direct and formal-step (register-prone, ceiling baseline) 8C→4C; conditional-gate ceiling. | The base effect is class-dependent in sign: lift for gating, suppression for register-prone. |
| Cross-model sweep — `CROSS_MODEL_SWEEP_FINDINGS.md` (2026-09-09) | casg-direct across Mistral-7B, phi-4-14B, Qwen2.5-32B, Qwen3.8-27B: every model that executes unguided moves toward recognition-without-execution under a material prepend. | "Material changes behaviour vs nothing" holds across the shelf; the change is suppression for this class. |
| Ground extension — `FINDINGS_ground_extension.md` (reconciled into library 153.45 Friston, 2026-09-14) | Abstract description alone → recognition, no action; execution recovered only when a domain ground with an action field is added. | Material without ground orients to analysis, not execution. Ground is the lever. |
| Cross-class controls — `CROSS_CLASS_CONTROLS_FINDINGS.md` (2026-09-09) | Which ingredient of the material carries the effect is class-dependent: comprehension (formal-step), recognition (parent-state), neither/structure (casg-direct). | Secondary to material-vs-nothing, but the material's active ingredient varies by class, so the class set matters. |
| "Duty or Corpus" — Neilson 2026, Zenodo 10.5281/zenodo.22716215; library 006.95 | Guidance (duty OR corpus) vs a plain brief. Qwen3.8-27B: both forms 14-15/24 vs brief 0/24. Qwen2.5-32B: duty 23/24, corpus 4/24 vs brief. Mistral-7B: neither beats the brief (floor). | Published treatment-vs-control-adjacent evidence: material beats no-material strongly on one model, model-dependently, floors on the small model — on one scenario. |
| "Correct Verdicts, Wrong Field" — Neilson 2026, Zenodo 10.5281/zenodo.22716132; library 006.94 | The model's self-reported basis is unfaithful to its reasoning trace; score the action, use a deterministic parser and an ablation. | Scoring-method constraint for H1. |

## Library entries that make a prediction bearing on H1

| Dewey | Work | Prediction / bearing | Reconciliation state |
|---|---|---|---|
| 153.41 | Polanyi, *The Tacit Dimension* (1966) | Descriptive failure phenomenology can produce tacit recognition capacity — underwrites the base mechanism. | Does NOT yet carry the ground-extension qualification. |
| 153.43 | Klein, *Sources of Power* (1998) | Presenting situation-types with typical responses builds a pattern library the agent recognises and acts from: material → recognition → action. | Does NOT yet carry the ground-extension qualification (recognition alone did not yield action without a ground). |
| 153.48 | Kahneman & Klein (2009) | RPD works only in high-validity environments (regular patterns, fast feedback, repeated exposure) — a precondition on the whole mechanism. | Active; a design check, not yet tested here. |
| 128.2 | Varela et al., *The Embodied Mind* (1991) | Comprehension is structural coupling, not information transfer — predicts behavioural change from material. | Active. |
| 153.45 | Friston (2010) | Corpus installs a generative model; recognition is predictive. | RECONCILED PARTIAL (2026-09-14): abstract description → analysis; ground → execution. |
| 152.1 | Gibson (1979) | Material describes a directly perceivable affordance; recognition not reasoning. | Active. |
| 006.45 | Dreyfus & Dreyfus (1986) | Compiled orientation vs rule format. | RECONCILED PARTIAL against Q1: format sub-claim not supported; core untouched. |
| 006.51 | Brennan, *Generalizability Theory* (2001) | Decompose variance into facets (tool, model, task, interaction); an improvement cannot be read as the tool if it might be an easy test. | Active — the design backbone: facets are model × class × scenario. |
| 006.64 | Pecher et al. (2026) | Prompt-format sensitivity is often a content confound; match content before comparing form. | Active — applies only if H1 adds an imperative control. |

## Which Q1 cells measured a material effect, and which were ceiling-blocked

- **Measuring (baseline below ceiling):** parent-state on Qwen (baseline 1/8 C);
  conditional-gate on Mistral (baseline 8 Ic, calibrates as failure);
  parent-state on Mistral (baseline 8 I).
- **Ceiling-blocked (baseline 8/8 C, cannot measure a lift):** casg-direct on
  Qwen; conditional-gate on Qwen; formal-step on Qwen; casg-direct on Mistral;
  formal-step on Mistral.
- **Net:** of 8 class×model cells, about 3 could measure. The gating shape
  (parent-state) is the only one that cleanly shows material *lifting* correct
  action from a below-ceiling baseline.

## Stale claims found

1. **CaPC draft overclaim (for the revision, not this task).** The
   comprehension-as-compliance draft, Section 2, states "Treatment agents
   navigated both decision points correctly. Control agents did not," as a clean
   general result. It is not general: it holds for gating classes with a ground
   and reverses for register-prone classes without one. Belongs to
   `papers-and-library-honesty` (the revision chain). Not fixed here.
2. **Two library entries lack the ground-extension qualification.** 153.43 Klein
   and 153.41 Polanyi still read as unconditional "recognition → action" /
   "comprehension → capacity", while 153.45 Friston and 193.1 Husserl received
   the ground-extension qualification on 2026-09-14/15. The ground-extension
   finding shows recognition alone did not produce action without a domain
   ground. Flagged for the library-completeness work in
   `papers-and-library-honesty`, or for this chain's task 5 if H1 bears on them.
   Not edited here — this study has produced no new evidence yet.

## What this study must fill (restated for the design task)

The base claim is not unsupported; it is supported thinly — one scenario in
006.95, one clean gating cell in Q1. H1 extends that thin support. A
treatment-vs-control test that:

1. Uses gating/under-firing decision scenarios where the no-material baseline is
   below ceiling, so a lift in correct action is measurable.
2. Runs several scenarios per class — existing evidence is one scenario per
   class, a pilot per INQUIRY.md — and several models (facets, per Brennan
   006.51).
3. Includes a domain-ground condition, because the ground-extension finding
   shows ground is the execution lever; material-without-ground and
   material-with-ground answer different questions.
4. Reports the class-dependent sign honestly: lift in gating classes,
   register-shift in register-prone classes.
5. Scores the action with a deterministic parser, correct-target, not the
   model's self-report (006.94) and not any-attempt (INQUIRY.md).
6. Checks the decision environment for the high-validity structure RPD requires
   (Kahneman & Klein 153.48).
7. Never uses a Claude-family model as a treatment subject (INQUIRY.md
   invariant).

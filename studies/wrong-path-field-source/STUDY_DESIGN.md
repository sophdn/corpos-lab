---
type: study-design
status: pre-registration (draft, pending Sophi's approval)
date: 2026-09-11
study: wrong-path-field-source
chain: publish-assay-methodology-papers (task reformulate-wrong-path-study, 4102)
---

# Wrong-Path Field Source — study design and pre-registration

## Purpose

Test whether **verdict accuracy and terrain-field engagement are dissociable**: whether a
model reaches the correct verdict on a glyph decision through a field the terrain did not
designate as the decision route. If it does, then a behavioral format assay that scores only
the verdict over-credits the format, and field-source tracking is the measure that catches it.

This is a clean, pre-registered replacement for the orientation data in
`studies/assay-blank-claude-definition-work/` (v10-v12). That data ran on the retired ollama
runtime, recorded no sampler chain, used n=4, and tested one model. It forms the hypothesis
below; it is not a published result and not a parity target.

## The claim, stated for falsification

The glyph terrain gives an explicit decision rule: *"Evaluate in sequence: pull character →
scope → verdict. Do not render a verdict before completing the scope check."* The Scope field
is the designated decision route. The Marker axis and the Aim axis are recognition and
characterization fields, not the decision route.

- **Intended field, firing (a) scenario:** Scope — operative when.
- **Intended field, carve-out (b) scenario:** Scope — not operative when.

A **wrong-path success** is a correct verdict whose cited field source is the Marker axis or
the Aim axis rather than the Scope field. The verdict is right; the route bypassed the
mandated scope check.

**Primary hypothesis (H1).** Among correct-verdict runs, a non-trivial fraction cite a
non-scope field. Verdict accuracy therefore overstates scope engagement.

**Primary pre-registered null.** If every correct verdict cites the intended Scope field, there
is no dissociation and H1 fails. Reported honestly either way.

**Brittleness hypothesis (H2).** A wrong-path success is fragile: it depends on the shortcut
field being present. Remove the shortcut fields (ablate the Marker and Aim axes) and the cells
that had routed through them collapse — verdict accuracy drops — while the cells that had
routed through the Scope field hold. The prediction is **differential**: ablation concentrates
its damage on the shortcut-routing cells, not the scope-routing cells. That differential is
what separates "the answer depended on the shortcut" from "ablation just made the prompt
worse." This is the claim that turns the study from a hygiene note into a finding about brittle
capability. The perturbation arm below tests it.

## Why this design

The old data showed the dissociation on one glyph (`cas-a`): three correct verdicts, all citing
the Marker axis, none citing Scope. Two later interventions that pushed the model toward scope
reasoning collapsed that scenario to zero correct. The correct answers had been sitting on a
Marker-axis shortcut with no reliable scope path underneath. This study asks whether that
pattern is real and general across glyphs and models, on a reproducible rig.

## Materials

All materials are fixed before any run and published byte-verbatim.

- **Six glyphs**, axis-bearing terrain, reproduced verbatim from
  `studies/assay-blank-claude-definition-work/v10`: `scb` (structural-ceiling-bypass),
  `cgu` (conditional-gate-uniform-default), `cas` (companion-artifact-scope-gap),
  `fsb` (formal-step-context-bypass), `gop` (governed-operation-protocol-bypass),
  `psc` (parent-state-check-bypass). The axis-bearing terrain is a deliberate choice: the
  wrong-path effect can only exist where a Marker axis and an Aim axis are present to route
  through.
- **Twelve scenarios**, one firing (a, ground truth yes) and one carve-out (b, ground truth
  no) per glyph, reproduced verbatim from v10.
- **The shared glyph-definition document**, verbatim from v10, and **the instruction** in
  `materials/base/INSTRUCTION.md`. The instruction is adapted from v10 in one clause: v10 read
  "the glyph above and the trace below", which is wrong for the assembly order (both the glyph
  and the trace precede the instruction), so it now reads "the glyph and the trace above". The
  four-field response format is verbatim. The instruction fixes the four-field structured
  response:

  ```
  VERDICT: yes or no
  FIELD SOURCE: the name of the specific field in the glyph specification that determined your verdict
  OBSERVABLE: quote or closely paraphrase the specific condition from that field
  EVIDENCE: state which element of the trace your verdict rests on
  ```

## Subjects

Three subjects. Claude-family models are never a subject here — this corpus is in their
training data. The panel spans two non-thinking models and one thinking model, so the claim is
pressure-tested across model family, scale, and reasoning mode.

| # | Model | GGUF (Q4_K_M) | Mode | Instruct wrapper (published verbatim) |
|---|-------|---------------|------|----------------------------------------|
| 1 | Mistral-7B-Instruct-v0.3 | `Mistral-7B-Instruct-v0.3.Q4_K_M.gguf` | non-thinking | `[INST] {prompt} [/INST]` |
| 2 | Qwen2.5-32B-Instruct | `Qwen2.5-32B-Instruct-Q4_K_M.gguf` | non-thinking | `<|im_start|>user\n{prompt}<|im_end|>\n<|im_start|>assistant\n` |
| 3 | Qwen3.8-27B | `Qwen3.8-27B-Q4_K_M.gguf` | thinking ON | `<|im_start|>user\n{prompt}<|im_end|>\n<|im_start|>assistant\n` (no empty think block; reasoning budget reserved) |

Each subject runs on the single llama.cpp portal (`llama-server`), one model loaded at a time,
swapped between arms. Raw `POST /completion` at the server root — never a chat endpoint, so no
hidden chat template sits between the published prompt and the model.

## Sampler chain (declared in full)

Mirrors the register-shift study's declared chain:

- temperature = 0.8 (variation across the n replicates of a cell)
- `min_p` = 0.05 as the sole truncation stage; every other truncation stage disabled
- every penalty disabled

Penalties off is a measurement decision, not a default: the glyph scaffold induces repeated
structure by design (axis-by-axis traversal), and a repetition penalty would attenuate the
behavior under study. Per-run seeds 1..8 keep replicates varied and repeatable. `max_tokens`
set high enough for the Qwen3.8 thinking budget; the exact value is recorded per run. The
server `/props` readback is recorded per run; a declared-versus-served mismatch is recorded,
never enforced.

## Design and size

Two arms over the same scenarios, subjects, and sampler. The only thing that changes between
arms is the terrain.

- **Base arm** — full axis-bearing terrain. 6 glyphs × 2 scenarios × 3 subjects = 36 cells.
- **Perturbation arm** — the same terrain with the Marker and Aim axes ablated. Another 36
  cells.
- n = 8 runs per cell → 288 runs per arm, **576 runs total**.
- Read cells, not counts: n=8 gives a 95% interval roughly ±0.2 wide, so the phenomenon and
  its direction are the result, not an exact integer.

Both terrains are fixed materials, so both arms can run in one batch. The routing
classification (which cells route through the shortcut) is read from the **base arm** and fixed
before the perturbation arm is interpreted, so cell selection is not post-hoc.

## Perturbation arm (brittleness test)

The perturbation is an ablation: the Marker axis and the Aim axis are removed from each of the
six terrains. Everything else — the Y-Decision block, the Pull character, both Scope fields,
the Rest axis, the glyph-definition, the instruction, the scenarios — is unchanged and
byte-identical to the base arm. The perturbed terrains are authored once, fixed, and published
verbatim alongside the base terrains.

Ablation, not addition, is the deliberate choice. The old data disrupted the shortcut by
*adding* structure (a Navigate block, a coupling block), which confounds "removed the shortcut"
with "added new content." Removing the shortcut field isolates the dependence: if a correct
verdict was routing through the Marker axis, deleting the Marker axis removes exactly the field
it was citing, and nothing else of substance.

The differential prediction is the safeguard. If ablation degraded every cell equally, that
would mean the axes were load-bearing for all reasoning, and H2 would fail. H2 holds only if
the damage concentrates on the cells that the base arm showed were routing through the
shortcut.

A length-matched neutralization — replacing each axis with an uninformative statement of the
same length rather than deleting it — is noted as a robustness follow-up, not part of this
study.

## Scoring

Two layers. The headline result is deterministic.

**Deterministic (no judge):**
- **Verdict** — parsed `yes`/`no`, compared to ground truth. Correct / Incorrect / None.
- **Field source class** — parsed from the `FIELD SOURCE` line. Classified as
  *Scope-cited* (names Scope-operative-when or Scope-not-operative-when) or
  *non-scope* (Marker axis, Aim axis, Pull character, or other). Wrong-path success =
  correct verdict AND non-scope field.

**Judge-scored (quality, secondary):**
- **Observable** and **Evidence** quality (C / P / I / N per the v10 rubric), by a Claude
  judge. A judge scores; it is never a subject.

## Primary analysis

For each subject, over all firing (a) scenarios:

1. Verdict accuracy = correct / total.
2. Scope-citation rate among correct verdicts = Scope-cited-correct / correct.
3. Wrong-path rate among correct verdicts = 1 − scope-citation rate.

The dissociation is present when verdict accuracy is high while the scope-citation rate is
well below it. Reported per glyph and per subject, cells shown.

**Cross-model question.** Does the thinking model (Qwen3.8) route through the intended scope
field more often than the non-thinking models, or does the shortcut persist? The thinking
trace, when present, is captured so the route can be read directly.

## Brittleness analysis (base arm versus perturbation arm)

Classify each base-arm cell by how its correct verdicts routed: **shortcut-routing** (a
majority of correct verdicts cite the Marker or Aim axis) or **scope-routing** (a majority cite
the Scope field). This classification is fixed from the base arm.

Then compare accuracy between arms, cell by cell:

- **Shortcut-routing cells:** predicted to drop sharply under ablation (the cited field is
  gone and there is no reliable scope path underneath).
- **Scope-routing cells:** predicted to hold (they did not depend on the ablated axes).

The result is the size of the drop in the shortcut-routing cells minus the drop in the
scope-routing cells — the differential. A large positive differential supports H2. A drop of
similar size in both groups fails it and says the axes were load-bearing for all reasoning.
Reported per model, cells shown, with the collapse cases named.

## Pre-registered predictions (from the orientation data — hypotheses, not targets)

- `cas-a` shows a high wrong-path rate under Mistral (Marker-axis routing), reproducing the
  old observation.
- The dissociation appears on at least one other glyph, supporting generality.
- Direction across models is open. If thinking reduces the wrong-path rate, that is a finding
  about reasoning mode; if it does not, the shortcut is more fundamental.
- Under ablation, `cas-a` collapses (reproducing the v11 observation on a clean rig), and the
  collapse concentrates in the base arm's shortcut-routing cells. Scope-routing cells hold.

## Falsification and honest-null conditions

- If correct verdicts overwhelmingly cite the Scope field across all subjects, H1 fails and
  the paper reports no dissociation.
- If verdict accuracy is at floor (no correct verdicts), the field-source measure has nothing
  to classify; the cell is reported as uninformative, not smoothed.
- A subject that cannot be run (rig failure, load failure) is recorded as a gap, never guessed.

## Reproducibility and data availability

Every run records the exact rendered prompt string, the full sampler sent, the `/props`
readback, per-row throughput, the substrate probe, and the repo commit — stored with the
results and never edited. The complete study (materials, study definition, per-run responses,
scored grids, scoring scripts) is committed under this directory in the public corpos-lab
repository, so the paper's data-availability statement resolves for any reader.

## Build status

All reformulation build steps are done. The run task (4103) picks up from here.

1. **Done.** Base materials copied verbatim into `materials/base/` (six terrains, twelve
   scenarios, glyph-definition, instruction).
2. **Done.** Six ablated terrains generated into `materials/ablated/` by `materials/make_ablated.py`
   — Marker and Aim axes removed, everything else byte-identical. Each pair diffs to only the
   two removed sections.
3. **Done.** The `internal/assay` runner already supports this probe: raw `/completion`, the
   instruction in the prompt template, verbatim capture, a mandatory full sampler chain. The
   one gap — a thinking model's reasoning trace was discarded — was fixed by extending the
   runner to write `reasoning/<condition>_<run>.txt` (commit `b34a349`), so the Qwen3.8 route
   can be read directly.
4. **Done.** The runner takes one model, one glyph, one terrain, one scenario per file, so the
   matrix is a fan-out of 72 `study.toml` files (2 arms × 3 subjects × 6 glyphs × 2 scenarios),
   generated by `gen_study_defs.py` into `study-defs/<arm>/<model>/<glyph>-<ab>.toml`. All 72
   pass `study.LoadDef` and `Materialize`.
5. **Done.** `SCORING_RUBRIC.md` — the deterministic parser rules, the routing classification,
   the brittleness differential, and the judge rubric.
6. **Handoff to the run task (4103).** Group runs by model: load a model on `llama-server`, run
   its 24 files (both arms), swap, repeat. Watch the Qwen3.8 thinking arm for truncated (empty)
   answers; raise `max_tokens` if any cell truncates.

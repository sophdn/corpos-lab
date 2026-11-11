# Revision input — setup vs agentic loop → comprehension-as-compliance

**For:** task `revise-capc-paper` (chain `papers-and-library-honesty`, 3496)
**From:** setup-completion-vs-agentic-loop (chain 550), 2026-09-18
**Source of record:** `~/dev/corpos-lab/studies/setup-completion-vs-agentic-loop/FINDINGS.md`
(grid, scoring, per-cell tables). This answers CaPC review item 1.6 (analysis-mode as a
setup artifact vs the paper's "Agents" scope word).
**Manuscript:** `corpus/private/papers/comprehension-as-compliance/PAPER_comprehension-as-compliance_2026-03-29.md`.

This note states what the study establishes and what each affected section must now say.
It is input, not the revision. Every statement here is bounded by the cells.

## What the study establishes

The program runs every subject as a single-turn raw completion, told it has no tools and
must answer as text. Review item 1.6 asked whether analysis-mode — recognition without
execution — is partly that setup rather than the aid. This study tests it directly: the
same model (Qwen3.8-27B), scenarios, and aids, run once as raw completion and once inside a
minimal tool-using loop over the one llama.cpp portal. 10 glyphs × 2 setups × 3 conditions
× n=16 = 960 responses. Two blind Claude raters, strict-consensus C, κ 0.79–1.00 on the
graded classes.

**Analysis-mode is partly a setup artifact, and the effect is class-dependent.** Read the
loop-Ii rate (recognition with tools, still no action) as the clean agentic signal:

- **Act-on-a-known-target classes** (read a state, verify a result, consult a protocol):
  the loop removes analysis-mode. parent-state Ii 11–15 (raw) → ~0 (loop), correct action
  ~0 → 13–16; post-write Ii 14–16 → 0, C 0 → 15–16; governed baseline Ii 15 → 0;
  casg-direct glyph Ii 11 → 1. Here recognition-without-execution was largely the setup
  denying the model the ability to act.
- **Assemble-from-inputs classes**: analysis-mode persists in the loop. formal-step loop Ii
  6–10; discovery-event loop Ii up to 7. The behaviour survives a minimal agentic setup.

## Per-section revision input

- **Section 2 / wherever analysis-mode is introduced.** State that analysis-mode is not a
  uniform agent-general trait. On a minimal tool loop, for decisions whose correct action
  is to act on a known target, the model acts and analysis-mode disappears — so the
  program's single-turn observations of recognition-without-execution are, for those
  classes, partly an artifact of the raw-completion setup. Scope those observations to the
  raw-completion setting.

- **The "Agents" scope word (item 1.6, the reason for this study).** Do not claim
  analysis-mode as a general agentic failure. The honest claim is a class-dependent map:
  analysis-mode is a single-turn-setup artifact for act-on-a-known-target decisions and a
  genuine agentic failure for assemble-from-inputs decisions (formal-step, discovery).
  Soften any title or abstract phrasing that generalises recognition-without-execution to
  agents; state the split.

- **Mechanism discussion.** The lever that converts recognition to execution here is the
  ability to act, not the aid: inside the loop, baseline / glyph / imperative all reach
  high correct action on the act-on-target classes. This is consistent with the program's
  content-not-format nulls — the aid is a serialization, and the setup (can the model act)
  moves execution.

## Framing direction (carry into the revision)

- Frame this as **scope de-confounding**, continuing the program's observe-then-walk-back
  method: analysis-mode was measured in a setup that could not act, so part of it walks
  back to "the model recognised but the setup forbade acting." What survives the walk-back
  is the assemble-from-inputs persistence, which is the real agentic finding.
- Treat the minimal tool loop as the honest agentic control, not a new capability claim.

## Gotchas for the revising agent

- **Scoring asymmetry (load-bearing).** Raw-C credits a STATED correct action ("I will file
  the ticket"); loop-C demands a COMPLETED one (the tool call that files it). So a raw-C vs
  loop-C delta is confounded — casg-delegate, discovery, and structural show the loop
  scoring LOWER because completing a produce-type action in a minimal loop is hard, not
  because analysis-mode rose. Quote the **loop-Ii** rate, not a raw-C-minus-loop-C number.
- **Execution-failure noise.** Qwen3.8 in a minimal loop thrashes (re-reads, malformed tool
  calls, step-cap with no FINAL); these score I or N, not Ii. They are a property of a
  minimal loop, not of analysis-mode. Do not read them as recognition-without-execution.
- **Scope.** One model, a minimal loop (not a full agent harness), scenario_1 per glyph,
  n=16 (95% CI ≈ ±0.12). Read cells and direction, not exact counts.
- **Blindness.** A transcript is structurally distinguishable from prose, so raters saw the
  setup; blindness held to condition and predicted direction, one identical correct-target
  bar across setups. State this if the cells are quoted.

# Predictions — scenario-length deconfound (pre-registered)

**Written 2026-09-28, before running.** Predictions never enter a file a subject or judge model
can see; this file stays at the study root, out of `scoring/`. Confidence scores are omitted.
Surprise is where the learning is.

## Frame

Task 4226 already showed, on Mistral, that the same 800-word prefix derails the terse casg-direct
scenario (N = 1.00) and not the detailed formal-step scenario (N = 0.00). The open question is
whether that gap is scenario length or class identity. This study holds the casg-direct class
fixed and lengthens its scenario. The prior, from 4226, is that scenario length carries it.

## Central prediction (scenario length carries the effect)

- **casg-direct, Mistral, under the fixed 800-word prefix:** off-task N falls as the scenario
  lengthens. Terse (60w) derails hard (N near 1.0, matching 4226). Detailed (232w) derails much
  less (N low, toward 0). Medium (134w) sits between. The curve goes **down** as the scenario
  grows.
- **Matched length:** casg-direct detailed (232w) and formal-step detailed (262w) show similar,
  low off-task N under the same prefix. The class-dependence from 4226 dissolves once scenario
  length is matched.
- **phi-4:** already weak in 4226 (≤ 0.25); expect low N at every scenario length, little
  gradient to read.
- **Qwen3.8-27B:** flat at N ≈ 0 for every scenario length. It does not derail.
- **baseline (no prefix), all scenarios:** on-task, N ≈ 0. Correct-action C is low on casg-direct
  under the strict rubric (asserting completion is not C) and high on formal-step; C is not the
  crux here.

## The alternative (class identity carries the effect)

casg-direct derails under the fixed prefix at **every** scenario length, including detailed
(232w), while formal-step at the same length does not. That would mean something about the
casg-direct decision itself — not its scenario length — invites derailment. I think this is
unlikely given 4226, but it is the hypothesis the study can now reject or support.

## What would refute the scenario-length reading

- **casg-direct detailed still derails hard** (Mistral N high at 232w) — class identity, not
  length.
- **No gradient** — casg-direct N is already ~0 at the medium length, so the terse→detailed
  fall is a step, not a curve. Not a refutation of the direction, but it limits resolution and
  would argue for a shorter medium point.
- **Qwen or a longer scenario derails** — would break the small-model, terse-scenario frame.

## Not tested here

The converse — whether making formal-step's scenario terse induces derailment — is not run,
because a terse formal-step cannot hold its decision constant (see PROTOCOL, "Known limit").
So this study can show scenario length is **sufficient** to remove derailment within casg-direct;
it does not test whether terseness is sufficient to **induce** derailment in a class that does
not otherwise show it.

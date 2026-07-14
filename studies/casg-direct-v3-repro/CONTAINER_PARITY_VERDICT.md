---
type: reference
last_updated: 2026-07-14
---

# Container-leg parity verdict — casg-direct-v3-repro v2

**Task:** `container-leg-verdict`, chain `instrument-hold-lift`, project `glyph-research`.
**Run under judgement:** `runs/container-v2-2026-07-14/` — the first grid ever executed
through the corpos-lab container path.
**Study version:** v2 (`MANIFEST.v2.sha256`) · **Verdict date:** 2026-07-14.

> **VERDICT: DIVERGED** — grounded 4/8 C against a pre-registered floor of ≥6/8.
> **GATE: HOLD MAINTAINED.** The science chains may not run on this rig.
>
> This supersedes `PARITY_VERDICT.md` §5's *path to lifting the hold*. It does **not**
> disturb that document's other finding: the **effect is still REPRODUCED**, on the
> evidence of the ollama positive control. What failed here is the **instrument**, which
> is precisely the thing this chain existed to test.

---

## 1. The tolerance — still not re-opened

Unchanged from task 3480's problem statement, committed **2026-07-07**, seven days before
this run existed:

> **glyph-only ≤2/8 C** and **grounded ≥6/8 C**.

It is stated here before the numbers, as it was for the control. A tolerance re-derived
after seeing container data would void the pre-registration, and a miss is exactly the
moment the temptation appears. It is declined.

## 2. Observed

Mistral-7B-Instruct-v0.3 Q4_K_M (`sha256:f5074b12…` — the control's own artifact),
temperature 0.8, seeds 1..8, max_tokens 512, executed through the digest-pinned assay
image `sha256:c0d44d61…` on corpos-net against llama-server.

| Condition | Grid | Score |
|---|---|---|
| baseline | C C C C C Ii C Ii | **6/8 C** · **0/8 correct-target** |
| glyph_only | I I I Ii I Ii Ii Ii | **0/8 C** (4 Ii, 4 I) |
| grounded_glyph | Ii C Ii Ii C C C Ii | **4/8 C** |

**Against the pre-registered tolerance:**

| Criterion | Required | Observed | Result |
|---|---|---|---|
| glyph-only | ≤2/8 C | 0/8 C | ✅ met |
| grounded | ≥6/8 C | **4/8 C** | ❌ **MISSED by 2** |

**Against the pinned gates:**

| Gate | Observed | Result |
|---|---|---|
| Calibration | 6/8 C, **0/8 correct-target** — every C uses a placeholder date, `v`-prefix, or bullet form; none writes the real Keep-a-Changelog entry | ✅ passed |
| Sufficiency | glyph_only 0/8 → not sufficient → ground triggered | ✅ matches v3 |
| Ground-lift | grounded 4/8 — clears "≥2 above cond 2" (+4) but **fails the ≥5/8 arm** | ❌ **NOT MET** |

## 3. Three-way comparison

| Condition | v3 original | ollama control | **container v2** | |
|---|---|---|---|---|
| baseline (correct-target C) | 7/8 C, **0/8 c-t** | 6/8 C, **0/8 c-t** | 6/8 C, **0/8 c-t** | ✅ exact |
| glyph_only | **0/8** (7 Ii, 1 I) | **0/8** (6 Ii, 2 I) | **0/8** (4 Ii, 4 I) | ✅ exact |
| grounded_glyph | 7/8 | 8/8 | **4/8** | ❌ **halved** |
| **ground lift** | **+7** | **+8** | **+4** | direction intact |

The shape of this table is the finding, and it is sharper than a flat "diverged".

**Two of three cells reproduce exactly.** Baseline lands 0/8 correct-target on all three
legs. Glyph-only lands **0/8 C on all three legs** — and the mechanism came with it: every
one of the 8 container responses entered axis-by-axis analysis mode, several walking
Marker/Aim/Rest explicitly, two concluding the obligation "remains unmet" and recommending
someone else act, and **not one writing an entry**. The register phenomenon that C2 is
built on reproduced on the container rig without qualification.

**Only the grounded conversion diverges**, and it degrades rather than vanishes: the
ground still converts analysis-mode to execution, in half the runs instead of nearly all.
The four failures are not refusals — they are *prose about the entry* ("the suggested
action is to prepend the following…") where the control produced *the entry itself*. The
strict cond-3 bar scores that Ii, and it is the same bar the control was held to.

## 4. Robustness of the verdict

**Second rater.** Qwen2.5-32B, temperature 0, rubric byte-identical to the control's:
**10/12 = 83%** agreement (control: 6/6). The sample was deliberately *harder* than the
control's — all 8 grounded runs plus 2 baseline and 2 glyph-only — because the grounded
C/Ii line is what the verdict turns on. Both disagreements are recorded in
`runs/container-v2-2026-07-14/double_score_result.json`:

- `glyph_only r3` — I (primary) vs Ii (Qwen). **Cannot matter**: both are non-C, so
  glyph-only is 0/8 either way.
- `grounded r7` — C (primary) vs Ii (Qwen). Moves grounded **down** to 3/8.

**The verdict survives every contestable call, and only gets stronger.** Under the
strictest reading grounded is 3/8; under the most generous it is 4/8. The floor is 6/8.
There is no scoring of this grid that reaches the tolerance.

**Is the gap distinguishable from noise?** Honestly: **not at n=8.**

| Comparison | Fisher two-sided |
|---|---|
| container 4/8 vs control 8/8 | p = 0.077 |
| container 4/8 vs v3 7/8 | p = 0.282 |
| container 3/8 vs control 8/8 (Qwen's reading) | p = 0.026 |

So ordinary sampling variance is **not excluded**. This is stated plainly because it cuts
against the conclusion: it would be easy, and wrong, to present a p=0.077 gap as proof the
rig is broken. **It does not change the verdict.** The tolerance is a pre-registered bright
line, not a significance test, and it was set at ≥6/8 in advance precisely so that a miss
could not be argued away afterwards with a test chosen after the fact. 4/8 misses. The
correct response to "the miss might be noise" is a matched re-run under the same fixed bar
— not a re-interpretation of this one.

## 5. Root cause

Per bug-fixing-discipline, root-caused before touching the instrument. **Two candidates
remain live; neither is proven.**

### Ruled OUT — delivery framing

The leading a-priori suspect (CHANGELOG.md KNOWN DELTA 2) was that llama.cpp's embedded
chat template differs from ollama's. **Checked directly and dismissed.** llama.cpp's
`/apply-template` returns:

```
[INST] HELLO_MARKER [/INST]
```

against ollama's `[INST] {prompt}[/INST]` — a **single space** before `[/INST]`, no
double-BOS. The prompts are materially the same. Delivery framing does not explain a
halved conversion rate, and the delta list is now shorter for having been tested rather
than assumed.

Stop sequences (KNOWN DELTA 3) are also unlikely: the responses terminate cleanly, with no
hallucinated `[INST]` follow-on turns for the missing stops to have prevented.

### Leading candidate — the sampler was never actually pinned

Filed: **`sampling-regime-underspecified-defeats-freeze-by-digest`** (high).

`model.GenParams` carries temperature, max_tokens, seed — and nothing else. Every other
sampler parameter inherits the **server's** default, and the two runtimes disagree:

| param | ollama (control) | llama.cpp (container) |
|---|---|---|
| top_p | 0.9 | **0.95** |
| min_p | 0.0 | **0.05** |
| repeat_penalty | 1.1 | **1.0** |
| top_k | 40 | 40 |

Read live from llama.cpp `GET /props`; the control's runner sent only
temperature/seed/num_predict and so inherited ollama's.

**The two legs sampled different distributions while both reporting "temperature 0.8, seed
N".** Temperature alone does not define a sampler. A looser tail (higher top_p, a min_p
floor) with **no repetition penalty** is a plausible mechanism for drift into
meta-commentary instead of entry production — which is exactly the failure mode of the four
Ii runs. Plausible, not demonstrated.

### The larger problem this exposes

`MANIFEST.v2.sha256` claims sampling is pinned because temperature/seeds/max_tokens ride in
`study.json` under `StudyDigest`. **That claim is false in substance.** The unpinned
majority of the sampler rides on the llama.cpp build's defaults. A future re-run could
match every digest in the manifest and still sample a different distribution — a study
that appears reproducible and is not. That is strictly worse than a loud mismatch, and it
would have propagated silently into C1, C2 and C3.

**The container leg earned its keep here even though it failed.** This hole was invisible
from the ollama control, invisible to the gate, and invisible to the manifest's own
self-description. It took running the rig against a second runtime to expose it.

## 6. Gate decision

### 🛑 HOLD MAINTAINED — science chains may NOT run on corpos-lab

`instrument-parity-reproduction` established the **effect is real** (positive control,
lift +8). This chain asked whether **corpos-lab can measure it**. On the evidence: **not
yet, and not to a known standard.** The rig produced a grid that misses a pre-registered
floor, under a sampler the manifest wrongly claims to have pinned. Lifting the hold now
would put C1/C2/C3 on an instrument that is measurably unable to reproduce the one result
it was checked against, with a freeze that does not cover what it says it covers.

A divergence is a finding about the rig, not a reason to move the bar.

### What the container leg DID establish

Not a wasted run — most of the rig works, and this is the first evidence of any of it:

- The container path executes end-to-end: digest-pinned image, corpos-net DNS to
  llama-server, `/in`→`/out` contract, 24/24 responses captured, run record written,
  persisted to the canonical toolkit DB.
- **The v2 sampling fix works.** All 8 replicates of every cell are **distinct**. Under
  v1's temp-0 instrument all 8 would have been byte-identical and a graded grid impossible.
- **The v2 scoring-honesty fix works.** All 24 rows landed `unscored`. Under v1 all 24
  would have been fabricated fails in the canonical DB.
- **Manifest verification, failure recording, and fail-closed behaviour all work** — the
  first attempt died on a real launcher bug and produced a correct failed-run record with
  digests intact.
- **glyph-only reproduced exactly (0/8, three legs) with its mechanism intact.**

### Path to lifting the hold — v3, one grid, same bar

1. Fix `sampling-regime-underspecified-defeats-freeze-by-digest`: carry the full sampler
   through `GenParams` → study def → `study.json`; read the effective sampler back from the
   server and record it; stop the manifest claiming a pin it does not have.
2. Fix `podman-launcher-relative-paths-become-named-volumes` (host-side, not a bump).
3. Bump to **v3** declaring the **control's** regime: top_p 0.9, min_p 0.0,
   repeat_penalty 1.1, top_k 40, temperature 0.8, seeds 1..8.
4. Re-run the container leg and judge against **this same tolerance, unchanged**.

**Binding, stated now rather than later:** v3 is **one grid**, and its result is reported
whichever way it falls. If a matched sampler restores grounded ≥6/8, the divergence was
instrument-attributable and the hold lifts. If it does not, that is a finding about
corpos-lab and the hold stays. **A third grid may not be run to obtain a better number.**
"Run until it reproduces" is the precise pathology the charter's stopping rules exist to
prevent, and the pull toward it is strongest exactly here — a near-miss with a ready
explanation. Declared in advance so it cannot be rationalised in the moment.

---

## Provenance

- Tolerance: task 3480, created 2026-07-07T23:17:44Z (`TaskCreated`), unmodified.
- Gates: `SCORING_RUBRIC.md` (MANIFEST-digested) + CHARTER.md v1, frozen 2026-07-13.
- Run: `runs/container-v2-2026-07-14/` — `out/results.json` (24 rows, all `unscored`),
  `out/responses/*.txt` (24), `run-record.json`, `double_score_result.json`.
- Control: `runs/ollama-repro-2026-07-13/SCORE_GRID.md` · Original: `PARITY_TARGET.md`.
- Deltas: `CHANGELOG.md` §KNOWN DELTAS (delta 2 now **ruled out** by `/apply-template`).
- Bugs filed: `sampling-regime-underspecified-defeats-freeze-by-digest` (high, root-cause
  candidate) · `podman-launcher-relative-paths-become-named-volumes` (high, host-side).
